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
	gitPath, subPath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		// Not a git repo
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		// if not authenticated do not return local/remote changes info
		st, err := gitutil.RunGitStatus(gitPath, subPath, "origin", "")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:     st.Branch,
			GithubUrl:  st.RemoteURL,
			Subpath:    subPath,
			ManagedGit: false,
		}), nil
	}

	// TODO: cache project inference
	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrInferProjectFailed) {
			return nil, err
		}
		// if not connected to a project do not return local/remote changes info
		st, err := gitutil.RunGitStatus(gitPath, subPath, "origin", "")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:     st.Branch,
			GithubUrl:  st.RemoteURL,
			Subpath:    subPath,
			ManagedGit: false,
		}), nil
	}
	project := projects[0]

	if subPath != project.Subpath {
		// unlikely but just in case
		return nil, connect.NewError(connect.CodeUnknown, errors.New("detected subpath within git repo does not match project subpath"))
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
	gs, err := gitutil.RunGitStatus(gitPath, subPath, config.RemoteName(), "")
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

func (s *Server) CreateGithubPullRequest(ctx context.Context, r *connect.Request[localv1.CreateGithubPullRequestRequest]) (*connect.Response[localv1.CreateGithubPullRequestResponse], error) {
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}
	project := projects[0]

	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateGithubPullRequest(ctx, &adminv1.CreateGithubPullRequestRequest{
		Org:     project.OrgName,
		Project: project.Name,
		Branch:  r.Msg.Branch,
		Title:   r.Msg.Title,
		Body:    r.Msg.Body,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.CreateGithubPullRequestResponse{
		PrUrl: resp.PrUrl,
	}), nil
}

func (s *Server) GetGithubPullRequest(ctx context.Context, r *connect.Request[localv1.GetGithubPullRequestRequest]) (*connect.Response[localv1.GetGithubPullRequestResponse], error) {
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}
	project := projects[0]

	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetGithubPullRequest(ctx, &adminv1.GetGithubPullRequestRequest{
		Org:     project.OrgName,
		Project: project.Name,
		Branch:  r.Msg.Branch,
	})
	if err != nil {
		return nil, err
	}

	state := localv1.GetGithubPullRequestResponse_STATE_UNSPECIFIED
	switch resp.PrState {
	case adminv1.GetGithubPullRequestResponse_STATE_OPEN:
		state = localv1.GetGithubPullRequestResponse_STATE_OPEN
	case adminv1.GetGithubPullRequestResponse_STATE_CLOSED_UNMERGED:
		state = localv1.GetGithubPullRequestResponse_STATE_CLOSED_UNMERGED
	case adminv1.GetGithubPullRequestResponse_STATE_MERGED:
		state = localv1.GetGithubPullRequestResponse_STATE_MERGED
	}
	return connect.NewResponse(&localv1.GetGithubPullRequestResponse{
		PrUrl:   resp.PrUrl,
		PrState: state,
	}), nil
}
