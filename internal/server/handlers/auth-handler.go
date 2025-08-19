package handlers

import (
	database2 "SmartSpend/internal/database"
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"
	_ "SmartSpend/internal/repository"
	"SmartSpend/internal/service"
	_ "SmartSpend/internal/service"
	"crypto/rsa"
	_ "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	database       database2.Service          = database2.New()
	userRepository repository.IUserRepository = repository.NewUserRepository(database)
	userService    service.IUserService       = service.NewUserService(userRepository)
	tokenService   service.ITokenService      = service.NewTokenService()
)

type TokenRequest struct {
	IDToken string `json:"id_token"`
}

func (s *Server) GoogleSignIn(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := idtoken.Validate(c, req.IDToken, os.Getenv("GOOGLE_WEB_CLIENT_ID"))
	if err != nil {
		errMsg := err.Error()

		if strings.Contains(errMsg, "token expired") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired id_token"})
		}

		return
	}

	user := model.User{
		ID:                     uuid.New(),
		FirstName:              payload.Claims["given_name"].(string),
		LastName:               payload.Claims["family_name"].(string),
		Username:               strings.Split(payload.Claims["email"].(string), "@")[0],
		GoogleEmail:            payload.Claims["email"].(string),
		AppleEmail:             "",
		RefreshToken:           tokenService.GenerateRefreshToken(),
		RefreshTokenExpiryDate: time.Now().Add(30 * 24 * time.Hour),
		AvatarURL:              payload.Claims["picture"].(string),
		CreatedAt:              time.Now(),
		Balance:                0.00,
		MonthlySavingGoal:      0.00,
	}

	userService.Save(user)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Signed In with Google!",
		"refresh_token": user.RefreshToken,
		"access_token":  tokenService.GenerateAccessToken(user.ID),
	})
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

	body, _ := ioutil.ReadAll(resp.Body)

	var jwks AppleJWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jwks: %w", err)
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			// Convert N/E (base64url) → RSA public key
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
		return nil, err
	}
	eBytes, err := jwt.DecodeSegment(eStr)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)

	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

func ValidateAppleIDToken(idToken string) (*jwt.MapClaims, error) {
	parser := new(jwt.Parser)
	claims := jwt.MapClaims{}

	token, _, err := parser.ParseUnverified(idToken, claims)
	if err != nil {
		return nil, fmt.Errorf("parse unverified: %w", err)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("kid not found in token header")
	}

	// Get Apple's public key
	pubKey, err := getApplePublicKey(kid)
	if err != nil {
		return nil, err
	}

	// Verify token
	verifiedClaims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(idToken, verifiedClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Validate claims
	if verifiedClaims["iss"] != "https://appleid.apple.com" {
		return nil, errors.New("invalid issuer")
	}
	if verifiedClaims["aud"] != os.Getenv("APPLE_CLIENT_ID") {
		return nil, errors.New("invalid audience")
	}
	if exp, ok := verifiedClaims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return nil, errors.New("token expired")
		}
	}

	return &verifiedClaims, nil
}

func (s *Server) AppleSignIn(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := ValidateAppleIDToken(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Always use "sub" as the unique user ID from Apple

	email, _ := (*claims)["email"].(string)

	user := model.User{
		ID:                     uuid.New(),
		FirstName:              "", // to be entered after
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
	}

	userService.Save(user)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Signed in with Apple!",
		"refresh_token": tokenService.GenerateRefreshToken(),
		"access_token":  tokenService.GenerateAccessToken(user.ID),
	})
}
