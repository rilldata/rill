package resolvers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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

// sqlIdenitifer is regex pattern to identify a SQL idenitifier. The idenitifier may be wrapped in double quotes.
// Additionally if double quotes are present in idenitifer, it is escaped with additional double quotes.
var sqlIdenitifer = `[a-zA-z_][a-zA-Z0-9_]*|"(?:[^"]|"")*"`

var (
	aggregateRegex = regexp.MustCompile(fmt.Sprintf(`(?i)AGGREGATE\((?:(%s)\.)?(%s)\)`, sqlIdenitifer, sqlIdenitifer))
	fromMVRegex    = regexp.MustCompile(fmt.Sprintf(`(?i)FROM\s+(%s)`, sqlIdenitifer))
)

func newMetricsViewSQL(ctx context.Context, opts *runtime.APIResolverOptions) (runtime.APIResolver, error) {
	sql := opts.API.Spec.ResolverProperties.Fields["sql"].GetStringValue()
	if sql == "" {
		return nil, errors.New("no sql query found for metrics sql resolver")
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

	// 2. Expand from metrics_view
	// if there is a match, it will be of the form `from (metrics_view)``
	// first is full match second is `metrics_view`
	matches := fromMVRegex.FindAllStringSubmatch(sql, -1)

	mvToMeasureExprMap := make(map[string]map[string]string)
	var mvConnector string
	for _, match := range matches {
		metricView := unquote(match[1])
		if _, ok := mvToMeasureExprMap[metricView]; ok {
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
		deps = append(deps, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricView})

		if mvConnector != "" && resource.GetMetricsView().Spec.Connector != mvConnector {
			return "", nil, fmt.Errorf("all referenced metrics views must use the same connector")
		}
		mvConnector = resource.GetMetricsView().Spec.Connector

		// change from metric view to underlying table
		fromQry, measureToExprMap, err := underlyingTableQuery(ctrl.Runtime, opts, resource.GetMetricsView(), resource.Meta.StateUpdatedOn.AsTime())
		if err != nil {
			return "", nil, err
		}
		mvToMeasureExprMap[metricView] = measureToExprMap
		sql = strings.ReplaceAll(sql, match[0], fromQry)
	}

	// 3. Expand aggregate expression
	if len(mvToMeasureExprMap) > 0 {
		// example query = select dim1, aggregate(mv."my measure") from mv
		// captures AGGREGATE("mv name"."my measure"), my name, measure
		aggMatches := aggregateRegex.FindAllStringSubmatch(sql, -1)
		for _, aggMatch := range aggMatches {
			metricView := unquote(aggMatch[1])
			var expr string
			var found bool
			if metricView == "" {
				if len(mvToMeasureExprMap) > 1 {
					return "", nil, fmt.Errorf("measure should be precedded with metric_view if api references more than one metric view")
				}

				expr, found = maps.Values(mvToMeasureExprMap)[0][unquote(aggMatch[2])]
			} else {
				measureToExprMap, ok := mvToMeasureExprMap[metricView]
				if !ok {
					return "", nil, fmt.Errorf("metric_view %q not found", metricView)
				}
				expr, found = measureToExprMap[unquote(aggMatch[2])]
			}

			if !found {
				return "", nil, fmt.Errorf("MetricsViewSQL: measure %q not found", aggMatch[2])
			}

			// TODO handle case when two different tables have same column name in the measure expression
			sql = strings.ReplaceAll(sql, aggMatch[0], expr)
		}
	}

	// 4. resolve all templates
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
			res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: name.Kind.String(), Name: name.Name}, false)
			if err != nil {
				return compilerv1.TemplateResource{}, err
			}

			switch name.Kind {
			case compilerv1.ResourceKindSource:
				return compilerv1.TemplateResource{
					Meta:  res.Meta,
					Spec:  res.GetSource().Spec,
					State: res.GetSource().State,
				}, nil
			case compilerv1.ResourceKindModel:
				return compilerv1.TemplateResource{
					Meta:  res.Meta,
					Spec:  res.GetModel().Spec,
					State: res.GetModel().State,
				}, nil
			case compilerv1.ResourceKindMetricsView:
				return compilerv1.TemplateResource{
					Meta:  res.Meta,
					Spec:  res.GetMetricsView().Spec,
					State: res.GetMetricsView().State,
				}, nil
			default:
				return compilerv1.TemplateResource{}, fmt.Errorf("can only lookup source, model or metrics_view")
			}
		},
	})
	if err != nil {
		return "", nil, err
	}

	return sql, deps, nil
}

func underlyingTableQuery(rt *runtime.Runtime, opts *runtime.APIResolverOptions, mv *runtimev1.MetricsViewV2, lastUpdatedTime time.Time) (string, map[string]string, error) {
	security, err := rt.ResolveMetricsViewSecurity(opts.UserAttributes, opts.InstanceID, mv.Spec, lastUpdatedTime)
	if err != nil {
		return "", nil, err
	}

	measures := make(map[string]string, len(mv.Spec.Measures))
	for _, measure := range mv.Spec.Measures {
		measures[measure.Name] = measure.Expression
	}

	if security == nil {
		return fmt.Sprintf("FROM %s", safeSQLName(mv.Spec.Table)), measures, nil
	}

	if !security.Access || security.ExcludeAll {
		return "", nil, fmt.Errorf("access forbidden")
	}

	dims := make(map[string]any, len(mv.Spec.Dimensions))
	for _, dim := range mv.Spec.Dimensions {
		dims[dim.Column] = nil
	}

	var finalMeasures map[string]string
	if len(security.Include) == 0 {
		finalMeasures = maps.Clone(measures)
	} else {
		finalMeasures = make(map[string]string)
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

	sql := "SELECT * FROM " + safeSQLName(mv.Spec.Table)
	if security.RowFilter != "" {
		sql += " WHERE " + security.RowFilter
	}
	return fmt.Sprintf("FROM (%s)", sql), finalMeasures, nil
}

func unquote(input string) string {
	return strings.Trim(strings.ReplaceAll(input, `""`, `"`), `"`)
}
