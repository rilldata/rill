package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/oauth"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"go.uber.org/zap"
)

func (a *Authenticator) handlePKCE(w http.ResponseWriter, r *http.Request, clientID, userID, codeChallenge, codeChallengeMethod, redirectURI string) {
	normalizedRedirect, err := normalizeRedirectURI(redirectURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authClient, err := a.admin.DB.FindAuthClient(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, fmt.Sprintf("invalid client_id %q", clientID), http.StatusBadRequest)
			return
		}
		internalServerError(w, fmt.Errorf("failed to lookup auth client, %w", err))
		return
	}
	if !hasGrantType(authClient.GrantTypes, authorizationCodeGrantType) {
		http.Error(w, "client is not permitted to use authorization codes", http.StatusBadRequest)
		return
	}

	if clientID != database.AuthClientIDRillWebLocal && len(authClient.RedirectURIs) == 0 {
		http.Error(w, "client has no registered redirect URIs", http.StatusBadRequest)
		return
	}

	if !isRedirectURIAllowed(authClient, normalizedRedirect) {
		http.Error(w, "invalid redirect URI", http.StatusBadRequest)
		return
	}

	// A signed-in user's browser request to an allowed callback is the consent
	// boundary for ordinary clients. Public registration cannot create privileged
	// clients, so those must have arrived through trusted provisioning.

	// Generate a unique authorization code
	code, err := generateRandomString(16) // 16 bytes, resulting in a 32-character hex string
	if err != nil {
		http.Error(w, "Failed to generate authorization code", http.StatusInternalServerError)
		return
	}

	// Set the expiration time for the authorization code to a minute from now. Note from https://www.oauth.com/oauth2-servers/authorization/the-authorization-response/ -
	// The authorization code must expire shortly after it is issued. The OAuth 2.0 spec recommends a maximum lifetime of 10 minutes, but in practice, most services set the expiration much shorter, around 30-60 seconds.
	expiration := time.Now().Add(1 * time.Minute)

	// Store the authorization code in the database
	_, err = a.admin.DB.InsertAuthorizationCode(r.Context(), code, userID, clientID, normalizedRedirect, codeChallenge, codeChallengeMethod, expiration)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to store authorization code, %w", err))
		return
	}

	// Build the redirection URI with the authorization code as per OAuth2 spec, state is URL-encoded
	redirectWithCode, err := urlutil.WithQuery(normalizedRedirect, map[string]string{
		"code":  code,
		"state": r.URL.Query().Get("state"),
	})
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to build redirect URI, %w", err))
		return
	}

	// Redirect the user agent to the redirect URI with the authorization code
	http.Redirect(w, r, redirectWithCode, http.StatusFound)
}

