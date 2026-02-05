package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	jwttoken "github.com/codepnw/go-starter-kit/pkg/jwttoken"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	token jwttoken.JWTToken
}

func InitMiddleware(token jwttoken.JWTToken) *Middleware {
	return &Middleware{token: token}
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
		if len(args) != 2 || args[0] != "Bearer" {
			response.ResponseError(c, http.StatusUnauthorized, errors.New("invalid token format"))
			c.Abort()
			return
		}

		claims, err := m.token.VerifyAccessToken(args[1])
		if err != nil {
			response.ResponseError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, config.ContextUserClaimsKey, claims)
		ctx = context.WithValue(ctx, config.ContextUserIDKey, claims.UserID)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
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
