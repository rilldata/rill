package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

// ProcessGithubInstallation tracks a confirmed relationship between a user and an installation of the Github App.
func (s *Service) ProcessUserGithubInstallation(ctx context.Context, userID string, installationID int64) error {
	return s.DB.UpsertUserGithubInstallation(ctx, userID, installationID)
}

// GetUserGithubInstallation returns a Github installation ID iff the Github App is installed on the repository AND we have a confirmed relationship between the user and that installation.
// The githubURL should be a HTTPS URL for a Github repository without the .git suffix.
func (s *Service) GetUserGithubInstallation(ctx context.Context, userID, githubURL string) (int64, bool, error) {
	account, repo, ok := splitGithubURL(githubURL)
	if !ok {
		return 0, false, fmt.Errorf("invalid Github URL %q", githubURL)
	}

	installation, resp, err := s.github.Apps.FindRepositoryInstallation(ctx, account, repo)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			// We don't have an installation on the repo
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to lookup repo info: %w", err)
	}

	installationID := installation.GetID()
	if installationID == 0 {
		// Do we have to check for this?
		return 0, false, fmt.Errorf("received invalid installation from Github")
	}

	_, err = s.DB.FindUserGithubInstallation(ctx, userID, installationID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// The user doesn't have access to the installation
			return 0, false, nil
		}
		return 0, false, err // Unexpected error
	}

	// The user has access to the installation
	return installationID, true, nil
}

// LookupGithubRepo calls the Github API using an installation token to get information about a Github repo.
// The githubURL should be a HTTPS URL for a Github repository without the .git suffix.
func (s *Service) LookupGithubRepo(ctx context.Context, installationID int64, githubURL string) (*github.Repository, error) {
	account, repo, ok := splitGithubURL(githubURL)
	if !ok {
		return nil, fmt.Errorf("invalid Github URL %q", githubURL)
	}

	gh, err := s.githubInstallationClient(installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create github installation client: %w", err)
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
	project, err := s.DB.FindProjectByGithubURL(ctx, githubURL)
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

	// Exit if push was not to production branch
	if branch != project.ProductionBranch {
		return nil
	}

	// Trigger reconcile (runs in the background - err means the deployment wasn't found, which is unlikely)
	if project.ProductionDeploymentID != nil {
		err = s.TriggerReconcile(ctx, *project.ProductionDeploymentID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) processGithubInstallationEvent(ctx context.Context, event *github.InstallationEvent) error {
	// We also get event.Repositories if needed
	installationID := *event.GetInstallation().ID

	switch event.GetAction() {
	case "created", "unsuspend", "suspend", "new_permissions_accepted":
		// TODO: Should we do anything for unsuspend?
	case "deleted":
		err := s.DB.DeleteUserGithubInstallations(ctx, installationID)
		if err != nil {
			s.logger.Error("failed to delete github installations", zap.Int64("installation_id", installationID), zap.Error(err), observability.ZapCtx(ctx))
		}
	}

	return nil
}

func (s *Service) processGithubInstallationRepositoriesEvent(ctx context.Context, event *github.InstallationRepositoriesEvent) error {
	// We can access event.RepositoriesAdded and event.RepositoriesRemoved
	return nil
}

// githubInstallationClient makes a Github client that authenticates as a specific installation.
// (As opposed to s.github, which authenticates as the Git App, and cannot access the contents of an installation.)
func (s *Service) githubInstallationClient(installationID int64) (*github.Client, error) {
	itr, err := ghinstallation.New(http.DefaultTransport, s.opts.GithubAppID, installationID, []byte(s.opts.GithubAppPrivateKey))
	if err != nil {
		return nil, err
	}
	return github.NewClient(&http.Client{Transport: itr}), nil
}

func splitGithubURL(githubURL string) (account, repo string, ok bool) {
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
