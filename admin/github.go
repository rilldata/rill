package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v50/github"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserIsNotCollaborator      = fmt.Errorf("user is not a collaborator for the repository")
	ErrGithubInstallationNotFound = fmt.Errorf("github installation not found")
)

// Github exposes the features we require from the Github API.
type Github interface {
	AppClient() *github.Client
	InstallationClient(installationID int64) (*github.Client, error)
}

// githubClient implements the Github interface.
type githubClient struct {
	appID         int64
	appPrivateKey string
	appClient     *github.Client

	cacheMu           sync.Mutex
	installationCache *simplelru.LRU
}

// NewGithub returns a new client for connecting to Github.
func NewGithub(appID int64, appPrivateKey string) (Github, error) {
	atr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, []byte(appPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create github app transport: %w", err)
	}
	appClient := github.NewClient(&http.Client{Transport: atr})

	lru, err := simplelru.NewLRU(100, nil)
	if err != nil {
		panic(err)
	}

	return &githubClient{
		appID:             appID,
		appPrivateKey:     appPrivateKey,
		appClient:         appClient,
		installationCache: lru,
	}, nil
}

func (g *githubClient) AppClient() *github.Client {
	return g.appClient
}

func (g *githubClient) InstallationClient(installationID int64) (*github.Client, error) {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()

	val, ok := g.installationCache.Get(installationID)
	if ok {
		return val.(*github.Client), nil
	}

	itr, err := ghinstallation.New(http.DefaultTransport, g.appID, installationID, []byte(g.appPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create github installation transport: %w", err)
	}
	installationClient := github.NewClient(&http.Client{Transport: itr})

	g.installationCache.Add(installationID, installationClient)
	return installationClient, nil
}

// GetGithubInstallation returns a non zero Github installation ID if the Github App is installed on the repository
// and is not in suspended state
// The githubURL should be a HTTPS URL for a Github repository without the .git suffix.
func (s *Service) GetGithubInstallation(ctx context.Context, githubURL string) (int64, error) {
	account, repo, ok := gitutil.SplitGithubURL(githubURL)
	if !ok {
		return 0, fmt.Errorf("invalid Github URL %q", githubURL)
	}

	installation, resp, err := s.github.AppClient().Apps.FindRepositoryInstallation(ctx, account, repo)
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
// The githubURL should be a HTTPS URL for a Github repository without the .git suffix.
func (s *Service) LookupGithubRepoForUser(ctx context.Context, installationID int64, githubURL, gitUsername string) (*github.Repository, error) {
	account, repo, ok := gitutil.SplitGithubURL(githubURL)
	if !ok {
		return nil, fmt.Errorf("invalid Github URL %q", githubURL)
	}

	if gitUsername == "" {
		return nil, fmt.Errorf("invalid gitUsername %q", gitUsername)
	}

	gh, err := s.github.InstallationClient(installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create github installation client: %w", err)
	}

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
	githubURL := *repo.HTMLURL
	projects, err := s.DB.FindProjectsByGithubURL(ctx, githubURL)
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
				s.logger.Error("process github event: could not find deployment", zap.String("project_id", project.ID), zap.Error(err), observability.ZapCtx(ctx))
				continue
			}

			err = s.TriggerReconcile(ctx, depl)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) processGithubInstallationEvent(ctx context.Context, event *github.InstallationEvent) error {
	switch event.GetAction() {
	case "created", "unsuspend", "new_permissions_accepted":
		// TODO: Should we do anything for unsuspend?
	case "suspend", "deleted":
		// the github installation ID will change if user re-installs the app deleting the project for now
		installation := event.GetInstallation()
		if installation == nil {
			return fmt.Errorf("nil installation")
		}

		if err := s.deleteProjectsForInstallation(ctx, installation.GetID()); err != nil {
			s.logger.Error("failed to delete deployment for installation", zap.Int64("installation_id", installation.GetID()), zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *Service) processGithubInstallationRepositoriesEvent(ctx context.Context, event *github.InstallationRepositoriesEvent) error {
	// We can access event.RepositoriesAdded and event.RepositoriesRemoved
	switch event.GetAction() {
	case "added":
		// no handling as of now
	case "removed":
		var multiErr error
		for _, repo := range event.RepositoriesRemoved {
			if err := s.deleteProjectsForRepo(ctx, repo); err != nil {
				multiErr = multierr.Combine(multiErr, err)
				s.logger.Error("failed to delete project for repo", zap.String("repo", *repo.HTMLURL), zap.Error(err))
			}
		}
		return multiErr
	}
	return nil
}

func (s *Service) deleteProjectsForInstallation(ctx context.Context, id int64) error {
	// Find Rill project for installationID
	projects, err := s.DB.FindProjectsByGithubInstallationID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// App is removed from repo not currently deployed. Do nothing.
			return nil
		}
		return err
	}

	var multiErr error
	for _, p := range projects {
		err := s.TeardownProject(ctx, p)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				multiErr = multierr.Combine(multiErr, err)
				s.logger.Error("unable to delete project", zap.String("project_id", p.ID), zap.Error(err))
			}
			continue
		}
	}
	return multiErr
}

func (s *Service) deleteProjectsForRepo(ctx context.Context, repo *github.Repository) error {
	// Find Rill project matching the repo that was pushed to
	githubURL := *repo.HTMLURL
	projects, err := s.DB.FindProjectsByGithubURL(ctx, githubURL)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// App is removed from repo not currently deployed. Do nothing.
			return nil
		}
		return err
	}

	var multiErr error
	for _, p := range projects {
		err := s.TeardownProject(ctx, p)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				multiErr = multierr.Combine(multiErr, err)
				s.logger.Error("unable to delete project", zap.String("project_id", p.ID), zap.Error(err))
			}
			continue
		}
	}
	return multiErr
}
