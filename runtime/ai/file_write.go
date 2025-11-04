package ai

import (
	"context"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

type WriteFile struct {
	Runtime *runtime.Runtime
}

var _ Tool[*WriteFileArgs, *WriteFileResult] = (*WriteFile)(nil)

type WriteFileArgs struct {
	Path     string `json:"path" jsonschema:"The path of the file to write"`
	Contents string `json:"contents" jsonschema:"The new contents to write to the file. If the file already exists, this will overwrite it."`
}

type WriteFileResult struct {
	Resources  []map[string]any `json:"resources,omitempty"`
	ParseError string           `json:"parse_error,omitempty"`
}

func (t *WriteFile) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "write_file",
		Title:       "Write file",
		Description: "Creates or updates a file in a Rill project. If the file already exists, it will be overwritten. If the file declares a Rill resource, it will wait for the resource to reconcile and return its kind, name and any errors encountered.",
	}
}

func (t *WriteFile) CheckAccess(ctx context.Context) bool {
	// NOTE: Disabled pending further improvements
	// s := GetSession(ctx)
	// return s.Claims().Can(runtime.EditRepo)
	return false
}

func (t *WriteFile) Handler(ctx context.Context, args *WriteFileArgs) (*WriteFileResult, error) {
	s := GetSession(ctx)

	if !strings.HasPrefix(args.Path, "/") {
		args.Path = "/" + args.Path
	}

	err := t.Runtime.PutFile(ctx, s.InstanceID(), args.Path, strings.NewReader(args.Contents), true, false)
	if err != nil {
		return nil, err
	}

	ctrl, err := t.Runtime.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	err = ctrl.Reconcile(ctx, runtime.GlobalProjectParserName) // TODO: Only if not streaming
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(time.Millisecond * 500):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	p, err := ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	if err != nil {
		return nil, err
	}
	for _, pe := range p.GetProjectParser().State.ParseErrors {
		if pe.FilePath == args.Path {
			return &WriteFileResult{
				ParseError: pe.Message,
			}, nil
		}
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if err != nil {
		return nil, err
	}

	rs, err := ctrl.List(ctx, "", args.Path, false)
	if err != nil {
		return nil, err
	}
	if len(rs) == 0 {
		return &WriteFileResult{}, nil
	}

	resources := []map[string]any{}
	for _, r := range rs {
		resources = append(resources, map[string]any{
			"kind":             r.Meta.Name.Kind,
			"name":             r.Meta.Name.Name,
			"reconcile_status": r.Meta.ReconcileStatus.String(),
			"reconcile_error":  r.Meta.ReconcileError,
		})
	}

	return &WriteFileResult{
		Resources: resources,
	}, nil
}
