package gitutil

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
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
	ManagedRepo       bool
}

func (g *Config) IsExpired() bool {
	return g.Password != "" && g.PasswordExpiresAt.Before(time.Now())
}

func (g *Config) FullyQualifiedRemote() (string, error) {
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

func (g *Config) RemoteName() string {
	if g.ManagedRepo {
		return "__rill_remote"
	}
	return "origin"
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

func CommitAndForcePush(ctx context.Context, projectPath string, config *Config, commitMsg string, author *object.Signature) error {
	// init git repo
	repo, err := git.PlainInitWithOptions(projectPath, &git.PlainInitOptions{
		InitOptions: git.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName(config.DefaultBranch),
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
	if commitMsg == "" {
		commitMsg = "Auto committed by Rill"
	}
	_, err = wt.Commit(commitMsg, &git.CommitOptions{All: true, Author: author, AllowEmptyCommits: true})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return fmt.Errorf("failed to commit files to git: %w", err)
		}
		// empty commit - nothing to cmmit
		return nil
	}

	if config.Username == "" {
		// If no credentials are provided we assume that is user's self managed repo and auth is already set in git
		// go-git does not support pushing to a private repo without auth so we will trigger the git command directly
		return RunGitPush(ctx, projectPath, config.RemoteName(), config.DefaultBranch)
	}

	// set remote and push the changes
	err = SetRemote(projectPath, config)
	if err != nil {
		return err
	}
	pushOpts := &git.PushOptions{
		RemoteName: config.RemoteName(),
		RemoteURL:  config.Remote,
		Force:      true,
	}
	if config.Username != "" && config.Password != "" {
		pushOpts.Auth = &githttp.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}
	}
	err = repo.PushContext(ctx, pushOpts)
	if err != nil {
		return fmt.Errorf("failed to push to remote : %w", err)
	}
	return nil
}

func Clone(ctx context.Context, path string, c *Config) (*git.Repository, error) {
	return git.PlainCloneContext(ctx, path, false, &git.CloneOptions{
		URL:           c.Remote,
		RemoteName:    c.RemoteName(),
		Auth:          &githttp.BasicAuth{Username: c.Username, Password: c.Password},
		ReferenceName: plumbing.NewBranchReferenceName(c.DefaultBranch),
		SingleBranch:  true,
	})
}

func NativeGitSignature(ctx context.Context, path string) (*object.Signature, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	cfg, err := repo.ConfigScoped(gitConfig.SystemScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get git config: %w", err)
	}
	if cfg.User.Email != "" && cfg.User.Name != "" {
		// user has git properly configured use that
		return &object.Signature{
			Name:  cfg.User.Name,
			Email: cfg.User.Email,
			When:  time.Now(),
		}, nil
	}
	return nil, fmt.Errorf("git user email or name is not set in git config")
}

func GitFetch(ctx context.Context, path string, config *Config) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}
	if config == nil {
		// uses default git configuration
		// go-git does not support fetching from a private repo without auth
		// so we will trigger the git command directly
		return RunGitFetch(ctx, path, "origin")
	}
	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: config.RemoteName(),
		RemoteURL:  config.Remote,
		Auth: &githttp.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		},
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			// no new changes to fetch, this is not an error
			return nil
		}
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}
	return nil
}

// SetRemote sets the remote by name Rill for the given repository to the provided remote URL.
func SetRemote(path string, config *Config) error {
	if config.Remote == "" {
		return nil
	}
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	remote, err := repo.Remote(config.RemoteName())
	if err != nil && !errors.Is(err, git.ErrRemoteNotFound) {
		return fmt.Errorf("failed to get remote: %w", err)
	}
	if remote != nil {
		if remote.Config().URLs[0] == config.Remote {
			// remote already exists with the same URL, no need to create it again
			return nil
		}
		// if the remote already exists with a different URL, delete it
		err = repo.DeleteRemote(config.RemoteName())
		if err != nil {
			return fmt.Errorf("failed to delete existing remote: %w", err)
		}
	}

	_, err = repo.CreateRemote(&gitConfig.RemoteConfig{
		Name: config.RemoteName(),
		URLs: []string{config.Remote},
	})
	return err
}

func IsGitRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
