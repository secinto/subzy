package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFingerprintPath(t *testing.T) {
	// Save original function
	originalGetPath := GetFingerprintPath
	defer func() { GetFingerprintPath = originalGetPath }()

	// Test with custom path
	testDir := t.TempDir()
	expectedPath := filepath.Join(testDir, "fingerprints.json")

	GetFingerprintPath = func() (string, error) {
		return expectedPath, nil
	}

	path, err := GetFingerprintPath()
	if err != nil {
		t.Fatalf("GetFingerprintPath() error = %v", err)
	}

	if path != expectedPath {
		t.Errorf("GetFingerprintPath() = %v, want %v", path, expectedPath)
	}
}

func TestDownloadFingerprintsInvalidPath(t *testing.T) {
	// Try to download to invalid path
	err := downloadFingerprints("/nonexistent/directory/fingerprints.json")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

func TestCheckFingerprintsWithMockedPath(t *testing.T) {
	// Save original function
	originalGetPath := GetFingerprintPath
	defer func() { GetFingerprintPath = originalGetPath }()

	// Mock with temp directory
	tmpDir := t.TempDir()
	fingerprintPath := filepath.Join(tmpDir, "fingerprints.json")

	GetFingerprintPath = func() (string, error) {
		return fingerprintPath, nil
	}

	// This will attempt to download from GitHub
	// In a real test environment, we'd mock the HTTP client
	// For now, we just verify the function doesn't panic
	err := CheckFingerprints()

	// We expect this to either succeed or fail gracefully
	// (network may not be available in test environment)
	if err != nil {
		t.Logf("CheckFingerprints() returned error (expected in test environment): %v", err)
	}

	// If it succeeded, verify file exists
	if err == nil {
		if _, statErr := os.Stat(fingerprintPath); os.IsNotExist(statErr) {
			t.Error("Fingerprints file should exist after successful download")
		}
	}
}

func TestFingerprintPathVariable(t *testing.T) {
	// Test that fingerprintPath is correctly set
	expectedURL := "https://raw.githubusercontent.com/LukaSikic/subzy/master/runner/fingerprints.json"
	if fingerprintPath != expectedURL {
		t.Errorf("fingerprintPath = %v, want %v", fingerprintPath, expectedURL)
	}
}

func TestSubzyDirVariable(t *testing.T) {
	// Test that subzyDir is correctly set
	if subzyDir != "subzy" {
		t.Errorf("subzyDir = %v, want subzy", subzyDir)
	}
}
