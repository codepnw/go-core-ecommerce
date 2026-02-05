package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	_ "github.com/lib/pq"
)

func ConnectPostgres(cfg *config.EnvConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("db connect failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}
	return db, nil
}
