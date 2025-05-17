// Package dotgit implements setting and getting git config for a project in a .git file in the project directory
package dotgit

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type GitConfig struct {
	Remote         string `yaml:"remote"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	PasswordExpiry string `yaml:"password_expiry"`

	DefaultBranch string `yaml:"default_branch"`
	Subpath       string `yaml:"subpath"`
}

func (g *GitConfig) IsEmpty() bool {
	return g.Remote == ""
}

func (g *GitConfig) CredentialsExpired() bool {
	if g.PasswordExpiry == "" {
		return false
	}

	expiry, err := time.Parse(time.RFC3339, g.PasswordExpiry)
	if err != nil {
		return false
	}

	return time.Now().After(expiry)
}

func (g *GitConfig) FullyQualifiedRemote() (string, error) {
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

type DotGit struct {
	projectDir string
}

func New(projectDir string) DotGit {
	return DotGit{projectDir: projectDir}
}

func (d DotGit) LoadGitCredentials() (*GitConfig, error) {
	data, err := os.ReadFile(d.dotGitFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &GitConfig{}, nil
		}
		return nil, err
	}

	creds := &GitConfig{}
	err = yaml.Unmarshal(data, creds)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

func (d DotGit) StoreGitCredentials(g *GitConfig) error {
	data, err := yaml.Marshal(g)
	if err != nil {
		return err
	}

	path := d.dotGitFilePath()
	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (d DotGit) dotGitFilePath() string {
	// May be there is a way to not hardcode tmp ?
	return filepath.Join(d.projectDir, "tmp", ".git")
}
