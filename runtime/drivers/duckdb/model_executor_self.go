package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
)

type selfToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &selfToSelfExecutor{}

func (e *selfToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *selfToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	inputProps := &ModelInputProperties{}
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
	if err := outputProps.validateAndApplyDefaults(opts, inputProps, outputProps); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
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

	asView := !materialize
	tableName := outputProps.Table

	// Backward compatibility for the old duckdb SQL:
	// It was possible to set a duckdb SQL which ingests data from an object store without setting the object store credentials.
	// We did rewriting for path to rewrite object store paths to paths of locally downloaded files.
	// The handling can now be done with duckdb's native connectors by setting a SQL that creates secret to access the object store.
	// However duckdb does not support GCS's native credentials(google_application_credentials) so we still maintain the hack for the same.
	// We expect to remove this rewriting once all users start using GCS's s3 compatibility API support.
	if scheme, createSecretSQL, dropSecretSQL, ast, ok := objectStoreRef(ctx, inputProps, opts); ok {
		if createSecretSQL != "" {
			inputProps.InternalCreateSecretSQL = createSecretSQL
			inputProps.InternalDropSecretSQL = dropSecretSQL
		} else if scheme == "gcs" {
			// rewrite duckdb sql with locally downloaded files
			handle, release, err := opts.Env.AcquireConnector(ctx, scheme)
			if err != nil {
				return nil, err
			}
			defer release()
			path := ast.GetTableRefs()[0].Paths[0]
			deleteFiles, err := rewriteDuckDBSQL(ctx, inputProps, handle, path, ast)
			if err != nil {
				return nil, err
			}
			defer deleteFiles()
		} else {
			rewrittenSQL, err := rewriteLocalPaths(ast, opts.Env.RepoRoot, opts.Env.AllowHostAccess)
			if err != nil {
				return nil, fmt.Errorf("invalid local path: %w", err)
			}
			inputProps.SQL = rewrittenSQL
		}
	}

	// Add PreExec statements that create temporary secrets for object store connectors.
	for _, connector := range e.c.config.secretConnectors() {
		createSecretSQL, dropSecretSQL, err := objectStoreSecretSQL(ctx, opts, connector, "", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create secret for connector %q: %w", connector, err)
		}
		if inputProps.InternalCreateSecretSQL == "" {
			inputProps.InternalCreateSecretSQL = createSecretSQL
		} else {
			inputProps.InternalCreateSecretSQL += ";" + createSecretSQL
		}
		if inputProps.InternalDropSecretSQL == "" {
			inputProps.InternalDropSecretSQL = dropSecretSQL
		} else {
			inputProps.InternalDropSecretSQL += ";" + dropSecretSQL
		}
	}

	originalPreExec := inputProps.PreExec
	if inputProps.InternalCreateSecretSQL != "" {
		if inputProps.PreExec == "" {
			inputProps.PreExec = inputProps.InternalCreateSecretSQL
		} else {
			inputProps.PreExec += ";" + inputProps.InternalCreateSecretSQL
		}
	}
	duration, err := e.createOrInsertIntoDuckDB(ctx, opts, inputProps, outputProps, tableName, asView)
	if err != nil {
		// On failure, try cleaning up secrets and retry without secrets for anonymous bucket access
		if inputProps.InternalDropSecretSQL != "" && (strings.Contains(err.Error(), "HTTP 403") || strings.Contains(err.Error(), "region being set incorrectly")) {
			if originalPreExec == "" {
				inputProps.PreExec = inputProps.InternalDropSecretSQL
			} else {
				inputProps.PreExec = originalPreExec + ";" + inputProps.InternalDropSecretSQL
			}
			var anonymErr error
			duration, anonymErr = e.createOrInsertIntoDuckDB(ctx, opts, inputProps, outputProps, tableName, asView)
			if anonymErr != nil {
				return nil, errors.Join(err, anonymErr)
			}
		} else {
			return nil, err
		}
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelResult{
		Connector:    opts.OutputConnector,
		Properties:   resultPropsMap,
		Table:        tableName,
		ExecDuration: duration,
	}, nil
}

func (e *selfToSelfExecutor) createOrInsertIntoDuckDB(ctx context.Context, opts *drivers.ModelExecuteOptions, inputProps *ModelInputProperties,
	outputProps *ModelOutputProperties, tableName string, asView bool,
) (time.Duration, error) {
	var duration time.Duration
	if !opts.IncrementalRun {
		// Prepare for ingesting into the staging view/table.
		// NOTE: This intentionally drops the end table if not staging changes.
		stagingTableName := tableName
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}
		_ = e.c.dropTable(ctx, stagingTableName)

		// Create the table
		if inputProps.Database != "" {
			// special handling for ingesting from an external database
			//
			// not handling incremental use cases since ingesting from an external database is mostly for small,experimental use cases
			if opts.Incremental {
				return 0, fmt.Errorf("`incremental` models are not supported when ingesting data from external db files")
			}
			var err error
			inputProps.Database, err = fileutil.ResolveLocalPath(inputProps.Database, opts.Env.RepoRoot, opts.Env.AllowHostAccess)
			if err != nil {
				return 0, err
			}
			res, err := e.createFromExternalDuckDB(ctx, inputProps, stagingTableName)
			if err != nil {
				_ = e.c.dropTable(ctx, stagingTableName)
				return 0, fmt.Errorf("failed to create model: %w", err)
			}
			duration = res.Duration
		} else {
			createTableOpts := &createTableOptions{
				view:         asView,
				beforeCreate: inputProps.PreExec,
				afterCreate:  inputProps.PostExec,
			}
			if inputProps.InitQueries != "" {
				createTableOpts.initQueries = []string{inputProps.InitQueries}
			}
			res, err := e.c.createTableAsSelect(ctx, stagingTableName, inputProps.SQL, createTableOpts)
			if err != nil {
				_ = e.c.dropTable(ctx, stagingTableName)
				return 0, fmt.Errorf("failed to create model: %w", err)
			}
			duration = res.duration
		}

		// Rename the staging table to the final table name
		if stagingTableName != tableName {
			err := e.c.forceRenameTable(ctx, stagingTableName, asView, tableName)
			if err != nil {
				return 0, fmt.Errorf("failed to rename staged model: %w", err)
			}
		}
	} else {
		// Insert into the table
		insertTableOpts := &InsertTableOptions{
			BeforeInsert: inputProps.PreExec,
			AfterInsert:  inputProps.PostExec,
			ByName:       false,
			Strategy:     outputProps.IncrementalStrategy,
			UniqueKey:    outputProps.UniqueKey,
			PartitionBy:  outputProps.PartitionBy,
		}
		if inputProps.InitQueries != "" {
			insertTableOpts.InitQueries = []string{inputProps.InitQueries}
		}
		res, err := e.c.insertTableAsSelect(ctx, tableName, inputProps.SQL, insertTableOpts)
		if err != nil {
			return 0, fmt.Errorf("failed to incrementally insert into table: %w", err)
		}
		duration = res.duration
	}
	return duration, nil
}

