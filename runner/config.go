package runner

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Config holds the configuration for subdomain scanning
type Config struct {
	HTTPS       bool
	VerifySSL   bool
	HideFails   bool
	OnlyVuln    bool
	Concurrency int
	Timeout     int
	Targets     string
	Target      string
	Output      string
	UserAgent   string

	// Logging configuration
	LogLevel    string
	LogFormat   string
	GraylogHost string
	GraylogApp  string
	LogToFile   bool
	LogFilePath string

	client       *http.Client
	fingerprints []Fingerprint
	logger       zerolog.Logger
}

func (s *Config) initHTTPClient() {
	// Note: InsecureSkipVerify defaults to true (!s.VerifySSL) for security scanning
	// because many targets have self-signed or expired certificates.
	// Use --verify_ssl flag to enable strict TLS verification.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !s.VerifySSL,
			MinVersion:         tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: s.Concurrency,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	timeout := time.Duration(s.Timeout) * time.Second
	client := &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	s.client = client
}

func (c *Config) loadFingerprints() error {
	fingerprints, err := Fingerprints()
	if err != nil {
		return err
	}
	c.fingerprints = fingerprints
	return nil
}
