package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mikespook/gorbac"
	"golang.org/x/oauth2"
)

const authSessionName = "auth"

// Authenticator is used to authenticate our users.
// Refereance link - https://auth0.com/docs/quickstart/webapp/golang/01-login for sample auth setup
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

type authorizer struct {
	users       Users
	rbac        *gorbac.RBAC
	permissions gorbac.Permissions
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
	logoutUrl, err := url.Parse("https://" + s.conf.AuthDomain + "/v2/logout")
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
	logoutUrl.RawQuery = parameters.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
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
	json.Unmarshal(profile, &profiles)

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

func (a *authorizer) HasPermission(userID, action, asset string) bool {
	user, ok := a.users[userID]
	if !ok {
		// Unknown userID
		log.Print("Unknown user:", userID)
		return false
	}

	for _, role := range user.Roles {
		permission := action + ":" + asset
		if a.rbac.IsGranted(role, a.permissions[permission], nil) {
			return true
		}
	}

	return false
}

func (s *Server) IsAuthorized(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		username, err := GetUserID(c)
		if err != nil {
			return sendError(c, http.StatusUnauthorized, "User not found")
		}

		asset := GetAssetName(c)
		action := actionFromMethod(c.Request().Method)

		if !s.authorizer.HasPermission(username, action, asset) {
			log.Printf("User '%s' is not allowed to '%s' resource '%s'", username, action, asset)
			c.Response().Writer.WriteHeader(http.StatusForbidden)
			return c.String(http.StatusUnauthorized, "Not Authenticated")
		}

		return next(c)
	}
}

func GetUserID(c echo.Context) (string, error) {
	authToken := c.Request().Header.Get("Authorization")
	if authToken != "" {
		tokenStr := strings.Split(authToken, " ")[1]
		token, err := jwt.Parse(tokenStr, nil)
		if token == nil {
			return "", err
		}
		claims, _ := token.Claims.(jwt.MapClaims)
		username := claims["nickname"].(string)
		return username, nil
	}

	sess, err := session.Get(authSessionName, c)
	if err != nil {
		return "", err
	}

	var profiles map[string]interface{}
	profile := sess.Values["profile"].([]byte)
	json.Unmarshal(profile, &profiles)
	username := profiles["nickname"].(string)
	return username, nil
}

// Need to check if there any better way for this i.e getting the asset names
func GetAssetName(ctx echo.Context) string {
	path := ctx.Path()
	// p:=strings.TrimSuffix(path, "/:name")
	paths := strings.Split(strings.TrimSuffix(path, "/:name"), "/")
	asset := paths[len(paths)-1]
	return asset
}

func actionFromMethod(httpMethod string) string {
	switch httpMethod {
	case "GET":
		return "read"
	case "POST":
		return "write"
	case "DELETE":
		return "delete"
	case "PUT":
		return "update"
	default:
		return ""
	}
}
