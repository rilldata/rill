package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	githubcookieName        = "github_auth"
	githubcookieFieldState  = "github_state"
	githubcookieFieldRemote = "github_remote"
)

func (s *Server) GetGithubRepoStatus(ctx context.Context, req *adminv1.GetGithubRepoStatusRequest) (*adminv1.GetGithubRepoStatusResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Check whether we have the access to the repo
	installationID, err := s.admin.GetGithubInstallation(ctx, req.GithubUrl)
	if err != nil {
		if !errors.Is(err, admin.ErrGithubInstallationNotFound) {
			return nil, status.Errorf(codes.InvalidArgument, "failed to check Github access: %s", err.Error())
		}

		// If no access, return instructions for granting access
		grantAccessURL, err := urlutil.WithQuery(s.urls.githubConnect, map[string]string{"remote": req.GithubUrl})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create redirect URL: %s", err)
		}

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
		redirectURL, err := urlutil.WithQuery(s.urls.githubAuth, map[string]string{"remote": req.GithubUrl})
		if err != nil {
			return nil, err
		}

		res := &adminv1.GetGithubRepoStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: redirectURL,
		}
		return res, nil
	}

	// Get repo info for user and return.
	repository, err := s.admin.LookupGithubRepoForUser(ctx, installationID, req.GithubUrl, user.GithubUsername)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsNotCollaborator) {
			// may be user authorised from another username
			redirectURL, err := urlutil.WithQuery(s.urls.githubAuthRetry, map[string]string{"remote": req.GithubUrl, "githubUsername": user.GithubUsername})
			if err != nil {
				return nil, err
			}

			res := &adminv1.GetGithubRepoStatusResponse{
				HasAccess:      false,
				GrantAccessUrl: redirectURL,
			}
			return res, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &adminv1.GetGithubRepoStatusResponse{
		HasAccess:     true,
		DefaultBranch: *repository.DefaultBranch,
	}
	return res, nil
}

// registerGithubEndpoints registers the non-gRPC endpoints for the Github integration.
func (s *Server) registerGithubEndpoints(mux *http.ServeMux) {
	// TODO: Add helper utils to clean this up
	inner := http.NewServeMux()
	inner.Handle("/github/webhook", otelhttp.WithRouteTag("/github/webhook", http.HandlerFunc(s.githubWebhook)))
	inner.Handle("/github/connect", otelhttp.WithRouteTag("/github/connect", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.githubConnect))))
	inner.Handle("/github/connect/callback", otelhttp.WithRouteTag("/github/connect/callback", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.githubConnectCallback))))
	inner.Handle("/github/auth/login", otelhttp.WithRouteTag("github/auth/login", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.githubAuthLogin))))
	inner.Handle("/github/auth/callback", otelhttp.WithRouteTag("github/auth/callback", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.githubAuthCallback))))
	inner.Handle("/github/repo_status", otelhttp.WithRouteTag("github/repo_status", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.githubRepoStatus))))
	mux.Handle("/github/", observability.Middleware("admin", s.logger, inner))
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
	remote := query.Get("remote")
	if remote == "" {
		http.Redirect(w, r, s.urls.githubAppInstallation, http.StatusTemporaryRedirect)
		return
	}

	redirectURL, err := urlutil.WithQuery(s.urls.githubAppInstallation, map[string]string{"state": remote})
	if err != nil {
		http.Error(w, "failed to generate URL", http.StatusInternalServerError)
		return
	}

	// Redirect to Github App for installation
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
		http.Error(w, "only authenticated users can connect to github", http.StatusUnauthorized)
		return
	}

	code := qry.Get("code")
	if code == "" {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	// exchange code to get an auth token and create a github client with user auth
	githubClient, err := s.userAuthGithubClient(ctx, code)
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

	_, err = s.admin.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
		DisplayName:    user.DisplayName,
		PhotoURL:       user.PhotoURL,
		GithubUsername: githubUser.GetLogin(),
	})
	if err != nil {
		s.logger.Error("failed to update user's github username")
	}

	remoteURL := qry.Get("state")
	account, repo, ok := gitutil.SplitGithubURL(remoteURL)
	if !ok {
		// request without state can come in multiple ways like
		// 	- if user changes app installation directly on the settings page
		//  - if admin user accepts the installation request
		http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
		return
	}

	if setupAction == "request" {
		// access requested
		redirectURL, err := urlutil.WithQuery(s.urls.githubConnectRequest, map[string]string{"remote": remoteURL})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create connect request url: %s", err.Error()), http.StatusInternalServerError)
			return
		}

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
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	// install/update setupAction
	// Verify that user installed the app on the right repo and we have access now
	_, err = s.admin.GetGithubInstallation(ctx, remoteURL)
	if err != nil {
		if !errors.Is(err, admin.ErrGithubInstallationNotFound) {
			http.Error(w, fmt.Sprintf("failed to check github repo status: %s", err), http.StatusInternalServerError)
			return
		}

		// no access
		// Redirect to UI retry page
		redirectURL, err := urlutil.WithQuery(s.urls.githubConnectRetry, map[string]string{"remote": remoteURL})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create retry request url: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Redirect to UI success page
	http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
}

