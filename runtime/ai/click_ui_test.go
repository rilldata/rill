package ai_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestClickUI(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)
	s := newSession(t, rt, instanceID)
	ctx := ai.WithUIContext(t.Context(), &ai.UIContext{
		PagePath: "/explore/orders",
		Actions: []ai.UIAction{
			{ID: "dashboard.start-pivot", Label: "Start pivot"},
		},
	})

	var res *ai.ClickUIResult
	_, err := s.CallTool(ctx, ai.RoleAssistant, ai.ClickUIName, &res, &ai.ClickUIArgs{
		ActionID: "dashboard.start-pivot",
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	_, err = s.CallTool(ctx, ai.RoleAssistant, ai.ClickUIName, nil, &ai.ClickUIArgs{
		ActionID: "dashboard.delete",
	})
	require.ErrorContains(t, err, "action_id is not available")
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
