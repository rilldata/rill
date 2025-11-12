package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v71/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/admin/server/auth"
	cligitutil "github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/archive"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	githubcookieName          = "github_auth"
	githubcookieFieldState    = "github_state"
	githubcookieFieldRemote   = "github_remote"
	githubcookieFieldRedirect = "github_redirect"
	archivePullTimeout        = 10 * time.Minute
	createRetries             = 3
)

func (s *Server) GetGithubUserStatus(ctx context.Context, req *adminv1.GetGithubUserStatusRequest) (*adminv1.GetGithubUserStatusResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}
	if user.GithubUsername == "" {
		// If we don't have user's github username we navigate user to installtion assuming they never installed github app
		return &adminv1.GetGithubUserStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: s.admin.URLs.GithubConnect(""),
		}, nil
	}
	token, err := s.userAccessToken(ctx, user)
	if err != nil {
		// token not valid or expired, take auth again
		return &adminv1.GetGithubUserStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: s.admin.URLs.GithubAuth(""),
		}, nil
	}

	userInstallationPermission := adminv1.GithubPermission_GITHUB_PERMISSION_UNSPECIFIED
	installation, _, err := s.admin.Github.AppClient().Apps.FindUserInstallation(ctx, user.GithubUsername)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return nil, fmt.Errorf("failed to get user installation: %w", err)
		}
	} else {
		// older git app would ask for Contents=read permission whereas new one asks for Contents=write and && Administration=write
		if installation.Permissions != nil && installation.Permissions.Contents != nil && strings.EqualFold(*installation.Permissions.Contents, "read") {
			userInstallationPermission = adminv1.GithubPermission_GITHUB_PERMISSION_READ
		}

		if installation.Permissions != nil && installation.Permissions.Contents != nil && installation.Permissions.Administration != nil && strings.EqualFold(*installation.Permissions.Administration, "write") && strings.EqualFold(*installation.Permissions.Contents, "write") {
			userInstallationPermission = adminv1.GithubPermission_GITHUB_PERMISSION_WRITE
		}
	}

	client := github.NewTokenClient(ctx, token)
	// List all the private organizations for the authenticated user
	orgs, _, err := client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}
	// List all the public organizations for the authenticated user
	publicOrgs, _, err := client.Organizations.List(ctx, user.GithubUsername, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	orgs = append(orgs, publicOrgs...)
	allOrgs := make([]string, 0)

	orgInstallationPermission := make(map[string]adminv1.GithubPermission)
	for _, org := range orgs {
		// dedupe orgs
		if _, ok := orgInstallationPermission[org.GetLogin()]; ok {
			continue
		}
		allOrgs = append(allOrgs, org.GetLogin())

		i, _, err := s.admin.Github.AppClient().Apps.FindOrganizationInstallation(ctx, org.GetLogin())
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				orgInstallationPermission[org.GetLogin()] = adminv1.GithubPermission_GITHUB_PERMISSION_UNSPECIFIED
				continue
			}
			return nil, fmt.Errorf("failed to get organization installation: %w", err)
		}
		permission := adminv1.GithubPermission_GITHUB_PERMISSION_UNSPECIFIED
		// older git app would ask for Contents=read permission whereas new one asks for Contents=write and && Administration=write
		if i.Permissions != nil && i.Permissions.Contents != nil && strings.EqualFold(*i.Permissions.Contents, "read") {
			permission = adminv1.GithubPermission_GITHUB_PERMISSION_READ
		}

		if i.Permissions != nil && i.Permissions.Contents != nil && i.Permissions.Administration != nil && strings.EqualFold(*i.Permissions.Administration, "write") && strings.EqualFold(*i.Permissions.Contents, "write") {
			permission = adminv1.GithubPermission_GITHUB_PERMISSION_WRITE
		}

		orgInstallationPermission[org.GetLogin()] = permission
	}

	return &adminv1.GetGithubUserStatusResponse{
		HasAccess:                           true,
		GrantAccessUrl:                      s.admin.URLs.GithubConnect(""),
		AccessToken:                         token,
		Account:                             user.GithubUsername,
		Orgs:                                allOrgs,
		UserInstallationPermission:          userInstallationPermission,
		OrganizationInstallationPermissions: orgInstallationPermission,
	}, nil
}

