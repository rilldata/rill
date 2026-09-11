package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
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

func TestUIContextCompletionMessage(t *testing.T) {
	ctx := ai.WithUIContext(t.Context(), &ai.UIContext{
		PagePath: "/canvas/revenue",
		Actions: []ai.UIAction{
			{ID: "canvas.edit", Label: "Edit"},
		},
	})

	msg := ai.UIContextCompletionMessage(ctx)
	require.NotNil(t, msg)
	require.Equal(t, "system", msg.Role)
	require.Contains(t, msg.Content[0].GetText(), "canvas.edit")
	require.Contains(t, msg.Content[0].GetText(), "/canvas/revenue")
}
