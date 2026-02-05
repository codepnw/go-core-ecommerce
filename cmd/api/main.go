package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/server"
	"github.com/codepnw/go-starter-kit/pkg/database"
)

const envPath = ".env"

func main() {
	// Load Config
	cfg, err := config.LoadConfig(envPath)
	if err != nil {
		log.Fatal(err)
	}

	// Connect Database
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// HTTP Server
	srv, err := server.NewServer(cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	httpSrv := &http.Server{
		Addr:         cfg.GetAppAddress(),
		Handler:      srv.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("server existing")
}
