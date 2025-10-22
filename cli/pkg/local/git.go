package local

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
)

func (s *Server) GitStatus(ctx context.Context, r *connect.Request[localv1.GitStatusRequest]) (*connect.Response[localv1.GitStatusResponse], error) {
	gitPath, subPath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	// Possibility not a git repo then throw a 400 error
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	// if there is a origin set, try with native git configurations
	remote, err := gitutil.ExtractGitRemote(gitPath, "origin", false)
	var remoteURL string
	if err == nil {
		remoteURL, _ = remote.Github()
	}
	if remoteURL == "" {
		// ignore subpath since git remote is non github and we can not use that
		subPath = ""
	}

	if err == nil && remoteURL != "" {
		err = gitutil.GitFetch(ctx, gitPath, nil)
		if err == nil {
			// if native git fetch succeeds, return the status
			gs, err := gitutil.RunGitStatus(gitPath, "origin")
			if err != nil {
				return nil, err
			}
			return connect.NewResponse(&localv1.GitStatusResponse{
				Branch:        gs.Branch,
				GithubUrl:     remoteURL,
				Subpath:       subPath,
				LocalChanges:  gs.LocalChanges,
				LocalCommits:  gs.LocalCommits,
				RemoteCommits: gs.RemoteCommits,
			}), nil
		}
	}

	// if native git fetch fails, try with ephemeral token - this may be a managed git project

	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		// if the user is not authenticated, we cannot fetch the project
		// return the best effort status
		gs, err := gitutil.RunGitStatus(gitPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:    gs.Branch,
			GithubUrl: remoteURL,
			Subpath:   subPath,
		}), nil
	}

	// to avoid asking user for inputs on UI simply used the last updated project for now
	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		// If the project is not found return the best effort status
		gs, err := gitutil.RunGitStatus(gitPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:    gs.Branch,
			GithubUrl: remoteURL,
			Subpath:   subPath,
		}), nil
	}
	project := projects[0]

	// get ephemeral git credentials
	config, err := s.app.ch.GitHelper(s.app.ch.Org, project.Name, gitPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	// set remote
	// usually not needed but the older flow did not set the remote by name `rill`
	err = gitutil.SetRemote(gitPath, config)
	if err != nil {
		return nil, err
	}
	err = gitutil.GitFetch(ctx, gitPath, config)
	if err != nil {
		return nil, err
	}
	gs, err := gitutil.RunGitStatus(gitPath, config.RemoteName())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitStatusResponse{
		Branch:        gs.Branch,
		GithubUrl:     remoteURL,
		Subpath:       subPath,
		ManagedGit:    config.ManagedRepo,
		LocalChanges:  gs.LocalChanges,
		LocalCommits:  gs.LocalCommits,
		RemoteCommits: gs.RemoteCommits,
	}), nil
}

func (s *Server) GithubRepoStatus(ctx context.Context, r *connect.Request[localv1.GithubRepoStatusRequest]) (*connect.Response[localv1.GithubRepoStatusResponse], error) {
	// Get an authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}
	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	// Forward the request to the admin server
	resp, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
		Remote: r.Msg.Remote,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GithubRepoStatusResponse{
		HasAccess:      resp.HasAccess,
		GrantAccessUrl: resp.GrantAccessUrl,
		DefaultBranch:  resp.DefaultBranch,
	}), nil
}

func (s *Server) GitPull(ctx context.Context, r *connect.Request[localv1.GitPullRequest]) (*connect.Response[localv1.GitPullResponse], error) {
	gitPath, err := gitutil.InferGitRepoRoot(s.app.ProjectPath)
	// Possibility not a git repo then throw a 400 error
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	origin, err := gitutil.ExtractGitRemote(gitPath, "origin", false)
	if err == nil && origin.URL != "" {
		out, err := gitutil.RunGitPull(ctx, gitPath, r.Msg.DiscardLocal, "", "origin")
		if err == nil && strings.Contains(out, "Already up to date") {
			return connect.NewResponse(&localv1.GitPullResponse{
				Output: out,
			}), nil
		}
	}
	// if native git pull fails, try with ephemeral token - this may be a managed git project

	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		return nil, errors.New("git credentials not set and repo is not connected to a project")
	}
	project := projects[0]

	config, err := s.app.ch.GitHelper(s.app.ch.Org, project.Name, gitPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	err = gitutil.SetRemote(gitPath, config)
	if err != nil {
		return nil, err
	}

	remote, err := config.FullyQualifiedRemote()
	if err != nil {
		return nil, err
	}

	out, err := gitutil.RunGitPull(ctx, gitPath, r.Msg.DiscardLocal, remote, config.RemoteName())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitPullResponse{
		Output: out,
	}), nil
}

func (s *Server) GitPush(ctx context.Context, r *connect.Request[localv1.GitPushRequest]) (*connect.Response[localv1.GitPushResponse], error) {
	gitPath, err := gitutil.InferGitRepoRoot(s.app.ProjectPath)
	// Possibility not a git repo then throw a 400 error
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	remote, err := gitutil.ExtractGitRemote(gitPath, "origin", false)
	if err == nil && remote.URL != "" {
		st, err := gitutil.RunGitStatus(gitPath, "origin")
		if err != nil {
			return nil, err
		}
		if st.RemoteCommits > 0 && !r.Msg.Force {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot push with remote commits present, please pull first"))
		}

		// generate git signature
		author, err := gitutil.NativeGitSignature(ctx, gitPath)
		if err == nil {
			err = gitutil.CommitAndForcePush(ctx, gitPath, &gitutil.Config{Remote: st.RemoteURL, DefaultBranch: st.Branch}, r.Msg.CommitMessage, author)
			if err == nil {
				return connect.NewResponse(&localv1.GitPushResponse{}), nil
			}
		}
	}
	// if native git push fails, try with ephemeral token - this may be a managed git project

	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		return nil, errors.New("git credentials not set and repo is not connected to a project")
	}
	project := projects[0]

	author, err := s.app.ch.GitSignature(ctx, gitPath)
	if err != nil {
		return nil, err
	}

	config, err := s.app.ch.GitHelper(s.app.ch.Org, project.Name, gitPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	err = gitutil.SetRemote(gitPath, config)
	if err != nil {
		return nil, err
	}

	// fetch the status again
	gs, err := gitutil.RunGitStatus(gitPath, config.RemoteName())
	if err != nil {
		return nil, err
	}
	if gs.RemoteCommits > 0 && !r.Msg.Force {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot push with remote commits present, please pull first"))
	}

	err = gitutil.CommitAndForcePush(ctx, gitPath, config, r.Msg.CommitMessage, author)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GitPushResponse{}), nil
}
