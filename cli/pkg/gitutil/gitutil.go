package gitutil

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	// "github.com/go-git/go-git/v5/plumbing"
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

	r, _ := ExtractRemotes(repoName)
	fmt.Println("Paths are", r)
	return repoName, nil
}

var ErrNotGitProject = errors.New("not a git project")

func ExtractRemotes(projectPath string) ([]string, error) {
	fmt.Println("inside ExtractRemotes ", projectPath)
	// Attempt to clone local repo
	// repo, err := git.PlainOpenWithOptions("/Users/rakeshrilldata/Workspace/gotrue", &git.PlainOpenOptions{DetectDotGit: true})
	// repo, err := git.PlainOpen("/Users/rakeshrilldata/Workspace/test/cabfinder")
	repo, err := git.PlainOpen(projectPath)
	if err != nil {
		fmt.Println("Error in plain open", err)
		return nil, ErrNotGitProject
	}

	// c, _ := repo.Config()
	// projectUrl := c.Raw.Section("remote").Subsection("origin").Options.Get("url")
	// fmt.Println("Raw is", projectUrl)

	// endpoint, err := transport.NewEndpoint(projectUrl)
	// if err != nil {
	// 	fmt.Println("error endpoint", err)
	// }

	// fmt.Println("endpoints", endpoint)

	// fmt.Println("repo details", c.Remotes)
	remote, err := repo.Remote("origin")
	if err != nil {
		fmt.Println("error in remote", err)
		return nil, err
	}

	refList, err := remote.List(&git.ListOptions{})
	if err != nil {
		fmt.Println("error in list", err)
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