func (s *Server) GetGithubRepoStatus(ctx context.Context, req *adminv1.GetGithubRepoStatusRequest) (*adminv1.GetGithubRepoStatusResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.remote", req.Remote),
	)

	// Backwards compatibility
	req.Remote = normalizeGitRemote(req.Remote)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Check whether we have the access to the repo
	installationID, err := s.admin.GetGithubInstallation(ctx, req.Remote)
	if err != nil {
		if !errors.Is(err, admin.ErrGithubInstallationNotFound) {
			return nil, status.Errorf(codes.InvalidArgument, "failed to check Github access: %s", err.Error())
		}

		// If no access, return instructions for granting access
		grantAccessURL := s.admin.URLs.GithubConnect(req.Remote)

		res := &adminv1.GetGithubRepoStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: grantAccessURL,
		}
		return res, nil
	}

	// we have access need to check if user is a collaborator and has authorised app on their account
	userID := claims.OwnerID()
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// user has not authorized github app
	if user.GithubUsername == "" {
		res := &adminv1.GetGithubRepoStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: s.admin.URLs.GithubAuth(req.Remote),
		}
		return res, nil
	}

	// Get repo info for user and return.
	repository, err := s.admin.LookupGithubRepoForUser(ctx, installationID, req.Remote, user.GithubUsername)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsNotCollaborator) {
			// may be user authorised from another username
			res := &adminv1.GetGithubRepoStatusResponse{
				HasAccess:      false,
				GrantAccessUrl: s.admin.URLs.GithubRetryAuthUI(req.Remote, user.GithubUsername, ""),
			}
			return res, nil
		}
		return nil, err
	}

	res := &adminv1.GetGithubRepoStatusResponse{
		HasAccess:     true,
		DefaultBranch: *repository.DefaultBranch,
	}
	return res, nil
}

func (s *Server) ListGithubUserRepos(ctx context.Context, req *adminv1.ListGithubUserReposRequest) (*adminv1.ListGithubUserReposResponse, error) {
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	userID := claims.OwnerID()
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// user has not authorized github app
	if user.GithubUsername == "" {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	token, err := s.userAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	client := github.NewTokenClient(ctx, token)

	// use a client with user's token to get installations
	repos, err := s.fetchReposForUser(ctx, client)
	if err != nil {
		return nil, err
	}

	return &adminv1.ListGithubUserReposResponse{
		Repos: repos,
	}, nil
}

func (s *Server) ConnectProjectToGithub(ctx context.Context, req *adminv1.ConnectProjectToGithubRequest) (*adminv1.ConnectProjectToGithubResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.remote", req.Remote),
	)

	// Find project
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser || !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update project's github connection")
	}

	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	var branch string
	if proj.ProdBranch != "" {
		branch = proj.ProdBranch
	} else {
		branch = "main"
	}
	token, err := s.createRepo(ctx, req.Remote, branch, user)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if proj.ArchiveAssetID != nil {
		author := &object.Signature{
			Name:  user.GithubUsername,
			Email: user.Email,
		}
		err := s.pushAssetToGit(ctx, *proj.ArchiveAssetID, req.Remote, branch, token, author)
		if err != nil {
			return nil, err
		}
	} else if proj.GitRemote != nil {
		mgdRepoToken, _, err := s.admin.Github.InstallationToken(ctx, *proj.GithubInstallationID, *proj.GithubRepoID)
		if err != nil {
			return nil, err
		}
		err = s.mirrorGitRepo(ctx, *proj.GitRemote, req.Remote, mgdRepoToken, token)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, status.Error(codes.Internal, "invalid project")
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	// TODO : migrate to use service rather than calling UpdateProject directly
	_, err = s.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
		Org:        org.Name,
		Project:    proj.Name,
		ProdBranch: &branch,
		GitRemote:  &req.Remote,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.ConnectProjectToGithubResponse{}, nil
}

