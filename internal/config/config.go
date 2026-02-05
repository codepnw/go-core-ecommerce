package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/codepnw/go-starter-kit/pkg/utils/validate"
	"github.com/joho/godotenv"
)

type contextKey string

const (
	// JWT token duration
	AccessTokenDuration  = time.Minute * 30
	RefreshTokenDuration = time.Hour * 24 * 7

	// Context keys
	ContextUserClaimsKey contextKey = "ctx-user-claims"
	ContextUserIDKey     contextKey = "ctx-user-id"

	ContextTimeout = time.Second * 10
)

type EnvConfig struct {
	APP AppConfig `envPrefix:"APP_"`
	DB  DBConfig  `envPrefix:"DB_"`
	JWT JWTConfig `envPrefix:"JWT_"`
}

type AppConfig struct {
	Host   string `env:"HOST" envDefault:"localhost"`
	Port   int    `env:"PORT" envDefault:"8080"`
	Prefix string `env:"PREFIX" envDefault:"/api/v1"`
}

type DBConfig struct {
	User     string `env:"USER" validate:"required"`
	Password string `env:"PASSWORD" validate:"required"`
	Name     string `env:"NAME" validate:"required"`
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"5432"`
	SSLMode  string `env:"SSL_MODE" envDefault:"disable"`
}

type JWTConfig struct {
	AppName    string `env:"APP_NAME" envDefault:"Go Starter Kit"`
	SecretKey  string `env:"SECRET_KEY" validate:"required"`
	RefreshKey string `env:"REFRESH_KEY" validate:"required"`
}

func LoadConfig(path string) (*EnvConfig, error) {
	// Load .env file
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("load env failed: %w", err)
	}

	cfg := new(EnvConfig)
	// Parse env tags
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env failed: %w", err)
	}

	// Validate config
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("validate env failed: %w", err)
	}
	return cfg, nil
}

// GetDatabaseDSN : Default PostgreSQL
func (cfg *EnvConfig) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)
}

func (cfg *EnvConfig) GetAppAddress() string {
	return fmt.Sprintf(":%d", cfg.APP.Port)
}
