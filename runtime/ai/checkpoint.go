package ai

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const CheckpointName = "checkpoint"

type Checkpoint struct {
	Runtime *runtime.Runtime
}

var _ Tool[*CheckpointArgs, *CheckpointResult] = (*Checkpoint)(nil)

type CheckpointArgs struct{}

type CheckpointResult struct {
	SHA        string `json:"sha" jsonschema:"The commit SHA of the checkpoint"`
	Message    string `json:"message" jsonschema:"The commit message or status message"`
	HadChanges bool   `json:"had_changes" jsonschema:"Indicates if new changes were committed (true) or if returning existing HEAD SHA (false)"`
}

func (t *Checkpoint) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        CheckpointName,
		Title:       "Checkpoint",
		Description: "Creates a git checkpoint by committing any uncommitted changes, or returns the current commit SHA if there are no changes. The commit SHA is automatically stored in conversation history.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Creating checkpoint...",
			"openai/toolInvocation/invoked":  "Created checkpoint",
		},
	}
}

func (t *Checkpoint) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAgentAccess(ctx, t.Runtime)
}

// ensureGitRepo initializes a git repository if one doesn't exist.
// TODO: This function should be removed once git initialization is moved to `rill init`.
func ensureGitRepo(ctx context.Context, repoRoot string) error {
	// Check if git is initialized
	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "rev-parse", "--git-dir")
	if err := cmd.Run(); err == nil {
		// Git repo already exists
		return nil
	}

	// Initialize git repository
	cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "init")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user if not already set (needed for commits)
	// Check if user.name is set
	cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "config", "user.name")
	if err := cmd.Run(); err != nil {
		// Set default git user
		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "config", "user.name", "Rill AI")
		_ = cmd.Run() // Ignore error, it's not critical
	}

	// Check if user.email is set
	cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "config", "user.email")
	if err := cmd.Run(); err != nil {
		// Set default git email
		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "config", "user.email", "ai@rilldata.com")
		_ = cmd.Run() // Ignore error, it's not critical
	}

	return nil
}

func (t *Checkpoint) Handler(ctx context.Context, args *CheckpointArgs) (*CheckpointResult, error) {
	s := GetSession(ctx)

	// Get the repo
	repo, release, err := t.Runtime.Repo(ctx, s.InstanceID())
	if err != nil {
		return nil, fmt.Errorf("failed to access repository: %w", err)
	}
	defer release()

	// Get the repo root path
	repoRoot, err := repo.Root(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}

	// Ensure git repository is initialized
	// TODO: Remove this call once git initialization is moved to `rill init`
	if err := ensureGitRepo(ctx, repoRoot); err != nil {
		return nil, err
	}

	// Check if there are uncommitted changes using git status --porcelain
	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to check git status: %w", err)
	}

	hasChanges := len(strings.TrimSpace(string(output))) > 0

	// Check if there are any commits yet using git directly
	cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "rev-parse", "HEAD")
	shaOutput, err := cmd.Output()
	hasCommits := err == nil && len(strings.TrimSpace(string(shaOutput))) > 0

	// If there are changes OR if there are no commits yet (fresh git init), create a commit
	if hasChanges || !hasCommits {
		// Stage all changes
		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "add", "-A")
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to stage changes: %w", err)
		}

		// Create commit with auto-generated message
		timestamp := time.Now().UTC().Format(time.RFC3339)
		var commitMsg string
		if !hasCommits {
			commitMsg = fmt.Sprintf("Initial commit at %s", timestamp)
		} else {
			commitMsg = fmt.Sprintf("AI checkpoint at %s", timestamp)
		}

		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "commit", "-m", commitMsg, "--allow-empty")
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to commit changes: %w", err)
		}

		// Get the new commit SHA using git directly
		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "rev-parse", "HEAD")
		shaOutput, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get commit hash: %w", err)
		}
		sha := strings.TrimSpace(string(shaOutput))

		return &CheckpointResult{
			SHA:        sha,
			Message:    commitMsg,
			HadChanges: hasChanges,
		}, nil
	}

	// No changes - return current HEAD SHA using git directly
	cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "rev-parse", "HEAD")
	shaOutput, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit hash: %w", err)
	}
	currentSHA := strings.TrimSpace(string(shaOutput))

	return &CheckpointResult{
		SHA:        currentSHA,
		Message:    "No changes to commit",
		HadChanges: false,
	}, nil
}
