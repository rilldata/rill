package gitutil

import (
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// SplitGithubRemote takes a GitHub HTTPS remote URL and extracts the Github account and repository name.
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
