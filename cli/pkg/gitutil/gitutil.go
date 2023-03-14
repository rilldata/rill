package gitutil

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	exec "golang.org/x/sys/execabs"
)

func CloneRepo(url string) (string, error) {
	endpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	repoName := fileutil.Stem(endpoint.Path)
	cmd := exec.Command("git", "clone", url)
	cmd.Stderr = os.Stderr
	_, err = cmd.Output()
	if err != nil {
		return "", err
	}

	return repoName, nil
}

type Remote struct {
	Name string
	URL  string
}

func ExtractRemotes(projectPath string) ([]Remote, error) {
	repo, err := git.PlainOpen(projectPath)
	if err != nil {
		return nil, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}

	res := make([]Remote, len(remotes))
	for idx, remote := range remotes {
		if len(remote.Config().URLs) == 0 {
			return nil, fmt.Errorf("no URL found for git remote %q", remote.Config().Name)
		}

		res[idx] = Remote{
			Name: remote.Config().Name,
			// The first URL in the slice is the URL Git fetches from (main one).
			// We'll make things easy for ourselves and only consider that.
			URL: remote.Config().URLs[0],
		}
	}

	return res, nil
}
