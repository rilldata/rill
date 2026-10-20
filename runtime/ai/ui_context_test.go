package ai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUIContextCompletionMessage(t *testing.T) {
	ctx := WithUIContext(t.Context(), &UIContext{
		PagePath: "/canvas/revenue",
		Actions: []UIAction{
			{ID: "canvas.edit", Label: "Edit"},
		},
	})

	msg := uiContextCompletionMessage(ctx)
	require.NotNil(t, msg)
	require.Equal(t, "system", msg.Role)
	require.Contains(t, msg.Content[0].GetText(), "canvas.edit")
	require.Contains(t, msg.Content[0].GetText(), "/canvas/revenue")
}
