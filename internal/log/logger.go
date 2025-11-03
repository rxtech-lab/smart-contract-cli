package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	Logger *zerolog.Logger
	file   *lumberjack.Logger
}

type Config struct {
	// LogFilePath is the file path where logs will be written
	// Default: "./logs/app.log"
	LogFilePath string
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated
	// Default: 10 MB
	MaxSize int
	// MaxBackups is the maximum number of old log files to retain
	// Default: 5
	MaxBackups int
	// MaxAge is the maximum number of days to retain old log files
	// Default: 30 days
	MaxAge int
	// Compress determines if the rotated log files should be compressed using gzip
	// Default: true
	Compress bool
	// ConsoleOutput determines if logs should also be written to console
	// Default: true
	ConsoleOutput bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		LogFilePath:   "./logs/app.log",
		MaxSize:       10,
		MaxBackups:    5,
		MaxAge:        30,
		Compress:      true,
		ConsoleOutput: true,
	}
}

// NewLogger creates a new logger that writes to console only (backward compatible).
func NewLogger() *Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &Logger{Logger: &logger}
}

// NewLoggerWithConfig creates a new logger with custom configuration.
// It supports writing to both file and console with log rotation.
func NewLoggerWithConfig(config Config) (*Logger, error) {
	// Ensure log directory exists
	logDir := filepath.Dir(config.LogFilePath)
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create lumberjack logger for file rotation
	fileLogger := &lumberjack.Logger{
		Filename:   config.LogFilePath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
		LocalTime:  true,
	}

	// Create multi-writer based on configuration
	var writers []io.Writer
	writers = append(writers, fileLogger)

	if config.ConsoleOutput {
		// Use console writer for pretty output to console
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		writers = append(writers, consoleWriter)
	}

	multi := io.MultiWriter(writers...)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	return &Logger{
		Logger: &logger,
		file:   fileLogger,
	}, nil
}

// NewFileLogger creates a logger that writes only to file (no console output).
func NewFileLogger(logFilePath string) (*Logger, error) {
	config := DefaultConfig()
	config.LogFilePath = logFilePath
	config.ConsoleOutput = false
	return NewLoggerWithConfig(config)
}

// Close closes the log file and flushes any pending writes.
func (l *Logger) Close() error {
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
	}
	return nil
}

// Rotate causes the log file to be rotated.
func (l *Logger) Rotate() error {
	if l.file != nil {
		if err := l.file.Rotate(); err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}
	}
	return nil
}

func (l *Logger) Info(message string, args ...any) {
	l.Logger.Info().Msgf(message, args...)
}

func (l *Logger) Error(message string, args ...any) {
	l.Logger.Error().Msgf(message, args...)
}

func (l *Logger) Debug(message string, args ...any) {
	l.Logger.Debug().Msgf(message, args...)
}

func (l *Logger) Warn(message string, args ...any) {
	l.Logger.Warn().Msgf(message, args...)
}

func (l *Logger) Fatal(message string, args ...any) {
	l.Logger.Fatal().Msgf(message, args...)
}
