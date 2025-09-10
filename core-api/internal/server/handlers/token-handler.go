package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func unauthorized(c *gin.Context, errorMessage string) {
	log.Println("Unauthorized:", errorMessage)
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
}

func (s *Server) RotateAccessToken(c *gin.Context) {
	refreshToken := c.GetHeader("Refresh-Token")
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		unauthorized(c, "Authorization required")
		return
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	_, err := tokenService.ValidateAccessToken(accessToken)
	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if ok && ve.Errors == jwt.ValidationErrorExpired {
			log.Println("Access token expired, rotating...")

			claims, err := tokenService.DecodeExpiredAccessToken(accessToken)
			if err != nil {
				unauthorized(c, "Cannot decode expired token: "+err.Error())
				return
			}

			userID, ok := claims["user-id"].(string)
			if !ok || userID == "" {
				unauthorized(c, "user-id not found in token")
				return
			}

			user, err := userService.FindById(userID)
			if err != nil {
				unauthorized(c, "User not found: "+err.Error())
				return
			}

			if user.RefreshToken != refreshToken {
				unauthorized(c, "Refresh token does not match")
				return
			}

			newAccessToken := tokenService.GenerateAccessToken(userID)
			c.JSON(http.StatusOK, gin.H{
				"message":      "Access token rotated successfully",
				"access_token": newAccessToken,
			})
			return
		}

		unauthorized(c, "Invalid Access Token. Error: "+err.Error())
		return
	}

	log.Println("Access token valid, no rotation needed")
	c.JSON(http.StatusOK, gin.H{
		"message":      "Access token is still valid, rotation not needed",
		"access_token": accessToken,
	})
}
