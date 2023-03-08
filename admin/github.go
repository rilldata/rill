package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
)

// TrackGithubInstallation TODO
func (s *Service) TrackGithubInstallation(ctx context.Context, userID string, installationID int64) error {
	return nil
}

// HasGithubInstallation
func (s *Service) HasGithubInstallation(ctx context.Context, userID, githubURL string) (bool, error) {
	// Parse SSH or HTTP endpoints
	// endpoint, err := transport.NewEndpoint(remote)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	// path := endpoint.Path
	// if strings.Contains(endpoint.Protocol, "http") {
	// 	_, path, _ = strings.Cut(path, "/")
	// }
	// fullName, _, _ = strings.Cut(path, ".git")

	// owner, repo, found := strings.Cut(fullName, "/")
	// if !found {
	// 	// invalid remote
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// installation, response, err := s.github.Apps.FindRepositoryInstallation(ctx, owner, repo)
	// if err != nil {
	// 	if response.StatusCode == http.StatusNotFound {
	// 		// Don't have access
	// 		return false, nil
	// 	}
	// 	// Unexpected
	// 	return false, err
	// }

	// installationID := installation.GetID()
	// if installationID == 0 {
	// 	// TODO: What does that even mean?
	// 	return false, fmt.Errorf("weird")
	// }

	// project.GithubAppInstallID = installation.GetID()
	// project.GitURL, err = s.githubAuthenticatedRemote(ctx, installationID, project.GitURL, owner, repo)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	return false, nil
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
	githubURL := *repo.CloneURL
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

	// TODO: Trigger deployment (unless currently deploying)
	// d := deployment.LocalDeployment{Logger: g.logger}
	// return d.DeployProject(project)

	return nil
}

func (s *Service) processGithubInstallationEvent(ctx context.Context, event *github.InstallationEvent) error {
	// We can access: event.Repositories and event.GetInstallation().GetID()

	switch event.GetAction() {
	case "created", "unsuspend":
		// TODO: Should we do anything for unsuspend?
	case "suspend", "deleted":
		// TODO: What to do about existing projects deploying from that installation?
	case "new_permissions_accepted":
		// TODO: Any caches to update here?
	}

	return nil
}

func (s *Service) processGithubInstallationRepositoriesEvent(ctx context.Context, event *github.InstallationRepositoriesEvent) error {
	// We can access event.RepositoriesAdded and event.RepositoriesRemoved

	// TODO: What to do about existing projects that have been removed?

	return nil
}

// githubAuthenticatedRemote builds an authenticated Git URL for a remote in an installation.
func (s *Service) githubAuthenticatedRemote(ctx context.Context, installationID int64, remote, owner, repoName string) (string, error) {
	client, err := s.githubInstallationClient(installationID)
	if err != nil {
		return "", err
	}

	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return "", err
	}

	httpURL := repo.GetCloneURL()
	if httpURL == "" {
		// should hopefully never happen
		return "", fmt.Errorf("no http url")
	}
	httpEndpoint, _ := transport.NewEndpoint(httpURL)
	httpEndpoint.User = "__githubapp_installation_id__"
	httpEndpoint.Password = fmt.Sprint(installationID)
	return httpEndpoint.String(), nil
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
