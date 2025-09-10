package handlers

import (
	"SmartSpend/internal/domain/model"
	_ "SmartSpend/internal/repository"
	_ "database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type GoogleTokenRequest struct {
	IDToken string `json:"id_token"`
}
type AppleUserRequest struct {
	UserId    string `json:"user_id"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

func (s *Server) GoogleAuth(c *gin.Context) {
	var req GoogleTokenRequest
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
		username := strings.Split(email, "@")[0]
		newUser := model.User{
			ID: uuid.New().String(),
			FirstName: func() string {
				if v, ok := payload.Claims["given_name"].(string); ok {
					return v
				}
				return ""
			}(),
			LastName: func() string {
				if v, ok := payload.Claims["family_name"].(string); ok {
					return v
				}
				return ""
			}(),
			Username:               username,
			GoogleEmail:            email,
			RefreshToken:           tokenService.GenerateRefreshToken(),
			RefreshTokenExpiryDate: time.Now().Add(30 * 24 * time.Hour),
			AvatarURL: func() string {
				if v, ok := payload.Claims["picture"].(string); ok {
					return v
				}
				return ""
			}(),
			CreatedAt:         time.Now(),
			Balance:           0.00,
			MonthlySavingGoal: 0.00,
			PreferredCurrency: "MKD",
		}

		userService.Save(newUser)
		userDto, _ := applicationUserService.FindById(newUser.ID)
		c.JSON(http.StatusOK, gin.H{
			"message":                   "Successfully signed up with Google!",
			"refresh_token":             newUser.RefreshToken,
			"refresh_token_expiry_date": newUser.RefreshTokenExpiryDate,
			"access_token":              tokenService.GenerateAccessToken(newUser.ID),
			"data":                      userDto,
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
			"data":                      userDto,
		})
		return
	}
}

func (s *Server) AppleSignIn(c *gin.Context) {
	var req AppleUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	existingUser, _ := userService.FindById(req.UserId)

	if existingUser != nil {
		// login
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

	// signup
	username := strings.Split(req.Email, "@")[0]
	newUser := model.User{
		ID:                     req.UserId,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Username:               username,
		GoogleEmail:            "",
		AppleEmail:             req.Email,
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
