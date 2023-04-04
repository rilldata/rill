package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-version"
)

// ReleaseInfo stores information about a release
type ReleaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

func CheckVersion(ctx context.Context, currentVersion string) (string, error) {
	addr := "https://api.github.com/repos/rilldata/rill-developer/releases/latest"
	info, err := latestVersion(ctx, addr)
	if err != nil {
		return "", err
	}

	v1, err := version.NewVersion(currentVersion)
	if err != nil {
		return "", err
	}

	v2, err := version.NewVersion(info.Version)
	if err != nil {
		return "", err
	}

	if v1.LessThan(v2) {
		return fmt.Sprintf("Latest version (%s) is greater than the current build version (%s)\n",
			info.Version, currentVersion), nil
	}
	return "", nil
}

func latestVersion(ctx context.Context, addr string) (*ReleaseInfo, error) {
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
		return nil, fmt.Errorf("error fetching latest release: %v", string(out))
	}

	var info *ReleaseInfo
	err = json.Unmarshal(out, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
