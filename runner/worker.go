package runner

import (
	"net/http"

	"github.com/logrusorgru/aurora"
	"io"
	"strings"
)

type resultStatus string

const (
	ResultHTTPError     resultStatus = "http error"
	ResultResponseError              = "response error"
	ResultVulnerable                 = "vulnerable"
	ResultNotVulnerable              = "not vulnerable"
)

type Result struct {
	resStatus resultStatus
	status    aurora.Value
	entry     Fingerprint
}

func (c *Config) checkSubdomain(subdomain string) Result {
	if !isValidUrl(subdomain) {
		if c.HTTPS {
			subdomain = "https://" + subdomain
		} else {
			subdomain = "http://" + subdomain
		}
	}

	req, err := http.NewRequest("GET", subdomain, nil)
	if err != nil {
		return Result{ResultHTTPError, aurora.Red("HTTP ERROR"), Fingerprint{}}
	}

	// Set User-Agent if configured
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	} else {
		req.Header.Set("User-Agent", "Subzy/1.1.0 (Subdomain Takeover Scanner)")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// Check for DNS resolution errors (NXDOMAIN)
		if strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "server misbehaving") {
			return c.matchNXDomain()
		}
		return Result{ResultHTTPError, aurora.Red("HTTP ERROR"), Fingerprint{}}
	}

	statusCode := resp.StatusCode

	// Limit response body to 1MB to prevent memory exhaustion
	limitedBody := io.LimitReader(resp.Body, 1024*1024)
	body, err := io.ReadAll(limitedBody)
	resp.Body.Close()
	if err != nil {
		return Result{ResultResponseError, aurora.Red("RESPONSE ERROR"), Fingerprint{}}
	}

	return c.matchResponse(string(body), statusCode)
}

func (c *Config) matchResponse(body string, statusCode int) Result {
	for _, fingerprint := range c.fingerprints {
		// Skip NXDOMAIN-only fingerprints
		if fingerprint.NXDomain && fingerprint.Fingerprint == "NXDOMAIN" {
			continue
		}

		matched := false

		// Check HTTP status code match if specified
		if fingerprint.HTTPStatus != nil {
			if statusCode == *fingerprint.HTTPStatus {
				matched = true
			}
		}

		// Check fingerprint string in body if specified
		if fingerprint.Fingerprint != "" && fingerprint.Fingerprint != "NXDOMAIN" {
			if strings.Contains(body, fingerprint.Fingerprint) {
				matched = true
			}
		}

		// If matched and marked as vulnerable, report it
		if matched && fingerprint.Vulnerable {
			return Result{ResultVulnerable, aurora.Green("VULNERABLE"), fingerprint}
		}
	}

	return Result{ResultNotVulnerable, aurora.Red("NOT VULNERABLE"), Fingerprint{}}
}

func (c *Config) matchNXDomain() Result {
	for _, fingerprint := range c.fingerprints {
		// Check for NXDOMAIN fingerprints
		if fingerprint.NXDomain && fingerprint.Fingerprint == "NXDOMAIN" {
			if fingerprint.Vulnerable {
				return Result{ResultVulnerable, aurora.Green("VULNERABLE"), fingerprint}
			}
		}
	}

	return Result{ResultHTTPError, aurora.Red("HTTP ERROR"), Fingerprint{}}
}
