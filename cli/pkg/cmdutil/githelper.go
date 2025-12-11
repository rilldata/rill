package cmdutil

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/sync/semaphore"
)

var gitignoreHasDotenvRegexp = regexp.MustCompile(`(?m)^\.env$`)

// GitHelper manages git operations for a project.
// It also caches the git credentials for the project.
// Do not use directly, use cmdutil.Helper to get an instance of GitHelper.
type GitHelper struct {
	h         *Helper
	org       string
	project   string
	localPath string

	// do not access gitConfig directly, use GitConfig and setGitConfig
	gitConfig   *gitutil.Config
	gitConfigMu *semaphore.Weighted
}

func newGitHelper(h *Helper, org, project, localPath string) *GitHelper {
	return &GitHelper{
		h:           h,
		org:         org,
		project:     project,
		localPath:   localPath,
		gitConfigMu: semaphore.NewWeighted(1),
	}
}

func (g *GitHelper) GitConfig(ctx context.Context) (*gitutil.Config, error) {
	err := g.gitConfigMu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer g.gitConfigMu.Release(1)
	if g.gitConfig != nil && !g.gitConfig.IsExpired() {
		return g.gitConfig, nil
	}

	c, err := g.h.Client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetCloneCredentials(ctx, &adminv1.GetCloneCredentialsRequest{
		Org:     g.org,
		Project: g.project,
	})
	if err != nil {
		return nil, err
	}
	if resp.GitRepoUrl == "" {
		return nil, fmt.Errorf("project %q is not connected to a git repository", g.project)
	}
	g.gitConfig = &gitutil.Config{
		Remote:            resp.GitRepoUrl,
		Username:          resp.GitUsername,
		Password:          resp.GitPassword,
		PasswordExpiresAt: resp.GitPasswordExpiresAt.AsTime(),
		DefaultBranch:     resp.GitProdBranch,
		Subpath:           resp.GitSubpath,
		ManagedRepo:       resp.GitManagedRepo,
	}
	return g.gitConfig, nil
}

func (g *GitHelper) PushToNewManagedRepo(ctx context.Context) (*adminv1.CreateManagedGitRepoResponse, error) {
	c, err := g.h.Client()
	if err != nil {
		return nil, err
	}

	gitRepo, err := c.CreateManagedGitRepo(ctx, &adminv1.CreateManagedGitRepoRequest{
		Org:  g.org,
		Name: g.project,
	})
	if err != nil {
		return nil, err
	}
	author, err := g.h.GitSignature(ctx, g.localPath)
	if err != nil {
		return nil, err
	}
	config := &gitutil.Config{
		Remote:            gitRepo.Remote,
		Username:          gitRepo.Username,
		Password:          gitRepo.Password,
		PasswordExpiresAt: gitRepo.PasswordExpiresAt.AsTime(),
		DefaultBranch:     gitRepo.DefaultBranch,
		ManagedRepo:       true,
	}
	err = gitutil.CommitAndPush(ctx, g.localPath, config, "", author)
	if err != nil {
		return nil, err
	}

	err = g.setGitConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return gitRepo, nil
}

func (g *GitHelper) PushToManagedRepo(ctx context.Context) error {
	gitConfig, err := g.GitConfig(ctx)
	if err != nil {
		return err
	}

	author, err := g.h.GitSignature(ctx, g.localPath)
	if err != nil {
		return err
	}
	err = g.h.CommitAndSafePush(ctx, g.localPath, gitConfig, "", author, "1")
	if err != nil {
		return err
	}
	return nil
}

func (g *GitHelper) setGitConfig(ctx context.Context, c *gitutil.Config) error {
	err := g.gitConfigMu.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer g.gitConfigMu.Release(1)

	g.gitConfig = c
	return nil
}

func SetupGitIgnore(ctx context.Context, repo drivers.RepoStore) error {
	// Ensure .gitignore exists and contains necessary entries
	contents, err := repo.Get(ctx, ".gitignore")
	if err != nil {
		if !strings.Contains(err.Error(), "no such file") {
			return err
		}
		// Create .gitignore if it does not exist
		err = repo.Put(ctx, ".gitignore", strings.NewReader(".DS_Store\n\n# Rill\n.env\ntmp\n"))
		if err != nil {
			return err
		}
		return nil
	}

	gitIgnoreContent := strings.ReplaceAll(contents, "\r\n", "\n")
	gitIgnoreEntries := strings.Split(gitIgnoreContent, "\n")
	var added bool
	for _, path := range []string{".DS_Store", ".env", "tmp"} {
		if slices.Contains(gitIgnoreEntries, path) {
			continue // already exists
		}
		added = true
		gitIgnoreContent += fmt.Sprintf("\n%s", path)
	}
	if !added {
		return nil // nothing to add
	}
	return repo.Put(ctx, ".gitignore", strings.NewReader(gitIgnoreContent))
}

func EnsureGitignoreHasDotenv(ctx context.Context, repo drivers.RepoStore) (bool, error) {
	return ensureGitignoreHas(ctx, repo, gitignoreHasDotenvRegexp, ".env")
}

func ensureGitignoreHas(ctx context.Context, repo drivers.RepoStore, regexp *regexp.Regexp, line string) (bool, error) {
	// Read .gitignore
	gitignore, _ := repo.Get(ctx, ".gitignore")

	// If .gitignore already has .env, do nothing
	if regexp.MatchString(gitignore) {
		return false, nil
	}

	// Add .env to the end of .gitignore
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
