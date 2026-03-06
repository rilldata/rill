package gitutil

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

var (
	gitignoreHasDotenvRegexp     = regexp.MustCompile(`(?m)^\.env$`)
	gitignoreHasDevDotenvRegexp  = regexp.MustCompile(`(?m)^\.dev\.env$`)
	gitignoreHasProdDotenvRegexp = regexp.MustCompile(`(?m)^\.prod\.env$`)
)

func EnsureGitignoreHasDotenv(ctx context.Context, repo drivers.RepoStore, path string) (bool, error) {
	var re *regexp.Regexp
	switch path {
	case ".env":
		re = gitignoreHasDotenvRegexp
	case ".dev.env":
		re = gitignoreHasDevDotenvRegexp
	case ".prod.env":
		re = gitignoreHasProdDotenvRegexp
	default:
		return false, fmt.Errorf("unsupported path %q, only `.env`, `.dev.env`, and `.prod.env` are supported", path)
	}
	return ensureGitignoreHas(ctx, repo, re, path)
}

func ensureGitignoreHas(ctx context.Context, repo drivers.RepoStore, regexp *regexp.Regexp, line string) (bool, error) {
	// Read .gitignore
	gitignore, _ := repo.Get(ctx, ".gitignore")

	// If .gitignore already has line, do nothing
	if regexp.MatchString(gitignore) {
		return false, nil
	}

	// Add line to the end of .gitignore
	if gitignore != "" {
		gitignore += "\n"
	}
	gitignore += line + "\n"

	// Write .gitignore
	err := repo.Put(ctx, ".gitignore", strings.NewReader(gitignore))
	if err != nil {
		return false, err
	}

	return true, nil
}
