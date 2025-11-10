package runner

import "testing"

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		setting  bool
		expected string
	}{
		{
			name:     "enabled returns Yes",
			setting:  true,
			expected: "[ Yes ]",
		},
		{
			name:     "disabled returns No",
			setting:  false,
			expected: "[ No ]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEnabled(tt.setting)
			if result != tt.expected {
				t.Errorf("isEnabled(%v) = %q, want %q", tt.setting, result, tt.expected)
			}
		})
	}
}

func TestIsValidUrl(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid http URL",
			url:      "http://example.com",
			expected: true,
		},
		{
			name:     "valid https URL",
			url:      "https://example.com",
			expected: true,
		},
		{
			name:     "valid URL with path",
			url:      "https://example.com/path/to/resource",
			expected: true,
		},
		{
			name:     "valid URL with port",
			url:      "http://example.com:8080",
			expected: true,
		},
		{
			name:     "invalid URL without scheme",
			url:      "example.com",
			expected: false,
		},
		{
			name:     "invalid URL with spaces",
			url:      "http://exam ple.com",
			expected: false,
		},
		{
			name:     "empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "subdomain without scheme",
			url:      "subdomain.example.com",
			expected: false,
		},
		{
			name:     "valid URL with subdomain",
			url:      "https://subdomain.example.com",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUrl(tt.url)
			if result != tt.expected {
				t.Errorf("isValidUrl(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}
