package gitutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrGitRemoteNotFound = errors.New("no git remotes found")

// Remote represents a Git remote with its name and URL.
type Remote struct {
	Name string
	URL  string
}

// Github returns a normalized HTTPS Github URL ending in .git for the remote.
func (r Remote) Github() (string, error) {
	if r.URL == "" {
		return "", fmt.Errorf("remote %q has no URL", r.Name)
	}
	return NormalizeGithubRemote(r.URL)
}

// ExtractGitRemote extracts the first Git remote from the Git repository at projectPath.
// If remoteName is provided, it will return the remote with that name.
// If detectDotGit is true, it will look for a .git directory in parent directories.
func ExtractGitRemote(projectPath, remoteName string, detectDotGit bool) (Remote, error) {
	remotes, err := ExtractRemotes(projectPath, detectDotGit)
	if err != nil {
		return Remote{}, err
	}
	if remoteName != "" {
		for _, remote := range remotes {
			if remote.Name == remoteName {
				return remote, nil
			}
		}
		return Remote{}, ErrGitRemoteNotFound
	}
	if len(remotes) == 0 {
		return Remote{}, ErrGitRemoteNotFound
	}
	return remotes[0], nil
}

// ExtractRemotes extracts all Git remotes from the Git repository at projectPath.
// If detectDotGit is true, it will look for a .git directory in parent directories.
func ExtractRemotes(projectPath string, detectDotGit bool) ([]Remote, error) {
	if !detectDotGit {
		// require a .git entry at exactly this path: a directory in a regular
		// checkout, or a file in a linked worktree.
		if _, err := os.Stat(filepath.Join(projectPath, ".git")); err != nil {
			if os.IsNotExist(err) {
				return nil, ErrNotAGitRepository
			}
			return nil, err
		}
	}

	out, err := Run(context.Background(), projectPath, "remote")
	if err != nil {
		if strings.Contains(err.Error(), "not a git repository") {
			return nil, ErrNotAGitRepository
		}
		return nil, err
	}

	names := strings.Fields(out)
	res := make([]Remote, 0, len(names))
	for _, name := range names {
		u, err := Run(context.Background(), projectPath, "remote", "get-url", name)
		if err != nil {
			return nil, fmt.Errorf("failed to get URL for git remote %q: %w", name, err)
		}
		if u == "" {
			return nil, fmt.Errorf("no URL found for git remote %q", name)
		}
		res = append(res, Remote{Name: name, URL: u})
	}

	return res, nil
}

// SetRemote sets the remote named config.RemoteName() for the repository at path to config.Remote.
// It is a no-op if the remote already has the wanted URL, or if it is not a Rill-managed remote
// (a user's own remote must never be overwritten).
func SetRemote(path string, config *Config) error {
	if config.Remote == "" {
		return nil
	}
	current, err := Run(context.Background(), path, "remote", "get-url", config.RemoteName())
	if err != nil {
		if !strings.Contains(err.Error(), "No such remote") {
			return fmt.Errorf("failed to get remote: %w", err)
		}
		// the remote does not exist yet: create it
		_, err = Run(context.Background(), path, "remote", "add", config.RemoteName(), config.Remote)
		return err
	}
	if current == config.Remote || !config.ManagedRepo {
		// remote already exists with the same URL, no need to update it
		// remote other than managed git exists, can't overwrite user's remote
		return nil
	}
	// the managed remote exists with a different URL: update it
	_, err = Run(context.Background(), path, "remote", "set-url", config.RemoteName(), config.Remote)
	return err
}

// RemoveRemote removes the named remote from the repository at path.
// It is a no-op if the remote does not exist.
func RemoveRemote(path, remoteName string) error {
	_, err := Run(context.Background(), path, "remote", "remove", remoteName)
	if err != nil && !strings.Contains(err.Error(), "No such remote") {
		return err
	}
	return nil
}
