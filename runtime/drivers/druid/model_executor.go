package druid

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) Exists(ctx context.Context, res *drivers.ModelResult) (bool, error) {
	return false, nil
}

func (c *connection) Delete(ctx context.Context, res *drivers.ModelResult) error {
	return nil
}

func (c *connection) Rename(ctx context.Context, res *drivers.ModelResult, newName string, env *drivers.ModelEnv) (*drivers.ModelResult, error) {
	resPropsMap := map[string]interface{}{}

	return &drivers.ModelResult{
		Connector:  res.Connector,
		Properties: resPropsMap,
		Table:      newName,
	}, nil
}
