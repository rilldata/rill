package gitutil

import (
	"os"
	"strings"

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

func ExtractRemotes(projectPath string) ([]string, error) {
	repo, err := git.PlainOpen(projectPath)
	if err != nil {
		return nil, err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, err
	}

	refList, err := remote.List(&git.ListOptions{InsecureSkipTLS: true})
	if err != nil {
		return nil, err
	}

	var branches []string
	refPrefix := "refs/heads/"

	for _, ref := range refList {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := refName[len(refPrefix):]
		branches = append(branches, branchName)
	}

	return branches, nil
}
