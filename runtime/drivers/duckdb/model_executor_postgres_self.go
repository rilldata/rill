package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type pgxToSelfInputProps struct {
	SQL         string `mapstructure:"sql"`
	DatabaseURL string `mapstructure:"database_url"`
}

type postgresToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &postgresToSelfExecutor{}

func (e *postgresToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *postgresToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	olap, _ := e.c.AsOLAP(e.c.instanceID)
	inputProps := &pgxToSelfInputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
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
		if t, err := olap.InformationSchema().Lookup(ctx, "", "", stagingTableName); err == nil {
			_ = olap.DropTable(ctx, stagingTableName, t.View)
		}
	}

	err := e.ingestFromPgx(ctx, inputProps, outputProps, opts, stagingTableName)
	if err != nil {
		if !opts.IncrementalRun {
			_ = olap.DropTable(ctx, stagingTableName, false)
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

func (e *postgresToSelfExecutor) ingestFromPgx(ctx context.Context, inputProps *pgxToSelfInputProps, outputProps *ModelOutputProperties, opts *drivers.ModelExecuteOptions, table string) error {
	// we first ingest data in a temporary table in the main db
	// and then copy it to the final table to ensure that the final table is always created using CRUD APIs which takes care
	// whether table goes in main db or in separate table specific db
	safeTmpTable := safeName(fmt.Sprintf("__%s_tmp_postgres", table))
	err := e.c.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, conn *sql.Conn) error {
		db, schema, err := scanCurrentDatabaseAndSchema(ctx, e.c)
		if err != nil {
			return fmt.Errorf("failed to scan current database and schema: %w", err)
		}

		// Attach postgres database to duckdb
		err = e.c.Exec(ctx, &drivers.Statement{
			Query: fmt.Sprintf("ATTACH %s AS pgx_db (TYPE POSTGRES, READ_ONLY)", safeSQLString(inputProps.DatabaseURL)),
		})
		if err != nil {
			return fmt.Errorf("failed to attach postgres database: %w", err)
		}

		// defer detach
		defer func() {
			// detach postgres database
			err := e.c.Exec(ensuredCtx, &drivers.Statement{
				Query: "DETACH pgx_db",
			})
			if err != nil {
				e.c.logger.Error("failed to detach postgres database", zap.Error(err))
			}
		}()

		// switch to postgres database - this is required to directly run queries on postgres table without pgx_db database prefix
		err = e.c.Exec(ctx, &drivers.Statement{
			Query: "USE pgx_db",
		})
		if err != nil {
			return fmt.Errorf("failed to switch to postgres database: %w", err)
		}

		defer func() {
			// switch back to local db and local schema
			err := e.c.Exec(ensuredCtx, &drivers.Statement{
				Query: fmt.Sprintf("USE %s.%s", safeName(db), safeName(schema)),
			})
			if err != nil {
				e.c.fatalInternalError(fmt.Errorf("failed to switch back to local db and schema: %w", err))
			}
		}()

		// ingest from postgres
		userQuery, _ := strings.CutSuffix(strings.TrimSpace(inputProps.SQL), ";") // trim trailing semi colon
		err = e.c.Exec(ctx, &drivers.Statement{
			Query: fmt.Sprintf("CREATE OR REPLACE TABLE %s.%s.%s AS (%s\n);", safeName(db), safeName(schema), safeTmpTable, userQuery),
		})
		if err != nil {
			return fmt.Errorf("failed to ingest data from postgres: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	defer func() {
		// drop temp table
		err := e.c.Exec(context.Background(), &drivers.Statement{
			Query: fmt.Sprintf("DROP TABLE %s", safeTmpTable),
		})
		if err != nil {
			e.c.logger.Info("failed to drop temp table", zap.Error(err))
		}
	}()
	// copy data from temp table to target table
	if opts.IncrementalRun {
		return e.c.InsertTableAsSelect(ctx, table, fmt.Sprintf("SELECT * FROM %s", safeTmpTable), false, true, outputProps.IncrementalStrategy, outputProps.UniqueKey)
	}
	return e.c.CreateTableAsSelect(ctx, table, false, fmt.Sprintf("SELECT * FROM %s", safeTmpTable), nil)
}

func scanCurrentDatabaseAndSchema(ctx context.Context, c *connection) (string, string, error) {
	res, err := c.Execute(ctx, &drivers.Statement{
		Query: "SELECT current_database(),current_schema()",
	})
	if err != nil {
		return "", "", err
	}
	defer res.Close()
	var localDB, localSchema string
	for res.Next() {
		if err := res.Scan(&localDB, &localSchema); err != nil {
			return "", "", err
		}
	}
	return localDB, localSchema, nil
}
