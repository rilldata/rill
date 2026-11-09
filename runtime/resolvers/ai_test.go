package resolvers

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/stretchr/testify/require"
)

func TestAIAlert(t *testing.T) {
	// A met condition returns a row with the reason, which makes the alert fire
	rows, err := aiAlertRows(&ai.AlertVerdict{Met: true, Reason: "Revenue fell 40%."}, "s1")
	require.NoError(t, err)
	require.Equal(t, []map[string]any{{"reason": "Revenue fell 40%.", "ai_session_id": "s1"}}, rows)

	// A condition that is not met returns no rows
	rows, err = aiAlertRows(&ai.AlertVerdict{Met: false, Reason: "Revenue is stable."}, "s1")
	require.NoError(t, err)
	require.Empty(t, rows)

	// A missing verdict is an error
	_, err = aiAlertRows(nil, "s1")
	require.Error(t, err)

	// The condition is never judged with claims that skip security checks
	r := &aiResolver{
		props:  &aiProps{Prompt: "Revenue dropped", IsAlert: true},
		args:   &aiArgs{},
		claims: &runtime.SecurityClaims{SkipChecks: true},
	}
	_, err = r.ResolveInteractive(t.Context())
	require.ErrorContains(t, err, "security checks")
}
