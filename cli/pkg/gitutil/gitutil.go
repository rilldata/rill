package gitutil

import (
	"fmt"
	"os"

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
	fmt.Println("repo name is", repoName)
	cmd := exec.Command("git", "clone", url)
	cmd.Stderr = os.Stderr
	_, err = cmd.Output()
	if err != nil {
		return "", err
	}

	return repoName, nil
}
