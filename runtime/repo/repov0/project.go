package repov0

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

func IsInit(ctx context.Context, repo drivers.RepoStore, instanceID string) bool {
	_, err := repo.Get(ctx, instanceID, "rill.yaml")
	return err == nil
}

func InitEmpty(ctx context.Context, repo drivers.RepoStore, instanceID string, name string) error {
	err := repo.PutBlob(ctx, instanceID, "rill.yaml", fmt.Sprintf("version: 0.0.1\n\nname: %s\n", name))
	if err != nil {
		return err
	}

	err = repo.PutBlob(ctx, instanceID, ".gitignore", "*.db\n")
	if err != nil {
		return err
	}

	err = repo.PutBlob(ctx, instanceID, "sources/.gitkeep", "")
	if err != nil {
		return err
	}

	err = repo.PutBlob(ctx, instanceID, "models/.gitkeep", "")
	if err != nil {
		return err
	}

	err = repo.PutBlob(ctx, instanceID, "dashboards/.gitkeep", "")
	if err != nil {
		return err
	}

	return nil
}
