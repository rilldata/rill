package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/go-github/v71/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/gitutil"
)

// gitCommitSignature is the author used for all Rill-managed commits.
var gitCommitSignature = gitutil.Signature{Name: "Rill", Email: "noreply@rilldata.com"}

// FileChange is a single file to write or delete in a Git commit.
// A nil Data means the file should be deleted.
type FileChange struct {
	Path string // repo-relative path, e.g. "user_files/<seg>/alerts/x.yaml"
	Data []byte
}

// GitWriteOptions configures a GitWriter.WriteFiles call.
type GitWriteOptions struct {
	// Config describes the remote and (ephemeral) credentials. Config.Remote is the clean URL;
	// Config.Username/Password hold the short-lived token. Credentials are never persisted to .git/config.
	Config *gitutil.Config
	// Branch is the branch to commit to and push.
	Branch string
	// BaseBranch is used to create Branch if it does not yet exist on the remote (e.g. the primary branch).
	BaseBranch string
	// Subpath is the project's subpath within the repo (for monorepos). Files are written relative to it.
	Subpath string
	// Message is the commit message.
	Message string
	// Changes is the set of files to write or delete.
	Changes []FileChange
}

// GitWriter commits files into project Git repos on behalf of the admin service.
// It serializes writes per (remote, branch) within the process; combined with the sync job being
// unique per project, this keeps concurrent writes to the same branch from racing on push.
type GitWriter struct {
	locks sync.Map // key: remote+"\x00"+branch -> *sync.Mutex
}

// NewGitWriter returns a new GitWriter.
func NewGitWriter() *GitWriter {
	return &GitWriter{}
}

func (w *GitWriter) lockFor(remote, branch string) *sync.Mutex {
	key := remote + "\x00" + branch
	mu, _ := w.locks.LoadOrStore(key, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// WriteFiles applies the given changes as a single commit on opts.Branch and pushes it.
// It returns the resulting commit hash. If there is nothing to commit, it returns ErrEmptyCommit.
func (w *GitWriter) WriteFiles(ctx context.Context, opts *GitWriteOptions) (string, error) {
	if opts.Config == nil || opts.Config.Remote == "" {
		return "", errors.New("gitwriter: remote is not configured")
	}
	if opts.Branch == "" {
		return "", errors.New("gitwriter: branch is required")
	}

	mu := w.lockFor(opts.Config.Remote, opts.Branch)
	mu.Lock()
	defer mu.Unlock()

	tokenURL, err := opts.Config.FullyQualifiedRemote()
	if err != nil {
		return "", err
	}

	dir, err := os.MkdirTemp("", "rill-userfiles-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	// Clone the target branch; if it doesn't exist yet, branch off the base branch.
	if err := gitutil.Clone(ctx, dir, tokenURL, opts.Branch, true, true); err != nil {
		if opts.BaseBranch == "" || opts.BaseBranch == opts.Branch {
			return "", fmt.Errorf("failed to clone branch %q: %w", opts.Branch, err)
		}
		if err := os.RemoveAll(dir); err != nil {
			return "", err
		}
		if err := gitutil.Clone(ctx, dir, tokenURL, opts.BaseBranch, true, true); err != nil {
			return "", fmt.Errorf("failed to clone base branch %q: %w", opts.BaseBranch, err)
		}
		if err := gitutil.Checkout(dir, opts.Branch, true, true, ""); err != nil {
			return "", fmt.Errorf("failed to create branch %q: %w", opts.Branch, err)
		}
	}

	// Never leave the credential-embedded URL in .git/config; subsequent ops pass the token URL explicitly.
	if _, err := gitutil.Run(ctx, dir, "remote", "set-url", "origin", opts.Config.Remote); err != nil {
		return "", err
	}

	if err := applyChanges(dir, opts.Subpath, opts.Changes); err != nil {
		return "", err
	}

	sha, err := gitutil.CommitAll(ctx, dir, "", opts.Message, gitCommitSignature)
	if err != nil {
		return "", err
	}

	if err := gitutil.Push(ctx, dir, tokenURL, opts.Branch); err != nil {
		return "", fmt.Errorf("failed to push branch %q: %w", opts.Branch, err)
	}

	return sha, nil
}

// ReadFiles shallow-clones branch and returns the content of the requested repo-relative paths that exist.
// Paths absent on the branch are simply omitted from the result. It is used to detect which staged files
// are already present on the served branch so they can be marked synced without an extra commit.
func (w *GitWriter) ReadFiles(ctx context.Context, cfg *gitutil.Config, branch, subpath string, paths []string) (map[string][]byte, error) {
	if cfg == nil || cfg.Remote == "" {
		return nil, errors.New("gitwriter: remote is not configured")
	}
	tokenURL, err := cfg.FullyQualifiedRemote()
	if err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp("", "rill-userfiles-read-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	if err := gitutil.Clone(ctx, dir, tokenURL, branch, true, true); err != nil {
		return nil, fmt.Errorf("failed to clone branch %q: %w", branch, err)
	}

	base := dir
	if subpath != "" {
		base = filepath.Join(dir, filepath.FromSlash(subpath))
	}

	out := make(map[string][]byte, len(paths))
	for _, p := range paths {
		b, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(p)))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		out[p] = b
	}
	return out, nil
}

// ReadUserFileFromGit returns the content of a repo-relative file on the project's primary branch.
// It uses the GitHub Contents API, avoiding a clone for single-file reads. It's used to get a fresh
// edit base for user files whose staged row is already synced (the Git copy is authoritative and may
// have been edited directly in Git since the sync).
// If the file does not exist on the branch (e.g. it was deleted directly in Git), the returned error
// wraps os.ErrNotExist so callers can distinguish a confirmed deletion from a transient read failure.
func (s *Service) ReadUserFileFromGit(ctx context.Context, proj *database.Project, filePath string) ([]byte, error) {
	if proj.GitRemote == nil || proj.GithubInstallationID == nil {
		return nil, fmt.Errorf("project %q is not connected to github", proj.Name)
	}
	account, repo, ok := gitutil.SplitGithubRemote(*proj.GitRemote)
	if !ok {
		return nil, fmt.Errorf("invalid github url %q stored for project", *proj.GitRemote)
	}
	repoID, err := s.githubRepoID(ctx, proj)
	if err != nil {
		return nil, err
	}
	client := s.Github.InstallationClient(*proj.GithubInstallationID, &repoID)

	if proj.Subpath != "" {
		filePath = path.Join(proj.Subpath, filePath)
	}
	fc, _, resp, err := client.Repositories.GetContents(ctx, account, repo, filePath, &github.RepositoryContentGetOptions{Ref: proj.PrimaryBranch})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%q does not exist on branch %q: %w", filePath, proj.PrimaryBranch, os.ErrNotExist)
		}
		return nil, fmt.Errorf("failed to read %q from github: %w", filePath, err)
	}
	if fc == nil {
		return nil, fmt.Errorf("path %q on github is not a file", filePath)
	}
	content, err := fc.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode %q from github: %w", filePath, err)
	}
	return []byte(content), nil
}

