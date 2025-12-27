package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name    string
		cfg     LogConfig
		wantErr bool
	}{
		{
			name: "default console logger",
			cfg: LogConfig{
				Level:  "info",
				Format: "console",
			},
			wantErr: false,
		},
		{
			name: "json logger",
			cfg: LogConfig{
				Level:  "info",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "debug level",
			cfg: LogConfig{
				Level:  "debug",
				Format: "console",
			},
			wantErr: false,
		},
		{
			name: "error level",
			cfg: LogConfig{
				Level:  "error",
				Format: "console",
			},
			wantErr: false,
		},
		{
			name: "invalid level defaults to info",
			cfg: LogConfig{
				Level:  "invalid",
				Format: "console",
			},
			wantErr: false,
		},
		{
			name: "empty format defaults to console",
			cfg: LogConfig{
				Level:  "info",
				Format: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := InitLogger(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Test that logger works
				logger.Info().Msg("test message")
			}
		})
	}
}

func TestInitLoggerWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	cfg := LogConfig{
		Level:      "info",
		Format:     "console",
		EnableFile: true,
		FilePath:   logFile,
	}

	logger, err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger() error = %v", err)
	}

	// Write a log message
	logger.Info().Msg("test file logging")

	// Check file was created with correct permissions
	info, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("Log file not created: %v", err)
	}

	// Check permissions (should be 0600)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Log file permissions = %v, want 0600", info.Mode().Perm())
	}
}

func TestInitLoggerWithInvalidFilePath(t *testing.T) {
	cfg := LogConfig{
		Level:      "info",
		Format:     "console",
		EnableFile: true,
		FilePath:   "/nonexistent/directory/test.log",
	}

	_, err := InitLogger(cfg)
	if err == nil {
		t.Error("Expected error for invalid file path, got nil")
	}
}

func TestDefaultLogger(t *testing.T) {
	logger := DefaultLogger()

	// Test that default logger works
	logger.Info().Msg("default logger test")
	logger.Debug().Msg("debug message")
	logger.Warn().Msg("warn message")
}

func TestGetHostname(t *testing.T) {
	hostname := getHostname()

	// Should return something (either actual hostname or "unknown")
	if hostname == "" {
		t.Error("getHostname() returned empty string")
	}
}

func TestLogConfigWithGraylogApp(t *testing.T) {
	cfg := LogConfig{
		Level:      "info",
		Format:     "console",
		GraylogApp: "test-app",
	}

	logger, err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger() error = %v", err)
	}

	// Logger should work with GraylogApp set
	logger.Info().Msg("test with graylog app name")
}
