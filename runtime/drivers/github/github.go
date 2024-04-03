package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	pullTimeout = 10 * time.Minute
	retryN      = 3
	retryWait   = 500 * time.Millisecond
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/drivers/github")

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

func (d driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("github driver can't be shared")
	}
	dsnStr, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open github connection")
	}

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
		config:       config,
		dsn:          dsn,
		tempdir:      tempdir,
		projectdir:   projectDir,
		singleflight: &singleflight.Group{},
		shared:       shared,
	}, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return drivers.Spec{}
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type connection struct {
	config              map[string]any
	dsn                 DSN
	shared              bool
	tempdir             string // tempdir path should be absolute
	projectdir          string
	cloneURLWithToken   string
	cloneURLRefreshedOn time.Time
	singleflight        *singleflight.Group
	cloned              atomic.Bool
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	err := os.RemoveAll(c.tempdir)
	if err != nil {
		return err
	}

	return nil
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return c.config
}

// AsRegistry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	if c.shared {
		return nil, false
	}
	return c, true
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
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

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(*structpb.Struct) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// cloneOrPull clones or pulls the repo with an exponential backoff retry on retryable errors.
// It's safe for concurrent calls, which are deduplicated.
func (c *connection) cloneOrPull(ctx context.Context, onlyClone bool) error {
	if onlyClone && c.cloned.Load() {
		return nil
	}

	ctx, span := tracer.Start(ctx, "cloneOrPull", trace.WithAttributes(attribute.Bool("onlyClone", onlyClone)))
	defer span.End()

	ch := c.singleflight.DoChan("pullOrClone", func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
		defer cancel()

		r := retrier.New(retrier.ExponentialBackoff(retryN, retryWait), retryErrClassifier{})
		err := r.Run(func() error { return c.cloneOrPullUnsafe(ctx) })
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-ch:
		return res.Err
	}
}

// cloneOrPullUnsafe pulls changes from the repo. Also clones the repo if it hasn't been cloned already.
func (c *connection) cloneOrPullUnsafe(ctx context.Context) error {
	if !c.cloned.Load() {
		err := c.cloneUnsafe(ctx)
		c.cloned.Store(err == nil)
		return err
	}

	return c.pullUnsafe(ctx)
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

	err = wt.Pull(&git.PullOptions{
		RemoteURL:     cloneURL,
		ReferenceName: plumbing.NewBranchReferenceName(c.dsn.Branch),
		SingleBranch:  true,
		Force:         true,
	})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	} else if errors.Is(err, git.ErrNonFastForwardUpdate) {
		head, err := repo.Head()
		if err != nil {
			return err
		}

		branch, err := repo.Branch(head.Name().Short())
		if err != nil {
			return err
		}

		rev, err := repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("remotes/%s/%s", branch.Remote, head.Name().Short())))
		if err != nil {
			return err
		}

		err = wt.Reset(&git.ResetOptions{
			Commit: *rev,
			Mode:   git.HardReset,
		})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// cloneUnsafe runs the initial clone of the repo.
func (c *connection) cloneUnsafe(ctx context.Context) error {
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
		return "", fmt.Errorf("failed to create github installation transport: %w", err)
	}
	token, err := itr.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
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

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// retryErrClassifier classifies Github request errors as retryable or not.
type retryErrClassifier struct{}

func (retryErrClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return retrier.Fail
	}

	ghinstallationErr := &ghinstallation.HTTPError{}
	if errors.As(err, &ghinstallationErr) && ghinstallationErr.Response != nil {
		statusCode := ghinstallationErr.Response.StatusCode
		if statusCode/100 == 4 && statusCode != 429 {
			// Any 4xx error apart from 429 is non retryable
			return retrier.Fail
		}
	}

	return retrier.Retry
}
