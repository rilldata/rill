package ai

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ResourceStatusName = "resource_status"

type ResourceStatus struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ResourceStatusArgs, *ResourceStatusResult] = (*ResourceStatus)(nil)

type ResourceStatusArgs struct {
	WhereError bool   `json:"where_error,omitempty" jsonschema:"Optional flag to only return resources that have reconcile errors."`
	Kind       string `json:"kind,omitempty" jsonschema:"Optional filter to only return resources of the specified kind."`
	Name       string `json:"name,omitempty" jsonschema:"Optional filter to only return the resource with the specified name."`
	Path       string `json:"path,omitempty" jsonschema:"Optional filter to only return resources declared in the specified file path."`
}

type ResourceStatusResult struct {
	Resources   []map[string]any `json:"resources" jsonschema:"List of resources and their status."`
	ParseErrors []map[string]any `json:"parse_errors" jsonschema:"List of parse errors encountered when parsing project files."`
}

func (t *ResourceStatus) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ResourceStatusName,
		Title:       "Get resource status",
		Description: "Returns the reconcile status of resources in the Rill project, including any parse errors.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Getting resource status...",
			"openai/toolInvocation/invoked":  "Got resource status",
		},
	}
}

func (t *ResourceStatus) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAgentAccess(ctx, t.Runtime)
}

func (t *ResourceStatus) Handler(ctx context.Context, args *ResourceStatusArgs) (*ResourceStatusResult, error) {
	s := GetSession(ctx)

	ctrl, err := t.Runtime.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}

	// List resources with optional filtering by kind and path
	rs, err := ctrl.List(ctx, args.Kind, args.Path, false)
	if err != nil {
		return nil, err
	}

	// Build the resources list with optional filtering
	resources := []map[string]any{}
	for _, r := range rs {
		// Apply where_error filter
		if args.WhereError && r.Meta.ReconcileError == "" {
			continue
		}

		// Apply name filter
		if args.Name != "" && r.Meta.Name.Name != args.Name {
			continue
		}

		// Build refs list
		refs := []map[string]any{}
		for _, ref := range r.Meta.Refs {
			refs = append(refs, map[string]any{
				"kind": ref.Kind,
				"name": ref.Name,
			})
		}

		// Get the first file path (resources can be declared in multiple files, but typically just one)
		var path string
		if len(r.Meta.FilePaths) > 0 {
			path = r.Meta.FilePaths[0]
		}

		resources = append(resources, map[string]any{
			"kind":             r.Meta.Name.Kind,
			"name":             r.Meta.Name.Name,
			"path":             path,
			"refs":             refs,
			"reconcile_status": r.Meta.ReconcileStatus.String(),
			"reconcile_error":  r.Meta.ReconcileError,
		})
	}

	// Get parse errors from the global project parser
	parseErrors := []map[string]any{}
	parser, err := ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	if err != nil {
		return nil, err
	}
	for _, pe := range parser.GetProjectParser().State.ParseErrors {
		// Apply path filter to parse errors as well
		if args.Path != "" && pe.FilePath != args.Path {
			continue
		}
		parseErrors = append(parseErrors, map[string]any{
			"path":    pe.FilePath,
			"message": pe.Message,
		})
	}

	return &ResourceStatusResult{
		Resources:   resources,
		ParseErrors: parseErrors,
	}, nil
}
