package local

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) GitStatus(ctx context.Context, r *connect.Request[localv1.GitStatusRequest]) (*connect.Response[localv1.GitStatusResponse], error) {
	gitPath, subPath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		// Not a git repo
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		// if not authenticated, still return local changes info for commit functionality
		st, err := gitutil.RunGitStatus(gitPath, subPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:        st.Branch,
			GithubUrl:     st.RemoteURL,
			Subpath:       subPath,
			ManagedGit:    false,
			LocalChanges:  st.LocalChanges,
			LocalCommits:  st.LocalCommits,
			RemoteCommits: st.RemoteCommits,
			HasUpstream:   st.HasUpstream,
		}), nil
	}

	// TODO: cache project inference
	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		// if not connected to a project, still return local changes info for commit functionality
		st, err := gitutil.RunGitStatus(gitPath, subPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:        st.Branch,
			GithubUrl:     st.RemoteURL,
			Subpath:       subPath,
			ManagedGit:    false,
			LocalChanges:  st.LocalChanges,
			LocalCommits:  st.LocalCommits,
			RemoteCommits: st.RemoteCommits,
			HasUpstream:   st.HasUpstream,
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
		HasUpstream:   gs.HasUpstream,
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
		return nil, errors.New("repo is not connected to a project")
	}
	project := projects[0]

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		// Not a git repo
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	if project.Subpath != subpath {
		return nil, errors.New("detected subpath within git repo does not match project subpath")
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

	out, err := gitutil.RunGitPull(ctx, gitPath, r.Msg.DiscardLocal, remote, config.RemoteName())
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
		return nil, errors.New("repo is not connected to a project")
	}
	project := projects[0]

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	// Not a git repo
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	if project.Subpath != subpath {
		return nil, errors.New("detected subpath within git repo does not match project subpath")
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

	var choice string
	if r.Msg.Force {
		choice = "2"
	} else {
		choice = "1"
	}
	err = s.app.ch.CommitAndSafePush(ctx, gitPath, config, r.Msg.CommitMessage, author, choice)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GitPushResponse{}), nil
}

func (s *Server) ListBranches(ctx context.Context, r *connect.Request[localv1.ListBranchesRequest]) (*connect.Response[localv1.ListBranchesResponse], error) {
	gitPath, _, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		// Not a git repo
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	result, err := gitutil.ListBranches(gitPath)
	if err != nil {
		return nil, err
	}

	// Convert to proto format
	branches := make([]*localv1.BranchInfo, len(result.Branches))
	for i, b := range result.Branches {
		branches[i] = &localv1.BranchInfo{
			Name:              b.Name,
			IsLocal:          b.IsLocal,
			IsRemote:         b.IsRemote,
			IsCurrent:        b.IsCurrent,
			LastCommitHash:   b.LastCommitHash,
			LastCommitMessage: b.LastCommitMessage,
			LastCommitTime:   timestamppb.New(b.LastCommitTime),
			Ahead:            b.Ahead,
			Behind:           b.Behind,
		}
	}

	return connect.NewResponse(&localv1.ListBranchesResponse{
		Branches:             branches,
		CurrentBranch:        result.CurrentBranch,
		HasUncommittedChanges: result.HasUncommittedChanges,
	}), nil
}

func (s *Server) CheckoutBranch(ctx context.Context, r *connect.Request[localv1.CheckoutBranchRequest]) (*connect.Response[localv1.CheckoutBranchResponse], error) {
	gitPath, _, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	err = gitutil.CheckoutBranch(gitPath, r.Msg.Branch, r.Msg.Force)
	if err != nil {
		return nil, err
	}

	// The file watcher in the runtime controller will automatically pick up changes
	// after branch checkout when files change on disk

	return connect.NewResponse(&localv1.CheckoutBranchResponse{}), nil
}

func (s *Server) CreateBranch(ctx context.Context, r *connect.Request[localv1.CreateBranchRequest]) (*connect.Response[localv1.CreateBranchResponse], error) {
	gitPath, _, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	branchInfo, err := gitutil.CreateBranch(gitPath, r.Msg.Name, r.Msg.Checkout)
	if err != nil {
		return nil, err
	}

	// The file watcher in the runtime controller will automatically pick up changes
	// if the checkout resulted in file changes

	return connect.NewResponse(&localv1.CreateBranchResponse{
		Branch: &localv1.BranchInfo{
			Name:              branchInfo.Name,
			IsLocal:          branchInfo.IsLocal,
			IsRemote:         branchInfo.IsRemote,
			IsCurrent:        branchInfo.IsCurrent,
			LastCommitHash:   branchInfo.LastCommitHash,
			LastCommitMessage: branchInfo.LastCommitMessage,
			LastCommitTime:   timestamppb.New(branchInfo.LastCommitTime),
		},
	}), nil
}

