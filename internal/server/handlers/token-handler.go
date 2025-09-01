package handlers

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func unauthorized(c *gin.Context, errorMessage string) {
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
			// extract UserId from token
			claims, err := tokenService.DecodeExpiredAccessToken(accessToken)
			if err != nil {
				unauthorized(c, "Cannot decode expired token: "+err.Error())
				return
			}

			userIDStr, ok := claims["UserID"].(string)
			if !ok {
				unauthorized(c, "UserID not found in token")
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				unauthorized(c, "Invalid UserID format")
				return
			}

			user, err := userService.FindById(userID)
			if err != nil {
				unauthorized(c, "User not found")
				return
			}

			if user.RefreshToken != refreshToken {
				unauthorized(c, "Refresh token does not match")
				return
			}

			newAccessToken := tokenService.GenerateAccessToken(userID)
			c.JSON(http.StatusOK, gin.H{
				"userId":       userID,
				"access_token": newAccessToken,
			})
			return
		}

		unauthorized(c, "Invalid Access Token. Cannot rotate, Error: "+err.Error())
		return
	}
}
