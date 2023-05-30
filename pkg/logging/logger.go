// Package logging provides wrappers around zap logging.
package logging

import (
	"log"

	"go.uber.org/zap"
)

// Logger wraps a zap.SugaredLogger.
type Logger struct {
	*zap.SugaredLogger
}

// Environment enumerates predefined logging configurations.
type Environment int

const (
	// ProductionLogger is a short hand for predefined logging configuration for production use.
	ProductionLogger Environment = iota
	// DevelopmentLogger is a short hand for predefined logging configuration for development use.
	DevelopmentLogger
)

func createLogger(logEnv Environment) *Logger {
	var logger *zap.Logger

	var err error

	switch logEnv {
	case ProductionLogger:
		logger, err = zap.NewProduction()

	case DevelopmentLogger:
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	//nolint: errcheck
	defer logger.Sync()

	return &Logger{SugaredLogger: logger.Sugar()}
}

// CreateProductionLogger creates a new production logger.
func CreateProductionLogger() *Logger {
	return createLogger(ProductionLogger)
}

// CreateDevelopmentLogger creates a new development logger.
func CreateDevelopmentLogger() *Logger {
	return createLogger(DevelopmentLogger)
}
