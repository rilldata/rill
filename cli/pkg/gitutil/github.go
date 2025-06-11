package gitutil

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// SplitGithubRemote splits a GitHub remote URL into a Github account and repository name.
func SplitGithubRemote(remote string) (account, repo string, ok bool) {
	ep, err := transport.NewEndpoint(remote)
	if err != nil {
		return "", "", false
	}

	if ep.Host != "github.com" {
		return "", "", false
	}

	account, repo = path.Split(ep.Path)
	account = strings.Trim(account, "/")
	repo = strings.TrimSuffix(repo, ".git")
	if account == "" || repo == "" || strings.Contains(account, "/") {
		return "", "", false
	}

	return account, repo, true
}

// NormalizeGithubRemote validates and converts a Git remote to a normalized HTTPS Github URL ending in .git.
func NormalizeGithubRemote(remote string) (string, error) {
	ep, err := transport.NewEndpoint(remote)
	if err != nil {
		return "", err
	}

	if ep.Host != "github.com" {
		return "", fmt.Errorf("remote %q is not a valid github.com remote", remote)
	}

	account, repo := path.Split(ep.Path)
	account = strings.Trim(account, "/")
	repo = strings.TrimSuffix(repo, ".git")
	if account == "" || repo == "" || strings.Contains(account, "/") {
		return "", fmt.Errorf("remote %q is not a valid github.com remote", remote)
	}

	// The .git suffix is not always required (e.g. Github has redirects if its missing), so we add it for consistency.
	if !strings.HasSuffix(ep.Path, ".git") {
		ep.Path += ".git"
	}

	githubRemote := &url.URL{
		Scheme: "https",
		Host:   ep.Host,
		Path:   ep.Path,
	}

	return githubRemote.String(), nil
}