func (e *selfToSelfExecutor) createFromExternalDuckDB(ctx context.Context, inputProps *ModelInputProperties, tbl string) (*rduckdb.TableWriteMetrics, error) {
	safeDBName := safeName(tbl + "_external_db_")
	safeTempTable := safeName(tbl + "__temp__")
	beforeCreateFn := func(ctx context.Context, conn *sqlx.Conn) error {
		if inputProps.PreExec != "" {
			if _, err := conn.ExecContext(ctx, inputProps.PreExec); err != nil {
				return err
			}
		}

		if _, err := conn.ExecContext(ctx, fmt.Sprintf("ATTACH %s AS %s (READ_ONLY)", safeSQLString(inputProps.Database), safeDBName)); err != nil {
			return err
		}

		var localDB, localSchema string
		if err := conn.QueryRowxContext(ctx, "SELECT current_database(),current_schema();").Scan(&localDB, &localSchema); err != nil {
			return err
		}

		if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE %s;", safeDBName)); err != nil {
			return err
		}

		userQuery := strings.TrimSpace(inputProps.SQL)
		userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s.%s.%s AS (%s\n);", safeName(localDB), safeName(localSchema), safeTempTable, userQuery)
		_, execErr := conn.ExecContext(ctx, query)
		// revert to localdb and schema before returning
		_, err := conn.ExecContext(ctx, fmt.Sprintf("USE %s.%s;", safeName(localDB), safeName(localSchema)))
		return errors.Join(execErr, err)
	}
	afterCreateFn := func(ctx context.Context, conn *sqlx.Conn) error {
		if inputProps.PostExec != "" {
			if _, err := conn.ExecContext(ctx, inputProps.PostExec); err != nil {
				return err
			}
		}
		_, err := conn.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", safeTempTable))
		return err
	}
	db, release, err := e.c.acquireDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = release()
	}()
	return db.CreateTableAsSelect(ctx, tbl, fmt.Sprintf("SELECT * FROM %s", safeTempTable), &rduckdb.CreateTableOptions{
		BeforeCreateFn: beforeCreateFn,
		AfterCreateFn:  afterCreateFn,
	})
}