func (s *Server) DeleteBranch(ctx context.Context, r *connect.Request[localv1.DeleteBranchRequest]) (*connect.Response[localv1.DeleteBranchResponse], error) {
	gitPath, _, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	err = gitutil.DeleteBranch(gitPath, r.Msg.Name, r.Msg.Force)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.DeleteBranchResponse{}), nil
}

func (s *Server) GitMerge(ctx context.Context, r *connect.Request[localv1.GitMergeRequest]) (*connect.Response[localv1.GitMergeResponse], error) {
	gitPath, _, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	result, err := gitutil.GitMerge(ctx, gitPath, r.Msg.SourceBranch)
	if err != nil {
		return nil, err
	}

	// Generate PR URL if there are conflicts and we have a remote
	var prURL string
	if result.HasConflicts {
		// Get current branch and remote URL
		status, err := gitutil.RunGitStatus(gitPath, "", "origin")
		if err == nil && status.RemoteURL != "" {
			prURL = gitutil.GeneratePullRequestURL(status.RemoteURL, r.Msg.SourceBranch, status.Branch)
		}
	}

	// The file watcher in the runtime controller will automatically pick up changes
	// after a successful merge

	return connect.NewResponse(&localv1.GitMergeResponse{
		Success:          result.Success,
		HasConflicts:     result.HasConflicts,
		ConflictingFiles: result.ConflictingFiles,
		PullRequestUrl:   prURL,
	}), nil
}

func (s *Server) GetCommitHistory(ctx context.Context, r *connect.Request[localv1.GetCommitHistoryRequest]) (*connect.Response[localv1.GetCommitHistoryResponse], error) {
	gitPath, _, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	limit := int(r.Msg.Limit)
	if limit <= 0 {
		limit = 50
	}

	commits, totalCount, err := gitutil.GetCommitHistory(gitPath, r.Msg.Branch, limit, int(r.Msg.Offset))
	if err != nil {
		return nil, err
	}

	// Convert to proto format
	protoCommits := make([]*localv1.CommitInfo, len(commits))
	for i, c := range commits {
		protoCommits[i] = &localv1.CommitInfo{
			Hash:        c.Hash,
			ShortHash:   c.ShortHash,
			Message:     c.Message,
			AuthorName:  c.AuthorName,
			AuthorEmail: c.AuthorEmail,
			Timestamp:   timestamppb.New(c.Timestamp),
		}
	}

	return connect.NewResponse(&localv1.GetCommitHistoryResponse{
		Commits:    protoCommits,
		TotalCount: int32(totalCount),
	}), nil
}

func (s *Server) GitCommit(ctx context.Context, r *connect.Request[localv1.GitCommitRequest]) (*connect.Response[localv1.GitCommitResponse], error) {
	gitPath, _, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	// Get author info - try to get from authenticated user, fallback to defaults
	authorName := "Rill Developer"
	authorEmail := "developer@rilldata.com"
	if s.app.ch.IsAuthenticated() {
		signature, err := s.app.ch.GitSignature(ctx, gitPath)
		if err == nil {
			authorName = signature.Name
			authorEmail = signature.Email
		}
	}

	commit, err := gitutil.GitCommit(gitPath, r.Msg.Message, authorName, authorEmail)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GitCommitResponse{
		Commit: &localv1.CommitInfo{
			Hash:        commit.Hash,
			ShortHash:   commit.ShortHash,
			Message:     commit.Message,
			AuthorName:  commit.AuthorName,
			AuthorEmail: commit.AuthorEmail,
			Timestamp:   timestamppb.New(commit.Timestamp),
		},
	}), nil
}

func (s *Server) PublishBranch(ctx context.Context, r *connect.Request[localv1.PublishBranchRequest]) (*connect.Response[localv1.PublishBranchResponse], error) {
	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	// Get current branch and status
	status, err := gitutil.RunGitStatus(gitPath, subpath, "origin")
	if err != nil {
		return nil, err
	}

	if status.HasUpstream {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("branch already has a remote tracking branch, use push instead"))
	}

	// Check if we have a remote configured
	if status.RemoteURL == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("no remote configured, please connect to a remote repository first"))
	}

	// Publish the branch
	remoteName := "origin"
	if s.app.ch.IsAuthenticated() {
		projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
		if err == nil && len(projects) > 0 {
			project := projects[0]
			config, err := s.app.ch.GitHelper(s.app.ch.Org, project.Name, gitPath).GitConfig(ctx)
			if err == nil {
				remoteName = config.RemoteName()
			}
		}
	}

	err = gitutil.PublishBranch(ctx, gitPath, remoteName, status.Branch)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.PublishBranchResponse{
		Branch:    status.Branch,
		RemoteUrl: status.RemoteURL,
	}), nil
}

