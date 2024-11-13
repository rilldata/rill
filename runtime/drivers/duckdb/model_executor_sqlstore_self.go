package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type sqlStoreToSelfInputProps struct {
	SQL         string `mapstructure:"sql"`
	DatabaseURL string `mapstructure:"database_url"`
	DSN         string `mapstructure:"dsn"`
}

func (p *sqlStoreToSelfInputProps) Validate() error {
	if p.SQL == "" {
		return fmt.Errorf("missing property 'sql'")
	}
	if p.DatabaseURL == "" && p.DSN == "" {
		return fmt.Errorf("missing property `dsn`")
	}
	if p.DatabaseURL != "" { // postgres connector calls it database_url
		p.DSN = p.DatabaseURL
	}
	return nil
}

type sqlStoreToSelfExecutor struct {
	c      *connection
	driver string // mysql or postgres
}

var _ drivers.ModelExecutor = &sqlStoreToSelfExecutor{}

func (e *sqlStoreToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *sqlStoreToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	olap, _ := e.c.AsOLAP(e.c.instanceID)
	inputProps := &sqlStoreToSelfInputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
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

	var err error
	switch e.driver {
	case "mysql":
		err = e.ingestFromMySQL(ctx, inputProps, outputProps, opts, stagingTableName)
	case "postgres":
		err = e.ingestFromPgx(ctx, inputProps, outputProps, opts, stagingTableName)
	default:
		return nil, fmt.Errorf("unsupported sql store driver: %s", e.driver)
	}
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

func (e *sqlStoreToSelfExecutor) ingestFromPgx(ctx context.Context, inputProps *sqlStoreToSelfInputProps, outputProps *ModelOutputProperties, opts *drivers.ModelExecuteOptions, table string) error {
	// Attach postgres database to duckdb
	err := e.c.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf("ATTACH %s AS pgx_db (TYPE POSTGRES, READ_ONLY)", safeSQLString(inputProps.DSN)),
	})
	if err != nil {
		return fmt.Errorf("failed to attach postgres database: %w", err)
	}

	defer func() {
		// detach postgres database
		err := e.c.Exec(context.Background(), &drivers.Statement{
			Query: "DETACH pgx_db",
		})
		if err != nil {
			e.c.logger.Error("failed to detach postgres database", zap.Error(err))
		}
	}()

	// ingest from postgres
	userQuery, _ := strings.CutSuffix(inputProps.SQL, ";") // trim trailing semi colon
	query := fmt.Sprintf("SELECT * FROM postgres_query('pgx_db', %s)", safeSQLString(userQuery))
	if opts.IncrementalRun {
		return e.c.InsertTableAsSelect(ctx, table, query, false, true, outputProps.IncrementalStrategy, outputProps.UniqueKey)
	}
	return e.c.CreateTableAsSelect(ctx, table, false, query, nil)
}

func (e *sqlStoreToSelfExecutor) ingestFromMySQL(ctx context.Context, inputProps *sqlStoreToSelfInputProps, outputProps *ModelOutputProperties, opts *drivers.ModelExecuteOptions, table string) error {
	// Install and load extension. Does not auto load.
	err := e.c.Exec(ctx, &drivers.Statement{
		Query: "INSTALL MYSQL; LOAD MYSQL;",
	})
	if err != nil {
		return fmt.Errorf("failed to install/load mysql extension: %w", err)
	}

	// Attach database to duckdb
	err = e.c.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf("ATTACH %s AS mysql_db (TYPE MYSQL, READ_ONLY)", safeSQLString(inputProps.DSN)),
	})
	if err != nil {
		return fmt.Errorf("failed to attach mysql database: %w", err)
	}

	defer func() {
		// detach database
		err := e.c.Exec(context.Background(), &drivers.Statement{
			Query: "DETACH mysql_db",
		})
		if err != nil {
			e.c.logger.Error("failed to detach mysql database", zap.Error(err))
		}
	}()

	// ingest from mysql
	userQuery, _ := strings.CutSuffix(inputProps.SQL, ";") // trim trailing semi colon
	query := fmt.Sprintf("SELECT * FROM mysql_query('mysql_db', %s)", safeSQLString(userQuery))
	if opts.IncrementalRun {
		return e.c.InsertTableAsSelect(ctx, table, query, false, true, outputProps.IncrementalStrategy, outputProps.UniqueKey)
	}
	return e.c.CreateTableAsSelect(ctx, table, false, query, nil)
}
