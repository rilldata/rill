package local

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
)

func (s *Server) GitStatus(ctx context.Context, r *connect.Request[localv1.GitStatusRequest]) (*connect.Response[localv1.GitStatusResponse], error) {
	// if it is not a git repo then throw a 400 error
	if !gitutil.IsGitRepo(s.app.ProjectPath) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("not a git repository"))
	}

	// if there is a origin set, try with native git configurations
	remote, err := gitutil.ExtractGitRemote(s.app.ProjectPath, "origin", false)
	if err == nil && remote.URL != "" {
		err = gitutil.GitFetch(ctx, s.app.ProjectPath, nil)
		if err == nil {
			// if native git fetch succeeds, return the status
			gs, err := gitutil.RunGitStatus(s.app.ProjectPath, "origin")
			if err != nil {
				return nil, err
			}
			return connect.NewResponse(&localv1.GitStatusResponse{
				Branch:        gs.Branch,
				GithubUrl:     gs.RemoteURL,
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
		gs, err := gitutil.RunGitStatus(s.app.ProjectPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:    gs.Branch,
			GithubUrl: gs.RemoteURL,
		}), nil
	}

	// to avoid asking user for inputs on UI simply used the last updated project for now
	name, err := inferRillManagedProjectName(ctx, s.app.ch, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		// If the project is not found return the best effort status
		gs, err := gitutil.RunGitStatus(s.app.ProjectPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:    gs.Branch,
			GithubUrl: gs.RemoteURL,
		}), nil
	}

	// get ephemeral git credentials
	config, err := s.app.ch.GitHelper(s.app.ch.Org, name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	// set remote
	// usually not needed but the older flow did not set the remote by name `rill`
	err = gitutil.SetRemote(s.app.ProjectPath, config)
	if err != nil {
		return nil, err
	}
	err = gitutil.GitFetch(ctx, s.app.ProjectPath, config)
	if err != nil {
		return nil, err
	}
	gs, err := gitutil.RunGitStatus(s.app.ProjectPath, config.RemoteName())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitStatusResponse{
		Branch:        gs.Branch,
		GithubUrl:     gs.RemoteURL,
		ManagedGit:    config.ManagedRepo,
		LocalChanges:  gs.LocalChanges,
		LocalCommits:  gs.LocalCommits,
		RemoteCommits: gs.RemoteCommits,
	}), nil
}

func (s *Server) GitPull(ctx context.Context, r *connect.Request[localv1.GitPullRequest]) (*connect.Response[localv1.GitPullResponse], error) {
	// if it is not a git repo then throw a 400 error
	if !gitutil.IsGitRepo(s.app.ProjectPath) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("not a git repository"))
	}

	origin, err := gitutil.ExtractGitRemote(s.app.ProjectPath, "origin", false)
	if err == nil && origin.URL != "" {
		out, err := gitutil.RunGitPull(ctx, s.app.ProjectPath, r.Msg.DiscardLocal, "", "origin")
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

	name, err := inferRillManagedProjectName(ctx, s.app.ch, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		return nil, errors.New("git credentials not set and repo is not connected to a project")
	}

	config, err := s.app.ch.GitHelper(s.app.ch.Org, name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	err = gitutil.SetRemote(s.app.ProjectPath, config)
	if err != nil {
		return nil, err
	}

	remote, err := config.FullyQualifiedRemote()
	if err != nil {
		return nil, err
	}

	out, err := gitutil.RunGitPull(ctx, s.app.ProjectPath, r.Msg.DiscardLocal, remote, config.RemoteName())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitPullResponse{
		Output: out,
	}), nil
}

func (s *Server) GitPush(ctx context.Context, r *connect.Request[localv1.GitPushRequest]) (*connect.Response[localv1.GitPushResponse], error) {
	// if it is not a git repo then throw a 400 error
	if !gitutil.IsGitRepo(s.app.ProjectPath) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("not a git repository"))
	}

	remote, err := gitutil.ExtractGitRemote(s.app.ProjectPath, "origin", false)
	if err == nil && remote.URL != "" {
		st, err := gitutil.RunGitStatus(s.app.ProjectPath, "origin")
		if err != nil {
			return nil, err
		}
		if st.RemoteCommits > 0 && !r.Msg.Force {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot push with remote commits present, please pull first"))
		}

		// generate git signature
		author, err := gitutil.NativeGitSignature(ctx, s.app.ProjectPath)
		if err == nil {
			err = gitutil.CommitAndForcePush(ctx, s.app.ProjectPath, &gitutil.Config{Remote: st.RemoteURL, DefaultBranch: st.Branch}, r.Msg.CommitMessage, author)
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

	name, err := inferRillManagedProjectName(ctx, s.app.ch, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		return nil, errors.New("git credentials not set and repo is not connected to a project")
	}

	author, err := s.app.ch.GitSignature(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	config, err := s.app.ch.GitHelper(s.app.ch.Org, name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	err = gitutil.SetRemote(s.app.ProjectPath, config)
	if err != nil {
		return nil, err
	}

	// fetch the status again
	gs, err := gitutil.RunGitStatus(s.app.ProjectPath, config.RemoteName())
	if err != nil {
		return nil, err
	}
	if gs.RemoteCommits > 0 && !r.Msg.Force {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot push with remote commits present, please pull first"))
	}

	err = gitutil.CommitAndForcePush(ctx, s.app.ProjectPath, config, r.Msg.CommitMessage, author)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GitPushResponse{}), nil
}

func inferRillManagedProjectName(ctx context.Context, h *cmdutil.Helper, org, pathToProject string) (string, error) {
	// Get the project name from the path
	projects, err := h.InferProjects(ctx, org, pathToProject)
	if err != nil {
		return "", err
	}

	if len(projects) == 1 {
		return projects[0].Name, nil
	}

	// in case of multiple projects, use the remote set in the current repo which will be set to the last used remote
	// this is to avoid asking the user for input on UI
	c := gitutil.Config{ManagedRepo: true}
	remote, _ := gitutil.ExtractGitRemote(pathToProject, c.RemoteName(), false)
	if remote.URL == "" {
		return projects[0].Name, nil
	}
	// filter projects by remote URL
	for _, p := range projects {
		if p.GitRemote == remote.URL {
			return p.Name, nil
		}
	}
	// if no project matches the remote URL, return the first project
	return projects[0].Name, nil
}
