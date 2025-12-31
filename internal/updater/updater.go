package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GitHub repository URL
const GitHubRepoURL = "https://github.com/vpoluyaktov/SmartCalc"
const gitHubAPIURL = "https://api.github.com/repos/vpoluyaktov/SmartCalc/releases/latest"

// ReleaseInfo contains information about a GitHub release
type ReleaseInfo struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Body        string `json:"body"`
}

// CheckForUpdates checks if there's a newer version available on GitHub
func CheckForUpdates(currentVersion string) (*ReleaseInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", gitHubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "SmartCalc-App")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Compare versions (strip 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVer := strings.TrimPrefix(currentVersion, "v")

	if latestVersion != currentVer && latestVersion > currentVer {
		return &release, nil
	}

	return nil, nil // No update available
}
