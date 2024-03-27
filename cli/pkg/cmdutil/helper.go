package cmdutil

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	telemetryServiceName    = "cli"
	telemetryIntakeURL      = "https://intake.rilldata.io/events/data-modeler-metrics"
	telemetryIntakeUser     = "data-modeler"
	telemetryIntakePassword = "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug==" // nolint:gosec // secret is safe for public use
)

type Helper struct {
	*printer.Printer
	Version            Version
	AdminURL           string
	AdminTokenOverride string
	AdminTokenDefault  string
	Org                string
	Interactive        bool

	adminClient        *client.Client
	adminClientHash    string
	activityClient     *activity.Client
	activityClientHash string
}

func (h *Helper) Close() error {
	grp := errgroup.Group{}

	if h.adminClient != nil {
		grp.Go(h.adminClient.Close)
	}

	if h.activityClient != nil {
		grp.Go(func() error {
			// We'll give ourselves 5s to flush any remaining events.
			// We don't use the cmd context because it might already be cancelled.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// We don't return the error because telemetry errors shouldn't become user-facing errors.
			_ = h.activityClient.Close(ctx)
			return nil
		})
	}

	return grp.Wait()
}

func (h *Helper) IsDev() bool {
	return h.Version.IsDev()
}

func (h *Helper) IsAuthenticated() bool {
	return h.AdminToken() != ""
}

func (h *Helper) AdminToken() string {
	if h.AdminTokenOverride != "" {
		return h.AdminTokenOverride
	}
	return h.AdminTokenDefault
}

func (h *Helper) Client() (*client.Client, error) {
	// We allow the admin token and URL to be changed (e.g. during login or env switching).
	// We compute and cache a hash of these values to detect changes.
	// If the hash has changed, we should close the existing client.
	hash := hashStr(h.AdminToken(), h.AdminURL)
	if h.adminClient != nil && h.adminClientHash != hash {
		_ = h.adminClient.Close()
		h.adminClient = nil
		h.adminClientHash = hash
	}
	h.adminClientHash = hash

	// Make a new client if we don't have one.
	if h.adminClient == nil {
		cliVersion := h.Version.Number
		if cliVersion == "" {
			cliVersion = "unknown"
		}

		userAgent := fmt.Sprintf("rill-cli/%v", cliVersion)
		c, err := client.New(h.AdminURL, h.AdminToken(), userAgent)
		if err != nil {
			return nil, err
		}

		h.adminClient = c
	}

	return h.adminClient, nil
}

// Telemetry returns a client for recording events.
// Note: It should only be used for parts of the CLI that run on users' local computer because:
// a) it accesses ~/.rill and adds information about the current user,
// b) it sends events to the public intake endpoint instead of directly to Kafka.
func (h *Helper) Telemetry(ctx context.Context) *activity.Client {
	// If the admin token or URL changes, the user ID of the telemetry client may have changed.
	// We compute and cache a hash of these values to detect changes.
	// If the hash has changed, we refetch the current user and update the client.
	hash := hashStr(h.AdminToken(), h.AdminURL)

	// Return the client if it's already created and the hash hasn't changed.
	if h.activityClient != nil && h.activityClientHash == hash {
		return h.activityClient
	}

	// Now we can update the hash. The user ID will be refetched below.
	h.activityClientHash = hash

	// Load telemetry config
	installID, analyticsEnabled, err := dotrill.AnalyticsInfo()
	if err != nil {
		analyticsEnabled = false
	}

	// Create a client if there isn't one
	if h.activityClient == nil {
		// If analytics are disabled, we'll use a no-op client.
		// We can set it and return early here.
		if !analyticsEnabled {
			h.activityClient = activity.NewNoopClient()
			return h.activityClient
		}

		// Create a sink that sends events to the intake server.
		intakeSink := activity.NewIntakeSink(zap.NewNop(), activity.IntakeSinkOptions{
			IntakeURL:      telemetryIntakeURL,
			IntakeUser:     telemetryIntakeUser,
			IntakePassword: telemetryIntakePassword,
			BufferSize:     50,
			SinkInterval:   time.Second,
		})

		// Wrap the intake sink in a filter sink that omits metrics events (since they are quite chatty and potentially sensitive).
		// (Remember, this telemetry client will only be used on local.)
		sink := activity.NewFilterSink(intakeSink, func(e activity.Event) bool {
			return e.EventType != activity.EventTypeMetric
		})

		// Create the telemetry client with metadata about the current environment.
		h.activityClient = activity.NewClient(sink, zap.NewNop())
		h.activityClient = h.activityClient.WithServiceName(telemetryServiceName)
		if h.Version.Number != "" || h.Version.Commit != "" {
			h.activityClient = h.activityClient.WithServiceVersion(h.Version.Number, h.Version.Commit)
		}
		if h.Version.IsDev() {
			h.activityClient = h.activityClient.WithIsDev()
		}
		h.activityClient = h.activityClient.WithInstallID(installID)
	}

	// Fetch the current user ID and set it on the telemetry client.
	// We do this outside of the client creation block to reset the user ID if the hash changes.
	var userID string
	if h.AdminToken() != "" {
		userID, _ = h.CurrentUserID(ctx)
	}
	h.activityClient = h.activityClient.WithUserID(userID)

	return h.activityClient
}

// CurrentUser fetches the ID of the current user.
// It caches the result in ~/.rill, along with a hash of the current admin token for cache invalidation in case of login/logout.
func (h *Helper) CurrentUserID(ctx context.Context) (string, error) {
	if h.AdminToken() == "" {
		return "", nil
	}

	newHash := hashStr(h.AdminToken(), h.AdminURL)

	oldHash, err := dotrill.GetUserCheckHash()
	if err != nil {
		return "", err
	}

	if oldHash == newHash {
		userID, err := dotrill.GetUserID()
		if err != nil {
			return "", err
		}
		return userID, nil
	}

	c, err := h.Client()
	if err != nil {
		return "", err
	}

	res, err := c.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return "", err
	}

	var userID string
	if res.User != nil {
		userID = res.User.Id
	}

	err = dotrill.SetUserID(userID)
	if err != nil {
		return "", err
	}

	err = dotrill.SetUserCheckHash(newHash)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (h *Helper) ProjectNamesByGithubURL(ctx context.Context, org, githubURL string) ([]string, error) {
	c, err := h.Client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{
		OrganizationName: org,
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, p := range resp.Projects {
		if strings.EqualFold(p.GithubUrl, githubURL) {
			names = append(names, p.Name)
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no project with githubURL %q exist in org %q", githubURL, org)
	}

	return names, nil
}

func (h *Helper) InferProjectName(ctx context.Context, org, path string) (string, error) {
	// Verify projectPath is a Git repo with remote on Github
	_, githubURL, err := gitutil.ExtractGitRemote(path, "", true)
	if err != nil {
		return "", err
	}

	// Fetch project names matching the Github URL
	names, err := h.ProjectNamesByGithubURL(ctx, org, githubURL)
	if err != nil {
		return "", err
	}

	if len(names) == 1 {
		return names[0], nil
	}

	return SelectPrompt("Select project", names, ""), nil
}

func hashStr(ss ...string) string {
	hash := md5.New()
	for _, s := range ss {
		_, err := hash.Write([]byte(s))
		if err != nil {
			panic(err)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
