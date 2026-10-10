package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/client"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/mcp"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file implements per-user authorization of external MCP servers (connectors with driver "mcp" and auth "oauth") in Rill Cloud.
//
// The browser-based flow runs here because the admin service owns user sessions:
// /mcp/oauth/connect discovers the server's authorization server, registers a client once per project, and redirects the user;
// /mcp/oauth/callback exchanges the code and stores the token encrypted per (project, connector, user).
// Runtimes never run the flow; they fetch a user's access token through the deployment-authenticated GetMCPOAuthToken RPC.

const (
	mcpOAuthCookieName = "mcp_oauth"
	// mcpOAuthFlowTTL is how long a pending authorization flow stays valid.
	mcpOAuthFlowTTL = 15 * time.Minute
	// mcpOAuthTokenRefreshLeeway is how long before expiry a token is refreshed when a runtime asks for it.
	mcpOAuthTokenRefreshLeeway = 60 * time.Second
	// mcpOAuthHTTPTimeout bounds discovery and token requests to authorization servers.
	mcpOAuthHTTPTimeout = 15 * time.Second
)

// mcpOAuthFlow is the state of a pending browser-based authorization flow, persisted in an encrypted cookie between the two requests.
type mcpOAuthFlow struct {
	State       string `json:"state"`
	Org         string `json:"org"`
	Project     string `json:"project"`
	Connector   string `json:"connector"`
	DisplayName string `json:"display_name"`
	Verifier    string `json:"verifier"`
	Redirect    string `json:"redirect"`
	Expires     int64  `json:"expires"`
}

// registerMCPOAuthEndpoints registers the browser-facing endpoints for authorizing external MCP servers.
func (s *Server) registerMCPOAuthEndpoints(mux *http.ServeMux) {
	inner := http.NewServeMux()
	observability.MuxHandle(inner, "/mcp/oauth/connect", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.mcpOAuthConnect)))
	observability.MuxHandle(inner, "/mcp/oauth/callback", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.mcpOAuthCallback)))
	mux.Handle("/mcp/", observability.Middleware("admin", s.logger, inner))
}

// mcpOAuthConnect starts an authorization flow for an MCP connector on behalf of the authenticated user.
// Query params: "org", "project" and "connector" (required); "branch" (optional, for edit-mode deployments);
// "redirect" (optional URL to return to when the page was not opened as a popup).
func (s *Server) mcpOAuthConnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		s.redirectLogin(w, r)
		return
	}

	qry := r.URL.Query()
	orgName, projectName, connector := qry.Get("org"), qry.Get("project"), qry.Get("connector")
	if orgName == "" || projectName == "" || connector == "" {
		http.Error(w, "missing org, project or connector", http.StatusBadRequest)
		return
	}
	redirect := qry.Get("redirect")
	if redirect == "" {
		redirect = s.admin.URLs.Project(orgName, projectName)
	} else if err := s.mcpOAuthValidateRedirect(ctx, redirect); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	proj, depl, permissions, err := s.mcpOAuthResolveDeployment(ctx, claims, orgName, projectName, qry.Get("branch"))
	if err != nil {
		writeMCPOAuthError(w, err)
		return
	}

	// Check that the user may use the server (its security policy is evaluated by the runtime for the user),
	// so Rill never collects a token for a user who cannot use the server.
	server, err := s.mcpOAuthUserServer(ctx, proj, depl, permissions, connector)
	if err != nil {
		writeMCPOAuthError(w, err)
		return
	}

	// Resolve the connector's OAuth configuration (pre-registered client and scopes), which may reference project secrets.
	cfg, err := s.mcpOAuthConnectorConfig(ctx, depl, connector)
	if err != nil {
		writeMCPOAuthError(w, err)
		return
	}

	oauthClient, err := s.mcpOAuthClient(ctx, proj, connector, cfg)
	if err != nil {
		s.logger.Warn("failed to prepare MCP OAuth client", zap.String("project_id", proj.ID), zap.String("connector", connector), zap.Error(err), observability.ZapCtx(ctx))
		writeMCPOAuthError(w, err)
		return
	}

	state, err := mcp.GenerateOAuthState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}
	verifier := mcp.GenerateOAuthVerifier()

	flow := &mcpOAuthFlow{
		State:       state,
		Org:         orgName,
		Project:     projectName,
		Connector:   connector,
		DisplayName: server.DisplayName,
		Verifier:    verifier,
		Redirect:    redirect,
		Expires:     time.Now().Add(mcpOAuthFlowTTL).Unix(),
	}
	flowJSON, err := json.Marshal(flow)
	if err != nil {
		http.Error(w, "failed to encode flow", http.StatusInternalServerError)
		return
	}
	sess := s.cookies.Get(r, mcpOAuthCookieName)
	sess.Values["flow"] = string(flowJSON)
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, oauthClient.AuthCodeURL(state, verifier), http.StatusFound)
}

