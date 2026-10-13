package server

import (
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestUIContextFromPB(t *testing.T) {
	ui, err := uiContextFromPB("/explore/orders", []*runtimev1.UIAction{
		{Id: "dashboard.start-pivot", Label: "Start pivot"},
	})
	require.NoError(t, err)
	require.Equal(t, "/explore/orders", ui.PagePath)
	require.Len(t, ui.Actions, 1)
	require.Equal(t, "dashboard.start-pivot", ui.Actions[0].ID)

	ui, err = uiContextFromPB("", nil)
	require.NoError(t, err)
	require.Nil(t, ui)
}

func TestUIContextFromPBRejectsMalformedInput(t *testing.T) {
	tests := []struct {
		name     string
		pagePath string
		actions  []*runtimev1.UIAction
		wantErr  string
	}{
		{name: "long page path", pagePath: strings.Repeat("x", 2049), wantErr: "ui_page_path exceeds"},
		{name: "too many actions", actions: make([]*runtimev1.UIAction, 101), wantErr: "ui_actions exceeds"},
		{name: "missing action", actions: []*runtimev1.UIAction{nil}, wantErr: "ui_actions[0] is missing"},
		{name: "missing ID", actions: []*runtimev1.UIAction{{Label: "Edit"}}, wantErr: "ui_actions[0].id is required"},
		{name: "long ID", actions: []*runtimev1.UIAction{{Id: strings.Repeat("x", 129)}}, wantErr: "ui_actions[0].id exceeds"},
		{name: "long label", actions: []*runtimev1.UIAction{{Id: "canvas.edit", Label: strings.Repeat("x", 257)}}, wantErr: "ui_actions[0].label exceeds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui, err := uiContextFromPB(tt.pagePath, tt.actions)
			require.Nil(t, ui)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}
