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
	mockUsersInfo := "# These are example mock users to test your security policies.\n# For more information, see the documentation: https://docs.rilldata.com/manage/security"
	mockUsers := "mock_users:\n- email: john@yourcompany.com\n- email: jane@partnercompany.com"
	ignoredPaths := "ignored_paths:\n- /.idea\n- /.vscode" // TODO: get a list from product
	rillYAML := fmt.Sprintf("compiler: %s\n\ntitle: %q\n\n%s\n\n%s\n\n%s", Version, title, mockUsersInfo, mockUsers, ignoredPaths)

	err := repo.Put(ctx, "rill.yaml", strings.NewReader(rillYAML))
	if err != nil {
		return err
	}

	gitignore, _ := repo.Get(ctx, ".gitignore")
	if gitignore != "" {
		gitignore += "\n"
	}
	gitignore += ".DS_Store\n\n# Rill\n.env\ntmp\n"

	err = repo.Put(ctx, ".gitignore", strings.NewReader(gitignore))
	if err != nil {
		return err
	}

	err = repo.Put(ctx, "sources/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	err = repo.Put(ctx, "models/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	err = repo.Put(ctx, "dashboards/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	return nil
}