// mcpOAuthCallback completes an authorization flow: it exchanges the code for a token, stores it for the user,
// and renders a page that notifies the chat (the opener window) and closes itself.
func (s *Server) mcpOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		http.Error(w, "unidentified user", http.StatusUnauthorized)
		return
	}

	// Load and clear the pending flow
	sess := s.cookies.Get(r, mcpOAuthCookieName)
	flowJSON, _ := sess.Values["flow"].(string)
	delete(sess.Values, "flow")
	_ = sess.Save(r, w)
	flow := &mcpOAuthFlow{}
	if flowJSON == "" || json.Unmarshal([]byte(flowJSON), flow) != nil || time.Now().Unix() > flow.Expires {
		http.Error(w, "invalid or expired authorization flow; start the flow again", http.StatusBadRequest)
		return
	}
	qry := r.URL.Query()
	if qry.Get("state") == "" || qry.Get("state") != flow.State {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	render := func(errMsg string) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(mcp.OAuthCallbackPage(flow.Connector, flow.DisplayName, errMsg, flow.Redirect)))
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

	proj, err := s.admin.DB.FindProjectByName(ctx, flow.Org, flow.Project)
	if err != nil {
		render(fmt.Sprintf("project not found: %s", err))
		return
	}
	row, err := s.admin.DB.FindMCPOAuthClient(ctx, proj.ID, flow.Connector)
	if err != nil {
		render("no OAuth client is registered for the server; start the flow again")
		return
	}
	oauthClient := &mcp.OAuthClient{}
	if err := json.Unmarshal(row.Client, oauthClient); err != nil {
		render("the registered OAuth client is invalid; start the flow again")
		return
	}

	tok, err := oauthClient.Exchange(ctx, code, flow.Verifier, mcp.NewHTTPClient(false, mcpOAuthHTTPTimeout))
	if err != nil {
		s.logger.Warn("failed to exchange MCP OAuth code", zap.String("project_id", proj.ID), zap.String("connector", flow.Connector), zap.Error(err), observability.ZapCtx(ctx))
		render(err.Error())
		return
	}
	tokJSON, err := json.Marshal(tok)
	if err != nil {
		render("failed to encode token")
		return
	}
	_, err = s.admin.DB.UpsertMCPOAuthToken(ctx, &database.UpsertMCPOAuthTokenOptions{
		ProjectID: proj.ID,
		Connector: flow.Connector,
		UserID:    claims.OwnerID(),
		Token:     tokJSON,
	})
	if err != nil {
		render(fmt.Sprintf("failed to store token: %s", err))
		return
	}

	render("")
}

