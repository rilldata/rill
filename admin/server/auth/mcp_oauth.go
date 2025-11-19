package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rilldata/rill/admin/pkg/oauth"
	"go.uber.org/zap"
)

// handleOAuthProtectedResourceMetadata serves the OAuth 2.0 Protected Resource Metadata
// as per RFC 8414 and MCP OAuth specification.
// This endpoint helps MCP clients discover the authorization server for this protected resource.
func (a *Authenticator) handleOAuthProtectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := oauth.ProtectedResourceMetadata{
		Resource:             a.admin.URLs.External(),
		AuthorizationServers: []string{a.admin.URLs.External()},
		BearerMethodsSupported: []string{
			"header", // Authorization: Bearer <token>
		},
		ResourceDocumentation: "https://docs.rilldata.com",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		internalServerError(w, fmt.Errorf("failed to encode metadata: %w", err))
		return
	}
}

// handleOAuthAuthorizationServerMetadata serves the OAuth 2.0 Authorization Server Metadata
// as per RFC 8414. This endpoint provides information about the OAuth 2.0 authorization server
// including supported flows, endpoints, and capabilities.
func (a *Authenticator) handleOAuthAuthorizationServerMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := oauth.AuthorizationServerMetadata{
		Issuer:                a.admin.URLs.External(),
		AuthorizationEndpoint: a.admin.URLs.OAuthAuthorize(),
		TokenEndpoint:         a.admin.URLs.OAuthToken(),
		RegistrationEndpoint:  a.admin.URLs.OAuthRegister(),
		JWKSURI:               a.admin.URLs.OAuthJWKS(),
		ScopesSupported: []string{
			"offline_access", // Refresh token support
		},
		ResponseTypesSupported: []string{
			"code", // Authorization code flow
		},
		ResponseModesSupported: []string{
			"query", // Response parameters in query string
		},
		GrantTypesSupported: []string{
			authorizationCodeGrantType, // Authorization code grant
			deviceCodeGrantType,        // Device code grant
			refreshTokenGrantType,      // Refresh token grant
		},
		TokenEndpointAuthMethodsSupported: []string{
			"none", // Public clients (PKCE)
		},
		CodeChallengeMethodsSupported: []string{
			"S256", // SHA-256 based PKCE
		},
		ServiceDocumentation: "https://docs.rilldata.com",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		internalServerError(w, fmt.Errorf("failed to encode metadata: %w", err))
		return
	}
}

// handleOAuthRegister handles OAuth 2.0 Dynamic Client Registration as per RFC 7591.
// This endpoint allows MCP clients like Claude Desktop or ChatGPT Desktop to dynamically
// register and obtain a client_id for use in OAuth flows.
func (a *Authenticator) handleOAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected a POST request", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to read request body: %w", err))
		return
	}

	var req oauth.ClientRegistrationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Use client_name if provided, otherwise use a default name
	displayName := req.ClientName
	if displayName == "" {
		displayName = "MCP Client"
	}

	// Validate redirect_uris - at least one is required for OAuth flows
	if len(req.RedirectURIs) == 0 {
		http.Error(w, "at least one redirect_uri is required", http.StatusBadRequest)
		return
	}

	scope := sanitizeScope(req.Scope)
	grantTypes := sanitizeGrantTypes(req.GrantTypes)
	if len(grantTypes) == 0 {
		// Default to authorization_code if none provided
		grantTypes = []string{authorizationCodeGrantType}
	}

	// Create a new auth client in the database
	client, err := a.admin.DB.InsertAuthClient(r.Context(), displayName, scope, grantTypes)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to create auth client: %w", err))
		return
	}

	// Build response - echo back the client's registration metadata as per RFC 7591
	resp := oauth.ClientRegistrationResponse{
		ClientID:                client.ID,
		ClientName:              client.DisplayName,
		Scope:                   client.Scope,
		GrantTypes:              client.GrantTypes,
		ClientIDIssuedAt:        client.CreatedOn.Unix(),
		RedirectURIs:            req.RedirectURIs,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		ResponseTypes:           req.ResponseTypes,
		ClientURI:               req.ClientURI,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		internalServerError(w, fmt.Errorf("failed to encode response: %w", err))
		return
	}

	a.logger.Info("Registered new OAuth client", zap.String("client_id", client.ID), zap.String("client_name", displayName))
}

// remove extra spaces from space separated scope string
func sanitizeScope(scope string) string {
	return strings.Join(strings.Fields(scope), " ")
}

// trims white spaces
func sanitizeGrantTypes(grants []string) []string {
	var sanitized []string
	for _, grant := range grants {
		trimmed := strings.TrimSpace(grant)
		if trimmed != "" {
			sanitized = append(sanitized, trimmed)
		}
	}
	return sanitized
}