func (s *Server) CreateManagedGitRepo(ctx context.Context, req *adminv1.CreateManagedGitRepoRequest) (*adminv1.CreateManagedGitRepoResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.name", req.Name),
	)

	// Find org
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).CreateProjects {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create projects")
	}

	repo, err := s.admin.CreateManagedGitRepo(ctx, org, req.Name, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	id, err := s.admin.Github.ManagedOrgInstallationID()
	if err != nil {
		return nil, err
	}
	token, expiresAt, err := s.admin.Github.InstallationToken(ctx, id, *repo.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.CreateManagedGitRepoResponse{
		Remote:            *repo.CloneURL,
		Username:          "x-access-token",
		Password:          token,
		DefaultBranch:     valOrDefault(repo.DefaultBranch, "main"),
		PasswordExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

// registerGithubEndpoints registers the non-gRPC endpoints for the Github integration.
func (s *Server) registerGithubEndpoints(mux *http.ServeMux) {
	// TODO: Add helper utils to clean this up
	inner := http.NewServeMux()
	observability.MuxHandle(inner, "/github/webhook", http.HandlerFunc(s.githubWebhook))
	observability.MuxHandle(inner, "/github/connect", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("/github/connect"), http.HandlerFunc(s.githubConnect))))
	observability.MuxHandle(inner, "/github/connect/callback", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("/github/connect/callback"), http.HandlerFunc(s.githubConnectCallback))))
	observability.MuxHandle(inner, "/github/auth/login", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("github/auth/login"), http.HandlerFunc(s.githubAuth))))
	observability.MuxHandle(inner, "/github/auth/callback", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("github/auth/callback"), http.HandlerFunc(s.githubAuthCallback))))
	observability.MuxHandle(inner, "/github/post-auth-redirect", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("github/post-auth-redirect"), http.HandlerFunc(s.githubStatus))))
	mux.Handle("/github/", observability.Middleware("admin", s.logger, inner))
}

type githubConnectState struct {
	Remote   string `json:"remote"`
	Redirect string `json:"redirect"`
}

func (g *githubConnectState) isEmpty() bool {
	return g.Remote == "" && g.Redirect == ""
}

