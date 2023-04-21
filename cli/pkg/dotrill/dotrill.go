// Package dotrill implements setting and getting key-value pairs in YAML files in ~/.rill.
package dotrill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/google/uuid"
)

// Constants for YAML files
const (
	ConfigFilename      = "config.yaml"      // For user-facing config
	CredentialsFilename = "credentials.yaml" // For access tokens
	StateFilename       = "state.yaml"       // For CLI state
)

// Constants for YAML keys
const (
	DefaultOrgConfigKey       = "org"
	AnalyticsEnabledConfigKey = "analytics_enabled"
	AccessTokenCredentialsKey = "token"
	InstallIDStateKey         = "install_id"
	VersionKey                = "latest_version"
	VersionUpdatedAtKey       = "latest_version_checked_at"
)

// homeDir is the user's home directory. We keep this as a global to override in unit tests.
var homeDir = ""

func init() {
	homeDir, _ = os.UserHomeDir()
}

// GetAll loads all values from ~/.rill/{filename}.
// It assumes filename identifies a YAML file.
func GetAll(filename string) (map[string]string, error) {
	filename, err := resolveFilename(filename, false)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	conf := map[string]string{}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// Get returns a single entry from ~/.rill/{filename}.
// It assumes filename identifies a YAML file.
func Get(filename, key string) (string, error) {
	conf, err := GetAll(filename)
	if err != nil {
		return "", err
	}

	return conf[key], nil
}

// Set sets a single value in ~/.rill/{filename}.
// It assumes filename identifies a YAML file.
func Set(filename, key, value string) error {
	if key == "" {
		return fmt.Errorf("cannot set empty key")
	}

	conf, err := GetAll(filename)
	if err != nil {
		return err
	}
	conf[key] = value

	filename, err = resolveFilename(filename, true)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0o644)
}

// GetDefaultOrg loads the default org
func GetDefaultOrg() (string, error) {
	return Get(ConfigFilename, DefaultOrgConfigKey)
}

// SetDefaultOrg saves the default org
func SetDefaultOrg(orgName string) error {
	return Set(ConfigFilename, DefaultOrgConfigKey, orgName)
}

// GetToken loads the current auth token
func GetAccessToken() (string, error) {
	return Get(CredentialsFilename, AccessTokenCredentialsKey)
}

// SetToken saves an auth token
func SetAccessToken(token string) error {
	return Set(CredentialsFilename, AccessTokenCredentialsKey, token)
}

func GetVersion() (string, error) {
	return Get(StateFilename, VersionKey)
}

func GetVersionUpdatedAt() (string, error) {
	return Get(StateFilename, VersionUpdatedAtKey)
}

func SetVersionUpdatedAt(updatedAt string) error {
	return Set(StateFilename, VersionUpdatedAtKey, updatedAt)
}

func SetVersion(version string) error {
	return Set(StateFilename, VersionKey, version)
}

// AnalyticsInfo returns analytics info.
// It loads a persistent install ID from ~/.rill/state.yaml (setting one if not found).
// It gets analytics enabled/disabled info from ~/.rill/config.yaml (key "analytics_enabled").
// It automatically migrates from the pre-v0.23 analytics config. See migrateOldAnalyticsConfig for details.
func AnalyticsInfo() (installID string, enabled bool, err error) {
	// Migrate from earlier analytics tracking, if necessary
	err = migrateOldAnalyticsConfig()
	if err != nil {
		fmt.Printf("state migration in ~/.rill failed: %s\n", err.Error())
	}

	// Get installID
	installID, err = Get(StateFilename, InstallIDStateKey)
	if err != nil {
		return "", false, err
	}

	// Trim space just to be safe
	installID = strings.TrimSpace(installID)

	// If installID was not found (or had been cleared), persist a new ID
	if installID == "" {
		installID = uuid.New().String()
		err := Set(StateFilename, InstallIDStateKey, installID)
		if err != nil {
			return "", false, err
		}
	}

	// Check if analytics is enabled
	enabledStr, err := Get(ConfigFilename, AnalyticsEnabledConfigKey)
	if err != nil {
		return "", false, err
	}
	// If not set, defaults to true
	enabled = enabledStr == "" || enabledStr == "1" || strings.EqualFold(enabledStr, "true")

	return installID, enabled, nil
}

// oldAnalyticsConfig represents the pre-v0.23 analytics info.
// See migrateOldAnalyticsConfig for details.
type oldAnalyticsConfig struct {
	InstallID        string `json:"installId"`
	AnalyticsEnabled *bool  `json:"analyticsEnabled"`
}

// migrateOldAnalyticsConfig migrates from the pre-v0.23 to the current analytics setup.
// It returns nil if there's nothing to migrate.
//
// Previously, analytics info was stored in ~/.rill/local.json. It included "installID" and "analyticsEnabled" fields.
// We are deprecating it to centralize user-facing config in config.yaml and to prevent confusion around the local.json file.
// It has been replaced with a config key ("analytics_enabled") and an install ID stored separately in ~/.rill/state.yaml.
func migrateOldAnalyticsConfig() error {
	filename, err := resolveFilename("local.json", false)
	if err != nil {
		return err
	}

	// Exit if file doesn't exist
	_, err = os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Parse file as JSON
	conf := &oldAnalyticsConfig{}
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, conf)
	if err != nil {
		return err
	}

	// Set install ID if applicable
	if conf.InstallID != "" {
		err := Set(StateFilename, InstallIDStateKey, conf.InstallID)
		if err != nil {
			return err
		}
	}

	// Set analytics_enabled if applicable
	if conf.AnalyticsEnabled != nil && !*conf.AnalyticsEnabled {
		err := Set(ConfigFilename, AnalyticsEnabledConfigKey, "false")
		if err != nil {
			return err
		}
	}

	// Delete the old local.json file
	err = os.Remove(filename)
	if err != nil {
		return err
	}

	return nil
}

// resolveFilename resolves a file name to a full path to ~/.rill.
// If mkdir is true, it will create the .rill directory if it doesn't exist.
func resolveFilename(name string, mkdir bool) (string, error) {
	if homeDir == "" {
		return "", fmt.Errorf("home directory not found")
	}

	dotrill := filepath.Join(homeDir, ".rill")
	if mkdir {
		err := os.MkdirAll(dotrill, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	filename := filepath.Join(dotrill, name)
	return filename, nil
}
