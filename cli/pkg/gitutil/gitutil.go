package gitutil

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
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

func ExtractRemotes(projectPath string) ([]Remote, error) {
	repo, err := git.PlainOpen(projectPath)
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

func RemotesToGithubURL(remotes []Remote) (string, error) {
	// Return the first Github URL found.
	// If no Github remotes were found, return the first error.
	var firstErr error
	for _, remote := range remotes {
		ghurl, err := remoteToGithubURL(remote.URL)
		if err == nil {
			// Found a Github remote. Success!
			return ghurl, nil
		}
		if firstErr == nil {
			firstErr = fmt.Errorf("invalid remote %q: %w", remote.URL, err)
		}
	}

	if firstErr == nil {
		return "", ErrGitRemoteNotFound
	}

	return "", firstErr
}

func remoteToGithubURL(remote string) (string, error) {
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

type SyncStatus int

const (
	SyncStatusUnspecified SyncStatus = iota
	SyncStatusModified               // Local branch has untracked/modified changes
	SyncStatusAhead                  // Local branch is ahead of remote branch
	SyncStatusSynced                 // Local branch is in sync with remote branch
)

// GetSyncStatus returns the status of current branch as compared to remote
// TODO: Need to implement cases like local branch is behind/diverged from remote branch
func GetSyncStatus(repoPath, branch string) (SyncStatus, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return SyncStatusUnspecified, err
	}

	ref, err := repo.Head()
	if err != nil {
		return SyncStatusUnspecified, err
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
	// no easy way to get it from go-get library so running git command directly and checking response
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
