package ai_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestClickUI(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)
	ctx := ai.WithUIContext(t.Context(), &ai.UIContext{
		PagePath: "/explore/orders",
		Actions: []ai.UIAction{
			{ID: "dashboard.start-pivot", Label: "Start pivot"},
		},
	})

	t.Run("browser confirms", func(t *testing.T) {
		r, s := newClickUISession(t, rt, instanceID)
		ackClickUI(t, s, func(callID string) {
			confirmClickUI(t, r, s, callID, "")
		})

		var res *ai.ClickUIResult
		_, err := s.CallTool(ctx, ai.RoleAssistant, ai.ClickUIName, &res, &ai.ClickUIArgs{
			ActionID: "dashboard.start-pivot",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("browser reports failure", func(t *testing.T) {
		r, s := newClickUISession(t, rt, instanceID)
		ackClickUI(t, s, func(callID string) {
			confirmClickUI(t, r, s, callID, "compile failed")
		})

		_, err := s.CallTool(ctx, ai.RoleAssistant, ai.ClickUIName, nil, &ai.ClickUIArgs{
			ActionID: "dashboard.start-pivot",
		})
		require.ErrorContains(t, err, "UI action failed: compile failed")
	})

	t.Run("report from another user or conversation is rejected", func(t *testing.T) {
		ai.SetClickUITimeout(t, 500*time.Millisecond)
		r, s := newClickUISession(t, rt, instanceID)
		ackClickUI(t, s, func(callID string) {
			require.ErrorIs(t, r.CompleteUIAction(s.ID(), uuid.NewString(), callID, ""), ai.ErrUIActionNotPending)
			require.ErrorIs(t, r.CompleteUIAction(uuid.NewString(), s.Claims().UserID, callID, ""), ai.ErrUIActionNotPending)
		})

		_, err := s.CallTool(ctx, ai.RoleAssistant, ai.ClickUIName, nil, &ai.ClickUIArgs{
			ActionID: "dashboard.start-pivot",
		})
		require.ErrorContains(t, err, "did not confirm")
	})

	t.Run("browser never confirms", func(t *testing.T) {
		ai.SetClickUITimeout(t, 100*time.Millisecond)
		r, s := newClickUISession(t, rt, instanceID)
		callIDs := make(chan string, 1)
		ackClickUI(t, s, func(id string) { callIDs <- id })

		_, err := s.CallTool(ctx, ai.RoleAssistant, ai.ClickUIName, nil, &ai.ClickUIArgs{
			ActionID: "dashboard.start-pivot",
		})
		require.ErrorContains(t, err, "did not confirm")
		// A late report finds nothing waiting.
		require.ErrorIs(t, r.CompleteUIAction(s.ID(), s.Claims().UserID, <-callIDs, ""), ai.ErrUIActionNotPending)
	})

	t.Run("unavailable action", func(t *testing.T) {
		_, s := newClickUISession(t, rt, instanceID)
		_, err := s.CallTool(ctx, ai.RoleAssistant, ai.ClickUIName, nil, &ai.ClickUIArgs{
			ActionID: "dashboard.delete",
		})
		require.ErrorContains(t, err, "action_id is not available")
	})
}

func newClickUISession(t *testing.T, rt *runtime.Runtime, instanceID string) (*ai.Runner, *ai.Session) {
	t.Helper()
	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims:     &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true},
		UserAgent:  "rill-test",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, s.Flush(t.Context()))
	})
	return r, s
}

// confirmClickUI reports an action's outcome, retrying until the click_ui handler has registered its waiter:
// the call message is emitted just before the handler runs.
func confirmClickUI(t *testing.T, r *ai.Runner, s *ai.Session, callID, errMsg string) {
	require.Eventually(t, func() bool {
		return r.CompleteUIAction(s.ID(), s.Claims().UserID, callID, errMsg) == nil
	}, 5*time.Second, 10*time.Millisecond)
}

// ackClickUI plays the browser: it calls ack with the ID of each click_ui call message the session emits.
func ackClickUI(t *testing.T, s *ai.Session, ack func(callID string)) {
	t.Helper()
	ch := s.Subscribe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for msg := range ch {
			if msg.Tool == ai.ClickUIName && msg.Type == ai.MessageTypeCall {
				ack(msg.ID)
			}
		}
	}()
	t.Cleanup(func() {
		s.Unsubscribe(ch)
		<-done
	})
}

func TestClickUICheckAccess(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	newSessionWithAccess := func(t *testing.T, userAgent string, permissions ...runtime.Permission) *ai.Session {
		t.Helper()
		r := ai.NewRunner(rt, activity.NewNoopClient())
		s, err := r.Session(t.Context(), &ai.SessionOptions{
			InstanceID: instanceID,
			Claims: &runtime.SecurityClaims{
				UserID:      uuid.NewString(),
				Permissions: permissions,
			},
			UserAgent: userAgent,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, s.Flush(t.Context()))
		})
		return s
	}

	ui := &ai.UIContext{
		PagePath: "/explore/orders",
		Actions:  []ai.UIAction{{ID: "dashboard.start-pivot", Label: "Start pivot"}},
	}
	tests := []struct {
		name        string
		userAgent   string
		permissions []runtime.Permission
		ui          *ai.UIContext
		want        bool
	}{
		{name: "first-party client with UI actions", userAgent: "rill-web", permissions: []runtime.Permission{runtime.UseAI}, ui: ui, want: true},
		{name: "first-party client without UI context", userAgent: "rill-web", permissions: []runtime.Permission{runtime.UseAI}, want: false},
		{name: "first-party client with empty UI actions", userAgent: "rill-web", permissions: []runtime.Permission{runtime.UseAI}, ui: &ai.UIContext{PagePath: "/explore/orders"}, want: false},
		{name: "external client", userAgent: "mcp-client", permissions: []runtime.Permission{runtime.UseAI}, ui: ui, want: false},
		{name: "missing AI permission", userAgent: "rill-web", ui: ui, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSessionWithAccess(t, tt.userAgent, tt.permissions...)
			tool, ok := s.Tool(ai.ClickUIName)
			require.True(t, ok)
			ctx := ai.WithSession(t.Context(), s)
			ctx = ai.WithUIContext(ctx, tt.ui)
			allowed, err := tool.CheckAccess(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, allowed)
		})
	}
}
