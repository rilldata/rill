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
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	cookieName                  = "auth"
	cookieFieldState            = "state"
	cookieFieldRedirect         = "redirect"
	cookieFieldCustomDomainFlow = "custom_domain_flow"
	cookieFieldAccessToken      = "access_token"
)

// RegisterEndpoints adds HTTP endpoints for auth.
// The mux must be served on the ExternalURL of the Authenticator since the logic in these handlers relies on knowing the full external URIs.
// Note that these are not gRPC handlers, just regular HTTP endpoints that we mount on the gRPC-gateway mux.
//
// Since we support optional custom domains for orgs, we need to jump through some hoops to ensure the auth redirects work correctly,
// and that the auth token cookie is set on the correct domain. Below follows an overview of the redirect flows for login and logout.
// Login flow without custom domain configured:
//
//	1 . UI requests <canonical-domain>/auth/login?redirect=<redirect>
//	   - Set a cookie with state=YYY
//	   - If a redirect URL is provided, store it in the cookie
//
//	2. Redirect to Auth0 /login?state=YYY
//	   - User completes login
//
//	3. Redirect to <canonical-domain>/auth/callback?state=YYY&<oauth-params>
//	   - Verify cookie state matches YYY
//	   - Save the long-lived token in a cookie
//	   - Redirect to the saved redirect URL (or the frontend root if none was provided)
//
// Login flow when a custom domain is configured:
//
//  1. UI requests <custom-domain>/auth/login?redirect=<redirect>
//     - Set a cookie with state=XXX
//     - If a redirect URL is provided, store it in the cookie
//
//  2. Redirect to <canonical-domain>/auth/login?redirect=https://<custom-domain>/auth/custom-domain-callback?state=XXX&custom_domain_flow=true
//     - Set a cookie with state=YYY and custom_domain_flow=true
//
//  3. Redirect to Auth0 /login?state=YYY
//     - User completes login
//
//  4. Redirect to <canonical-domain>/auth/callback?state=YYY&<oauth-params>
//     - Verify cookie state matches YYY
//     - Issue a noune token
//
//  5. Redirect to <custom-domain>/auth/custom-domain-callback?state=XXX&noune_token=<short-lived-token>
//     - Verify cookie state matches XXX
//     - Exchange short-lived token for a long-lived token
//     - Save the long-lived token in a cookie
//     - Redirect to the saved redirect URL (or the frontend root if none was provided)
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
func (a *Authenticator) RegisterEndpoints(mux *http.ServeMux, limiter ratelimit.Limiter, issuer *auth.Issuer) {
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
	observability.MuxHandle(inner, "/auth/custom-domain-callback", middleware.Check(checkLimit("/auth/custom-domain-callback"), http.HandlerFunc(a.authLoginCustomDomainCallback)))
	observability.MuxHandle(inner, "/auth/logout", middleware.Check(checkLimit("/auth/logout"), http.HandlerFunc(a.authLogout)))
	observability.MuxHandle(inner, "/auth/logout/provider", middleware.Check(checkLimit("/auth/logout/provider"), http.HandlerFunc(a.authLogoutProvider)))
	observability.MuxHandle(inner, "/auth/logout/callback", middleware.Check(checkLimit("/auth/logout/callback"), http.HandlerFunc(a.authLogoutCallback)))
	observability.MuxHandle(inner, "/auth/assume-open", a.HTTPMiddlewareLenient(middleware.Check(checkLimit("/auth/assume-open"), http.HandlerFunc(a.authAssumeOpen)))) // NOTE: Uses auth middleware
	observability.MuxHandle(inner, "/auth/oauth/device_authorization", middleware.Check(checkLimit("/auth/oauth/device_authorization"), http.HandlerFunc(a.handleDeviceCodeRequest)))
	observability.MuxHandle(inner, "/auth/oauth/device", a.HTTPMiddleware(middleware.Check(checkLimit("/auth/oauth/device"), http.HandlerFunc(a.handleUserCodeConfirmation))))   // NOTE: Uses auth middleware
	observability.MuxHandle(inner, "/auth/oauth/authorize", a.HTTPMiddleware(middleware.Check(checkLimit("/auth/oauth/authorize"), http.HandlerFunc(a.handleAuthorizeRequest)))) // NOTE: Uses auth middleware
	observability.MuxHandle(inner, "/auth/oauth/token", middleware.Check(checkLimit("/auth/oauth/token"), http.HandlerFunc(a.getAccessToken)))
	observability.MuxHandle(inner, "/auth/oauth/register", middleware.Check(checkLimit("/auth/oauth/register"), http.HandlerFunc(a.handleOAuthRegister)))
	mux.Handle("/auth/", observability.Middleware("admin", a.logger, inner))
	// Register well known endpoints
	wellKnownMux := http.NewServeMux()
	// Serve public JWKS for runtime JWT verification
	wellKnownMux.Handle("/.well-known/jwks.json", issuer.WellKnownHandler())
	// OAuth discovery endpoints for MCP support
	observability.MuxHandle(wellKnownMux, "/.well-known/oauth-protected-resource", middleware.Check(checkLimit("/.well-known/oauth-protected-resource"), http.HandlerFunc(a.handleOAuthProtectedResourceMetadata)))
	observability.MuxHandle(wellKnownMux, "/.well-known/oauth-authorization-server", middleware.Check(checkLimit("/.well-known/oauth-authorization-server"), http.HandlerFunc(a.handleOAuthAuthorizationServerMetadata)))
	mux.Handle("/.well-known/", observability.Middleware("admin", a.logger, wellKnownMux))
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

// authStart begins an OAuth/OIDC flow when called on the canonical domain, redirecting the user
// to the authentication provider. After authentication, the user is redirected back to authLoginCallback.
//
// At the end of authLoginCallback, the user is redirected to the root of the frontend URL
// (configured in admin.New). This destination can be overridden by providing a "redirect"
// query parameter to authStart.
//
// If an org has a custom domain configured (see the type doc on admin.URLs for details),
// authStart stores the state and redirect URL in a cookie, then redirects the user to
// authStart on the canonical domain.
func (a *Authenticator) authStart(w http.ResponseWriter, r *http.Request, signup bool) {
	// Generate random state for CSRF
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
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

	// If this is part of the custom domain login flow, save that info in the cookie since we need that info when handling the auth callback.
	customDomainFlow := r.URL.Query().Get("custom_domain_flow")
	if b, err := strconv.ParseBool(customDomainFlow); err == nil && b {
		sess.Values[cookieFieldCustomDomainFlow] = true
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	// Redirect to <canonical-domain>/auth/login (custom domain flow)
	host := originalHost(r)
	if a.admin.URLs.IsCustomDomain(host) {
		customCallbackURL := a.admin.URLs.WithCustomDomain(host).AuthCustomDomainCallback(state)
		canonicalLoginURL := a.admin.URLs.AuthLogin(customCallbackURL, true)
		http.Redirect(w, r, canonicalLoginURL, http.StatusTemporaryRedirect)
		return
	}

	// Redirect to auth provider (canonical domain flow)
	redirectURL := a.oauth2.AuthCodeURL(state)
	if signup {
		// Set custom parameters for signup using AuthCodeOption
		customOption := oauth2.SetAuthURLParam("screen_hint", "signup")
		redirectURL = a.oauth2.AuthCodeURL(state, customOption)
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// authLoginCallback is invoked after the user has successfully authenticated with the auth provider.
// See authStart for details on how the flow is initiated.
//
// authLoginCallback validates the OAuth response, retrieves the user’s profile information, and
// creates or updates the user in the database.
//
// If a custom domain is configured, authLoginCallback issues a short-lived token and redirects
// the user to authLoginCustomDomainCallback with that token. Otherwise, it issues a new user
// auth token, saves it in a cookie, and redirects the user to the location specified in the
// initial authLogin request.
//
// authLoginCallback does not set the auth cookie directly in the custom domain case, since the
// cookie must be set on the custom domain rather than the canonical domain. Instead, it creates
// a nonce token and redirects the user to authLoginCustomDomainCallback on the custom domain.
// That handler sets the actual auth cookie and performs the final redirect back to the UI
// (either to the saved redirect location or to the frontend root).
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
		redirect = a.admin.URLs.Frontend()
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

	// If it's part of a custom domain login flow, redirect back to the custom domain with a short-lived access token for the user.
	customDomainFlow, ok := sess.Values[cookieFieldCustomDomainFlow].(bool)
	delete(sess.Values, cookieFieldCustomDomainFlow)
	if ok && customDomainFlow {
		// save the cookies
		if err := sess.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Issue a short-lived nonce token (2-minute TTL) for browser auth callback
		ttl := 2 * time.Minute
		authNonceToken, err := a.admin.IssueUserAuthToken(r.Context(), user.ID, database.AuthClientIDRillWeb, "Nonce Token", nil, &ttl, false)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to issue API token: %s", err), http.StatusInternalServerError)
			return
		}
		redirectWithNonceToken := urlutil.MustWithQuery(redirect, map[string]string{"nonce_token": authNonceToken.Token().String()})
		http.Redirect(w, r, redirectWithNonceToken, http.StatusTemporaryRedirect)
		return
	}

	// Issue a new persistent auth token
	authToken, err := a.admin.IssueUserAuthToken(r.Context(), user.ID, database.AuthClientIDRillWeb, "Browser session", nil, nil, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to issue API token: %s", err), http.StatusInternalServerError)
		return
	}
	newAuthToken := authToken.Token().String()

	// If there's already a token in the cookie, and it's not the same one, revoke it (since we're now setting a new one).
	oldAuthToken, ok := sess.Values[cookieFieldAccessToken].(string)
	if ok && oldAuthToken != "" && oldAuthToken != newAuthToken {
		err := a.admin.RevokeAuthToken(r.Context(), oldAuthToken)
		if err != nil {
			a.logger.Info("failed to revoke old user auth token during new auth", zap.Error(err), observability.ZapCtx(r.Context()))
			// The old token was probably manually revoked. We can still continue.
		}
	}

	// Set auth token in cookie
	sess.Values[cookieFieldAccessToken] = newAuthToken

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

// authLoginCustomDomainCallback first verifies the state for CSRF protection, then extracts
// a nonce token from the query parameters and validates it. If valid, it issues a new
// long-lived token, stores it in a cookie, and redirects the user to the frontend.
//
// The redirect destination can be overridden by providing a "redirect" query parameter.
func (a *Authenticator) authLoginCustomDomainCallback(w http.ResponseWriter, r *http.Request) {
	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)

	// Check that random state matches (for CSRF protection)
	if r.URL.Query().Get("state") != sess.Values[cookieFieldState] {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	delete(sess.Values, cookieFieldState)

	// Get redirect destination
	redirect, ok := sess.Values[cookieFieldRedirect].(string)
	if !ok || redirect == "" {
		redirect = a.admin.URLs.Frontend()
	}
	delete(sess.Values, cookieFieldRedirect)

	nonceToken := r.URL.Query().Get("nonce_token")
	if nonceToken == "" {
		http.Error(w, "token not provided", http.StatusBadRequest)
		return
	}
	validated, err := a.admin.ValidateAuthToken(r.Context(), nonceToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	newAuthToken, err := a.admin.IssueUserAuthToken(r.Context(), validated.OwnerID(), database.AuthClientIDRillWeb, "Browser session", nil, nil, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to issue API token: %s", err), http.StatusInternalServerError)
		return
	}
	newToken := newAuthToken.Token().String()

	err = a.admin.RevokeAuthToken(r.Context(), nonceToken)
	if err != nil {
		a.logger.Info("failed to revoke nonce token during auth", zap.Error(err), observability.ZapCtx(r.Context()))
		// The nonce token was probably manually revoked. We can still continue.
	}

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

// authAssumeOpen allows a superuser to assume the identity of another user for support purposes.
// It checks if the current user is a superuser, then updates the auth cookie to contain a token that represents the target user.
func (a *Authenticator) authAssumeOpen(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := GetClaims(ctx)

	// If the user is not authenticated, redirect to login.
	// NOTE: Since we use the HTTPMiddlewareLenient middleware, this also handles users with expired tokens (e.g. if you're assuming different users in a row).
	if claims.OwnerType() == OwnerTypeAnon {
		http.Redirect(w, r, a.admin.URLs.AuthLogin(r.URL.String(), false), http.StatusTemporaryRedirect)
		return
	}

	// If the caller is not a user, error out.
	if claims.OwnerType() != OwnerTypeUser {
		http.Error(w, "not authenticated as a user", http.StatusBadRequest)
		return
	}

	// Grab the DB model for the current token. We need to do some deeper checks.
	tokenMdl, ok := claims.AuthTokenModel().(*database.UserAuthToken)
	if !ok {
		http.Error(w, "invalid user auth token model", http.StatusBadRequest)
		return
	}

	// If you switch between assumed users without un-assuming first, this call may be made with a representative token.
	// Let's handle that gracefully by logging you back in as yourself first.
	if tokenMdl.RepresentingUserID != nil {
		http.Redirect(w, r, a.admin.URLs.AuthLogin(r.URL.String(), false), http.StatusTemporaryRedirect)
		return
	}

	// Only superuser can do assume open
	if !claims.Superuser(ctx) {
		http.Error(w, "not authorized: only superusers can assume another user", http.StatusUnauthorized)
		return
	}

	// Validate the user to represent
	representEmail := r.URL.Query().Get("representing_user")
	if representEmail == "" {
		http.Error(w, "representing user not provided", http.StatusBadRequest)
		return
	}
	u, err := a.admin.DB.FindUserByEmail(ctx, representEmail)
	if err != nil {
		http.Error(w, fmt.Sprintf("user with email %q not found", representEmail), http.StatusBadRequest)
		return
	}
	if u.ID == tokenMdl.UserID {
		http.Error(w, "representing user cannot represent yourself", http.StatusBadRequest)
		return
	}
	representingUserID := &u.ID

	// Parse the TTL for the representative token (if any)
	ttlMinutesStr := r.URL.Query().Get("ttl_minutes")
	var ttl *time.Duration
	if ttlMinutesStr != "" {
		mins, err := strconv.Atoi(ttlMinutesStr)
		if err != nil || mins <= 0 {
			http.Error(w, "invalid ttl_minutes parameter", http.StatusBadRequest)
			return
		}
		d := time.Duration(mins) * time.Minute
		ttl = &d
	}

	// Issue a new token for the representing user.
	// We use tokenMdl.UserID here instead of claims.OwnerID() because OwnerID() could return the representing user's ID if the token is already an assumed token.
	newAuthToken, err := a.admin.IssueUserAuthToken(r.Context(), tokenMdl.UserID, database.AuthClientIDRillSupport, fmt.Sprintf("Support for %s", representEmail), representingUserID, ttl, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to issue API token: %s", err), http.StatusInternalServerError)
		return
	}
	newToken := newAuthToken.Token().String()

	// If there's already a token in the cookie, and it's not the same one, revoke it (since we're now setting a new one).
	sess := a.cookies.Get(r, cookieName)
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
	http.Redirect(w, r, a.admin.URLs.Frontend(), http.StatusTemporaryRedirect)
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
	if grantType == deviceCodeGrantType {
		a.getAccessTokenForDeviceCode(w, r, values)
	} else if grantType == authorizationCodeGrantType {
		a.getAccessTokenForAuthorizationCode(w, r, values)
	} else if grantType == refreshTokenGrantType {
		a.getAccessTokenForRefreshToken(w, r, values)
	} else {
		http.Error(w, fmt.Sprintf("unexpected grant_type: %q", grantType), http.StatusBadRequest)
		return
	}
}

func originalHost(r *http.Request) string {
	if xfHost := r.Header.Get("Rill-Custom-Domain"); xfHost != "" {
		return xfHost
	}
	return r.Host
}
