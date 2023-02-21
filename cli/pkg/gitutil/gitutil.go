package gitutil

import (
	"os"
	"strings"

	// "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	exec "golang.org/x/sys/execabs"
)

func CloneRepo(url string) (string, error) {
	endpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	repoName := strings.TrimSuffix(endpoint.Path[strings.LastIndex(endpoint.Path, "/")+1:], ".git")
	cmd := exec.Command("git", "clone", url)
	cmd.Stderr = os.Stderr
	_, err = cmd.Output()
	if err != nil {
		return "", err
	}

	return repoName, nil
}
