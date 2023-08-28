// Package logging provides wrappers around zap logging.
package logging

import (
	"log"

	"go.uber.org/zap"
)

// LogLevel enumerates log levels
type LogLevel int

const (
	// DEBUG is the equivalent of zap.DebugLevel
	DEBUG LogLevel = iota

	// INFO is the equivalent of zap.InfoLevel
	INFO

	// WARN is the equivalent of zap.WarnLevel
	WARN

	// ERROR is the equivalent of zap.ErrorLevel
	ERROR
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

// Log is a prepared wrapper to harmonize the logging entrypoint.
func (l *Logger) Log(logLevel LogLevel, msg string, keysAndValues ...interface{}) {
	var method func(msg string, keysAndValues ...interface{})
	switch logLevel {
	case DEBUG:
		method = l.Debugw
	case INFO:
		method = l.Infow
	case WARN:
		method = l.Warnw
	case ERROR:
		method = l.Errorw
	}

	method(msg, keysAndValues...)
}

// AuditLog is a prepared wrapper to harmonize the audit logging format.
func (l *Logger) AuditLog(logLevel LogLevel, phase string, msg string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, "log-type", "audit", "phase", phase)
	l.Log(logLevel, msg, keysAndValues...)
}
