package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/pkg/oauth"
	"github.com/stretchr/testify/require"
)

func TestOAuthProtectedResourceMetadata(t *testing.T) {
	// Create minimal URLs for testing
	urls, err := admin.NewURLs("http://localhost:8080", "http://localhost:3000")
	require.NoError(t, err)

	// Create minimal authenticator for testing
	auth := &Authenticator{
		admin: &admin.Service{
			URLs: urls,
		},
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
	w := httptest.NewRecorder()

	// Call handler
	auth.handleOAuthProtectedResourceMetadata(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var metadata oauth.ProtectedResourceMetadata
	err = json.NewDecoder(w.Body).Decode(&metadata)
	require.NoError(t, err)

	// Verify fields
	require.Equal(t, "http://localhost:8080", metadata.Resource)
	require.Equal(t, []string{"http://localhost:8080"}, metadata.AuthorizationServers)
	require.Contains(t, metadata.BearerMethodsSupported, "header")
}

func TestOAuthAuthorizationServerMetadata(t *testing.T) {
	// Create minimal URLs for testing
	urls, err := admin.NewURLs("http://localhost:8080", "http://localhost:3000")
	require.NoError(t, err)

	// Create minimal authenticator for testing
	auth := &Authenticator{
		admin: &admin.Service{
			URLs: urls,
		},
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
	w := httptest.NewRecorder()

	// Call handler
	auth.handleOAuthAuthorizationServerMetadata(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var metadata oauth.AuthorizationServerMetadata
	err = json.NewDecoder(w.Body).Decode(&metadata)
	require.NoError(t, err)

	// Verify fields
	require.Equal(t, "http://localhost:8080", metadata.Issuer)
	require.Equal(t, "http://localhost:8080/auth/oauth/authorize", metadata.AuthorizationEndpoint)
	require.Equal(t, "http://localhost:8080/auth/oauth/token", metadata.TokenEndpoint)
	require.Equal(t, "http://localhost:8080/auth/oauth/register", metadata.RegistrationEndpoint)
	require.Contains(t, metadata.ResponseTypesSupported, "code")
	require.Contains(t, metadata.GrantTypesSupported, authorizationCodeGrantType)
	require.Contains(t, metadata.GrantTypesSupported, refreshTokenGrantType)
	require.Contains(t, metadata.GrantTypesSupported, deviceCodeGrantType)
	require.Contains(t, metadata.CodeChallengeMethodsSupported, "S256")
	require.Contains(t, metadata.TokenEndpointAuthMethodsSupported, "none")
}

func TestOAuthRegister(t *testing.T) {
	// Create minimal URLs for testing
	urls, err := admin.NewURLs("http://localhost:8080", "http://localhost:3000")
	require.NoError(t, err)

	// Create mock admin service
	mockAdmin := &admin.Service{
		URLs: urls,
	}

	auth := &Authenticator{
		admin: mockAdmin,
	}

	t.Run("rejects non-POST requests", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/oauth/register", nil)
		w := httptest.NewRecorder()

		auth.handleOAuthRegister(w, req)

		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/oauth/register", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		auth.handleOAuthRegister(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}
