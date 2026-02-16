package auth

import (
	"context"
	"errors"

	"github.com/codepnw/go-starter-kit/internal/config"
	jwttoken "github.com/codepnw/go-starter-kit/pkg/jwttoken"
)

func GetUserFromContext(ctx context.Context) (*jwttoken.UserClaims, error) {
	claims, ok := ctx.Value(config.ContextUserClaimsKey).(*jwttoken.UserClaims)
	if !ok {
		return nil, errors.New("get user context failed")
	}
	return claims, nil
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(config.ContextUserIDKey).(string)
	if !ok {
		return "", errors.New("get user id context failed")
	}
	return userID, nil
}

func GetUserRoleFromContext(ctx context.Context) (string, error) {
	role, ok := ctx.Value(config.ContextUserRoleKey).(string)
	if !ok {
		return "", errors.New("get role context failed")
	}
	return role, nil
}

func SetContextUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, config.ContextUserIDKey, userID)
}