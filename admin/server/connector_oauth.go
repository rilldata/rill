package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/sessions"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/oauthclient"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/safehttp"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file implements per-user OAuth authorization of connectors (such as MCP servers with auth "oauth") in Rill Cloud.
//
// The browser-based flow runs here because the admin service owns user sessions:
// /connectors/oauth/connect asks the deployment's runtime whether the user may use the connector and for its OAuth configuration,
// discovers the authorization server, registers a client once per configuration, and redirects the user;
// /connectors/oauth/callback exchanges the code and stores the token encrypted per (client, user).
// A client is bound to the hash of the connector's OAuth configuration, so a configuration change never reuses existing authorizations.
// Runtimes never run the flow; they fetch a user's access token through the deployment-authenticated GetConnectorOAuthToken RPC.

const (
	connectorOAuthCookieName = "connector_oauth"
	// connectorOAuthFlowTTL is how long a pending authorization flow stays valid.
	connectorOAuthFlowTTL = 15 * time.Minute
	// connectorOAuthHTTPTimeout bounds discovery and token requests to authorization servers.
	connectorOAuthHTTPTimeout = 15 * time.Second
)

// connectorOAuthFlow is the state of a pending browser-based authorization flow.
// It is persisted between the two requests in an encrypted cookie of its own, named after the flow's OAuth state parameter.
type connectorOAuthFlow struct {
	// UserID is the user who started the flow; only that user can complete it.
	UserID    string `json:"user_id"`
	Org       string `json:"org"`
	Project   string `json:"project"`
	Connector string `json:"connector"`
	// OAuthClientID is the database ID of the ConnectorOAuthClient the authorization was started with; the callback completes it with the same client.
	OAuthClientID string `json:"oauth_client_id"`
	Verifier      string `json:"verifier"`
	Redirect      string `json:"redirect"`
	Expires       int64  `json:"expires"`
}

// registerConnectorOAuthEndpoints registers the browser-facing endpoints for authorizing connectors on behalf of users.
func (s *Server) registerConnectorOAuthEndpoints(mux *http.ServeMux) {
	inner := http.NewServeMux()
	observability.MuxHandle(inner, "/connectors/oauth/connect", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.connectorOAuthConnect)))
	observability.MuxHandle(inner, "/connectors/oauth/callback", s.authenticator.HTTPMiddleware(http.HandlerFunc(s.connectorOAuthCallback)))
	mux.Handle("/connectors/oauth/", observability.Middleware("admin", s.logger, inner))
}

