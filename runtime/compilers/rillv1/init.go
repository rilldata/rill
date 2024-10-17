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
func InitEmpty(ctx context.Context, repo drivers.RepoStore, instanceID, displayName string) error {
	mockUsersInfo := "# These are example mock users to test your security policies.\n# Learn more: https://docs.rilldata.com/manage/security"
	mockUsers := "mock_users:\n- email: john@yourcompany.com\n- email: jane@partnercompany.com"
	rillYAML := fmt.Sprintf("compiler: %s\n\ndisplay_name: %q\n\n%s\n%s", Version, displayName, mockUsersInfo, mockUsers)

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

	return nil
}
