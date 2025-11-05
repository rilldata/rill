package local

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
)

func (s *Server) GitStatus(ctx context.Context, r *connect.Request[localv1.GitStatusRequest]) (*connect.Response[localv1.GitStatusResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		connect.NewError(connect.CodeFailedPrecondition, errors.New("must authenticate before performing this action"))
	}

	// TODO: cache project inference
	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		connect.NewError(connect.CodeFailedPrecondition, errors.New("not connected to any rill project"))
	}
	project := projects[0]

	gitPath, subPath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		// Possibility not a git repo then throw a 400 error
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	if subPath != project.Subpath {
		// TODO: subpath does not match
		return nil, connect.NewError(connect.CodeUnknown, errors.New("current path is not within the project's git subpath"))
	}

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
	gs, err := gitutil.RunGitStatus(gitPath, subPath, config.RemoteName())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitStatusResponse{
		Branch:        gs.Branch,
		GithubUrl:     gs.RemoteURL,
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

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	// Possibility not a git repo then throw a 400 error
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	if project.Subpath != subpath {
		return nil, errors.New("detected path does not match project git subpath")
	}

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

	out, err := gitutil.RunGitPull(ctx, gitPath, r.Msg.DiscardLocal, false, remote, config.RemoteName())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitPullResponse{
		Output: out,
	}), nil
}

func (s *Server) GitPush(ctx context.Context, r *connect.Request[localv1.GitPushRequest]) (*connect.Response[localv1.GitPushResponse], error) {
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

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	// Possibility not a git repo then throw a 400 error
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	if project.Subpath != subpath {
		return nil, errors.New("detected path does not match project git subpath")
	}

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
	gs, err := gitutil.RunGitStatus(gitPath, subpath, config.RemoteName())
	if err != nil {
		return nil, err
	}
	if gs.RemoteCommits > 0 && !r.Msg.Force {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot push with remote commits present, please pull first"))
	}

	err = s.app.ch.CommitAndSafePush(ctx, gitPath, config, r.Msg.CommitMessage, author)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GitPushResponse{}), nil
}