// GetConnectorOAuthToken implements AdminService
func (s *Server) GetConnectorOAuthToken(ctx context.Context, req *adminv1.GetConnectorOAuthTokenRequest) (*adminv1.GetConnectorOAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.connector", req.Connector),
		attribute.String("args.config_hash", req.ConfigHash),
		attribute.String("args.user_id", req.UserId),
		attribute.Bool("args.force_refresh", req.ForceRefresh),
	)

	depl, err := s.connectorOAuthCallerDeployment(ctx)
	if err != nil {
		return nil, err
	}
	// The client is bound to the connector's OAuth configuration.
	clientRow, err := s.admin.DB.FindConnectorOAuthClient(ctx, depl.ProjectID, req.Connector, req.ConfigHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "the user has not authorized the connector")
		}
		return nil, err
	}

	var tok *oauth2.Token
	// Set when the refresh showed that the authorization is gone: the token is deleted inside the transaction,
	// and the error is only returned after the deletion has been committed.
	var expired bool
	err = s.connectorOAuthWithTokenLock(ctx, clientRow.ID, req.UserId, func(ctx context.Context) error {
		row, err := s.admin.DB.FindConnectorOAuthToken(ctx, clientRow.ID, req.UserId)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return status.Error(codes.NotFound, "the user has not authorized the connector")
			}
			return err
		}
		tok = &oauth2.Token{}
		if err := json.Unmarshal(row.Token, tok); err != nil {
			return fmt.Errorf("stored token is invalid: %w", err)
		}
		if !req.ForceRefresh {
			// The token row is written whenever a token is obtained or refreshed, so its update time is when the token was issued.
			if !oauthclient.TokenExpiring(tok, row.UpdatedOn) {
				return nil
			}
			// A token without a refresh token can't be renewed early, so it is served until it actually expires.
			if tok.RefreshToken == "" && tok.Valid() {
				return nil
			}
		}

		oauthClient := &oauthclient.Client{}
		if err := json.Unmarshal(clientRow.Client, oauthClient); err != nil {
			return fmt.Errorf("stored OAuth client is invalid: %w", err)
		}
		// The refresh deliberately runs while the lock and its transaction are held, unlike client registration.
		// Many servers rotate refresh tokens and treat a reused one as theft, revoking the whole grant, so concurrent refreshes must be serialized.
		// A user's token is refreshed at most about once per lifetime, which bounds how many connections this holds.
		newTok, err := oauthClient.Refresh(ctx, tok, safehttp.NewHTTPClient(false, connectorOAuthHTTPTimeout))
		if err != nil {
			if errors.Is(err, oauthclient.ErrAuthorizationRequired) {
				// The authorization is gone for good; forget it so the user is asked to connect again.
				expired = true
				return s.admin.DB.DeleteConnectorOAuthToken(ctx, clientRow.ID, req.UserId)
			}
			if !req.ForceRefresh && tok.Valid() {
				// The refresh was only early; the current token still works until it expires, and the next request retries.
				s.logger.Warn("failed to refresh connector OAuth token early", zap.String("connector", req.Connector), zap.Error(err), observability.ZapCtx(ctx))
				return nil
			}
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		tokJSON, err := json.Marshal(newTok)
		if err != nil {
			return err
		}
		_, err = s.admin.DB.UpsertConnectorOAuthToken(ctx, &database.UpsertConnectorOAuthTokenOptions{
			OAuthClientID: clientRow.ID,
			UserID:        req.UserId,
			Token:         tokJSON,
		})
		if err != nil {
			return fmt.Errorf("failed to store refreshed token: %w", err)
		}
		tok = newTok
		return nil
	})
	if err != nil {
		return nil, err
	}
	if expired {
		return nil, status.Error(codes.NotFound, "the user's authorization has expired")
	}

	res := &adminv1.GetConnectorOAuthTokenResponse{AccessToken: tok.AccessToken}
	if !tok.Expiry.IsZero() {
		res.ExpiresOn = timestamppb.New(tok.Expiry)
	}
	return res, nil
}

