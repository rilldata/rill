// Package dotrill implements setting and getting key-value pairs in YAML files in ~/.rill.
package dotrill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

// Constants for YAML files
const (
	ConfigFilename      = "config.yaml"      // For user-facing config
	CredentialsFilename = "credentials.yaml" // For access tokens
	StateFilename       = "state.yaml"       // For CLI state
)

// Constants for YAML keys
const (
	DefaultOrgConfigKey                             = "org"
	BackupDefaultOrgConfigKey                       = "backup_org"
	DefaultAdminURLConfigKey                        = "api_url"
	AnalyticsEnabledConfigKey                       = "analytics_enabled"
	AccessTokenCredentialsKey                       = "token"
	InstallIDStateKey                               = "install_id"
	RepresentingUserCredentialsKey                  = "representing_user"
	RepresentingUserAccessTokenExpiryCredentialsKey = "representing_user_token_expiry"
	BackupTokenCredentialsKey                       = "backup_token"
	LatestVersionStateKey                           = "latest_version"
	LatestVersionCheckedAtStateKey                  = "latest_version_checked_at"
	UserIDStateKey                                  = "user_id"
	UserCheckHashStateKey                           = "user_check_hash"
)

// DotRill encapsulates access to .rill.
type DotRill struct {
	homeDir string
}

// New creates a new Dotrill instance.
// If homeDir is empty, it creates `.rill` in the user's home directory.
func New(homeDir string) DotRill {
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	return DotRill{homeDir: homeDir}
}