// githubConnect starts an installation flow of the Github App.
// It's implemented as a non-gRPC endpoint mounted directly on /github/connect.
// It redirects the user to Github to authorize Rill to access one or more repositories.
// After the Github flow completes, the user is redirected back to githubConnectCallback.
func (s *Server) githubConnect(w http.ResponseWriter, r *http.Request) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		// redirect to the auth site, with a redirect back to here after successful auth.
		s.redirectLogin(w, r)
		return
	}

	query := r.URL.Query()

	remote := query.Get("remote") // May not be set
	redirect, err := url.QueryUnescape(query.Get("redirect"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unescape redirect param: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	// Ignore escape error, param will be omitted.

	// Redirect to Github App for installation
	redirectURL, err := s.githubAppInstallationURL(githubConnectState{
		Remote:   remote,
		Redirect: redirect,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create redirect url: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// githubConnectCallback is called after a Github App authorization flow initiated by githubConnect has completed.
// This call can originate from users who are not logged in in cases like admin user accepting installation request, removing existing installation etc.
// It's implemented as a non-gRPC endpoint mounted directly on /github/connect/callback.
// High level flow:
// User installation
//   - Save user's github username in the users table
//   - verify the user is a collaborator else return unauthorised
//   - verify the user installed the app on the right repo else navigate to retry
//   - navigate to success page
//
// If user requests the app
//   - Save user's github username in the users table
//   - navigate to request page
func (s *Server) githubConnectCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract info from query string
	qry := r.URL.Query()
	setupAction := qry.Get("setup_action")
	if setupAction != "install" && setupAction != "update" && setupAction != "request" {
		http.Error(w, fmt.Sprintf("unexpected setup_action=%q", setupAction), http.StatusBadRequest)
		return
	}

	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		s.redirectLogin(w, r)
		return
	}

	code := qry.Get("code")
	if code == "" {
		if setupAction == "install" || !qry.Has("state") {
			http.Error(w, "unable to verify user's identity", http.StatusInternalServerError)
			return
		}

		remoteURL := qry.Get("state")
		remoteURL = normalizeGitRemote(remoteURL) // Backwards compatibility
		redirectURL := s.admin.URLs.GithubConnectRequestUI(remoteURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// exchange code to get an auth token and create a github client with user auth
	githubClient, githubToken, err := s.userAuthGithubClient(ctx, code)
	if err != nil {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	githubUser, _, err := githubClient.Users.Get(ctx, "")
	if err != nil {
		// todo :: can this throw Requires authentication error ??
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	// save github user name
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		// user is always guaranteed to exist if it reaches here
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, err = s.admin.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
		DisplayName:          user.DisplayName,
		PhotoURL:             user.PhotoURL,
		GithubUsername:       githubUser.GetLogin(),
		GithubToken:          githubToken.AccessToken,
		GithubTokenExpiresOn: &githubToken.Expiry,
		GithubRefreshToken:   githubToken.RefreshToken,
		QuotaSingleuserOrgs:  user.QuotaSingleuserOrgs,
		QuotaTrialOrgs:       user.QuotaTrialOrgs,
		PreferenceTimeZone:   user.PreferenceTimeZone,
	})
	if err != nil {
		s.logger.Error("failed to update user's github username")
	}

	remoteURL := qry.Get("state")
	remoteURL = normalizeGitRemote(remoteURL) // Backwards compatibility
	if remoteURL == "autoclose" {
		// signal from UI flow to autoclose the confirmation dialog
		// TODO: if we ever want more complex signals, we should consider converting this to an object using proto or json
		redirectURL := s.admin.URLs.GithubConnectSuccessUI(true)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	remoteURL, err = url.QueryUnescape(remoteURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse state for remote_url=%s: %s", remoteURL, err.Error()), http.StatusInternalServerError)
		return
	}

	if remoteURL == "" {
		// request without state can come in multiple ways like
		// 	- if user changes app installation directly on the settings page
		//  - if admin user accepts the installation request
		http.Redirect(w, r, s.admin.URLs.GithubConnectSuccessUI(false), http.StatusTemporaryRedirect)
		return
	}

	var state githubConnectState
	err = json.Unmarshal([]byte(remoteURL), &state)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse state for remote_url=%s: %s", remoteURL, err.Error()), http.StatusInternalServerError)
		return
	}

	account, repo, ok := gitutil.SplitGithubRemote(state.Remote)
	if !ok {
		if state.Redirect != "" {
			http.Redirect(w, r, state.Redirect, http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, s.admin.URLs.GithubConnectSuccessUI(false), http.StatusTemporaryRedirect)
		}
		return
	}

	if setupAction == "request" {
		// access requested
		redirectURL := s.admin.URLs.GithubConnectRequestUI(state.Remote)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// install/update setupAction
	// Verify that user installed the app on the right repo and we have access now
	// This needs to come before collaborator check for private repos.
	_, err = s.admin.GetGithubInstallation(ctx, state.Remote)
	if err != nil {
		if !errors.Is(err, admin.ErrGithubInstallationNotFound) {
			http.Error(w, fmt.Sprintf("failed to check github repo status: %s", err), http.StatusInternalServerError)
			return
		}

		// no access
		// Redirect to UI retry page
		redirectURL := s.admin.URLs.GithubConnectRetryUI(state.Remote, state.Redirect)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// verify there is no spoofing and the user is a collaborator to the repo
	isCollaborator, err := s.isCollaborator(ctx, account, repo, githubClient, githubUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify ownership: %s", err), http.StatusUnauthorized)
		return
	}

	if !isCollaborator {
		// Redirect to retry page
		redirectURL := s.admin.URLs.GithubRetryAuthUI(state.Remote, user.GithubUsername, state.Redirect)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Redirect to UI success page or the redirect param
	if state.Redirect != "" {
		http.Redirect(w, r, state.Redirect, http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, s.admin.URLs.GithubConnectSuccessUI(false), http.StatusTemporaryRedirect)
	}
}

// githubAuthLogin starts user authorization of github app.
// In case github app is installed by another user, other users of the repo need to separately authorise github app
// where this flow comes into picture.
// Some implementation details are copied from auth package.
// It's implemented as a non-gRPC endpoint mounted directly on /github/auth/login.
func (s *Server) githubAuth(w http.ResponseWriter, r *http.Request) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		// Redirect to the auth site, with a redirect back to here after successful auth.
		s.redirectLogin(w, r)
		return
	}

	// Generate random state for CSRF
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Get auth cookie
	sess := s.cookies.Get(r, githubcookieName)
	// Set state in cookie
	sess.Values[githubcookieFieldState] = state
	remote := r.URL.Query().Get("remote")
	remote = normalizeGitRemote(remote) // Backwards compatibility
	if remote != "" {
		sess.Values[githubcookieFieldRemote] = remote
	}
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		sess.Values[githubcookieFieldRedirect] = redirect
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
		RedirectURL:  s.admin.URLs.GithubAuthCallback(),
	}
	// Redirect to github login page
	http.Redirect(w, r, oauthConf.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusTemporaryRedirect)
}

// githubAuthCallback is called after a user authorizes github app on their account
// It's implemented as a non-gRPC endpoint mounted directly on /github/auth/callback.
func (s *Server) githubAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		http.Error(w, "unidentified user", http.StatusUnauthorized)
		return
	}

	// Get auth cookie
	sess := s.cookies.Get(r, githubcookieName)
	// Check that random state matches (for CSRF protection)
	qry := r.URL.Query()
	if qry.Get("state") != sess.Values[githubcookieFieldState] {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	delete(sess.Values, githubcookieFieldState)

	// verify user's identity with github
	code := qry.Get("code")
	if code == "" {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	// exchange code to get an auth token and create a github client with user auth
	c, ghToken, err := s.userAuthGithubClient(ctx, code)
	if err != nil {
		// todo :: check for unauthorised user error
		http.Error(w, fmt.Sprintf("internal error %s", err.Error()), http.StatusInternalServerError)
		return
	}

	gitUser, _, err := c.Users.Get(ctx, "")
	if err != nil {
		// todo :: check for unauthorised user error
		http.Error(w, fmt.Sprintf("internal error %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// save the github user name
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		// can this happen ??
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, "unidentified user", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("internal error %s", err.Error()), http.StatusInternalServerError)
		return
	}

	_, err = s.admin.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
		DisplayName:          user.DisplayName,
		PhotoURL:             user.PhotoURL,
		GithubUsername:       gitUser.GetLogin(),
		GithubRefreshToken:   ghToken.RefreshToken,
		GithubToken:          ghToken.AccessToken,
		GithubTokenExpiresOn: &ghToken.Expiry,
		QuotaSingleuserOrgs:  user.QuotaSingleuserOrgs,
		QuotaTrialOrgs:       user.QuotaTrialOrgs,
		PreferenceTimeZone:   user.PreferenceTimeZone,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to save user information %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// if there is a remote set, verify the user is a collaborator the repo
	remote := ""
	if value, ok := sess.Values[githubcookieFieldRemote]; ok {
		remote = value.(string)
	}
	delete(sess.Values, githubcookieFieldRemote)
	remote = normalizeGitRemote(remote) // Backwards compatibility

	if remote == "autoclose" {
		// signal from UI flow to autoclose the confirmation dialog
		// TODO: if we ever want more complex signals, we should consider converting this to an object using proto or json
		redirectURL := s.admin.URLs.GithubConnectSuccessUI(true)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	account, repo, ok := gitutil.SplitGithubRemote(remote)
	if !ok {
		http.Redirect(w, r, s.admin.URLs.GithubConnectSuccessUI(false), http.StatusTemporaryRedirect)
		return
	}

	redirect := ""
	if value, ok := sess.Values[githubcookieFieldRedirect]; ok {
		if strVal, ok := value.(string); ok {
			redirect = strVal
		}
	}
	delete(sess.Values, githubcookieFieldRedirect)

	ok, err = s.isCollaborator(ctx, account, repo, c, gitUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("user identification failed with error %s", err.Error()), http.StatusUnauthorized)
		return
	}

	if !ok {
		// Redirect to retry page
		redirectURL := s.admin.URLs.GithubRetryAuthUI(remote, user.GithubUsername, redirect)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI success page or the redirect param
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, s.admin.URLs.GithubConnectSuccessUI(false), http.StatusTemporaryRedirect)
	}
}

// githubWebhook is called by Github to deliver events about new pushes, pull requests, changes to a repository, etc.
// It's implemented as a non-gRPC endpoint mounted directly on /github/webhook.
// Note that Github webhooks have a timeout of 10 seconds. Webhook processing is moved to the background to prevent timeouts.
func (s *Server) githubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected a POST request", http.StatusBadRequest)
		return
	}

	payload, err := github.ValidatePayload(r, []byte(s.opts.GithubAppWebhookSecret))
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid github payload: %s", err), http.StatusUnauthorized)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid webhook payload: %s", err), http.StatusBadRequest)
		return
	}

	err = s.admin.ProcessGithubEvent(context.Background(), event)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to process event: %s", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// githubStatus is a http wrapper over [GetGithubRepoStatus]/[GetGithubUserStatus] depending upon whether `remote` query is passed.