// DeleteConnectorOAuthToken implements AdminService
func (s *Server) DeleteConnectorOAuthToken(ctx context.Context, req *adminv1.DeleteConnectorOAuthTokenRequest) (*adminv1.DeleteConnectorOAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.connector", req.Connector),
		attribute.String("args.config_hash", req.ConfigHash),
		attribute.String("args.user_id", req.UserId),
	)

	depl, err := s.connectorOAuthCallerDeployment(ctx)
	if err != nil {
		return nil, err
	}
	// The user may hold tokens for earlier configurations of the connector too, such as before a scope change, or for other branches.
	// They are all removed, so no grant outlives the disconnect.
	clientRows, err := s.admin.DB.FindConnectorOAuthClients(ctx, depl.ProjectID, req.Connector)
	if err != nil {
		return nil, err
	}
	type deletedToken struct {
		client *database.ConnectorOAuthClient
		token  *database.ConnectorOAuthToken
	}
	var deleted []deletedToken
	for _, clientRow := range clientRows {
		err = s.connectorOAuthWithTokenLock(ctx, clientRow.ID, req.UserId, func(ctx context.Context) error {
			row, err := s.admin.DB.FindConnectorOAuthToken(ctx, clientRow.ID, req.UserId)
			if err != nil {
				if errors.Is(err, database.ErrNotFound) {
					return nil
				}
				return err
			}
			err = s.admin.DB.DeleteConnectorOAuthToken(ctx, clientRow.ID, req.UserId)
			if err != nil {
				return err
			}
			deleted = append(deleted, deletedToken{client: clientRow, token: row})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Clients of earlier configurations that no longer hold tokens are not needed anymore.
	err = s.admin.DB.DeleteConnectorOAuthClientsWithoutTokens(ctx, depl.ProjectID, req.Connector, req.ConfigHash)
	if err != nil {
		s.logger.Warn("failed to delete unused connector OAuth clients", zap.String("connector", req.Connector), zap.Error(err), observability.ZapCtx(ctx))
	}

	// Revoke the deleted tokens, so the user's grants end at the authorization server too. It is best-effort and runs after the commits.
	for _, d := range deleted {
		oauthClient := &oauthclient.Client{}
		tok := &oauth2.Token{}
		if json.Unmarshal(d.client.Client, oauthClient) != nil || json.Unmarshal(d.token.Token, tok) != nil {
			continue
		}
		err = oauthClient.Revoke(ctx, tok, safehttp.NewHTTPClient(false, connectorOAuthHTTPTimeout))
		if err != nil {
			s.logger.Info("failed to revoke connector OAuth token", zap.String("connector", req.Connector), zap.Error(err), observability.ZapCtx(ctx))
		}
	}
	return &adminv1.DeleteConnectorOAuthTokenResponse{}, nil
}

// connectorOAuthConnect starts an authorization flow for a connector on behalf of the authenticated user.
// Query params: "org", "project" and "connector" (required); "branch" (optional, for edit-mode deployments);
// "redirect" (optional URL to return to when the page was not opened as a popup).
func (s *Server) connectorOAuthConnect(w http.ResponseWriter, r *http.Request) {
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
	} else if err := s.connectorOAuthValidateRedirect(ctx, redirect); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	proj, depl, permissions, err := s.connectorOAuthResolveDeployment(ctx, claims, orgName, projectName, qry.Get("branch"))
	if err != nil {
		writeConnectorOAuthError(w, err)
		return
	}

	// Ask the runtime whether the user may use the connector (it evaluates the connector's security policy for the user),
	// so Rill never collects a token for a user who cannot use the connector, and for the connector's OAuth configuration.
	cfg, err := s.connectorOAuthConfig(ctx, proj, depl, permissions, claims.OwnerID(), connector)
	if err != nil {
		writeConnectorOAuthError(w, err)
		return
	}

	clientRow, oauthClient, err := s.connectorOAuthClient(ctx, proj, connector, cfg)
	if err != nil {
		s.logger.Warn("failed to prepare connector OAuth client", zap.String("project_id", proj.ID), zap.String("connector", connector), zap.Error(err), observability.ZapCtx(ctx))
		writeConnectorOAuthError(w, err)
		return
	}

	state, err := oauthclient.GenerateState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}
	verifier := oauthclient.GenerateVerifier()

	flow := &connectorOAuthFlow{
		UserID:        claims.OwnerID(),
		Org:           orgName,
		Project:       projectName,
		Connector:     connector,
		OAuthClientID: clientRow.ID,
		Verifier:      verifier,
		Redirect:      redirect,
		Expires:       time.Now().Add(connectorOAuthFlowTTL).Unix(),
	}
	flowJSON, err := json.Marshal(flow)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to encode flow: %s", err), http.StatusInternalServerError)
		return
	}
	sess := s.connectorOAuthFlowSession(r, state)
	sess.Values["flow"] = string(flowJSON)
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, oauthClient.AuthCodeURL(state, verifier), http.StatusFound)
}

