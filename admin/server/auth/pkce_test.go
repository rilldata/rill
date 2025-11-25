package auth

import (
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRedirectURI(t *testing.T) {
	uri, err := normalizeRedirectURI("https://example.com/auth/callback ")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/auth/callback", uri)
}

func TestIsRedirectURIAllowed(t *testing.T) {
	t.Run("third-party client must be in list", func(t *testing.T) {
		client := &database.AuthClient{
			ID:           "client-123",
			RedirectURIs: []string{"https://app.example.com/auth/callback"},
		}
		require.True(t, isRedirectURIAllowed(client, "https://app.example.com/auth/callback"))
		require.False(t, isRedirectURIAllowed(client, "https://other.example.com/auth/callback"))
	})

	t.Run("localhost client allows any port on /auth/callback", func(t *testing.T) {
		client := &database.AuthClient{ID: database.AuthClientIDRillWebLocal}
		require.True(t, isRedirectURIAllowed(client, "http://localhost:3000/auth/callback"))
		require.True(t, isRedirectURIAllowed(client, "https://localhost:12345/auth/callback"))
		require.False(t, isRedirectURIAllowed(client, "http://localhost:3000/other"))
		require.False(t, isRedirectURIAllowed(client, "http://127.0.0.1:3000/auth/callback"))
	})
}
