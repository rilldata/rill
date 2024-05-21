package resolvers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	metricssqlparser "github.com/rilldata/rill/runtime/pkg/metricssql"
	"go.uber.org/zap"
)

var errForbidden = errors.New("access to metrics view is forbidden")

func init() {
	runtime.RegisterResolverInitializer("builtin_pg_catalog_sql", newBuiltinInformationSchemaSQL)
	runtime.RegisterBuiltinAPI("pg-catalog-sql", "builtin_pg_catalog_sql", nil)
}

type args struct {
	SQL      string `mapstructure:"sql"`
	Priority int    `mapstructure:"priority"`
	DataDir  string `mapstructure:"temp_dir"`
}

var extraCharRe = regexp.MustCompile(`[\n\t\r]`)

// newBuiltinMetricsSQL is the resolver for the built-in /metrics-sql API.
// It executes a metrics SQL query provided dynamically through the args.
// It errors if the user identified by the attributes is not an admin.
func newBuiltinInformationSchemaSQL(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// Decode the args
	args := &args{}
	if err := mapstructure.Decode(opts.Args, args); err != nil {
		return nil, err
	}

	args.SQL = extraCharRe.ReplaceAllString(args.SQL, "\n")
	resolver, ok := parse(args.SQL)
	if ok {
		return resolver, nil
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	resources, err := ctrl.List(ctx, runtime.ResourceKindMetricsView, "", false)
	if err != nil {
		return nil, err
	}

	dbDir := filepath.Join(args.DataDir, strconv.FormatInt(time.Now().UnixMilli(), 10))
	if err := os.Mkdir(dbDir, fs.ModePerm); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dbDir, "catalog.db")
	dbName := fmt.Sprintf("pg_catalog_db_%v", rand.Int())
	args.SQL = rewriteSQL(args.SQL)
	// loop over all resources and create corresponding table in duckdb so that these can be queried with information_schema
	for _, resource := range resources {
		metricSQL, err := fromQueryForMetricsView(ctx, ctrl, opts, resource)
		if err != nil {
			if errors.Is(err, errForbidden) {
				continue
			}
			return nil, err
		}

		compiler := metricssqlparser.New(ctrl, opts.InstanceID, opts.UserAttributes, args.Priority)
		sql, connector, _, err := compiler.Compile(ctx, metricSQL)
		if err != nil {
			return nil, err
		}

		if err := populate(ctx, opts, connector, sql, dbPath, dbName, resource.Meta.Name.Name); err != nil {
			return nil, err
		}
	}

	// init the duckdb
	handle, err := drivers.Open("duckdb", opts.InstanceID, map[string]any{"path": dbPath}, activity.NewNoopClient(), zap.NewNop())
	if err != nil {
		return nil, err
	}

	olap, ok := handle.AsOLAP(opts.InstanceID)
	if !ok {
		return nil, fmt.Errorf("developer error : handle is not an OLAP driver")
	}

	if err := populateCatalogTables(ctx, olap); err != nil {
		return nil, err
	}
	fmt.Printf("final SQL: %v\n", args.SQL)
	return &catalogSQLResolver{
		olap: olap,
		sql:  args.SQL,
	}, nil
}

func populate(ctx context.Context, opts *runtime.ResolverOptions, connector, query, path, dbName, mvName string) error {
	olap, release, err := opts.Runtime.OLAP(ctx, opts.InstanceID, connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("only duckdb is supported")
	}

	err = olap.WithConnection(ctx, 1, false, false, func(ctx, ensuredCtx context.Context, _ *sql.Conn) error {
		if err := olap.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ATTACH '%s' AS %s", path, dbName)}); err != nil {
			return err
		}

		defer func() {
			_ = olap.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("DETACH %s", dbName)})
		}()

		qry := fmt.Sprintf("CREATE TABLE %s.%s AS SELECT * FROM (%s) LIMIT 0", dbName, olap.Dialect().EscapeIdentifier(mvName), query)
		return olap.Exec(ctx, &drivers.Statement{Query: qry})
	})
	return err
}

func populateCatalogTables(ctx context.Context, olap drivers.OLAPStore) error {
	return olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE TABLE catalog.pg_catalog.pg_matviews(schemaname VARCHAR, matviewname VARCHAR, matviewowner VARCHAR, tablespace VARCHAR, hasindexes BOOLEAN, ispopulated BOOLEAN, definition VARCHAR)",
	})
}

