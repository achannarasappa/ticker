package updater

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/achannarasappa/ticker/v5/internal/cache"
	"github.com/spf13/afero"
)

const (
	checkInterval         = 3 * time.Hour
	cacheKeyLatestVersion = "latest-version"
)

// CacheFilePath returns the path to the cache file
func CacheFilePath() string {
	return cache.FilePath()
}

// Check returns the latest version if one is available and newer than currentVersion,
// otherwise returns an empty string. Results are cached for 3 hours.
// Skips the check for dev builds (currentVersion == "v0.0.0").
func Check(currentVersion, releasesURL, cacheFilePath string, fs afero.Fs) string {
	if currentVersion == "v0.0.0" {
		return ""
	}

	c := cache.New(fs, cacheFilePath, true)

	var latest string
	if c.Get(cacheKeyLatestVersion, &latest) {
		return newerVersion(latest, currentVersion)
	}

	latest, err := fetchLatest(releasesURL)
	if err != nil {
		return ""
	}

	c.Set(cacheKeyLatestVersion, latest, checkInterval)

	return newerVersion(latest, currentVersion)
}

// newerVersion returns latest when it differs from currentVersion, otherwise an
// empty string.
func newerVersion(latest, currentVersion string) string {
	if latest != currentVersion {
		return latest
	}

	return ""
}

func fetchLatest(releasesURL string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(releasesURL) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}
