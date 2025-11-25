package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/server/cookies"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	authorizationCodeGrantType = "authorization_code"
	refreshTokenGrantType      = "refresh_token"
	deviceCodeGrantType        = "urn:ietf:params:oauth:grant-type:device_code"
	longLivedAccessTokenScope  = "long_lived_access_token" // nolint:gosec // custom scope to indicate long-lived access token
)

// AuthenticatorOptions provides options for Authenticator
type AuthenticatorOptions struct {
	AuthDomain       string
	AuthClientID     string
	AuthClientSecret string
}

// Authenticator wraps functionality for admin server auth.
// It provides endpoints for login/logout, creates users, issues cookie-based auth tokens, and provides middleware for authenticating requests.
// The implementation was derived from: https://auth0.com/docs/quickstart/webapp/golang/01-login.
type Authenticator struct {
	logger  *zap.Logger
	admin   *admin.Service
	cookies *cookies.Store
	opts    *AuthenticatorOptions
	oidc    *oidc.Provider
	oauth2  oauth2.Config
}

// NewAuthenticator creates an Authenticator.
func NewAuthenticator(logger *zap.Logger, adm *admin.Service, cookieStore *cookies.Store, opts *AuthenticatorOptions) (*Authenticator, error) {
	oidcProvider, err := oidc.NewProvider(context.Background(), "https://"+opts.AuthDomain+"/")
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     opts.AuthClientID,
		ClientSecret: opts.AuthClientSecret,
		RedirectURL:  adm.URLs.AuthLoginCallback(),
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	a := &Authenticator{
		logger:  logger,
		admin:   adm,
		cookies: cookieStore,
		opts:    opts,
		oidc:    oidcProvider,
		oauth2:  oauth2Config,
	}

	return a, nil
}
