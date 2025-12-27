package runner

import (
	"encoding/json"
	"fmt"
	"os"
)

type Fingerprint struct {
	Engine        string   `json:"engine"`
	Status        string   `json:"status"`
	Fingerprint   string   `json:"fingerprint"`
	Discussion    string   `json:"discussion"`
	Documentation string   `json:"documentation"`
	FalsePositive []string `json:"False_Positive"`
}

func Fingerprints() ([]Fingerprint, error) {

	var fingerprints []Fingerprint

	fingerPrintsPath, err := GetFingerprintPath()
	if err != nil {
		return nil, fmt.Errorf("Fingerprints: %v", err)
	}
	file, err := os.ReadFile(fingerPrintsPath)
	if err != nil {
		return nil, fmt.Errorf("Fingerprints: %v", err)
	}

	err = json.Unmarshal(file, &fingerprints)
	if err != nil {
		return nil, fmt.Errorf("Fingerprints: %v", err)
	}

	return fingerprints, nil
}
