package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReadSubdomains(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdomains.txt")

	// Create test file with subdomains
	content := "subdomain1.example.com\nsubdomain2.example.com\nsubdomain3.example.com\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	subdomains, err := readSubdomains(testFile)
	if err != nil {
		t.Fatalf("readSubdomains() error = %v", err)
	}

	expectedCount := 3
	if len(subdomains) != expectedCount {
		t.Errorf("Expected %d subdomains, got %d", expectedCount, len(subdomains))
	}

	expected := []string{
		"subdomain1.example.com",
		"subdomain2.example.com",
		"subdomain3.example.com",
	}

	for i, subdomain := range subdomains {
		if subdomain != expected[i] {
			t.Errorf("Subdomain[%d] = %q, want %q", i, subdomain, expected[i])
		}
	}
}

func TestReadSubdomainsEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	// Create empty file
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	subdomains, err := readSubdomains(testFile)
	if err != nil {
		t.Fatalf("readSubdomains() error = %v", err)
	}

	if len(subdomains) != 0 {
		t.Errorf("Expected 0 subdomains from empty file, got %d", len(subdomains))
	}
}

func TestReadSubdomainsWithEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdomains.txt")

	// Create test file with empty lines
	content := "subdomain1.example.com\n\nsubdomain2.example.com\n\n\nsubdomain3.example.com\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	subdomains, err := readSubdomains(testFile)
	if err != nil {
		t.Fatalf("readSubdomains() error = %v", err)
	}

	// Empty lines are included as empty strings
	expectedCount := 6
	if len(subdomains) != expectedCount {
		t.Errorf("Expected %d entries (including empty lines), got %d", expectedCount, len(subdomains))
	}
}

func TestReadSubdomainsFileNotFound(t *testing.T) {
	_, err := readSubdomains("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Expected error when file doesn't exist")
	}
}

func TestReadSubdomainsWithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdomains.txt")

	// Create test file with whitespace
	content := "  subdomain1.example.com  \nsubdomain2.example.com\n\tsubdomain3.example.com\t\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	subdomains, err := readSubdomains(testFile)
	if err != nil {
		t.Fatalf("readSubdomains() error = %v", err)
	}

	// Scanner preserves whitespace
	if subdomains[0] != "  subdomain1.example.com  " {
		t.Errorf("Expected whitespace to be preserved, got %q", subdomains[0])
	}
}

func TestReadSubdomainsLargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create a file with many subdomains
	var content string
	count := 1000
	for i := 0; i < count; i++ {
		if i > 0 {
			content += "\n"
		}
		content += fmt.Sprintf("subdomain%d.example.com", i)
	}

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	subdomains, err := readSubdomains(testFile)
	if err != nil {
		t.Fatalf("readSubdomains() error = %v", err)
	}

	if len(subdomains) != count {
		t.Errorf("Expected %d subdomains, got %d", count, len(subdomains))
	}
}
