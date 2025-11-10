package runner

import "net/url"

func isEnabled(setting bool) string {
	if setting {
		return "[ Yes ]"
	}
	return "[ No ]"
}

func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}

type subdomainResult struct {
	Subdomain     string `json:"subdomain"`
	Status        string `json:"status"`
	Engine        string `json:"engine"`
	Documentation string `json:"documentation"`
	Discussion    string `json:"discussion"`
}
