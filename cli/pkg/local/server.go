package local

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/dotrillcloud"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/pkce"
	"github.com/rilldata/rill/cli/pkg/update"
	"github.com/rilldata/rill/cli/pkg/web"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
	"github.com/rilldata/rill/proto/gen/rill/local/v1/localv1connect"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const retries = 3

// Server implements endpoints for the local Rill app (usually served on localhost).
type Server struct {
	logger   *zap.Logger
	app      *App
	metadata *localMetadata
}

var _ localv1connect.LocalServiceHandler = (*Server)(nil)

// RegisterHandlers registers the server's handlers on the provided ServeMux.
func (s *Server) RegisterHandlers(mux *http.ServeMux, httpPort int, secure, enableUI bool) {
	// Register local Connect (gRPC) service
	route, handler := localv1connect.NewLocalServiceHandler(s)
	mux.Handle(route, handler)

	// Register the local UI
	if enableUI {
		mux.Handle("/", web.StaticHandler())
	}

	// Register auth endpoints (starts and OAuth flow that leads to a token being set in ~/.rill)
	mux.Handle("/auth", s.authHandler(httpPort, secure))
	mux.Handle("/auth/callback", s.authCallbackHandler())
	mux.Handle("/auth/logout", s.logoutHandler())

	// Register telemetry proxy endpoint
	mux.Handle("/local/track", s.trackingHandler())

	// Deprecated: use proto RPCs instead
	mux.Handle("/local/config", s.metadataHandler())
	mux.Handle("/local/version", s.versionHandler())
}

// Ping implements localv1connect.LocalServiceHandler.
func (s *Server) Ping(ctx context.Context, r *connect.Request[localv1.PingRequest]) (*connect.Response[localv1.PingResponse], error) {
	return connect.NewResponse(&localv1.PingResponse{
		Time: timestamppb.Now(),
	}), nil
}

// GetMetadata implements localv1connect.LocalServiceHandler.
func (s *Server) GetMetadata(ctx context.Context, r *connect.Request[localv1.GetMetadataRequest]) (*connect.Response[localv1.GetMetadataResponse], error) {
	return connect.NewResponse(&localv1.GetMetadataResponse{
		InstanceId:       s.metadata.InstanceID,
		ProjectPath:      s.metadata.ProjectPath,
		InstallId:        s.metadata.InstallID,
		UserId:           s.metadata.UserID,
		Version:          s.metadata.Version,
		BuildCommit:      s.metadata.BuildCommit,
		BuildTime:        s.metadata.BuildTime,
		IsDev:            s.metadata.IsDev,
		AnalyticsEnabled: s.metadata.AnalyticsEnabled,
		Readonly:         s.metadata.Readonly,
		GrpcPort:         int32(s.metadata.GRPCPort),
		LoginUrl:         s.app.localURL + "/auth",
		AdminUrl:         s.app.ch.AdminURL(),
	}), nil
}

