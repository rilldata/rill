package cmdutil

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/rilldata/rill/cli/pkg/version"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	defaultAdminURL = "https://admin.rilldata.com"

	telemetryServiceName    = "cli"
	telemetryIntakeURL      = "https://intake.rilldata.io/events/data-modeler-metrics"
	telemetryIntakeUser     = "data-modeler"
	telemetryIntakePassword = "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug==" // nolint:gosec // secret is safe for public use
)

var ErrNoMatchingProject = fmt.Errorf("no matching project found")

type Helper struct {
	*printer.Printer
	Version            version.Version
	DotRill            dotrill.DotRill
	Interactive        bool
	Org                string
	AdminURLDefault    string
	AdminURLOverride   string
	AdminTokenDefault  string
	AdminTokenOverride string

	adminClient        *client.Client
	adminClientHash    string
	activityClient     *activity.Client
	activityClientHash string

	gitHelper   *GitHelper
	gitHelperMu sync.Mutex
}

func NewHelper(ver version.Version, homeDir string) (*Helper, error) {
	// Create it
	ch := &Helper{
		Printer:     printer.NewPrinter(printer.FormatHuman),
		DotRill:     dotrill.New(homeDir),
		Version:     ver,
		Interactive: true,
	}

	// Load base admin config from ~/.rill
	err := ch.ReloadAdminConfig()
	if err != nil {
		return nil, err
	}

	// Load default org
	defaultOrg, err := ch.DotRill.GetDefaultOrg()
	if err != nil {
		return nil, fmt.Errorf("could not parse default org from ~/.rill: %w", err)
	}
	ch.Org = defaultOrg

	return ch, nil
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

func (h *Helper) SetOrg(org string) error {
	if h.Org == org {
		return nil
	}
	h.Org = org
	err := h.DotRill.SetDefaultOrg(org)
	if err != nil {
		return fmt.Errorf("failed to set default org: %w", err)
	}

	h.gitHelperMu.Lock()
	defer h.gitHelperMu.Unlock()
	h.gitHelper = nil // Invalidate the git helper since the org has changed.
	return nil
}

func (h *Helper) IsDev() bool {
	return h.Version.IsDev()
}

func (h *Helper) IsAuthenticated() bool {
	return h.AdminToken() != ""
}

// ReloadAdminConfig populates the helper's AdminURL, AdminTokenDefault, and Org properties from ~/.rill.
func (h *Helper) ReloadAdminConfig() error {
	adminToken, err := h.DotRill.GetAccessToken()
	if err != nil {
		return fmt.Errorf("could not parse access token from ~/.rill: %w", err)
	}

	adminURL, err := h.DotRill.GetDefaultAdminURL()
	if err != nil {
		return fmt.Errorf("could not parse default api URL from ~/.rill: %w", err)
	}
	if adminURL == "" {
		adminURL = defaultAdminURL
	}

	h.AdminURLDefault = adminURL
	h.AdminTokenDefault = adminToken

	return nil
}

func (h *Helper) AdminToken() string {
	if h.AdminTokenOverride != "" {
		return h.AdminTokenOverride
	}
	return h.AdminTokenDefault
}

func (h *Helper) AdminURL() string {
	if h.AdminURLOverride != "" {
		return h.AdminURLOverride
	}
	return h.AdminURLDefault
}

func (h *Helper) Client() (*client.Client, error) {
	// The admin token and URL may have changed (e.g. if the user did a separate login or env switch).
	// Reload the admin config from disk to get the latest values.
	err := h.ReloadAdminConfig()
	if err != nil {
		return nil, err
	}

	// Compute and cache a hash of the admin config values to detect changes.
	// If the hash has changed, we should close the existing client.
	hash := hashStr(h.AdminToken(), h.AdminURL())
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
		c, err := client.New(h.AdminURL(), h.AdminToken(), userAgent)
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
	hash := hashStr(h.AdminToken(), h.AdminURL())

	// Return the client if it's already created and the hash hasn't changed.
	if h.activityClient != nil && h.activityClientHash == hash {
		return h.activityClient
	}

	// Now we can update the hash. The user ID will be refetched below.
	h.activityClientHash = hash

	// Load telemetry config
	installID, analyticsEnabled, err := h.DotRill.AnalyticsInfo()
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

		// Wrap the intake sink in a filter sink that omits events we don't want to send from local.
		// (Remember, this telemetry client will only be used on local.)
		sink := activity.NewFilterSink(intakeSink, func(e activity.Event) bool {
			// Omit metrics events (since they are quite chatty and potentially sensitive).
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

// CurrentUserID fetches the ID of the current user.
// It caches the result in ~/.rill, along with a hash of the current admin token for cache invalidation in case of login/logout.
func (h *Helper) CurrentUserID(ctx context.Context) (string, error) {
	if h.AdminToken() == "" {
		return "", nil
	}

	newHash := hashStr(h.AdminToken(), h.AdminURL())

	oldHash, err := h.DotRill.GetUserCheckHash()
	if err != nil {
		return "", err
	}

	if oldHash == newHash {
		userID, err := h.DotRill.GetUserID()
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

	err = h.DotRill.SetUserID(userID)
	if err != nil {
		return "", err
	}

	err = h.DotRill.SetUserCheckHash(newHash)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (h *Helper) ProjectNamesByGitRemote(ctx context.Context, org, remote, subPath string) ([]string, error) {
	if org == "" || remote == "" {
		return nil, errors.New("org, remote cannot be blank")
	}

	c, err := h.Client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{
		Org: org,
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, p := range resp.Projects {
		if strings.EqualFold(p.GitRemote, remote) && (subPath == "" || strings.EqualFold(p.Subpath, subPath)) {
			names = append(names, p.Name)
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no project with Git remote %q exists in the org %q", remote, org)
	}

	return names, nil
}

// InferProjectName infers the project name from the given path.
// If multiple projects are found, it prompts the user to select one.
func (h *Helper) InferProjectName(ctx context.Context, org, pathToProject string) (string, error) {
	projects, err := h.InferProjects(ctx, org, pathToProject)
	if err != nil {
		return "", err
	}
	if len(projects) == 1 {
		return projects[0].Name, nil
	}

	var names []string
	for _, p := range projects {
		names = append(names, p.Name)
	}
	return SelectPrompt("Select project", names, "")
}

func (h *Helper) InferProjects(ctx context.Context, org, path string) ([]*adminv1.Project, error) {
	path, err := fileutil.ExpandHome(path)
	if err != nil {
		return nil, err
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Build request
	req := &adminv1.ListProjectsForFingerprintRequest{
		DirectoryName: filepath.Base(path),
	}

	// extract subpath
	repoRoot, subpath, err := gitutil.InferRepoRootAndSubpath(path)
	if err == nil {
		req.SubPath = subpath
	}

	// extract remotes
	remote, err := gitutil.ExtractRemotes(repoRoot, false)
	if err == nil {
		for _, r := range remote {
			if r.Name == "__rill_remote" {
				req.RillMgdGitRemote = r.URL
			} else {
				gitRemote, err := r.Github()
				if err == nil {
					req.GitRemote = gitRemote
				}
			}
		}
	}
	c, err := h.Client()
	if err != nil {
		return nil, err
	}
	resp, err := c.ListProjectsForFingerprint(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Projects) == 0 {
		return nil, ErrNoMatchingProject
	}

	if org == "" {
		return resp.Projects, nil
	}

	orgFiltered := make([]*adminv1.Project, 0)
	for _, p := range resp.Projects {
		if p.OrgName == org {
			orgFiltered = append(orgFiltered, p)
		}
	}
	if len(orgFiltered) == 0 {
		return nil, ErrNoMatchingProject
	}
	// cleanup rill managed remote
	if len(orgFiltered) == 1 && orgFiltered[0].ManagedGitId == "" && req.RillMgdGitRemote != "" {
		err := h.HandleRepoTransfer(repoRoot, req.GitRemote)
		if err != nil {
			return nil, err
		}
	}
	return orgFiltered, nil
}

// OpenRuntimeClient opens a client for the production deployment for the given project.
// If local is true, it connects to the locally running runtime instead of the deployed project's runtime.
// It returns the runtime client and instance ID for the project.
func (h *Helper) OpenRuntimeClient(ctx context.Context, org, project string, local bool) (*runtimeclient.Client, string, error) {
	var host, instanceID, jwt string
	if local {
		// This is the default port that Rill localhost uses for gRPC.
		// TODO: In the future, we should capture the gRPC port in ~/.rill and use it here.
		host = "http://localhost:49009"
		instanceID = "default"
	} else {
		adm, err := h.Client()
		if err != nil {
			return nil, "", err
		}

		proj, err := adm.GetProject(ctx, &adminv1.GetProjectRequest{
			Org:     org,
			Project: project,
		})
		if err != nil {
			return nil, "", err
		}

		depl := proj.ProdDeployment
		if depl == nil {
			return nil, "", fmt.Errorf("project %q is not currently deployed", project)
		}
		if depl.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING {
			return nil, "", fmt.Errorf("deployment status not RUNNING: %s", depl.Status.String())
		}

		host = depl.RuntimeHost
		instanceID = depl.RuntimeInstanceId
		jwt = proj.Jwt
	}

	rt, err := runtimeclient.New(host, jwt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect to runtime: %w", err)
	}

	return rt, instanceID, nil
}

func (h *Helper) GitHelper(org, project, localPath string) *GitHelper {
	h.gitHelperMu.Lock()
	defer h.gitHelperMu.Unlock()

	// If the git helper is nil or the org, project or local path has changed, create a new one.
	if h.gitHelper == nil || h.gitHelper.org != org || h.gitHelper.project != project || h.gitHelper.localPath != localPath {
		h.gitHelper = newGitHelper(h, h.Org, project, localPath)
	}
	return h.gitHelper
}

func (h *Helper) GitSignature(ctx context.Context, path string) (*object.Signature, error) {
	repo, err := git.PlainOpen(path)
	if err == nil {
		cfg, err := repo.ConfigScoped(config.SystemScope)
		if err == nil && cfg.User.Email != "" && cfg.User.Name != "" {
			// user has git properly configured use that
			return &object.Signature{
				Name:  cfg.User.Name,
				Email: cfg.User.Email,
				When:  time.Now(),
			}, nil
		}
	}

	// use email of rill user
	c, err := h.Client()
	if err != nil {
		return nil, err
	}
	userResp, err := c.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		if strings.Contains(err.Error(), "not authenticated as a user") {
			return &object.Signature{
				Name:  "service-account",
				Email: "service-account@rilldata.com", // not an actual email
				When:  time.Now(),
			}, nil
		}
		return nil, err
	}

	return &object.Signature{
		Name:  userResp.User.DisplayName,
		Email: userResp.User.Email,
		When:  time.Now(),
	}, nil
}

func (h *Helper) HandleRepoTransfer(path, remote string) error {
	// clear cache
	h.gitHelperMu.Lock()
	h.gitHelper = nil
	h.gitHelperMu.Unlock()

	// remove rill managed remote
	err := removeRemote(path, "__rill_remote")
	if err != nil {
		return err
	}

	// set origin to remote
	err = gitutil.SetRemote(path, &gitutil.Config{
		Remote: remote,
	})
	if err != nil {
		return err
	}

	return nil
}

// CommitAndSafePush commits changes and safely pushes them to the remote repository.
// It fetches the latest remote changes, checks for conflicts, and handles them based on defaultPushChoice:
//   - "1": Pull remote changes and merge (fails on conflicts)
//   - "2": Overwrite remote changes with local changes using merge with favourLocal=true (not supported for monorepos)
//   - "3": Abort the push operation
//
// If h.Interactive is true and there are remote commits, the user will be prompted to choose how to proceed.
func (h *Helper) CommitAndSafePush(ctx context.Context, root string, config *gitutil.Config, commitMsg string, author *object.Signature, defaultPushChoice string) error {
	// 1. Fetch latest from remote
	err := gitutil.GitFetch(ctx, root, config)
	if err != nil {
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// 2. Check status of the subpath
	status, err := gitutil.RunGitStatus(root, config.Subpath, config.RemoteName())
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}
	if status.Branch != config.DefaultBranch {
		return fmt.Errorf("current branch %q does not match expected branch %q", status.Branch, config.DefaultBranch)
	}

	// 3. Warn if there are remote commits
	choice := defaultPushChoice
	if status.RemoteCommits != 0 {
		if h.Interactive {
			h.PrintfWarn("Warning: There are changes on the remote branch that are not in your local branch.")
			h.PrintfWarn("It's recommended to pull the latest changes before pushing to avoid overwriting remote changes.\n")
			h.PrintfWarn("Please choose one of the following options to proceed:\n")
			h.PrintfWarn("1: Pull remote changes to your local branch and fail on conflicts\n")
			h.PrintfWarn("2: Overwrite remote changes with your local changes(Not supported for monorepos)\n")
			h.PrintfWarn("3: Abort deploy and merge manually\n")
			choice, err = SelectPrompt("Choose how to resolve remote changes", []string{"1", "2", "3"}, "1")
			if err != nil {
				return err
			}
		}
	}

	// 4. Merge + push
	// The push can still fail if there were new remote commits since the fetch. But that's okay, the user can just retry.
	switch choice {
	case "1":
		err := gitutil.RunUpstreamMerge(ctx, config.RemoteName(), root, status.Branch, false)
		if err != nil {
			return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
		}
		return gitutil.CommitAndPush(ctx, root, config, commitMsg, author)
	case "2":
		// Instead of a force push, we do a merge with favourLocal=true to ensure we don't loose history.
		// This is not euivalent to a force push but is safer for users.
		if config.Subpath != "" {
			// force pushing in a monorepo can overwrite other subpaths
			// we can check for changes in other subpaths but it is tricky and error prone
			// monorepo setups are advanced use cases and we can require users to manually resolve remote changes
			return fmt.Errorf("cannot overwrite remote changes in a monorepo setup. Merge remote changes manually")
		}
		err := gitutil.RunUpstreamMerge(ctx, config.RemoteName(), root, status.Branch, true)
		if err != nil {
			return fmt.Errorf("local is behind remote and failed to sync with remote: %w", err)
		}
		return gitutil.CommitAndPush(ctx, root, config, commitMsg, author)
	default:
		return fmt.Errorf("aborting deploy")
	}
}

func removeRemote(path, remoteName string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	err = repo.DeleteRemote(remoteName)
	if err != nil && !errors.Is(err, git.ErrRemoteNotFound) {
		return err
	}
	return nil
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
