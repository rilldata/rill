package auth

import (
	"context"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"github.com/rilldata/rill/admin"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"golang.org/x/oauth2"
)

// AuthenticatorOptions provides options for Authenticator
type AuthenticatorOptions struct {
	AuthDomain       string
	AuthClientID     string
	AuthClientSecret string
	ExternalURL      string
	FrontendURL      string
}

// Authenticator wraps functionality for admin server auth.
// It provides endpoints for login/logout, creates users, issues cookie-based auth tokens, and provides middleware for authenticating requests.
// The implementation was derived from: https://auth0.com/docs/quickstart/webapp/golang/01-login.
type Authenticator struct {
	logger  *otelzap.Logger
	admin   *admin.Service
	cookies *sessions.CookieStore
	opts    *AuthenticatorOptions
	oidc    *oidc.Provider
	oauth2  oauth2.Config
}

// NewAuthenticator creates an Authenticator.
func NewAuthenticator(logger *otelzap.Logger, adm *admin.Service, cookies *sessions.CookieStore, opts *AuthenticatorOptions) (*Authenticator, error) {
	oidcProvider, err := oidc.NewProvider(context.Background(), "https://"+opts.AuthDomain+"/")
	if err != nil {
		return nil, err
	}

	// Auth callback URL is fixed. See RegisterEndpoints.
	redirectURL, err := url.JoinPath(opts.ExternalURL, "/auth/callback")
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     opts.AuthClientID,
		ClientSecret: opts.AuthClientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	a := &Authenticator{
		logger:  logger,
		admin:   adm,
		cookies: cookies,
		opts:    opts,
		oidc:    oidcProvider,
		oauth2:  oauth2Config,
	}

	return a, nil
}
