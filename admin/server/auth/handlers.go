package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	cookieName             = "auth"
	cookieFieldState       = "state"
	cookieFieldRedirect    = "redirect"
	cookieFieldAccessToken = "access_token"
)

// RegisterEndpoints adds HTTP endpoints for auth.
// The mux must be served on the ExternalURL of the Authenticator since the logic in these handlers relies on knowing the full external URIs.
// Note that these are not gRPC handlers, just regular HTTP endpoints that we mount on the gRPC-gateway mux.
//
// Since we support optional custom domains for orgs, we need to jump through some hoops to ensure the auth redirects work correctly,
// and that the auth token cookie is set on the correct domain. Below follows an overview of the redirect flows for login and logout.

// Login flow:
//  1. Frontend calls <canonical domain>/auth/login?redirect=<frontend return URL>
//  2. It redirects to <auth provider> for login
//  3. The auth provider redirects to <canonical domain>/auth/callback
//  4. It redirects to <custom domain>/auth/with-token
//  5. It redirects to <frontend return URL>
//
// Logout flow:
//  1. Frontend calls <custom domain>/auth/logout?redirect=<frontend return URL>
//  2. It redirects to <canonical domain>/auth/logout/provider?redirect=<frontend return URL>
//  3. It redirects to <auth provider> for logout
//  4. The auth provider redirects to <canonical domain>/auth/logout/callback
//  5. It redirects to <frontend return URL>
//
// The "canonical domain" is the Rill-managed external URL of the current service (e.g. "admin.rilldata.com").
// The "custom domain" is the custom domain of the current org with path suffix for the admin service (e.g. "myorg.com/api").
// If the current org doesn't have a custom domain, the custom domain can be substituted for the canonical domain (without path suffix).
//
// There are more details in the type doc for `admin.URLs` and in the individual handler docstrings below.
func (a *Authenticator) RegisterEndpoints(mux *http.ServeMux, limiter ratelimit.Limiter) {
	// checkLimit needs access to limiter
	checkLimit := func(route string) middleware.CheckFunc {
		return func(req *http.Request) error {
			claims := GetClaims(req.Context())
			if claims == nil || claims.OwnerType() == OwnerTypeAnon {
				limitKey := ratelimit.AnonLimitKey(route, observability.HTTPPeer(req))
				if err := limiter.Limit(req.Context(), limitKey, ratelimit.Sensitive); err != nil {
					if errors.As(err, &ratelimit.QuotaExceededError{}) {
						return httputil.Error(http.StatusTooManyRequests, err)
					}
					return err
				}
			}
			return nil
		}
	}

	// TODO: Add helper utils to clean this up
	inner := http.NewServeMux()
	observability.MuxHandle(inner, "/auth/signup", middleware.Check(checkLimit("/auth/signup"), http.HandlerFunc(a.authSignup)))
	observability.MuxHandle(inner, "/auth/login", middleware.Check(checkLimit("/auth/login"), http.HandlerFunc(a.authLogin)))
	observability.MuxHandle(inner, "/auth/callback", middleware.Check(checkLimit("/auth/callback"), http.HandlerFunc(a.authLoginCallback)))
	observability.MuxHandle(inner, "/auth/with-token", middleware.Check(checkLimit("/auth/with-token"), http.HandlerFunc(a.authWithToken)))
	observability.MuxHandle(inner, "/auth/logout", middleware.Check(checkLimit("/auth/logout"), http.HandlerFunc(a.authLogout)))
	observability.MuxHandle(inner, "/auth/logout/provider", middleware.Check(checkLimit("/auth/logout/provider"), http.HandlerFunc(a.authLogoutProvider)))
	observability.MuxHandle(inner, "/auth/logout/callback", middleware.Check(checkLimit("/auth/logout/callback"), http.HandlerFunc(a.authLogoutCallback)))
	observability.MuxHandle(inner, "/auth/oauth/device_authorization", middleware.Check(checkLimit("/auth/oauth/device_authorization"), http.HandlerFunc(a.handleDeviceCodeRequest)))
	observability.MuxHandle(inner, "/auth/oauth/device", a.HTTPMiddleware(middleware.Check(checkLimit("/auth/oauth/device"), http.HandlerFunc(a.handleUserCodeConfirmation))))   // NOTE: Uses auth middleware
	observability.MuxHandle(inner, "/auth/oauth/authorize", a.HTTPMiddleware(middleware.Check(checkLimit("/auth/oauth/authorize"), http.HandlerFunc(a.handleAuthorizeRequest)))) // NOTE: Uses auth middleware
	observability.MuxHandle(inner, "/auth/oauth/token", middleware.Check(checkLimit("/auth/oauth/token"), http.HandlerFunc(a.getAccessToken)))
	mux.Handle("/auth/", observability.Middleware("admin", a.logger, inner))
}