// GetVersion implements localv1connect.LocalServiceHandler.
func (s *Server) GetVersion(ctx context.Context, r *connect.Request[localv1.GetVersionRequest]) (*connect.Response[localv1.GetVersionResponse], error) {
	latestVersion, err := update.LatestVersion(ctx)
	if err != nil {
		s.logger.Warn("error finding latest version", zap.Error(err))
	}

	return connect.NewResponse(&localv1.GetVersionResponse{
		Current: s.app.ch.Version.Number,
		Latest:  latestVersion,
	}), nil
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
	_, err = wt.Commit("Auto committed by Rill", &git.CommitOptions{All: true})
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

// DeployProject implements localv1connect.LocalServiceHandler.
func (s *Server) DeployProject(ctx context.Context, r *connect.Request[localv1.DeployProjectRequest]) (*connect.Response[localv1.DeployProjectResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}
	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	// check if rill org exists
	_, err = c.GetOrganization(ctx, &adminv1.GetOrganizationRequest{
		Name: r.Msg.Org,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			// create org if not exists
			_, err = c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
				Name:        r.Msg.Org,
				Description: "Auto created by Rill",
			})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var projRequest *adminv1.CreateProjectRequest
	if r.Msg.Upload { // upload repo to rill managed storage instead of github
		repo, release, err := s.app.Runtime.Repo(ctx, s.app.Instance.ID)
		if err != nil {
			return nil, err
		}
		defer release()

		assetID, err := cmdutil.UploadRepo(ctx, repo, s.app.ch, r.Msg.Org, r.Msg.ProjectName)
		if err != nil {
			return nil, err
		}

		// create project request
		projRequest = &adminv1.CreateProjectRequest{
			OrganizationName: r.Msg.Org,
			Name:             r.Msg.ProjectName,
			Description:      "Auto created by Rill",
			Provisioner:      "",
			ProdVersion:      "",
			ProdOlapDriver:   "duckdb",
			ProdOlapDsn:      "",
			ProdSlots:        int64(DefaultProdSlots(s.app.ch)),
			Public:           false,
			ArchiveAssetId:   assetID,
		}
	} else {
		userStatus, err := c.GetGithubUserStatus(ctx, &adminv1.GetGithubUserStatusRequest{})
		if err != nil {
			return nil, err
		}
		if !userStatus.HasAccess {
			// generally this should not happen as IsGithubConnected should be true before deploying
			return nil, fmt.Errorf("rill git app should be installed/authorized by user before deploying, please visit %s", userStatus.GrantAccessUrl)
		}

		// check if project is a git repo
		remote, ghURL, err := gitutil.ExtractGitRemote(s.app.ProjectPath, "", false)
		if err != nil {
			if errors.Is(err, gitutil.ErrGitRemoteNotFound) || errors.Is(err, git.ErrRepositoryNotExists) {
				return nil, errors.New("project is not a valid git repository or not connected to a remote")
			}
			return nil, err
		}

		// check if there are uncommitted changes
		// ignore errors since check is best effort and can fail in multiple cases
		syncStatus, _ := gitutil.GetSyncStatus(s.app.ProjectPath, "", remote.Name)
		if syncStatus == gitutil.SyncStatusModified || syncStatus == gitutil.SyncStatusAhead {
			return nil, errors.New("project has uncommitted changes")
		}

		// Get github repo status
		repoStatus, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
			GithubUrl: ghURL,
		})
		if err != nil {
			return nil, err
		}
		if !repoStatus.HasAccess {
			// generally this should not happen as IsRepoAccessGranted should be true before deploying
			return nil, fmt.Errorf("need access to the repository before deploying, please visit %s to grant access", repoStatus.GrantAccessUrl)
		}
		projRequest = &adminv1.CreateProjectRequest{
			OrganizationName: r.Msg.Org,
			Name:             r.Msg.ProjectName,
			Description:      "Auto created by Rill",
			Provisioner:      "",
			ProdVersion:      "",
			ProdOlapDriver:   "duckdb",
			ProdOlapDsn:      "",
			ProdSlots:        int64(DefaultProdSlots(s.app.ch)),
			Public:           false,
			GithubUrl:        ghURL,
			Subpath:          "",
			ProdBranch:       repoStatus.DefaultBranch,
		}
	}

	// create project
	suffix := 0
	var projResp *adminv1.CreateProjectResponse
	err = retrier.New(retrier.ConstantBackoff(retries, 1), nameConflictRetryErrClassifier{}).RunCtx(ctx, func(ctx context.Context) error {
		name := r.Msg.ProjectName
		if suffix > 0 {
			name = fmt.Sprintf("%s-%d", r.Msg.ProjectName, suffix)
		}
		projRequest.Name = name
		projResp, err = c.CreateProject(ctx, projRequest)
		suffix++
		return err
	})
	if err != nil {
		return nil, err
	}

	err = dotrillcloud.SetAll(s.app.ProjectPath, s.app.ch.AdminURL(), &dotrillcloud.Config{
		ProjectID: projResp.Project.Id,
	})
	if err != nil {
		return nil, err
	}

	// Parse .env and push it as variables
	dotenv, err := ParseDotenv(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .env: %w", err)
	}
	if len(dotenv) > 0 {
		_, err = c.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
			Organization: r.Msg.Org,
			Project:      r.Msg.ProjectName,
			Variables:    dotenv,
		})
		if err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&localv1.DeployProjectResponse{
		DeployId:    projResp.Project.ProdDeploymentId,
		Org:         projResp.Project.OrgName,
		Project:     projResp.Project.Name,
		FrontendUrl: projResp.Project.FrontendUrl,
	}), nil
}

// RedeployProject implements localv1connect.LocalServiceHandler.
func (s *Server) RedeployProject(ctx context.Context, r *connect.Request[localv1.RedeployProjectRequest]) (*connect.Response[localv1.RedeployProjectResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}
	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	projResp, err := c.GetProjectByID(ctx, &adminv1.GetProjectByIDRequest{
		Id: r.Msg.ProjectId,
	})
	if err != nil {
		return nil, err
	}

	if r.Msg.Reupload {
		repo, release, err := s.app.Runtime.Repo(ctx, s.app.Instance.ID)
		if err != nil {
			return nil, err
		}
		defer release()

		assetID, err := cmdutil.UploadRepo(ctx, repo, s.app.ch, projResp.Project.OrgName, projResp.Project.Name)
		if err != nil {
			return nil, err
		}
		_, err = c.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
			ArchiveAssetId:   &assetID,
			OrganizationName: projResp.Project.OrgName,
			Name:             projResp.Project.Name,
		})
		if err != nil {
			return nil, err
		}
	}

	// Parse .env and push it as variables
	dotenv, err := ParseDotenv(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .env: %w", err)
	}
	if len(dotenv) > 0 {
		_, err = c.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
			Organization: projResp.Project.OrgName,
			Project:      projResp.Project.Name,
			Variables:    dotenv,
		})
		if err != nil {
			return nil, err
		}
	}

	// TODO : Add other update project fields
	return connect.NewResponse(&localv1.RedeployProjectResponse{
		FrontendUrl: projResp.Project.FrontendUrl,
	}), nil
}