func objectStoreRef(ctx context.Context, props *ModelInputProperties, opts *drivers.ModelExecuteOptions) (string, string, string, *duckdbsql.AST, bool) {
	// Skip path parsing and secret creation in the following cases:
	// 1. InternalCreateSecretSQL is already set.
	// 2. pre_exec is set we assume user have set the secret.
	// 3. non-DuckDB connectors which typically don't require secrets, and if they do, the secret is already defined in InternalCreateSecretSQL
	if props.InternalCreateSecretSQL != "" || props.PreExec != "" || opts.InputConnector != "duckdb" {
		return "", props.InternalCreateSecretSQL, props.InternalDropSecretSQL, nil, false
	}
	// Parse AST
	ast, err := duckdbsql.Parse(props.SQL)
	if err != nil {
		// If we can't parse the SQL just let duckdb run on it and give a sql parse error.
		return "", "", "", nil, false
	}

	// If there is a single table reference check if it is an object store reference.
	refs := ast.GetTableRefs()
	if len(refs) != 1 {
		return "", "", "", nil, false
	}
	ref := refs[0]
	// Parse the path as a URL (also works for local paths)
	if len(ref.Paths) == 0 {
		return "", "", "", nil, false
	}
	uri, err := url.Parse(ref.Paths[0])
	if err != nil {
		return "", "", "", nil, false
	}

	if uri.Scheme == "s3" || uri.Scheme == "azure" || uri.Scheme == "gcs" || uri.Scheme == "gs" {
		if uri.Scheme == "gs" {
			uri.Scheme = "gcs"
		}
		// for s3 and azure we can just set a duckdb secret and ingest data using duckdb's native support for s3 and azure
		createSecretSQL, dropSecretSQL, err := objectStoreSecretSQL(ctx, opts, uri.Scheme, ref.Paths[0], opts.InputProperties)
		if err != nil {
			if errors.Is(err, errGCSUsesNativeCreds) {
				return uri.Scheme, "", "", ast, true
			}
			return "", "", "", nil, false
		}
		return uri.Scheme, createSecretSQL, dropSecretSQL, ast, true
	}
	if uri.Scheme == "" && uri.Host == "" {
		// local file reference
		return "local_file", "", "", ast, true
	}
	return "", "", "", nil, false
}

func rewriteDuckDBSQL(ctx context.Context, props *ModelInputProperties, inputHandle drivers.Handle, path string, ast *duckdbsql.AST) (release func(), retErr error) {
	fs, ok := inputHandle.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("internal error: expected object store connector")
	}

	iter, err := fs.DownloadFiles(ctx, path)
	if err != nil {
		return nil, err
	}
	defer func() {
		// closing the iterator deletes the files
		// only delete the files if there was an error
		// the caller will call release once the files are no longer needed
		if retErr != nil {
			_ = iter.Close()
		}
	}()

	// We want to batch all the files to avoid issues with schema compatibility and partition_overwrite inserts.
	// If a user encounters performance issues, we should encourage them to use `partitions:` without `incremental:` to break ingestion into smaller batches.
	iter.SetKeepFilesUntilClose()
	var files []string
	for {
		localFiles, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		files = append(files, localFiles...)
	}
	if len(files) == 0 {
		return nil, drivers.ErrNoRows
	}

	// Rewrite the SQL
	props.SQL, err = rewriteSQL(ast, files)
	return func() { _ = iter.Close() }, err
}

func rewriteSQL(ast *duckdbsql.AST, allFiles []string) (string, error) {
	err := ast.RewriteTableRefs(func(table *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
		return &duckdbsql.TableRef{
			Paths:      allFiles,
			Function:   table.Function,
			Properties: table.Properties,
			Params:     table.Params,
		}, true
	})
	if err != nil {
		return "", err
	}
	sql, err := ast.Format()
	if err != nil {
		return "", err
	}
	return sql, nil
}

// rewriteLocalPaths rewrites a DuckDB SQL statement such that relative paths become absolute paths relative to the basePath,
// and if allowHostAccess is false, returns an error if any of the paths resolve to a path outside of the basePath.
func rewriteLocalPaths(ast *duckdbsql.AST, basePath string, allowHostAccess bool) (string, error) {
	var resolveErr error
	err := ast.RewriteTableRefs(func(t *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
		res := make([]string, 0)
		for _, p := range t.Paths {
			resolved, err := fileutil.ResolveLocalPath(p, basePath, allowHostAccess)
			if err != nil {
				resolveErr = err
				return nil, false
			}
			res = append(res, resolved)
		}
		return &duckdbsql.TableRef{
			Function:   t.Function,
			Paths:      res,
			Properties: t.Properties,
			Params:     t.Params,
		}, true
	})
	if resolveErr != nil {
		return "", resolveErr
	}
	if err != nil {
		return "", err
	}

	return ast.Format()
}
