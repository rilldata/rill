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
	Diff                 string           `json:"diff,omitempty" jsonschema:"Diff of the file contents."`
	IsNewFile            bool             `json:"is_new_file,omitempty" jsonschema:"Indicates if the tool created a new file."`
	Resources            []map[string]any `json:"resources,omitempty" jsonschema:"The Rill resources declared in the file, if any."`
	ParseError           string           `json:"parse_error,omitempty" jsonschema:"Parse error encountered when parsing the file, if any."`
	CheckpointCommitHash string           `json:"checkpoint_commit_hash,omitempty" jsonschema:"The commit hash of the checkpoint just before writing any file in the current message chain."`
}

func (t *WriteFile) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        WriteFileName,
		Title:       "Write file",
		Description: "Creates, updates or deletes a file in a Rill project. If the file already exists, it will be overwritten. If the file declares a Rill resource, it will wait for the resource to reconcile and return its kind, name and any errors encountered.",
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

	checkpointCommitHash, err := t.maybeCreateCheckpoint(ctx, s)
	if err != nil {
		return nil, err
	}

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

		resources, parseErr, err = t.reconcileAndGetStatus(ctx, args.Path)
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
		Diff:                 diff,
		IsNewFile:            isNewFile,
		Resources:            resources,
		ParseError:           parseErr,
		CheckpointCommitHash: checkpointCommitHash,
	}, nil
}

// reconcileAndGetStatus waits until reconciliation is done, then returns the status of resources declared in the file at the given path.
func (t *WriteFile) reconcileAndGetStatus(ctx context.Context, path string) (resources []map[string]any, parseError string, err error) {
	s := GetSession(ctx)
	ctrl, err := t.Runtime.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, "", err
	}
	err = ctrl.Reconcile(ctx, runtime.GlobalProjectParserName) // TODO: Only if not streaming
	if err != nil {
		return nil, "", err
	}

	select {
	case <-time.After(time.Millisecond * 500):
	case <-ctx.Done():
		return nil, "", ctx.Err()
	}

	p, err := ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	if err != nil {
		return nil, "", err
	}
	for _, pe := range p.GetProjectParser().State.ParseErrors {
		if pe.FilePath == path {
			return nil, pe.Message, nil
		}
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if err != nil {
		return nil, "", err
	}

	rs, err := ctrl.List(ctx, "", path, false)
	if err != nil {
		return nil, "", err
	}
	if len(rs) == 0 {
		return nil, "", nil
	}

	resources = []map[string]any{}
	for _, r := range rs {
		resources = append(resources, map[string]any{
			"kind":             r.Meta.Name.Kind,
			"name":             r.Meta.Name.Name,
			"reconcile_status": r.Meta.ReconcileStatus.String(),
			"reconcile_error":  r.Meta.ReconcileError,
		})
	}
	return resources, "", nil
}

// maybeCreateCheckpoint creates a checkpoint if this is the 1st write file message in the current message chain.
func (t *WriteFile) maybeCreateCheckpoint(ctx context.Context, s *Session) (string, error) {
	// Find a write file message in the current message chain.
	var msg *Message
	for i := len(s.messages) - 1; i >= 0; i-- {
		if s.messages[i].Tool == WriteFileName && s.messages[i].Type == MessageTypeResult {
			msg = s.messages[i]
		}

		if s.messages[i].Role == RoleUser {
			break
		}
	}

	// If there is already a write file message then we dont need to create a checkpoint.
	if msg != nil {
		return "", nil
	}

	repo, release, err := t.Runtime.Repo(ctx, s.InstanceID())
	if err != nil {
		return "", err
	}
	defer release()

	// Get the status of the repo
	gitStatus, err := repo.Status(ctx)
	if err != nil {
		return "", err
	}

	// If there are local changes, commit them. Otherwise, just return the current commit hash.
	var hash string
	if gitStatus.LocalChanges {
		hash, err = repo.Commit(ctx, "Checkpoint") // TODO: message
	} else {
		hash, err = repo.CommitHash(ctx)
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}
