package middleware

import (
	database2 "SmartSpend/internal/database"
	"SmartSpend/internal/repository"
	"SmartSpend/internal/service/domain"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var (
	database       database2.Service          = database2.New()
	userRepository repository.IUserRepository = repository.NewUserRepository(database)
	userService    domain.IUserService        = domain.NewUserService(userRepository)
	tokenService   domain.ITokenService       = domain.NewTokenService()
)

func unauthorized(c *gin.Context, errorMessage string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
	return
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(c, "Authorization required")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			unauthorized(c, "Missing Bearer prefix")
			return
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := tokenService.ValidateAccessToken(accessToken)

		if err != nil {
			ve, ok := err.(*jwt.ValidationError)
			if ok && ve.Errors == jwt.ValidationErrorExpired {
				unauthorized(c, "Access token expired")
				return
			}
			unauthorized(c, "Invalid Access Token")
			return
		}

		if !token.Valid {
			unauthorized(c, "Invalid Access Token")
			return
		}

		// Token valid-> continue
		c.Next()
	}
}
