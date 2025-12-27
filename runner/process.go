package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/logrusorgru/aurora"
)

func Process(config *Config) error {
	// Initialize logger
	logger, err := InitLogger(LogConfig{
		Level:       config.LogLevel,
		Format:      config.LogFormat,
		GraylogHost: config.GraylogHost,
		GraylogApp:  config.GraylogApp,
		EnableFile:  config.LogToFile,
		FilePath:    config.LogFilePath,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	config.logger = logger

	config.initHTTPClient()
	if err := config.loadFingerprints(); err != nil {
		return fmt.Errorf("Process: %v", err)
	}

	subdomains, err := getSubdomains(config)
	if err != nil {
		return fmt.Errorf("Process: failed to get subdomains: %w", err)
	}

	// Log scan configuration
	logger.Info().
		Int("target_count", len(subdomains)).
		Int("fingerprint_count", len(config.fingerprints)).
		Str("output_file", config.Output).
		Bool("only_vulnerable", config.OnlyVuln).
		Bool("https_default", config.HTTPS).
		Int("concurrency", config.Concurrency).
		Bool("verify_ssl", config.VerifySSL).
		Int("timeout_seconds", config.Timeout).
		Bool("hide_fails", config.HideFails).
		Msg("Starting subdomain takeover scan")

	// Keep console output for user feedback (when not in JSON mode)
	if config.LogFormat != "json" {
		fmt.Println("[ * ]", "Loaded", len(subdomains), "targets")
		fmt.Println("[ * ]", "Loaded", len(config.fingerprints), "fingerprints")
		if config.Output != "" {
			fmt.Printf("[ * ] Output filename: %s\n", config.Output)
			fmt.Println(isEnabled(config.OnlyVuln), "Save only vulnerable subdomains")
		}

		fmt.Println(isEnabled(config.HTTPS), "HTTPS by default (--https)")
		fmt.Println("[", config.Concurrency, "]", "Concurrent requests (--concurrency)")
		fmt.Println(isEnabled(config.VerifySSL), "Check target only if SSL is valid (--verify_ssl)")
		fmt.Println("[", config.Timeout, "]", "HTTP request timeout (in seconds) (--timeout)")
		fmt.Println(isEnabled(config.HideFails), "Show only potentially vulnerable subdomains (--hide_fails)")
	}

	subdomainCh := make(chan string, config.Concurrency*2)
	resCh := make(chan *subdomainResult, config.Concurrency)

	var wg sync.WaitGroup
	wg.Add(config.Concurrency)

	// Results collector - single goroutine, no mutex needed
	var results []*subdomainResult
	var resultsWg sync.WaitGroup
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for r := range resCh {
			if config.Output != "" {
				if config.OnlyVuln && r.Status != string(ResultVulnerable) {
					continue
				}
				results = append(results, r)
			}
		}
	}()

	for i := 0; i < config.Concurrency; i++ {
		go processor(subdomainCh, resCh, config, &wg)
	}

	go func() {
		for _, subdomain := range subdomains {
			subdomainCh <- subdomain
		}
		close(subdomainCh)
	}()

	wg.Wait()
	close(resCh)
	resultsWg.Wait()

	if config.Output != "" {
		f, err := os.OpenFile(config.Output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")

		if err := enc.Encode(results); err != nil {
			return err
		}

		logger.Info().
			Str("output_file", config.Output).
			Int("result_count", len(results)).
			Msg("Saved scan results to file")

		if config.LogFormat != "json" {
			fmt.Printf("[ * ] Saved output to %q\n", config.Output)
		}
	}

	logger.Info().Msg("Scan completed")
	return nil
}

func processor(subdomainCh chan string, resCh chan *subdomainResult, c *Config, wg *sync.WaitGroup) {
	defer wg.Done()
	for subdomain := range subdomainCh {
		result := c.checkSubdomain(subdomain)
		resCh <- &subdomainResult{
			Subdomain:     subdomain,
			Status:        string(result.resStatus),
			Engine:        result.entry.Engine,
			Documentation: result.entry.Documentation,
			Discussion:    result.entry.Discussion,
		}

		// Use resStatus for comparison instead of aurora.Value
		if result.resStatus == ResultVulnerable {
			// Log vulnerability with structured data
			c.logger.Error().
				Str("subdomain", subdomain).
				Str("status", "vulnerable").
				Str("engine", result.entry.Engine).
				Str("documentation", result.entry.Documentation).
				Str("discussion", result.entry.Discussion).
				Msg("Vulnerable subdomain detected")

			// Console output for user
			if c.LogFormat != "json" {
				fmt.Print("-----------------\r\n")
				fmt.Println("[ ", result.status, " ]", " - ", subdomain, " [ ", result.entry.Engine, " ] ")
				fmt.Println("[ ", aurora.Blue("DISCUSSION"), " ]", " - ", result.entry.Discussion)
				fmt.Println("[ ", aurora.Blue("DOCUMENTATION"), " ]", " - ", result.entry.Documentation)
				fmt.Print("-----------------\r\n")
			}
		} else {
			// Log check result
			if result.resStatus == ResultHTTPError {
				c.logger.Warn().
					Str("subdomain", subdomain).
					Str("status", string(result.resStatus)).
					Msg("HTTP error checking subdomain")
			} else {
				c.logger.Debug().
					Str("subdomain", subdomain).
					Str("status", string(result.resStatus)).
					Msg("Subdomain check completed")
			}

			// Console output
			if !c.HideFails && c.LogFormat != "json" {
				fmt.Println("[ ", result.status, " ]", " - ", subdomain)
			}
		}
	}
}

// getSubdomains returns list of subdomains from config
func getSubdomains(c *Config) ([]string, error) {
	if c.Target == "" {
		return readSubdomains(c.Targets)
	}
	return strings.Split(c.Target, ","), nil
}
