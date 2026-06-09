package file

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/filewatcher"
	rtgitutil "github.com/rilldata/rill/runtime/pkg/gitutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestListCommits(t *testing.T) {
	t.Run("returns commits in reverse chronological order with all fields populated", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 3)
		c := &connection{root: tempDir}

		commits, nextToken, err := c.ListCommits(context.Background(), "", 0)
		require.NoError(t, err)
		require.Empty(t, nextToken, "no next page token expected when limit is 0")
		require.Len(t, commits, 3)

		// Newest commit first.
		require.Equal(t, "commit 3", trimMessage(commits[0].CommitMessage))
		require.Equal(t, "commit 2", trimMessage(commits[1].CommitMessage))
		require.Equal(t, "commit 1", trimMessage(commits[2].CommitMessage))

		for _, commit := range commits {
			require.Len(t, commit.CommitSha, 40, "commit sha should be 40 chars")
			require.Equal(t, "Test User", commit.AuthorName)
			require.Equal(t, "test@rilldata.com", commit.AuthorEmail)
			require.NotNil(t, commit.CommittedOn)
			require.False(t, commit.CommittedOn.AsTime().IsZero())
		}
	})

	t.Run("respects limit and returns nextPageToken when more commits exist", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 5)
		c := &connection{root: tempDir}

		commits, nextToken, err := c.ListCommits(context.Background(), "", 2)
		require.NoError(t, err)
		require.Len(t, commits, 2)
		require.NotEmpty(t, nextToken, "next page token must be populated when more commits exist")
		require.Len(t, nextToken, 40)
	})

	t.Run("returns empty nextPageToken when no more commits", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 2)
		c := &connection{root: tempDir}

		commits, nextToken, err := c.ListCommits(context.Background(), "", 2)
		require.NoError(t, err)
		require.Len(t, commits, 2)
		require.Empty(t, nextToken)
	})

	t.Run("paging through all commits yields each commit exactly once", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 5)
		c := &connection{root: tempDir}

		var all []string
		token := ""
		for {
			commits, next, err := c.ListCommits(context.Background(), token, 2)
			require.NoError(t, err)
			for _, commit := range commits {
				all = append(all, trimMessage(commit.CommitMessage))
			}
			if next == "" {
				break
			}
			token = next
		}
		require.Equal(t, []string{"commit 5", "commit 4", "commit 3", "commit 2", "commit 1"}, all)
	})

	t.Run("pageToken starting from an arbitrary commit returns that commit and its ancestors", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 4)
		c := &connection{root: tempDir}

		all, _, err := c.ListCommits(context.Background(), "", 0)
		require.NoError(t, err)
		require.Len(t, all, 4)

		// Start from the 2nd-newest commit; should return 3 commits (it + 2 ancestors).
		commits, nextToken, err := c.ListCommits(context.Background(), all[1].CommitSha, 0)
		require.NoError(t, err)
		require.Empty(t, nextToken)
		require.Len(t, commits, 3)
		require.Equal(t, all[1].CommitSha, commits[0].CommitSha)
		require.Equal(t, all[2].CommitSha, commits[1].CommitSha)
		require.Equal(t, all[3].CommitSha, commits[2].CommitSha)
	})

	t.Run("preserves multi-line commit messages", func(t *testing.T) {
		tempDir := initRepo(t)
		commitWithMessage(t, tempDir, "file.txt", "x", "subject line\n\nbody line one\nbody line two\n")
		c := &connection{root: tempDir}

		commits, _, err := c.ListCommits(context.Background(), "", 0)
		require.NoError(t, err)
		require.Len(t, commits, 1)
		// git normalizes trailing whitespace but preserves the body structure.
		require.Contains(t, commits[0].CommitMessage, "subject line")
		require.Contains(t, commits[0].CommitMessage, "body line one")
		require.Contains(t, commits[0].CommitMessage, "body line two")
	})

	t.Run("preserves commit messages containing ASCII unit and record separators", func(t *testing.T) {
		tempDir := initRepo(t)
		// These bytes are what our parsing uses internally; they must survive round-trip via `-z` mode.
		message := "weird message with \x1f and \x1e bytes inside"
		commitWithMessage(t, tempDir, "file.txt", "x", message)
		c := &connection{root: tempDir}

		commits, _, err := c.ListCommits(context.Background(), "", 0)
		require.NoError(t, err)
		require.Len(t, commits, 1)
		require.Contains(t, commits[0].CommitMessage, message)
	})

	t.Run("populates CommittedOn from the committer timestamp", func(t *testing.T) {
		tempDir := initRepo(t)
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "f.txt"), []byte("x"), 0644))
		runGit(t, tempDir, "add", "f.txt")

		// Pin author and committer dates to a known value.
		fixed := "2024-01-02T03:04:05+00:00"
		cmd := exec.Command("git", "-C", tempDir, "commit", "-m", "fixed-time commit")
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_DATE="+fixed,
			"GIT_COMMITTER_DATE="+fixed,
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))

		c := &connection{root: tempDir}
		commits, _, err := c.ListCommits(context.Background(), "", 0)
		require.NoError(t, err)
		require.Len(t, commits, 1)

		expected, err := time.Parse(time.RFC3339, fixed)
		require.NoError(t, err)
		require.True(t, commits[0].CommittedOn.AsTime().Equal(expected),
			"expected committed_on=%s, got %s", expected, commits[0].CommittedOn.AsTime())
	})
}