// GetCurrentUser implements localv1connect.LocalServiceHandler.
func (s *Server) GetCurrentUser(ctx context.Context, r *connect.Request[localv1.GetCurrentUserRequest]) (*connect.Response[localv1.GetCurrentUserResponse], error) {
	if !s.app.ch.IsAuthenticated() {
		return connect.NewResponse(&localv1.GetCurrentUserResponse{
			User: nil,
		}), nil
	}

	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	userResp, err := c.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return nil, err
	}
	if userResp.User == nil {
		return nil, errors.New("failed to get current user")
	}

	// get rill user orgs
	resp, err := c.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{PageSize: 1000})
	if err != nil {
		return nil, err
	}

	userOrgs := make([]string, 0, len(resp.Organizations))
	for _, org := range resp.Organizations {
		userOrgs = append(userOrgs, org.Name)
	}

	representingUser, err := dotrill.GetRepresentingUser()
	if err != nil {
		return nil, errors.New("failed to get assumed user email")
	}
	isRepresentingUser := false
	if representingUser != "" {
		isRepresentingUser = true
	}

	return connect.NewResponse(&localv1.GetCurrentUserResponse{
		User: &adminv1.User{
			Id:          userResp.User.Id,
			Email:       userResp.User.Email,
			DisplayName: userResp.User.DisplayName,
			PhotoUrl:    userResp.User.PhotoUrl,
		},
		RillUserOrgs:       userOrgs,
		IsRepresentingUser: isRepresentingUser,
	}), nil
}

// GetCurrentProject implements localv1connect.LocalServiceHandler.
func (s *Server) GetCurrentProject(ctx context.Context, r *connect.Request[localv1.GetCurrentProjectRequest]) (*connect.Response[localv1.GetCurrentProjectResponse], error) {
	localProjectName := filepath.Base(s.app.ProjectPath)

	// Return early if the user isn't logged in
	if !s.app.ch.IsAuthenticated() {
		return connect.NewResponse(&localv1.GetCurrentProjectResponse{
			LocalProjectName: localProjectName,
		}), nil
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GetCurrentProjectResponse{
		LocalProjectName: localProjectName,
		Project:          project,
	}), nil
}

func (s *Server) ListOrganizationsAndBillingMetadata(ctx context.Context, r *connect.Request[localv1.ListOrganizationsAndBillingMetadataRequest]) (*connect.Response[localv1.ListOrganizationsAndBillingMetadataResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}
	c, err := s.app.ch.Client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{PageSize: 1000})
	if err != nil {
		return nil, err
	}

	orgsMetadata := make([]*localv1.ListOrganizationsAndBillingMetadataResponse_OrgMetadata, len(resp.Organizations))
	for i, org := range resp.Organizations {
		issues, err := c.ListOrganizationBillingIssues(ctx, &adminv1.ListOrganizationBillingIssuesRequest{
			Organization: org.Name,
		})
		if err != nil {
			return nil, err
		}

		orgsMetadata[i] = &localv1.ListOrganizationsAndBillingMetadataResponse_OrgMetadata{
			Name:   org.Name,
			Issues: issues.Issues,
		}
	}

	return connect.NewResponse(&localv1.ListOrganizationsAndBillingMetadataResponse{
		Orgs: orgsMetadata,
	}), nil
}

// authHandler starts the OAuth2 PKCE flow to authenticate the user and get a rill access token.
func (s *Server) authHandler(httpPort int, secure bool) http.Handler {
	scheme := "http"
	if secure {
		scheme = "https"
	}
	redirectURL := fmt.Sprintf("%s://localhost:%d/auth/callback", scheme, httpPort)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// generate random state
		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
			return
		}
		state := base64.URLEncoding.EncodeToString(b)

		// check the request for redirect query param, we will use this to redirect back to this after auth
		origin := r.URL.Query().Get("redirect")
		if origin == "" {
			origin = "/"
		}

		authenticator, err := pkce.NewAuthenticator(s.app.ch.AdminURL(), redirectURL, database.AuthClientIDRillWebLocal, origin)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to generate pkce authenticator: %s", err), http.StatusInternalServerError)
			return
		}
		s.app.pkceAuthenticators[state] = authenticator
		authURL := authenticator.GetAuthURL(state)
		http.Redirect(w, r, authURL, http.StatusFound)
	})
}

