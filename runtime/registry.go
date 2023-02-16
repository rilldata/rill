package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

func (r *Runtime) FindInstances(ctx context.Context) ([]*drivers.Instance, error) {
	return r.Registry().FindInstances(ctx)
}

func (r *Runtime) FindInstance(ctx context.Context, instanceID string) (*drivers.Instance, error) {
	return r.Registry().FindInstance(ctx, instanceID)
}

func (r *Runtime) CreateInstance(ctx context.Context, inst *drivers.Instance, envString string) error {
	// Check OLAP connection
	olap, err := drivers.Open(inst.OLAPDriver, inst.OLAPDSN, r.logger)
	if err != nil {
		return err
	}
	_, ok := olap.OLAPStore()
	if !ok {
		return fmt.Errorf("not a valid OLAP driver: '%s'", inst.OLAPDriver)
	}

	// Check repo connection
	repo, err := drivers.Open(inst.RepoDriver, inst.RepoDSN, r.logger)
	if err != nil {
		return err
	}
	repoStore, ok := repo.RepoStore()
	if !ok {
		return fmt.Errorf("not a valid repo driver: '%s'", inst.RepoDriver)
	}

	// Check that it's a driver that supports embedded catalogs
	if inst.EmbedCatalog {
		_, ok := olap.CatalogStore()
		if !ok {
			return errors.New("driver does not support embedded catalogs")
		}
	}

	// Prepare connections for use
	err = olap.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare instance: %w", err)
	}
	err = repo.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare instance: %w", err)
	}

	file, _ := repoStore.Get(ctx, inst.ID, "rill.yaml")

	env, err := drivers.NewEnvVariables(ctx, file, envString)
	if err != nil {
		return fmt.Errorf("failed to parse env variables %w", err)
	}
	inst.Env = &env

	// Create instance
	err = r.Registry().CreateInstance(ctx, inst)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) DeleteInstance(ctx context.Context, instanceID string) error {
	err := r.Registry().DeleteInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	return nil
}
