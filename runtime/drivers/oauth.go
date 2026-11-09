package drivers

// OAuthConfigurer is implemented by handles whose connector can use per-user OAuth.
type OAuthConfigurer interface {
	// OAuthConfig returns the per-user OAuth configuration, or nil if the connector doesn't use per-user OAuth.
	OAuthConfig() *OAuthConfig
}

// OAuthConfig is a connector's per-user OAuth configuration.
type OAuthConfig struct {
	ServerURL string
	// Issuer is the authorization server, for resources without protected resource metadata.
	Issuer       string
	ClientID     string
	ClientSecret string
	Scopes       []string
	// AuthorizationParams are extra query parameters for the authorization endpoint.
	AuthorizationParams map[string]string
	// Hash identifies the configuration. It changes when any other field changes, which invalidates existing authorizations.
	Hash string
}
