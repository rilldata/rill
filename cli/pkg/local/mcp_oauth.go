package local

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/mcp"
	"go.uber.org/zap"
)

// This file implements the browser-based OAuth flow for external MCP servers in Rill Developer.
// The flow mirrors the /auth handlers: /mcp/oauth/connect starts it and /mcp/oauth/callback completes it.
// Tokens are persisted by the MCP driver in the connector's local data dir; there is a single local user.

// mcpOAuthFlowTTL is how long a pending authorization flow stays valid.
const mcpOAuthFlowTTL = 15 * time.Minute

// mcpOAuthFlow is a pending browser-based authorization flow, keyed by its OAuth state parameter.
type mcpOAuthFlow struct {
	connector   string
	displayName string
	verifier    string
	redirect    string
	expires     time.Time
}

// mcpOAuthFlows tracks pending flows.
type mcpOAuthFlows struct {
	mu    sync.Mutex
	flows map[string]*mcpOAuthFlow
}

func (f *mcpOAuthFlows) add(state string, flow *mcpOAuthFlow) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.flows == nil {
		f.flows = make(map[string]*mcpOAuthFlow)
	}
	for k, v := range f.flows {
		if time.Now().After(v.expires) {
			delete(f.flows, k)
		}
	}
	f.flows[state] = flow
}

func (f *mcpOAuthFlows) take(state string) (*mcpOAuthFlow, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	flow, ok := f.flows[state]
	if !ok {
		return nil, false
	}
	delete(f.flows, state)
	if time.Now().After(flow.expires) {
		return nil, false
	}
	return flow, true
}

// mcpOAuthConnectHandler starts the authorization flow for an MCP connector and redirects the browser to the authorization server.
// Query params: "connector" (required) and "redirect" (optional URL to return to when the page was not opened as a popup).
func (s *Server) mcpOAuthConnectHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		connector := r.URL.Query().Get("connector")
		if connector == "" {
			http.Error(w, "missing connector", http.StatusBadRequest)
			return
		}
		redirect := s.safeLocalRedirect(r.URL.Query().Get("redirect"), s.app.localURL)

		handle, release, err := s.app.Runtime.AcquireHandle(ctx, s.app.Instance.ID, connector)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to open connector %q: %s", connector, err), http.StatusBadRequest)
			return
		}
		defer release()
		manager, ok := handle.(mcp.OAuthManager)
		if !ok || !manager.OAuthEnabled() {
			http.Error(w, fmt.Sprintf("connector %q is not an MCP server with 'auth: oauth'", connector), http.StatusBadRequest)
			return
		}

		state, err := mcp.GenerateOAuthState()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
			return
		}
		authURL, verifier, err := manager.BeginOAuth(ctx, s.mcpOAuthRedirectURL(), state)
		if err != nil {
			s.logger.Warn("failed to start MCP OAuth flow", zap.String("connector", connector), zap.Error(err))
			http.Error(w, fmt.Sprintf("failed to start authorization: %s", err), http.StatusBadGateway)
			return
		}

		s.app.mcpOAuthFlows.add(state, &mcpOAuthFlow{
			connector:   connector,
			displayName: mcpDisplayName(handle.Config(), connector),
			verifier:    verifier,
			redirect:    redirect,
			expires:     time.Now().Add(mcpOAuthFlowTTL),
		})
		http.Redirect(w, r, authURL, http.StatusFound)
	})
}

// mcpOAuthCallbackHandler completes the authorization flow: it exchanges the code for a token, persists it,
// and shows a page that notifies the chat (the opener window) and closes itself.
func (s *Server) mcpOAuthCallbackHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		qry := r.URL.Query()
		flow, ok := s.app.mcpOAuthFlows.take(qry.Get("state"))
		if !ok {
			http.Error(w, "invalid or expired authorization flow; start the flow again", http.StatusBadRequest)
			return
		}

		render := func(errMsg string) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(mcp.OAuthCallbackPage(flow.connector, flow.displayName, errMsg, flow.redirect)))
		}

		if errCode := qry.Get("error"); errCode != "" {
			msg := errCode
			if desc := qry.Get("error_description"); desc != "" {
				msg += ": " + desc
			}
			render(msg)
			return
		}
		code := qry.Get("code")
		if code == "" {
			render("the authorization server did not return a code")
			return
		}

		handle, release, err := s.app.Runtime.AcquireHandle(ctx, s.app.Instance.ID, flow.connector)
		if err != nil {
			render(fmt.Sprintf("failed to open connector %q: %s", flow.connector, err))
			return
		}
		defer release()
		manager, ok := handle.(mcp.OAuthManager)
		if !ok {
			render(fmt.Sprintf("connector %q is no longer an MCP server", flow.connector))
			return
		}
		err = manager.CompleteOAuth(ctx, code, flow.verifier)
		if err != nil {
			s.logger.Warn("failed to complete MCP OAuth flow", zap.String("connector", flow.connector), zap.Error(err))
			render(err.Error())
			return
		}

		// Reconcile the connector so its tools show up in the connector state.
		ctrl, err := s.app.Runtime.Controller(ctx, s.app.Instance.ID)
		if err == nil {
			err = ctrl.Reconcile(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindConnector, Name: flow.connector})
		}
		if err != nil {
			s.logger.Warn("failed to reconcile MCP connector after authorization", zap.String("connector", flow.connector), zap.Error(err))
		}

		render("")
	})
}

// mcpOAuthRedirectURL is the redirect URI registered with authorization servers.
// It is derived from the local server's URL, so it follows the port Rill Developer actually runs on.
func (s *Server) mcpOAuthRedirectURL() string {
	return s.app.localURL + "/mcp/oauth/callback"
}

// safeLocalRedirect returns redirect if it points back to Rill Developer, otherwise fallback.
// Accepted are relative paths, loopback origins (the UI runs on another port in development) and the configured allowed origins.
// The callback page links to the URL, so javascript: and third-party URLs must never pass.
func (s *Server) safeLocalRedirect(redirect, fallback string) string {
	if redirect == "" {
		return fallback
	}
	u, err := url.Parse(redirect)
	if err != nil {
		return fallback
	}
	if u.Scheme == "" && u.Host == "" {
		if strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//") {
			return redirect
		}
		return fallback
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fallback
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return redirect
	}
	if local, err := url.Parse(s.app.localURL); err == nil && strings.EqualFold(local.Host, u.Host) {
		return redirect
	}
	origin := u.Scheme + "://" + u.Host
	for _, allowed := range s.app.allowedOrigins {
		if strings.EqualFold(strings.TrimSuffix(allowed, "/"), origin) {
			return redirect
		}
	}
	return fallback
}

// mcpDisplayName returns the display name configured on an MCP connector, falling back to its name.
func mcpDisplayName(config map[string]any, connector string) string {
	if name, ok := config["display_name"].(string); ok && name != "" {
		return name
	}
	return connector
}
