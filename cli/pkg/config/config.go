package config

import "fmt"

type Config struct {
	Version            Version
	AdminURL           string
	AdminTokenDefault  string
	AdminTokenOverride string
	DefaultOrg         string
}

func (c Config) IsDev() bool {
	return c.Version.IsDev()
}

func (c Config) AdminToken() string {
	if c.AdminTokenOverride != "" {
		return c.AdminTokenOverride
	}
	return c.AdminTokenDefault
}

type Version struct {
	Number    string
	Commit    string
	Timestamp string
}

func (v Version) String() string {
	if v.Number == "" {
		return "unknown (built from source)"
	}
	return fmt.Sprintf("%s (build commit: %s date: %s)", v.Number, v.Commit, v.Timestamp)
}

func (v Version) IsDev() bool {
	return v.Number == ""
}
