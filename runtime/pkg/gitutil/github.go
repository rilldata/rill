package gitutil

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// scpLikeRegexp matches scp-style git remotes such as "git@github.com:org/repo.git", optionally
// with a numeric port segment ("host:22:path"). Paths starting with a backslash are excluded so
// Windows paths like `C:\dir` parse as local paths instead.
var scpLikeRegexp = regexp.MustCompile(`^(?:([^@]+)@)?([^:\s]+):(?:(\d{1,5}):)?([^\\].*)$`)

// SplitGithubRemote splits a GitHub remote URL into a Github account and repository name.
func SplitGithubRemote(remote string) (account, repo string, ok bool) {
	host, remotePath, ok := splitRemote(remote)
	if !ok || !strings.EqualFold(host, "github.com") {
		return "", "", false
	}

	account, repo = path.Split(remotePath)
	account = strings.Trim(account, "/")
	repo = strings.TrimSuffix(repo, ".git")
	if account == "" || repo == "" || strings.Contains(account, "/") {
		return "", "", false
	}

	return account, repo, true
}

// NormalizeGithubRemote validates and converts a Git remote to a normalized HTTPS Github URL ending in .git.
// Any credentials embedded in the remote are stripped.
func NormalizeGithubRemote(remote string) (string, error) {
	account, repo, ok := SplitGithubRemote(remote)
	if !ok {
		return "", fmt.Errorf("remote %q is not a valid github.com remote", remote)
	}
	// The .git suffix is not always required (e.g. Github has redirects if its missing), so we add it for consistency.
	return fmt.Sprintf("https://github.com/%s/%s.git", account, repo), nil
}

// splitRemote splits a git remote into its host and path components, dropping credentials, ports,
// and the scheme. It accepts scheme URLs ("https://", "ssh://", "git://"), scp-style remotes
// ("git@github.com:org/repo.git"), and local file paths (for which host is empty).
func splitRemote(remote string) (host, remotePath string, ok bool) {
	if !strings.Contains(remote, "://") {
		if m := scpLikeRegexp.FindStringSubmatch(remote); m != nil {
			return m[2], m[4], true
		}
		// a local file path
		return "", remote, true
	}
	u, err := url.Parse(remote)
	if err != nil {
		return "", "", false
	}
	return u.Hostname(), u.Path, true
}
