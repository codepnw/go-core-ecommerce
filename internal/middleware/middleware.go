package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/codepnw/go-starter-kit/internal/auth"
	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/features/user"
	jwttoken "github.com/codepnw/go-starter-kit/pkg/jwttoken"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Middleware struct {
	token jwttoken.JWTToken
	redis *redis.Client
}

func InitMiddleware(token jwttoken.JWTToken, redis *redis.Client) *Middleware {
	return &Middleware{
		token: token,
		redis: redis,
	}
}

func (m *Middleware) Authorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ResponseError(c, http.StatusUnauthorized, errors.New("header is misstion"))
			c.Abort()
			return
		}

		args := strings.Fields(authHeader)
		authToken := args[1]

		if len(args) != 2 || args[0] != "Bearer" {
			response.ResponseError(c, http.StatusUnauthorized, errors.New("invalid token format"))
			c.Abort()
			return
		}

		// Check Redis Token Blacklist
		isBlacklisted, err := m.isTokenBlacklisted(c.Request.Context(), authToken)
		if err != nil {
			response.ResponseError(c, http.StatusInternalServerError, err)
			c.Abort()
			return
		}
		if isBlacklisted {
			response.ResponseError(c, http.StatusUnauthorized, errors.New("token is been revoked (logged out)"))
			c.Abort()
			return
		}

		// Verify JWT Token
		claims, err := m.token.VerifyAccessToken(authToken)
		if err != nil {
			response.ResponseError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, config.ContextUserClaimsKey, claims)
		ctx = context.WithValue(ctx, config.ContextUserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, config.ContextUserRoleKey, claims.Role)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (m *Middleware) AdminOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRole, err := auth.GetUserRoleFromContext(ctx.Request.Context())
		if err != nil {
			log.Println("Error get context role:", err)
			response.ResponseError(ctx, http.StatusForbidden, errors.New("no permissions"))
			ctx.Abort()
			return
		}

		if userRole != string(user.RoleAdmin) {
			response.ResponseError(ctx, http.StatusForbidden, errors.New("no permissions"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func (m *Middleware) Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		ctx.Next()

		duration := time.Since(start)
		status := ctx.Writer.Status()
		method := ctx.Request.Method
		clientIP := ctx.ClientIP()

		attrs := []any{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("ip", clientIP),
			slog.Duration("latency", duration),
		}

		if raw != "" {
			attrs = append(attrs, slog.String("query", raw))
		}

		if status >= 500 {
			slog.Error("Request failed", attrs...)
		} else if status >= 400 {
			slog.Warn("Bad request", attrs...)
		} else {
			slog.Info("Request success", attrs...)
		}
	}
}

func (m *Middleware) isTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)

	_, err := m.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Not Found in Blacklist, Can Logout
	}
	if err != nil {
		return false, err // Error from redis
	}

	return true, nil // Found in Blacklist
}
