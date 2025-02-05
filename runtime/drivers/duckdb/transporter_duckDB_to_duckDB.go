package duckdb

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/url"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
	"go.uber.org/zap"
)

type duckDBToDuckDB struct {
	from   drivers.Handle
	to     *connection
	logger *zap.Logger
}

func newDuckDBToDuckDB(from drivers.Handle, c *connection, logger *zap.Logger) drivers.Transporter {
	return &duckDBToDuckDB{
		from:   from,
		to:     c,
		logger: logger,
	}
}

var _ drivers.Transporter = &duckDBToDuckDB{}

func (t *duckDBToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	var props map[string]any
	if t.from.Driver() != "duckdb" {
		// ingest from external db which can also be configured separately via a connector
		props = maps.Clone(t.from.Config())
		maps.Copy(props, srcProps)
	} else {
		props = srcProps
	}
	srcCfg, err := parseDBSourceProperties(props)
	if err != nil {
		return err
	}

	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	t.logger = t.logger.With(zap.String("source", sinkCfg.Table))

	if srcCfg.Database != "" { // query to be run against an external DB
		if t.from.Driver() == "duckdb" {
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

	return t.to.CreateTableAsSelect(ctx, sinkCfg.Table, srcCfg.SQL, &drivers.CreateTableOptions{})
}

func (t *duckDBToDuckDB) transferFromExternalDB(ctx context.Context, srcProps *dbSourceProperties, sinkProps *sinkProperties) error {
	var initSQL []string
	safeDBName := safeName(sinkProps.Table + "_external_db_")
	safeTempTable := safeName(sinkProps.Table + "__temp__")
	switch t.from.Driver() {
	case "mysql":
		dsn := rewriteMySQLDSN(srcProps.Database)
		initSQL = append(initSQL, "INSTALL 'MYSQL'; LOAD 'MYSQL';", fmt.Sprintf("ATTACH %s AS %s (TYPE mysql, READ_ONLY)", safeSQLString(dsn), safeDBName))
	case "postgres":
		initSQL = append(initSQL, "INSTALL 'POSTGRES'; LOAD 'POSTGRES';", fmt.Sprintf("ATTACH %s AS %s (TYPE postgres, READ_ONLY)", safeSQLString(srcProps.Database), safeDBName))
	case "duckdb":
		initSQL = append(initSQL, fmt.Sprintf("ATTACH %s AS %s (READ_ONLY)", safeSQLString(srcProps.Database), safeDBName))
	default:
		return fmt.Errorf("internal error: unsupported external database: %s", t.from.Driver())
	}
	beforeCreateFn := func(ctx context.Context, conn *sqlx.Conn) error {
		for _, sql := range initSQL {
			_, err := conn.ExecContext(ctx, sql)
			if err != nil {
				return err
			}
		}

		var localDB, localSchema string
		err := conn.QueryRowxContext(ctx, "SELECT current_database(),current_schema();").Scan(&localDB, &localSchema)
		if err != nil {
			return err
		}

		_, err = conn.ExecContext(ctx, fmt.Sprintf("USE %s;", safeDBName))
		if err != nil {
			return err
		}

		userQuery := strings.TrimSpace(srcProps.SQL)
		userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s.%s.%s AS (%s\n);", safeName(localDB), safeName(localSchema), safeTempTable, userQuery)
		_, err = conn.ExecContext(ctx, query)
		// first revert back to localdb
		if err != nil {
			return err
		}
		// revert to localdb and schema before returning
		_, err = conn.ExecContext(ctx, fmt.Sprintf("USE %s.%s;", safeName(localDB), safeName(localSchema)))
		return err
	}
	afterCreateFn := func(ctx context.Context, conn *sqlx.Conn) error {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", safeTempTable))
		return err
	}
	db, release, err := t.to.acquireDB()
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()
	return db.CreateTableAsSelect(ctx, sinkProps.Table, fmt.Sprintf("SELECT * FROM %s", safeTempTable), &rduckdb.CreateTableOptions{
		BeforeCreateFn: beforeCreateFn,
		AfterCreateFn:  afterCreateFn,
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

// rewriteMySQLDSN rewrites a MySQL DSN to a format that DuckDB expects.
// DuckDB does not support the URI based DSN format yet. It expects the DSN to be in the form of key=value pairs.
// This function parses the MySQL URI based DSN and converts it to the key=value format. It only converts the common parameters.
// For more advanced parameters like SSL configs, the user should manually convert the DSN to the key=value format.
// If there is an error parsing the DSN, it returns the DSN as is.
func rewriteMySQLDSN(dsn string) string {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		// If we can't parse the DSN, just return it as is. May be it is already in the form duckdb expects.
		return dsn
	}

	var sb strings.Builder

	if cfg.User != "" {
		sb.WriteString(fmt.Sprintf("user=%s ", cfg.User))
	}
	if cfg.Passwd != "" {
		sb.WriteString(fmt.Sprintf("password=%s ", cfg.Passwd))
	}
	if cfg.DBName != "" {
		sb.WriteString(fmt.Sprintf("database=%s ", cfg.DBName))
	}
	switch cfg.Net {
	case "unix":
		sb.WriteString(fmt.Sprintf("socket=%s ", cfg.Addr))
	case "tcp", "tcp6":
		host, port, err := net.SplitHostPort(cfg.Addr)
		if err != nil {
			return dsn
		}
		sb.WriteString(fmt.Sprintf("host=%s ", host))
		if port != "" {
			sb.WriteString(fmt.Sprintf("port=%s ", port))
		}
	default:
		return dsn
	}
	return sb.String()
}