type catalogSQLResolver struct {
	olap drivers.OLAPStore
	sql  string
}

func (r *catalogSQLResolver) Close() error {
	return nil
}

func (r *catalogSQLResolver) Key() string {
	return r.sql
}

func (r *catalogSQLResolver) Refs() []*runtimev1.ResourceName {
	return nil
}

func (r *catalogSQLResolver) Validate(ctx context.Context) error {
	_, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:  r.sql,
		DryRun: true,
	})
	return err
}

func (r *catalogSQLResolver) ResolveInteractive(ctx context.Context, opts *runtime.ResolverInteractiveOptions) (*runtime.ResolverResult, error) {
	res, err := r.olap.Execute(ctx, &drivers.Statement{
		Query: r.sql,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	if opts != nil && opts.Format == runtime.GOOBJECTS {
		return r.scanAsGoObjects(res)
	}

	var out []map[string]any
	for res.Rows.Next() {
		row := make(map[string]any)
		err = res.Rows.MapScan(row)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	data, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &runtime.ResolverResult{
		Data:   data,
		Schema: res.Schema,
		Cache:  false, // never cache information schema queries
	}, nil
}

func (r *catalogSQLResolver) scanAsGoObjects(res *drivers.Result) (*runtime.ResolverResult, error) {
	var out [][]any
	for res.Rows.Next() {
		row, err := res.Rows.SliceScan()
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	return &runtime.ResolverResult{
		Rows:   out,
		Schema: res.Schema,
		Cache:  false,
	}, nil
}

func (r *catalogSQLResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return fmt.Errorf("not implemented")
}

func fromQueryForMetricsView(ctx context.Context, ctrl *runtime.Controller, opts *runtime.ResolverOptions, mv *runtimev1.Resource) (string, error) {
	spec := mv.GetMetricsView().State.ValidSpec
	if spec == nil {
		return "", fmt.Errorf("metrics view %q is not ready for querying, reconcile status: %q", mv.Meta.GetName(), mv.Meta.ReconcileStatus)
	}

	olap, release, err := ctrl.Runtime.OLAP(ctx, opts.InstanceID, spec.Connector)
	if err != nil {
		return "", err
	}
	defer release()
	dialect := olap.Dialect()

	security, err := ctrl.Runtime.ResolveMetricsViewSecurity(opts.UserAttributes, opts.InstanceID, spec, mv.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return "", err
	}

	var cols []string
	for _, measure := range spec.Measures {
		cols = append(cols, measure.Name)
	}
	for _, dim := range spec.Dimensions {
		cols = append(cols, dim.Name)
	}

	if security == nil {
		if spec.TimeDimension != "" {
			cols = append(cols, spec.TimeDimension)
		}
		return fmt.Sprintf("SELECT %s FROM %s", strings.Join(cols, ","), dialect.EscapeIdentifier(mv.Meta.Name.Name)), nil
	}

	if !security.Access || security.ExcludeAll {
		return "", errForbidden
	}

	var final []string
	if len(security.Include) != 0 {
		for _, measure := range cols {
			if slices.Contains(security.Include, measure) { // only include the included cols if include is set
				final = append(final, measure)
			}
		}
	}
	if len(final) > 0 {
		cols = final
	}

	for _, col := range cols {
		if !slices.Contains(security.Exclude, col) {
			final = append(final, col)
		}
	}

	if spec.TimeDimension != "" {
		final = append(final, spec.TimeDimension)
	}

	var sql strings.Builder
	sql.WriteString("SELECT ")
	sql.WriteString(strings.Join(final, ","))
	sql.WriteString(" FROM ")
	sql.WriteString(dialect.EscapeIdentifier(mv.Meta.Name.Name))
	if security.RowFilter != "" {
		sql.WriteString(" WHERE ")
		sql.WriteString(security.RowFilter)
	}
	return sql.String(), nil
}

var (
	functions         string         = "has_any_column_privilege|has_column_privilege|has_database_privilege|has_foreign_data_wrapper_privilege|has_function_privilege|has_language_privilege|has_parameter_privilege|has_schema_privilege|has_sequence_privilege|has_server_privilege|has_table_privilege|has_tablespace_privilege|has_type_privilege|pg_has_role"
	re                *regexp.Regexp = regexp.MustCompile(fmt.Sprintf(`pg_catalog.(%s)\(([^,]+), ([^,]+), ([^)]+)\)`, functions))
	dbRe              *regexp.Regexp = regexp.MustCompile(`pg_catalog\.(\w+)`)
	regclassRe        *regexp.Regexp = regexp.MustCompile(`'pg_class'::regclass`)
	versionRe         *regexp.Regexp = regexp.MustCompile(`pg_catalog\.version\(\)`)
	pgBackendPid      *regexp.Regexp = regexp.MustCompile(`(?:pg_catalog\.)?pg_backend_pid\([^)]*\)`)
	indexRe           *regexp.Regexp = regexp.MustCompile(`(?:pg_catalog\.)?pg_get_indexdef\([^)]*\)`)
	identifyOptionsRe *regexp.Regexp = regexp.MustCompile(`(?is)\(SELECT\s+json_build_object\([^)]*\)\s*FROM[^)]*\)\s+as\s+identity_options`)
	serialSequenceRe  *regexp.Regexp = regexp.MustCompile(`pg_catalog\.pg_get_serial_sequence\([^\)]*\)`)
)

func rewriteSQL(input string) string {
	result := re.ReplaceAllString(input, `(select pg_catalog.$1($3, $4))`)
	result = serialSequenceRe.ReplaceAllString(result, "NULL")
	result = pgBackendPid.ReplaceAllString(result, `(SELECT 1234) AS pg_backend_pid`)
	result = indexRe.ReplaceAllString(result, "NULL")
	result = versionRe.ReplaceAllString(result, `(SELECT 'PostgreSQL 10.0 (Debian 10.0-1.pgdg110+1) on aarch64-unknown-linux-gnu, compiled by gcc (Debian 10.2.1-6) 10.2.1 20210110, 64-bit') AS version`)
	result = dbRe.ReplaceAllString(result, `catalog.pg_catalog.$1`)
	result = regclassRe.ReplaceAllString(result, `(SELECT oid FROM pg_class WHERE relname = 'pg_class')`)
	result = identifyOptionsRe.ReplaceAllString(result, " NULL AS identity_options")
	return result
}

var (
	showVarRe *regexp.Regexp = regexp.MustCompile(`(?i)SHOW\s+(.+)`)
)

func parse(sql string) (runtime.Resolver, bool) {
	sql = strings.TrimSuffix(sql, ";")
	matches := showVarRe.FindStringSubmatch(sql)
	if len(matches) <= 1 {
		return nil, false
	}
	return &postgresSQLResolver{variable: matches[1]}, true
}

type postgresSQLResolver struct {
	variable string
}

// Close implements runtime.Resolver.
func (p *postgresSQLResolver) Close() error {
	return nil
}

// Key implements runtime.Resolver.
func (p *postgresSQLResolver) Key() string {
	return ""
}

// Refs implements runtime.Resolver.
func (p *postgresSQLResolver) Refs() []*runtimev1.ResourceName {
	return nil
}

// ResolveExport implements runtime.Resolver.
func (p *postgresSQLResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	panic("unimplemented")
}

// ResolveInteractive implements runtime.Resolver.
func (p *postgresSQLResolver) ResolveInteractive(ctx context.Context, opts *runtime.ResolverInteractiveOptions) (*runtime.ResolverResult, error) {
	fields := make([]*runtimev1.StructType_Field, 1)
	fields[0] = &runtimev1.StructType_Field{
		Name: p.variable,
		Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING, Nullable: false},
	}

	row := make([][]any, 1)
	row[0] = make([]any, 1)
	row[0][0] = value(p.variable)
	return &runtime.ResolverResult{
		Rows:   row,
		Schema: &runtimev1.StructType{Fields: fields},
		Cache:  false,
	}, nil
}

// Validate implements runtime.Resolver.
func (p *postgresSQLResolver) Validate(ctx context.Context) error {
	return nil
}

func value(variable string) string {
	switch strings.ToLower(variable) {
	case "standard_conforming_string":
		return "on"
	case "transaction isolation level":
		return "read committed"
	default:
		return "tbd"
	}
}

var _ runtime.Resolver = &postgresSQLResolver{}
