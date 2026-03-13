package logger

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Initialize provides zap logger instance based environment
func Initialize(env string) *zap.Logger {
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
