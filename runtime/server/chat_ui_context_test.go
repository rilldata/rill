package server

import (
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestUIContextFromPB(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	s := &Server{logger: zap.New(core)}

	ui := s.uiContextFromPB([]*runtimev1.UIAction{
		{Id: "dashboard.start-pivot", Label: "Start pivot"},
	})
	require.Len(t, ui.Actions, 1)
	require.Equal(t, "dashboard.start-pivot", ui.Actions[0].ID)
	require.Equal(t, "Start pivot", ui.Actions[0].Label)
	require.Equal(t, 0, logs.Len())

	require.Nil(t, s.uiContextFromPB(nil))
}

// The UI context is best-effort metadata, so oversized or malformed input must be sanitized rather than fail the chat request.
func TestUIContextFromPBSanitizesActions(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	s := &Server{logger: zap.New(core)}

	// The browser limits labels to 256 characters, not bytes, so a non-ASCII label of 256 characters must be accepted as is.
	nonASCIILabel := strings.Repeat("é", maxUIActionLabelRunes)
	require.Greater(t, len(nonASCIILabel), maxUIActionLabelRunes)

	ui := s.uiContextFromPB([]*runtimev1.UIAction{
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
	// Dropping malformed actions is logged.
	require.Equal(t, 1, logs.Len())
	require.Equal(t, int64(6), logs.All()[0].ContextMap()["received"])
	require.Equal(t, int64(3), logs.All()[0].ContextMap()["kept"])
}

func TestUIContextFromPBCapsActionsAfterValidation(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	s := &Server{logger: zap.New(core)}

	// Malformed entries come first so that capping before validation would wrongly crowd out valid actions.
	pbActions := make([]*runtimev1.UIAction, 0, maxUIActions+10)
	for i := 0; i < 5; i++ {
		pbActions = append(pbActions, &runtimev1.UIAction{Label: "missing id"})
	}
	for i := 0; i < maxUIActions+5; i++ {
		pbActions = append(pbActions, &runtimev1.UIAction{Id: fmt.Sprintf("action.%d", i)})
	}

	ui := s.uiContextFromPB(pbActions)
	require.Len(t, ui.Actions, maxUIActions)
	require.Equal(t, "action.0", ui.Actions[0].ID)
	require.Equal(t, fmt.Sprintf("action.%d", maxUIActions-1), ui.Actions[maxUIActions-1].ID)

	// Hitting the cap is logged so it does not go unnoticed.
	require.Equal(t, 1, logs.Len())
	require.Equal(t, int64(maxUIActions), logs.All()[0].ContextMap()["limit"])

	// Exactly at the cap nothing is dropped and nothing is logged.
	ui = s.uiContextFromPB(pbActions[5 : 5+maxUIActions])
	require.Len(t, ui.Actions, maxUIActions)
	require.Equal(t, 1, logs.Len())
}
