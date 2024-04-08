package admin

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/admin/client"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const (
	pullTimeout         = 10 * time.Minute
	pullRetryN          = 3
	pullRetryWait       = 500 * time.Millisecond
	pullVirtualPageSize = 100
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/drivers/admin")

var spec = drivers.Spec{
	DisplayName: "Rill Admin",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "access_token",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
	},
}

func init() {
	drivers.Register("admin", driver{})
}

type driver struct{}

var _ drivers.Driver = driver{}

type configProperties struct {
	AdminURL    string `mapstructure:"admin_url"`
	AccessToken string `mapstructure:"access_token"`
	ProjectID   string `mapstructure:"project_id"`
	Branch      string `mapstructure:"branch"`
	DataDir     string `mapstructure:"data_dir"`
}

func (d driver) Open(cfgMap map[string]any, shared bool, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("admin driver can't be shared")
	}

	cfg := &configProperties{}
	err := mapstructure.WeakDecode(cfgMap, cfg)
	if err != nil {
		return nil, err
	}

	admin, err := client.New(cfg.AdminURL, cfg.AccessToken, "rill-runtime")
	if err != nil {
		return nil, fmt.Errorf("failed to open admin client: %w", err)
	}

	h := &Handle{
		config:       cfg,
		logger:       logger,
		admin:        admin,
		singleflight: &singleflight.Group{},
	}

	return h, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type Handle struct {
	config               *configProperties
	logger               *zap.Logger
	admin                *client.Client
	singleflight         *singleflight.Group
	cloned               atomic.Bool
	repoPath             string
	projPath             string
	gitURL               string
	gitURLExpiresOn      time.Time
	virtualNextPageToken string
}

var _ drivers.Handle = &Handle{}

// Driver implements drivers.Handle.
func (h *Handle) Driver() string {
	return "admin"
}

// Config implements drivers.Handle.
func (h *Handle) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(h.config, &m)
	return m
}

