package service

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/repository"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
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
	RotateRefreshToken(refreshToken string, userId uuid.UUID)

	GenerateAccessToken(userId uuid.UUID) string
	ValidateAccessToken(jwtToken string) (*jwt.Token, error)
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
	user := userService.FindById(userId)
	if user == nil {
		return errors.New("user not found"), false
	}
	if user.RefreshToken != refreshToken {
		return errors.New("refresh token does not match"), false
	}
	return nil, true
}

func (t *TokenService) RotateRefreshToken(refreshToken string, userId uuid.UUID) {
	err, valid := t.ValidateRefreshToken(refreshToken, userId)
	if err == nil && valid == true {
		user := userService.FindById(userId)
		user.RefreshToken = t.GenerateRefreshToken()
		user.RefreshTokenExpiryDate = time.Now().Add(30 * 24 * time.Hour)
		userRepository.Save(*user)
	}
}

func (t *TokenService) GenerateAccessToken(userId uuid.UUID) string {
	return jwtService.GenerateAccessToken(userId)
}

func (t *TokenService) ValidateAccessToken(jwtToken string) (*jwt.Token, error) {
	return jwtService.ValidateAccessToken(jwtToken)
}
