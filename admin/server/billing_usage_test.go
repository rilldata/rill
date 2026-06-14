package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnonymousUserID(t *testing.T) {
	// Deterministic and independent of map iteration order.
	a := anonymousUserID(map[string]any{"team": "acme", "region": "us", "embed": true})
	b := anonymousUserID(map[string]any{"region": "us", "embed": true, "team": "acme"})
	require.Equal(t, a, b)

	// Different attributes produce a different identifier.
	c := anonymousUserID(map[string]any{"team": "globex", "region": "us", "embed": true})
	require.NotEqual(t, a, c)

	// Empty attributes are handled.
	require.NotEmpty(t, anonymousUserID(map[string]any{}))
}
