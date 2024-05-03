package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type ModelInputProperties struct {
	SQL string `mapstructure:"sql"`
}

func (p *ModelInputProperties) Validate() error {
	if p.SQL == "" {
		return fmt.Errorf("missing property 'sql'")
	}
	return nil
}

type ModelOutputProperties struct {
	Table       string `mapstructure:"table"`
	Materialize *bool  `mapstructure:"materialize"`
}

type ModelResultProperties struct {
	Table         string `mapstructure:"table"`
	View          bool   `mapstructure:"view"`
	UsedModelName bool   `mapstructure:"used_model_name"`
}

func (p *ModelOutputProperties) Validate() error {
	return nil
}

func (c *connection) Supports(ctx context.Context, opts *drivers.ModelExecuteOptions) (bool, error) {
	input, output, release, err := c.acquireHandles(ctx, opts)
	if err != nil {
		return false, err
	}
	defer release()

	if input.Driver() == "duckdb" && output.Driver() == "duckdb" && opts.InputConnector == opts.OutputConnector {
		return true, nil
	}

	return false, nil
}

func (c *connection) Run(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelExecuteResult, error) {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return nil, fmt.Errorf("output connector is not OLAP")
	}

	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	materialize := opts.Env.DefaultMaterialize
	if outputProps.Materialize != nil {
		materialize = *outputProps.Materialize
	}

	tableName := outputProps.Table
	stagingTableName := tableName
	if opts.Env.StageChanges {
		stagingTableName = stagingTableNameFor(tableName)
	}

	// Drop the staging view/table if it exists.
	// NOTE: This intentionally drops the end table if not staging changes.
	stagingTable, err := olap.InformationSchema().Lookup(ctx, "", "", stagingTableName)
	if err == nil {
		_ = olap.DropTable(ctx, stagingTableName, stagingTable.View)
	}

	err = olap.CreateTableAsSelect(ctx, stagingTableName, !materialize, inputProps.SQL)
	if err != nil {
		_ = olap.DropTable(ctx, stagingTableName, !materialize)
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// Rename the staging table to the final table name
	if stagingTableName != tableName {
		err = olapForceRenameTable(ctx, olap, stagingTableName, !materialize, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to rename staged model: %w", err)
		}
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          !materialize,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelExecuteResult{
		Connector:  opts.OutputConnector,
		Properties: resultPropsMap,
		Table:      tableName,
	}, nil
}

func (c *connection) Rename(ctx context.Context, opts *drivers.ModelRenameOptions) (*drivers.ModelExecuteResult, error) {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return nil, fmt.Errorf("connector is not an OLAP")
	}

	prevResultProps := &ModelResultProperties{}
	if err := mapstructure.WeakDecode(opts.PreviousResult.Properties, prevResultProps); err != nil {
		return nil, fmt.Errorf("failed to parse previous result properties: %w", err)
	}

	if !prevResultProps.UsedModelName {
		return opts.PreviousResult, nil
	}

	err := olapForceRenameTable(ctx, olap, prevResultProps.Table, prevResultProps.View, opts.NewName)
	if err != nil {
		return nil, fmt.Errorf("failed to rename model: %w", err)
	}

	prevResultProps.Table = opts.NewName
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(prevResultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	return &drivers.ModelExecuteResult{
		Connector:  opts.PreviousResult.Connector,
		Properties: resultPropsMap,
		Table:      opts.NewName,
	}, nil
}

func (c *connection) Exists(ctx context.Context, res *drivers.ModelExecuteResult) (bool, error) {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return false, fmt.Errorf("connector is not an OLAP")
	}

	_, err := olap.InformationSchema().Lookup(ctx, "", "", res.Table)
	return err == nil, nil
}

func (c *connection) Delete(ctx context.Context, res *drivers.ModelExecuteResult) error {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return fmt.Errorf("connector is not an OLAP")
	}

	stagingTable, err := olap.InformationSchema().Lookup(ctx, "", "", stagingTableNameFor(res.Table))
	if err == nil {
		_ = olap.DropTable(ctx, stagingTable.Name, stagingTable.View)
	}

	table, err := olap.InformationSchema().Lookup(ctx, "", "", res.Table)
	if err != nil {
		return err
	}

	return olap.DropTable(ctx, table.Name, table.View)
}

func (c *connection) acquireHandles(ctx context.Context, opts *drivers.ModelExecuteOptions) (drivers.Handle, drivers.Handle, func(), error) {
	input, releaseInput, err := opts.Env.AcquireConnector(ctx, opts.InputConnector)
	if err != nil {
		return nil, nil, nil, err
	}

	if opts.InputConnector == opts.OutputConnector {
		return input, input, releaseInput, nil
	}

	output, releaseOutput, err := opts.Env.AcquireConnector(ctx, opts.OutputConnector)
	if err != nil {
		releaseInput()
		return nil, nil, nil, err
	}

	release := func() {
		releaseInput()
		releaseOutput()
	}

	return input, output, release, nil
}

// stagingTableName returns a stable temporary table name for a destination table.
// By using a stable temporary table name, we can ensure proper garbage collection without managing additional state.
func stagingTableNameFor(table string) string {
	return "__rill_tmp_model_" + table
}

// olapForceRenameTable renames a table or view from fromName to toName in the OLAP connector.
// If a view or table already exists with toName, it is overwritten.
func olapForceRenameTable(ctx context.Context, olap drivers.OLAPStore, fromName string, fromIsView bool, toName string) error {
	if fromName == "" || toName == "" {
		return fmt.Errorf("cannot rename empty table name: fromName=%q, toName=%q", fromName, toName)
	}

	if fromName == toName {
		return nil
	}

	// Infer SQL keyword for the table type
	var typ string
	if fromIsView {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	// Renaming a table to the same name with different casing is not supported. Workaround by renaming to a temporary name first.
	if strings.EqualFold(fromName, toName) {
		tmpName := fmt.Sprintf("__rill_tmp_rename_%s_%s", typ, toName)
		err := olap.RenameTable(ctx, fromName, tmpName, fromIsView)
		if err != nil {
			return err
		}
		fromName = tmpName
	}

	// Do the rename
	return olap.RenameTable(ctx, fromName, toName, fromIsView)
}
