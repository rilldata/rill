package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/sync/semaphore"
)

var (
	gitignoreHasDotenvRegexp       = regexp.MustCompile(`(?m)^\.env$`)
	gitignoreHasDotRillCloudRegexp = regexp.MustCompile(`(?m)^\s*\.rillcloud/`)
)

// GitHelper manages git operations for a project
// It also caches the git credentials for the project
type GitHelper struct {
	c         *client.Client
	org       string
	project   string
	localPath string

	// do not access gitConfig directly, use FetchGitConfig and setGitConfig
	gitConfig   *gitutil.Config
	gitConfigMu *semaphore.Weighted
}

func NewGitHelper(adminClient *client.Client, org, project, localPath string) *GitHelper {
	return &GitHelper{
		c:           adminClient,
		org:         org,
		project:     project,
		localPath:   localPath,
		gitConfigMu: semaphore.NewWeighted(1),
	}
}

func (g *GitHelper) FetchGitConfig(ctx context.Context) (*gitutil.Config, error) {
	err := g.gitConfigMu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer g.gitConfigMu.Release(1)
	if g.gitConfig != nil && !g.gitConfig.IsExpired() {
		return g.gitConfig, nil
	}

	resp, err := g.c.GetCloneCredentials(ctx, &adminv1.GetCloneCredentialsRequest{
		Organization: g.org,
		Project:      g.project,
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
	}
	return g.gitConfig, nil
}

func (g *GitHelper) PushToNewManagedRepo(ctx context.Context) (*adminv1.CreateManagedGitRepoResponse, error) {
	gitRepo, err := g.c.CreateManagedGitRepo(ctx, &adminv1.CreateManagedGitRepoRequest{
		Organization: g.org,
		Name:         g.project,
	})
	if err != nil {
		return nil, err
	}
	author, err := AutoCommitGitSignature(ctx, g.c, g.localPath)
	if err != nil {
		return nil, err
	}
	err = gitutil.CommitAndForcePush(ctx, g.localPath, gitRepo.Remote, gitRepo.Username, gitRepo.Password, gitRepo.DefaultBranch, author)
	if err != nil {
		return nil, err
	}

	err = g.setGitConfig(ctx, &gitutil.Config{
		Remote:            gitRepo.Remote,
		Username:          gitRepo.Username,
		Password:          gitRepo.Password,
		PasswordExpiresAt: gitRepo.PasswordExpiresAt.AsTime(),
		DefaultBranch:     gitRepo.DefaultBranch,
		Subpath:           "",
	})
	if err != nil {
		return nil, err
	}

	return gitRepo, nil
}

func (g *GitHelper) PushToManagedRepo(ctx context.Context) error {
	gitConfig, err := g.FetchGitConfig(ctx)
	if err != nil {
		return err
	}

	author, err := AutoCommitGitSignature(ctx, g.c, g.localPath)
	if err != nil {
		return err
	}
	err = gitutil.CommitAndForcePush(ctx, g.localPath, gitConfig.Remote, gitConfig.Username, gitConfig.Password, gitConfig.DefaultBranch, author)
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

func AutoCommitGitSignature(ctx context.Context, c adminv1.AdminServiceClient, path string) (*object.Signature, error) {
	repo, err := git.PlainOpen(path)
	if err == nil {
		cfg, err := repo.ConfigScoped(config.SystemScope)
		if err == nil && cfg.User.Email != "" && cfg.User.Name != "" {
			// user has git properly configured use that
			return &object.Signature{
				Name:  cfg.User.Name,
				Email: cfg.User.Email,
				When:  time.Now(),
			}, nil
		}
	}

	// use email of rill user
	userResp, err := c.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return nil, err
	}
	if userResp.User == nil {
		return nil, errors.New("failed to get current user")
	}

	return &object.Signature{
		Name:  userResp.User.DisplayName,
		Email: userResp.User.Email,
		When:  time.Now(),
	}, nil
}

func EnsureGitignoreHasDotenv(ctx context.Context, repo drivers.RepoStore) (bool, error) {
	return ensureGitignoreHas(ctx, repo, gitignoreHasDotenvRegexp, ".env")
}

func EnsureGitignoreHasDotRillCloud(ctx context.Context, repo drivers.RepoStore) (bool, error) {
	return ensureGitignoreHas(ctx, repo, gitignoreHasDotRillCloudRegexp, ".rillcloud/")
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
