package gitutil

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	exec "golang.org/x/sys/execabs"
)

var ErrGitRemoteNotFound = errors.New("no git remotes found")

func CloneRepo(repoURL string) (string, error) {
	endpoint, err := transport.NewEndpoint(repoURL)
	if err != nil {
		return "", err
	}

	repoName := fileutil.Stem(endpoint.Path)
	cmd := exec.Command("git", "clone", repoURL)
	cmd.Stderr = os.Stderr
	_, err = cmd.Output()
	if err != nil {
		return "", err
	}

	return repoName, nil
}

type Remote struct {
	Name string
	URL  string
}

func ExtractRemotes(projectPath string, detectDotGit bool) ([]Remote, error) {
	repo, err := git.PlainOpenWithOptions(projectPath, &git.PlainOpenOptions{
		DetectDotGit: detectDotGit,
	})
	if err != nil {
		return nil, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}

	res := make([]Remote, len(remotes))
	for idx, remote := range remotes {
		if len(remote.Config().URLs) == 0 {
			return nil, fmt.Errorf("no URL found for git remote %q", remote.Config().Name)
		}

		res[idx] = Remote{
			Name: remote.Config().Name,
			// The first URL in the slice is the URL Git fetches from (main one).
			// We'll make things easy for ourselves and only consider that.
			URL: remote.Config().URLs[0],
		}
	}

	return res, nil
}

func RemotesToGithubURL(remotes []Remote) (*Remote, string, error) {
	// Return the first Github URL found.
	// If no Github remotes were found, return the first error.
	var firstErr error
	for _, remote := range remotes {
		ghurl, err := RemoteToGithubURL(remote.URL)
		if err == nil {
			// Found a Github remote. Success!
			return &remote, ghurl, nil
		}
		if firstErr == nil {
			firstErr = fmt.Errorf("invalid remote %q: %w", remote.URL, err)
		}
	}

	if firstErr == nil {
		return nil, "", ErrGitRemoteNotFound
	}

	return nil, "", firstErr
}

func RemoteToGithubURL(remote string) (string, error) {
	ep, err := transport.NewEndpoint(remote)
	if err != nil {
		return "", err
	}

	if ep.Host != "github.com" {
		return "", fmt.Errorf("must be a git remote on github.com")
	}

	account, repo := path.Split(ep.Path)
	account = strings.Trim(account, "/")
	repo = strings.TrimSuffix(repo, ".git")
	if account == "" || repo == "" || strings.Contains(account, "/") {
		return "", fmt.Errorf("not a valid github.com remote")
	}

	githubURL := &url.URL{
		Scheme: "https",
		Host:   ep.Host,
		Path:   strings.TrimSuffix(ep.Path, ".git"),
	}

	return githubURL.String(), nil
}

func SplitGithubURL(githubURL string) (account, repo string, ok bool) {
	ep, err := transport.NewEndpoint(githubURL)
	if err != nil {
		return "", "", false
	}

	if ep.Host != "github.com" {
		return "", "", false
	}

	account, repo = path.Split(ep.Path)
	account = strings.Trim(account, "/")
	if account == "" || repo == "" || strings.Contains(account, "/") {
		return "", "", false
	}

	return account, repo, true
}

func ExtractGitRemote(projectPath, remoteName string, detectDotGit bool) (*Remote, string, error) {
	remotes, err := ExtractRemotes(projectPath, detectDotGit)
	if err != nil {
		return nil, "", err
	}
	if remoteName != "" {
		for _, remote := range remotes {
			if remote.Name == remoteName {
				return RemotesToGithubURL([]Remote{remote})
			}
		}
	}

	// Parse into a https://github.com/account/repo (no .git) format
	return RemotesToGithubURL(remotes)
}

type SyncStatus int

const (
	SyncStatusUnspecified SyncStatus = iota
	SyncStatusModified               // Local branch has untracked/modified changes
	SyncStatusAhead                  // Local branch is ahead of remote branch
	SyncStatusSynced                 // Local branch is in sync with remote branch
)

// GetSyncStatus returns the status of current branch as compared to remote/branch
// TODO: Need to implement cases like local branch is behind/diverged from remote branch
func GetSyncStatus(repoPath, branch, remote string) (SyncStatus, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return SyncStatusUnspecified, err
	}

	ref, err := repo.Head()
	if err != nil {
		return SyncStatusUnspecified, err
	}

	if branch == "" {
		// try to infer default branch from local repo
		remoteRef, err := repo.Reference(plumbing.NewRemoteHEADReferenceName(remote), true)
		if err != nil {
			return SyncStatusUnspecified, err
		}

		_, branch, _ = strings.Cut(remoteRef.Name().Short(), fmt.Sprintf("%s/", remote))
	}

	// if user is not on required branch
	if !ref.Name().IsBranch() || ref.Name().Short() != branch {
		return SyncStatusUnspecified, fmt.Errorf("not on required branch")
	}

	w, err := repo.Worktree()
	if err != nil {
		if errors.Is(err, git.ErrIsBareRepository) {
			// no commits can be made in bare repository
			return SyncStatusSynced, nil
		}
		return SyncStatusUnspecified, err
	}

	repoStatus, err := w.Status()
	if err != nil {
		return SyncStatusUnspecified, err
	}

	// check all files are in unmodified state
	if !repoStatus.IsClean() {
		return SyncStatusModified, nil
	}

	// check if there are local commits not pushed to remote yet
	// no easy way to get it from go-git library so running git command directly and checking response
	cmd := exec.Command("git", "-C", repoPath, "log", "@{u}..")
	data, err := cmd.Output()
	if err != nil {
		return SyncStatusUnspecified, err
	}

	if len(data) != 0 {
		return SyncStatusAhead, nil
	}
	return SyncStatusSynced, nil
}

