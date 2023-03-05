package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"github.com/rilldata/rill/admin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// AuthenticatorOptions provides options for Authenticator
type AuthenticatorOptions struct {
	AuthDomain       string
	AuthClientID     string
	AuthClientSecret string
	AuthCallbackURL  string
}

// Authenticator wraps functionality for admin server auth.
// It provides endpoints for login/logout, creates users, issues cookie-based auth tokens, and provides middleware for authenticating requests.
type Authenticator struct {
	logger  *zap.Logger
	admin   *admin.Service
	cookies *sessions.CookieStore
	opts    *AuthenticatorOptions
	oidc    *oidc.Provider
	oauth2  oauth2.Config
}

// NewAuthenticator creates an Authenticator.
func NewAuthenticator(logger *zap.Logger, adm *admin.Service, cookies *sessions.CookieStore, opts *AuthenticatorOptions) (*Authenticator, error) {
	oidcProvider, err := oidc.NewProvider(context.Background(), "https://"+opts.AuthDomain+"/")
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     opts.AuthClientID,
		ClientSecret: opts.AuthClientSecret,
		RedirectURL:  opts.AuthCallbackURL,
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
