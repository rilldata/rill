package rillv1

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// IsInit returns true if a Rill project exists in the repo
func IsInit(ctx context.Context, repo drivers.RepoStore, instanceID string) bool {
	_, err := ParseRillYAML(ctx, repo, instanceID)
	return err == nil
}

// InitEmpty initializes an empty project
func InitEmpty(ctx context.Context, repo drivers.RepoStore, instanceID, title string) error {
	rillYAML := fmt.Sprintf("compiler: %s\n\ntitle: %q\n", Version, title)

	err := repo.Put(ctx, instanceID, "rill.yaml", strings.NewReader(rillYAML))
	if err != nil {
		return err
	}

	gitignore, _ := repo.Get(ctx, instanceID, ".gitignore")
	if gitignore != "" {
		gitignore += "\n"
	}
	gitignore += "# Rill\n*.db\n*.db.tmp\n*.db.wal\n.env\n"

	err = repo.Put(ctx, instanceID, ".gitignore", strings.NewReader(gitignore))
	if err != nil {
		return err
	}

	err = repo.Put(ctx, instanceID, "sources/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	err = repo.Put(ctx, instanceID, "models/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	err = repo.Put(ctx, instanceID, "dashboards/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	return nil
}
