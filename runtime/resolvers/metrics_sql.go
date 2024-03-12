package resolvers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	runtime.RegisterResolverInitializer("MetricsSQL", newMetricsSQL)
}

type metricsSQLProps struct {
	SQL string `mapstructure:"sql"`
}

// newMetricsSQL creates a resolver for evaluating metrics SQL.
// It wraps the regular SQL resolver and compiles the metrics SQL to a regular SQL query first.
// The compiler preserves templating in the SQL, allowing the regular SQL resolver to handle SQL templating rules.
func newMetricsSQL(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &metricsSQLProps{}
	if err := mapstructure.Decode(opts.Properties, props); err != nil {
		return nil, err
	}

	if props.SQL == "" {
		return nil, errors.New(`metrics SQL: missing required property "sql"`)
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	olap, release, err := opts.Runtime.OLAP(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	compiler := &metricsSQLCompiler{
		instanceID:     opts.InstanceID,
		ctrl:           ctrl,
		sql:            props.SQL,
		dialect:        olap.Dialect(),
		args:           opts.Args,
		userAttributes: opts.UserAttributes,
	}

	sql, connector, refs, err := compiler.compile(ctx)
	if err != nil {
		release()
		return nil, err
	}

	// Build the options for the regular SQL resolver
	sqlResolverOpts := &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: map[string]any{
			"connector": connector,
			"sql":       sql,
		},
		Args:           opts.Args,
		UserAttributes: opts.UserAttributes,
		ForExport:      opts.ForExport,
	}

	return newSQLWithRefs(ctx, sqlResolverOpts, refs)
}

var (
	// sqlIdentifier is regex pattern to identify a SQL identifier. The identifier may be wrapped in double quotes.
	// Additionally if double quotes are present in identifier, it is escaped with additional double quotes.
	sqlIdentifier = `[a-zA-z_][a-zA-Z0-9_]*|"(?:[^"]|"")*"`

	// aggregateRegex is regex pattern to identify an AGGREGATE function in SQL.
	aggregateRegex = regexp.MustCompile(fmt.Sprintf(`(?i)AGGREGATE\(\s*(%s)\s*\)$`, sqlIdentifier))

	// sqlRegex identifies a SQL query of the form: "SELECT... FROM table...".
	sqlRegex = regexp.MustCompile(fmt.Sprintf(`(?i)SELECT\s+((?:.|\n)*)\s+FROM\s+(%s)((?:.|\n)*)`, sqlIdentifier))
)

// metricsSQLCompiler parses a metrics SQL query and compiles it to a regular SQL query.
// Metrics SQL is a superset of SQL that supports querying Rill's metrics views.
// The syntax is inspired by Calcite's measure columns: https://issues.apache.org/jira/browse/CALCITE-4496.
//
// This is a simple implementation that uses regular expressions. It does not support all SQL features.
// It works by:
//
// 1. Expanding AGGREGATE(measure) into actual aggregate expressions from the metrics view definition.
// 2. Expanding dimension into dimension expression or underlying column name.
// 3. Converting "FROM metrics_view" clauses to nested SELECTs on the underlying table with filters based on the metrics view's security policy.
//
// - Input: SELECT dim1, AGGREGATE(measure1) FROM metrics_view GROUP BY dim1
// - Output: SELECT col1 AS dim1, SUM(col2) AS measure1 FROM underlying_model GROUP BY dim1
type metricsSQLCompiler struct {
	instanceID     string
	ctrl           *runtime.Controller
	sql            string
	dialect        drivers.Dialect
	args           map[string]any
	userAttributes map[string]any
}

// compile compiles the metrics SQL to a regular SQL query. It maintains template tags in the SQL.
// It returns the compiled SQL, the connector to use, and the refs to metrics views.
// It does not return other refs (like sources or models). The regular SQL resolver will handle those.
func (c *metricsSQLCompiler) compile(ctx context.Context) (string, string, []*runtimev1.ResourceName, error) {
	sql := strings.TrimSpace(c.sql)
	matches := sqlRegex.FindAllStringSubmatch(sql, -1)
	if len(matches) != 1 {
		// TODO Add a doc link for the supported syntax
		return "", "", nil, fmt.Errorf("invalid metrics_sql: %q", sql)
	}

	metricView := unquote(strings.TrimSpace(matches[0][2]))
	resource, err := c.ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricView}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return "", "", nil, fmt.Errorf("metrics_view %q not found. Metric_sql can only target metrics_view, use sql for other user cases", metricView)
		}
		return "", "", nil, fmt.Errorf("error fetching resource %v: %w", metricView, err)
	}

	fromSQL, dimensions, measures, err := c.fromQueryForMetricsView(resource)
	if err != nil {
		return "", "", nil, err
	}

	selectList := strings.Split(matches[0][1], ",")
	resolvedSelectList := make([]string, len(selectList))
	for i, col := range selectList {
		col = strings.Trim(strings.TrimSpace(col), "\n\t")
		aggMatch := aggregateRegex.FindAllStringSubmatch(col, -1)
		if len(aggMatch) > 0 {
			col = strings.TrimSpace(aggMatch[0][1])
			expr, ok := measures[unquote(col)]
			if !ok {
				return "", "", nil, fmt.Errorf("aggregate column %q must be a measure in metrics_view %q", col, metricView)
			}
			resolvedSelectList[i] = fmt.Sprintf("%s AS %s", expr, col)
		} else {
			expr, ok := dimensions[unquote(col)]
			if !ok {
				return "", "", nil, fmt.Errorf("non aggregate column %q must be a dimension in metrics_view %q", col, metricView)
			}
			resolvedSelectList[i] = fmt.Sprintf("%s AS %s", expr, col)
		}
	}

	sql = fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(resolvedSelectList, ", "), fromSQL, matches[0][3])
	return sql, resource.GetMetricsView().State.ValidSpec.Connector, []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: metricView}}, nil
}

