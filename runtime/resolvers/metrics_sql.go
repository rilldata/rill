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
	"golang.org/x/exp/maps"
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
	aggregateRegex = regexp.MustCompile(fmt.Sprintf(`(?i)AGGREGATE\((?:(%s)\.)?(%s)\)`, sqlIdentifier, sqlIdentifier))

	// fromRegex is regex pattern to identify a FROM clause in SQL.
	fromRegex = regexp.MustCompile(fmt.Sprintf(`(?i)FROM\s+(%s)`, sqlIdentifier))
)

// metricsSQLCompiler parses a metrics SQL query and compiles it to a regular SQL query.
// Metrics SQL is a superset of SQL that supports querying Rill's metrics views.
// The syntax is inspired by Calcite's measure columns: https://issues.apache.org/jira/browse/CALCITE-4496.
//
// This is a simple implementation that uses regular expressions. It does not support all SQL features.
// It works by:
//
// 1. Expanding AGGREGATE(measure) into actual aggregate expressions from the metrics view definition.
// 2. Converting "FROM metrics_view" clauses to nested SELECTs on the underlying table with filters based on the metrics view's security policy.
//
// TODO: This implementation does not resolve dimension names to underlying columns/expressions. Here is an example of the desired transformation:
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
	// Expand "FROM metrics_view".
	// For each match, match[1] will contain the metrics_view identifier.
	sql := c.sql
	matches := fromRegex.FindAllStringSubmatch(sql, -1)
	mvToMeasureExprMap := make(map[string]map[string]string)
	var mvConnector string
	var refs []*runtimev1.ResourceName
	for _, match := range matches {
		metricView := unquote(match[1])
		if _, ok := mvToMeasureExprMap[metricView]; ok {
			continue
		}

		resource, err := c.ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricView}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				continue
			}
			return "", "", nil, fmt.Errorf("error fetching resource %v: %w", metricView, err)
		}

		refs = append(refs, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricView})

		if mvConnector != "" && resource.GetMetricsView().Spec.Connector != mvConnector {
			return "", "", nil, fmt.Errorf("all referenced metrics views must use the same connector")
		}
		mvConnector = resource.GetMetricsView().Spec.Connector

		// Replace "FROM metric_view" with a query to the underlying table
		fromQry, measureToExprMap, err := c.fromQueryForMetricsView(resource)
		if err != nil {
			return "", "", nil, err
		}
		mvToMeasureExprMap[metricView] = measureToExprMap
		sql = strings.ReplaceAll(sql, match[0], fromQry)
	}

	// Expand AGGREGATE expressions
	if len(mvToMeasureExprMap) > 0 {
		// Example: SELECT dim1, AGGREGATE(mv."my measure") FROM mv
		// The regex captures [AGGREGATE("mv name"."my measure"), mv name, measure]
		aggMatches := aggregateRegex.FindAllStringSubmatch(sql, -1)
		for _, aggMatch := range aggMatches {
			metricView := unquote(aggMatch[1])
			var expr string
			var found bool
			if metricView == "" {
				if len(mvToMeasureExprMap) > 1 {
					return "", "", nil, fmt.Errorf("ambiguous reference to measure %q: use a fully qualified name such as \"metrics_view.measure\"", unquote(aggMatch[2]))
				}

				expr, found = maps.Values(mvToMeasureExprMap)[0][unquote(aggMatch[2])]
			} else {
				measureToExprMap, ok := mvToMeasureExprMap[metricView]
				if !ok {
					return "", "", nil, fmt.Errorf("metric_view %q not found", metricView)
				}
				expr, found = measureToExprMap[unquote(aggMatch[2])]
			}

			if !found {
				return "", "", nil, fmt.Errorf("MetricsViewSQL: measure %q not found", aggMatch[2])
			}

			// TODO handle case when two different tables have same column name in the measure expression
			sql = strings.ReplaceAll(sql, aggMatch[0], expr)
		}
	}

	return sql, mvConnector, normalizeRefs(refs), nil
}

func (c *metricsSQLCompiler) fromQueryForMetricsView(mv *runtimev1.Resource) (string, map[string]string, error) {
	spec := mv.GetMetricsView().State.ValidSpec

	security, err := c.ctrl.Runtime.ResolveMetricsViewSecurity(c.userAttributes, c.instanceID, spec, mv.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return "", nil, err
	}

	measures := make(map[string]string, len(spec.Measures))
	for _, measure := range spec.Measures {
		measures[measure.Name] = measure.Expression
	}

	if security == nil {
		return fmt.Sprintf("FROM %s", c.dialect.EscapeIdentifier(spec.Table)), measures, nil
	}

	if !security.Access || security.ExcludeAll {
		return "", nil, fmt.Errorf("access to metrics view %q forbidden", mv.Meta.Name)
	}

	dims := make(map[string]any, len(spec.Dimensions))
	for _, dim := range spec.Dimensions {
		dims[dim.Column] = nil
	}

	finalMeasures := maps.Clone(measures)
	if len(security.Include) != 0 {
		for measure := range finalMeasures {
			if !slices.Contains(security.Include, measure) { // measures not part of include clause should not be accessible
				finalMeasures[measure] = "null"
			}
		}
	}

	for _, include := range security.Include {
		if _, ok := dims[include]; ok {
			return "", nil, fmt.Errorf("metrics SQL does not support metrics views with an include/exclude security policy that applies to dimensions")
		}
		finalMeasures[include] = measures[include]
	}

	for _, exclude := range security.Exclude {
		if _, ok := dims[exclude]; ok {
			return "", nil, fmt.Errorf("metrics SQL does not support metrics views with an include/exclude security policy that applies to dimensions")
		}
		finalMeasures[exclude] = "null"
	}

	sql := "SELECT * FROM " + c.dialect.EscapeIdentifier(spec.Table)
	if security.RowFilter != "" {
		sql += " WHERE " + security.RowFilter
	}
	return fmt.Sprintf("FROM (%s)", sql), finalMeasures, nil
}

func unquote(input string) string {
	return strings.Trim(strings.ReplaceAll(input, `""`, `"`), `"`)
}
