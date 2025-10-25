package ai

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

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
		Name:        "read_file",
		Title:       "Read file",
		Description: "Reads the contents of a file in the Rill project",
	}
}

func (t *ReadFile) CheckAccess(ctx context.Context) bool {
	// NOTE: Disabled pending further improvements
	// s := GetSession(ctx)
	// return s.Claims().Can(runtime.ReadRepo)
	return false
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
