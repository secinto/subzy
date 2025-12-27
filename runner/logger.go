package runner

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level       string // debug, info, warn, error
	Format      string // json, console
	GraylogHost string // e.g., "graylog.example.com:12201"
	GraylogApp  string // Application name
	EnableFile  bool   // Log to file
	FilePath    string // Log file path
}

// InitLogger initializes and configures the logger with multiple outputs
func InitLogger(cfg LogConfig) (zerolog.Logger, error) {
	var writers []io.Writer

	// Console output (development/debugging)
	if cfg.Format == "console" || cfg.Format == "" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
		writers = append(writers, consoleWriter)
	} else if cfg.Format == "json" {
		// JSON output to stdout
		writers = append(writers, os.Stdout)
	}

	// Graylog GELF output (production monitoring)
	if cfg.GraylogHost != "" {
		gelfWriter, err := gelf.NewUDPWriter(cfg.GraylogHost)
		if err != nil {
			return zerolog.Logger{}, fmt.Errorf("failed to create Graylog writer: %w", err)
		}

		// Set Graylog facility (application name)
		if cfg.GraylogApp != "" {
			gelfWriter.Facility = cfg.GraylogApp
		} else {
			gelfWriter.Facility = "subzy"
		}

		// Wrap GELF writer to convert zerolog format to GELF
		writers = append(writers, &gelfLogWriter{writer: gelfWriter})
	}

	// File output with secure permissions (owner read/write only)
	if cfg.EnableFile && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return zerolog.Logger{}, fmt.Errorf("failed to open log file: %w", err)
		}
		writers = append(writers, file)
	}

	// If no writers configured, default to console
	if len(writers) == 0 {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		})
	}

	// Create multi-writer
	multi := zerolog.MultiLevelWriter(writers...)

	// Create logger with timestamp and app name
	logger := zerolog.New(multi).With().
		Timestamp().
		Str("app", cfg.GraylogApp).
		Logger()

	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger = logger.Level(level)

	return logger, nil
}

// gelfLogWriter wraps a GELF writer to make it compatible with zerolog
type gelfLogWriter struct {
	writer *gelf.UDPWriter
}

// Write implements io.Writer for GELF
func (w *gelfLogWriter) Write(p []byte) (n int, err error) {
	// Create GELF message
	msg := &gelf.Message{
		Version:  "1.1",
		Host:     getHostname(),
		Short:    string(p),
		Full:     string(p),
		TimeUnix: float64(time.Now().Unix()),
		Level:    6, // Info level by default
		Extra: map[string]interface{}{
			"_facility": "subzy",
		},
	}

	// Send to Graylog
	if err := w.writer.WriteMessage(msg); err != nil {
		return 0, err
	}

	return len(p), nil
}

// getHostname returns the hostname or "unknown" if unavailable
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// DefaultLogger creates a default console logger
func DefaultLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}).With().
		Timestamp().
		Str("app", "subzy").
		Logger()
}
