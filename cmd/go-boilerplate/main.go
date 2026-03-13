package main

import (
	"context"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kevalsabhani/go-boilerplate/internal/config"
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

	shutdown, err := observability.InitTracer(cfg.Primary.Env)
	if err != nil {
		logger.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer shutdown(context.Background())
}