// getAccessTokenForAuthorizationCode exchanges an authorization code for an access token
func (a *Authenticator) getAccessTokenForAuthorizationCode(w http.ResponseWriter, r *http.Request, values url.Values) {
	// Extract the authorization code
	code := values.Get("code")
	if code == "" {
		http.Error(w, "authorization code is required", http.StatusBadRequest)
		return
	}

	// Extract the client ID
	clientID := values.Get("client_id")
	if clientID == "" {
		http.Error(w, "client ID is required", http.StatusBadRequest)
		return
	}

	// Extract the redirect URI
	redirectURI := values.Get("redirect_uri")
	if redirectURI == "" {
		http.Error(w, "redirect URI is required", http.StatusBadRequest)
		return
	}

	// Extract the code verifier
	codeVerifier := values.Get("code_verifier")
	if codeVerifier == "" {
		http.Error(w, "code verifier is required", http.StatusBadRequest)
		return
	}

	responseVersion := values.Get("token_response_version")

	// A successful exchange consumes its code and persists every token in one
	// transaction. A failed validation or persistence step rolls everything back.
	txCtx, tx, err := a.admin.DB.NewTx(r.Context(), false)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to start authorization code transaction, %w", err))
		return
	}
	defer func() { _ = tx.Rollback() }()

	authCode, err := a.admin.DB.FindAuthorizationCode(txCtx, code)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, "no such authorization code found", http.StatusBadRequest)
		} else {
			internalServerError(w, fmt.Errorf("failed to get authorization code, %w", err))
		}
		return
	}

	userID := authCode.UserID
	if userID == "" {
		http.Error(w, "no user found for authorization code", http.StatusInternalServerError)
		return
	}

	// These checks bind the code to the client, callback, lifetime, and PKCE
	// secret chosen at authorization time. Invalid attempts mint nothing and do
	// not consume the code, preventing an attacker from burning a valid code.
	if authCode.ClientID != clientID {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}

	if authCode.RedirectURI != redirectURI {
		http.Error(w, "invalid redirect URI", http.StatusBadRequest)
		return
	}

	authClient, err := a.admin.DB.FindAuthClient(txCtx, clientID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, fmt.Sprintf("invalid client_id %q", clientID), http.StatusBadRequest)
			return
		}
		internalServerError(w, fmt.Errorf("failed to lookup auth client, %w", err))
		return
	}

	if !authCode.Expiration.After(time.Now()) {
		http.Error(w, "authorization code has expired", http.StatusBadRequest)
		return
	}

	if !verifyCodeChallenge(codeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
		http.Error(w, "invalid code verifier", http.StatusBadRequest)
		return
	}

	// Deleting inside the transaction is the single-use claim. Concurrent valid
	// exchanges can both read the code, but PostgreSQL lets only one delete it;
	// the loser observes not-found and cannot persist a token.
	err = a.admin.DB.DeleteAuthorizationCode(txCtx, code)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, "no such authorization code found", http.StatusBadRequest)
		} else {
			internalServerError(w, fmt.Errorf("failed to consume authorization code, %w", err))
		}
		return
	}

	scope := authClient.Scope
	longLivedAccess := hasScope(scope, longLivedAccessTokenScope)
	refreshAllowed := hasGrantType(authClient.GrantTypes, refreshTokenGrantType)
	var respBytes []byte
	// Issue long-lived access token for clients explicitly whitelisted
	if longLivedAccess {
		authToken, err := a.admin.IssueUserAuthToken(txCtx, userID, authCode.ClientID, "", nil, nil, false)
		if err != nil {
			if errors.Is(err, r.Context().Err()) {
				http.Error(w, "request cancelled or timeout", http.StatusRequestTimeout)
				return
			}
			internalServerError(w, fmt.Errorf("failed to issue access token, %w", err))
			return
		}

		if responseVersion == "standard" {
			resp := oauth.TokenResponse{
				AccessToken: authToken.Token().String(),
				TokenType:   "Bearer",
				ExpiresIn:   0, // never expires
				Scope:       scope,
				UserID:      userID,
			}
			respBytes, err = json.Marshal(resp)
			if err != nil {
				internalServerError(w, fmt.Errorf("failed to marshal response, %w", err))
				return
			}
		} else {
			resp := oauth.LegacyTokenResponse{
				AccessToken: authToken.Token().String(),
				TokenType:   "Bearer",
				ExpiresIn:   0, // never expires
				UserID:      userID,
			}
			respBytes, err = json.Marshal(resp)
			if err != nil {
				internalServerError(w, fmt.Errorf("failed to marshal response, %w", err))
				return
			}
		}
	} else {
		// Issue access token (60 minutes TTL)
		accessTTL := 60 * time.Minute
		accessToken, err := a.admin.IssueUserAuthToken(txCtx, userID, authCode.ClientID, "Access Token", nil, &accessTTL, false)
		if err != nil {
			if errors.Is(err, r.Context().Err()) {
				http.Error(w, "request cancelled or timeout", http.StatusRequestTimeout)
				return
			}
			internalServerError(w, fmt.Errorf("failed to issue access token, %w", err))
			return
		}

		resp := oauth.TokenResponse{
			AccessToken: accessToken.Token().String(),
			ExpiresIn:   int64(accessTTL.Seconds()),
			TokenType:   "Bearer",
			Scope:       scope,
			UserID:      userID,
		}
		if refreshAllowed {
			// Issue refresh token (365 days TTL)
			refreshTTL := 365 * 24 * time.Hour
			refreshToken, err := a.admin.IssueUserAuthToken(txCtx, userID, authCode.ClientID, "Refresh Token", nil, &refreshTTL, true)
			if err != nil {
				if errors.Is(err, r.Context().Err()) {
					http.Error(w, "request cancelled or timeout", http.StatusRequestTimeout)
					return
				}
				internalServerError(w, fmt.Errorf("failed to issue refresh token, %w", err))
				return
			}

			resp.RefreshToken = refreshToken.Token().String()
		}
		respBytes, err = json.Marshal(resp)
		if err != nil {
			internalServerError(w, fmt.Errorf("failed to marshal response, %w", err))
			return
		}
	}

	// Commit is the all-or-nothing boundary: the consumed code and every access
	// or refresh token become durable together, so no failed exchange can leave
	// an orphan token behind.
	if err := tx.Commit(); err != nil {
		internalServerError(w, fmt.Errorf("failed to commit authorization code exchange, %w", err))
		return
	}
	a.admin.Used.Client(clientID)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to write response, %w", err))
		return
	}
	a.logger.Debug("Exchanged authorization code for tokens", zap.String("userID", userID), zap.String("clientID", clientID), zap.Bool("long_lived_access_token", longLivedAccess), zap.Bool("refresh_token_grant", refreshAllowed))
}

