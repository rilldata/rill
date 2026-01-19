package ai

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const NavigateName = "navigate"

type Navigate struct {
}

var _ Tool[*NavigateArgs, *NavigateResult] = (*Navigate)(nil)

type NavigateArgs struct {
	Kind string `json:"kind" jsonschema:"The kind of navigation to perform. Supported values: 'file', 'explore', 'canvas'."`
	Name string `json:"name" jsonschema:"The name of the item to navigate to."`
}

type NavigateResult struct {
}

func (t *Navigate) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        NavigateName,
		Title:       "Navigate UI",
		Description: "Navigate to a specific UI element in the Rill UI. Supported kinds: 'file', 'explore', 'canvas'.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Navigating...",
			"openai/toolInvocation/invoked":  "Navigated",
		},
	}
}

func (t *Navigate) CheckAccess(ctx context.Context) (bool, error) {
	// Must be allowed to use AI features
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}

	// Only allow for rill user agents since it's not functional in MCP contexts.
	if !strings.HasPrefix(s.CatalogSession().UserAgent, "rill") {
		return false, nil
	}
	return true, nil
}

func (t *Navigate) Handler(ctx context.Context, args *NavigateArgs) (*NavigateResult, error) {
	if args.Kind == "" {
		return nil, errors.New("kind is required")
	}

	if args.Name == "" {
		return nil, errors.New("name is required")
	}

	return &NavigateResult{}, nil
}
