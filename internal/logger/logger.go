package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog.Logger for application-wide logging
type Logger struct {
	*zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Pretty     bool   // Enable pretty console output
	OutputFile string // Optional file output path
}

// New creates a new logger with the given configuration
func New(cfg Config) *Logger {
	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer = os.Stdout

	// Pretty console output (for development)
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// File output (optional)
	if cfg.OutputFile != "" {
		file, err := os.OpenFile(cfg.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			// Write to both stdout and file
			output = io.MultiWriter(output, file)
		}
	}

	// Create logger
	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{Logger: &logger}
}

// NewDefault creates a logger with default settings
func NewDefault() *Logger {
	return New(Config{
		Level:  "info",
		Pretty: true,
	})
}

// WithComponent returns a logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := l.With().Str("component", component).Logger()
	return &Logger{Logger: &newLogger}
}

// WithRequestID returns a logger with a request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	newLogger := l.With().Str("request_id", requestID).Logger()
	return &Logger{Logger: &newLogger}
}

// WithIP returns a logger with an IP address field
func (l *Logger) WithIP(ip string) *Logger {
	newLogger := l.With().Str("ip", ip).Logger()
	return &Logger{Logger: &newLogger}
}

// Global returns the global logger instance
func Global() *Logger {
	return &Logger{Logger: &log.Logger}
}
