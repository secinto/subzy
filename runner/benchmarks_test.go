package runner

import (
	"testing"
)

func BenchmarkMatchResponse(b *testing.B) {
	config := &Config{
		fingerprints: []Fingerprint{
			{
				Engine:        "Service1",
				Fingerprint:   "error-message-1",
				FalsePositive: []string{},
			},
			{
				Engine:        "Service2",
				Fingerprint:   "error-message-2",
				FalsePositive: []string{"false-positive"},
			},
			{
				Engine:        "Service3",
				Fingerprint:   "error-message-3",
				FalsePositive: []string{},
			},
		},
	}

	body := "This is a test response body that contains error-message-2 and some other text to make it realistic"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.matchResponse(body)
	}
}

func BenchmarkMatchResponseNoMatch(b *testing.B) {
	config := &Config{
		fingerprints: []Fingerprint{
			{
				Engine:        "Service1",
				Fingerprint:   "error-message-1",
				FalsePositive: []string{},
			},
			{
				Engine:        "Service2",
				Fingerprint:   "error-message-2",
				FalsePositive: []string{},
			},
		},
	}

	body := "This is a test response body with no matching fingerprints"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.matchResponse(body)
	}
}

func BenchmarkMatchResponseWithFalsePositive(b *testing.B) {
	config := &Config{
		fingerprints: []Fingerprint{
			{
				Engine:        "Service1",
				Fingerprint:   "error-occurred",
				FalsePositive: []string{"but-its-ok", "no-worries", "all-good"},
			},
		},
	}

	body := "error-occurred but-its-ok so everything is fine"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.matchResponse(body)
	}
}

func BenchmarkMatchResponseManyFingerprints(b *testing.B) {
	// Simulate realistic scenario with 44 fingerprints
	fingerprints := make([]Fingerprint, 44)
	for i := 0; i < 44; i++ {
		fingerprints[i] = Fingerprint{
			Engine:        "Service",
			Fingerprint:   "unique-error",
			FalsePositive: []string{},
		}
	}

	config := &Config{
		fingerprints: fingerprints,
	}

	body := "This is a test response body with no matching fingerprints but lots of text to process"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.matchResponse(body)
	}
}

func BenchmarkIsValidUrl(b *testing.B) {
	urls := []string{
		"http://example.com",
		"https://subdomain.example.com/path",
		"example.com",
		"http://example.com:8080/path?query=value",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, url := range urls {
			isValidUrl(url)
		}
	}
}

func BenchmarkInitHTTPClient(b *testing.B) {
	config := &Config{
		Concurrency: 10,
		Timeout:     10,
		VerifySSL:   false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.initHTTPClient()
	}
}

func BenchmarkIsEnabled(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isEnabled(true)
		isEnabled(false)
	}
}
