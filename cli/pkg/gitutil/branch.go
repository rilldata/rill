package gitutil

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// BranchInfo contains information about a git branch
type BranchInfo struct {
	Name              string
	IsLocal           bool
	IsRemote          bool
	IsCurrent         bool
	LastCommitHash    string
	LastCommitMessage string
	LastCommitTime    time.Time
	Ahead             int32
	Behind            int32
}

// ListBranchesResult contains the result of ListBranches operation
type ListBranchesResult struct {
	Branches              []BranchInfo
	CurrentBranch         string
	HasUncommittedChanges bool
}

// ListBranches returns all local and remote branches in the repository
func ListBranches(repoPath string) (*ListBranchesResult, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current branch
	headRef, err := repo.Head()
	currentBranch := ""
	if err == nil && headRef.Name().IsBranch() {
		currentBranch = headRef.Name().Short()
	}

	// Track branches we've seen (to merge local/remote info)
	branchMap := make(map[string]*BranchInfo)

	// Get local branches
	localBranches, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to list local branches: %w", err)
	}

	err = localBranches.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return nil
		}

		branchMap[branchName] = &BranchInfo{
			Name:              branchName,
			IsLocal:           true,
			IsRemote:          false,
			IsCurrent:         branchName == currentBranch,
			LastCommitHash:    ref.Hash().String()[:7],
			LastCommitMessage: firstLine(commit.Message),
			LastCommitTime:    commit.Author.When,
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate local branches: %w", err)
	}

	// Get remote branches
	remoteBranches, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to list references: %w", err)
	}

	err = remoteBranches.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Name().IsRemote() {
			return nil
		}
		// Extract branch name from remote ref (e.g., "origin/main" -> "main")
		remoteName := ref.Name().Short()
		// Skip HEAD refs
		if remoteName == "origin/HEAD" {
			return nil
		}
		// Get just the branch name part
		branchName := remoteName
		if len(remoteName) > 7 && remoteName[:7] == "origin/" {
			branchName = remoteName[7:]
		}

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return nil
		}

		if existing, ok := branchMap[branchName]; ok {
			existing.IsRemote = true
		} else {
			branchMap[branchName] = &BranchInfo{
				Name:              branchName,
				IsLocal:           false,
				IsRemote:          true,
				IsCurrent:         false,
				LastCommitHash:    ref.Hash().String()[:7],
				LastCommitMessage: firstLine(commit.Message),
				LastCommitTime:    commit.Author.When,
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate remote branches: %w", err)
	}

	// Convert map to slice
	branches := make([]BranchInfo, 0, len(branchMap))
	for _, branch := range branchMap {
		branches = append(branches, *branch)
	}

	// Check for uncommitted changes
	wt, err := repo.Worktree()
	hasUncommittedChanges := false
	if err == nil {
		status, err := wt.Status()
		if err == nil {
			hasUncommittedChanges = !status.IsClean()
		}
	}

	return &ListBranchesResult{
		Branches:              branches,
		CurrentBranch:         currentBranch,
		HasUncommittedChanges: hasUncommittedChanges,
	}, nil
}

// CheckoutBranch switches to the specified branch
func CheckoutBranch(repoPath, branchName string, force bool) error {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Check for uncommitted changes if not forcing
	if !force {
		status, err := wt.Status()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}
		if !status.IsClean() {
			return fmt.Errorf("you have uncommitted changes. Commit or stash them before switching branches, or use force")
		}
	}

	// Try to checkout as local branch first
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Force:  force,
	})
	if err != nil {
		// If local branch doesn't exist, try to checkout from remote
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branchName),
			Create: true,
			Force:  force,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch %q: %w", branchName, err)
		}
	}

	return nil
}

