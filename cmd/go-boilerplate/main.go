package main

import (
	"context"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kevalsabhani/go-boilerplate/internal/config"
	"github.com/kevalsabhani/go-boilerplate/internal/database"
	"github.com/kevalsabhani/go-boilerplate/internal/logger"
	"github.com/kevalsabhani/go-boilerplate/internal/observability"
	"go.uber.org/zap"
)

// Main entrypoint
func main() {
	// Load config
	cfg := config.MustLoad("")

	// Initialize logger
	logger := logger.Initialize(cfg.Primary.Env)
	defer func() {
		_ = logger.Sync()
	}()

	// app context - cancelled on OS signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize tracer
	shutdown, err := observability.InitTracer(cfg.Primary.Env)
	if err != nil {
		logger.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer shutdown(context.Background())

	// Initialize database
	db, err := database.New(ctx, cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()
}
