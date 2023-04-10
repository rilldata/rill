package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/go-github/v50/github"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetGithubRepoStatus(ctx context.Context, req *adminv1.GetGithubRepoStatusRequest) (*adminv1.GetGithubRepoStatusResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Check whether user has granted access
	installationID, ok, err := s.admin.GetUserGithubInstallation(ctx, claims.OwnerID(), req.GithubUrl)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to check Github access: %s", err.Error())
	}

	// If the user has not granted access, return instructions for granting access
	if !ok {
		grantAccessURL, err := url.JoinPath(s.opts.ExternalURL, "/github/connect")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create redirect URL: %s", err)
		}

		res := &adminv1.GetGithubRepoStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: grantAccessURL,
		}
		return res, nil
	}

	// The user has granted access. Get repo info and return.
	repo, err := s.admin.LookupGithubRepo(ctx, installationID, req.GithubUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &adminv1.GetGithubRepoStatusResponse{
		HasAccess:     true,
		DefaultBranch: *repo.DefaultBranch,
	}
	return res, nil
}

// registerGithubEndpoints registers the non-gRPC endpoints for the Github integration.
func (s *Server) registerGithubEndpoints(mux *gateway.ServeMux) error {
	err := mux.HandlePath("POST", "/github/webhook", s.githubWebhook)
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/github/connect", s.authenticator.HTTPMiddleware(s.githubConnect))
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/github/connect/callback", s.authenticator.HTTPMiddleware(s.githubConnectCallback))
	if err != nil {
		return err
	}

	return nil
}

// githubConnect starts an installation flow of the Github App.
// It's implemented as a non-gRPC endpoint mounted directly on /github/connect.
// It redirects the user to Github to authorize Rill to access one or more repositories.
// After the Github flow completes, the user is redirected back to githubConnectCallback.
func (s *Server) githubConnect(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		// TODO: It should redirect to the auth site, with a redirect back to here after successful auth.
		http.Error(w, "only authenticated users can connect to github", http.StatusUnauthorized)
		return
	}

	// NOTE: If needed, we can add a `state` query parameter that will be passed through to githubConnectCallback.

	// Redirect to Github App for installation
	redirectURL := fmt.Sprintf("https://github.com/apps/%s/installations/new", s.opts.GithubAppName)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// githubConnectCallback is called after a Github App authorization flow initiated by githubConnect has completed.
// It's implemented as a non-gRPC endpoint mounted directly on /github/connect/callback.
func (s *Server) githubConnectCallback(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// TODO: Enable user authorization and verify user per https://roadie.io/blog/avoid-leaking-github-org-data/

	// Extract info from query string
	qry := r.URL.Query()
	setupAction := qry.Get("setup_action")
	if setupAction != "install" && setupAction != "update" { // TODO: Also handle "request"
		http.Error(w, fmt.Sprintf("unexpected setup_action=%q", setupAction), http.StatusBadRequest)
		return
	}
	installationIDStr := qry.Get("installation_id")
	installationID, err := strconv.Atoi(installationIDStr)
	if err != nil || installationID == 0 {
		http.Error(w, fmt.Sprintf("unexpected installation_id=%q", installationIDStr), http.StatusBadRequest)
		return
	}

	// Check there's an authenticated user (this should always be the case for flows initiated by githubConnect)
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		http.Error(w, "only authenticated users can connect to github", http.StatusUnauthorized)
		return
	}

	// Associate the user with the installation
	err = s.admin.ProcessUserGithubInstallation(r.Context(), claims.OwnerID(), int64(installationID))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to track github install: %s", err), http.StatusInternalServerError)
		return
	}

	// Redirect to UI success page
	redirectURL, err := url.JoinPath(s.opts.FrontendURL, "/-/github/connect/success")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// githubWebhook is called by Github to deliver events about new pushes, pull requests, changes to a repository, etc.
// It's implemented as a non-gRPC endpoint mounted directly on /github/webhook.
// Note that Github webhooks have a timeout of 10 seconds. Webhook processing is moved to the background to prevent timeouts.
func (s *Server) githubWebhook(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
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
