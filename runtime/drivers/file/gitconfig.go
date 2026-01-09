package file

import (
	"context"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

// loadGitConfig loads the git configuration for the repository
// Should be called with c.gitMu held.
func (c *connection) loadGitConfig(ctx context.Context) (*gitutil.Config, error) {
	if c.gitConfig != nil && !c.gitConfig.IsExpired() {
		return c.gitConfig, nil
	}

	// Build request
	req := &adminv1.ListProjectsForFingerprintRequest{
		DirectoryName: filepath.Base(c.root),
	}

	// extract subpath
	repoRoot, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err == nil {
		req.SubPath = subpath
	}

	// extract remotes
	remote, err := gitutil.ExtractRemotes(repoRoot, false)
	if err == nil {
		for _, r := range remote {
			if r.Name == "__rill_remote" {
				req.RillMgdGitRemote = r.URL
			} else {
				gitRemote, err := r.Github()
				if err == nil {
					req.GitRemote = gitRemote
				}
			}
		}
	}
	resp, err := c.admin.ListProjectsForFingerprint(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Projects) == 0 {
		return nil, nil
	}

	orgFiltered := make([]*adminv1.Project, 0)
	for _, p := range resp.Projects {
		if p.OrgName == c.driverConfig.Org {
			orgFiltered = append(orgFiltered, p)
		}
	}
	if len(orgFiltered) == 0 {
		return nil, nil
	}
	p := orgFiltered[0]
	creds, err := c.admin.GetCloneCredentials(ctx, &adminv1.GetCloneCredentialsRequest{
		Org:     p.OrgName,
		Project: p.Name,
	})
	if err != nil {
		return nil, err
	}

	c.gitConfig = &gitutil.Config{
		Remote:            creds.GitRepoUrl,
		Username:          creds.GitUsername,
		Password:          creds.GitPassword,
		PasswordExpiresAt: creds.GitPasswordExpiresAt.AsTime(),
		DefaultBranch:     creds.GitPrimaryBranch,
		Subpath:           creds.GitSubpath,
		ManagedRepo:       creds.GitManagedRepo,
	}
	return c.gitConfig, nil
}

func (c *connection) gitSignature(ctx context.Context, path string) (*object.Signature, error) {
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
	userResp, err := c.admin.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return nil, err
	}

	return &object.Signature{
		Name:  userResp.User.DisplayName,
		Email: userResp.User.Email,
		When:  time.Now(),
	}, nil
}