func (c *metricsSQLCompiler) fromQueryForMetricsView(mv *runtimev1.Resource) (string, map[string]string, map[string]string, error) {
	spec := mv.GetMetricsView().State.ValidSpec

	security, err := c.ctrl.Runtime.ResolveMetricsViewSecurity(c.userAttributes, c.instanceID, spec, mv.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return "", nil, nil, err
	}

	measures := make(map[string]string, len(spec.Measures))
	for _, measure := range spec.Measures {
		measures[measure.Name] = measure.Expression
	}

	dimensions := make(map[string]string, len(spec.Dimensions))
	for _, dim := range spec.Dimensions {
		if dim.Expression != "" {
			dimensions[dim.Name] = dim.Expression
		} else {
			dimensions[dim.Name] = c.dialect.EscapeIdentifier(dim.Column)
		}
	}

	if security == nil {
		return c.dialect.EscapeIdentifier(spec.Table), dimensions, measures, nil
	}

	if !security.Access || security.ExcludeAll {
		return "", nil, nil, fmt.Errorf("access to metrics view %q forbidden", mv.Meta.Name.Name)
	}

	if len(security.Include) != 0 {
		for measure := range measures {
			if !slices.Contains(security.Include, measure) { // measures not part of include clause should not be accessible
				measures[measure] = "null"
			}
		}

		for dimension := range dimensions {
			if !slices.Contains(security.Include, dimension) { // dimensions not part of include clause should not be accessible
				dimensions[dimension] = "null"
			}
		}
	}

	for _, exclude := range security.Exclude {
		if _, ok := dimensions[exclude]; ok {
			dimensions[exclude] = "null"
		} else {
			measures[exclude] = "null"
		}
	}

	sql := "SELECT * FROM " + c.dialect.EscapeIdentifier(spec.Table)
	if security.RowFilter != "" {
		sql += " WHERE " + security.RowFilter
	}
	return fmt.Sprintf("(%s)", sql), dimensions, measures, nil
}

func unquote(input string) string {
	return strings.Trim(strings.ReplaceAll(input, `""`, `"`), `"`)
}
