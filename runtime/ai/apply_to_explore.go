package ai

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ApplyToExploreName = "apply_to_explore"

type ApplyToExplore struct{}

var _ Tool[*ApplyToExploreArgs, *ApplyToExploreResult] = (*ApplyToExplore)(nil)

type ApplyToExploreArgs struct {
	Name       string   `json:"name" jsonschema:"The name of the explore to navigate to."`
	Dimensions []string `json:"dimensions,omitempty" jsonschema:"Optional dimensions to preview in the explore."`
	Measures   []string `json:"measures,omitempty" jsonschema:"Optional measures to preview in the explore."`
	SortBy     string   `json:"sort_by,omitempty" jsonschema:"Optional measure to sort by in the explore."`
	SortDesc   bool     `json:"sort_desc,omitempty" jsonschema:"Optional flag to sort in descending order in the explore."`
}

type ApplyToExploreResult struct{}

func (t *ApplyToExplore) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ApplyToExploreName,
		Title:       "Apply to explore",
		Description: "Apply settings to a specific explore.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Applying...",
			"openai/toolInvocation/invoked":  "Applied",
		},
	}
}

func (t *ApplyToExplore) CheckAccess(ctx context.Context) (bool, error) {
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

func (t *ApplyToExplore) Handler(ctx context.Context, args *ApplyToExploreArgs) (*ApplyToExploreResult, error) {
	if args.Name == "" {
		return nil, errors.New("name is required")
	}

	return &ApplyToExploreResult{}, nil
}