// authSignup redirects the users to the auth provider's signup page.
// See authStart for details.
func (a *Authenticator) authSignup(w http.ResponseWriter, r *http.Request) {
	a.authStart(w, r, true)
}

// authLogin redirects the users to the auth provider's login page.
// See authStart for details.
func (a *Authenticator) authLogin(w http.ResponseWriter, r *http.Request) {
	a.authStart(w, r, false)
}

// authStart starts an OAuth and OIDC flow that redirects the user for authentication with the auth provider.
// After auth, the user is redirected back to authLoginCallback.
//
// At the end of authLoginCallback, the user will be redirected to the root of the frontend URL (configured in admin.New).
// The redirect destination can be overridden by passing a "redirect" query parameter to authStart.
//
// If an org has a custom domain configured (see the type doc on admin.URLs for details),
// the frontend must still use the primary domain when redirecting to authStart (e.g. "admin.rilldata.com/auth/login").
// This is because the auth service will always redirect back to a fixed callback URL on the primary domain,
// and we need the cookies set in this handler to be available there.
//
// For orgs with a custom domain configured, to eventually redirect the user back to the custom domain,
// the frontend should set the "redirect" query parameter to a full URL containing the custom domain URL.
func (a *Authenticator) authStart(w http.ResponseWriter, r *http.Request, signup bool) {
	// Generate random state for CSRF
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)

	// Set state in cookie
	sess.Values[cookieFieldState] = state

	// Set redirect URL in cookie to enable custom redirects after auth has completed
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		sess.Values[cookieFieldRedirect] = redirect
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	// Redirect to auth provider
	redirectURL := a.oauth2.AuthCodeURL(state)
	if signup {
		// Set custom parameters for signup using AuthCodeOption
		customOption := oauth2.SetAuthURLParam("screen_hint", "signup")
		redirectURL = a.oauth2.AuthCodeURL(state, customOption)
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// authLoginCallback is called after the user has successfully authenticated with the auth provider.
// See authStart for details about how the flow is initiated.
//
// authLoginCallback validates the OAuth info, fetches user profile info, and creates/updates the user in our DB.
// It then issues a new user auth token and saves it in a cookie.
// Finally, it redirects the user to the location specified in the initial call to authLogin.
//
// authLoginCallback doesn't set the auth cookie directly because the final redirect destination may be an org with a custom domain.
// So we need to ensure that the auth token is set in a cookie on the custom domain instead of the primary domain.
// So after creating a new token, it redirects to authWithToken on the same domain as the final redirect destination (if any, else it stays on the same domain).
// authWithToken will then set the actual auth cookie and do the final redirect back to the UI.
func (a *Authenticator) authLoginCallback(w http.ResponseWriter, r *http.Request) {
	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)

	// Check that random state matches (for CSRF protection)
	if r.URL.Query().Get("state") != sess.Values[cookieFieldState] {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	delete(sess.Values, cookieFieldState)

	// Check for errors in the auth flow
	if errStr := r.URL.Query().Get("error"); errStr != "" {
		description := r.URL.Query().Get("error_description")
		http.Error(w, fmt.Sprintf("auth error of type %q: %s", errStr, description), http.StatusUnauthorized)
		return
	}

	// Exchange authorization code for an oauth2 token
	oauthToken, err := a.oauth2.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to convert authorization code into a token: %s", err), http.StatusUnauthorized)
		return
	}

	// Extract and verify ID token (which contains the user's identity info)
	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token field in oauth2 token", http.StatusUnauthorized)
		return
	}

	oidcConfig := &oidc.Config{
		ClientID: a.oauth2.ClientID,
	}

	idToken, err := a.oidc.Verifier(oidcConfig).Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify ID token: %s", err), http.StatusInternalServerError)
		return
	}

	// Extract user profile information
	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	email, ok := profile["email"].(string)
	if !ok || email == "" {
		http.Error(w, "claim 'email' not found", http.StatusInternalServerError)
		return
	}
	emailVerified, ok := profile["email_verified"].(bool)
	if !ok {
		// For SAML flows, it is passed as a string
		emailVerifiedStr, ok := profile["email_verified"].(string)
		if !ok {
			http.Error(w, "claim 'email_verified' not found", http.StatusInternalServerError)
			return
		}
		emailVerified, err = strconv.ParseBool(emailVerifiedStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("claim 'email_verified' could not be parsed as a boolean (got %q)", emailVerifiedStr), http.StatusInternalServerError)
			return
		}
	}
	name, ok := profile["name"].(string)
	if !ok {
		http.Error(w, "claim 'name' not found", http.StatusInternalServerError)
		return
	}
	photoURL, ok := profile["picture"].(string)
	if !ok {
		http.Error(w, "claim 'picture' not found", http.StatusInternalServerError)
		return
	}

	// Get redirect destination
	redirect, ok := sess.Values[cookieFieldRedirect].(string)
	if !ok || redirect == "" {
		redirect = "/"
	}
	delete(sess.Values, cookieFieldRedirect)

	// Check that the user's email is verified
	if !emailVerified {
		redirectURL := a.admin.URLs.WithCustomDomainFromRedirectURL(redirect).AuthVerifyEmailUI()
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Create (or update) user in our DB
	user, err := a.admin.CreateOrUpdateUser(r.Context(), email, name, photoURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update user: %s", err), http.StatusInternalServerError)
		return
	}

	// Issue a new persistent auth token
	authToken, err := a.admin.IssueUserAuthToken(r.Context(), user.ID, database.AuthClientIDRillWeb, "Browser session", nil, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to issue API token: %s", err), http.StatusInternalServerError)
		return
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to authWithToken to set the token in a cookie.
	// We don't set the cookie directly here because the auth flow may have started from an org with a custom domain (see authStart),
	// in which case the token needs to be set in a cookie on the custom domain instead of the primary domain (where authLoginCallback is called).
	tokenStr := authToken.Token().String()
	withTokenURL := a.admin.URLs.WithCustomDomainFromRedirectURL(redirect).AuthWithToken(tokenStr, redirect)
	http.Redirect(w, r, withTokenURL, http.StatusTemporaryRedirect)
}

// authWithToken extracts an auth token from the query params and sets it in the auth cookie.
// It then redirects the user to the frontend. The redirect destination can be overridden by passing a "redirect" query parameter.
func (a *Authenticator) authWithToken(w http.ResponseWriter, r *http.Request) {
	// Get new auth token
	newToken := r.URL.Query().Get("token")
	if newToken == "" {
		http.Error(w, "token not provided", http.StatusBadRequest)
		return
	}

	// Get redirect destination after setting the token in the cookie
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = a.admin.URLs.Frontend()
	}

	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)

	// If there's already a token in the cookie, and it's not the same one, revoke it (since we're now setting a new one).
	oldAuthToken, ok := sess.Values[cookieFieldAccessToken].(string)
	if ok && oldAuthToken != "" && oldAuthToken != newToken {
		err := a.admin.RevokeAuthToken(r.Context(), oldAuthToken)
		if err != nil {
			a.logger.Info("failed to revoke old user auth token during new auth", zap.Error(err), observability.ZapCtx(r.Context()))
			// The old token was probably manually revoked. We can still continue.
		}
	}

	// Set auth token in cookie
	sess.Values[cookieFieldAccessToken] = newToken

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