// connectorOAuthCallback completes an authorization flow: it exchanges the code for a token, stores it for the user,
// and renders a page that notifies the opener window and closes itself.
func (s *Server) connectorOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		http.Error(w, "unidentified user", http.StatusUnauthorized)
		return
	}

	// Take the pending flow out of the cookie, so it can only be used once
	qry := r.URL.Query()
	state := qry.Get("state")
	if state == "" {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	sess := s.connectorOAuthFlowSession(r, state)
	flowJSON, _ := sess.Values["flow"].(string)
	if flowJSON != "" {
		sess.Options.MaxAge = -1
		_ = sess.Save(r, w)
	}
	flow := &connectorOAuthFlow{}
	if flowJSON == "" || json.Unmarshal([]byte(flowJSON), flow) != nil || time.Now().Unix() > flow.Expires {
		http.Error(w, "invalid or expired authorization flow; start the flow again", http.StatusBadRequest)
		return
	}
	if flow.UserID != claims.OwnerID() {
		http.Error(w, "the authorization flow was started by another user; start the flow again", http.StatusBadRequest)
		return
	}

	render := func(errMsg string) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(oauthclient.CallbackPage(flow.Connector, errMsg, flow.Redirect)))
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
	clientRow, err := s.admin.DB.FindConnectorOAuthClientByID(ctx, flow.OAuthClientID)
	if err != nil || clientRow.ProjectID != proj.ID || clientRow.Connector != flow.Connector {
		render("the OAuth client the flow was started with no longer exists; start the flow again")
		return
	}
	oauthClient := &oauthclient.Client{}
	if err := json.Unmarshal(clientRow.Client, oauthClient); err != nil {
		render("the registered OAuth client is invalid; start the flow again")
		return
	}
	if err := oauthClient.ValidateIssuer(qry.Get("iss")); err != nil {
		render(err.Error())
		return
	}

	hc := safehttp.NewHTTPClient(false, connectorOAuthHTTPTimeout)
	tok, err := oauthClient.Exchange(ctx, code, flow.Verifier, hc)
	if err != nil {
		s.logger.Warn("failed to exchange connector OAuth code", zap.String("project_id", proj.ID), zap.String("connector", flow.Connector), zap.Error(err), observability.ZapCtx(ctx))
		render(err.Error())
		return
	}
	tokJSON, err := json.Marshal(tok)
	if err != nil {
		render("failed to encode token")
		return
	}
	err = s.connectorOAuthWithTokenLock(ctx, clientRow.ID, claims.OwnerID(), func(ctx context.Context) error {
		_, err := s.admin.DB.UpsertConnectorOAuthToken(ctx, &database.UpsertConnectorOAuthTokenOptions{
			OAuthClientID: clientRow.ID,
			UserID:        claims.OwnerID(),
			Token:         tokJSON,
		})
		return err
	})
	if err != nil {
		render(fmt.Sprintf("failed to store token: %s", err))
		return
	}

	render("")
}

