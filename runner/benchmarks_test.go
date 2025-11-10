package runner

import (
	"testing"
)

func BenchmarkMatchResponse(b *testing.B) {
	config := &Config{
		fingerprints: []Fingerprint{
			{
				Service:     "Service1",
				Fingerprint: "error-message-1",
				Vulnerable:  true,
			},
			{
				Service:     "Service2",
				Fingerprint: "error-message-2",
				Vulnerable:  true,
			},
			{
				Service:     "Service3",
				Fingerprint: "error-message-3",
				Vulnerable:  true,
			},
		},
	}

	body := "This is a test response body that contains error-message-2 and some other text to make it realistic"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.matchResponse(body, 200)
	}
}

func BenchmarkMatchResponseNoMatch(b *testing.B) {
	config := &Config{
		fingerprints: []Fingerprint{
			{
				Service:     "Service1",
				Fingerprint: "error-message-1",
				Vulnerable:  true,
			},
			{
				Service:     "Service2",
				Fingerprint: "error-message-2",
				Vulnerable:  true,
			},
		},
	}

	body := "This is a test response body with no matching fingerprints"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.matchResponse(body, 200)
	}
}

func BenchmarkMatchResponseWithNonVulnerable(b *testing.B) {
	config := &Config{
		fingerprints: []Fingerprint{
			{
				Service:     "Service1",
				Fingerprint: "error-occurred",
				Vulnerable:  false, // Not vulnerable despite matching
			},
		},
	}

	body := "error-occurred but not actually vulnerable"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.matchResponse(body, 200)
	}
}

func BenchmarkMatchResponseManyFingerprints(b *testing.B) {
	// Simulate realistic scenario with 44 fingerprints
	fingerprints := make([]Fingerprint, 44)
	for i := 0; i < 44; i++ {
		fingerprints[i] = Fingerprint{
			Service:     "Service",
			Fingerprint: "unique-error",
			Vulnerable:  true,
		}
	}

	config := &Config{
		fingerprints: fingerprints,
	}

	body := "This is a test response body with no matching fingerprints but lots of text to process"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.matchResponse(body, 200)
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
