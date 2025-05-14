package local

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"connectrpc.com/connect"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v52/github"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
)

func (s *Server) GitStatus(ctx context.Context, r *connect.Request[localv1.GitStatusRequest]) (*connect.Response[localv1.GitStatusResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	err = gitutil.RunGitFetch(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	status, err := gitutil.RunGitStatus(s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GitStatusResponse{
		Branch:        status.Branch,
		ManagedGit:    project.ManagedGitId != "",
		LocalChanges:  status.LocalChanges,
		RemoteChanges: status.RemoteChanges,
	}), nil
}

func (s *Server) GitPull(ctx context.Context, r *connect.Request[localv1.GitPullRequest]) (*connect.Response[localv1.GitPullResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	var creds gitutil.GitRemoteCredentials
	if project.ManagedGitId == "" {
		creds = gitutil.GitRemoteCredentials{}
	} else {
		creds, err = s.app.ch.GitCredentials(ctx, project.OrgName, project.Name, s.app.ProjectPath)
		if err != nil {
			return nil, err
		}
	}
	_, err = gitutil.RunGitPull(ctx, s.app.ProjectPath, r.Msg.DiscardLocal, creds)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitPullResponse{}), nil
}

func (s *Server) GitPush(ctx context.Context, r *connect.Request[localv1.GitPushRequest]) (*connect.Response[localv1.GitPushResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	var creds gitutil.GitRemoteCredentials
	if project.ManagedGitId == "" {
		creds = gitutil.GitRemoteCredentials{}
	} else {
		creds, err = s.app.ch.GitCredentials(ctx, project.OrgName, project.Name, s.app.ProjectPath)
		if err != nil {
			return nil, err
		}
	}
	err = gitutil.CommitAndForcePush(ctx, s.app.ProjectPath, creds.Remote, creds.Username, creds.Password, r.Msg.Force)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitPushResponse{}), nil
}

// PushToGithub implements localv1connect.LocalServiceHandler.
// It assumes that the current project is not a git repo, it should generally be called after DeployValidation.
func (s *Server) PushToGithub(ctx context.Context, r *connect.Request[localv1.PushToGithubRequest]) (*connect.Response[localv1.PushToGithubResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}
	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	initGit := false
	// check if project has a git repo
	remote, _, err := gitutil.ExtractGitRemote(s.app.ProjectPath, "", false)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			initGit = true
		} else if !errors.Is(err, gitutil.ErrGitRemoteNotFound) {
			return nil, err
		}
	}
	if remote != nil {
		return nil, errors.New("git repository is already initialized with a remote")
	}

	gitStatus, err := c.GetGithubUserStatus(ctx, &adminv1.GetGithubUserStatusRequest{})
	if err != nil {
		return nil, err
	}
	if !gitStatus.HasAccess {
		// generally this should not happen as IsGithubConnected should be true before pushing to git
		return nil, fmt.Errorf("rill git app should be installed by user before pushing by visiting %s", gitStatus.GrantAccessUrl)
	}

	// if r.Msg.Account is empty, githubAccount will be "" which is equivalent to using default github account which is same as github username
	var githubAccount string
	if r.Msg.Account == "" || r.Msg.Account == gitStatus.Account {
		githubAccount = ""
	} else {
		githubAccount = r.Msg.Account
	}

	// check if we have write permission on the github account
	// this is a safety check as DeployValidation should take care of this
	if githubAccount == "" {
		if gitStatus.UserInstallationPermission != adminv1.GithubPermission_GITHUB_PERMISSION_WRITE {
			return nil, fmt.Errorf("rill github app should be installed with write permission on user personal account by visiting %s", gitStatus.GrantAccessUrl)
		}
	} else {
		valid := false
		for o, p := range gitStatus.OrganizationInstallationPermissions {
			if o == githubAccount && p == adminv1.GithubPermission_GITHUB_PERMISSION_WRITE {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("rill github app should be installed with write permission on organization %q by visiting %s", githubAccount, gitStatus.GrantAccessUrl)
		}
	}

	repoName := filepath.Base(s.app.ProjectPath)
	if r.Msg.Repo != "" {
		repoName = r.Msg.Repo
	}
	githubClient := github.NewTokenClient(ctx, gitStatus.AccessToken)
	defaultBranch := "main"

	// create remote repo
	suffix := 0
	var githubRepo *github.Repository
	name := repoName
	err = retrier.New(retrier.ConstantBackoff(retries, 1), nameConflictRetryErrClassifier{}).RunCtx(ctx, func(ctx context.Context) error {
		if suffix > 0 {
			name = fmt.Sprintf("%s-%d", repoName, suffix)
		}
		githubRepo, _, err = githubClient.Repositories.Create(ctx, githubAccount, &github.Repository{Name: &name, DefaultBranch: &defaultBranch})
		suffix++
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	var repo *git.Repository
	// init git repo
	if initGit {
		repo, err = git.PlainInitWithOptions(s.app.ProjectPath, &git.PlainInitOptions{
			InitOptions: git.InitOptions{
				DefaultBranch: plumbing.NewBranchReferenceName("main"),
			},
			Bare: false,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to init git repo: %w", err)
		}
	} else {
		repo, err = git.PlainOpen(s.app.ProjectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open git repo: %w", err)
		}
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// git add .
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return nil, fmt.Errorf("failed to add files to git: %w", err)
	}

	// git commit -m
	author, err := autoCommitGitSignature(ctx, c, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate git commit signature: %w", err)
	}
	_, err = wt.Commit("Auto committed by Rill", &git.CommitOptions{All: true, Author: author})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return nil, fmt.Errorf("failed to commit files to git: %w", err)
		}
	}

	// Create the remote
	_, err = repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{*githubRepo.HTMLURL}})
	if err != nil {
		return nil, fmt.Errorf("failed to create remote: %w", err)
	}

	// push the changes
	if err := repo.PushContext(ctx, &git.PushOptions{Auth: &githttp.BasicAuth{Username: "x-access-token", Password: gitStatus.AccessToken}}); err != nil {
		return nil, fmt.Errorf("failed to push to remote %q : %w", *githubRepo.HTMLURL, err)
	}

	account := githubAccount
	if account == "" {
		account = gitStatus.Account
	}

	return connect.NewResponse(&localv1.PushToGithubResponse{
		GithubUrl: *githubRepo.HTMLURL,
		Account:   account,
		Repo:      name,
	}), nil
}