// connectorOAuthWithTokenLock runs fn in a transaction that holds the lock on a user's token for a client.
// Every write of a token (refresh, disconnect, a completed authorization) goes through it,
// so a refresh can never resurrect a token that a disconnect deleted or overwrite a fresh authorization, and two refreshes never race on a rotating refresh token.
// The lock lives in the database, so it also covers multiple admin processes.
func (s *Server) connectorOAuthWithTokenLock(ctx context.Context, oauthClientID, userID string, fn func(ctx context.Context) error) error {
	ctx, tx, err := s.admin.DB.NewTx(ctx, false)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	err = s.admin.DB.AcquireConnectorOAuthTokenLock(ctx, oauthClientID, userID)
	if err != nil {
		return err
	}
	err = fn(ctx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// connectorOAuthValidateRedirect checks that a return URL points back to Rill: a relative path, the admin or frontend hosts,
// or an organization's custom domain. The callback page links to it, so it must never be an attacker-controlled or javascript: URL.
func (s *Server) connectorOAuthValidateRedirect(ctx context.Context, redirect string) error {
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

func (s *Server) connectorOAuthCallerDeployment(ctx context.Context) (*database.Deployment, error) {
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeDeployment {
		return nil, status.Error(codes.PermissionDenied, "only deployments can access connector OAuth tokens")
	}
	depl, err := s.admin.DB.FindDeployment(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}
	return depl, nil
}

// connectorOAuthResolveDeployment finds the project and the deployment (prod, or the given branch) the user wants to authorize a connector for,
// and checks the user may read it.
func (s *Server) connectorOAuthResolveDeployment(ctx context.Context, claims auth.Claims, orgName, projectName, branch string) (*database.Project, *database.Deployment, *adminv1.ProjectPermissions, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, orgName, projectName)
	if err != nil {
		return nil, nil, nil, connectorOAuthHTTPError{status: http.StatusNotFound, msg: "project not found"}
	}
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}
	if !permissions.ReadProject {
		return nil, nil, nil, connectorOAuthHTTPError{status: http.StatusForbidden, msg: "you do not have access to this project"}
	}

	var depl *database.Deployment
	if branch != "" {
		depls, err := s.admin.DB.FindDeploymentsForProject(ctx, proj.ID, "", branch)
		if err != nil {
			return nil, nil, nil, err
		}
		if len(depls) != 1 {
			return nil, nil, nil, connectorOAuthHTTPError{status: http.StatusNotFound, msg: fmt.Sprintf("no unique deployment found for branch %q", branch)}
		}
		depl = depls[0]
	} else {
		if proj.PrimaryDeploymentID == nil {
			return nil, nil, nil, connectorOAuthHTTPError{status: http.StatusNotFound, msg: "the project has no deployment"}
		}
		depl, err = s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if depl.Environment == "dev" && !permissions.ReadDev {
		return nil, nil, nil, connectorOAuthHTTPError{status: http.StatusForbidden, msg: "you do not have access to this deployment"}
	}
	if depl.Environment != "dev" && !permissions.ReadProd {
		return nil, nil, nil, connectorOAuthHTTPError{status: http.StatusForbidden, msg: "you do not have access to this deployment"}
	}
	return proj, depl, permissions, nil
}

// connectorOAuthConfig asks the deployment's runtime whether the user may use the connector, and for the connector's OAuth configuration.
// The runtime evaluates the connector's security policy with the same attributes and rules it would see in the user's own JWT.
func (s *Server) connectorOAuthConfig(ctx context.Context, proj *database.Project, depl *database.Deployment, permissions *adminv1.ProjectPermissions, userID, connector string) (*runtimev1.GetConnectorOAuthConfigResponse, error) {
	rt, err := s.admin.OpenRuntimeClient(depl)
	if err != nil {
		return nil, connectorOAuthHTTPError{status: http.StatusServiceUnavailable, msg: err.Error()}
	}
	defer rt.Close()

	attr, rules, err := s.jwtAttributesAndRulesForUser(ctx, proj, userID, permissions)
	if err != nil {
		return nil, err
	}
	attrPB, err := structpb.NewStruct(attr)
	if err != nil {
		return nil, fmt.Errorf("failed to encode user attributes: %w", err)
	}
	res, err := rt.GetConnectorOAuthConfig(ctx, &runtimev1.GetConnectorOAuthConfigRequest{
		InstanceId:     depl.RuntimeInstanceID,
		Connector:      connector,
		UserAttributes: attrPB,
		UserRules:      rules,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, err
		}
		switch st.Code() {
		case codes.PermissionDenied:
			return nil, connectorOAuthHTTPError{status: http.StatusForbidden, msg: "the connector does not exist or you do not have access to it"}
		case codes.NotFound:
			return nil, connectorOAuthHTTPError{status: http.StatusNotFound, msg: st.Message()}
		case codes.FailedPrecondition, codes.InvalidArgument:
			return nil, connectorOAuthHTTPError{status: http.StatusBadRequest, msg: st.Message()}
		case codes.Unavailable:
			return nil, connectorOAuthHTTPError{status: http.StatusServiceUnavailable, msg: "the deployment is not available"}
		default:
			return nil, fmt.Errorf("failed to get the connector's OAuth configuration: %w", err)
		}
	}
	if res.ConfigHash == "" || res.Config == nil || res.Config.ResourceUrl == "" {
		return nil, fmt.Errorf("the runtime returned an incomplete OAuth configuration for connector %q", connector)
	}
	return res, nil
}

// connectorOAuthClient returns the project's OAuth client for the connector's current configuration, discovering and registering one if none is stored.
// Stored clients are never modified, because tokens and pending authorizations are bound to them.
//
// Registration is deliberately not serialized: two users connecting the same new configuration at the same moment is rare,
// and holding a lock (and a transaction) across discovery would keep a database connection busy for the duration of the HTTP calls.
// If it does happen, both requests discover and register a client, but the unique index on (project, connector, config hash)
// lets only the first insert win; the other request gets the winner's client back and uses it, so both flows complete with the same client.
// The cost is one unused registration at the authorization server (none for pre-registered clients, where discovery only reads metadata).
func (s *Server) connectorOAuthClient(ctx context.Context, proj *database.Project, connector string, cfg *runtimev1.GetConnectorOAuthConfigResponse) (*database.ConnectorOAuthClient, *oauthclient.Client, error) {
	redirectURL := s.admin.URLs.ConnectorOAuthCallback()

	row, err := s.admin.DB.FindConnectorOAuthClient(ctx, proj.ID, connector, cfg.ConfigHash)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, nil, err
	}
	if row != nil {
		existing := &oauthclient.Client{}
		if json.Unmarshal(row.Client, existing) == nil && existing.RedirectURL == redirectURL {
			return row, existing, nil
		}
		// The client was registered with another redirect URL (the admin service's URL changed), so it can no longer start flows.
		// There is one client per configuration, so it is replaced; its tokens go with it and users connect again.
		// A concurrent request may have deleted it already.
		err = s.admin.DB.DeleteConnectorOAuthClient(ctx, row.ID)
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, nil, err
		}
	}

	oauthClient, err := oauthclient.Discover(ctx, &oauthclient.DiscoverOptions{
		ServerURL:           cfg.Config.ResourceUrl,
		RedirectURL:         redirectURL,
		ClientID:            cfg.Config.ClientId,
		ClientSecret:        cfg.Config.ClientSecret,
		Scopes:              cfg.Config.Scopes,
		AuthorizationParams: cfg.Config.AuthorizationParams,
		HTTPClient:          safehttp.NewHTTPClient(false, connectorOAuthHTTPTimeout),
	})
	if err != nil {
		return nil, nil, connectorOAuthHTTPError{status: http.StatusBadGateway, msg: err.Error()}
	}
	clientJSON, err := json.Marshal(oauthClient)
	if err != nil {
		return nil, nil, err
	}
	row, err = s.admin.DB.FindOrInsertConnectorOAuthClient(ctx, &database.InsertConnectorOAuthClientOptions{
		ProjectID:  proj.ID,
		Connector:  connector,
		ConfigHash: cfg.ConfigHash,
		Client:     clientJSON,
	})
	if err != nil {
		return nil, nil, err
	}
	// The row is another request's client if that request inserted one first, so use the stored client rather than the one just registered.
	oauthClient = &oauthclient.Client{}
	if err := json.Unmarshal(row.Client, oauthClient); err != nil {
		return nil, nil, fmt.Errorf("stored OAuth client is invalid: %w", err)
	}
	return row, oauthClient, nil
}

// connectorOAuthFlowSession returns the session that holds the pending flow with the given OAuth state parameter.
// Each flow has its own cookie, so several flows can be pending at once (e.g. when connecting two connectors in a row).
// The cookie expires with the flow and is only sent to the callback.
func (s *Server) connectorOAuthFlowSession(r *http.Request, state string) *sessions.Session {
	sess := s.cookies.Get(r, connectorOAuthCookieName+"_"+state)
	// Copy the options, which may be shared with the cookie store
	opts := *sess.Options
	if callback, err := url.Parse(s.admin.URLs.ConnectorOAuthCallback()); err == nil {
		opts.Path = callback.Path
	}
	opts.MaxAge = int(connectorOAuthFlowTTL.Seconds())
	sess.Options = &opts
	return sess
}

// connectorOAuthHTTPError is an error with an HTTP status for the browser-facing handlers.
type connectorOAuthHTTPError struct {
	status int
	msg    string
}

func (e connectorOAuthHTTPError) Error() string {
	return e.msg
}

func writeConnectorOAuthError(w http.ResponseWriter, err error) {
	var httpErr connectorOAuthHTTPError
	if errors.As(err, &httpErr) {
		http.Error(w, httpErr.msg, httpErr.status)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