// githubAuthLogin starts user authorization of github app.
// In case github app is installed by another user, other users of the repo need to separately authorise github app
// where this flow comes into picture.
// Some implementation details are copied from auth package.
// It's implemented as a non-gRPC endpoint mounted directly on /github/auth/login.
func (s *Server) githubAuthLogin(w http.ResponseWriter, r *http.Request) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		// Redirect to the auth site, with a redirect back to here after successful auth.
		redirectURL, err := urlutil.WithQuery(s.urls.authLogin, map[string]string{"redirect": r.URL.RequestURI()})
		if err != nil {
			http.Error(w, "failed to generate URL", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
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
	sess, err := s.cookies.Get(r, githubcookieName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	// Set state in cookie
	sess.Values[githubcookieFieldState] = state
	remote := r.URL.Query().Get("remote")
	if remote != "" {
		sess.Values[githubcookieFieldRemote] = remote
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
		RedirectURL:  s.urls.githubAuthCallback,
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
	sess, err := s.cookies.Get(r, githubcookieName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

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
	c, err := s.userAuthGithubClient(ctx, code)
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
		DisplayName:    user.DisplayName,
		PhotoURL:       user.PhotoURL,
		GithubUsername: gitUser.GetLogin(),
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

	account, repo, ok := gitutil.SplitGithubURL(remote)
	if !ok {
		http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
		return
	}

	ok, err = s.isCollaborator(ctx, account, repo, c, gitUser)
	if err != nil {
		http.Error(w, "unidentified user", http.StatusUnauthorized)
		return
	}

	if !ok {
		redirectURL, err := urlutil.WithQuery(s.urls.githubAuthRetry, map[string]string{"remote": remote, "githubUsername": user.GithubUsername})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to retry page
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI success page
	http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
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

func (s *Server) githubRepoStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		s.redirectLogin(w, r)
		return
	}

	githubURL := r.URL.Query().Get("remote")
	// Check whether we have the access to the repo
	installationID, err := s.admin.GetGithubInstallation(ctx, githubURL)
	if err != nil {
		if !errors.Is(err, admin.ErrGithubInstallationNotFound) {
			http.Error(w, fmt.Sprintf("failed to check Github access: %s", err), http.StatusBadRequest)
			return
		}

		// If no access, redirect to github connect page
		grantAccessURL, err := urlutil.WithQuery(s.urls.githubConnect, map[string]string{"remote": githubURL})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
			return
		}

		redirectURL, err := urlutil.WithQuery(s.urls.githubConnectUI, map[string]string{"redirect_url": grantAccessURL})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// we have access need to check if user is a collaborator and has authorised app on their account
	userID := claims.OwnerID()
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to find user: %s", err), http.StatusInternalServerError)
		return
	}

	// user has not authorized github app
	if user.GithubUsername == "" {
		grantAccessURL, err := urlutil.WithQuery(s.urls.githubAuth, map[string]string{"remote": githubURL})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
			return
		}

		redirectURL, err := urlutil.WithQuery(s.urls.githubConnectUI, map[string]string{"redirect_url": grantAccessURL})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Get repo info for user and return.
	_, err = s.admin.LookupGithubRepoForUser(ctx, installationID, githubURL, user.GithubUsername)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsNotCollaborator) {
			// may be user authorised from another username
			retryURL, err := urlutil.WithQuery(s.urls.githubAuthRetry, map[string]string{"remote": githubURL, "githubUsername": user.GithubUsername})
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
				return
			}

			redirectURL, err := urlutil.WithQuery(s.urls.githubConnectUI, map[string]string{"redirect_url": retryURL})
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}

		http.Error(w, fmt.Sprintf("failed to lookup repo for user: %s", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
}

func (s *Server) userAuthGithubClient(ctx context.Context, code string) (*github.Client, error) {
	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
	}

	token, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	oauthClient := oauthConf.Client(ctx, token)
	return github.NewClient(oauthClient), nil
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
		// user client will not have access to the repository
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return false, nil
		}
		return false, err
	}
	return isCollaborator, nil
}

func (s *Server) redirectLogin(w http.ResponseWriter, r *http.Request) {
	redirectURL, err := urlutil.WithQuery(s.urls.authLogin, map[string]string{"redirect": r.URL.RequestURI()})
	if err != nil {
		http.Error(w, "failed to generate URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
