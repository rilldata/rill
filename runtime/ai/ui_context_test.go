package ai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUIContextCompletionMessage(t *testing.T) {
	ctx := WithUIContext(t.Context(), &UIContext{
		Actions: []UIAction{
			{ID: "canvas.edit", Label: "Edit"},
		},
	})

	msg := uiContextCompletionMessage(ctx)
	require.NotNil(t, msg)
	require.Equal(t, "system", msg.Role)
	require.Contains(t, msg.Content[0].GetText(), `{"id":"canvas.edit","label":"Edit"}`)

	require.Nil(t, uiContextCompletionMessage(t.Context()))
	require.Nil(t, uiContextCompletionMessage(WithUIContext(t.Context(), &UIContext{})))
}
