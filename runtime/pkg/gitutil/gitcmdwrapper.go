package gitutil

import (
	"errors"
	"fmt"
	"os/exec"
)

// MergeWithTheirsStrategy merge a branch into the current branch using the "theirs" strategy.
func MergeWithTheirsStrategy(path, branch string) error {
	cmd := exec.Command("git", "-C", path, "merge", "-X", "theirs", branch)
	_, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return fmt.Errorf("git merge failed: %s", string(execErr.Stderr))
		}
		return err
	}
	return nil
}

// MergeWithBailOnConflict attempts to merge a branch into the current branch and aborts if there are conflicts.
// Returns true if merge was successful, false if there were conflicts (but abort succeeded).
// Returns an error only if both merge and abort fail.
func MergeWithBailOnConflict(path, branch string) (bool, error) {
	// First try the merge
	cmd := exec.Command("git", "-C", path, "merge", "--no-ff", branch)
	_, err := cmd.Output()
	if err == nil {
		// Merge succeeded
		return true, nil
	}

	// Merge failed, try to abort
	abortCmd := exec.Command("git", "-C", path, "merge", "--abort")
	abortErr := abortCmd.Run()
	if abortErr != nil {
		// Both merge and abort failed
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return false, fmt.Errorf("git merge failed and abort failed: %s", string(execErr.Stderr))
		}
		return false, fmt.Errorf("git merge failed and abort failed: %w", err)
	}

	// Merge failed but abort succeeded
	return false, nil
}
