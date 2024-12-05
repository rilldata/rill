package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type warehouseToSelfExecutor struct {
	c *connection
	w drivers.Warehouse
}

var _ drivers.ModelExecutor = &warehouseToSelfExecutor{}

func (e *warehouseToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *warehouseToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	olap, ok := e.c.AsOLAP(e.c.instanceID)
	if !ok {
		return nil, fmt.Errorf("output connector is not OLAP")
	}

	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(opts); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	tableName := outputProps.Table
	stagingTableName := tableName
	if !opts.IncrementalRun {
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}

		// NOTE: This intentionally drops the end table if not staging changes.
		_ = olap.DropTable(ctx, stagingTableName)
	}

	err := e.queryAndInsert(ctx, opts, olap, stagingTableName, outputProps)
	if err != nil {
		if !opts.IncrementalRun {
			_ = olap.DropTable(ctx, stagingTableName)
		}
		return nil, err
	}

	if !opts.IncrementalRun {
		if stagingTableName != tableName {
			err = olapForceRenameTable(ctx, olap, stagingTableName, false, tableName)
			if err != nil {
				return nil, fmt.Errorf("failed to rename staged model: %w", err)
			}
		}
	}

	resultProps := &ModelResultProperties{
		Table:         tableName,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: resultPropsMap,
		Table:      tableName,
	}, nil
}

func (e *warehouseToSelfExecutor) queryAndInsert(ctx context.Context, opts *drivers.ModelExecuteOptions, olap drivers.OLAPStore, outputTable string, outputProps *ModelOutputProperties) (err error) {
	start := time.Now()
	e.c.logger.Debug("duckdb: warehouse transfer started", zap.String("model", opts.ModelName), observability.ZapCtx(ctx))
	defer func() {
		e.c.logger.Debug("duckdb: warehouse transfer finished", zap.Duration("elapsed", time.Since(start)), zap.Bool("success", err == nil), zap.Error(err), observability.ZapCtx(ctx))
	}()

	iter, err := e.w.QueryAsFiles(ctx, opts.InputProperties)
	if err != nil {
		return err
	}
	defer iter.Close()

	create := !opts.IncrementalRun
	for {
		files, err := iter.Next()
		if err != nil {
			// TODO: Why is this not just one error?
			if errors.Is(err, io.EOF) || errors.Is(err, drivers.ErrNoRows) {
				break
			}
			return err
		}

		format := fileutil.FullExt(files[0])
		if iter.Format() != "" {
			format += "." + iter.Format()
		}

		from, err := sourceReader(files, format, make(map[string]any))
		if err != nil {
			return err
		}
		qry := fmt.Sprintf("SELECT * FROM %s", from)

		if !create && opts.IncrementalRun {
			err := olap.InsertTableAsSelect(ctx, outputTable, qry, false, true, outputProps.IncrementalStrategy, outputProps.UniqueKey)
			if err != nil {
				return fmt.Errorf("failed to incrementally insert into table: %w", err)
			}
			continue
		}

		if !create {
			err := olap.InsertTableAsSelect(ctx, outputTable, qry, false, true, drivers.IncrementalStrategyAppend, nil)
			if err != nil {
				return fmt.Errorf("failed to insert into table: %w", err)
			}
			continue
		}

		err = olap.CreateTableAsSelect(ctx, outputTable, false, qry, nil)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}

		create = false
	}

	// We were supposed to create the table, but didn't get any data
	if create {
		return drivers.ErrNoRows
	}

	return nil
}
