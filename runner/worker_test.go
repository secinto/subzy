package runner

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMatchResponse(t *testing.T) {
	config := &Config{
		fingerprints: []Fingerprint{
			{
				Service:       "TestService",
				Fingerprint:   "unique-error-message",
				Vulnerable:    true,
				Discussion:    "https://example.com/discussion",
				Documentation: "https://example.com/docs",
				CICDPass:      false,
				CName:         []string{},
				NXDomain:      false,
				HTTPStatus:    nil,
				Status:        "Vulnerable",
			},
			{
				Service:       "FalsePositiveService",
				Fingerprint:   "error-occurred",
				Vulnerable:    false,
				Discussion:    "https://example.com/discussion2",
				Documentation: "https://example.com/docs2",
				CICDPass:      false,
				CName:         []string{},
				NXDomain:      false,
				HTTPStatus:    nil,
				Status:        "Not vulnerable",
			},
		},
	}

	tests := []struct {
		name           string
		body           string
		statusCode     int
		expectedStatus resultStatus
	}{
		{
			name:           "vulnerable when fingerprint matches",
			body:           "This page contains unique-error-message",
			statusCode:     200,
			expectedStatus: ResultVulnerable,
		},
		{
			name:           "not vulnerable when no fingerprint matches",
			body:           "This page has no matching fingerprint",
			statusCode:     200,
			expectedStatus: ResultNotVulnerable,
		},
		{
			name:           "not vulnerable when fingerprint matches but not marked vulnerable",
			body:           "error-occurred",
			statusCode:     200,
			expectedStatus: ResultNotVulnerable,
		},
		{
			name:           "empty body returns not vulnerable",
			body:           "",
			statusCode:     200,
			expectedStatus: ResultNotVulnerable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.matchResponse(tt.body, tt.statusCode)
			if result.resStatus != tt.expectedStatus {
				t.Errorf("matchResponse() status = %v, want %v", result.resStatus, tt.expectedStatus)
			}
		})
	}
}

func TestCheckSubdomain(t *testing.T) {
	// Create a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check User-Agent header
		ua := r.Header.Get("User-Agent")
		if ua == "" {
			t.Error("User-Agent header not set")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test-fingerprint"))
	}))
	defer ts.Close()

	config := &Config{
		HTTPS:     false,
		VerifySSL: false,
		Timeout:   10,
		fingerprints: []Fingerprint{
			{
				Service:       "TestEngine",
				Fingerprint:   "test-fingerprint",
				Vulnerable:    true,
				CICDPass:      false,
				CName:         []string{},
				NXDomain:      false,
				HTTPStatus:    nil,
				Status:        "Vulnerable",
				Discussion:    "",
				Documentation: "",
			},
		},
	}
	config.initHTTPClient()

	tests := []struct {
		name           string
		subdomain      string
		expectedStatus resultStatus
	}{
		{
			name:           "valid URL with server response",
			subdomain:      ts.URL,
			expectedStatus: ResultVulnerable,
		},
		{
			name:           "URL without scheme adds http",
			subdomain:      ts.URL[7:], // Remove http://
			expectedStatus: ResultVulnerable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.checkSubdomain(tt.subdomain)
			if result.resStatus != tt.expectedStatus {
				t.Errorf("checkSubdomain() status = %v, want %v", result.resStatus, tt.expectedStatus)
			}
		})
	}
}

func TestCheckSubdomainHTTPError(t *testing.T) {
	config := &Config{
		HTTPS:     false,
		VerifySSL: false,
		Timeout:   1,
	}
	config.initHTTPClient()

	result := config.checkSubdomain("http://nonexistent-domain-12345.invalid")
	if result.resStatus != ResultHTTPError {
		t.Errorf("checkSubdomain() with invalid domain should return HTTP error, got %v", result.resStatus)
	}
}

func TestCheckSubdomainWithHTTPSFlag(t *testing.T) {
	config := &Config{
		HTTPS:     true,
		VerifySSL: false,
		Timeout:   10,
		fingerprints: []Fingerprint{
			{
				Service:       "TestEngine",
				Fingerprint:   "not-present",
				Vulnerable:    true,
				CICDPass:      false,
				CName:         []string{},
				NXDomain:      false,
				HTTPStatus:    nil,
				Status:        "Vulnerable",
				Discussion:    "",
				Documentation: "",
			},
		},
	}
	config.initHTTPClient()

	// Test that HTTPS flag is respected
	// This will fail with connection error, but we're just checking the URL construction
	subdomain := "example.com"
	result := config.checkSubdomain(subdomain)

	// Should get an HTTP error trying to connect
	if result.resStatus != ResultHTTPError {
		t.Errorf("Expected HTTP error for unreachable HTTPS domain, got %v", result.resStatus)
	}
}

func TestResponseBodyLimit(t *testing.T) {
	// Create a test server that returns a very large response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write 2MB of data (more than the 1MB limit)
		data := make([]byte, 2*1024*1024)
		for i := range data {
			data[i] = 'A'
		}
		w.Write(data)
	}))
	defer ts.Close()

	config := &Config{
		HTTPS:        false,
		VerifySSL:    false,
		Timeout:      10,
		fingerprints: []Fingerprint{},
	}
	config.initHTTPClient()

	// Should not cause memory issues due to 1MB limit
	result := config.checkSubdomain(ts.URL)

	// Should complete without error
	if result.resStatus == ResultResponseError {
		t.Error("Should handle large responses without error")
	}
}
