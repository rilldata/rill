package gitutil

import (
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

func SplitGithubURL(githubURL string) (account, repo string, ok bool) {
	ep, err := transport.NewEndpoint(githubURL)
	if err != nil {
		return "", "", false
	}

	if ep.Host != "github.com" {
		return "", "", false
	}

	account, repo = path.Split(ep.Path)
	account = strings.Trim(account, "/")
	if account == "" || repo == "" || strings.Contains(account, "/") {
		return "", "", false
	}

	return account, repo, true
}
