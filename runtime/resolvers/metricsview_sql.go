package resolvers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	runtime.RegisterAPIResolverInitializer("Metrics", newMetricsViewSQL)
}

var (
	aggRegexWithoutMV = regexp.MustCompile(`(?i)AGGREGATE\(([a-zA-z_][a-zA-Z0-9_]*|"(?:[^"]|"")*")\)`)
	fromMVRegex       = regexp.MustCompile(`(?i)FROM\s+([a-zA-z_][a-zA-Z0-9_]*|"(?:[^"]|"")*")`)
)

func newMetricsViewSQL(ctx context.Context, opts *runtime.APIResolverOptions) (runtime.APIResolver, error) {
	sql := opts.API.Spec.ResolverProperties.Fields["sql"].GetStringValue()
	if sql == "" {
		return nil, errors.New("no sql query found for sql resolver")
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	resolvedSQL, deps, err := resolveSQLAndDeps(ctx, sql, opts)
	if err != nil {
		return nil, err
	}

	parsedSQL, err := parseMetricsViewSQL(ctx, ctrl, resolvedSQL)
	if err != nil {
		return nil, err
	}

	olap, release, err := opts.Runtime.OLAP(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	// once we have the resolved SQL we can directly use SQLResolver
	return &SQLResolver{
		resolvedSQL: parsedSQL,
		deps:        deps,
		olap:        olap,
		releaseFunc: release,
	}, nil
}

// parseMetricsViewSQL parses the metrics SQL where it
//
// 1. expands AGGREGATE(metric) into actual aggregate definition from model_view definition
// 2. converts the FROM model_view clause to the underlying FROM table
//
// example transformation:
// input : SELECT dim_col, AGGREGATE(metric) FROM metrics_view GROUP BY dim_col
// output : SELECT dim_name, sum(col2) FROM metrics_view_model  GROUP BY dim_col
func parseMetricsViewSQL(ctx context.Context, ctrl *runtime.Controller, sql string) (string, error) {
	// if there is a match, it will be of the form `from (metrics_view)``
	// first is full match second is `metrics_view`
	matches := fromMVRegex.FindAllStringSubmatch(sql, -1)

	seenMV := make(map[string]any)
	var mvConnector string
	for _, match := range matches {
		metricView := unquote(match[1])
		if _, ok := seenMV[metricView]; ok {
			continue
		}
		seenMV[metricView] = nil

		resource, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricView}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				continue
			}
			return "", fmt.Errorf("error fetching resource %v: %w", metricView, err)
		}
		if resource.GetMetricsView() == nil { // resource is not a metrics view
			continue
		}

		if mvConnector != "" && resource.GetMetricsView().Spec.Connector != mvConnector {
			return "", fmt.Errorf("all referenced metrics views must use the same connector")
		}
		mvConnector = resource.GetMetricsView().Spec.Connector

		// change from metric view to underlying table
		sql = strings.ReplaceAll(sql, match[0], fmt.Sprintf("FROM %s", safeSQLName(resource.GetMetricsView().Spec.Table)))

		measures := resource.GetMetricsView().Spec.Measures
		nameToExprMap := make(map[string]string, len(measures))
		for _, m := range measures {
			nameToExprMap[m.Name] = m.Expression
		}

		// example query = select dim1, aggregate(mv."my measure") from mv
		aggRegex, err := regexp.Compile(fmt.Sprintf(`(?i)AGGREGATE\((%s|"%s").([a-zA-z_][a-zA-Z0-9_]*|"(?:[^"]|"")*")\)`, metricView, metricView))
		if err != nil {
			return "", err
		}

		// captures AGGREGATE("mv name"."my measure"), my name, measure
		aggMatches := aggRegex.FindAllStringSubmatch(sql, -1)
		for _, aggMatch := range aggMatches {
			expr, ok := nameToExprMap[unquote(aggMatch[2])]
			if !ok {
				return "", fmt.Errorf("MetricsViewSQL: measure %v not found", aggMatch[2])
			}

			// TODO handle case when two different tables have same column name in the measure expression
			sql = strings.ReplaceAll(sql, aggMatch[0], expr)
		}

		// additionally also handle the case when only one `from mv` found
		// in which case user can submit query without mv name appended to measure
		// select dim1, aggregate("my measure") from mv
		if len(matches) == 1 {
			aggMatches = aggRegexWithoutMV.FindAllStringSubmatch(sql, -1)
			for _, aggMatch := range aggMatches {
				expr, ok := nameToExprMap[unquote(aggMatch[1])]
				if !ok {
					return "", fmt.Errorf("MetricsViewSQL: measure %v not found", aggMatch[1])
				}

				sql = strings.ReplaceAll(sql, aggMatch[0], expr)
			}
		}
	}
	return sql, nil
}

func unquote(input string) string {
	return strings.Trim(strings.ReplaceAll(input, `""`, `"`), `"`)
}