// It redirects to the grantAccessURL if there is no access.
// It's implemented as a non-gRPC endpoint mounted directly on /github/post-auth-redirect.
func (s *Server) githubStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		s.redirectLogin(w, r)
		return
	}

	var hasAccess bool
	var grantAccessURL string
	remote := r.URL.Query().Get("remote")
	remote = normalizeGitRemote(remote) // Backwards compatibility
	if remote == "" {
		resp, err := s.GetGithubUserStatus(ctx, &adminv1.GetGithubUserStatusRequest{})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch user status: %s", err), http.StatusInternalServerError)
			return
		}
		hasAccess = resp.HasAccess
		grantAccessURL = resp.GrantAccessUrl
	} else {
		resp, err := s.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{Remote: remote})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch github repo status: %s", err), http.StatusInternalServerError)
			return
		}
		hasAccess = resp.HasAccess
		grantAccessURL = resp.GrantAccessUrl
	}

	if hasAccess {
		http.Redirect(w, r, s.admin.URLs.GithubConnectSuccessUI(false), http.StatusTemporaryRedirect)
		return
	}

	redirectURL := s.admin.URLs.GithubConnectUI(grantAccessURL)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (s *Server) userAuthGithubClient(ctx context.Context, code string) (*github.Client, *admin.GithubToken, error) {
	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
	}

	token, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	oauthClient := oauthConf.Client(ctx, token)
	return github.NewClient(oauthClient), &admin.GithubToken{AccessToken: token.AccessToken, Expiry: token.Expiry, RefreshToken: token.RefreshToken}, nil
}