// getAccessTokenForRefreshToken exchanges a refresh token for a new access token and refresh token
func (a *Authenticator) getAccessTokenForRefreshToken(w http.ResponseWriter, r *http.Request, values url.Values) {
	// Extract the refresh token
	refreshTokenStr := values.Get("refresh_token")
	if refreshTokenStr == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	// Extract the client ID
	clientID := values.Get("client_id")
	if clientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	// Rotation is one transaction: consuming the old refresh token is the
	// single-use claim, and both replacement tokens commit with that claim.
	txCtx, tx, err := a.admin.DB.NewTx(r.Context(), false)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to start refresh token transaction, %w", err))
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Validate the refresh token.
	refreshToken, err := a.admin.ValidateAuthToken(txCtx, refreshTokenStr)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Ensure it's actually a refresh token
	userToken, ok := refreshToken.TokenModel().(*database.UserAuthToken)
	if !ok {
		http.Error(w, "invalid token type", http.StatusBadRequest)
		return
	}

	if !userToken.Refresh {
		http.Error(w, "token is not a refresh token", http.StatusBadRequest)
		return
	}
	// ValidateAuthToken has a short cache, so enforce expiry again at the
	// rotation boundary in case the token expired after it was cached.
	if userToken.ExpiresOn != nil && !userToken.ExpiresOn.After(time.Now()) {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	userID := refreshToken.OwnerID()
	tknClientID := ""
	if userToken.AuthClientID != nil {
		tknClientID = *userToken.AuthClientID
	}
	if tknClientID != clientID {
		http.Error(w, "client_id does not match token's client ID", http.StatusBadRequest)
		return
	}

	authClient, err := a.admin.DB.FindAuthClient(txCtx, clientID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, fmt.Sprintf("invalid client_id %q", clientID), http.StatusBadRequest)
			return
		}
		internalServerError(w, fmt.Errorf("failed to lookup auth client, %w", err))
		return
	}
	if !hasGrantType(authClient.GrantTypes, refreshTokenGrantType) {
		http.Error(w, "client is not permitted to use refresh tokens", http.StatusBadRequest)
		return
	}

	// Deleting inside the transaction atomically claims this token. Concurrent
	// rotations can both validate it, but only one can delete and commit it.
	err = a.admin.DB.DeleteUserAuthToken(txCtx, userToken.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		} else {
			internalServerError(w, fmt.Errorf("failed to consume refresh token, %w", err))
		}
		return
	}

	// Issue new access token (60 minutes TTL)
	accessTTL := 60 * time.Minute
	accessToken, err := a.admin.IssueUserAuthToken(txCtx, userID, clientID, "Access Token", nil, &accessTTL, false)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			http.Error(w, "request cancelled or timeout", http.StatusRequestTimeout)
			return
		}
		internalServerError(w, fmt.Errorf("failed to issue access token, %w", err))
		return
	}

	// Issue new refresh token (365 days TTL)
	newRefreshTTL := 365 * 24 * time.Hour
	newRefreshToken, err := a.admin.IssueUserAuthToken(txCtx, userID, clientID, "Refresh Token", nil, &newRefreshTTL, true)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			http.Error(w, "request cancelled or timeout", http.StatusRequestTimeout)
			return
		}
		internalServerError(w, fmt.Errorf("failed to issue refresh token, %w", err))
		return
	}

	resp := oauth.TokenResponse{
		AccessToken:  accessToken.Token().String(),
		ExpiresIn:    int64(accessTTL.Seconds()),
		RefreshToken: newRefreshToken.Token().String(),
		TokenType:    "Bearer",
		Scope:        authClient.Scope,
		UserID:       userID,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to marshal response, %w", err))
		return
	}

	// Commit before writing the response. If the client disconnects while the
	// response is written, the old token stays consumed and cannot be replayed
	// to create a second durable branch.
	if err := tx.Commit(); err != nil {
		internalServerError(w, fmt.Errorf("failed to commit refresh token rotation, %w", err))
		return
	}
	a.admin.PurgeAuthTokenCache()
	a.admin.Used.Client(clientID)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to write response, %w", err))
		return
	}
	a.logger.Debug("Refreshed access and refresh tokens", zap.String("userID", userID), zap.String("clientID", clientID))
}

// verifyCodeChallenge validates the code verifier with the stored code challenge
func verifyCodeChallenge(verifier, challenge, method string) bool {
	switch method {
	case "S256":
		s256 := sha256.Sum256([]byte(verifier))
		computedChallenge := base64.RawURLEncoding.EncodeToString(s256[:])
		return computedChallenge == challenge
	default:
		return false
	}
}

func hasScope(scopes, scope string) bool {
	for _, token := range strings.Fields(scopes) {
		if token == scope {
			return true
		}
	}
	return false
}

func hasGrantType(grants []string, grant string) bool {
	for _, g := range grants {
		if g == grant {
			return true
		}
	}
	return false
}

func normalizeRedirectURI(uri string) (string, error) {
	sanitized, err := sanitizeRedirectURIs([]string{uri})
	if err != nil {
		return "", err
	}
	if len(sanitized) == 0 {
		return "", fmt.Errorf("redirect_uri is required")
	}
	return sanitized[0], nil
}

func isRedirectURIAllowed(client *database.AuthClient, normalizedRedirect string) bool {
	if client == nil {
		return false
	}

	// For localhost rill dev, allow any localhost redirect with /auth/callback path
	if client.ID == database.AuthClientIDRillWebLocal {
		parsed, err := url.Parse(normalizedRedirect)
		if err != nil {
			return false
		}
		if parsed.Hostname() != "localhost" {
			return false
		}
		if parsed.Path != "/auth/callback" {
			return false
		}
		return true
	}

	return slices.Contains(client.RedirectURIs, normalizedRedirect)
}

// Generates a random string for use as the authorization code
func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
