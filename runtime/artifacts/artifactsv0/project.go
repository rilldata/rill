package artifactsv0

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

func IsInit(ctx context.Context, repo drivers.RepoStore, instanceID string) bool {
	_, err := repo.Get(ctx, instanceID, "rill.yaml")
	return err == nil
}

func InitEmpty(ctx context.Context, repo drivers.RepoStore, instanceID string, name string) error {
	err := repo.Put(ctx, instanceID, "rill.yaml", strings.NewReader(fmt.Sprintf("version: 0.0.1\n\nname: %s\n", name)))
	if err != nil {
		return err
	}

	err = repo.Put(ctx, instanceID, ".gitignore", strings.NewReader("*.db\n"))
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