// isCollaborator checks if the user is a collaborator of the repository identified by owner and repo
// client must be authorized with user's auth token
func (s *Server) isCollaborator(ctx context.Context, owner, repo string, client *github.Client, user *github.User) (bool, error) {
	githubUserName := user.GetLogin()
	// repo belongs to the user's personal account
	if owner == githubUserName {
		return true, nil
	}

	// repo belongs to an org
	isCollaborator, resp, err := client.Repositories.IsCollaborator(ctx, owner, repo, user.GetLogin())
	if err != nil {
		// user client does not have access to the repository
		if resp != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
			return false, nil
		}
		return false, err
	}
	return isCollaborator, nil
}

func (s *Server) redirectLogin(w http.ResponseWriter, r *http.Request) {
	redirectURL := s.admin.URLs.AuthLogin(r.URL.RequestURI(), false)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (s *Server) checkGithubRateLimit(route string) middleware.CheckFunc {
	return func(req *http.Request) error {
		claims := auth.GetClaims(req.Context())
		if claims == nil || claims.OwnerType() == auth.OwnerTypeAnon {
			limitKey := ratelimit.AnonLimitKey(route, observability.HTTPPeer(req))
			if err := s.limiter.Limit(req.Context(), limitKey, ratelimit.Sensitive); err != nil {
				if errors.As(err, &ratelimit.QuotaExceededError{}) {
					return httputil.Error(http.StatusTooManyRequests, err)
				}
				return err
			}
		}
		return nil
	}
}

func (s *Server) userAccessToken(ctx context.Context, user *database.User) (string, error) {
	if user.GithubTokenExpiresOn != nil && user.GithubTokenExpiresOn.After(time.Now().Add(5*time.Minute)) {
		return user.GithubToken, nil
	}

	if user.GithubRefreshToken == "" {
		return "", errors.New("refresh token is empty")
	}

	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
	}

	src := oauthConf.TokenSource(ctx, &oauth2.Token{RefreshToken: user.GithubRefreshToken})
	oauthToken, err := src.Token()
	if err != nil {
		return "", err
	}

	// refresh token changes after using it for getting a new token
	_, err = s.admin.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
		DisplayName:          user.DisplayName,
		PhotoURL:             user.PhotoURL,
		GithubUsername:       user.GithubUsername,
		GithubToken:          oauthToken.AccessToken,
		GithubTokenExpiresOn: &oauthToken.Expiry,
		GithubRefreshToken:   oauthToken.RefreshToken,
		QuotaSingleuserOrgs:  user.QuotaSingleuserOrgs,
		QuotaTrialOrgs:       user.QuotaTrialOrgs,
		PreferenceTimeZone:   user.PreferenceTimeZone,
	})
	if err != nil {
		s.logger.Error("failed to update user's github refresh token")
		return "", err
	}

	return oauthToken.AccessToken, nil
}

