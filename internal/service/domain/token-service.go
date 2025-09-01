package domain

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/repository"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var (
	dbService                                 = database.New()
	userRepository repository.IUserRepository = repository.NewUserRepository(dbService)
	userService    IUserService               = NewUserService(userRepository)
	jwtService     IJWTService                = NewJWTService()
)

type ITokenService interface {
	GenerateRefreshToken() string
	ValidateRefreshToken(refreshToken string, userId uuid.UUID) (error, bool)
	RotateRefreshToken(refreshToken string, userId uuid.UUID) string

	GenerateAccessToken(userId uuid.UUID) string
	ValidateAccessToken(accessToken string) (*jwt.Token, error)
	DecodeAccessToken(accessToken string) (map[string]interface{}, error)
	DecodeExpiredAccessToken(tokenString string) (map[string]interface{}, error)
}

type TokenService struct {
}

func NewTokenService() ITokenService {
	return &TokenService{}
}

func (t *TokenService) GenerateRefreshToken() string {
	bytes := make([]byte, 64)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (t *TokenService) ValidateRefreshToken(refreshToken string, userId uuid.UUID) (error, bool) {
	user, _ := userService.FindById(userId)
	if user == nil {
		return errors.New("user not found"), false
	}
	if user.RefreshToken != refreshToken {
		return errors.New("refresh token does not match"), false
	}
	return nil, true
}

func (t *TokenService) RotateRefreshToken(refreshToken string, userId uuid.UUID) string {
	err, valid := t.ValidateRefreshToken(refreshToken, userId)
	if err == nil && valid == true {
		newRefreshToken := t.GenerateRefreshToken()
		user, _ := userService.FindById(userId)
		user.RefreshToken = newRefreshToken
		user.RefreshTokenExpiryDate = time.Now().Add(30 * 24 * time.Hour)
		userRepository.Save(*user)
		return newRefreshToken
	}
	return ""
}

func (t *TokenService) GenerateAccessToken(userId uuid.UUID) string {
	return jwtService.GenerateAccessToken(userId)
}

func (t *TokenService) ValidateAccessToken(accessToken string) (*jwt.Token, error) {
	return jwtService.ValidateAccessToken(accessToken)
}

func (t *TokenService) DecodeAccessToken(tokenString string) (map[string]interface{}, error) {
	token, err := t.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

func (t *TokenService) DecodeExpiredAccessToken(tokenString string) (map[string]interface{}, error) {
	// Parse the token WITHOUT validation (this allows expired tokens)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure the token method conforms to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_TOKEN_SECRET_KEY")), nil // Use your secret key here
	})

	// Check if the error is ONLY due to expiration
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			// If it's only expired, we can still extract claims
			if ve.Errors == jwt.ValidationErrorExpired {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					return claims, nil
				}
			}
		}
		// If it's any other error, return it
		return nil, err
	}

	// If no error, extract claims normally
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}
