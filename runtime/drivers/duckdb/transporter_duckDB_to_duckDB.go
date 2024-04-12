package duckdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
)

type duckDBToDuckDB struct {
	to     drivers.OLAPStore
	logger *zap.Logger
}

func NewDuckDBToDuckDB(to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &duckDBToDuckDB{
		to:     to,
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
		return t.transferFromExternalDB(ctx, srcCfg, sinkCfg)
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

	return t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, srcCfg.SQL)
}

func (t *duckDBToDuckDB) transferFromExternalDB(ctx context.Context, srcProps *dbSourceProperties, sinkProps *sinkProperties) error {
	return t.to.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, _ *sql.Conn) error {
		res, err := t.to.Execute(ctx, &drivers.Statement{Query: "SELECT current_database(),current_schema();"})
		if err != nil {
			return err
		}

		var localDB, localSchema string
		for res.Next() {
			if err := res.Scan(&localDB, &localSchema); err != nil {
				_ = res.Close()
				return err
			}
		}
		_ = res.Close()

		// duckdb considers everything before first . as db name
		// alternative solution can be to query `show databases()` before and after to identify db name
		dbName, _, _ := strings.Cut(filepath.Base(srcProps.Database), ".")
		if dbName == "main" {
			return fmt.Errorf("`main` is a reserved db name")
		}

		if err = t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ATTACH %s AS %s", safeSQLString(srcProps.Database), safeSQLName(dbName))}); err != nil {
			return fmt.Errorf("failed to attach db %q: %w", srcProps.Database, err)
		}

		defer func() {
			err := t.to.WithConnection(ensuredCtx, 100, false, true, func(wrappedCtx, ensuredCtx context.Context, conn *sql.Conn) error {
				return t.to.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("DETACH %s;", safeSQLName(dbName))})
			})
			if err != nil {
				t.logger.Error("failed to detach db", zap.Error(err))
			}
		}()

		if err := t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("USE %s;", safeName(dbName))}); err != nil {
			return err
		}

		defer func() { // revert back to localdb
			if err = t.to.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("USE %s.%s;", safeName(localDB), safeName(localSchema))}); err != nil {
				t.logger.Error("failed to switch to local database", zap.Error(err))
			}
		}()

		userQuery := strings.TrimSpace(srcProps.SQL)
		userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s.%s.%s AS (%s\n);", safeName(localDB), safeName(localSchema), safeName(sinkProps.Table), userQuery)
		return t.to.Exec(ctx, &drivers.Statement{Query: query})
	})
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
