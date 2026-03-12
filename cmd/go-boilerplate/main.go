package main

import (
	"context"
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kevalsabhani/go-boilerplate/internal/config"
	"github.com/kevalsabhani/go-boilerplate/internal/observability"
	"go.uber.org/zap"
)

const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Main entrypoint
func main() {
	// Load config
	cfg := config.MustLoad("")

	// Initialize logger
	logger := initLogger(cfg.Primary.Env)
	defer func() {
		_ = logger.Sync()
	}()

	shutdown, err := observability.InitTracer(cfg.Primary.Env)
	if err != nil {
		logger.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer shutdown(context.Background())
}

// initLogger provides zap logger instance based environment
func initLogger(env string) *zap.Logger {
	var (
		logger *zap.Logger
		err    error
	)
	switch env {
	case EnvStaging, EnvProduction:
		logger, err = zap.NewProduction()
	default:
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}

	return logger
}
