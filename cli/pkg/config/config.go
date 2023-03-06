package config

import "fmt"

type Config struct {
	Version             Version
	AdminURL            string
	OverridenAdminToken string
	DefaultAdminToken   string
	DefaultOrg          string
}

func (c Config) IsDev() bool {
	return c.Version.IsDev()
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

func (c *Config) IsAuthenticated() bool {
	return (c.GetAdminToken() != "" && c.AdminURL != "")
}

func (c *Config) GetAdminToken() string {
	if c.OverridenAdminToken != "" {
		return c.OverridenAdminToken
	}
	return c.DefaultAdminToken
}
