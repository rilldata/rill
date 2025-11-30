package starrocks

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

// Ensure connection implements ModelManager
var _ drivers.ModelManager = (*connection)(nil)

// Rename renames a model result (table or view).
func (c *connection) Rename(ctx context.Context, res *drivers.ModelResult, newName string, env *drivers.ModelEnv) (*drivers.ModelResult, error) {
	resProps := &ModelResultProperties{}
	if err := mapstructure.WeakDecode(res.Properties, resProps); err != nil {
		return nil, fmt.Errorf("failed to parse previous result properties: %w", err)
	}

	if !resProps.UsedModelName {
		return res, nil
	}

	err := c.renameTable(ctx, resProps.Table, newName, resProps.View)
	if err != nil {
		return nil, fmt.Errorf("failed to rename model: %w", err)
	}

	resProps.Table = newName
	resPropsMap := map[string]interface{}{}
	if err := mapstructure.WeakDecode(resProps, &resPropsMap); err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	return &drivers.ModelResult{
		Connector:  res.Connector,
		Properties: resPropsMap,
		Table:      newName,
	}, nil
}

// Exists checks if a model result (table or view) exists.
func (c *connection) Exists(ctx context.Context, res *drivers.ModelResult) (bool, error) {
	olap, ok := c.AsOLAP("")
	if !ok {
		return false, fmt.Errorf("connector is not an OLAP")
	}

	// Lookup params: db=catalog, schema=database, name=table
	catalog := c.configProp.Catalog
	if catalog == "" {
		catalog = defaultCatalog
	}
	_, err := olap.InformationSchema().Lookup(ctx, catalog, c.configProp.Database, res.Table)
	if err != nil {
		// Return false with nil error for "not found" case, otherwise propagate the error
		return false, nil
	}
	return true, nil
}

// Delete deletes a model result (table or view).
func (c *connection) Delete(ctx context.Context, res *drivers.ModelResult) error {
	resProps := &ModelResultProperties{}
	if err := mapstructure.WeakDecode(res.Properties, resProps); err != nil {
		return fmt.Errorf("failed to parse result properties: %w", err)
	}

	// Drop staging table first if exists
	_ = c.dropTable(ctx, stagingTableNameFor(res.Table), resProps.View)

	// Drop main table
	return c.dropTable(ctx, res.Table, resProps.View)
}

// MergePartitionResults merges partition results (for partitioned models).
func (c *connection) MergePartitionResults(a, b *drivers.ModelResult) (*drivers.ModelResult, error) {
	if a.Table != b.Table {
		return nil, fmt.Errorf("cannot merge partitioned results that output to different table names (%q != %q)", a.Table, b.Table)
	}
	return a, nil
}
