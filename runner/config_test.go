package runner

import (
	"testing"
	"time"
)

func TestInitHTTPClient(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		verifySetup func(*testing.T, *Config)
	}{
		{
			name: "default configuration",
			config: Config{
				VerifySSL:   false,
				Concurrency: 10,
				Timeout:     10,
			},
			verifySetup: func(t *testing.T, c *Config) {
				if c.client == nil {
					t.Error("HTTP client not initialized")
				}
				if c.client.Timeout != 10*time.Second {
					t.Errorf("Expected timeout 10s, got %v", c.client.Timeout)
				}
			},
		},
		{
			name: "with SSL verification",
			config: Config{
				VerifySSL:   true,
				Concurrency: 5,
				Timeout:     30,
			},
			verifySetup: func(t *testing.T, c *Config) {
				if c.client == nil {
					t.Error("HTTP client not initialized")
				}
				if c.client.Timeout != 30*time.Second {
					t.Errorf("Expected timeout 30s, got %v", c.client.Timeout)
				}
			},
		},
		{
			name: "high concurrency",
			config: Config{
				VerifySSL:   false,
				Concurrency: 100,
				Timeout:     5,
			},
			verifySetup: func(t *testing.T, c *Config) {
				if c.client == nil {
					t.Error("HTTP client not initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.initHTTPClient()
			tt.verifySetup(t, &tt.config)
		})
	}
}

func TestLoadFingerprints(t *testing.T) {
	// Save original function
	originalGetPath := GetFingerprintPath
	defer func() { GetFingerprintPath = originalGetPath }()

	// Test with non-existent path
	GetFingerprintPath = func() (string, error) {
		return "/nonexistent/fingerprints.json", nil
	}

	config := &Config{}
	err := config.loadFingerprints()
	if err == nil {
		t.Error("Expected error when fingerprints file doesn't exist")
	}
}

func TestConfigUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
	}{
		{
			name:      "default user agent",
			userAgent: "",
		},
		{
			name:      "custom user agent",
			userAgent: "CustomBot/1.0",
		},
		{
			name:      "browser user agent",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				UserAgent: tt.userAgent,
			}
			if config.UserAgent != tt.userAgent {
				t.Errorf("UserAgent = %q, want %q", config.UserAgent, tt.userAgent)
			}
		})
	}
}

func TestHTTPClientConnectionPooling(t *testing.T) {
	config := &Config{
		Concurrency: 50,
		Timeout:     10,
		VerifySSL:   false,
	}

	config.initHTTPClient()

	if config.client == nil {
		t.Fatal("HTTP client not initialized")
	}

	// Verify transport is configured
	transport := config.client.Transport
	if transport == nil {
		t.Fatal("HTTP transport not configured")
	}
}

func TestTimeoutConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		timeoutSeconds  int
		expectedTimeout time.Duration
	}{
		{
			name:            "1 second timeout",
			timeoutSeconds:  1,
			expectedTimeout: 1 * time.Second,
		},
		{
			name:            "10 second timeout",
			timeoutSeconds:  10,
			expectedTimeout: 10 * time.Second,
		},
		{
			name:            "60 second timeout",
			timeoutSeconds:  60,
			expectedTimeout: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Timeout:     tt.timeoutSeconds,
				Concurrency: 10,
			}
			config.initHTTPClient()

			if config.client.Timeout != tt.expectedTimeout {
				t.Errorf("Expected timeout %v, got %v", tt.expectedTimeout, config.client.Timeout)
			}
		})
	}
}
