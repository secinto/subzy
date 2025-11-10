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
		return Result{ResultHTTPError, aurora.Red("HTTP ERROR"), Fingerprint{}}
	}
	// Limit response body to 1MB to prevent memory exhaustion
	limitedBody := io.LimitReader(resp.Body, 1024*1024)
	body, err := io.ReadAll(limitedBody)
	resp.Body.Close()
	if err != nil {
		return Result{ResultResponseError, aurora.Red("RESPONSE ERROR"), Fingerprint{}}
	}

	return c.matchResponse(string(body))
}

func (c *Config) matchResponse(body string) Result {
	for _, fingerprint := range c.fingerprints {
		if strings.Contains(body, fingerprint.Fingerprint) {
			for _, falsePositiveString := range fingerprint.FalsePositive {
				if len(falsePositiveString) > 0 {
					if strings.Contains(body, falsePositiveString) {
						return Result{ResultNotVulnerable, aurora.Red("NOT VULNERABLE"), Fingerprint{}}
					}
				}
			}
			return Result{ResultVulnerable, aurora.Green("VULNERABLE"), fingerprint}
		}
	}

	return Result{ResultNotVulnerable, aurora.Red("NOT VULNERABLE"), Fingerprint{}}
}
