package eventhandler

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
)

type Handler interface {
	Process(ctx context.Context, eventData any) error
}

var ErrInvalidEvent = errors.New("invalid payload")

type githubHandler struct {
	db database.DB
}

func NewGithubHandler(db database.DB) (Handler, error) {
	return &githubHandler{db: db}, nil
}

func (g *githubHandler) Process(ctx context.Context, raw any) error {
	switch event := raw.(type) {
	case *github.PushEvent:
		return g.processPushEvent(ctx, event)
	case *github.InstallationEvent: // triggered during first installation of app to org or some repos
		return g.processInstallationEvent(ctx, event)
	case *github.InstallationRepositoriesEvent: // triggered when new repos are added to the org
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

	gitURL := repo.GetGitURL()
	if gitURL == "" {
		return ErrInvalidEvent
	}

	project, err := g.db.FindProjectByGithubURL(ctx, gitURL)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// app installed on repo not existing in our db
			return nil
		}
		return err
	}

	// format is refs/heads/main or refs/tags/v3.14.1
	ref := event.GetRef()
	if ref == "" {
		return ErrInvalidEvent
	}
	_, branch, found := strings.Cut(ref, "refs/heads/")
	if !found || branch != project.ProductionBranch.String {
		// a tag push or a push on another branch
		return nil
	}

	installID := event.GetInstallation().GetID()
	if installID != 0 && project.GithubAppInstallID.Valid && project.GithubAppInstallID.Int64 != event.GetInstallation().GetID() {
		// missed install event, update installation ID
		project.GithubAppInstallID = sql.NullInt64{Int64: installID, Valid: true}
		_, err = g.db.UpdateProject(ctx, project)
		if err != nil {
			return err
		}
	}

	// todo :: trigger app's reconcile
	return nil
}

func (g *githubHandler) processInstallationEvent(ctx context.Context, event *github.InstallationEvent) error {
	// todo :: add or remove installation id from project
	return nil
}

func (g *githubHandler) processInstallationRepositoriesEvent(ctx context.Context, event *github.InstallationRepositoriesEvent) error {
	// todo :: add or remove installation id from project
	return nil
}