func (s *Server) fetchReposForUser(ctx context.Context, client *github.Client) ([]*adminv1.ListGithubUserReposResponse_Repo, error) {
	repos := make([]*adminv1.ListGithubUserReposResponse_Repo, 0)
	page := 1

	for {
		installations, httpResp, err := client.Apps.ListUserInstallations(ctx, &github.ListOptions{Page: page, PerPage: 100})
		if err != nil {
			return nil, err
		}

		// TODO: fill in permission

		for _, installation := range installations {
			reposForInst, err := s.fetchReposForInstallation(ctx, client, *installation.ID)
			if err != nil {
				return nil, err
			}
			repos = append(repos, reposForInst...)
		}

		if httpResp.NextPage == 0 {
			break
		}
		page = httpResp.NextPage
	}

	return repos, nil
}

func (s *Server) fetchReposForInstallation(ctx context.Context, client *github.Client, instID int64) ([]*adminv1.ListGithubUserReposResponse_Repo, error) {
	repos := make([]*adminv1.ListGithubUserReposResponse_Repo, 0)
	page := 1

	for {
		reposResp, httpResp, err := client.Apps.ListUserRepos(ctx, instID, &github.ListOptions{Page: page, PerPage: 100})
		if err != nil {
			return nil, err
		}

		for _, repo := range reposResp.Repositories {
			var owner string
			if repo.Owner != nil {
				owner = fromStringPtr(repo.Owner.Login)
			}
			var branch string
			if repo.DefaultBranch != nil {
				branch = fromStringPtr(repo.DefaultBranch)
			} else {
				branch = fromStringPtr(repo.MasterBranch)
			}
			repos = append(repos, &adminv1.ListGithubUserReposResponse_Repo{
				Name:          fromStringPtr(repo.Name),
				Owner:         owner,
				Description:   fromStringPtr(repo.Description),
				Remote:        fromStringPtr(repo.CloneURL),
				DefaultBranch: branch,
			})
		}

		if httpResp.NextPage == 0 {
			break
		}
		page = httpResp.NextPage
	}

	return repos, nil
}