// applyChanges writes and deletes files in the working tree under <dir>/<subpath>.
func applyChanges(dir, subpath string, changes []FileChange) error {
	base := dir
	if subpath != "" {
		base = filepath.Join(dir, filepath.FromSlash(subpath))
	}
	for _, c := range changes {
		// Guard against path traversal in stored paths.
		clean := filepath.Clean(filepath.FromSlash(c.Path))
		if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.IsAbs(clean) {
			return fmt.Errorf("gitwriter: invalid file path %q", c.Path)
		}
		full := filepath.Join(base, clean)
		if c.Data == nil {
			if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, c.Data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// IsProtectedBranchErr reports whether err looks like a GitHub protected-branch push rejection.
func IsProtectedBranchErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "protected branch") ||
		strings.Contains(msg, "gh006") ||
		strings.Contains(msg, "required status check") ||
		(strings.Contains(msg, "push declined") && strings.Contains(msg, "branch"))
}

// GitConfigForProject builds a gitutil.Config with a short-lived installation token for the project's repo.
func (s *Service) GitConfigForProject(ctx context.Context, proj *database.Project) (*gitutil.Config, error) {
	if proj.GitRemote == nil || proj.GithubInstallationID == nil {
		return nil, fmt.Errorf("project %q is not connected to github", proj.Name)
	}

	repoID, err := s.githubRepoID(ctx, proj)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := s.Github.InstallationToken(ctx, *proj.GithubInstallationID, repoID)
	if err != nil {
		return nil, err
	}

	cfg := &gitutil.Config{
		Remote:            *proj.GitRemote,
		PasswordExpiresAt: expiresAt,
		DefaultBranch:     proj.PrimaryBranch,
		Subpath:           proj.Subpath,
		ManagedRepo:       proj.ManagedGitRepoID != nil,
	}
	// An empty token means credential-less access (e.g. local file:// remotes in tests).
	if token != "" {
		cfg.Username = "x-access-token"
		cfg.Password = token
	}
	return cfg, nil
}

// githubRepoID returns the numeric GitHub repo ID for a project, resolving it via the GitHub API if not stored.
func (s *Service) githubRepoID(ctx context.Context, proj *database.Project) (int64, error) {
	if proj.GithubRepoID != nil {
		return *proj.GithubRepoID, nil
	}
	if proj.GithubInstallationID == nil || proj.GitRemote == nil {
		return 0, fmt.Errorf("project %q is not connected to github", proj.Name)
	}

	client := s.Github.InstallationClient(*proj.GithubInstallationID, nil)
	account, repo, ok := gitutil.SplitGithubRemote(*proj.GitRemote)
	if !ok {
		return 0, fmt.Errorf("invalid github url %q stored for project", *proj.GitRemote)
	}
	ghRepo, _, err := client.Repositories.Get(ctx, account, repo)
	if err != nil {
		return 0, fmt.Errorf("failed to get github repo: %w", err)
	}
	return ghRepo.GetID(), nil
}
