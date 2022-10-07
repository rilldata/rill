package sqlite

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

// FindInstances implements drivers.RegistryStore
func (c *connection) FindInstances(ctx context.Context) []*drivers.Instance {
	panic(fmt.Errorf("not implemented"))
}

// FindInstance implements drivers.RegistryStore
func (c *connection) FindInstance(ctx context.Context, id string) (*drivers.Instance, bool) {
	panic(fmt.Errorf("not implemented"))
}

// CreateInstance implements drivers.RegistryStore
func (c *connection) CreateInstance(ctx context.Context, instance *drivers.Instance) error {
	panic(fmt.Errorf("not implemented"))
}

// DeleteInstance implements drivers.RegistryStore
func (c *connection) DeleteInstance(ctx context.Context, id string) error {
	panic(fmt.Errorf("not implemented"))
}

// FindRepos implements drivers.RegistryStore
func (c *connection) FindRepos(ctx context.Context) []*drivers.Repo {
	panic(fmt.Errorf("not implemented"))
}

// FindRepo implements drivers.RegistryStore
func (c *connection) FindRepo(ctx context.Context, id string) (*drivers.Repo, bool) {
	panic(fmt.Errorf("not implemented"))
}

// CreateRepo implements drivers.RegistryStore
func (c *connection) CreateRepo(ctx context.Context, repo *drivers.Repo) error {
	panic(fmt.Errorf("not implemented"))
}

// DeleteRepo implements drivers.RegistryStore
func (c *connection) DeleteRepo(ctx context.Context, id string) error {
	panic(fmt.Errorf("not implemented"))
}