// GetAll loads all values from ~/.rill/{filename}.
// It assumes filename identifies a YAML file.
func (d DotRill) GetAll(filename string) (map[string]string, error) {
	filename, err := d.ResolveFilename(filename, false)
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
func (d DotRill) Get(filename, key string) (string, error) {
	conf, err := d.GetAll(filename)
	if err != nil {
		return "", err
	}

	return conf[key], nil
}

// Set sets a single value in ~/.rill/{filename}.
// It assumes filename identifies a YAML file.
func (d DotRill) Set(filename, key, value string) error {
	if key == "" {
		return fmt.Errorf("cannot set empty key")
	}

	conf, err := d.GetAll(filename)
	if err != nil {
		return err
	}
	conf[key] = value

	filename, err = d.ResolveFilename(filename, true)
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
func (d DotRill) GetDefaultOrg() (string, error) {
	return d.Get(ConfigFilename, DefaultOrgConfigKey)
}

// SetDefaultOrg saves the default org
func (d DotRill) SetDefaultOrg(orgName string) error {
	return d.Set(ConfigFilename, DefaultOrgConfigKey, orgName)
}

// GetBackupDefaultOrg loads the backedup default org
func (d DotRill) GetBackupDefaultOrg() (string, error) {
	return d.Get(ConfigFilename, BackupDefaultOrgConfigKey)
}

// SetBackupDefaultOrg saves the backedup default org
func (d DotRill) SetBackupDefaultOrg(orgName string) error {
	return d.Set(ConfigFilename, BackupDefaultOrgConfigKey, orgName)
}

// SetDefaultAdminURL loads the default admin URL (if set)
func (d DotRill) SetDefaultAdminURL(url string) error {
	return d.Set(ConfigFilename, DefaultAdminURLConfigKey, url)
}

// GetDefaultAdminURL loads the default admin URL (if set)
func (d DotRill) GetDefaultAdminURL() (string, error) {
	return d.Get(ConfigFilename, DefaultAdminURLConfigKey)
}

// GetToken loads the current auth token
func (d DotRill) GetAccessToken() (string, error) {
	return d.Get(CredentialsFilename, AccessTokenCredentialsKey)
}

// SetToken saves an auth token
func (d DotRill) SetAccessToken(token string) error {
	return d.Set(CredentialsFilename, AccessTokenCredentialsKey, token)
}

// GetBackupToken loads the original auth token
func (d DotRill) GetBackupToken() (string, error) {
	return d.Get(CredentialsFilename, BackupTokenCredentialsKey)
}

// SetBackupToken saves original auth token
func (d DotRill) SetBackupToken(token string) error {
	return d.Set(CredentialsFilename, BackupTokenCredentialsKey, token)
}

// GetRepresentingUser loads the current representing user email
func (d DotRill) GetRepresentingUser() (string, error) {
	return d.Get(CredentialsFilename, RepresentingUserCredentialsKey)
}

// GetRepresentingUserAccessTokenExpiry loads the current auth token expiry
func (d DotRill) GetRepresentingUserAccessTokenExpiry() (time.Time, error) {
	expiryStr, err := d.Get(CredentialsFilename, RepresentingUserAccessTokenExpiryCredentialsKey)
	if err != nil {
		return time.Time{}, err
	}
	if expiryStr == "" {
		return time.Time{}, nil
	}
	expiry, err := time.Parse(time.RFC3339Nano, expiryStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse token expiry: %w", err)
	}
	return expiry, nil
}

// SetRepresentingUserAccessTokenExpiry saves an auth token expiry
func (d DotRill) SetRepresentingUserAccessTokenExpiry(expiry time.Time) error {
	var expiryStr string
	if !expiry.IsZero() {
		expiryStr = expiry.Format(time.RFC3339Nano)
	}
	return d.Set(CredentialsFilename, RepresentingUserAccessTokenExpiryCredentialsKey, expiryStr)
}

// SetRepresentingUser saves representing user email
func (d DotRill) SetRepresentingUser(email string) error {
	return d.Set(CredentialsFilename, RepresentingUserCredentialsKey, email)
}

func (d DotRill) SetVersion(version string) error {
	return d.Set(StateFilename, LatestVersionStateKey, version)
}

func (d DotRill) GetVersion() (string, error) {
	return d.Get(StateFilename, LatestVersionStateKey)
}

func (d DotRill) SetVersionUpdatedAt(updatedAt string) error {
	return d.Set(StateFilename, LatestVersionCheckedAtStateKey, updatedAt)
}

func (d DotRill) GetVersionUpdatedAt() (string, error) {
	return d.Get(StateFilename, LatestVersionCheckedAtStateKey)
}

// SetEnvToken backup the token for given env
func (d DotRill) SetEnvToken(env, token string) error {
	key := fmt.Sprintf("tokens.%s", env)
	return d.Set(CredentialsFilename, key, token)
}

// GetEnvToken loads the token for given env
func (d DotRill) GetEnvToken(env string) (string, error) {
	key := fmt.Sprintf("tokens.%s", env)
	return d.Get(CredentialsFilename, key)
}

// GetCurrentUserID gets the current user ID
func (d DotRill) GetUserID() (string, error) {
	return d.Get(StateFilename, UserIDStateKey)
}

// SetCurrentUserID saves the current user ID
func (d DotRill) SetUserID(userID string) error {
	return d.Set(StateFilename, UserIDStateKey, userID)
}

// GetUserCheckHash gets the hash used to determine whether to re-fetch the user ID.
func (d DotRill) GetUserCheckHash() (string, error) {
	return d.Get(StateFilename, UserCheckHashStateKey)
}

// SetUserCheckHash sets the hash used to determine whether to re-fetch the user ID.
func (d DotRill) SetUserCheckHash(hash string) error {
	return d.Set(StateFilename, UserCheckHashStateKey, hash)
}

// AnalyticsInfo returns analytics info.
// It loads a persistent install ID from ~/.rill/state.yaml (setting one if not found).
// It gets analytics enabled/disabled info from ~/.rill/config.yaml (key "analytics_enabled").
// It automatically migrates from the pre-v0.23 analytics config. See migrateOldAnalyticsConfig for details.
func (d DotRill) AnalyticsInfo() (installID string, enabled bool, err error) {
	// Migrate from earlier analytics tracking, if necessary
	err = d.migrateOldAnalyticsConfig()
	if err != nil {
		fmt.Printf("state migration in ~/.rill did not succeed: %s\n", err.Error())
	}

	// Get installID
	installID, err = d.Get(StateFilename, InstallIDStateKey)
	if err != nil {
		return "", false, err
	}

	// Trim space just to be safe
	installID = strings.TrimSpace(installID)

	// If installID was not found (or had been cleared), persist a new ID
	if installID == "" {
		installID = uuid.New().String()
		err := d.Set(StateFilename, InstallIDStateKey, installID)
		if err != nil {
			return "", false, err
		}
	}

	// Check if analytics is enabled
	enabledStr, err := d.Get(ConfigFilename, AnalyticsEnabledConfigKey)
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
func (d DotRill) migrateOldAnalyticsConfig() error {
	filename, err := d.ResolveFilename("local.json", false)
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
		err := d.Set(StateFilename, InstallIDStateKey, conf.InstallID)
		if err != nil {
			return err
		}
	}

	// Set analytics_enabled if applicable
	if conf.AnalyticsEnabled != nil && !*conf.AnalyticsEnabled {
		err := d.Set(ConfigFilename, AnalyticsEnabledConfigKey, "false")
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

// ResolveFilename resolves a file name to a full path to ~/.rill.
// If mkdir is true, it will create the .rill directory if it doesn't exist.
func (d DotRill) ResolveFilename(name string, mkdir bool) (string, error) {
	if d.homeDir == "" {
		return "", fmt.Errorf("home directory not found")
	}

	dotrill := filepath.Join(d.homeDir, ".rill")
	if mkdir {
		err := os.MkdirAll(dotrill, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	filename := filepath.Join(dotrill, name)
	return filename, nil
}
