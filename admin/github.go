package admin

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v71/github"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserIsNotCollaborator      = fmt.Errorf("user is not a collaborator for the repository")
	ErrGithubInstallationNotFound = fmt.Errorf("github installation not found")
)

type GithubToken struct {
	AccessToken  string
	Expiry       time.Time
	RefreshToken string
}

// Github exposes the features we require from the Github API.
type Github interface {
	AppClient() *github.Client
	InstallationClient(installationID int64, repoID *int64) *github.Client
	// InstallationToken returns a token for the installation ID limited to the repoID.
	InstallationToken(ctx context.Context, installationID, repoID int64) (token string, expiresAt time.Time, err error)
	InstallationTokenForOrg(ctx context.Context, org string) (token string, expiresAt time.Time, err error)

	CreateManagedRepo(ctx context.Context, repoPrefix string) (*github.Repository, error)
	ManagedOrgInstallationID() (int64, error)
}

// githubClient implements the Github interface.
type githubClient struct {
	appID         int64
	appPrivateKey string
	appClient     *github.Client
	// appTransport is the transport used to create the app client.
	// It can used across multiple installation clients to reuse TCP connections to Github.
	appTransport *ghinstallation.AppsTransport
	managedAcct  string

	// managedOrgInstallationID is usually populated when the client is created.
	// But we do not return an error if there is any error in fetching the installation ID.
	// This is to let admin server start even if there is an issue with Github service.
	managedOrgInstallationID int64
	managedOrgFetchError     error

	cacheMu           sync.Mutex
	installationCache *simplelru.LRU
}

