package runner

import (
	"bytes"
	"io"
	"net/http"

	"github.com/logrusorgru/aurora"
)

// Version is the application version used in User-Agent
const Version = "1.1.0"

type resultStatus string

const (
	ResultHTTPError     resultStatus = "http error"
	ResultResponseError resultStatus = "response error"
	ResultVulnerable    resultStatus = "vulnerable"
	ResultNotVulnerable resultStatus = "not vulnerable"
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
		req.Header.Set("User-Agent", "Subzy/"+Version+" (Subdomain Takeover Scanner)")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{ResultHTTPError, aurora.Red("HTTP ERROR"), Fingerprint{}}
	}
	defer resp.Body.Close()

	// Limit response body to 1MB to prevent memory exhaustion
	limitedBody := io.LimitReader(resp.Body, 1024*1024)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return Result{ResultResponseError, aurora.Red("RESPONSE ERROR"), Fingerprint{}}
	}

	return c.matchResponse(body)
}

// matchResponse checks response body against fingerprints
// Using []byte instead of string to avoid memory allocation
func (c *Config) matchResponse(body []byte) Result {
	for _, fingerprint := range c.fingerprints {
		fingerprintBytes := []byte(fingerprint.Fingerprint)
		if bytes.Contains(body, fingerprintBytes) {
			for _, falsePositiveString := range fingerprint.FalsePositive {
				if len(falsePositiveString) > 0 {
					if bytes.Contains(body, []byte(falsePositiveString)) {
						return Result{ResultNotVulnerable, aurora.Red("NOT VULNERABLE"), Fingerprint{}}
					}
				}
			}
			return Result{ResultVulnerable, aurora.Green("VULNERABLE"), fingerprint}
		}
	}

	return Result{ResultNotVulnerable, aurora.Red("NOT VULNERABLE"), Fingerprint{}}
}
