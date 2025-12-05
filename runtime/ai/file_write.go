package ai

import (
	"context"
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
	Contents string `json:"contents" jsonschema:"The new contents to write to the file. If the file already exists, this will overwrite it."`
}

type WriteFileResult struct {
	Diff       string           `json:"diff,omitempty"` // Unified diff (empty for new files)
	IsNewFile  bool             `json:"is_new_file"`    // True if file didn't exist before
	Resources  []map[string]any `json:"resources,omitempty"`
	ParseError string           `json:"parse_error,omitempty"`
}

func (t *WriteFile) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        WriteFileName,
		Title:       "Write file",
		Description: "Creates or updates a file in a Rill project. If the file already exists, it will be overwritten. If the file declares a Rill resource, it will wait for the resource to reconcile and return its kind, name and any errors encountered.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Writing file...",
			"openai/toolInvocation/invoked":  "Wrote file",
		},
	}
}

func (t *WriteFile) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAgentAccess(ctx, t.Runtime)
}

func (t *WriteFile) Handler(ctx context.Context, args *WriteFileArgs) (*WriteFileResult, error) {
	s := GetSession(ctx)

	if !strings.HasPrefix(args.Path, "/") {
		args.Path = "/" + args.Path
	}

	// Read existing content before writing (for diff computation)
	originalContent, _, err := t.Runtime.GetFile(ctx, s.InstanceID(), args.Path)
	isNewFile := err != nil // File doesn't exist if there's an error
	if isNewFile {
		originalContent = ""
	}

	err = t.Runtime.PutFile(ctx, s.InstanceID(), args.Path, strings.NewReader(args.Contents), true, false)
	if err != nil {
		return nil, err
	}

	// Compute unified diff (for new files, originalContent is empty so all lines show as additions)
	diff := computeUnifiedDiff(args.Path, originalContent, args.Contents, isNewFile)

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
				Diff:       diff,
				IsNewFile:  isNewFile,
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
		return &WriteFileResult{
			Diff:      diff,
			IsNewFile: isNewFile,
		}, nil
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
		Diff:      diff,
		IsNewFile: isNewFile,
		Resources: resources,
	}, nil
}

// computeUnifiedDiff generates a unified diff between original and new content
func computeUnifiedDiff(path, original, updated string, isNewFile bool) string {
	fromFile := "a" + path
	if isNewFile {
		fromFile = "/dev/null"
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(original),
		B:        difflib.SplitLines(updated),
		FromFile: fromFile,
		ToFile:   "b" + path,
		Context:  3,
	}
	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return ""
	}
	return text
}