// NewGithub returns a new client for connecting to Github.
func NewGithub(ctx context.Context, appID int64, appPrivateKey, githubManagedAcct string, logger *zap.Logger) (Github, error) {
	atr, err := ghinstallation.NewAppsTransport(retryableHTTPRoundTripper(), appID, []byte(appPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create github app transport: %w", err)
	}
	appClient := github.NewClient(&http.Client{Transport: atr})

	lru, err := simplelru.NewLRU(100, nil)
	if err != nil {
		panic(err)
	}

	g := &githubClient{
		appID:             appID,
		appPrivateKey:     appPrivateKey,
		appClient:         appClient,
		appTransport:      atr,
		installationCache: lru,
		managedAcct:       githubManagedAcct,
	}

	// Set the managed org installation client
	if githubManagedAcct == "" {
		g.managedOrgFetchError = fmt.Errorf("managed Git repositories are not configured for this environment")
		return g, nil
	}
	i, _, err := appClient.Apps.FindOrganizationInstallation(ctx, githubManagedAcct)
	if err != nil {
		logger.Error("failed to get managed org installation ID", zap.Error(err), observability.ZapCtx(ctx))
		g.managedOrgFetchError = err
		return g, nil
	}
	g.managedOrgInstallationID = *i.ID

	return g, nil
}

func (g *githubClient) AppClient() *github.Client {
	return g.appClient
}

func (g *githubClient) InstallationClient(installationID int64, repoID *int64) *github.Client {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()

	// lookup cache
	cacheKey := installationCacheKey(installationID, repoID)
	val, ok := g.installationCache.Get(cacheKey)
	if ok {
		return val.(*github.Client)
	}

	// create transport for the installation from the app transport
	itr := ghinstallation.NewFromAppsTransport(g.appTransport, installationID)
	if repoID != nil {
		// set the repository ID in the transport options
		opts := itr.InstallationTokenOptions
		if opts == nil {
			opts = &github.InstallationTokenOptions{}
		}
		opts.RepositoryIDs = []int64{*repoID}
	}
	// create the installation client
	installationClient := github.NewClient(&http.Client{Transport: itr})

	// add to cache
	g.installationCache.Add(cacheKey, installationClient)
	return installationClient
}

func (g *githubClient) InstallationToken(ctx context.Context, installationID, repoID int64) (string, time.Time, error) {
	client := g.InstallationClient(installationID, &repoID)
	return g.token(ctx, client)
}

func (g *githubClient) InstallationTokenForOrg(ctx context.Context, org string) (string, time.Time, error) {
	installation, _, err := g.appClient.Apps.FindOrganizationInstallation(ctx, org)
	if err != nil {
		return "", time.Time{}, err
	}
	client := g.InstallationClient(*installation.ID, nil)
	return g.token(ctx, client)
}

func (g *githubClient) CreateManagedRepo(ctx context.Context, name string) (*github.Repository, error) {
	repoName := fmt.Sprintf("%s-%v", name, uuid.New().String()[0:8])

	// get managed org client
	id, err := g.ManagedOrgInstallationID()
	if err != nil {
		return nil, fmt.Errorf("failed to get managed org installation ID: %w", err)
	}
	client := g.InstallationClient(id, nil)

	// create the repo
	repo, _, err := client.Repositories.Create(ctx, g.managedAcct, &github.Repository{
		Name:    github.Ptr(repoName),
		Private: github.Ptr(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create managed repo: %w", err)
	}

	// the create repo API does not wait for repo creation to be fully processed on server. Need to verify by making a get call in a loop
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	for {
		select {
		case <-pollCtx.Done():
			return nil, pollCtx.Err()
		case <-time.After(2 * time.Second):
			// Ready to check again.
		}
		_, _, err := client.Repositories.Get(ctx, g.managedAcct, repoName)
		if err == nil {
			break
		}
	}

	return repo, nil
}

func (g *githubClient) ManagedOrgInstallationID() (int64, error) {
	return g.managedOrgInstallationID, g.managedOrgFetchError
}

func (g *githubClient) token(ctx context.Context, client *github.Client) (string, time.Time, error) {
	tr, ok := client.Client().Transport.(*ghinstallation.Transport)
	if !ok {
		return "", time.Time{}, fmt.Errorf("transport is not of type *ghinstallation.Transport")
	}
	t, err := tr.Token(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	_, expiry, err := tr.Expiry()
	if err != nil {
		return "", time.Time{}, err
	}
	return t, expiry, nil
}

func (s *Service) CreateManagedGitRepo(ctx context.Context, org *database.Organization, name, ownerID string) (*github.Repository, error) {
	if org.QuotaProjects >= 0 {
		count, err := s.DB.CountManagedGitRepos(ctx, org.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count managed repos: %w", err)
		}

		quota := quotaManagedRepos(org)
		if count >= quota {
			return nil, fmt.Errorf("managed repo quota exceeded: %d/%d", count, quota)
		}
	}

	repo, err := s.Github.CreateManagedRepo(ctx, fmt.Sprintf("%s-%s", org.Name, name))
	if err != nil {
		return nil, fmt.Errorf("failed to create managed repo: %w", err)
	}
	_, err = s.DB.InsertManagedGitRepo(ctx, &database.InsertManagedGitRepoOptions{
		OrgID:   org.ID,
		Remote:  repo.GetCloneURL(),
		OwnerID: ownerID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert managed repo meta: %w", err)
	}

	return repo, nil
}

// GetGithubInstallation returns a non zero Github installation ID if the Github App is installed on the repository and is not in suspended state.
// The remote should be a HTTPS URL for a github.com repository with the .git suffix.
func (s *Service) GetGithubInstallation(ctx context.Context, remote string) (int64, error) {
	account, repo, ok := gitutil.SplitGithubRemote(remote)
	if !ok {
		return 0, fmt.Errorf("invalid Github remote %q", remote)
	}

	installation, resp, err := s.Github.AppClient().Apps.FindRepositoryInstallation(ctx, account, repo)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			// We don't have an installation on the repo
			return 0, ErrGithubInstallationNotFound
		}
		return 0, fmt.Errorf("failed to lookup repo info: %w", err)
	}

	if installation.SuspendedAt != nil {
		return 0, ErrGithubInstallationNotFound
	}

	installationID := installation.GetID()
	if installationID == 0 {
		// Do we have to check for this?
		return 0, fmt.Errorf("received invalid installation from Github")
	}

	// The user has access to the installation
	return installationID, nil
}

// LookupGithubRepoForUser returns a Github repository iff the Github App is installed on the repository and user is a collaborator of the project.
// The remote should be a HTTPS URL for a github.com repository with the .git suffix.
func (s *Service) LookupGithubRepoForUser(ctx context.Context, installationID int64, remote, gitUsername string) (*github.Repository, error) {
	account, repo, ok := gitutil.SplitGithubRemote(remote)
	if !ok {
		return nil, fmt.Errorf("invalid Github remote %q", remote)
	}

	if gitUsername == "" {
		return nil, fmt.Errorf("invalid gitUsername %q", gitUsername)
	}

	gh := s.Github.InstallationClient(installationID, nil)

	isColab, resp, err := gh.Repositories.IsCollaborator(ctx, account, repo, gitUsername)
	if err != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, ErrUserIsNotCollaborator
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !isColab {
		return nil, ErrUserIsNotCollaborator
	}

	repository, _, err := gh.Repositories.Get(ctx, account, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get github repository: %w", err)
	}

	return repository, nil
}

// ProcessGithubEvent processes a Github event (usually received over webhooks).
// After validating that the event is a valid Github event, it moves further processing to the background and returns a nil error.
func (s *Service) ProcessGithubEvent(ctx context.Context, rawEvent any) error {
	switch event := rawEvent.(type) {
	// Triggered on push to repository
	case *github.PushEvent:
		return s.processGithubPush(ctx, event)
	// Triggered during first installation of app to an account (org or user) or one or more repos
	case *github.InstallationEvent:
		return s.processGithubInstallationEvent(ctx, event)
	// Triggered when new repos are added to the account (org or user), and the installation has full access to account
	case *github.InstallationRepositoriesEvent:
		return s.processGithubInstallationRepositoriesEvent(ctx, event)
	default:
		return nil
	}
}

func (s *Service) processGithubPush(ctx context.Context, event *github.PushEvent) error {
	// Find Rill project matching the repo that was pushed to
	repo := event.GetRepo()
	projects, err := s.DB.FindProjectsByGitRemote(ctx, *repo.CloneURL)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// App is installed on repo not currently deployed. Do nothing.
			return nil
		}
		return err
	}

	// Parse the branch that was pushed to
	// The format is refs/heads/main or refs/tags/v3.14.1
	ref := event.GetRef()
	_, branch, found := strings.Cut(ref, "refs/heads/")
	if !found {
		// We ignore tag pushes
		return nil
	}

	// Iterate over all projects and trigger reconcile
	for _, project := range projects {
		if branch != project.ProdBranch {
			// Ignore if push was not to production branch
			continue
		}

		// Trigger reconcile (runs in the background)
		if project.ProdDeploymentID != nil {
			depl, err := s.DB.FindDeployment(ctx, *project.ProdDeploymentID)
			if err != nil {
				s.Logger.Error("process github event: could not find deployment", zap.String("project_id", project.ID), zap.Error(err), observability.ZapCtx(ctx))
				continue
			}

			err = s.TriggerParser(ctx, depl)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) processGithubInstallationEvent(_ context.Context, event *github.InstallationEvent) error {
	switch event.GetAction() {
	case "created", "unsuspend", "new_permissions_accepted":
		// TODO: Should we do anything for unsuspend?
	case "suspend", "deleted":
		// no handling as of now
		// previously we were deleting the projects
		// but that means if there is an accidental suspend we delete all projects
	}
	return nil
}

func (s *Service) processGithubInstallationRepositoriesEvent(_ context.Context, event *github.InstallationRepositoriesEvent) error {
	// We can access event.RepositoriesAdded and event.RepositoriesRemoved
	switch event.GetAction() {
	case "added":
		// no handling as of now
	case "removed":
		// no handling as of now
		// previously we were deleting the project for the repo
		// but that means if there is an accidental removal we delete all projects
	}
	return nil
}

func quotaManagedRepos(org *database.Organization) int {
	if org.QuotaProjects >= 0 {
		// allow additional 10 repos for cases where we provision a github repo but it is not used because of errors/user bailed out etc/unused repos were not garbage collected
		return org.QuotaProjects + 10
	}
	return math.MaxInt
}

func retryableHTTPRoundTripper() http.RoundTripper {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 10 * time.Second
	retryClient.Logger = nil // Disable inbuilt logger
	return retryClient.StandardClient().Transport
}

func installationCacheKey(installationID int64, repoID *int64) string {
	if repoID != nil {
		return fmt.Sprintf("%d-%d", installationID, *repoID)
	}
	return fmt.Sprintf("%d", installationID)
}
