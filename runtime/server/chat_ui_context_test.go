package server

import (
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestUIContextFromPB(t *testing.T) {
	ui := uiContextFromPB([]*runtimev1.UIAction{
		{Id: "dashboard.start-pivot", Label: "Start pivot"},
	})
	require.Len(t, ui.Actions, 1)
	require.Equal(t, "dashboard.start-pivot", ui.Actions[0].ID)
	require.Equal(t, "Start pivot", ui.Actions[0].Label)

	require.Nil(t, uiContextFromPB(nil))
}

// The UI context is best-effort metadata, so oversized or malformed input must be sanitized rather than fail the chat request.
func TestUIContextFromPBSanitizesActions(t *testing.T) {
	// The browser limits labels to 256 characters, not bytes, so a non-ASCII label of 256 characters must be accepted as is.
	nonASCIILabel := strings.Repeat("é", maxUIActionLabelRunes)
	require.Greater(t, len(nonASCIILabel), maxUIActionLabelRunes)

	ui := uiContextFromPB([]*runtimev1.UIAction{
		{Id: "canvas.edit", Label: strings.Repeat("x", maxUIActionLabelRunes+1)},
		{Id: "canvas.share", Label: nonASCIILabel},
		nil,
		{Label: "missing id"},
		{Id: strings.Repeat("x", maxUIActionIDBytes+1), Label: "long id"},
		{Id: "dashboard.start-pivot", Label: "Start pivot"},
	})
	require.Len(t, ui.Actions, 3)
	require.Equal(t, "canvas.edit", ui.Actions[0].ID)
	require.Equal(t, strings.Repeat("x", maxUIActionLabelRunes), ui.Actions[0].Label)
	require.Equal(t, "canvas.share", ui.Actions[1].ID)
	require.Equal(t, nonASCIILabel, ui.Actions[1].Label)
	require.Equal(t, "dashboard.start-pivot", ui.Actions[2].ID)

	// Only the first 100 actions are kept.
	many := make([]*runtimev1.UIAction, maxUIActions+5)
	for i := range many {
		many[i] = &runtimev1.UIAction{Id: fmt.Sprintf("action.%d", i)}
	}
	ui = uiContextFromPB(many)
	require.Len(t, ui.Actions, maxUIActions)
	require.Equal(t, "action.0", ui.Actions[0].ID)
	require.Equal(t, fmt.Sprintf("action.%d", maxUIActions-1), ui.Actions[maxUIActions-1].ID)
}
