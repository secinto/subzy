package runner

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

var (
	fingerprintPath = "https://raw.githubusercontent.com/LukaSikic/subzy/master/runner/fingerprints.json"
	subzyDir        = "subzy"
	// GetFingerprintPath is a variable to allow overriding in tests
	GetFingerprintPath = func() (string, error) {
		home, err := homedir.Dir()
		if err != nil {
			return "", fmt.Errorf("GetFingerprintPath: %v", err)
		}
		dirPath := filepath.Join(home, subzyDir)
		if _, err := os.Stat(dirPath); errors.Is(err, fs.ErrNotExist) {
			if err := os.Mkdir(dirPath, 0700); err != nil {
				return "", err
			}
		}
		return path.Join(dirPath, "fingerprints.json"), nil
	}
)

func downloadFingerprints(fingerprintsPath string) error {
	out, err := os.OpenFile(fingerprintsPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("downloadFingerprints: %v", err)
	}
	defer out.Close()

	resp, err := http.Get(fingerprintPath)
	if err != nil {
		return fmt.Errorf("downloadFingerprints: %v", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("downloadFingerprints: %v", err)
	}

	return nil
}

func CheckFingerprints() error {
	filePath, err := GetFingerprintPath()
	if err != nil {
		return fmt.Errorf("CheckFingerprints: %v", err)
	}
	return downloadFingerprints(filePath)
}
