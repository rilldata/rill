package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsSensitiveHeaderKey(t *testing.T) {
	sensitive := []string{"Authorization", "X-API-Key", "API-Key", "Token", "X-Token", "X-Auth", "X-Secret", "Proxy-Authorization"}
	for _, key := range sensitive {
		require.True(t, IsSensitiveHeaderKey(key), "expected %q to be sensitive", key)
	}

	notSensitive := []string{"Content-Type", "Accept", "X-Request-ID", "User-Agent"}
	for _, key := range notSensitive {
		require.False(t, IsSensitiveHeaderKey(key), "expected %q to not be sensitive", key)
	}
}

func TestSplitAuthSchemePrefix(t *testing.T) {
	scheme, secret, ok := SplitAuthSchemePrefix("Bearer my-token-123")
	require.True(t, ok)
	require.Equal(t, "Bearer ", scheme)
	require.Equal(t, "my-token-123", secret)

	scheme, secret, ok = SplitAuthSchemePrefix("Basic dXNlcjpwYXNz")
	require.True(t, ok)
	require.Equal(t, "Basic ", scheme)
	require.Equal(t, "dXNlcjpwYXNz", secret)

	// No prefix
	_, _, ok = SplitAuthSchemePrefix("plain-token")
	require.False(t, ok)

	// Prefix only (no token part)
	_, _, ok = SplitAuthSchemePrefix("Bearer")
	require.False(t, ok)
}

func TestHeaderKeyToEnvSegment(t *testing.T) {
	require.Equal(t, "x_api_key", HeaderKeyToEnvSegment("X-API-Key"))
	require.Equal(t, "authorization", HeaderKeyToEnvSegment("Authorization"))
	require.Equal(t, "content_type", HeaderKeyToEnvSegment("Content-Type"))
}

func TestResolveHeaderEnvVarName(t *testing.T) {
	existing := make(map[string]bool)
	name := ResolveHeaderEnvVarName("my_conn", "authorization", existing)
	require.Equal(t, "connector.my_conn.authorization", name)

	// Conflict
	existing[name] = true
	name2 := ResolveHeaderEnvVarName("my_conn", "authorization", existing)
	require.Equal(t, "connector.my_conn.authorization_1", name2)
}