func TestListBranches(t *testing.T) {
	t.Run("lists local branches and reports the current branch", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		runGit(t, tempDir, "branch", "feature")
		runGit(t, tempDir, "branch", "experimental")
		c := newFileConnection(t, tempDir)

		branches, current, err := c.ListBranches(context.Background())
		require.NoError(t, err)
		require.Equal(t, "main", current)
		require.ElementsMatch(t, []string{"main", "feature", "experimental"}, branches)
	})

	t.Run("includes remote branches and dedupes against locals", func(t *testing.T) {
		remote := setupBareRemote(t)
		local := setupRepoWithCommits(t, 1)
		runGit(t, local, "remote", "add", "origin", remote)
		runGit(t, local, "push", "origin", "main")

		// Add a branch only on the remote.
		runGit(t, local, "checkout", "-b", "remote-only")
		commitWithMessage(t, local, "f.txt", "x", "remote-only commit")
		runGit(t, local, "push", "origin", "remote-only")
		runGit(t, local, "checkout", "main")
		runGit(t, local, "branch", "-D", "remote-only")
		runGit(t, local, "fetch", "origin")

		c := newFileConnection(t, local)
		branches, current, err := c.ListBranches(context.Background())
		require.NoError(t, err)
		require.Equal(t, "main", current)
		require.Contains(t, branches, "main")
		require.Contains(t, branches, "remote-only", "remote-only branch should be surfaced via refs/remotes/origin/")
		// HEAD ref should not leak through as a branch name.
		require.NotContains(t, branches, "HEAD")
	})
}

func TestSwitchBranch(t *testing.T) {
	t.Run("switches to an existing branch", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		runGit(t, tempDir, "branch", "feature")
		c := &connection{root: tempDir}

		require.NoError(t, c.SwitchBranch(context.Background(), "feature", false, false))
		require.Equal(t, "feature", currentBranchName(t, tempDir))
	})

	t.Run("returns ErrRefNotFound when branch is missing and create is false", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		c := &connection{root: tempDir}

		err := c.SwitchBranch(context.Background(), "does-not-exist", false, false)
		require.ErrorIs(t, err, rtgitutil.ErrRefNotFound)
	})

	t.Run("creates the branch when createIfNotExists is true", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		c := &connection{root: tempDir}

		require.NoError(t, c.SwitchBranch(context.Background(), "new-branch", true, false))
		require.Equal(t, "new-branch", currentBranchName(t, tempDir))
	})

	t.Run("force discards local changes when switching", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		runGit(t, tempDir, "branch", "feature")
		// Dirty the working tree on main.
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("local mutation"), 0644))
		c := &connection{root: tempDir}

		require.NoError(t, c.SwitchBranch(context.Background(), "feature", false, true))
		require.Equal(t, "feature", currentBranchName(t, tempDir))
		// After force checkout the working tree should match the branch, not the local mutation.
		content, err := os.ReadFile(filepath.Join(tempDir, "file1.txt"))
		require.NoError(t, err)
		require.Equal(t, "content 1", string(content))
	})
}

func TestCommit(t *testing.T) {
	t.Run("commits all working-tree changes and returns the commit hash", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "new.txt"), []byte("hello"), 0644))
		c := newFileConnection(t, tempDir)

		hash, err := c.Commit(context.Background(), "add new.txt")
		require.NoError(t, err)
		require.Len(t, hash, 40)

		// The file should be present in the latest commit.
		files := strings.TrimSpace(runGitOutput(t, tempDir, "show", "--name-only", "--pretty=format:", "HEAD"))
		require.Contains(t, files, "new.txt")
	})

	t.Run("uses the default message when message is empty", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "new.txt"), []byte("hello"), 0644))
		c := newFileConnection(t, tempDir)

		_, err := c.Commit(context.Background(), "")
		require.NoError(t, err)
		msg := strings.TrimSpace(runGitOutput(t, tempDir, "log", "-1", "--pretty=format:%s"))
		require.Equal(t, "Auto committed by Rill", msg)
	})

	t.Run("returns empty hash and no error when there is nothing to commit", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		c := newFileConnection(t, tempDir)

		hash, err := c.Commit(context.Background(), "noop")
		require.NoError(t, err)
		require.Empty(t, hash)
	})
}

