package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL = 15 * time.Minute

type TokenService struct {
	secret []byte
}

func NewTokenServiceFromEnv() (*TokenService, error) {
	secret := os.Getenv("AUTH_JWT_SECRET")
	if secret == "" {
		return nil, errors.New("AUTH_JWT_SECRET is required")
	}
	return &TokenService{secret: []byte(secret)}, nil
}

func (s *TokenService) GenerateAccessToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *TokenService) ParseAccessToken(token string) (*jwt.RegisteredClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(_ *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
