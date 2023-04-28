package update

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
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
)

const (
	addr            = "https://api.github.com/repos/rilldata/rill-developer/releases/latest"
	versionCheckTTL = 24 * time.Hour
)

func CheckVersion(ctx context.Context, currentVersion string) error {
	// Check if build from source
	if currentVersion == "" {
		return nil
	}

	latestVersion, err := LatestVersion(ctx)
	if err != nil {
		return err
	}

	v1, err := version.NewVersion(currentVersion)
	if err != nil {
		return err
	}

	v2, err := version.NewVersion(latestVersion)
	if err != nil {
		return err
	}

	if v1.LessThan(v2) {
		fmt.Printf("\n%s %s → %s\n\n",
			color.YellowString("A new version of rill is available:"),
			color.CyanString(currentVersion),
			color.CyanString(latestVersion))
		return nil
	}

	return nil
}

// This will return the latest version available in cache or fetch it from github if its older than 24h
func LatestVersion(ctx context.Context) (string, error) {
	cachedVersion, err := dotrill.GetVersion()
	if err != nil {
		return "", err
	}

	if cachedVersion != "" {
		cachedVersionUpdatedAt, err := dotrill.GetVersionUpdatedAt()
		if err != nil {
			return "", err
		}

		updatedAt, err := time.Parse(cmdutil.TSFormatLayout, cachedVersionUpdatedAt)
		if err != nil {
			return "", err
		}

		if time.Since(updatedAt.UTC()).Hours() < versionCheckTTL.Hours() {
			return cachedVersion, nil
		}
	}

	// Check with latest release on github if cached version is not available
	info, err := fetchLatestVersion(ctx)
	if err != nil {
		return "", err
	}

	err = dotrill.SetVersionUpdatedAt(time.Now().UTC().Format(cmdutil.TSFormatLayout))
	if err != nil {
		return "", err
	}

	err = dotrill.SetVersion(info.Version)
	if err != nil {
		return "", err
	}

	return info.Version, nil
}

// ReleaseInfo stores information about a release
type githubReleaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

// This will fetch the latest version available for rill on github releases
func fetchLatestVersion(ctx context.Context) (*githubReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, http.NoBody)
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
