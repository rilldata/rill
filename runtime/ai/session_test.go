package ai_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// TestSessionConcurrentCreate ensures concurrent requests addressing the same new session ID all get that session.
// It happens when an MCP client makes parallel tool calls on a session the runtime has not seen before,
// e.g. the first calls the unified MCP server in the admin service forwards for a project.
func TestSessionConcurrentCreate(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)
	r := ai.NewRunner(rt, activity.NewNoopClient())
	sessionID := uuid.NewString()

	start := make(chan struct{})
	var g errgroup.Group
	for range 10 {
		g.Go(func() error {
			<-start
			s, err := r.Session(t.Context(), &ai.SessionOptions{
				InstanceID:        instanceID,
				SessionID:         sessionID,
				CreateIfNotExists: true,
				Claims:            &runtime.SecurityClaims{SkipChecks: true},
				UserAgent:         "test",
			})
			if err != nil {
				return err
			}
			if s.ID() != sessionID {
				return fmt.Errorf("got session %q, want %q", s.ID(), sessionID)
			}
			return nil
		})
	}
	close(start)
	require.NoError(t, g.Wait())

	catalog, release, err := rt.Catalog(t.Context(), instanceID)
	require.NoError(t, err)
	defer release()
	sessions, err := catalog.FindAISessions(t.Context(), "", "")
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	require.Equal(t, sessionID, sessions[0].ID)
}