// Migrate implements drivers.Handle.
func (h *Handle) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (h *Handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Close implements drivers.Handle.
func (h *Handle) Close() error {
	if h.repoPath != "" {
		_ = os.RemoveAll(h.repoPath)
	}
	return nil
}

// AsRegistry implements drivers.Handle.
func (h *Handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (h *Handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (h *Handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return h, true
}

// AsAdmin implements drivers.Handle.
func (h *Handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return h, true
}

// AsAI implements drivers.Handle.
func (h *Handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return h, true
}

// AsOLAP implements drivers.Handle.
func (h *Handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (h *Handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (h *Handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsTransporter implements drivers.Handle.
func (h *Handle) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Handle.
func (h *Handle) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (h *Handle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// cloneOrPull clones or pulls the repo with an exponential backoff retry on retryable errors.
// If onlyClone is false, it's a cheap operation on anything but the first call.
// After it returns successfully, h.repoPath is safe to access.
// It's safe for concurrent calls, which are deduplicated.
func (h *Handle) cloneOrPull(ctx context.Context, onlyClone bool) error {
	if onlyClone && h.cloned.Load() {
		return nil
	}

	ctx, span := tracer.Start(ctx, "cloneOrPull", trace.WithAttributes(attribute.Bool("onlyClone", onlyClone)))
	defer span.End()

	ch := h.singleflight.DoChan("cloneOrPull", func() (interface{}, error) {
		if onlyClone && h.cloned.Load() {
			return nil, nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
		defer cancel()

		r := retrier.New(retrier.ExponentialBackoff(pullRetryN, pullRetryWait), retryErrClassifier{})
		err := r.Run(func() error { return h.cloneOrPullUnsafe(ctx) })
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
func (h *Handle) cloneOrPullUnsafe(ctx context.Context) error {
	err := h.checkHandshake(ctx)
	if err != nil {
		return fmt.Errorf("repo handshake failed: %w", err)
	}

	if !h.cloned.Load() {
		err := h.cloneUnsafeGit()
		if err == nil {
			err = h.pullUnsafeVirtual(ctx)
		}
		h.cloned.Store(err == nil)
		return err
	}

	err = h.pullUnsafeGit()
	if err == nil {
		err = h.pullUnsafeVirtual(ctx)
	}
	return err
}

// checkHandshake checks and possibly renews the repo details handshake with the admin server.
func (h *Handle) checkHandshake(ctx context.Context) error {
	if h.gitURLExpiresOn.After(time.Now()) {
		return nil
	}

	meta, err := h.admin.GetRepoMeta(ctx, &adminv1.GetRepoMetaRequest{
		ProjectId: h.config.ProjectID,
		Branch:    h.config.Branch,
	})
	if err != nil {
		return err
	}

	if h.repoPath == "" {
		if err := os.Mkdir(h.config.DataDir, os.ModePerm); err != nil && !errors.Is(err, fs.ErrExist) {
			return err
		}
		h.repoPath, err = os.MkdirTemp(h.config.DataDir, "admin_driver_repo")
		if err != nil {
			return err
		}

		h.repoPath, err = filepath.Abs(h.repoPath)
		if err != nil {
			return err
		}
	}

	if meta.GitSubpath == "" {
		h.projPath = h.repoPath
	} else {
		h.projPath = filepath.Join(h.repoPath, meta.GitSubpath)
	}

	h.gitURL = meta.GitUrl
	if meta.GitUrlExpiresOn != nil {
		h.gitURLExpiresOn = meta.GitUrlExpiresOn.AsTime()
	} else {
		// Should never happen, unless there is no connected Github repo, which is not allowed today.
		h.gitURLExpiresOn = time.Now().Add(time.Hour)
	}

	return nil
}

// cloneUnsafe clones the Git repository. It removes any existing repository at the repoPath (in case a previous clone failed in a dirty state).
// Unsafe for concurrent use.
func (h *Handle) cloneUnsafeGit() error {
	_, err := os.Stat(h.repoPath)
	if err == nil {
		_ = os.RemoveAll(h.repoPath)
	}

	_, err = git.PlainClone(h.repoPath, false, &git.CloneOptions{
		URL:           h.gitURL,
		ReferenceName: plumbing.NewBranchReferenceName(h.config.Branch),
		SingleBranch:  true,
	})
	return err
}

// pullUnsafeGit pulls changes from the Git repo. It must run after a successful call to cloneUnsafeGit.
// Unsafe for concurrent use.
func (h *Handle) pullUnsafeGit() error {
	repo, err := git.PlainOpen(h.repoPath)
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = wt.Pull(&git.PullOptions{
		RemoteURL:     h.gitURL,
		ReferenceName: plumbing.NewBranchReferenceName(h.config.Branch),
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

// pullUnsafeVirtual syncs changes from the admin server's virtual repository.
// It places files from the virtual repo in a sub-directory __virtual__ of the Git repository.
// It must run after a successful call to cloneUnsafeGit (which creates the directory).
// Unsafe for concurrent use.
func (h *Handle) pullUnsafeVirtual(ctx context.Context) error {
	dst := filepath.Join(h.projPath, "__virtual__")

	i := 0
	n := 500
	for i = 0; i < n; i++ { // Just a failsafe to avoid infinite loops
		res, err := h.admin.PullVirtualRepo(ctx, &adminv1.PullVirtualRepoRequest{
			ProjectId: h.config.ProjectID,
			Branch:    h.config.Branch,
			PageSize:  pullVirtualPageSize,
			PageToken: h.virtualNextPageToken,
		})
		if err != nil {
			return fmt.Errorf("failed to sync virtual repo: %w", err)
		}

		for _, vf := range res.Files {
			path := filepath.Join(dst, vf.Path)

			if vf.Deleted {
				err = os.Remove(path)
				if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove virtual file %q: %w", path, err)
				}
				continue
			}

			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return fmt.Errorf("could not create directory for virtual file %q: %w", path, err)
			}

			err = os.WriteFile(path, vf.Data, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to write virtual file %q: %w", path, err)
			}
		}

		h.virtualNextPageToken = res.NextPageToken

		// If there are no more files, we're done for now.
		// We can't just check NextPageToken because it will still be set, enabling us to pull new changes next time pullUnsafeVirtual is called.
		if len(res.Files) == 0 {
			break
		}
	}

	if i == n {
		return fmt.Errorf("internal: pullUnsafeVirtual ran for over %d iterations", n)
	}

	return nil
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
