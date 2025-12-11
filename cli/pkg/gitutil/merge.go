package gitutil

import (
	"context"
	"fmt"
	"strings"

	exec "golang.org/x/sys/execabs"
)

// MergeResult contains the result of a merge operation
type MergeResult struct {
	Success          bool
	HasConflicts     bool
	ConflictingFiles []string
}

// GitMerge merges the source branch into the current branch
// This uses the git command directly for better conflict handling
func GitMerge(ctx context.Context, repoPath, sourceBranch string) (*MergeResult, error) {
	// First, try the merge
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "merge", sourceBranch, "--no-edit")
	output, err := cmd.CombinedOutput()

	if err == nil {
		// Merge succeeded
		return &MergeResult{
			Success:      true,
			HasConflicts: false,
		}, nil
	}

	// Check if it's a conflict
	outputStr := string(output)
	if strings.Contains(outputStr, "CONFLICT") || strings.Contains(outputStr, "Automatic merge failed") {
		// Get the list of conflicting files
		conflictingFiles := getConflictingFiles(ctx, repoPath)

		// Abort the merge to leave the repo in a clean state
		abortCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "merge", "--abort")
		_ = abortCmd.Run() // Ignore error, best effort

		return &MergeResult{
			Success:          false,
			HasConflicts:     true,
			ConflictingFiles: conflictingFiles,
		}, nil
	}

	// Some other error
	return nil, fmt.Errorf("merge failed: %s", outputStr)
}

// getConflictingFiles returns a list of files with conflicts
func getConflictingFiles(ctx context.Context, repoPath string) []string {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

// GeneratePullRequestURL generates a GitHub PR URL for creating a PR
func GeneratePullRequestURL(remoteURL, sourceBranch, targetBranch string) string {
	// Parse the remote URL to get owner and repo
	// Format: https://github.com/owner/repo.git or git@github.com:owner/repo.git
	if remoteURL == "" {
		return ""
	}

	var owner, repo string

	if strings.HasPrefix(remoteURL, "https://github.com/") {
		// HTTPS format
		parts := strings.TrimPrefix(remoteURL, "https://github.com/")
		parts = strings.TrimSuffix(parts, ".git")
		split := strings.Split(parts, "/")
		if len(split) >= 2 {
			owner = split[0]
			repo = split[1]
		}
	} else if strings.HasPrefix(remoteURL, "git@github.com:") {
		// SSH format
		parts := strings.TrimPrefix(remoteURL, "git@github.com:")
		parts = strings.TrimSuffix(parts, ".git")
		split := strings.Split(parts, "/")
		if len(split) >= 2 {
			owner = split[0]
			repo = split[1]
		}
	}

	if owner == "" || repo == "" {
		return ""
	}

	// Generate the compare URL for creating a PR
	return fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s?expand=1", owner, repo, targetBranch, sourceBranch)
}

