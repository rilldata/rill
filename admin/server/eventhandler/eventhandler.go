package eventhandler

import (
	"context"
	"errors"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/deployment"
)

// Handler processes web hook events
type Handler interface {
	// Process the event
	Process(ctx context.Context, eventData any) error
}

var ErrInvalidEvent = errors.New("invalid payload")

type githubHandler struct {
	db database.DB
}

// NewGithubHandler returns a handler that processes github web hook events
func NewGithubHandler(db database.DB) (Handler, error) {
	return &githubHandler{db: db}, nil
}

func (g *githubHandler) Process(ctx context.Context, raw any) error {
	switch event := raw.(type) {
	case *github.PushEvent:
		return g.processPushEvent(ctx, event)
	case *github.InstallationEvent: // triggered during first installation of app to org or some repos
		return g.processInstallationEvent(ctx, event)
	case *github.InstallationRepositoriesEvent: // triggered when new repos are added to the org/account and installation access given on full org
		return g.processInstallationRepositoriesEvent(ctx, event)
	default:
		return nil
	}
}

func (g *githubHandler) processPushEvent(ctx context.Context, event *github.PushEvent) error {
	// can move these validations into validate if processing event in async
	repo := event.GetRepo()
	if repo == nil {
		return ErrInvalidEvent
	}

	fullName := repo.GetFullName()
	if fullName == "" {
		return ErrInvalidEvent
	}

	project, err := g.db.FindProjectByGitFullName(ctx, fullName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// app installed on repo not existing in our db
			return nil
		}
		return err
	}

	if !isDeployBranch(event, project.ProductionBranch) {
		return nil
	}

	// some cases when this can happen
	// 1. missed install event (unlikely since we update id in setup callback as well)
	// 2. app installed on repo first and project connected later (may be navigate user to project connect in setup callback)
	// 3. Events out of order where installation removed event came first and github push later (redeploy in private repo will anyways fail since access is no longer present)
	// should we handle this ?
	installID := event.GetInstallation().GetID()
	if installID != 0 && project.GithubAppInstallID != installID {
		project.GithubAppInstallID = installID
		_, err = g.db.UpdateProject(ctx, project)
		if err != nil {
			return err
		}
	}

	// this is just for MVP
	d := deployment.LocalDeployment{}
	return d.DeployProject(project)
}

func (g *githubHandler) processInstallationEvent(ctx context.Context, event *github.InstallationEvent) error {
	installationID := event.GetInstallation().GetID()
	if installationID == 0 {
		return ErrInvalidEvent
	}

	for _, repo := range event.Repositories {
		project, err := g.db.FindProjectByGitFullName(ctx, repo.GetFullName())
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				// app installed on repo not existing in our db
				continue
			}
			return err
		}
		// The action that was performed. Can be either "created", "deleted", "suspend", "unsuspend" or "new_permissions_accepted".
		switch event.GetAction() {
		case "created", "unsuspend":
			project.GithubAppInstallID = installationID
			// ignoring error
			_, _ = g.db.UpdateProject(ctx, project)
		case "suspend", "deleted":
			project.GithubAppInstallID = 0
			// ignoring error
			_, _ = g.db.UpdateProject(ctx, project)
		case "new_permissions_accepted":
			// do nothing for now
		}
	}
	return nil
}

func (g *githubHandler) processInstallationRepositoriesEvent(ctx context.Context, event *github.InstallationRepositoriesEvent) error {
	installationID := event.GetInstallation().GetID()
	if installationID == 0 {
		return ErrInvalidEvent
	}

	for _, repo := range event.RepositoriesAdded {
		project, err := g.db.FindProjectByGitFullName(ctx, repo.GetFullName())
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				// app installed on repo not existing in our db
				continue
			}
			return err
		}
		project.GithubAppInstallID = installationID
		// ignoring error
		_, _ = g.db.UpdateProject(ctx, project)
	}

	for _, repo := range event.RepositoriesRemoved {
		project, err := g.db.FindProjectByGitFullName(ctx, repo.GetFullName())
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				// app installed on repo not existing in our db
				continue
			}
			return err
		}
		project.GithubAppInstallID = 0
		// ignoring error
		_, _ = g.db.UpdateProject(ctx, project)
	}
	return nil
}

// use either user-provied branch or default branch of the repo
func isDeployBranch(event *github.PushEvent, prodBranch string) bool {
	ref := event.GetRef()
	if ref == "" {
		return false
	}
	// format is refs/heads/main or refs/tags/v3.14.1
	_, branch, found := strings.Cut(ref, "refs/heads/")
	if !found {
		// a tag push
		return false
	}

	if prodBranch != "" {
		return branch == prodBranch
	}
	return branch == event.GetRepo().GetDefaultBranch()
}
