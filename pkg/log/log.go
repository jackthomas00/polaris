package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	// Set default log level from environment or default to info
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Configure output with timestamp
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	logger = zerolog.New(output).With().Timestamp().Logger()
}

// SetLevel sets the global log level
func SetLevel(level string) {
	if logLevel, err := zerolog.ParseLevel(level); err == nil {
		zerolog.SetGlobalLevel(logLevel)
	}
}

// Debug logs a debug message
func Debug(msg string) {
	logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	logger.Debug().Msg(fmt.Sprintf(format, v...))
}

// Info logs an info message
func Info(msg string) {
	logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	logger.Info().Msg(fmt.Sprintf(format, v...))
}

// Warn logs a warning message
func Warn(msg string) {
	logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func Warnf(format string, v ...interface{}) {
	logger.Warn().Msg(fmt.Sprintf(format, v...))
}

// Error logs an error message
func Error(msg string) {
	logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	logger.Error().Msg(fmt.Sprintf(format, v...))
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, v ...interface{}) {
	logger.Fatal().Msg(fmt.Sprintf(format, v...))
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) zerolog.Logger {
	return logger.With().Logger()
}

// WithField returns a logger with a field
func WithField(key string, value interface{}) zerolog.Logger {
	return logger.With().Interface(key, value).Logger()
}

// WithFields returns a logger with multiple fields
func WithFields(fields map[string]interface{}) zerolog.Logger {
	event := logger.With()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	return event.Logger()
}

// WithError returns a logger with an error field
func WithError(err error) zerolog.Logger {
	return logger.With().Err(err).Logger()
}
