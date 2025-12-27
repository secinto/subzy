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
			Engine:        "TestEngine1",
			Status:        "vulnerable",
			Fingerprint:   "error-message-1",
			Discussion:    "https://example.com/discussion1",
			Documentation: "https://example.com/docs1",
			FalsePositive: []string{"false-positive-1"},
		},
		{
			Engine:        "TestEngine2",
			Status:        "vulnerable",
			Fingerprint:   "error-message-2",
			Discussion:    "https://example.com/discussion2",
			Documentation: "https://example.com/docs2",
			FalsePositive: []string{},
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
	if fingerprints[0].Engine != "TestEngine1" {
		t.Errorf("Expected Engine 'TestEngine1', got %q", fingerprints[0].Engine)
	}

	if fingerprints[0].Fingerprint != "error-message-1" {
		t.Errorf("Expected Fingerprint 'error-message-1', got %q", fingerprints[0].Fingerprint)
	}

	if len(fingerprints[0].FalsePositive) != 1 {
		t.Errorf("Expected 1 false positive, got %d", len(fingerprints[0].FalsePositive))
	}

	// Verify second fingerprint
	if fingerprints[1].Engine != "TestEngine2" {
		t.Errorf("Expected Engine 'TestEngine2', got %q", fingerprints[1].Engine)
	}

	if len(fingerprints[1].FalsePositive) != 0 {
		t.Errorf("Expected 0 false positives, got %d", len(fingerprints[1].FalsePositive))
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
		"engine": "TestEngine",
		"status": "vulnerable",
		"fingerprint": "test-fp",
		"discussion": "test-discussion",
		"documentation": "test-docs",
		"false_positive": ["fp1", "fp2"]
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
		{"Engine", fp.Engine, "TestEngine"},
		{"Status", fp.Status, "vulnerable"},
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

	if len(fp.FalsePositive) != 2 {
		t.Errorf("Expected 2 false positives, got %d", len(fp.FalsePositive))
	}
}
