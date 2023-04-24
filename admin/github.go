package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserIsNotCollaborator      = fmt.Errorf("user is not a collaborator for the repository")
	ErrGithubInstallationNotFound = fmt.Errorf("github installation not found")
)

// GetGithubInstallation returns a non zero Github installation ID iff the Github App is installed on the repository.
// The githubURL should be a HTTPS URL for a Github repository without the .git suffix.
func (s *Service) GetGithubInstallation(ctx context.Context, githubURL string) (int64, error) {
	account, repo, ok := gitutil.SplitGithubURL(githubURL)
	if !ok {
		return 0, fmt.Errorf("invalid Github URL %q", githubURL)
	}

	// TODO :: handle suspended case
	installation, resp, err := s.github.Apps.FindRepositoryInstallation(ctx, account, repo)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			// We don't have an installation on the repo
			return 0, ErrGithubInstallationNotFound
		}
		return 0, fmt.Errorf("failed to lookup repo info: %w", err)
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

	gh, err := s.githubInstallationClient(installationID)
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

		// Trigger reconcile (runs in the background - err means the deployment wasn't found, which is unlikely)
		if project.ProdDeploymentID != nil {
			err = s.TriggerReconcile(ctx, *project.ProdDeploymentID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) processGithubInstallationEvent(ctx context.Context, event *github.InstallationEvent) error {
	switch event.GetAction() {
	case "created", "unsuspend", "suspend", "new_permissions_accepted":
		// TODO: Should we do anything for unsuspend?
	case "deleted":
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
