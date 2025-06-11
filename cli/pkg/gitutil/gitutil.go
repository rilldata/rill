package gitutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	exec "golang.org/x/sys/execabs"
)

var ErrGitRemoteNotFound = errors.New("no git remotes found")

type Config struct {
	Remote            string
	Username          string
	Password          string
	PasswordExpiresAt time.Time
	DefaultBranch     string
	Subpath           string
}

func (g *Config) IsExpired() bool {
	return g.Password != "" && g.PasswordExpiresAt.Before(time.Now())
}

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

// Remote represents a Git remote with its name and URL.
// The URL is normalized to a HTTPS URL with a .git suffix.
type Remote struct {
	Name string
	URL  string
}

// Github returns a normalized HTTPS Github URL ending in .git for the remote.
func (r Remote) Github() (string, error) {
	if r.URL == "" {
		return "", fmt.Errorf("remote %q has no URL", r.Name)
	}
	return NormalizeGithubRemote(r.URL)
}

// ExtractGitRemote extracts the first Git remote from the Git repository at projectPath.
// If remoteName is provided, it will return the remote with that name.
// If detectDotGit is true, it will look for a .git directory in parent directories.
func ExtractGitRemote(projectPath, remoteName string, detectDotGit bool) (Remote, error) {
	remotes, err := ExtractRemotes(projectPath, detectDotGit)
	if err != nil {
		return Remote{}, err
	}
	if remoteName != "" {
		for _, remote := range remotes {
			if remote.Name == remoteName {
				return remote, nil
			}
		}
		return Remote{}, ErrGitRemoteNotFound
	}
	if len(remotes) == 0 {
		return Remote{}, ErrGitRemoteNotFound
	}
	return remotes[0], nil
}

// ExtractRemotes extracts all Git remotes from the Git repository at projectPath.
// The returned remotes are normalized with NormalizeGithubRemote.
// If detectDotGit is true, it will look for a .git directory in parent directories.
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

func CommitAndForcePush(ctx context.Context, projectPath, remote, username, password, branch string, author *object.Signature, allowEmptyCommits bool) error {
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
	_, err = wt.Commit("Auto committed by Rill", &git.CommitOptions{All: true, Author: author, AllowEmptyCommits: allowEmptyCommits})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return fmt.Errorf("failed to commit files to git: %w", err)
		}
		// empty commit - nothing to cmmit
		return nil
	}

	// set remote
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remote},
	})
	if err != nil {
		if !errors.Is(err, git.ErrRemoteExists) {
			return fmt.Errorf("failed to create remote: %w", err)
		}
		// remote already exists do nothing we can override the URL while pushing
	}

	// push the changes
	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RemoteURL:  remote,
		Auth:       &githttp.BasicAuth{Username: username, Password: password},
		Force:      true,
	})
	if err != nil {
		return fmt.Errorf("failed to push to remote : %w", err)
	}
	return nil
}

func Clone(ctx context.Context, path string, c *Config) (*git.Repository, error) {
	return git.PlainCloneContext(ctx, path, false, &git.CloneOptions{
		URL:           c.Remote,
		Auth:          &githttp.BasicAuth{Username: c.Username, Password: c.Password},
		ReferenceName: plumbing.NewBranchReferenceName(c.DefaultBranch),
		SingleBranch:  true,
	})
}
