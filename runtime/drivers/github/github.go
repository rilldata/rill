package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

const (
	retryN    = 3
	retryWait = 500 * time.Millisecond
)

type DSN struct {
	GithubURL      string `json:"github_url"`
	Subpath        string `json:"subpath"`
	Branch         string `json:"branch"`
	InstallationID int64  `json:"installation_id"`
}

func init() {
	drivers.Register("github", driver{})
}

type driver struct{}

func (d driver) Open(dsnStr string, logger *zap.Logger) (drivers.Connection, error) {
	var dsn DSN
	err := json.Unmarshal([]byte(dsnStr), &dsn)
	if err != nil {
		return nil, err
	}

	tempdir, err := os.MkdirTemp("", "github_repo_driver")
	if err != nil {
		return nil, err
	}

	tempdir, err = filepath.Abs(tempdir)
	if err != nil {
		return nil, err
	}

	projectDir := tempdir
	if dsn.Subpath != "" {
		projectDir = filepath.Join(tempdir, dsn.Subpath)
	}

	// NOTE :: project isn't cloned yet
	return &connection{
		dsnStr:     dsnStr,
		dsn:        dsn,
		tempdir:    tempdir,
		projectdir: projectDir,
	}, nil
}

type connection struct {
	dsnStr string
	dsn    DSN
	// tempdir path should be absolute
	tempdir             string
	projectdir          string
	cloneURLWithToken   string
	cloneURLRefreshedOn time.Time
	mu                  sync.Mutex

	// cloned is set to true once github repo has been cloned successfully.
	cloned  bool
	pullErr error
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	err := os.RemoveAll(c.tempdir)
	if err != nil {
		return err
	}

	return nil
}

// Registry implements drivers.Connection.
func (c *connection) RegistryStore() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return c, true
}

// OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// pull pulls changes from the repo. Also clones the repo if it hasn't been cloned already.
func (c *connection) pull(ctx context.Context) error {
	var deduplicated bool
	if !c.mu.TryLock() {
		deduplicated = true
		c.mu.Lock()
	}
	defer c.mu.Unlock()

	if deduplicated {
		return c.pullErr
	}

	if !c.cloned {
		c.pullErr = c.clone(ctx)
		if c.pullErr == nil { // cloned successfully
			c.cloned = true
		}
		return c.pullErr
	}

	c.pullErr = c.pullUnsafe(ctx)
	return c.pullErr
}

// pullUnsafe pulls changes from the repo. Requires repo to be cloned already.
// Unsafe for concurrent use.
func (c *connection) pullUnsafe(ctx context.Context) error {
	repo, err := git.PlainOpen(c.tempdir)
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	cloneURL, err := c.cloneURL(ctx)
	if err != nil {
		return err
	}

	err = wt.Pull(&git.PullOptions{RemoteURL: cloneURL})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

// clone runs the initial clone of the repo.
func (c *connection) clone(ctx context.Context) error {
	cloneURL, err := c.cloneURL(ctx)
	if err != nil {
		return err
	}

	_, err = git.PlainClone(c.tempdir, false, &git.CloneOptions{
		URL:           cloneURL,
		ReferenceName: plumbing.NewBranchReferenceName(c.dsn.Branch),
		SingleBranch:  true,
	})
	return err
}

const cloneURLTTL = 30 * time.Minute

// renewCloneURL retrieves a new clone URL containing an authentication token for the repo.
func (c *connection) cloneURL(ctx context.Context) (string, error) {
	// Return cached token if not expired
	if c.cloneURLWithToken != "" && time.Since(c.cloneURLRefreshedOn) < cloneURLTTL {
		return c.cloneURLWithToken, nil
	}

	// This driver expects Github App credentials to be available from the environment.
	// TODO: The parsing probably should not happen here.
	appID, _ := strconv.ParseInt(os.Getenv("RILL_RUNTIME_GITHUB_APP_ID"), 10, 64)
	if appID == 0 {
		return "", fmt.Errorf("invalid value provided for RILL_RUNTIME_GITHUB_APP_ID")
	}
	privateKey := os.Getenv("RILL_RUNTIME_GITHUB_APP_PRIVATE_KEY")
	if privateKey == "" {
		return "", fmt.Errorf("invalid value provided for RILL_RUNTIME_GITHUB_APP_PRIVATE_KEY")
	}

	// Get a Github token for this installation ID
	itr, err := ghinstallation.New(http.DefaultTransport, appID, c.dsn.InstallationID, []byte(privateKey))
	if err != nil {
		log.Fatal(err)
	}
	token, err := itr.Token(ctx)
	if err != nil {
		return "", err
	}

	// Create clone URL
	ep, err := transport.NewEndpoint(c.dsn.GithubURL + ".git") // TODO: Can the clone URL be different from the HTTP URL of a Github repo?
	if err != nil {
		return "", fmt.Errorf("failed to create endpoint from %q: %w", c.dsn.GithubURL, err)
	}
	ep.User = "x-access-token"
	ep.Password = token
	cloneURL := ep.String()

	// Cache it
	c.cloneURLWithToken = cloneURL
	c.cloneURLRefreshedOn = time.Now()

	// Done
	return cloneURL, nil
}
