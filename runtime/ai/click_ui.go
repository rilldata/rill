package ai

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ClickUIName = "click_ui"

type ClickUI struct{}

var _ Tool[*ClickUIArgs, *ClickUIResult] = (*ClickUI)(nil)

type ClickUIArgs struct {
	ActionID string `json:"action_id" jsonschema:"Exact ID of a currently available UI action. Only use IDs provided in the current UI context."`
}

type ClickUIResult struct{}

func (t *ClickUI) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ClickUIName,
		Title:       "Click UI action",
		Description: "Request the Rill browser UI to click a safe action explicitly exposed in the current UI context. Only use an exact available action ID.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Clicking UI action...",
			"openai/toolInvocation/invoked":  "UI action requested",
		},
	}
}

func (t *ClickUI) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}

	// Browser actions are only meaningful for the first-party Rill client.
	if !strings.HasPrefix(s.CatalogSession().UserAgent, "rill") {
		return false, nil
	}
	ui, ok := UIContextFromContext(ctx)
	return ok && len(ui.Actions) > 0, nil
}

func (t *ClickUI) Handler(ctx context.Context, args *ClickUIArgs) (*ClickUIResult, error) {
	if args.ActionID == "" {
		return nil, errors.New("action_id is required")
	}

	ui, ok := UIContextFromContext(ctx)
	if !ok {
		return nil, errors.New("no UI actions are available")
	}
	matches := 0
	for _, action := range ui.Actions {
		if action.ID == args.ActionID {
			matches++
		}
	}
	if matches == 1 {
		return &ClickUIResult{}, nil
	}

	return nil, errors.New("action_id is not available on the current page")
}
