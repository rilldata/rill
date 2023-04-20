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
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"gopkg.in/yaml.v2"
)

const addr = "https://api.github.com/repos/rilldata/rill-developer/releases/latest"

type UpdateInfo struct {
	Update      bool
	Message     string
	ReleaseInfo *ReleaseInfo
}

// ReleaseInfo stores information about a release
type ReleaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

func CheckVersion(ctx context.Context, currentVersion string) error {
	// Check if build from source
	if currentVersion == "" {
		return nil
	}

	updateInfo, err := checkVersion(ctx, currentVersion)
	if err != nil {
		return err
	}

	if updateInfo.Update {
		fmt.Printf("\n%s %s → %s\n\n",
			color.YellowString("A new version of rill is available:"),
			color.CyanString(currentVersion),
			color.CyanString(updateInfo.ReleaseInfo.Version))
	}

	return nil
}

func checkVersion(ctx context.Context, currentVersion string) (*UpdateInfo, error) {
	stateEntry, _ := getVersionInfo()

	if stateEntry != nil && time.Since(stateEntry.UpdateAt).Hours() < 24 {
		v1, err := version.NewVersion(currentVersion)
		if err != nil {
			return nil, err
		}

		v2, err := version.NewVersion(stateEntry.LatestRelease.Version)
		if err != nil {
			return nil, err
		}

		if v1.LessThan(v2) {
			return &UpdateInfo{
				Update: true,
				Message: fmt.Sprintf("Latest version (%s) is greater than the current build version (%s)\n",
					stateEntry.LatestRelease.Version, currentVersion),
				ReleaseInfo: &stateEntry.LatestRelease,
			}, nil
		}

		return &UpdateInfo{
			Update:  false,
			Message: "Skip checking the latest version",
		}, nil
	}

	// Check with latest release on github
	info, err := LatestVersion(ctx)
	if err != nil {
		return nil, err
	}

	err = setVersionInfo(time.Now(), *info)
	if err != nil {
		return nil, err
	}

	v1, err := version.NewVersion(currentVersion)
	if err != nil {
		return nil, err
	}

	v2, err := version.NewVersion(info.Version)
	if err != nil {
		return nil, err
	}

	if v1.LessThan(v2) {
		return &UpdateInfo{
			Update: true,
			Message: fmt.Sprintf("Latest version (%s) is greater than the current build version (%s)\n",
				info.Version, currentVersion),
			ReleaseInfo: info,
		}, nil
	}

	return &UpdateInfo{
		Update: false,
		Message: fmt.Sprintf("Latest version (%s) is less than or equal to current build version (%s)",
			info.Version, currentVersion),
		ReleaseInfo: info,
	}, nil
}

// This will fetch the latest version available for rill on github releases
func LatestVersion(ctx context.Context) (*ReleaseInfo, error) {
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

	var info *ReleaseInfo
	err = json.Unmarshal(out, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

type versionInfo struct {
	UpdateAt      time.Time   `yaml:"update_at"`
	LatestRelease ReleaseInfo `yaml:"latest_release"`
}

func getVersionInfo() (*versionInfo, error) {
	content, err := dotrill.GetVersionInfo()
	if err != nil {
		return nil, err
	}

	var verionInfo versionInfo
	err = yaml.Unmarshal([]byte(content), &verionInfo)
	if err != nil {
		return nil, err
	}

	return &verionInfo, nil
}

func setVersionInfo(t time.Time, r ReleaseInfo) error {
	data := versionInfo{
		UpdateAt:      t,
		LatestRelease: r,
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	err = dotrill.SetVersionInfo(string(content))
	if err != nil {
		return err
	}

	return nil
}
