package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/rilldata/rill/admin/pkg/oauth"
	"go.uber.org/zap"
)

const maxOAuthRegistrationBodyBytes = 1 << 20

// handleOAuthProtectedResourceMetadata serves the OAuth 2.0 Protected Resource Metadata as per RFC 9728 and MCP OAuth specification.
// This endpoint helps MCP clients discover the authorization server for this protected resource. https://www.rfc-editor.org/rfc/rfc9728.html
func (a *Authenticator) handleOAuthProtectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := oauth.ProtectedResourceMetadata{
		Resource:             a.admin.URLs.OAuthExternalResourceURL(r),
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

	// Bound this public endpoint before decoding so a registration request cannot
	// consume an arbitrary amount of server memory.
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxOAuthRegistrationBodyBytes))
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
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

	redirectURIs, err := sanitizeRedirectURIs(req.RedirectURIs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(redirectURIs) == 0 {
		http.Error(w, "at least one valid redirect_uri is required", http.StatusBadRequest)
		return
	}

	// Dynamic registration is untrusted. In particular, it must never mint a
	// client eligible for non-expiring access tokens; privileged clients are
	// provisioned through trusted internal paths instead.
	scope, err := validateDynamicRegistrationScope(req.Scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	grantTypes, err := validateDynamicRegistrationGrantTypes(req.GrantTypes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(grantTypes) == 0 {
		// Default to authorization_code if none provided
		grantTypes = []string{authorizationCodeGrantType}
	}

	// Create a new auth client in the database
	client, err := a.admin.DB.InsertAuthClient(r.Context(), displayName, scope, grantTypes, redirectURIs)
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
		RedirectURIs:            redirectURIs,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		ResponseTypes:           req.ResponseTypes,
		ClientURI:               req.ClientURI,
	}

	// Marshal before committing the status. Once a response or redirect starts,
	// an error must not append a second HTTP error body to it.
	respBytes, err := json.Marshal(resp)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to encode response: %w", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(respBytes); err != nil {
		a.logger.Error("Failed to write OAuth registration response", zap.Error(err))
		return
	}

	a.logger.Info("Registered new OAuth client", zap.String("client_id", client.ID), zap.String("client_name", displayName))
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

func validateDynamicRegistrationScope(scope string) (string, error) {
	seen := make(map[string]struct{})
	var sanitized []string
	for _, token := range strings.Fields(scope) {
		switch token {
		case "offline_access":
			// Safe for public clients when paired with an approved refresh grant.
		case longLivedAccessTokenScope:
			return "", fmt.Errorf("scope %q requires trusted registration", token)
		default:
			return "", fmt.Errorf("unsupported scope %q", token)
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		sanitized = append(sanitized, token)
	}
	return strings.Join(sanitized, " "), nil
}

func validateDynamicRegistrationGrantTypes(grants []string) ([]string, error) {
	seen := make(map[string]struct{})
	var sanitized []string
	for _, grant := range sanitizeGrantTypes(grants) {
		switch grant {
		case authorizationCodeGrantType, refreshTokenGrantType, deviceCodeGrantType:
		default:
			return nil, fmt.Errorf("unsupported grant_type %q", grant)
		}
		if _, ok := seen[grant]; ok {
			continue
		}
		seen[grant] = struct{}{}
		sanitized = append(sanitized, grant)
	}
	return sanitized, nil
}

// sanitizeRedirectURIs validates redirect URIs and normalizes them.
// Requires absolute HTTP(S) URLs without fragments.
func sanitizeRedirectURIs(uris []string) ([]string, error) {
	seen := make(map[string]struct{})
	var sanitized []string
	for _, raw := range uris {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		parsed, err := url.Parse(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid redirect_uri %q: %w", raw, err)
		}
		parsed.Scheme = strings.ToLower(parsed.Scheme)
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return nil, fmt.Errorf("redirect_uri %q must use http or https", raw)
		}
		if parsed.Host == "" {
			return nil, fmt.Errorf("redirect_uri %q must include a host", raw)
		}
		// Authorization codes sent over plain HTTP are exposed in transit. Native
		// clients may use HTTP only on the local machine; remote callbacks use TLS.
		if parsed.Scheme == "http" && !isLoopbackHostname(parsed.Hostname()) {
			return nil, fmt.Errorf("redirect_uri %q must use https unless it is loopback", raw)
		}
		if parsed.Fragment != "" {
			return nil, fmt.Errorf("redirect_uri %q must not include a fragment", raw)
		}

		normalized := parsed.String()
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		sanitized = append(sanitized, normalized)
	}
	return sanitized, nil
}

func isLoopbackHostname(hostname string) bool {
	if strings.EqualFold(hostname, "localhost") {
		return true
	}
	ip := net.ParseIP(hostname)
	return ip != nil && ip.IsLoopback()
}
