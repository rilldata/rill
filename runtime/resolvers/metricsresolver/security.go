package metricsresolver

// import (
// 	"context"
// 	"time"

// 	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
// 	"github.com/rilldata/rill/runtime"
// 	"github.com/rilldata/rill/runtime/queries"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// // resolveMVAndSecurityFromAttributes resolves the metrics view and security policy from the attributes
// func resolveMVAndSecurityFromAttributes(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string, attrs map[string]any, dims []*runtimev1.MetricsViewAggregationDimension, measures []*runtimev1.MetricsViewAggregationMeasure) (*runtimev1.MetricsViewSpec, *runtime.ResolvedMetricsViewSecurity, error) {
// 	mv, lastUpdatedOn, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	resolvedSecurity, err := rt.ResolveMetricsViewSecurity(attrs, instanceID, mv, lastUpdatedOn)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	if resolvedSecurity != nil && !resolvedSecurity.Access {
// 		return nil, nil, queries.ErrForbidden
// 	}

// 	for _, dim := range dims {
// 		if dim.Name == mv.TimeDimension {
// 			// checkFieldAccess doesn't currently check the time dimension
// 			continue
// 		}
// 		if !checkFieldAccess(dim.Name, resolvedSecurity) {
// 			return nil, nil, queries.ErrForbidden
// 		}
// 	}

// 	for _, m := range measures {
// 		if m.BuiltinMeasure != runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED {
// 			continue
// 		}
// 		if !checkFieldAccess(m.Name, resolvedSecurity) {
// 			return nil, nil, queries.ErrForbidden
// 		}
// 	}

// 	return mv, resolvedSecurity, nil
// }
