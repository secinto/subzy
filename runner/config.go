package runner

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	HTTPS        bool
	VerifySSL    bool
	Emoji        bool
	HideFails    bool
	OnlyVuln     bool
	Concurrency  int
	Timeout      int
	Targets      string
	Target       string
	Output       string
	UserAgent    string
	RateLimit    int

	// Logging configuration
	LogLevel     string
	LogFormat    string
	GraylogHost  string
	GraylogApp   string
	LogToFile    bool
	LogFilePath  string

	client       *http.Client
	fingerprints []Fingerprint
	logger       zerolog.Logger
}

func (s *Config) initHTTPClient() {
	// Optimize connection pooling for concurrent requests
	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: !s.VerifySSL},
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