func CommitAndForcePush(ctx context.Context, projectPath, remote, username, password, branch string, author *object.Signature) error {
	// init git repo
	repo, err := git.PlainInitWithOptions(projectPath, &git.PlainInitOptions{
		InitOptions: git.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName(branch),
		},
		Bare: false,
	})
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return fmt.Errorf("failed to init git repo: %w", err)
		}
		repo, err = git.PlainOpen(projectPath)
		if err != nil {
			return fmt.Errorf("failed to open git repo: %w", err)
		}
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// git add .
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// git commit -m
	_, err = wt.Commit("Auto committed by Rill", &git.CommitOptions{All: true, Author: author})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return fmt.Errorf("failed to commit files to git: %w", err)
		}
		// empty commit - nothing to cmmit
		return nil
	}

	err = RunGitPush(ctx, projectPath, force, GitRemoteCredentials{
		Remote:   remote,
		Username: username,
		Password: password,
	})
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return fmt.Errorf("git push failed: %s", string(execErr.Stderr))
		}
		return err
	}
	return nil
}

type GitStatus struct {
	Branch        string
	LocalChanges  bool
	RemoteChanges bool
}

func RunGitStatus(path string) (*GitStatus, error) {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain=v2", "--branch")
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// parse the output
	// Format is
	// # branch.oid 4954f542d4b1f652bba02064aa8ee64ece38d02a
	// # branch.head mgd_repo_poc
	// # branch.upstream origin/mgd_repo_poc
	// # branch.ab +0 -0
	// lines describing the status of the working tree
	status := &GitStatus{}
	lines := strings.SplitSeq(strings.TrimSpace(string(data)), "\n")
	for line := range lines {
		switch {
		// standard headers - all may not be present
		case strings.HasPrefix(line, "# branch.oid "):
		case strings.HasPrefix(line, "# branch.head "):
			// Should handle detached state ?
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.upstream "):
		case strings.HasPrefix(line, "# branch.ab "):
			s := strings.Split(line, " ")

			ahead, err := strconv.Atoi(s[2])
			if err != nil {
				return nil, err
			}
			if ahead != 0 {
				status.LocalChanges = true
			}

			behind, err := strconv.Atoi(s[3])
			if err != nil {
				return nil, err
			}
			if behind != 0 {
				status.RemoteChanges = true
			}
		default:
			// any non header line means staged, unstaged or untracked changes
			status.LocalChanges = true
			return status, nil
		}
	}
	return status, nil
}

type GitRemoteCredentials struct {
	Remote   string
	Username string
	Password string
}

func (g GitRemoteCredentials) FullyQualifiedRemote() (string, error) {
	if g.Remote == "" {
		return "", fmt.Errorf("remote is not set")
	}
	u, err := url.Parse(g.Remote)
	if err != nil {
		return "", err
	}
	if g.Username != "" {
		if g.Password != "" {
			u.User = url.UserPassword(g.Username, g.Password)
		} else {
			u.User = url.User(g.Username)
		}
	}
	return u.String(), nil
}

func RunGitFetch(ctx context.Context, path string) error {
	args := []string{"-C", path, "fetch"}
	cmd := exec.CommandContext(ctx, "git", args...)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func RunGitPull(ctx context.Context, path string, discardLocal bool, g GitRemoteCredentials) (string, error) {
	if discardLocal {
		// instead of doing a hard clean, do a stash instead
		cmd := exec.CommandContext(ctx, "git", "-C", path, "stash", "--include-untracked")
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to remove local changes: %w", err)
		}
	}

	args := []string{"-C", path, "pull"}
	if g.Remote != "" {
		u, err := g.FullyQualifiedRemote()
		if err != nil {
			return "", err
		}
		args = append(args, u)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return "", fmt.Errorf("git pull failed: %s", string(execErr.Stderr))
		}
		return "", err
	}
	return string(out), nil
}

func RunGitPush(ctx context.Context, path string, force bool, g GitRemoteCredentials) error {
	args := []string{"-C", path, "push", "--set-upstream"}
	if force {
		args = append(args, "--force")
	}
	if g.Remote != "" {
		u, err := g.FullyQualifiedRemote()
		if err != nil {
			return err
		}
		args = append(args, u)
	}
	st, err := RunGitStatus(path)
	if err != nil {
		return err
	}
	// TODO :: handle detached state
	args = append(args, st.Branch)

	cmd := exec.CommandContext(ctx, "git", args...)
	_, err = cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return fmt.Errorf("git push failed: %s", string(execErr.Stderr))
		}
		return err
	}
	return nil
}

func RunGitClone(ctx context.Context, path, branch string, g GitRemoteCredentials) error {
	u, err := g.FullyQualifiedRemote()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, u, path)
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
