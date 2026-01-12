package ai

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ProjectStatusName = "project_status"

type ProjectStatus struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ProjectStatusArgs, *ProjectStatusResult] = (*ProjectStatus)(nil)

type ProjectStatusArgs struct {
	WhereError bool   `json:"where_error,omitempty" jsonschema:"Optional flag to only return resources that have reconcile errors."`
	Kind       string `json:"kind,omitempty" jsonschema:"Optional filter to only return resources of the specified kind."`
	Name       string `json:"name,omitempty" jsonschema:"Optional filter to only return the resource with the specified name."`
	Path       string `json:"path,omitempty" jsonschema:"Optional filter to only return resources declared in the specified file path."`
}

type ProjectStatusResult struct {
	DefaultOLAPConnector string           `json:"default_olap_connector,omitempty" jsonschema:"The default OLAP connector configured in rill.yaml. May or may not exist as an explicit connector resource."`
	Env                  []string         `json:"env,omitempty" jsonschema:"List of environment variable names present in the project. Their values are omitted for security."`
	Resources            []map[string]any `json:"resources" jsonschema:"List of resources and their status."`
	ParseErrors          []map[string]any `json:"parse_errors" jsonschema:"List of parse errors encountered when parsing project files."`
}

func (t *ProjectStatus) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ProjectStatusName,
		Title:       "Get project status",
		Description: "Returns the reconcile status of resources in the Rill project, including any parse errors.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Getting project status...",
			"openai/toolInvocation/invoked":  "Got project status",
		},
	}
}

func (t *ProjectStatus) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAccess(ctx, t.Runtime, false)
}

func (t *ProjectStatus) Handler(ctx context.Context, args *ProjectStatusArgs) (*ProjectStatusResult, error) {
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

	// Get instance info
	instance, err := t.Runtime.Instance(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	var varNames []string
	for k, v := range instance.ResolveVariables(false) {
		// Skip empty variables and internal ones
		if v == "" || strings.HasPrefix(k, "rill.") {
			continue
		}
		varNames = append(varNames, k)
	}

	return &ProjectStatusResult{
		DefaultOLAPConnector: instance.ResolveOLAPConnector(),
		Env:                  varNames,
		Resources:            resources,
		ParseErrors:          parseErrors,
	}, nil
}
