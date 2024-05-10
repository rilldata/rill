package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type sqlStoreToSelfExecutor struct {
	c        *connection
	sqlstore drivers.SQLStore
	opts     *drivers.ModelExecutorOptions
}

var _ drivers.ModelExecutor = &sqlStoreToSelfExecutor{}

func (e *sqlStoreToSelfExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
	olap, ok := e.c.AsOLAP(e.c.instanceID)
	if !ok {
		return nil, fmt.Errorf("output connector is not OLAP")
	}

	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(e.opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(e.opts); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = e.opts.ModelName
		usedModelName = true
	}

	tableName := outputProps.Table
	stagingTableName := tableName
	if !e.opts.IncrementalRun {
		if e.opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}

		// NOTE: This intentionally drops the end table if not staging changes.
		if t, err := olap.InformationSchema().Lookup(ctx, "", "", stagingTableName); err == nil {
			_ = olap.DropTable(ctx, stagingTableName, t.View)
		}
	}

	err := e.queryAndInsert(ctx, olap, stagingTableName, outputProps)
	if err != nil {
		if !e.opts.IncrementalRun {
			_ = olap.DropTable(ctx, stagingTableName, false)
		}
		return nil, err
	}

	if !e.opts.IncrementalRun {
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
		Connector:  e.opts.OutputConnector,
		Properties: resultPropsMap,
		Table:      tableName,
	}, nil
}

func (e *sqlStoreToSelfExecutor) queryAndInsert(ctx context.Context, olap drivers.OLAPStore, outputTable string, outputProps *ModelOutputProperties) (err error) {
	start := time.Now()
	e.c.logger.Debug("duckdb: sqlstore transfer started", zap.String("model", e.opts.ModelName), observability.ZapCtx(ctx))
	defer func() {
		e.c.logger.Debug("duckdb: sqlstore transfer finished", zap.Duration("elapsed", time.Since(start)), zap.Bool("success", err == nil), zap.Error(err), observability.ZapCtx(ctx))
	}()

	storageLimitBytes := e.c.config.StorageLimitBytes
	if storageLimitBytes == 0 {
		storageLimitBytes = math.MaxInt64
	}

	iter, err := e.sqlstore.QueryAsFiles(ctx, e.opts.InputProperties, &drivers.QueryOption{TotalLimitInBytes: storageLimitBytes}, drivers.NoOpProgress{})
	if err != nil {
		return err
	}
	defer iter.Close()

	create := !e.opts.IncrementalRun
	for {
		files, err := iter.Next()
		if err != nil {
			// TODO: Why is this not just one error?
			if errors.Is(err, io.EOF) || errors.Is(err, drivers.ErrNoRows) || errors.Is(err, drivers.ErrIteratorDone) {
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

		if !create && e.opts.IncrementalRun {
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

		err = olap.CreateTableAsSelect(ctx, outputTable, false, qry)
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
