package ai

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

type ListFiles struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListFilesArgs, *ListFilesResult] = (*ListFiles)(nil)

type ListFilesArgs struct{}

type ListFilesResult struct {
	Files []map[string]any `json:"files"`
}

func (t *ListFiles) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_files",
		Title:       "List project files",
		Description: "Lists all the files in the Rill project, as well as the resources they declare and the current status of those resources",
	}
}

func (t *ListFiles) CheckAccess(ctx context.Context) bool {
	// NOTE: Disabled pending further improvements
	// s := GetSession(ctx)
	// return s.Claims().Can(runtime.ReadRepo)
	return false
}

func (t *ListFiles) Handler(ctx context.Context, args *ListFilesArgs) (*ListFilesResult, error) {
	s := GetSession(ctx)

	ctrl, err := t.Runtime.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	rs, err := ctrl.List(ctx, "", "", false)
	if err != nil {
		return nil, err
	}
	resourcesByPath := make(map[string][]*runtimev1.Resource)
	for _, r := range rs {
		for _, p := range r.Meta.FilePaths {
			resourcesByPath[p] = append(resourcesByPath[p], r)
		}
	}

	parser, err := ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	if err != nil {
		return nil, err
	}
	parseErrorsByPath := make(map[string][]string)
	for _, e := range parser.GetProjectParser().State.ParseErrors {
		parseErrorsByPath[e.FilePath] = append(parseErrorsByPath[e.FilePath], e.Message)
	}

	files, err := t.Runtime.ListFiles(ctx, s.InstanceID(), "**")
	if err != nil {
		return nil, err
	}

	var res []map[string]any
	for _, file := range files {
		if file.IsDir {
			continue
		}

		resources := []map[string]any{}
		for _, r := range resourcesByPath[file.Path] {
			resources = append(resources, map[string]any{
				"kind":             r.Meta.Name.Kind,
				"name":             r.Meta.Name.Name,
				"reconcile_status": r.Meta.ReconcileStatus.String(),
				"reconcile_error":  r.Meta.ReconcileError,
			})
		}

		data := make(map[string]any)
		data["path"] = file.Path
		if len(resources) > 0 {
			data["resources"] = resources
		}
		if len(parseErrorsByPath[file.Path]) > 0 {
			data["parse_errors"] = parseErrorsByPath[file.Path]
		}

		res = append(res, data)
	}

	return &ListFilesResult{
		Files: res,
	}, nil
}
