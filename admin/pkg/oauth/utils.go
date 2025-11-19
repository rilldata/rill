package oauth

const (
	FormMediaType = "application/x-www-form-urlencoded"
	JSONMediaType = "application/json"
)

// TokenResponse contains the information returned after fetching an access token from the OAuth server.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`
	UserID       string `json:"user_id"`
}

// LegacyTokenResponse for backwards compatibility with older Rill CLI client that expect expires_in as a string.
// TODO remove this after 2-3 releases and only keep TokenResponse, also remove sending of token_response_version in exchange token requests.
type LegacyTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in,string"`
	TokenType   string `json:"token_type"`
	UserID      string `json:"user_id"`
}

// ProtectedResourceMetadata contains the OAuth 2.0 Protected Resource Metadata as per RFC 8414
// See: https://www.rfc-editor.org/rfc/rfc8414.html
type ProtectedResourceMetadata struct {
	// Resource is the protected resource identifier
	Resource string `json:"resource,omitempty"`
	// AuthorizationServers is an array of strings representing authorization server issuer identifiers
	AuthorizationServers []string `json:"authorization_servers"`
	// BearerMethodsSupported is an array of the OAuth 2.0 Bearer Token methods supported by this resource
	BearerMethodsSupported []string `json:"bearer_methods_supported,omitempty"`
	// ResourceSigningAlgValuesSupported is an array of the JWS signing algorithms supported by the resource
	ResourceSigningAlgValuesSupported []string `json:"resource_signing_alg_values_supported,omitempty"`
	// ResourceDocumentation is a URL of a page containing human-readable information about the resource
	ResourceDocumentation string `json:"resource_documentation,omitempty"`
}

// AuthorizationServerMetadata contains OAuth 2.0 Authorization Server Metadata as per RFC 8414
// See: https://www.rfc-editor.org/rfc/rfc8414.html#section-2
type AuthorizationServerMetadata struct {
	// Issuer is the authorization server's issuer identifier URL
	Issuer string `json:"issuer"`
	// AuthorizationEndpoint is the URL of the authorization server's authorization endpoint
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	// TokenEndpoint is the URL of the authorization server's token endpoint
	TokenEndpoint string `json:"token_endpoint"`
	// RegistrationEndpoint is the URL of the authorization server's dynamic client registration endpoint
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`
	// JWKSURI is the URL of the authorization server's JSON Web Key Set document
	JWKSURI string `json:"jwks_uri,omitempty"`
	// ScopesSupported is an array of the OAuth 2.0 scope values that this authorization server supports
	ScopesSupported []string `json:"scopes_supported,omitempty"`
	// ResponseTypesSupported is an array of the OAuth 2.0 response_type values that this authorization server supports
	ResponseTypesSupported []string `json:"response_types_supported"`
	// ResponseModesSupported is an array of the OAuth 2.0 response_mode values that this authorization server supports
	ResponseModesSupported []string `json:"response_modes_supported,omitempty"`
	// GrantTypesSupported is an array of the OAuth 2.0 grant type values that this authorization server supports
	GrantTypesSupported []string `json:"grant_types_supported"`
	// TokenEndpointAuthMethodsSupported is an array of client authentication methods supported by this token endpoint
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	// CodeChallengeMethodsSupported is an array of PKCE code challenge methods supported
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported,omitempty"`
	// ServiceDocumentation is a URL of a page containing human-readable information about the authorization server
	ServiceDocumentation string `json:"service_documentation,omitempty"`
}

// ClientRegistrationRequest contains the OAuth 2.0 Dynamic Client Registration request as per RFC 7591
// See: https://www.rfc-editor.org/rfc/rfc7591.html
type ClientRegistrationRequest struct {
	// RedirectURIs is an array of redirection URIs for use in redirect-based flows
	RedirectURIs []string `json:"redirect_uris,omitempty"`
	// Scope indicates the scope values that the client is requesting
	Scope string `json:"scope,omitempty"`
	// TokenEndpointAuthMethod indicates the requested authentication method for the token endpoint
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method,omitempty"`
	// GrantTypes is an array of OAuth 2.0 grant types that the client will use
	GrantTypes []string `json:"grant_types,omitempty"`
	// ResponseTypes is an array of OAuth 2.0 response types that the client will use
	ResponseTypes []string `json:"response_types,omitempty"`
	// ClientName is a human-readable name for the client
	ClientName string `json:"client_name,omitempty"`
	// ClientURI is a URL of the home page of the client
	ClientURI string `json:"client_uri,omitempty"`
}

// ClientRegistrationResponse contains the OAuth 2.0 Dynamic Client Registration response as per RFC 7591
// See: https://www.rfc-editor.org/rfc/rfc7591.html
type ClientRegistrationResponse struct {
	// ClientID is the unique client identifier
	ClientID string `json:"client_id"`
	// ClientName is a human-readable name for the client
	ClientName string `json:"client_name,omitempty"`
	// Scope indicates the scope values that the client is registered for
	Scope string `json:"scope,omitempty"`
	// ClientIDIssuedAt is the time at which the client identifier was issued (Unix timestamp)
	ClientIDIssuedAt int64 `json:"client_id_issued_at,omitempty"`
	// RedirectURIs is an array of redirection URIs for use in redirect-based flows
	RedirectURIs []string `json:"redirect_uris"`
	// TokenEndpointAuthMethod indicates the authentication method for the token endpoint
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method,omitempty"`
	// GrantTypes is an array of OAuth 2.0 grant types that the client may use
	GrantTypes []string `json:"grant_types,omitempty"`
	// ResponseTypes is an array of OAuth 2.0 response types that the client may use
	ResponseTypes []string `json:"response_types,omitempty"`
	// ClientURI is a URL of the home page of the client
	ClientURI string `json:"client_uri,omitempty"`
}
