package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/server/auth"
)

// githubConnect starts an installation flow of the Github App.
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
	url := fmt.Sprintf("https://github.com/apps/%s/installations/new", s.opts.GithubAppName)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// githubConnectCallback is called after a Github App authorization flow initiated by githubConnect has completed.
func (s *Server) githubConnectCallback(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// TODO: Enable user authorization and verify user per https://roadie.io/blog/avoid-leaking-github-org-data/

	// Extract info from query string
	qry := r.URL.Query()
	setupAction := qry.Get("setup_action")
	if setupAction != "install" { // TODO: Can also be update, request
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
	err = s.admin.TrackGithubInstallation(r.Context(), claims.OwnerID(), int64(installationID))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to track github install: %s", err), http.StatusInternalServerError)
		return
	}

	// TODO: Redirect to UI success page
	// http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
	w.WriteHeader(http.StatusOK)
}

// githubWebhook is called by Github to deliver events about new pushes, pull requests, changes to a repository, etc.
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
