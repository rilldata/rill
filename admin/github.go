package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrUserIsNotCollaborator = fmt.Errorf("user is not a collaborator for the repository")

// GetGithubInstallation returns a Github installation ID iff the Github App is installed on the repository AND we have a confirmed relationship between the user and that installation.
// The githubURL should be a HTTPS URL for a Github repository without the .git suffix.
func (s *Service) GetGithubInstallation(ctx context.Context, userID, githubURL string) (int64, bool, error) {
	account, repo, ok := gitutil.SplitGithubURL(githubURL)
	if !ok {
		return 0, false, fmt.Errorf("invalid Github URL %q", githubURL)
	}

	// TODO :: handle suspended case
	installation, resp, err := s.Github.Apps.FindRepositoryInstallation(ctx, account, repo)
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

	// The user has access to the installation
	return installationID, true, nil
}

// LookupGithubRepoForUser calls the Github API using an installation token to get information about a Github repo.
// The githubURL should be a HTTPS URL for a Github repository without the .git suffix.
func (s *Service) LookupGithubRepoForUser(ctx context.Context, installationID int64, githubURL, gitUserName string) (*github.Repository, error) {
	account, repo, ok := gitutil.SplitGithubURL(githubURL)
	if !ok {
		return nil, fmt.Errorf("invalid Github URL %q", githubURL)
	}

	gh, err := s.githubInstallationClient(installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create github installation client: %w", err)
	}

	isColab, resp, err := gh.Repositories.IsCollaborator(ctx, account, repo, gitUserName)
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

// IsUserExist checks if user with userName exists on Github
func (s *Service) IsUserExist(ctx context.Context, userName string) (bool, error) {
	if userName == "" {
		return false, nil
	}

	user, resp, err := s.Github.Users.Get(ctx, userName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}

	return user.GetLogin() == userName, nil
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
