package ai

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/rilldata/rill/runtime"
)

const WriteFileName = "write_file"

type WriteFile struct {
	Runtime *runtime.Runtime
}

var _ Tool[*WriteFileArgs, *WriteFileResult] = (*WriteFile)(nil)

type WriteFileArgs struct {
	Path     string `json:"path" jsonschema:"The path of the file to write"`
	Contents string `json:"contents,omitempty" jsonschema:"Optional new contents to write to the file. If the file already exists, this will overwrite it."`
	Remove   bool   `json:"remove,omitempty" jsonschema:"Optional flag to remove the file instead of writing to it. Defaults to false."`
}

type WriteFileResult struct {
	Diff          string           `json:"diff,omitempty" jsonschema:"Diff of the file contents."`
	IsNewFile     bool             `json:"is_new_file,omitempty" jsonschema:"Indicates if the tool created a new file."`
	Resources     []map[string]any `json:"resources,omitempty" jsonschema:"The Rill resources declared in the file, if any."`
	ParseError    string           `json:"parse_error,omitempty" jsonschema:"Parse error encountered when parsing the file, if any."`
	ParseWarnings []string         `json:"parse_warnings,omitempty" jsonschema:"Parse warnings encountered when parsing the file, if any. The file may still be successfully reconciled if there are warnings."`
}

func (t *WriteFile) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        WriteFileName,
		Title:       "Write file",
		Description: "Creates, updates or deletes a file in a Rill project. If the file already exists, it will be overwritten. If the file declares a Rill resource, it will wait for the resource to reconcile and return its kind, name and any errors encountered.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(true),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Writing file...",
			"openai/toolInvocation/invoked":  "Wrote file",
		},
	}
}

func (t *WriteFile) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAccess(ctx, t.Runtime, true)
}

func (t *WriteFile) Handler(ctx context.Context, args *WriteFileArgs) (*WriteFileResult, error) {
	s := GetSession(ctx)

	if !strings.HasPrefix(args.Path, "/") {
		args.Path = "/" + args.Path
	}

	// Read existing content before writing (for diff computation)
	var isNewFile bool
	originalContent, _, err := t.Runtime.GetFile(ctx, s.InstanceID(), args.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		isNewFile = true
	}

	// Write the file
	var resources []map[string]any
	var parseErr string
	var parseWarnings []string
	if args.Remove {
		err = t.Runtime.DeleteFile(ctx, s.InstanceID(), args.Path, false)
		if err != nil {
			return nil, err
		}
	} else {
		err = t.Runtime.PutFile(ctx, s.InstanceID(), args.Path, strings.NewReader(args.Contents), true, false)
		if err != nil {
			return nil, err
		}

		resources, parseErr, parseWarnings, err = t.reconcileAndGetStatus(ctx, args.Path)
		if err != nil {
			return nil, err
		}
	}

	// Compute a unified diff
	var diff string
	diff, _ = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(originalContent),
		FromFile: args.Path,
		B:        difflib.SplitLines(args.Contents),
		ToFile:   args.Path,
		Context:  3,
	})

	// Done
	return &WriteFileResult{
		Diff:          diff,
		IsNewFile:     isNewFile,
		Resources:     resources,
		ParseError:    parseErr,
		ParseWarnings: parseWarnings,
	}, nil
}

// reconcileAndGetStatus waits until reconciliation is done, then returns the status of resources declared in the file at the given path.
func (t *WriteFile) reconcileAndGetStatus(ctx context.Context, path string) (resources []map[string]any, parseError string, parseWarnings []string, err error) {
	s := GetSession(ctx)
	ctrl, err := t.Runtime.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, "", nil, err
	}
	err = ctrl.Reconcile(ctx, runtime.GlobalProjectParserName) // TODO: Only if not streaming
	if err != nil {
		return nil, "", nil, err
	}

	select {
	case <-time.After(time.Millisecond * 500):
	case <-ctx.Done():
		return nil, "", nil, ctx.Err()
	}

	p, err := ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	if err != nil {
		return nil, "", nil, err
	}
	for _, e := range p.GetProjectParser().State.ParseErrors {
		if e.FilePath == path && e.Warning {
			parseWarnings = append(parseWarnings, e.Message)
		}
	}
	for _, pe := range p.GetProjectParser().State.ParseErrors {
		if pe.FilePath == path && !pe.Warning {
			return nil, pe.Message, parseWarnings, nil
		}
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if err != nil {
		return nil, "", nil, err
	}

	rs, err := ctrl.List(ctx, "", path, false)
	if err != nil {
		return nil, "", nil, err
	}
	if len(rs) == 0 {
		return nil, "", nil, nil
	}

	resources = []map[string]any{}
	for _, r := range rs {
		resources = append(resources, map[string]any{
			"kind":               r.Meta.Name.Kind,
			"name":               r.Meta.Name.Name,
			"reconcile_status":   r.Meta.ReconcileStatus.String(),
			"reconcile_error":    r.Meta.ReconcileError,
			"reconcile_warnings": r.Meta.ReconcileWarnings,
		})
	}
	return resources, "", parseWarnings, nil
}
