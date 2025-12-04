package gitutil

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

var (
	ErrGitRemoteNotFound = errors.New("no git remotes found")
	ErrNotAGitRepository = errors.New("not a git repository")
)

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

func CommitAndPush(ctx context.Context, projectPath string, config *Config, commitMsg string, author *object.Signature) error {
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

	// check current branch matches deployed branch
	headRef, err := repo.Head()
	if err == nil {
		if !headRef.Name().IsBranch() {
			return fmt.Errorf("detached HEAD state detected. Checkout a branch")
		}
		branch := headRef.Name().Short()
		if headRef.Name().Short() != config.DefaultBranch {
			return fmt.Errorf("current branch %q does not match deployed branch %q", branch, config.DefaultBranch)
		}
	} else if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		// ErrReferenceNotFound happens when looking for HEAD on a fresh repo
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// git add subpath/**
	var stagingPath string
	if config.Subpath != "" {
		stagingPath = filepath.Join(config.Subpath, "**")
	} else {
		stagingPath = "."
	}
	if err := wt.AddWithOptions(&git.AddOptions{Glob: stagingPath}); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// git commit -m
	if commitMsg == "" {
		commitMsg = "Auto committed by Rill"
	}
	_, err = wt.Commit(commitMsg, &git.CommitOptions{Author: author, AllowEmptyCommits: true})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return fmt.Errorf("failed to commit files to git: %w", err)
		}
		// empty commit - nothing to cmmit
		return nil
	}

	// set remote and push the changes
	err = SetRemote(projectPath, config)
	if err != nil {
		return err
	}

	if config.Username == "" {
		// If no credentials are provided we assume that is user's self managed repo and auth is already set in git
		// go-git does not support pushing to a private repo without auth so we will trigger the git command directly
		return RunGitPush(ctx, projectPath, config.RemoteName(), config.DefaultBranch)
	}

	u, err := url.Parse(config.Remote)
	if err != nil {
		return fmt.Errorf("failed to parse remote URL: %w", err)
	}
	u.User = url.UserPassword(config.Username, config.Password)
	return RunGitPush(ctx, projectPath, u.String(), config.DefaultBranch)
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
	if config == nil || config.Username == "" {
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
		if remote.Config().URLs[0] == config.Remote || !config.ManagedRepo {
			// remote already exists with the same URL, no need to create it again
			// remote other than managed git exists, can't overwrite user's remote
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
	_, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	return err == nil
}

// InferRepoRootAndSubpath infers the root of the Git repository and the subpath from the given path.
// Since the extraction stops at first .git directory it means that if a subpath in a github monorepo is deployed as a rill managed project it will prevent the subpath from being inferred.
// This means :
// - user will need to explicitly set the subpath if they want to connect this to Github.
// - When finding matching projects it will only list the rill managed projects for that subpath.
func InferRepoRootAndSubpath(path string) (string, string, error) {
	// check if is a git repository
	repoRoot, err := InferGitRepoRoot(path)
	if err != nil {
		return "", "", err
	}

	// infer subpath if it exists
	subPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		// should never happen because repoRoot is detected from path
		return "", "", err
	}
	if subPath == "." || subPath == "" {
		// no subpath
		return repoRoot, "", nil
	}
	// check if subpath is in .gitignore
	ignored, err := isGitIgnored(repoRoot, subPath)
	if err != nil {
		return "", "", err
	}
	if ignored {
		// if subpath is ignored this is not a valid git path
		return "", "", ErrNotAGitRepository
	}
	return repoRoot, subPath, nil
}