func TestMergeToBranch(t *testing.T) {
	t.Run("merges current branch into the target and restores the original branch", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		runGit(t, tempDir, "checkout", "-b", "feature")
		commitWithMessage(t, tempDir, "feature.txt", "feature content", "add feature")

		c := newFileConnection(t, tempDir)
		require.NoError(t, c.MergeToBranch(context.Background(), "main", false))

		// We started on feature, so we should still be on feature after the call returns.
		require.Equal(t, "feature", currentBranchName(t, tempDir))
		// And main should now contain feature.txt.
		runGit(t, tempDir, "checkout", "main")
		require.FileExists(t, filepath.Join(tempDir, "feature.txt"))
	})

	t.Run("returns MergeFailedError on conflicts and aborts the merge", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		// Diverge feature and main on the same file.
		runGit(t, tempDir, "checkout", "-b", "feature")
		commitWithMessage(t, tempDir, "file1.txt", "feature side", "feature edit")
		runGit(t, tempDir, "checkout", "main")
		commitWithMessage(t, tempDir, "file1.txt", "main side", "main edit")
		runGit(t, tempDir, "checkout", "feature")

		c := newFileConnection(t, tempDir)
		err := c.MergeToBranch(context.Background(), "main", false)
		var mergeErr *drivers.MergeFailedError
		require.ErrorAs(t, err, &mergeErr)
		require.Equal(t, "feature", mergeErr.MergedBranch, "MergedBranch should refer to the incoming branch")

		// Main should not have been polluted by an abandoned merge state.
		runGit(t, tempDir, "checkout", "main")
		content, err := os.ReadFile(filepath.Join(tempDir, "file1.txt"))
		require.NoError(t, err)
		require.Equal(t, "main side", string(content))
		require.NoFileExists(t, filepath.Join(tempDir, ".git", "MERGE_HEAD"))
	})

	t.Run("force=true picks the current branch's version", func(t *testing.T) {
		tempDir := setupRepoWithCommits(t, 1)
		runGit(t, tempDir, "checkout", "-b", "feature")
		commitWithMessage(t, tempDir, "file1.txt", "feature wins", "feature edit")
		runGit(t, tempDir, "checkout", "main")
		commitWithMessage(t, tempDir, "file1.txt", "main loses", "main edit")
		runGit(t, tempDir, "checkout", "feature")

		c := newFileConnection(t, tempDir)
		require.NoError(t, c.MergeToBranch(context.Background(), "main", true))

		runGit(t, tempDir, "checkout", "main")
		content, err := os.ReadFile(filepath.Join(tempDir, "file1.txt"))
		require.NoError(t, err)
		require.Equal(t, "feature wins", string(content), "force merge should resolve in favour of the incoming (current) branch")
	})
}

func setupRepoWithCommits(t *testing.T, n int) string {
	t.Helper()
	tempDir := initRepo(t)
	for i := 1; i <= n; i++ {
		commitWithMessage(t, tempDir,
			fmt.Sprintf("file%d.txt", i),
			fmt.Sprintf("content %d", i),
			fmt.Sprintf("commit %d", i),
		)
	}
	return tempDir
}

func initRepo(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	runGit(t, "", "init", tempDir)
	runGit(t, tempDir, "checkout", "-b", "main")
	runGit(t, tempDir, "config", "user.name", "Test User")
	runGit(t, tempDir, "config", "user.email", "test@rilldata.com")
	runGit(t, tempDir, "config", "commit.gpgsign", "false")
	return tempDir
}

func commitWithMessage(t *testing.T, repoPath, fileName, fileContent, message string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(repoPath, fileName), []byte(fileContent), 0644))
	runGit(t, repoPath, "add", fileName)
	runGit(t, repoPath, "commit", "-m", message)
}

func runGit(t *testing.T, repoPath string, args ...string) {
	t.Helper()
	runGitOutput(t, repoPath, args...)
}

func trimMessage(s string) string {
	// `git log %B` includes a trailing newline; tests compare against the subject line.
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}

// newFileConnection builds a *connection sufficient for exercising Status/Commit/MergeToBranch
// without spinning up the full driver Open path (which would require admin auth / a watcher).
func newFileConnection(t *testing.T, root string) *connection {
	t.Helper()
	return &connection{
		logger:       zap.NewNop(),
		root:         root,
		driverConfig: &configProperties{HomeDir: t.TempDir()},
		driverName:   "file",
		dotRill:      dotrill.New(t.TempDir()),
		watcher:      filewatcher.NewLazyWatcher(root, nil, zap.NewNop()),
	}
}

// setupBareRemote creates a bare git repository in a temp directory and returns its path.
// Suitable as a `git remote add origin <path>` target for tests.
func setupBareRemote(t *testing.T) string {
	t.Helper()
	remote := t.TempDir()
	runGit(t, "", "init", "--bare", remote)
	return remote
}

func currentBranchName(t *testing.T, repoPath string) string {
	t.Helper()
	return strings.TrimSpace(runGitOutput(t, repoPath, "rev-parse", "--abbrev-ref", "HEAD"))
}

func runGitOutput(t *testing.T, repoPath string, args ...string) string {
	t.Helper()
	full := args
	if repoPath != "" {
		full = append([]string{"-C", repoPath}, args...)
	}
	cmd := exec.Command("git", full...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, string(out))
	return string(out)
}
