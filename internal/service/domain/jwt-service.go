package domain

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var ()

type IJWTService interface {
	GenerateAccessToken(userId uuid.UUID) string
	ValidateAccessToken(tokenString string) (*jwt.Token, error)
	ExtractClaims(tokenString string) (uuid.UUID, error)
}

type JWTCustomClaims struct {
	UserID uuid.UUID `json:"UserID"`
	jwt.StandardClaims
}

type JwtService struct {
	secretKey string
	issuer    string
}

func NewJWTService() IJWTService {
	return &JwtService{
		secretKey: os.Getenv("ACCESS_TOKEN_SECRET_KEY"),
		issuer:    "Smart Spend Team",
	}
}

func (jwtSrv *JwtService) GenerateAccessToken(userId uuid.UUID) string {
	claims := &JWTCustomClaims{
		UserID: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
			Issuer:    jwtSrv.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(jwtSrv.secretKey))
	if err != nil {
		panic(err)
	}
	return t
}

func (jwtSrv *JwtService) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSrv.secretKey), nil
	})
}

func (jwtSrv *JwtService) ExtractClaims(tokenString string) (uuid.UUID, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTCustomClaims{})
	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*JWTCustomClaims); ok {
		return claims.UserID, nil
	}

	return uuid.Nil, fmt.Errorf("invalid token claims")
}
