package updater

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/afero"
)

const checkInterval = 3 * time.Hour

type cacheEntry struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

// CacheFilePath returns the path to the update check cache file
func CacheFilePath() string {
	return filepath.Join(xdg.CacheHome, "ticker", "update-check.json")
}

// Check returns the latest version if one is available and newer than currentVersion,
// otherwise returns an empty string. Results are cached for 3 hours.
// Skips the check for dev builds (currentVersion == "v0.0.0").
func Check(currentVersion, releasesURL, cacheFilePath string, fs afero.Fs) string {
	if currentVersion == "v0.0.0" {
		return ""
	}

	if entry, err := readCache(cacheFilePath, fs); err == nil {
		if time.Since(entry.CheckedAt) < checkInterval {
			if entry.LatestVersion != currentVersion {
				return entry.LatestVersion
			}

			return ""
		}
	}

	latest, err := fetchLatest(releasesURL)
	if err != nil {
		return ""
	}

	_ = writeCache(cacheFilePath, latest, fs)

	if latest != currentVersion {
		return latest
	}

	return ""
}

func readCache(path string, fs afero.Fs) (cacheEntry, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return cacheEntry{}, err
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}, err
	}

	return entry, nil
}

func writeCache(path, latestVersion string, fs afero.Fs) error {
	data, err := json.Marshal(cacheEntry{
		CheckedAt:     time.Now(),
		LatestVersion: latestVersion,
	})
	if err != nil {
		return err
	}
	if err := fs.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return afero.WriteFile(fs, path, data, 0644)
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
