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
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/oauth"
)

const authorizationCodeGrantType = "authorization_code"

func (a *Authenticator) handlePKCE(w http.ResponseWriter, r *http.Request, clientID, userID, codeChallenge, codeChallengeMethod, redirectURI string) {
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
	_, err = a.admin.DB.InsertAuthorizationCode(r.Context(), code, userID, clientID, redirectURI, codeChallenge, codeChallengeMethod, expiration)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to store authorization code, %w", err))
		return
	}

	// Build the redirection URI with the authorization code as per OAuth2 spec, state is URL-encoded
	redirectWithCode := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, r.URL.Query().Get("state"))

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

	// get the authorization code from the database
	authCode, err := a.admin.DB.FindAuthorizationCode(r.Context(), code)
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

	// remove the authorization code from the database to prevent reuse
	err = a.admin.DB.DeleteAuthorizationCode(r.Context(), code)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to delete authorization code, %w", err))
		return
	}

	// Check if the client ID matches the stored client ID
	if authCode.ClientID != clientID {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}

	// Check if the redirect URI matches the stored redirect URI
	if authCode.RedirectURI != redirectURI {
		http.Error(w, "invalid redirect URI", http.StatusBadRequest)
		return
	}

	// Check if the authorization code has expired
	if time.Now().After(authCode.Expiration) {
		http.Error(w, "authorization code has expired", http.StatusBadRequest)
		return
	}

	// Verify the code verifier against the stored code challenge
	if !verifyCodeChallenge(codeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
		http.Error(w, "invalid code verifier", http.StatusBadRequest)
		return
	}

	// Issue an access token
	authToken, err := a.admin.IssueUserAuthToken(r.Context(), userID, authCode.ClientID, "", nil, nil)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			http.Error(w, "request cancelled or timeout", http.StatusRequestTimeout)
			return
		}
		internalServerError(w, fmt.Errorf("failed to issue access token, %w", err))
		return
	}

	resp := oauth.TokenResponse{
		AccessToken: authToken.Token().String(),
		TokenType:   "Bearer",
		ExpiresIn:   0, // never expires
		UserID:      userID,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to marshal response, %w", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to write response, %w", err))
		return
	}
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

// Generates a random string for use as the authorization code
func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
