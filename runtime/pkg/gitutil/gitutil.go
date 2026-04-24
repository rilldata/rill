package gitutil

import (
	"context"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// EnsureGitignoreHas ensures the given path is listed in .gitignore.
// Returns true if .gitignore was modified.
func EnsureGitignoreHas(ctx context.Context, repo drivers.RepoStore, path string) (bool, error) {
	gitignore, _ := repo.Get(ctx, ".gitignore")

	// Check if any existing line already matches
	for _, line := range strings.Split(gitignore, "\n") {
		if strings.TrimSpace(line) == path {
			return false, nil
		}
	}

	// Append the path to .gitignore
	if gitignore != "" {
		gitignore += "\n"
	}
	gitignore += path + "\n"

	err := repo.Put(ctx, ".gitignore", strings.NewReader(gitignore))
	if err != nil {
		return false, err
	}

	return true, nil
}
