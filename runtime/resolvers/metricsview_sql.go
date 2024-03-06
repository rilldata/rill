package resolvers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/exp/maps"
)

func init() {
	runtime.RegisterAPIResolverInitializer("Metrics", newMetricsViewSQL)
}

var (
	aggWithoutMVRegex = regexp.MustCompile(`(?i)AGGREGATE\(([a-zA-z_][a-zA-Z0-9_]*|"(?:[^"]|"")*")\)`)
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

	parsedSQL, deps, err := expandMetricsViewSQL(ctx, ctrl, opts, sql)
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

// expandMetricsViewSQL parses the metrics SQL where it
//
// 1. expands AGGREGATE(metric) into actual aggregate definition from model_view definition
// 2. converts the FROM model_view clause to the underlying FROM table
//
// example transformation:
// input : SELECT dim_col, AGGREGATE(metric) FROM metrics_view GROUP BY dim_col
// output : SELECT dim_name, sum(col2) FROM metrics_view_model  GROUP BY dim_col
func expandMetricsViewSQL(ctx context.Context, ctrl *runtime.Controller, opts *runtime.APIResolverOptions, sql string) (string, []*runtimev1.ResourceName, error) {
	// 1. get the dependencies
	meta, err := compilerv1.AnalyzeTemplate(sql)
	if err != nil {
		return "", nil, err
	}
	var deps []*runtimev1.ResourceName
	for _, ref := range meta.Refs {
		deps = append(deps, &runtimev1.ResourceName{Kind: ref.Kind.String(), Name: ref.Name})
	}

	// 2. Expand metricsview SQL
	// if there is a match, it will be of the form `from (metrics_view)``
	// first is full match second is `metrics_view`
	matches := fromMVRegex.FindAllStringSubmatch(sql, -1)

	seenMV := make(map[string]*runtimev1.ResourceName)
	var mvConnector string
	for _, match := range matches {
		metricView := unquote(match[1])
		if _, ok := seenMV[metricView]; ok {
			continue
		}

		resource, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricView}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				continue
			}
			return "", nil, fmt.Errorf("error fetching resource %v: %w", metricView, err)
		}
		if resource.GetMetricsView() == nil { // resource is not a metrics view
			continue
		}
		seenMV[metricView] = &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricView}

		if mvConnector != "" && resource.GetMetricsView().Spec.Connector != mvConnector {
			return "", nil, fmt.Errorf("all referenced metrics views must use the same connector")
		}
		mvConnector = resource.GetMetricsView().Spec.Connector

		// change from metric view to underlying table
		fromQry, err := underlyingTableQuery(ctrl.Runtime, opts, resource.GetMetricsView(), resource.Meta.StateUpdatedOn.AsTime())
		if err != nil {
			return "", nil, err
		}
		sql = strings.ReplaceAll(sql, match[0], fromQry)

		measures := resource.GetMetricsView().Spec.Measures
		nameToExprMap := make(map[string]string, len(measures))
		for _, m := range measures {
			nameToExprMap[m.Name] = m.Expression
		}

		// example query = select dim1, aggregate(mv."my measure") from mv
		// captures AGGREGATE("mv name"."my measure"), my name, measure
		aggRegex, err := regexp.Compile(fmt.Sprintf(`(?i)AGGREGATE\((%s|"%s").([a-zA-z_][a-zA-Z0-9_]*|"(?:[^"]|"")*")\)`, metricView, metricView))
		if err != nil {
			return "", nil, err
		}

		aggMatches := aggRegex.FindAllStringSubmatch(sql, -1)
		for _, aggMatch := range aggMatches {
			expr, ok := nameToExprMap[unquote(aggMatch[2])]
			if !ok {
				return "", nil, fmt.Errorf("MetricsViewSQL: measure %v not found", aggMatch[2])
			}

			// TODO handle case when two different tables have same column name in the measure expression
			sql = strings.ReplaceAll(sql, aggMatch[0], expr)
		}

		// additionally also handle the case when only one `from mv` found
		// in which case user can submit query without mv name appended to measure
		// select dim1, aggregate("my measure") from mv
		if len(matches) == 1 {
			aggMatches = aggWithoutMVRegex.FindAllStringSubmatch(sql, -1)
			for _, aggMatch := range aggMatches {
				expr, ok := nameToExprMap[unquote(aggMatch[1])]
				if !ok {
					return "", nil, fmt.Errorf("MetricsViewSQL: measure %v not found", aggMatch[1])
				}

				sql = strings.ReplaceAll(sql, aggMatch[0], expr)
			}
		}
	}
	deps = append(deps, maps.Values(seenMV)...)

	// 3. resolver all templates
	sql, err = compilerv1.ResolveTemplate(sql, compilerv1.TemplateData{
		User:       opts.UserAttributes,
		ExtraProps: opts.Args,
		Self: compilerv1.TemplateResource{
			Meta:  &runtimev1.ResourceMeta{}, // TODO: Fill in with actual metadata
			Spec:  opts.API.Spec,
			State: opts.API.State,
		},
		Resolve: func(ref compilerv1.ResourceName) (string, error) {
			return safeSQLName(ref.Name), nil
		},
		Lookup: func(name compilerv1.ResourceName) (compilerv1.TemplateResource, error) {
			// TODO: Implement this, do we need to support this?
			return compilerv1.TemplateResource{}, nil
		},
	})
	if err != nil {
		return "", nil, err
	}

	return sql, deps, nil
}

func underlyingTableQuery(rt *runtime.Runtime, opts *runtime.APIResolverOptions, mv *runtimev1.MetricsViewV2, lastUpdatedTime time.Time) (string, error) {
	security, err := rt.ResolveMetricsViewSecurity(opts.UserAttributes, opts.InstanceID, mv.Spec, lastUpdatedTime)
	if err != nil {
		return "", err
	}

	if security == nil {
		return fmt.Sprintf("FROM %s", safeSQLName(mv.Spec.Table)), nil
	}

	if !security.Access || security.ExcludeAll {
		return "", fmt.Errorf("access forbidden")
	}

	var sql string
	if len(security.Include) > 0 {
		var cols []string
		for _, col := range security.Include {
			cols = append(cols, safeSQLName(col))
		}
		sql = "SELECT " + strings.Join(cols, ", ") + " FROM " + safeSQLName(mv.Spec.Table)
	} else if len(security.Exclude) > 0 {
		var cols []string
		for _, col := range mv.Spec.Dimensions {
			if !slices.Contains(security.Exclude, col.Column) {
				cols = append(cols, safeSQLName(col.Column))
			}
		}
		sql = "SELECT " + strings.Join(cols, ", ") + " FROM " + safeSQLName(mv.Spec.Table)
	} else {
		sql = "SELECT * FROM " + safeSQLName(mv.Spec.Table)
	}

	if security.RowFilter != "" {
		sql += " WHERE " + security.RowFilter
	}
	return fmt.Sprintf("FROM (%s)", sql), nil
}

func unquote(input string) string {
	return strings.Trim(strings.ReplaceAll(input, `""`, `"`), `"`)
}