// GetMCPOAuthToken implements AIService. It can only be called by deployments, for users of their own project.
func (s *Server) GetMCPOAuthToken(ctx context.Context, req *adminv1.GetMCPOAuthTokenRequest) (*adminv1.GetMCPOAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.connector", req.Connector),
		attribute.String("args.user_id", req.UserId),
		attribute.Bool("args.force_refresh", req.ForceRefresh),
	)

	depl, err := s.mcpOAuthCallerDeployment(ctx)
	if err != nil {
		return nil, err
	}
	if req.Connector == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "connector and user_id are required")
	}

	row, err := s.admin.DB.FindMCPOAuthToken(ctx, depl.ProjectID, req.Connector, req.UserId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "the user has not authorized the connector")
		}
		return nil, err
	}
	tok := &oauth2.Token{}
	if err := json.Unmarshal(row.Token, tok); err != nil {
		return nil, fmt.Errorf("stored token is invalid: %w", err)
	}

	expiring := !tok.Expiry.IsZero() && time.Until(tok.Expiry) < mcpOAuthTokenRefreshLeeway
	if req.ForceRefresh || expiring {
		clientRow, err := s.admin.DB.FindMCPOAuthClient(ctx, depl.ProjectID, req.Connector)
		if err != nil {
			return nil, fmt.Errorf("failed to find the connector's OAuth client: %w", err)
		}
		oauthClient := &mcp.OAuthClient{}
		if err := json.Unmarshal(clientRow.Client, oauthClient); err != nil {
			return nil, fmt.Errorf("stored OAuth client is invalid: %w", err)
		}
		newTok, err := oauthClient.Refresh(ctx, tok, mcp.NewHTTPClient(false, mcpOAuthHTTPTimeout))
		if err != nil {
			if errors.Is(err, drivers.ErrMCPAuthorizationRequired) {
				// The authorization is gone for good; forget it so the user is asked to connect again.
				_ = s.admin.DB.DeleteMCPOAuthToken(ctx, depl.ProjectID, req.Connector, req.UserId)
				return nil, status.Error(codes.NotFound, "the user's authorization has expired")
			}
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		tokJSON, err := json.Marshal(newTok)
		if err != nil {
			return nil, err
		}
		_, err = s.admin.DB.UpsertMCPOAuthToken(ctx, &database.UpsertMCPOAuthTokenOptions{
			ProjectID: depl.ProjectID,
			Connector: req.Connector,
			UserID:    req.UserId,
			Token:     tokJSON,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to store refreshed token: %w", err)
		}
		tok = newTok
	}

	res := &adminv1.GetMCPOAuthTokenResponse{AccessToken: tok.AccessToken}
	if !tok.Expiry.IsZero() {
		res.ExpiresOn = timestamppb.New(tok.Expiry)
	}
	return res, nil
}

// DeleteMCPOAuthToken implements AIService. It can only be called by deployments, for users of their own project.
func (s *Server) DeleteMCPOAuthToken(ctx context.Context, req *adminv1.DeleteMCPOAuthTokenRequest) (*adminv1.DeleteMCPOAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.connector", req.Connector),
		attribute.String("args.user_id", req.UserId),
	)

	depl, err := s.mcpOAuthCallerDeployment(ctx)
	if err != nil {
		return nil, err
	}
	if req.Connector == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "connector and user_id are required")
	}

	row, err := s.admin.DB.FindMCPOAuthToken(ctx, depl.ProjectID, req.Connector, req.UserId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &adminv1.DeleteMCPOAuthTokenResponse{}, nil
		}
		return nil, err
	}

	// Best-effort revocation at the authorization server
	tok := &oauth2.Token{}
	if json.Unmarshal(row.Token, tok) == nil {
		if clientRow, err := s.admin.DB.FindMCPOAuthClient(ctx, depl.ProjectID, req.Connector); err == nil {
			oauthClient := &mcp.OAuthClient{}
			if json.Unmarshal(clientRow.Client, oauthClient) == nil {
				if err := oauthClient.Revoke(ctx, tok, mcp.NewHTTPClient(false, mcpOAuthHTTPTimeout)); err != nil {
					s.logger.Info("failed to revoke MCP OAuth token", zap.String("connector", req.Connector), zap.Error(err), observability.ZapCtx(ctx))
				}
			}
		}
	}

	err = s.admin.DB.DeleteMCPOAuthToken(ctx, depl.ProjectID, req.Connector, req.UserId)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	return &adminv1.DeleteMCPOAuthTokenResponse{}, nil
}

// mcpOAuthValidateRedirect checks that a return URL points back to Rill: a relative path, the admin or frontend hosts,
// or an organization's custom domain. The callback page links to it, so it must never be an attacker-controlled or javascript: URL.
func (s *Server) mcpOAuthValidateRedirect(ctx context.Context, redirect string) error {
	if s.admin.URLs.IsSafeRedirectURL(redirect) {
		return nil
	}
	parsed, err := url.Parse(redirect)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("invalid redirect %q", redirect)
	}
	_, err = s.admin.DB.FindOrganizationByCustomDomain(ctx, parsed.Host)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("redirect to %q is not allowed", redirect)
		}
		return err
	}
	return nil
}

