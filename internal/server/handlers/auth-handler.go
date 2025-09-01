package handlers

import (
	"SmartSpend/internal/domain/model"
	_ "SmartSpend/internal/repository"
	"crypto/rsa"
	_ "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type TokenRequest struct {
	IDToken string `json:"id_token"`
}

func (s *Server) GoogleAuth(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate token against one of your client IDs
	validClients := []string{
		os.Getenv("GOOGLE_WEB_CLIENT_ID"),
		os.Getenv("GOOGLE_IOS_CLIENT_ID"),
	}

	var payload *idtoken.Payload
	var err error
	for _, clientID := range validClients {
		payload, err = idtoken.Validate(c, req.IDToken, clientID)
		if err == nil {
			break
		}
	}
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "token expired") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired id_token"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid id_token"})
		return
	}

	// double check aud
	aud := payload.Claims["aud"].(string)
	ok := false
	for _, clientID := range validClients {
		if aud == clientID {
			ok = true
			break
		}
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid audience"})
		return
	}

	// __________________________
	email := payload.Claims["email"].(string)

	existingUser := userService.FindByGoogleEmail(email)

	if existingUser == nil {
		// Signup flow
		username := strings.Split(email, "@")[0]
		newUser := model.User{
			ID:                     uuid.New(),
			FirstName:              payload.Claims["given_name"].(string),
			LastName:               payload.Claims["family_name"].(string),
			Username:               username,
			GoogleEmail:            email,
			RefreshToken:           tokenService.GenerateRefreshToken(),
			RefreshTokenExpiryDate: time.Now().Add(30 * 24 * time.Hour),
			AvatarURL:              payload.Claims["picture"].(string),
			CreatedAt:              time.Now(),
			Balance:                0.00,
			MonthlySavingGoal:      0.00,
			PreferredCurrency:      "MKD",
		}

		userService.Save(newUser)
		userDto, _ := applicationUserService.FindById(newUser.ID)
		c.JSON(http.StatusOK, gin.H{
			"message":                   "Successfully signed up with Google!",
			"refresh_token":             newUser.RefreshToken,
			"refresh_token_expiry_date": newUser.RefreshTokenExpiryDate,
			"access_token":              tokenService.GenerateAccessToken(newUser.ID),
			"user_data":                 userDto,
		})
		return
	}

	// user already exists
	existingUser.RefreshToken = tokenService.GenerateRefreshToken()
	existingUser.RefreshTokenExpiryDate = time.Now().Add(30 * 24 * time.Hour)
	userDto, _ := applicationUserService.FindById(existingUser.ID)

	if err := userService.Update(*existingUser); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":                   "Successfully logged in! Welcome back.",
			"refresh_token":             existingUser.RefreshToken,
			"refresh_token_expiry_date": existingUser.RefreshTokenExpiryDate,
			"access_token":              tokenService.GenerateAccessToken(existingUser.ID),
			"user_data":                 userDto,
		})
		return
	}
}

var appleJWKSURL = "https://appleid.apple.com/auth/keys"

type AppleJWKS struct {
	Keys []AppleJWK `json:"keys"`
}

type AppleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func getApplePublicKey(kid string) (*rsa.PublicKey, error) {
	resp, err := http.Get(appleJWKSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks endpoint returned status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read jwks response: %w", err)
	}

	var jwks AppleJWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jwks: %w", err)
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			if key.Kty != "RSA" {
				return nil, fmt.Errorf("unsupported key type: %s", key.Kty)
			}
			if key.Alg != "RS256" {
				return nil, fmt.Errorf("unsupported algorithm: %s", key.Alg)
			}

			pubKey, err := parseRSAPublicKey(key.N, key.E)
			if err != nil {
				return nil, fmt.Errorf("failed to parse rsa key: %w", err)
			}
			return pubKey, nil
		}
	}
	return nil, errors.New("no matching key found")
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := jwt.DecodeSegment(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N parameter: %w", err)
	}
	eBytes, err := jwt.DecodeSegment(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E parameter: %w", err)
	}

	if len(nBytes) == 0 || len(eBytes) == 0 {
		return nil, errors.New("invalid RSA parameters: empty N or E")
	}

	n := new(big.Int).SetBytes(nBytes)
	if n.Sign() <= 0 {
		return nil, errors.New("invalid RSA parameter N: must be positive")
	}

	// Convert E bytes to int more safely
	var e int
	if len(eBytes) > 4 {
		return nil, errors.New("E parameter too large")
	}
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e <= 0 {
		return nil, errors.New("invalid RSA parameter E: must be positive")
	}

	pubKey := &rsa.PublicKey{N: n, E: e}

	// Validate key size (Apple uses 2048-bit keys)
	if pubKey.N.BitLen() < 2048 {
		return nil, errors.New("RSA key too small")
	}

	return pubKey, nil
}

