package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSubdomains(t *testing.T) {
	// Create a temporary file with test data
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "targets.txt")

	content := []byte("example.com\ntest.com\nfoo.bar.com")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		config   *Config
		expected []string
		wantErr  bool
	}{
		{
			name: "single target from CLI",
			config: &Config{
				Target: "example.com",
			},
			expected: []string{"example.com"},
			wantErr:  false,
		},
		{
			name: "multiple targets from CLI",
			config: &Config{
				Target: "example.com,test.com,foo.bar.com",
			},
			expected: []string{"example.com", "test.com", "foo.bar.com"},
			wantErr:  false,
		},
		{
			name: "targets from file",
			config: &Config{
				Targets: testFile,
			},
			expected: []string{"example.com", "test.com", "foo.bar.com"},
			wantErr:  false,
		},
		{
			name: "non-existent file",
			config: &Config{
				Targets: "/nonexistent/file.txt",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getSubdomains(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSubdomains() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("getSubdomains() got %d results, want %d", len(result), len(tt.expected))
				}
				for i, v := range result {
					if v != tt.expected[i] {
						t.Errorf("getSubdomains()[%d] = %v, want %v", i, v, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestSubdomainResult(t *testing.T) {
	result := &subdomainResult{
		Subdomain:     "test.example.com",
		Status:        "vulnerable",
		Engine:        "TestEngine",
		Documentation: "https://docs.example.com",
		Discussion:    "https://github.com/example/discussion",
	}

	if result.Subdomain != "test.example.com" {
		t.Errorf("Subdomain = %v, want test.example.com", result.Subdomain)
	}
	if result.Status != "vulnerable" {
		t.Errorf("Status = %v, want vulnerable", result.Status)
	}
	if result.Engine != "TestEngine" {
		t.Errorf("Engine = %v, want TestEngine", result.Engine)
	}
}

func TestResultStatus(t *testing.T) {
	tests := []struct {
		status   resultStatus
		expected string
	}{
		{ResultHTTPError, "http error"},
		{ResultResponseError, "response error"},
		{ResultVulnerable, "vulnerable"},
		{ResultNotVulnerable, "not vulnerable"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("resultStatus = %v, want %v", string(tt.status), tt.expected)
			}
		})
	}
}
