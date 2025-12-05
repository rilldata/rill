package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/azure"
	"github.com/rilldata/rill/runtime/drivers/gcs"
	"github.com/rilldata/rill/runtime/drivers/https"
	"github.com/rilldata/rill/runtime/drivers/s3"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/globutil"
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

var createSecretRegex = regexp.MustCompile(`(?i)\bcreate\b(?:\s+\w+)*?\s+secret\b`)

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

	// Secret already exists if:
	// 1. selfToSelfExecutor called from an explicit connector executor (e.g. model_executor_objectstore_self) that has already set InternalCreateSecretSQL,
	// 2. The user has defined create secret in pre_exec.
	secretAlreadyExists := inputProps.InternalCreateSecretSQL != "" || createSecretRegex.MatchString(inputProps.PreExec)

	var parsedPath string
	var parsedConnectorType string
	var parsedAST *duckdbsql.AST
	// Perform sql parsing only in the following cases:
	// 1. secret not exists already
	// 2. The input connector is DuckDB (other connectors which need secrets have already set it).
	if !secretAlreadyExists && opts.InputConnector == "duckdb" {
		// We donot want to parser sql but there are still some use case where we need to parse the path from SQL
		// 1. if we have gcs connector with native cred, DuckDB doesn't support native GCS credentials(google_credentials_json)
		// 2. if we have s3 connector without region,  DuckDB doesn't support automatic region detection
		// 3. if we have local path
		var ok bool
		if parsedConnectorType, parsedPath, parsedAST, ok = parseRefFromSQL(inputProps.SQL); ok {
			if parsedConnectorType == "local_file" {
				rewrittenSQL, err := rewriteLocalPaths(parsedAST, opts.Env.RepoRoot, opts.Env.AllowHostAccess)
				if err != nil {
					return nil, fmt.Errorf("invalid local path: %w", err)
				}
				inputProps.SQL = rewrittenSQL
			}
		}
	}

	connectorSecretsAvailable := make(map[string]bool)
	if !secretAlreadyExists {
		connectorsForSecrets, autoDetected := connectorsForSecrets(inputProps.CreateSecretsFromConnectors, e.c.config.CreateSecretsFromConnectors, opts.Env.Connectors)
		var createSecretSQLs, dropSecretSQLs []string
		for _, connector := range connectorsForSecrets {
			// we donnot need to pass the parsedPath we are using because of s3 region detection
			createSecretSQL, dropSecretSQL, connectorType, err := generateSecretSQL(ctx, opts, connector, parsedPath, nil)
			if err != nil {
				// Skip creating secrets when:
				// - autoDetected: user didn't explicitly configure the secret container
				// - errGCSUsesNativeCreds: DuckDB doesn't support native GCS credentials
				if autoDetected || errors.Is(err, errGCSUsesNativeCreds) {
					continue
				}

				return nil, fmt.Errorf("failed to create secret for connector %q: %w", connector, err)
			}
			connectorSecretsAvailable[connectorType] = true
			createSecretSQLs = append(createSecretSQLs, createSecretSQL)
			dropSecretSQLs = append(dropSecretSQLs, dropSecretSQL)
		}
		inputProps.InternalCreateSecretSQL = strings.Join(createSecretSQLs, "; ")
		inputProps.InternalDropSecretSQL = strings.Join(dropSecretSQLs, "; ")
	}

	if parsedConnectorType == "gcs" && !connectorSecretsAvailable["gcs"] {
		// rewrite duckdb sql with locally downloaded files
		handle, release, err := opts.Env.AcquireConnector(ctx, parsedConnectorType)
		if err != nil {
			return nil, err
		}
		defer release()
		deleteFiles, err := rewriteDuckDBSQL(ctx, inputProps, handle, parsedPath, parsedAST)
		if err != nil {
			return nil, err
		}
		defer deleteFiles()
	}

	// If host access is allowed, ensure DuckDB has fallback secrets for missing cloud providers.
	// This allows DuckDB to access s3, gcs, azure creds from env if  when no explicit connectors exist.
	if !secretAlreadyExists && opts.Env.AllowHostAccess {
		var fallbackSecrets []string
		var fallbackDrops []string

		// Helper to add a fallback secret for a connector type
		addFallbackSecret := func(connector string) {
			secretName := safeName(fmt.Sprintf("%s__%s__secret", opts.ModelName, connector))
			validation := ", VALIDATION 'none'"
			// azure does not support this property
			if connector == "azure" {
				validation = ""
			}
			fallbackSecrets = append(fallbackSecrets, fmt.Sprintf(`
			CREATE OR REPLACE TEMPORARY SECRET  %s (
			TYPE %s,
			PROVIDER credential_chain%s
			)`, secretName, connector, validation))
			fallbackDrops = append(fallbackDrops, fmt.Sprintf(`DROP SECRET IF EXISTS %s`, secretName))
		}

		// Add missing secrets individually
		if !connectorSecretsAvailable["s3"] {
			addFallbackSecret("s3")
		}
		if !connectorSecretsAvailable["gcs"] {
			addFallbackSecret("gcs")
		}
		if !connectorSecretsAvailable["azure"] {
			addFallbackSecret("azure")
		}
		if len(fallbackSecrets) > 0 {
			if inputProps.InternalCreateSecretSQL != "" {
				inputProps.InternalCreateSecretSQL += "; "
			}
			inputProps.InternalCreateSecretSQL += strings.Join(fallbackSecrets, "; ")

			if inputProps.InternalDropSecretSQL != "" {
				inputProps.InternalDropSecretSQL += "; "
			}
			inputProps.InternalDropSecretSQL += strings.Join(fallbackDrops, "; ")
		}
	}

	// Save the original PreExec so we can restore it later if the query fails
	originalPreExec := inputProps.PreExec
	if inputProps.InternalCreateSecretSQL != "" {
		if inputProps.PreExec != "" {
			inputProps.PreExec += "\n;" // adding \n if there is comment in pre_exec
		}
		inputProps.PreExec += inputProps.InternalCreateSecretSQL
	}
	duration, err := e.createOrInsertIntoDuckDB(ctx, opts, inputProps, outputProps, tableName, asView)
	if err != nil {
		// On failure, try cleaning up secrets and retry without secrets for anonymous bucket access
		if inputProps.InternalDropSecretSQL != "" && (strings.Contains(err.Error(), "HTTP 403") || strings.Contains(err.Error(), "IO Error") || strings.Contains(err.Error(), "region being set incorrectly")) {
			// Restore original PreExec, append drop secret, and retry.
			inputProps.PreExec = originalPreExec
			if inputProps.PreExec != "" {
				inputProps.PreExec += "\n;" // adding \n if there is comment in pre_exec
			}
			inputProps.PreExec += inputProps.InternalDropSecretSQL
			var anonymErr error
			duration, anonymErr = e.createOrInsertIntoDuckDB(ctx, opts, inputProps, outputProps, tableName, asView)
			if anonymErr != nil {
				// throwing the original error
				return nil, err
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

func parseRefFromSQL(sql string) (string, string, *duckdbsql.AST, bool) {
	// Parse AST
	ast, err := duckdbsql.Parse(sql)
	if err != nil {
		// If we can't parse the SQL just let duckdb run on it and give a sql parse error.
		return "", "", nil, false
	}

	// If there is a single table reference check if it is an object store reference.
	refs := ast.GetTableRefs()
	if len(refs) != 1 {
		return "", "", nil, false
	}
	ref := refs[0]
	// Parse the path as a URL (also works for local paths)
	if len(ref.Paths) == 0 {
		return "", "", nil, false
	}
	path := ref.Paths[0]
	uri, err := url.Parse(path)
	if err != nil {
		return "", "", nil, false
	}

	switch uri.Scheme {
	case "s3", "azure", "gcs", "gs", "http", "https":
		if uri.Scheme == "gs" {
			uri.Scheme = "gcs"
		}
		return uri.Scheme, path, ast, true
	}
	if uri.Scheme == "" && uri.Host == "" {
		// local file reference
		return "local_file", path, ast, true
	}
	return "", "", nil, false
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

// connectorsForSecrets returns the list of connectors to be used for secret creation.
// Priority:
// 1. If the model configuration specifies connector names, use those.
// 2. if duckdb connector configuration specifies connector names, use those
// 3. If neither is configured, automatically detect all connectors of type s3, azure, gcs, or https.
// The boolean return value is true if the list of connectors was automatically detected.
func connectorsForSecrets(modelSecrets, duckdbSecrets []string, allConnectors []*runtimev1.Connector) ([]string, bool) {
	var configuredConnectorsForSecrets []string
	if len(modelSecrets) > 0 {
		configuredConnectorsForSecrets = append(configuredConnectorsForSecrets, modelSecrets...)
	} else if len(duckdbSecrets) > 0 {
		configuredConnectorsForSecrets = append(configuredConnectorsForSecrets, duckdbSecrets...)
	}

	// If no connectors are configured, automatically detect all connectors of type s3, azure, gcs, or https from the project.
	// If a single configured value contains a comma-separated list of connector names, split it into individual entries.
	// Otherwise, return the explicitly configured list of connectors.
	if len(configuredConnectorsForSecrets) == 0 {
		var res []string
		for _, c := range allConnectors {
			if c.Type == "s3" || c.Type == "azure" || c.Type == "gcs" || c.Type == "https" {
				res = append(res, c.Name)
			}
		}
		return res, true
	} else if len(configuredConnectorsForSecrets) == 1 && strings.Contains(configuredConnectorsForSecrets[0], ",") {
		res := strings.Split(configuredConnectorsForSecrets[0], ",")
		for i, s := range res {
			res[i] = strings.TrimSpace(s)
		}
		return res, false
	}
	return configuredConnectorsForSecrets, false
}

func generateSecretSQL(ctx context.Context, opts *drivers.ModelExecuteOptions, connector, optionalBucketURL string, optionalAdditionalConfig map[string]any) (string, string, string, error) {
	handle, release, err := opts.Env.AcquireConnector(ctx, connector)
	if err != nil {
		return "", "", "", err
	}
	defer release()

	safeSecretName := safeName(fmt.Sprintf("%s__%s__secret", opts.ModelName, connector))
	dropSecretSQL := fmt.Sprintf("DROP SECRET IF EXISTS %s", safeSecretName)
	connectorType := handle.Driver()

	switch connectorType {
	case "s3":
		conn, ok := handle.(*s3.Connection)
		if !ok {
			return "", "", "", fmt.Errorf("internal error: expected s3 connector handle")
		}
		s3Config := conn.ParsedConfig()
		err := mapstructure.WeakDecode(optionalAdditionalConfig, s3Config)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse s3 config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE S3")

		if s3Config.AccessKeyID != "" {
			fmt.Fprintf(&sb, ", KEY_ID %s, SECRET %s", safeSQLString(s3Config.AccessKeyID), safeSQLString(s3Config.SecretAccessKey))
		} else if s3Config.AllowHostAccess {
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN, VALIDATION 'none'")
		}

		if s3Config.SessionToken != "" {
			fmt.Fprintf(&sb, ", SESSION_TOKEN %s", safeSQLString(s3Config.SessionToken))
		}
		if s3Config.Endpoint != "" {
			uri, err := url.Parse(s3Config.Endpoint)
			if err == nil && uri.Scheme != "" { // let duckdb raise an error if the endpoint is invalid
				// for duckdb the endpoint should not have a scheme
				s3Config.Endpoint = strings.TrimPrefix(s3Config.Endpoint, uri.Scheme+"://")
				if uri.Scheme == "http" {
					sb.WriteString(", USE_SSL false")
				}
			}
			sb.WriteString(", ENDPOINT ")
			sb.WriteString(safeSQLString(s3Config.Endpoint))
			sb.WriteString(", URL_STYLE path")
		}
		if s3Config.Region != "" {
			sb.WriteString(", REGION ")
			sb.WriteString(safeSQLString(s3Config.Region))
		} else if optionalBucketURL != "" {
			// DuckDB does not automatically resolve the region as of 1.2.0 so we try to detect and set the region.
			uri, err := globutil.ParseBucketURL(optionalBucketURL)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to parse path %q: %w", optionalBucketURL, err)
			}
			reg, err := s3.BucketRegion(ctx, s3Config, uri.Host)
			if err != nil {
				return "", "", "", err
			}
			sb.WriteString(", REGION ")
			sb.WriteString(safeSQLString(reg))
		}
		writeScope(&sb, s3Config.PathPrefixes)
		sb.WriteRune(')')
		return sb.String(), dropSecretSQL, connectorType, nil
	case "gcs":
		// GCS works via S3 compatibility mode.
		// This means we that gcsConfig.KeyID and gcsConfig.Secret should be set instead of gcsConfig.SecretJSON.
		gcsConnectorProp := handle.Config()
		gcsConfig, err := gcs.NewConfigProperties(gcsConnectorProp)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to load gcs base config: %w", err)
		}
		if err := mapstructure.WeakDecode(optionalAdditionalConfig, gcsConfig); err != nil {
			return "", "", "", fmt.Errorf("failed to parse gcs config properties: %w", err)
		}
		// If no credentials are provided we assume that the user wants to use the native credentials
		if gcsConfig.SecretJSON != "" || (gcsConfig.KeyID == "" && gcsConfig.Secret == "") {
			return "", "", "", errGCSUsesNativeCreds
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE GCS")
		if gcsConfig.KeyID != "" {
			fmt.Fprintf(&sb, ", KEY_ID %s, SECRET %s", safeSQLString(gcsConfig.KeyID), safeSQLString(gcsConfig.Secret))
		} else if gcsConfig.AllowHostAccess {
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN, VALIDATION 'none'")
		}
		writeScope(&sb, gcsConfig.PathPrefixes)
		sb.WriteRune(')')
		return sb.String(), dropSecretSQL, connectorType, nil
	case "azure":
		conn, ok := handle.(*azure.Connection)
		if !ok {
			return "", "", "", fmt.Errorf("internal error: expected azure connector handle")
		}
		azureConfig := conn.ParsedConfig()
		err := mapstructure.WeakDecode(optionalAdditionalConfig, azureConfig)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse azure config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE AZURE")
		// if connection string is set then use that and fall back to env credentials only if host access is allowed and connection string is not set
		connectionString := azureConfig.GetConnectionString()
		if connectionString != "" {
			fmt.Fprintf(&sb, ", CONNECTION_STRING %s", safeSQLString(connectionString))
		} else if azureConfig.AllowHostAccess {
			// duckdb will use default defaultazurecredential https://github.com/Azure/azure-sdk-for-cpp/blob/azure-identity_1.6.0/sdk/identity/azure-identity/README.md#defaultazurecredential
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN")
		}
		if azureConfig.GetAccount() != "" {
			fmt.Fprintf(&sb, ", ACCOUNT_NAME %s", safeSQLString(azureConfig.GetAccount()))
		}
		writeScope(&sb, azureConfig.PathPrefixes)
		sb.WriteRune(')')
		return sb.String(), dropSecretSQL, connectorType, nil
	case "https":
		httpConfig, err := https.NewConfigProperties(handle.Config())
		if err != nil {
			return "", "", "", fmt.Errorf("failed to load http connector properties: %w", err)
		}
		if err := mapstructure.WeakDecode(optionalAdditionalConfig, httpConfig); err != nil {
			return "", "", "", fmt.Errorf("failed to parse http model properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE HTTP")
		if len(httpConfig.Headers) > 0 {
			var headerStrings []string
			for key, value := range httpConfig.Headers {
				headerStrings = append(headerStrings, fmt.Sprintf("%s : %s", safeSQLString(key), safeSQLString(value)))
			}
			fmt.Fprintf(&sb, ", EXTRA_HTTP_HEADERS MAP { %s } ", strings.Join(headerStrings, ", "))
		}
		writeScope(&sb, httpConfig.PathPrefixes)
		sb.WriteRune(')')
		return sb.String(), dropSecretSQL, connectorType, nil
	default:
		return "", "", "", fmt.Errorf("internal error: secret generation is not supported for connector %q", handle.Driver())
	}
}

func writeScope(sb *strings.Builder, prefixes []string) {
	if len(prefixes) == 0 {
		return
	}
	sb.WriteString(", SCOPE [")
	for i, p := range prefixes {
		sb.WriteString(safeSQLString(p))
		if i < len(prefixes)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
}