func ValidateAppleIDToken(idToken string) (*jwt.MapClaims, error) {
	if strings.TrimSpace(idToken) == "" {
		return nil, errors.New("empty ID token")
	}

	parser := jwt.Parser{
		ValidMethods: []string{"RS256"}, // only RS256
	}
	claims := jwt.MapClaims{}

	token, _, err := parser.ParseUnverified(idToken, claims)
	if err != nil {
		return nil, fmt.Errorf("parse unverified: %w", err)
	}

	// validate header
	alg, ok := token.Header["alg"].(string)
	if !ok || alg != "RS256" {
		return nil, errors.New("invalid or missing algorithm in token header")
	}

	typ, ok := token.Header["typ"].(string)
	if !ok || typ != "JWT" {
		return nil, errors.New("invalid or missing token type in header")
	}

	kid, ok := token.Header["kid"].(string)
	if !ok || strings.TrimSpace(kid) == "" {
		return nil, errors.New("kid not found in token header")
	}

	// get Apple's public key
	pubKey, err := getApplePublicKey(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// verify token with proper validation
	verifiedClaims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(idToken, verifiedClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	if !parsedToken.Valid {
		return nil, errors.New("token is not valid")
	}

	// validate required claims
	iss, ok := verifiedClaims["iss"].(string)
	if !ok || iss != "https://appleid.apple.com" {
		return nil, errors.New("invalid or missing issuer")
	}

	aud, ok := verifiedClaims["aud"].(string)
	if !ok || aud != os.Getenv("APPLE_CLIENT_ID") {
		return nil, errors.New("invalid or missing audience")
	}

	// Validate expiration
	exp, ok := verifiedClaims["exp"].(float64)
	if !ok {
		return nil, errors.New("missing expiration claim")
	}
	if int64(exp) <= time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	// validate issued at time (iat)
	iat, ok := verifiedClaims["iat"].(float64)
	if !ok {
		return nil, errors.New("missing issued at claim")
	}
	if int64(iat) > time.Now().Unix() {
		return nil, errors.New("token issued in the future")
	}

	// Validate not before (if present)
	if nbf, ok := verifiedClaims["nbf"].(float64); ok {
		if int64(nbf) > time.Now().Unix() {
			return nil, errors.New("token not yet valid")
		}
	}

	// Validate subject
	sub, ok := verifiedClaims["sub"].(string)
	if !ok || strings.TrimSpace(sub) == "" {
		return nil, errors.New("missing or empty subject claim")
	}

	return &verifiedClaims, nil
}

func (s *Server) AppleSignIn(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	claims, err := ValidateAppleIDToken(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Apple ID token"})
		return
	}

	email, _ := (*claims)["email"].(string)

	existingUser := userService.FindByAppleEmail(email)

	if existingUser != nil {
		// Login
		existingUser.RefreshToken = tokenService.GenerateRefreshToken()
		existingUser.RefreshTokenExpiryDate = time.Now().Add(30 * 24 * time.Hour)

		if err := userService.Update(*existingUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		userDto, _ := applicationUserService.FindById(existingUser.ID)
		c.JSON(http.StatusOK, gin.H{
			"message":                   "Successfully logged in with Apple!",
			"refresh_token":             existingUser.RefreshToken,
			"refresh_token_expiry_date": existingUser.RefreshTokenExpiryDate,
			"access_token":              tokenService.GenerateAccessToken(existingUser.ID),
			"user_data":                 userDto,
		})
		return
	}

	// Create new user
	newUser := model.User{
		ID:                     uuid.New(),
		FirstName:              "",
		LastName:               "",
		Username:               "",
		GoogleEmail:            "",
		AppleEmail:             email,
		RefreshToken:           tokenService.GenerateRefreshToken(),
		RefreshTokenExpiryDate: time.Now().Add(30 * 24 * time.Hour),
		AvatarURL:              "",
		CreatedAt:              time.Now(),
		Balance:                0.00,
		MonthlySavingGoal:      0.00,
		PreferredCurrency:      "MKD",
	}

	userService.Save(newUser)

	userDto, _ := applicationUserService.FindById(newUser.ID)
	c.JSON(http.StatusOK, gin.H{
		"message":                   "Successfully signed up with Apple!",
		"refresh_token":             newUser.RefreshToken,
		"refresh_token_expiry_date": newUser.RefreshTokenExpiryDate,
		"access_token":              tokenService.GenerateAccessToken(newUser.ID),
		"user_data":                 userDto,
	})
}

func (s *Server) Logout(c *gin.Context) {
	_, userId := getUserFromDatabase(c)
	user, err := userService.FindById(userId)
	if err == nil {
		user.RefreshToken = ""
		user.RefreshTokenExpiryDate = time.Time{}
		err := userService.Update(*user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully logged-out!",
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	return
}
