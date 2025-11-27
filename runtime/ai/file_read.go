package ai

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ReadFileName = "read_file"

type ReadFile struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ReadFileArgs, *ReadFileResult] = (*ReadFile)(nil)

type ReadFileArgs struct {
	Path string `json:"path" jsonschema:"The path of the file to read"`
}

type ReadFileResult struct {
	Contents string
}

func (t *ReadFile) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ReadFileName,
		Title:       "Read file",
		Description: "Reads the contents of a file in the Rill project",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Reading file...",
			"openai/toolInvocation/invoked":  "Read file",
		},
	}
}

func (t *ReadFile) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAgentAccess(ctx, t.Runtime)
}

func (t *ReadFile) Handler(ctx context.Context, args *ReadFileArgs) (*ReadFileResult, error) {
	s := GetSession(ctx)

	if !strings.HasPrefix(args.Path, "/") {
		args.Path = "/" + args.Path
	}

	blob, _, err := t.Runtime.GetFile(ctx, s.InstanceID(), args.Path)
	if err != nil {
		return nil, err
	}

	return &ReadFileResult{Contents: blob}, nil
}
