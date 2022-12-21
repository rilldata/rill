package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const authSessionName = "auth"

// Authenticator is used to authenticate our users.
// Refereance link - https://auth0.com/docs/quickstart/webapp/golang/01-login for sample auth setup.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

func newAuthenticator(c context.Context, conf Config) (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		c,
		"https://"+conf.AuthDomain+"/",
	)
	if err != nil {
		return nil, err
	}

	config := oauth2.Config{
		ClientID:     conf.AuthClientID,
		ClientSecret: conf.AuthClientSecret,
		RedirectURL:  conf.AuthCallbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   config,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func (s *Server) authLogin(c echo.Context) error {
	state, err := generateRandomState()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sess, err := session.Get(authSessionName, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sess.Values["state"] = state

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusTemporaryRedirect, s.auth.AuthCodeURL(state))
}

func (s *Server) callback(c echo.Context) error {
	sess, err := session.Get(authSessionName, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if c.QueryParam("state") != sess.Values["state"] {
		return c.String(http.StatusBadRequest, "Invalid state parameter.")
	}

	// Exchange an authorization code for a token.
	token, err := s.auth.Exchange(c.Request().Context(), c.QueryParam("code"))
	if err != nil {
		return c.String(http.StatusUnauthorized, "Failed to convert an authorization code into a token.")
	}

	idToken, err := s.auth.VerifyIDToken(c.Request().Context(), token)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to verify ID Token.")
	}

	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	profileBytes, err := json.Marshal(profile)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to Serialize profile.")
	}

	sess.Values["access_token"] = token.AccessToken
	sess.Values["profile"] = profileBytes

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	// Redirect to logged in page.
	return c.Redirect(http.StatusTemporaryRedirect, "/auth/user")
}

func (s *Server) logout(c echo.Context) error {
	logoutURL, err := url.Parse("https://" + s.conf.AuthDomain + "/v2/logout")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	scheme := "http"
	if c.Request().TLS != nil {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + c.Request().Host + "/auth/logout/callback")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", s.conf.AuthClientID)
	logoutURL.RawQuery = parameters.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, logoutURL.String())
}

func (s *Server) logoutCallback(c echo.Context) error {
	sess, err := session.Get(authSessionName, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sess.Values["access_token"] = nil
	sess.Values["profile"] = nil

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "user logged out")
}

func (s *Server) user(c echo.Context) error {
	sess, err := session.Get(authSessionName, c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var profiles map[string]interface{}
	profile := sess.Values["profile"].([]byte)
	err = json.Unmarshal(profile, &profiles)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, profiles)
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)
	return state, nil
}

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(authSessionName, c)
		if err != nil {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		if sess.Values["profile"] == nil {
			return c.String(http.StatusUnauthorized, "Not Authenticated")
		}
		return next(c)
	}
}
