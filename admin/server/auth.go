package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authCookieName = "auth"

// authLogin starts an OAuth and OIDC flow that redirects the user for authentication with the auth provider.
// After auth, the user is redirected back to authLoginCallback, which in turn will redirect the user to "/".
// You can override the redirect destination by passing a `?redirect=URI` query to this endpoint.
// The implementation was derived from: https://auth0.com/docs/quickstart/webapp/golang/01-login.
func (s *Server) authLogin(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// Generate random state for CSRF
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Get auth cookie
	sess, err := s.cookies.Get(r, authCookieName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	// Set state in cookie
	sess.Values["state"] = state

	// Set redirect URL in cookie to enable custom redirects after auth has completed
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		sess.Values["redirect"] = redirect
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	// Redirect to auth provider
	http.Redirect(w, r, s.oauth2.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// authLoginCallback is called after the user has successfully authenticated with the auth provider.
// It validates the OAuth info, fetches user profile info, and creates/updates the user in our DB.
// It then issues a new user auth token and saves it in a cookie.
func (s *Server) authLoginCallback(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// Get auth cookie
	sess, err := s.cookies.Get(r, authCookieName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	// Check that random state matches (for CSRF protection)
	if r.URL.Query().Get("state") != sess.Values["state"] {
		http.Error(w, fmt.Sprintf("Invalid state parameter: %s", err), http.StatusBadRequest)
		return
	}
	delete(sess.Values, "state")

	// Exchange authorization code for an oauth2 token
	oauthToken, err := s.oauth2.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert an authorization code into a token: %s", err), http.StatusUnauthorized)
		return
	}

	// Extract and verify ID token (which contains the user's identity info)
	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token field in oauth2 token", http.StatusUnauthorized)
		return
	}

	oidcConfig := &oidc.Config{
		ClientID: s.oauth2.ClientID,
	}

	idToken, err := s.oidc.Verifier(oidcConfig).Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify ID Token: %s", err), http.StatusInternalServerError)
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

	// Create (or update) user in our DB
	user, err := s.admin.CreateOrUpdateUser(r.Context(), email, name, photoURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// If there's already a token in the cookie, revoke it (not stopping on errors)
	oldAuthToken, ok := sess.Values["access_token"].(string)
	if ok && oldAuthToken != "" {
		err := s.admin.RevokeAuthToken(r.Context(), oldAuthToken)
		if err != nil {
			s.logger.Error("failed to revoke old user auth token during new auth", zap.Error(err))
		}
	}

	// Issue a new persistent auth token
	authToken, err := s.admin.IssueUserAuthToken(r.Context(), user.ID, database.RillWebClientID, "Browser session")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to issue API token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Set auth token in cookie
	sess.Values["access_token"] = authToken.Token().String()

	// Get redirect destination
	redirect, ok := sess.Values["redirect"].(string)
	if !ok || redirect == "" {
		redirect = "/"
	}
	delete(sess.Values, "redirect")

	// Update cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI (usually)
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

// authLogout implements user logout. It revokes the current user auth token, then redirects to the auth providers logout flow.
// Once the logout has completed, the auth provider will redirect the user to authLogoutCallback.
func (s *Server) authLogout(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// Get auth cookie
	sess, err := s.cookies.Get(r, authCookieName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	// Revoke access token and clear in cookie
	authToken, ok := sess.Values["access_token"].(string)
	if ok && authToken != "" {
		err := s.admin.RevokeAuthToken(r.Context(), authToken)
		if err != nil {
			s.logger.Error("failed to revoke user auth token during logout", zap.Error(err))
		}
	}
	delete(sess.Values, "access_token")

	// Set redirect URL in cookie to enable custom redirects after logout
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		sess.Values["redirect"] = redirect
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	// Build auth provider logout URL
	logoutURL, err := url.Parse("https://" + s.conf.AuthDomain + "/v2/logout")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build callback endpoint for authLogoutCallback
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	returnTo, err := url.Parse(scheme + "://" + r.Host + "/auth/logout/callback")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to auth provider's logout
	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", s.conf.AuthClientID)
	logoutURL.RawQuery = parameters.Encode()

	http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
}

// authLogoutCallback is called when a logout flow iniated by authLogout has completed.
func (s *Server) authLogoutCallback(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// Get auth cookie
	sess, err := s.cookies.Get(r, authCookieName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	// Get redirect destination
	redirect, ok := sess.Values["redirect"].(string)
	if !ok || redirect == "" {
		redirect = "/"
	}
	delete(sess.Values, "redirect")

	// Save updated cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI (usually)
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

// Claims resolves permissions for a requester.
type Claims interface {
	Subject() string
}

// claimsContextKey is used to set and get Claims on a request context.
type claimsContextKey struct{}

// GetClaims retrieves Claims from a request context.
// It should only be used in handlers intercepted by UnaryServerInterceptor or StreamServerInterceptor.
func GetClaims(ctx context.Context) Claims {
	claims, ok := ctx.Value(claimsContextKey{}).(Claims)
	if !ok {
		return nil
	}

	return claims
}

// anonClaims represents claims for an unauthenticated user.
type anonClaims struct{}

func (c anonClaims) Subject() string {
	return ""
}

// tokenClaims represents claims for an auth token.
type tokenClaims struct {
	token admin.AuthToken
}

func (c *tokenClaims) Subject() string {
	return c.token.OwnerID()
}

// CookieAuthAnnotator is a gRPC-gateway annotator that moves access tokens in HTTP cookies to the "authorization" gRPC metadata.
func (s *Server) CookieAuthAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	// Get auth cookie
	sess, err := s.cookies.Get(r, authCookieName)
	if err != nil {
		return metadata.Pairs()
	}

	token, ok := sess.Values["access_token"].(string)
	if ok && token != "" {
		return metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
	}

	return metadata.Pairs()
}

// AuthUnaryServerInterceptor is a middleware for setting claims on runtime server requests.
// The assigned claims can be retrieved using GetClaims. If the interceptor succeeds, a Claims value is guaranteed to be set on the ctx.
func (s *Server) AuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		authHeader := metautils.ExtractIncoming(ctx).Get("authorization")
		newCtx, err := s.parseClaimsFromBearer(ctx, authHeader)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return handler(newCtx, req)
	}
}

// AuthStreamServerInterceptor is the streaming variant of UnaryServerInterceptor.
func (s *Server) AuthStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		authHeader := metautils.ExtractIncoming(ss.Context()).Get("authorization")
		newCtx, err := s.parseClaimsFromBearer(ss.Context(), authHeader)
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}

		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx

		return handler(srv, wrapped)
	}
}

// AuthHTTPMiddleware is a HTTP middleware variant of UnaryServerInterceptor.
// It additionally supports reading access tokens from cookies.
// It should be used for non-gRPC HTTP endpoints (CookieAuthAnnotator takes care of handling cookies in gRPC-gateway requests).
func (s *Server) AuthHTTPMiddleware(next gateway.HandlerFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		// Handle authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			newCtx, err := s.parseClaimsFromBearer(r.Context(), authHeader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			next(w, r.WithContext(newCtx), pathParams)
			return
		}

		// There was no authorization header. Try the cookie.
		sess, err := s.cookies.Get(r, authCookieName)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
			return
		}

		// Read access token from cookie
		authToken, ok := sess.Values["access_token"].(string)
		if ok && authToken != "" {
			newCtx, err := s.parseClaimsFromToken(r.Context(), authToken)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			next(w, r.WithContext(newCtx), pathParams)
			return
		}

		// No token was found. Set anonClaims.
		newCtx := context.WithValue(r.Context(), claimsContextKey{}, anonClaims{})
		next(w, r.WithContext(newCtx), pathParams)
	}
}

func (s *Server) parseClaimsFromBearer(ctx context.Context, authorizationHeader string) (context.Context, error) {
	// If authorization header is not set, we set anonClaims.
	if authorizationHeader == "" {
		ctx = context.WithValue(ctx, claimsContextKey{}, anonClaims{})
		return ctx, nil
	}

	// Extract bearer token
	bearerToken := ""
	if len(authorizationHeader) >= 6 && strings.EqualFold(authorizationHeader[0:6], "bearer") {
		bearerToken = strings.TrimSpace(authorizationHeader[6:])
	}
	if bearerToken == "" {
		return nil, errors.New("no bearer token found in authorization header")
	}

	return s.parseClaimsFromToken(ctx, bearerToken)
}

func (s *Server) parseClaimsFromToken(ctx context.Context, token string) (context.Context, error) {
	validated, err := s.admin.ValidateAuthToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Create claims
	claims := &tokenClaims{
		token: validated,
	}

	ctx = context.WithValue(ctx, claimsContextKey{}, claims)
	return ctx, nil
}
