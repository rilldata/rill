package drivers

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrOAuthUserRequired is returned when a connector needs a per-user OAuth token but the caller is not a signed-in user.
var ErrOAuthUserRequired = errors.New("the connector uses per-user OAuth and requires a signed-in user")

// OAuthAuthorizationRequiredError is returned when the user has not authorized a connector that uses per-user OAuth, or the authorization expired.
// The frontend parses its message.
type OAuthAuthorizationRequiredError struct {
	Connector string
}

func (e *OAuthAuthorizationRequiredError) Error() string {
	return fmt.Sprintf("sign in to connector %q to view this data", e.Connector)
}

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

// OAuthToken is a user's OAuth access token for a connector.
type OAuthToken struct {
	UserID      string
	AccessToken string
	// ExpiresOn is zero if the token does not expire.
	ExpiresOn time.Time
}

// OAuthCaller fetches the OAuth tokens of the user a request runs on behalf of.
type OAuthCaller interface {
	// Token returns the caller's token for a connector.
	// It returns ErrNotImplemented where per-user OAuth is unavailable (Rill Developer),
	// ErrOAuthUserRequired if the caller is not a signed-in user,
	// and *OAuthAuthorizationRequiredError if the user has not authorized the connector.
	Token(ctx context.Context, instanceID, connector, configHash string, forceRefresh bool) (*OAuthToken, error)
}

type oauthCallerKey struct{}

// WithOAuthCaller returns a context for queries made on behalf of the caller.
func WithOAuthCaller(ctx context.Context, c OAuthCaller) context.Context {
	return context.WithValue(ctx, oauthCallerKey{}, c)
}

// OAuthCallerFromContext returns the caller set with WithOAuthCaller, or nil.
func OAuthCallerFromContext(ctx context.Context) OAuthCaller {
	c, _ := ctx.Value(oauthCallerKey{}).(OAuthCaller)
	return c
}

type serviceCredentialKey struct{}

// WithServiceCredential marks a context as system work with no caller (such as reconciling resources),
// which uses the connector's own credentials even if the connector uses per-user OAuth.
func WithServiceCredential(ctx context.Context) context.Context {
	return context.WithValue(ctx, serviceCredentialKey{}, true)
}

// WithoutServiceCredential clears the mark set by WithServiceCredential,
// for work done on behalf of a user inside system work.
func WithoutServiceCredential(ctx context.Context) context.Context {
	return context.WithValue(ctx, serviceCredentialKey{}, false)
}

// UsesServiceCredential reports whether the context was marked with WithServiceCredential.
func UsesServiceCredential(ctx context.Context) bool {
	v, _ := ctx.Value(serviceCredentialKey{}).(bool)
	return v
}