func (s *Server) createRepo(ctx context.Context, remote, branch string, user *database.User) (string, error) {
	org, repo, ok := gitutil.SplitGithubRemote(remote)
	if !ok {
		return "", status.Error(codes.InvalidArgument, fmt.Sprintf("invalid remote: %q", remote))
	}

	var err error
	var token, ghAcct string
	var client *github.Client
	if org == user.GithubUsername {
		// if expectation is to create in user's personal account then we need to use user access token
		token, err = s.userAccessToken(ctx, user)
		if err != nil {
			return "", err
		}
		// We need to pass empty org if the org to be created in is same as the authenticated user.
		ghAcct = ""
		client = github.NewTokenClient(ctx, token)
	} else {
		// get the installation access token for that org
		token, _, err = s.admin.Github.InstallationTokenForOrg(ctx, org)
		if err != nil {
			return "", err
		}
		ghAcct = org
		client = github.NewTokenClient(ctx, token)
		// check user should be a member of the org to create a repo
		ok, _, err := client.Organizations.IsMember(ctx, ghAcct, user.GithubUsername)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", status.Errorf(codes.PermissionDenied, "user is not a member of the organization %q", org)
		}
	}

	_, _, err = client.Repositories.Create(ctx, ghAcct, &github.Repository{
		Name:          &repo,
		DefaultBranch: &branch,
		Private:       github.Ptr(true),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create repo: %w", err)
	}

	// github.Repositories.Create returns before actually creating the repo. So do an exponential backoff check
	err = retrier.New(retrier.ExponentialBackoff(createRetries, time.Second), nil).RunCtx(ctx, func(ctx context.Context) error {
		_, _, err := client.Repositories.Get(ctx, org, repo)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to verify repo creation: %w", err)
	}

	return token, nil
}

func (s *Server) mirrorGitRepo(ctx context.Context, srcGitRemote, destGitRemote, srcToken, destToken string) error {
	gitPath, err := os.MkdirTemp(os.TempDir(), "projects")
	if err != nil {
		return err
	}
	defer os.RemoveAll(gitPath)

	repo, err := git.PlainCloneContext(ctx, gitPath, false, &git.CloneOptions{
		URL:  srcGitRemote,
		Auth: &githttp.BasicAuth{Username: "x-access-token", Password: srcToken},
	})
	if err != nil {
		return fmt.Errorf("failed to clone git repo: %w", err)
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "dest",
		URLs: []string{destGitRemote},
	})
	if err != nil {
		return fmt.Errorf("failed to create remote: %w", err)
	}

	// Push everything (all refs, like git push --mirror)
	err = repo.PushContext(ctx, &git.PushOptions{
		Auth:       &githttp.BasicAuth{Username: "x-access-token", Password: destToken},
		RemoteName: "dest",
		RefSpecs: []config.RefSpec{
			"+refs/*:refs/*", // force-push all refs
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) pushAssetToGit(ctx context.Context, assetID, remote, branch, token string, author *object.Signature) error {
	asset, err := s.admin.DB.FindAsset(ctx, assetID)
	if err != nil {
		return err
	}

	downloadURL, err := s.generateSignedDownloadURL(asset)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	downloadDir, err := os.MkdirTemp(os.TempDir(), "extracted_archives")
	if err != nil {
		return err
	}
	defer os.RemoveAll(downloadDir)
	downloadDst := filepath.Join(downloadDir, "zipped_repo.tar.gz")

	projPath := filepath.Join(downloadDir, "projects")
	err = archive.Download(ctx, downloadURL, downloadDst, projPath, false, true)
	if err != nil {
		return err
	}

	config := &cligitutil.Config{
		Remote:        remote,
		Username:      "x-access-token",
		Password:      token,
		DefaultBranch: branch,
	}
	return cligitutil.CommitAndPush(ctx, projPath, config, "", author)
}

func (s *Server) githubAppInstallationURL(state githubConnectState) (string, error) {
	res := fmt.Sprintf("https://github.com/apps/%s/installations/new", s.opts.GithubAppName)
	if state.isEmpty() {
		return res, nil
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return res, fmt.Errorf("failed to marshal github app installation state: %w", err)
	}

	return urlutil.MustWithQuery(res, map[string]string{"state": string(stateJSON)}), nil
}

func fromStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// normalizeGitRemote adds a .git suffix to the Git remote URL if it doesn't already have one.
// If it's not a Github URL, it returns the string as is.
// This is for backwards compatibility with old CLIs that sent Github HTML URLs instead of Github remote URLs.
func normalizeGitRemote(remote string) string {
	if !strings.HasPrefix(remote, "https://github.com") {
		return remote // Not a Github remote, return as is
	}
	if strings.HasSuffix(remote, ".git") {
		return remote
	}
	return remote + ".git"
}