// authLogout implements user logout. It revokes the current user auth token, then redirects to the auth provider's logout flow.
// Once the logout has completed, the auth provider will redirect the user to authLogoutCallback.
//
// For orgs with a custom domain configured, the frontend should call authLogout on the custom domain's admin endpoint.
// Note that this is different from authStart, which must always be called on the canonical external domain.
// The reason logout must be called on the custom domain is that the auth token cookie must be cleared from the custom domain.
// authLogout itself will then handle the redirects with the auth provider from the canonical domain.
func (a *Authenticator) authLogout(w http.ResponseWriter, r *http.Request) {
	// Revoke access token and clear in cookie
	sess := a.cookies.Get(r, cookieName)
	authToken, ok := sess.Values[cookieFieldAccessToken].(string)
	if ok && authToken != "" {
		err := a.admin.RevokeAuthToken(r.Context(), authToken)
		if err != nil {
			a.logger.Info("failed to revoke user auth token during logout", zap.Error(err), observability.ZapCtx(r.Context()))
			// We should still continue to ensure the user is logged out on the auth provider as well.
		}
	}
	delete(sess.Values, cookieFieldAccessToken)

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	// Extract custom redirect destination (if any)
	redirect := r.URL.Query().Get("redirect")

	// Redirect to authLogoutProvider (see its docstring below for details on why we do this).
	http.Redirect(w, r, a.admin.URLs.AuthLogoutProvider(redirect), http.StatusTemporaryRedirect)
}

