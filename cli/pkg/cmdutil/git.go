package cmdutil

import (
	"context"
	"errors"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/dotgit"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

func (h *Helper) PushToNewManagedRepo(ctx context.Context, c *client.Client, org, project, path string) (*adminv1.CreateManagedGitRepoResponse, error) {
	gitRepo, err := c.CreateManagedGitRepo(ctx, &adminv1.CreateManagedGitRepoRequest{
		Organization: org,
		Name:         project,
	})
	if err != nil {
		return nil, err
	}
	author, err := AutoCommitGitSignature(ctx, c, path)
	if err != nil {
		return nil, err
	}
	err = gitutil.CommitAndForcePush(ctx, path, gitRepo.Remote, gitRepo.Username, gitRepo.Password, gitRepo.DefaultBranch, author)
	if err != nil {
		return nil, err
	}

	// also save the credentials in .git
	creds := &dotgit.GitConfig{
		Remote:         gitRepo.Remote,
		Username:       gitRepo.Username,
		Password:       gitRepo.Password,
		PasswordExpiry: gitRepo.PasswordExpiresAt.AsTime().Format(time.RFC3339),
		DefaultBranch:  gitRepo.DefaultBranch,
	}
	g := dotgit.New(path)
	err = g.StoreGitCredentials(creds)
	if err != nil {
		return nil, err
	}
	return gitRepo, nil
}

func (h *Helper) PushToManagedRepo(ctx context.Context, c *client.Client, org, project, path string) error {
	g := dotgit.New(path)
	gitConfig, err := g.LoadGitCredentials()
	if err != nil {
		return err
	}
	author, err := AutoCommitGitSignature(ctx, c, path)
	if err != nil {
		return err
	}
	err = gitutil.CommitAndForcePush(ctx, path, gitConfig.Remote, gitConfig.Username, gitConfig.Password, gitConfig.DefaultBranch, author)
	if err != nil {
		return err
	}
	return nil
}

func (h *Helper) GitCredentials(ctx context.Context, org, name, localPath string) (*dotgit.GitConfig, error) {
	g := dotgit.New(localPath)

	// Check if we have the git credentials in .git
	creds, err := g.LoadGitCredentials()
	if err != nil {
		return nil, err
	}
	if !creds.IsEmpty() && !creds.CredentialsExpired() {
		return creds, nil
	}

	resp, err := h.adminClient.GetCloneCredentials(ctx, &adminv1.GetCloneCredentialsRequest{
		Organization: org,
		Project:      name,
	})
	if err != nil {
		return nil, err
	}
	if resp.ArchiveDownloadUrl != "" {
		// Maybe download and automigrate to managed repo ??
		return nil, gitutil.ErrGitRemoteNotFound
	}
	creds = &dotgit.GitConfig{
		Remote:         resp.GitRepoUrl,
		Username:       resp.GitUsername,
		Password:       resp.GitPassword,
		PasswordExpiry: resp.GitPasswordExpiresAt.AsTime().Format(time.RFC3339),
		DefaultBranch:  resp.GitProdBranch,
		Subpath:        resp.GitSubpath,
	}
	err = g.StoreGitCredentials(creds)
	if err != nil {
		return nil, err
	}
	return creds, nil
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