// mcpOAuthCallerDeployment returns the deployment making the request, or a PermissionDenied error if the caller is not a deployment.
func (s *Server) mcpOAuthCallerDeployment(ctx context.Context) (*database.Deployment, error) {
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeDeployment {
		return nil, status.Error(codes.PermissionDenied, "only deployments can access MCP OAuth tokens")
	}
	depl, err := s.admin.DB.FindDeployment(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}
	return depl, nil
}

// mcpOAuthResolveDeployment finds the project and the deployment (prod, or the given branch) the user wants to authorize a server for,
// and checks the user may read it.
func (s *Server) mcpOAuthResolveDeployment(ctx context.Context, claims auth.Claims, orgName, projectName, branch string) (*database.Project, *database.Deployment, *adminv1.ProjectPermissions, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, orgName, projectName)
	if err != nil {
		return nil, nil, nil, mcpOAuthHTTPError{status: http.StatusNotFound, msg: "project not found"}
	}
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}
	if !permissions.ReadProject {
		return nil, nil, nil, mcpOAuthHTTPError{status: http.StatusForbidden, msg: "you do not have access to this project"}
	}

	var depl *database.Deployment
	if branch != "" {
		depls, err := s.admin.DB.FindDeploymentsForProject(ctx, proj.ID, "", branch)
		if err != nil {
			return nil, nil, nil, err
		}
		if len(depls) != 1 {
			return nil, nil, nil, mcpOAuthHTTPError{status: http.StatusNotFound, msg: fmt.Sprintf("no unique deployment found for branch %q", branch)}
		}
		depl = depls[0]
	} else {
		if proj.PrimaryDeploymentID == nil {
			return nil, nil, nil, mcpOAuthHTTPError{status: http.StatusNotFound, msg: "the project has no deployment"}
		}
		depl, err = s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if depl.Environment == "dev" && !permissions.ReadDev {
		return nil, nil, nil, mcpOAuthHTTPError{status: http.StatusForbidden, msg: "you do not have access to this deployment"}
	}
	if depl.Environment != "dev" && !permissions.ReadProd {
		return nil, nil, nil, mcpOAuthHTTPError{status: http.StatusForbidden, msg: "you do not have access to this deployment"}
	}
	return proj, depl, permissions, nil
}

// mcpOAuthUserServer asks the deployment's runtime, as the user, whether the MCP server exists and the user may use it.
// The runtime evaluates the connector's security policy for the user's attributes.
func (s *Server) mcpOAuthUserServer(ctx context.Context, proj *database.Project, depl *database.Deployment, permissions *adminv1.ProjectPermissions, connector string) (*runtimev1.MCPServer, error) {
	if depl.RuntimeHost == "" {
		return nil, mcpOAuthHTTPError{status: http.StatusServiceUnavailable, msg: "the deployment is not ready"}
	}
	jwt, err := s.issueRuntimeToken(ctx, &issueRuntimeTokenOptions{
		project:            proj,
		deployment:         depl,
		projectPermissions: permissions,
		forOwner:           true,
		ttl:                5 * time.Minute,
	})
	if err != nil {
		return nil, err
	}
	rt, err := client.New(depl.RuntimeHost, jwt)
	if err != nil {
		return nil, err
	}
	defer rt.Close()
	res, err := rt.ListMCPServers(ctx, &runtimev1.ListMCPServersRequest{InstanceId: depl.RuntimeInstanceID})
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP servers: %w", err)
	}
	for _, srv := range res.Servers {
		if srv.Connector != connector {
			continue
		}
		if !srv.Oauth {
			return nil, mcpOAuthHTTPError{status: http.StatusBadRequest, msg: "the server does not use per-user OAuth"}
		}
		return srv, nil
	}
	return nil, mcpOAuthHTTPError{status: http.StatusForbidden, msg: "the server does not exist or you do not have access to it"}
}