// authLogoutProvider redirects to the auth provider's logout flow.
// This is separated from authLogout to support orgs with custom domains where the auth token cookie must be cleared from the custom domain,
// but the redirect destination must be set in a cookie on the primary domain because the auth provider will redirect to authLogoutCallback on the primary domain.
func (a *Authenticator) authLogoutProvider(w http.ResponseWriter, r *http.Request) {
	// Set custom redirect destination in cookie for when the logout flow is over (if any)
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		// Update cookie
		sess := a.cookies.Get(r, cookieName)
		sess.Values[cookieFieldRedirect] = redirect

		// Save cookie
		if err := sess.Save(r, w); err != nil {
			http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
			return
		}
	}

	// Build and redirect to the auth provider logout URL.
	logoutURL, err := url.Parse("https://" + a.opts.AuthDomain + "/v2/logout")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	parameters := url.Values{}
	parameters.Add("returnTo", a.admin.URLs.AuthLogoutCallback())
	parameters.Add("client_id", a.opts.AuthClientID)
	logoutURL.RawQuery = parameters.Encode()
	http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
}

// authLogoutCallback is called by the auth provider when a logout flow iniated by authLogout has completed.
//
// For orgs with a custom domain configured, the auth provider will still redirect back to the canonical domain's authLogoutCallback.
// The user will then be sent back to the custom domain if the redirect destination initially provided in authLogout is on the custom domain.
func (a *Authenticator) authLogoutCallback(w http.ResponseWriter, r *http.Request) {
	// Get redirect destination
	sess := a.cookies.Get(r, cookieName)
	redirect, ok := sess.Values[cookieFieldRedirect].(string)
	if !ok || redirect == "" {
		redirect = a.admin.URLs.Frontend()
	}
	delete(sess.Values, cookieFieldRedirect)

	// Save updated cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI (usually)
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

// handleAuthorizeRequest handles the incoming OAuth2 Authorization request, if the user is not logged redirect to login, currently only PKCE based authorization code flow is supported
func (a *Authenticator) handleAuthorizeRequest(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		internalServerError(w, fmt.Errorf("did not find any claims, %w", errors.New("server error")))
		return
	}
	if claims.OwnerType() == OwnerTypeAnon {
		// not logged in, redirect to login
		// after login redirect back to same path so encode the current URL as a redirect parameter
		encodedURL := url.QueryEscape(r.URL.String())
		http.Redirect(w, r, "/auth/login?redirect="+encodedURL, http.StatusTemporaryRedirect)
	}
	if claims.OwnerType() != OwnerTypeUser {
		http.Error(w, "only users can be authorized", http.StatusBadRequest)
		return
	}
	userID := claims.OwnerID()

	// Extract necessary details from the query parameters
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")

	if clientID == "" || redirectURI == "" || responseType == "" {
		http.Error(w, "Missing required parameters - client_id or redirect_uri or response_type", http.StatusBadRequest)
		return
	}

	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")

	if codeChallenge != "" {
		if codeChallengeMethod == "" {
			http.Error(w, "Missing code challenge method", http.StatusBadRequest)
			return
		}
		if responseType != "code" {
			http.Error(w, "Invalid response type", http.StatusBadRequest)
			return
		}
		a.handlePKCE(w, r, clientID, userID, codeChallenge, codeChallengeMethod, redirectURI)
	} else {
		http.Error(w, "only PKCE based authorization code flow is supported", http.StatusBadRequest)
		return
	}
}

// getAccessToken depending on the grant_type either verifies the device code and returns an access token if the request is approved or exchanges the authorization code for an access token
func (a *Authenticator) getAccessToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected a POST request", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to read request body: %w", err))
		return
	}
	bodyStr := string(body)
	values, err := url.ParseQuery(bodyStr)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to parse query: %w", err))
		return
	}

	grantType := values.Get("grant_type")
	if !(grantType == deviceCodeGrantType || grantType == authorizationCodeGrantType) {
		http.Error(w, "invalid grant_type", http.StatusBadRequest)
		return
	}

	if grantType == deviceCodeGrantType {
		a.getAccessTokenForDeviceCode(w, r, values)
	} else {
		a.getAccessTokenForAuthorizationCode(w, r, values)
	}
}
