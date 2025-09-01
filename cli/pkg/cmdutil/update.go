package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
)

const (
	versionCheckURI = "https://api.github.com/repos/rilldata/rill/releases/latest"
	versionCheckTTL = 24 * time.Hour
)

func (h *Helper) CheckVersion(ctx context.Context) error {
	// Check if build from source
	if h.Version.Number == "" {
		return nil
	}

	// Add a timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	latestVersion, err := h.LatestVersion(ctx)
	if err != nil {
		return err
	}

	v1, err := version.NewVersion(h.Version.Number)
	if err != nil {
		return err
	}

	v2, err := version.NewVersion(latestVersion)
	if err != nil {
		// Set version as empty if any parse errors
		_ = h.DotRill.SetVersion("")
		return err
	}

	if v1.LessThan(v2) {
		fmt.Printf("%s %s → %s\n\n",
			color.YellowString("A new version of rill is available (run `rill upgrade`):"),
			color.CyanString(h.Version.Number),
			color.CyanString(latestVersion))
		return nil
	}

	return nil
}

// LatestVersion returns the latest available version of Rill (cached for up to 24 hours).
func (h *Helper) LatestVersion(ctx context.Context) (string, error) {
	cachedVersion, err := h.DotRill.GetVersion()
	if err != nil {
		return "", err
	}

	cachedVersionUpdatedAt, err := h.DotRill.GetVersionUpdatedAt()
	if err != nil {
		return "", err
	}

	if cachedVersion != "" && cachedVersionUpdatedAt != "" {
		updatedAt, err := time.Parse(time.RFC3339, cachedVersionUpdatedAt)
		if err != nil {
			// Set versionTs as empty if any parse errors
			_ = h.DotRill.SetVersionUpdatedAt("")
			return "", err
		}

		if time.Since(updatedAt) < versionCheckTTL {
			return cachedVersion, nil
		}
	}

	// Check with latest release on github if cached version is not available
	info, err := fetchLatestVersion(ctx)
	if err != nil {
		return "", err
	}

	err = h.DotRill.SetVersionUpdatedAt(time.Now().Format(time.RFC3339))
	if err != nil {
		return "", err
	}

	err = h.DotRill.SetVersion(info.Version)
	if err != nil {
		return "", err
	}

	return info.Version, nil
}

// githubReleaseInfo represents information about a release.
type githubReleaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

// fetchLatestVersion fetches the latest version of Rill from Github releases.
func fetchLatestVersion(ctx context.Context) (*githubReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionCheckURI, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	getToken := func() string {
		if t := os.Getenv("GH_TOKEN"); t != "" {
			return t
		}
		return os.Getenv("GITHUB_TOKEN")
	}

	if token := getToken(); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	client := &http.Client{Timeout: time.Second * 15}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !success {
		return nil, fmt.Errorf("error fetching latest release: %s", string(out))
	}

	var info *githubReleaseInfo
	err = json.Unmarshal(out, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