func (s *Server) DiscardChanges(ctx context.Context, r *connect.Request[localv1.DiscardChangesRequest]) (*connect.Response[localv1.DiscardChangesResponse], error) {
	gitPath, _, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	if len(r.Msg.Paths) > 0 {
		err = gitutil.DiscardChangesInPaths(ctx, gitPath, r.Msg.Paths)
	} else {
		err = gitutil.DiscardAllChanges(ctx, gitPath)
	}
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.DiscardChangesResponse{}), nil
}

func (s *Server) CreatePreviewDeployment(ctx context.Context, r *connect.Request[localv1.CreatePreviewDeploymentRequest]) (*connect.Response[localv1.CreatePreviewDeploymentResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	gitPath, subpath, err := gitutil.InferRepoRootAndSubpath(s.app.ProjectPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	// Get current branch
	status, err := gitutil.RunGitStatus(gitPath, subpath, "origin")
	if err != nil {
		return nil, err
	}
	if status.Branch == "" {
		return nil, errors.New("cannot create preview deployment: not on a branch")
	}

	// Get admin client
	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	// Push the branch to remote first (required for preview deployment)
	projects, err := s.app.ch.InferProjects(ctx, r.Msg.Org, s.app.ProjectPath)
	if err != nil {
		if !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return nil, err
		}
		return nil, errors.New("repo is not connected to a project")
	}
	project := projects[0]

	config, err := s.app.ch.GitHelper(r.Msg.Org, project.Name, gitPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	err = gitutil.SetRemote(gitPath, config)
	if err != nil {
		return nil, err
	}

	// Push the current branch
	author, err := s.app.ch.GitSignature(ctx, gitPath)
	if err != nil {
		return nil, err
	}
	err = s.app.ch.CommitAndSafePush(ctx, gitPath, config, "Preview deployment", author, "1")
	if err != nil {
		return nil, err
	}

	// Create preview deployment via admin API
	resp, err := c.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{
		Org:         r.Msg.Org,
		Project:     r.Msg.Project,
		Environment: "preview",
		Branch:      status.Branch,
	})
	if err != nil {
		return nil, err
	}

	// Construct frontend URL
	frontendURL := s.app.ch.AdminURL() + "/" + r.Msg.Org + "/" + r.Msg.Project + "/-/deployments/" + resp.Deployment.Id

	return connect.NewResponse(&localv1.CreatePreviewDeploymentResponse{
		DeploymentId: resp.Deployment.Id,
		FrontendUrl:  frontendURL,
		Branch:       status.Branch,
	}), nil
}

func (s *Server) ListPreviewDeployments(ctx context.Context, r *connect.Request[localv1.ListPreviewDeploymentsRequest]) (*connect.Response[localv1.ListPreviewDeploymentsResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	// Get deployments filtered by preview environment
	resp, err := c.ListDeployments(ctx, &adminv1.ListDeploymentsRequest{
		Org:         r.Msg.Org,
		Project:     r.Msg.Project,
		Environment: "preview",
	})
	if err != nil {
		return nil, err
	}

	// Convert to local proto format
	deployments := make([]*localv1.PreviewDeploymentInfo, len(resp.Deployments))
	for i, d := range resp.Deployments {
		frontendURL := s.app.ch.AdminURL() + "/" + r.Msg.Org + "/" + r.Msg.Project + "/-/deployments/" + d.Id
		deployments[i] = &localv1.PreviewDeploymentInfo{
			Id:          d.Id,
			Branch:      d.Branch,
			Status:      d.Status.String(),
			FrontendUrl: frontendURL,
			CreatedOn:   d.CreatedOn,
		}
	}

	return connect.NewResponse(&localv1.ListPreviewDeploymentsResponse{
		Deployments: deployments,
	}), nil
}

func (s *Server) DeletePreviewDeployment(ctx context.Context, r *connect.Request[localv1.DeletePreviewDeploymentRequest]) (*connect.Response[localv1.DeletePreviewDeploymentResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	// Delete the deployment via admin API
	resp, err := c.DeleteDeployment(ctx, &adminv1.DeleteDeploymentRequest{
		DeploymentId: r.Msg.DeploymentId,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.DeletePreviewDeploymentResponse{
		DeploymentId: resp.DeploymentId,
	}), nil
}
