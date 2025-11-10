package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFingerprintsLoading(t *testing.T) {
	// Create a temporary directory for test
	tmpDir := t.TempDir()

	// Create a test fingerprints.json file
	testFingerprints := []Fingerprint{
		{
			Service:       "TestEngine1",
			Status:        "Vulnerable",
			Fingerprint:   "error-message-1",
			Discussion:    "https://example.com/discussion1",
			Documentation: "https://example.com/docs1",
			Vulnerable:    true,
			CICDPass:      false,
			CName:         []string{},
			NXDomain:      false,
			HTTPStatus:    nil,
		},
		{
			Service:       "TestEngine2",
			Status:        "Vulnerable",
			Fingerprint:   "error-message-2",
			Discussion:    "https://example.com/discussion2",
			Documentation: "https://example.com/docs2",
			Vulnerable:    true,
			CICDPass:      false,
			CName:         []string{},
			NXDomain:      false,
			HTTPStatus:    nil,
		},
	}

	testData, err := json.MarshalIndent(testFingerprints, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	testFile := filepath.Join(tmpDir, "fingerprints.json")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Save original function and restore after test
	originalGetPath := GetFingerprintPath
	defer func() { GetFingerprintPath = originalGetPath }()

	// Override GetFingerprintPath to return our test path
	GetFingerprintPath = func() (string, error) {
		return testFile, nil
	}

	// Test loading fingerprints
	fingerprints, err := Fingerprints()
	if err != nil {
		t.Fatalf("Fingerprints() error = %v", err)
	}

	if len(fingerprints) != 2 {
		t.Errorf("Expected 2 fingerprints, got %d", len(fingerprints))
	}

	// Verify first fingerprint
	if fingerprints[0].Service != "TestEngine1" {
		t.Errorf("Expected Service 'TestEngine1', got %q", fingerprints[0].Service)
	}

	if fingerprints[0].Fingerprint != "error-message-1" {
		t.Errorf("Expected Fingerprint 'error-message-1', got %q", fingerprints[0].Fingerprint)
	}

	if !fingerprints[0].Vulnerable {
		t.Errorf("Expected Vulnerable true, got false")
	}

	// Verify second fingerprint
	if fingerprints[1].Service != "TestEngine2" {
		t.Errorf("Expected Service 'TestEngine2', got %q", fingerprints[1].Service)
	}

	if !fingerprints[1].Vulnerable {
		t.Errorf("Expected Vulnerable true, got false")
	}
}

func TestFingerprintsFileNotFound(t *testing.T) {
	// Save original function and restore after test
	originalGetPath := GetFingerprintPath
	defer func() { GetFingerprintPath = originalGetPath }()

	// Override to return non-existent path
	GetFingerprintPath = func() (string, error) {
		return "/nonexistent/path/fingerprints.json", nil
	}

	_, err := Fingerprints()
	if err == nil {
		t.Error("Expected error when fingerprints file doesn't exist")
	}
}

func TestFingerprintsInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fingerprints.json")

	// Write invalid JSON
	if err := os.WriteFile(testFile, []byte("invalid json {{{"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Save original function and restore after test
	originalGetPath := GetFingerprintPath
	defer func() { GetFingerprintPath = originalGetPath }()

	GetFingerprintPath = func() (string, error) {
		return testFile, nil
	}

	_, err := Fingerprints()
	if err == nil {
		t.Error("Expected error when fingerprints file contains invalid JSON")
	}
}

func TestFingerprintStructTags(t *testing.T) {
	// Test that JSON tags work correctly
	jsonData := `[{
		"cicd_pass": true,
		"cname": ["example.com"],
		"service": "TestEngine",
		"status": "Vulnerable",
		"fingerprint": "test-fp",
		"discussion": "test-discussion",
		"documentation": "test-docs",
		"http_status": 404,
		"nxdomain": false,
		"vulnerable": true
	}]`

	var fingerprints []Fingerprint
	err := json.Unmarshal([]byte(jsonData), &fingerprints)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(fingerprints) != 1 {
		t.Fatalf("Expected 1 fingerprint, got %d", len(fingerprints))
	}

	fp := fingerprints[0]

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Service", fp.Service, "TestEngine"},
		{"Status", fp.Status, "Vulnerable"},
		{"Fingerprint", fp.Fingerprint, "test-fp"},
		{"Discussion", fp.Discussion, "test-discussion"},
		{"Documentation", fp.Documentation, "test-docs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}

	if !fp.Vulnerable {
		t.Errorf("Expected Vulnerable true, got false")
	}

	if !fp.CICDPass {
		t.Errorf("Expected CICDPass true, got false")
	}

	if len(fp.CName) != 1 {
		t.Errorf("Expected 1 cname, got %d", len(fp.CName))
	}

	if fp.HTTPStatus == nil || *fp.HTTPStatus != 404 {
		t.Errorf("Expected HTTPStatus 404, got %v", fp.HTTPStatus)
	}
}
