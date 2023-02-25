package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
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

func (s *Server) authLogin(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	state, err := generateRandomState()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
		return
	}

	sess, err := sessions.NewCookieStore([]byte(s.conf.SessionSecret)).Get(req, authSessionName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	sess.Values["state"] = state

	if err := sess.Save(req, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, s.auth.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (s *Server) callback(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	sess, err := sessions.NewCookieStore([]byte(s.conf.SessionSecret)).Get(req, authSessionName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	if req.URL.Query().Get("state") != sess.Values["state"] {
		http.Error(w, fmt.Sprintf("Invalid state parameter: %s", err), http.StatusBadRequest)
		return
	}

	// Exchange an authorization code for a token.
	token, err := s.auth.Exchange(req.Context(), req.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert an authorization code into a token: %s", err), http.StatusUnauthorized)
		return
	}

	idToken, err := s.auth.VerifyIDToken(req.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify ID Token: %s", err), http.StatusInternalServerError)
		return
	}

	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	profileBytes, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to Serialize profile: %s", err), http.StatusInternalServerError)
		return
	}

	sess.Values["access_token"] = token.AccessToken
	sess.Values["profile"] = profileBytes

	if err := sess.Save(req, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Redirect to logged in page.
	http.Redirect(w, req, "/auth/user", http.StatusTemporaryRedirect)
}

func (s *Server) logout(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	logoutURL, err := url.Parse("https://" + s.conf.AuthDomain + "/v2/logout")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + req.Host + "/auth/logout/callback")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", s.conf.AuthClientID)
	logoutURL.RawQuery = parameters.Encode()

	http.Redirect(w, req, logoutURL.String(), http.StatusTemporaryRedirect)
}

func (s *Server) logoutCallback(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	sess, _ := sessions.NewCookieStore([]byte(s.conf.SessionSecret)).Get(req, authSessionName)

	sess.Values["access_token"] = nil
	sess.Values["profile"] = nil

	if err := sess.Save(req, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "user logged out")
}

func (s *Server) user(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	sess, _ := sessions.NewCookieStore([]byte(s.conf.SessionSecret)).Get(req, authSessionName)

	var profiles map[string]interface{}
	if sess.Values["profile"] == nil {
		http.Error(w, "Not Authenticated", http.StatusUnauthorized)
		return
	}
	profile := sess.Values["profile"].([]byte)
	err := json.Unmarshal(profile, &profiles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(profiles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func (s *Server) IsAuthenticated1(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := sessions.NewCookieStore([]byte(s.conf.SessionSecret)).Get(r, authSessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if sess.Values["profile"] == nil {
			http.Error(w, "Not Authenticated", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
