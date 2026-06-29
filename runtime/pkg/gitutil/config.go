package gitutil

import (
	"fmt"
	"net/url"
	"time"
)

// Config describes a git remote and the (usually ephemeral) credentials to access it.
type Config struct {
	Remote            string
	Username          string
	Password          string
	PasswordExpiresAt time.Time
	DefaultBranch     string
	Subpath           string
	ManagedRepo       bool
}

func (g *Config) IsExpired() bool {
	return g.Password != "" && g.PasswordExpiresAt.Before(time.Now())
}

// FullyQualifiedRemote returns the remote URL with the credentials embedded in it.
// The result may be passed to git commands as an argument but must never be persisted to .git/config.
func (g *Config) FullyQualifiedRemote() (string, error) {
	if g.Remote == "" {
		return "", fmt.Errorf("remote is not set")
	}
	u, err := url.Parse(g.Remote)
	if err != nil {
		return "", err
	}
	if g.Username != "" {
		if g.Password != "" {
			u.User = url.UserPassword(g.Username, g.Password)
		} else {
			u.User = url.User(g.Username)
		}
	}
	return u.String(), nil
}

func (g *Config) RemoteName() string {
	if g.ManagedRepo {
		return "__rill_remote"
	}
	return "origin"
}

// Signature identifies the author of a git commit.
type Signature struct {
	Name  string
	Email string
}
