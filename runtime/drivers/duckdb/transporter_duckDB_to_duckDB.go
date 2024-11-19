package duckdb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
)

type duckDBToDuckDB struct {
	to     *connection
	logger *zap.Logger
}

func newDuckDBToDuckDB(c *connection, logger *zap.Logger) drivers.Transporter {
	return &duckDBToDuckDB{
		to:     c,
		logger: logger,
	}
}

var _ drivers.Transporter = &duckDBToDuckDB{}

func (t *duckDBToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	srcCfg, err := parseDBSourceProperties(srcProps)
	if err != nil {
		return err
	}

	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	t.logger = t.logger.With(zap.String("source", sinkCfg.Table))

	if srcCfg.Database != "" { // query to be run against an external DB
		if !strings.HasPrefix(srcCfg.Database, "md:") {
			srcCfg.Database, err = fileutil.ResolveLocalPath(srcCfg.Database, opts.RepoRoot, opts.AllowHostAccess)
			if err != nil {
				return err
			}
		}
		// return t.transferFromExternalDB(ctx, srcCfg, sinkCfg)
	}

	// We can't just pass the SQL statement to DuckDB outright.
	// We need to do some rewriting for certain table references (currently object stores and local files).

	// Parse AST
	ast, err := duckdbsql.Parse(srcCfg.SQL)
	if err != nil {
		return fmt.Errorf("failed to parse sql: %w", err)
	}

	// Error if there isn't exactly one table reference
	refs := ast.GetTableRefs()
	if len(refs) != 1 {
		return errors.New("sql sources should have exactly one table reference")
	}
	ref := refs[0]
	if len(ref.Paths) == 0 {
		return errors.New("only read_* functions with a single path is supported")
	}
	if len(ref.Paths) > 1 {
		return errors.New("invalid source, only a single path for source is supported")
	}

	// Parse the path as a URL (also works for local paths)
	uri, err := url.Parse(ref.Paths[0])
	if err != nil {
		return fmt.Errorf("could not parse table function path %q: %w", ref.Paths[0], err)
	}

	// If the path is an object store reference, rewrite to objectStoreToDuckDB transporter.
	// TODO: This is pretty hacky and we should ideally break the relevant object store functionality out into a util function that we can use here.
	// (Or consider rethinking how object store connectors work in general.)
	if uri.Scheme == "s3" || uri.Scheme == "gs" || uri.Scheme == "azure" {
		if uri.Scheme == "gs" {
			uri.Scheme = "gcs"
		}

		conn, release, err := opts.AcquireConnector(uri.Scheme)
		if err != nil {
			return fmt.Errorf("sql references %q, but not able to acquire connector: %w", uri.Scheme, err)
		}
		defer release()

		objStore, ok := conn.AsObjectStore()
		if !ok {
			return fmt.Errorf("expected connector %q to implement ObjectStore", uri.Scheme)
		}

		srcProps["path"] = ref.Paths[0]
		return NewObjectStoreToDuckDB(objStore, t.to, t.logger).Transfer(ctx, srcProps, sinkProps, opts)
	}

	// If the path is a local file reference, rewrite to a safe and repo-relative path.
	if uri.Scheme == "" && uri.Host == "" {
		rewrittenSQL, err := rewriteLocalPaths(ast, opts.RepoRoot, opts.AllowHostAccess)
		if err != nil {
			return fmt.Errorf("invalid local path: %w", err)
		}
		srcCfg.SQL = rewrittenSQL
	}

	return t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, srcCfg.SQL, nil)
}

// func (t *duckDBToDuckDB) transferFromExternalDB(ctx context.Context, srcProps *dbSourceProperties, sinkProps *sinkProperties) error {
// 	t.to.db.CreateTableAsSelect(ctx, sinkProps.Table, )

// 	var localDB, localSchema string
// 	err = conn.QueryRowContext(ctx, "SELECT current_database(),current_schema()").Scan(&localDB, &localSchema)
// 	if err != nil {
// 		return err
// 	}

// 	// duckdb considers everything before first . as db name
// 	// alternative solution can be to query `show databases()` before and after to identify db name
// 	dbName, _, _ := strings.Cut(filepath.Base(srcProps.Database), ".")
// 	if dbName == "main" {
// 		return fmt.Errorf("`main` is a reserved db name")
// 	}

// 	if _, err = conn.ExecContext(ctx, fmt.Sprintf("ATTACH %s AS %s", safeSQLString(srcProps.Database), safeSQLName(dbName))); err != nil {
// 		return fmt.Errorf("failed to attach db %q: %w", srcProps.Database, err)
// 	}

// 	defer func() {
// 		_, err = conn.ExecContext(context.Background(), fmt.Sprintf("DETACH %s", safeSQLName(dbName)))
// 	}()

// 	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE %s;", safeName(dbName))); err != nil {
// 		return err
// 	}

// 	defer func() {
// 		_, err = conn.ExecContext(context.Background(), fmt.Sprintf("USE %s.%s;", safeName(localDB), safeName(localSchema)))
// 		if err != nil {
// 			t.logger.Error("failed to switch back to original database", zap.Error(err))
// 		}
// 	}()

// 	userQuery := strings.TrimSpace(srcProps.SQL)
// 	userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
// 	safeTempTable := safeName(fmt.Sprintf("%s_tmp_", sinkProps.Table))
// 	defer func() {
// 		// ensure temporary table is cleaned
// 		_, err := conn.ExecContext(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", safeTempTable))
// 		if err != nil {
// 			t.logger.Error("failed to drop temp table", zap.String("table", safeTempTable), zap.Error(err))
// 		}
// 	}()

// 	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s.%s.%s AS (%s\n);", safeName(localDB), safeName(localSchema), safeTempTable, userQuery)
// 	_, err = conn.ExecContext(ctx, query)
// 	// first revert to original database
// 	if _, switchErr := conn.ExecContext(context.Background(), fmt.Sprintf("USE %s.%s;", safeName(localDB), safeName(localSchema))); switchErr != nil {
// 		t.to.fatalInternalError(fmt.Errorf("failed to switch back to original database: %w", err))
// 	}
// 	// check for the original error
// 	if err != nil {
// 		return fmt.Errorf("failed to create table: %w", err)
// 	}

// 	// create permanent table from temp table using crud API
// 	return rwConn.CreateTableAsSelect(ctx, sinkProps.Table, fmt.Sprintf("SELECT * FROM %s", safeTempTable), nil)
// }

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
