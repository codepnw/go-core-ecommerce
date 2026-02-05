package jwttoken

import (
	"errors"
	"fmt"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/features/user"
	"github.com/golang-jwt/jwt/v5"
)

//go:generate mockgen -source=jwt.go -destination=jwt_mock.go -package=jwttoken
type JWTToken interface {
	GenerateAccessToken(u *user.User) (string, error)
	GenerateRefreshToken(u *user.User) (string, error)
	VerifyAccessToken(tokenStr string) (*UserClaims, error)
	VerifyRefreshToken(tokenStr string) (*UserClaims, error)
}

type token struct {
	appName    string
	secretKey  string
	refreshKey string
}

func NewJWTToken(appName, secretKey, refreshKey string) (JWTToken, error) {
	if secretKey == "" || refreshKey == "" {
		return nil, errors.New("secret & refresh key is required")
	}
	return &token{
		appName:    appName,
		secretKey:  secretKey,
		refreshKey: refreshKey,
	}, nil
}

type UserClaims struct {
	UserID string
	Email  string
	*jwt.RegisteredClaims
}

// ------------- Generate Token ----------------

func (j *token) GenerateAccessToken(u *user.User) (string, error) {
	return j.generateToken(j.secretKey, u, config.AccessTokenDuration)
}

func (j *token) GenerateRefreshToken(u *user.User) (string, error) {
	return j.generateToken(j.refreshKey, u, config.RefreshTokenDuration)
}

func (j *token) generateToken(key string, u *user.User, duration time.Duration) (string, error) {
	claims := &UserClaims{
		UserID: u.ID,
		Email:  u.Email,
		RegisteredClaims: &jwt.RegisteredClaims{
			Subject:   u.ID,
			Issuer:    j.appName,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("sign token failed: %w", err)
	}
	return ss, nil
}

// ------------- Verify Token ----------------

func (j *token) VerifyAccessToken(tokenStr string) (*UserClaims, error) {
	return j.verifyToken(j.secretKey, tokenStr)
}

func (j *token) VerifyRefreshToken(tokenStr string) (*UserClaims, error) {
	return j.verifyToken(j.refreshKey, tokenStr)
}

func (j *token) verifyToken(key, tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(key), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token failed: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("type assertion claims failed")
	}
	return claims, nil
}