// mcpOAuthConnectorConfig is the OAuth-related configuration of an MCP connector, with project secrets resolved.
type mcpOAuthConnectorConfig struct {
	URL          string
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// mcpOAuthConnectorConfig fetches the connector's resolved configuration from the deployment's runtime with a management token.
func (s *Server) mcpOAuthConnectorConfig(ctx context.Context, depl *database.Deployment, connector string) (*mcpOAuthConnectorConfig, error) {
	rt, err := s.admin.OpenRuntimeClient(depl)
	if err != nil {
		return nil, err
	}
	defer rt.Close()
	res, err := rt.AnalyzeConnectors(ctx, &runtimev1.AnalyzeConnectorsRequest{InstanceId: depl.RuntimeInstanceID})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze connectors: %w", err)
	}
	for _, c := range res.Connectors {
		if c.Name != connector {
			continue
		}
		if c.ErrorMessage != "" {
			return nil, mcpOAuthHTTPError{status: http.StatusBadRequest, msg: fmt.Sprintf("the connector has an error: %s", c.ErrorMessage)}
		}
		if c.Driver == nil || c.Driver.Name != "mcp" {
			return nil, mcpOAuthHTTPError{status: http.StatusBadRequest, msg: "the connector is not an MCP server"}
		}
		cfg := c.Config.AsMap()
		out := &mcpOAuthConnectorConfig{}
		out.URL, _ = cfg["url"].(string)
		out.ClientID, _ = cfg["oauth_client_id"].(string)
		out.ClientSecret, _ = cfg["oauth_client_secret"].(string)
		switch scopes := cfg["oauth_scopes"].(type) {
		case string:
			if scopes != "" {
				out.Scopes = []string{scopes}
			}
		case []any:
			for _, sc := range scopes {
				if str, ok := sc.(string); ok {
					out.Scopes = append(out.Scopes, str)
				}
			}
		}
		if out.URL == "" {
			return nil, mcpOAuthHTTPError{status: http.StatusBadRequest, msg: "the connector has no url"}
		}
		return out, nil
	}
	return nil, mcpOAuthHTTPError{status: http.StatusNotFound, msg: "connector not found"}
}

// mcpOAuthClient returns the project's OAuth client for the connector, discovering and registering one if there is none
// or the stored one no longer matches the connector's configuration.
func (s *Server) mcpOAuthClient(ctx context.Context, proj *database.Project, connector string, cfg *mcpOAuthConnectorConfig) (*mcp.OAuthClient, error) {
	redirectURL := s.admin.URLs.MCPOAuthCallback()

	row, err := s.admin.DB.FindMCPOAuthClient(ctx, proj.ID, connector)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		existing := &mcp.OAuthClient{}
		if json.Unmarshal(row.Client, existing) == nil && existing.Matches(cfg.URL, redirectURL) && (cfg.ClientID == "" || existing.ClientID == cfg.ClientID) {
			return existing, nil
		}
	}

	oauthClient, err := mcp.DiscoverOAuthClient(ctx, &mcp.DiscoverOAuthClientOptions{
		ServerURL:    cfg.URL,
		RedirectURL:  redirectURL,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scopes,
		HTTPClient:   mcp.NewHTTPClient(false, mcpOAuthHTTPTimeout),
	})
	if err != nil {
		return nil, mcpOAuthHTTPError{status: http.StatusBadGateway, msg: err.Error()}
	}
	clientJSON, err := json.Marshal(oauthClient)
	if err != nil {
		return nil, err
	}
	_, err = s.admin.DB.UpsertMCPOAuthClient(ctx, &database.UpsertMCPOAuthClientOptions{
		ProjectID: proj.ID,
		Connector: connector,
		Client:    clientJSON,
	})
	if err != nil {
		return nil, err
	}
	return oauthClient, nil
}

// mcpOAuthHTTPError is an error with an HTTP status for the browser-facing handlers.
type mcpOAuthHTTPError struct {
	status int
	msg    string
}

func (e mcpOAuthHTTPError) Error() string {
	return e.msg
}

func writeMCPOAuthError(w http.ResponseWriter, err error) {
	var httpErr mcpOAuthHTTPError
	if errors.As(err, &httpErr) {
		http.Error(w, httpErr.msg, httpErr.status)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
