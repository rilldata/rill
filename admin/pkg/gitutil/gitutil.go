package gitutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitConfig struct {
	Remote   string
	Username string
	Password string
	Branch   string
	Subpath  string
}

var allowedPaths = []string{
	".git",
	"README.md",
	"LICENSE",
}

func SplitGithubURL(githubURL string) (account, repo string, ok bool) {
	githubURL = strings.TrimSuffix(githubURL, ".git")
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

// CopyRepoContents copies the content of the repo to the projPath.
// If subpath is provided, it will copy the content from that subpath within the git root.
func CopyRepoContents(ctx context.Context, projPath string, config *GitConfig) error {
	srcGitPath, err := os.MkdirTemp(os.TempDir(), "src_git_repos")
	if err != nil {
		return err
	}
	defer os.RemoveAll(srcGitPath)

	// srcProjPath is actual path for project including any subpath within the git root
	srcProjPath := srcGitPath
	if config.Subpath != "" {
		srcProjPath = filepath.Join(srcProjPath, config.Subpath)
	}
	err = os.MkdirAll(srcProjPath, fs.ModePerm)
	if err != nil {
		return err
	}

	_, err = git.PlainCloneContext(ctx, srcGitPath, false, &git.CloneOptions{
		URL:           config.Remote,
		Auth:          &githttp.BasicAuth{Username: config.Username, Password: config.Password},
		ReferenceName: plumbing.NewBranchReferenceName(config.Branch),
		SingleBranch:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to clone source git repo: %w", err)
	}

	err = copyDirExclDotGit(srcProjPath, projPath)
	if err != nil {
		return fmt.Errorf("failed to read root files: %w", err)
	}

	return nil
}

// PushContentsToRepo pushes the contents copied by copyData to the git repository at remote.
func PushContentsToRepo(ctx context.Context, copyData func(path string) error, c *GitConfig, force bool, author *object.Signature) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// generate a temp dir to extract the archive
	gitPath, err := os.MkdirTemp(os.TempDir(), "projects")
	if err != nil {
		return err
	}
	defer os.RemoveAll(gitPath)

	// projPath is the target for extracting the archive
	projPath := gitPath
	if c.Subpath != "" {
		projPath = filepath.Join(projPath, c.Subpath)
	}
	err = os.MkdirAll(projPath, fs.ModePerm)
	if err != nil {
		return err
	}

	gitAuth := &githttp.BasicAuth{Username: c.Username, Password: c.Password}

	var ghRepo *git.Repository
	empty := false
	ghRepo, err = git.PlainClone(gitPath, false, &git.CloneOptions{
		URL:           c.Remote,
		Auth:          gitAuth,
		ReferenceName: plumbing.NewBranchReferenceName(c.Branch),
		SingleBranch:  true,
	})
	if err != nil {
		if !errors.Is(err, transport.ErrEmptyRemoteRepository) {
			return fmt.Errorf("failed to init git repo: %w", err)
		}

		empty = true
		ghRepo, err = git.PlainInitWithOptions(gitPath, &git.PlainInitOptions{
			InitOptions: git.InitOptions{
				DefaultBranch: plumbing.NewBranchReferenceName(c.Branch),
			},
			Bare: false,
		})
		if err != nil {
			return fmt.Errorf("failed to init git repo: %w", err)
		}
	}

	wt, err := ghRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	var wtc worktreeContents
	if !empty {
		wtc, err = readWorktree(wt, c.Subpath)
		if err != nil {
			return fmt.Errorf("failed to read worktree: %w", err)
		}

		if len(wtc.otherPaths) > 0 && !force {
			return fmt.Errorf("worktree has additional contents")
		}
	}

	// remove all the other paths
	for _, path := range wtc.otherPaths {
		err = os.RemoveAll(filepath.Join(projPath, path))
		if err != nil {
			return err
		}
	}

	err = copyData(projPath)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// add back the older gitignore contents if present
	if wtc.gitignore != "" {
		gi, err := os.ReadFile(filepath.Join(projPath, ".gitignore"))
		if err != nil {
			return err
		}

		// if the new gitignore is not the same then it was overwritten during extract
		if string(gi) != wtc.gitignore {
			// append the new contents to the end
			gi = append([]byte(fmt.Sprintf("%s\n", wtc.gitignore)), gi...)

			err = os.WriteFile(filepath.Join(projPath, ".gitignore"), gi, fs.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	// git add .
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// git commit -m
	_, err = wt.Commit("Auto committed by Rill", &git.CommitOptions{
		All:    true,
		Author: author,
	})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return fmt.Errorf("failed to commit files to git: %w", err)
		}
	}

	if empty {
		// we need to add a remote as the new repo if the repo was completely empty
		_, err = ghRepo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{c.Remote}})
		if err != nil {
			return fmt.Errorf("failed to create remote: %w", err)
		}
	}

	if err := ghRepo.PushContext(ctx, &git.PushOptions{Auth: gitAuth}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return fmt.Errorf("failed to push to remote %q : %w", c.Remote, err)
		}
	}
	return nil
}

func copyDirExclDotGit(srcDir, destDir string) error {
	_, err := os.Stat(destDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(destDir, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			err = copyDirExclDotGit(srcPath, destPath)
		} else {
			err = copyFile(srcPath, destPath)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(srcFile, destFile string) error {
	src, err := os.Create(destFile)
	if err != nil {
		return err
	}

	defer src.Close()

	dest, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer dest.Close()

	_, err = io.Copy(src, dest)
	if err != nil {
		return err
	}

	return nil
}

type worktreeContents struct {
	gitignore  string
	otherPaths []string
}

func readWorktree(wt *git.Worktree, subpath string) (worktreeContents, error) {
	var wtc worktreeContents

	files, err := wt.Filesystem.ReadDir(subpath)
	if err != nil {
		return worktreeContents{}, err
	}
	for _, file := range files {
		if file.Name() == ".gitignore" {
			f, err := wt.Filesystem.Open(filepath.Join(subpath, file.Name()))
			if err != nil {
				return worktreeContents{}, err
			}
			wtc.gitignore, err = readFile(f)
			if err != nil {
				return worktreeContents{}, err
			}
		} else {
			found := false
			for _, path := range allowedPaths {
				if file.Name() == path {
					found = true
					break
				}
			}
			if !found {
				wtc.otherPaths = append(wtc.otherPaths, file.Name())
			}
		}
	}

	return wtc, nil
}

func readFile(f billy.File) (string, error) {
	defer f.Close()
	buf := make([]byte, 0, 32*1024)
	c := ""

	for {
		n, err := f.Read(buf[:cap(buf)])
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		if n == 0 {
			continue
		}
		buf = buf[:n]
		c += string(buf)
	}

	return c, nil
}