// CreateBranch creates a new branch from the current HEAD
// Uncommitted changes will be preserved and carried to the new branch
func CreateBranch(repoPath, branchName string, checkout bool) (*BranchInfo, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current HEAD
	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create the new branch reference
	branchRef := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(branchRef, headRef.Hash())
	err = repo.Storer.SetReference(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch %q: %w", branchName, err)
	}

	// Get commit info for the response
	commit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	branchInfo := &BranchInfo{
		Name:              branchName,
		IsLocal:           true,
		IsRemote:          false,
		IsCurrent:         checkout,
		LastCommitHash:    headRef.Hash().String()[:7],
		LastCommitMessage: firstLine(commit.Message),
		LastCommitTime:    commit.Author.When,
	}

	// Checkout if requested - use Keep to preserve uncommitted changes
	if checkout {
		wt, err := repo.Worktree()
		if err != nil {
			return branchInfo, fmt.Errorf("branch created but failed to get worktree: %w", err)
		}
		// Keep: true preserves uncommitted changes in the worktree
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Keep:   true, // Preserve uncommitted changes
		})
		if err != nil {
			return branchInfo, fmt.Errorf("branch created but failed to checkout: %w", err)
		}
	}

	return branchInfo, nil
}

// GitCommit creates a new commit with all staged and unstaged changes
func GitCommit(repoPath, message, authorName, authorEmail string) (*CommitInfo, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Check if there are any changes to commit
	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	if status.IsClean() {
		return nil, fmt.Errorf("nothing to commit, working tree clean")
	}

	// Add all changes (staged and unstaged)
	err = wt.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create the commit
	commitHash, err := wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}

	// Get the commit object for response
	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	return &CommitInfo{
		Hash:        commit.Hash.String(),
		ShortHash:   commit.Hash.String()[:7],
		Message:     firstLine(commit.Message),
		AuthorName:  commit.Author.Name,
		AuthorEmail: commit.Author.Email,
		Timestamp:   commit.Author.When,
	}, nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(repoPath, branchName string, force bool) error {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Check if trying to delete current branch
	headRef, err := repo.Head()
	if err == nil && headRef.Name().Short() == branchName {
		return fmt.Errorf("cannot delete the currently checked out branch")
	}

	branchRef := plumbing.NewBranchReferenceName(branchName)

	// Check if branch exists
	_, err = repo.Reference(branchRef, true)
	if err != nil {
		return fmt.Errorf("branch %q does not exist", branchName)
	}

	// Delete the branch
	err = repo.Storer.RemoveReference(branchRef)
	if err != nil {
		return fmt.Errorf("failed to delete branch %q: %w", branchName, err)
	}

	return nil
}

// GetCommitHistory returns the commit history for a branch
func GetCommitHistory(repoPath, branch string, limit, offset int) ([]CommitInfo, int, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open repository: %w", err)
	}

	// If no branch specified, use current HEAD
	var startHash plumbing.Hash
	if branch == "" {
		headRef, err := repo.Head()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get HEAD: %w", err)
		}
		startHash = headRef.Hash()
	} else {
		// Try local branch first, then remote
		ref, err := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
		if err != nil {
			ref, err = repo.Reference(plumbing.NewRemoteReferenceName("origin", branch), true)
			if err != nil {
				return nil, 0, fmt.Errorf("branch %q not found: %w", branch, err)
			}
		}
		startHash = ref.Hash()
	}

	// Default limit
	if limit <= 0 {
		limit = 50
	}

	// Get commit iterator
	iter, err := repo.Log(&git.LogOptions{
		From: startHash,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get log: %w", err)
	}

	var commits []CommitInfo
	count := 0
	totalCount := 0

	err = iter.ForEach(func(c *object.Commit) error {
		totalCount++

		// Skip offset commits
		if totalCount <= offset {
			return nil
		}

		// Stop if we've collected enough
		if count >= limit {
			return nil
		}

		commits = append(commits, CommitInfo{
			Hash:        c.Hash.String(),
			ShortHash:   c.Hash.String()[:7],
			Message:     firstLine(c.Message),
			AuthorName:  c.Author.Name,
			AuthorEmail: c.Author.Email,
			Timestamp:   c.Author.When,
		})
		count++
		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, totalCount, nil
}

// CommitInfo contains information about a commit
type CommitInfo struct {
	Hash        string
	ShortHash   string
	Message     string
	AuthorName  string
	AuthorEmail string
	Timestamp   time.Time
}

// firstLine returns the first line of a string
func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' || c == '\r' {
			return s[:i]
		}
	}
	return s
}