// authCallbackHandler handles the OAuth2 PKCE callback to exchange the authorization code for a rill access token.
func (s *Server) authCallbackHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		state := r.URL.Query().Get("state")
		if code == "" {
			http.Error(w, "missing state", http.StatusBadRequest)
			return
		}

		authenticator, ok := s.app.pkceAuthenticators[state]
		if !ok {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		// remove authenticator from map
		delete(s.app.pkceAuthenticators, state)

		if authenticator == nil {
			http.Error(w, "failed to get authenticator", http.StatusInternalServerError)
			return
		}

		// Exchange the code for an access token
		token, err := authenticator.ExchangeCodeForToken(code)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to exchange code for token: %s", err), http.StatusInternalServerError)
			return
		}

		// Save token and reload config
		err = dotrill.SetAccessToken(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to save access token: %s", err), http.StatusInternalServerError)
			return
		}
		err = s.app.ch.ReloadAdminConfig()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to reload admin config: %s", err), http.StatusInternalServerError)
			return
		}

		// Redirect back to url provided by caller when initiating auth flow
		http.Redirect(w, r, authenticator.OriginURL, http.StatusFound)
	})
}

// logoutHandler logs out the user and unsets the token stored
func (s *Server) logoutHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Logout the CLI
		err := auth.Logout(r.Context(), s.app.ch)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to logout: %s", err), http.StatusInternalServerError)
			return
		}

		// Get URL for cloud auth.
		// NOTE: This is temporary until we migrate to a server that can host HTTP and gRPC on the same port.
		authURL := s.app.ch.AdminURL()
		if strings.Contains(authURL, "http://localhost:9090") {
			authURL = "http://localhost:8080"
		}

		// Logout on cloud as well
		var qry map[string]string
		if r.URL.Query().Get("redirect") != "" {
			qry = map[string]string{"redirect": r.URL.Query().Get("redirect")}
		}
		logoutURL, err := urlutil.WithQuery(urlutil.MustJoinURL(authURL, "auth", "logout"), qry)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create logout URL: %s", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, logoutURL, http.StatusFound)
	})
}

// trackingHandler proxies events to intake.rilldata.io.
func (s *Server) trackingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read entire body up front (since it may be closed before the request is sent in the goroutine below)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.logger.Info("failed to read telemetry request", zap.Error(err))
			w.WriteHeader(http.StatusOK)
			return
		}

		// Parse the body as JSON
		var event map[string]any
		err = json.Unmarshal(body, &event)
		if err != nil {
			s.logger.Info("failed to parse telemetry request", zap.Error(err))
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass as raw event to the telemetry client
		err = s.app.ch.Telemetry(r.Context()).RecordRaw(event)
		if err != nil {
			s.logger.Info("failed to proxy telemetry event from UI", zap.Error(err))
		}
		w.WriteHeader(http.StatusOK)
	})
}

// localMetadata contains metadata about the current project and Rill configuration.
type localMetadata struct {
	InstanceID       string `json:"instance_id"`
	ProjectPath      string `json:"project_path"`
	InstallID        string `json:"install_id"`
	UserID           string `json:"user_id"`
	Version          string `json:"version"`
	BuildCommit      string `json:"build_commit"`
	BuildTime        string `json:"build_time"`
	IsDev            bool   `json:"is_dev"`
	AnalyticsEnabled bool   `json:"analytics_enabled"`
	Readonly         bool   `json:"readonly"`
	GRPCPort         int    `json:"grpc_port"`
}

// metadataHandler serves the metadata of the local Rill instance.
func (s *Server) metadataHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(s.metadata)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, err = w.Write(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write response data: %s", err), http.StatusInternalServerError)
			return
		}
	})
}

// versionResponse is the response format for versionHandler.
type versionResponse struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
}

// versionHandler servers the current and latest version of the Rill CLI.
func (s *Server) versionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the latest version available
		latestVersion, err := update.LatestVersion(r.Context())
		if err != nil {
			s.logger.Warn("error finding latest version", zap.Error(err))
		}

		inf := &versionResponse{
			CurrentVersion: s.app.ch.Version.Number,
			LatestVersion:  latestVersion,
		}

		data, err := json.Marshal(inf)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		_, err = w.Write(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write response data: %s", err), http.StatusInternalServerError)
			return
		}
	})
}

// nameConflictRetryErrClassifier classifies name already exists errors as retryable, works for both github repo and project name
type nameConflictRetryErrClassifier struct{}

func (nameConflictRetryErrClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}

	if strings.Contains(err.Error(), "name already exists") {
		return retrier.Retry
	}

	return retrier.Fail
}
