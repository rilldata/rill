package metricsview

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type Executor struct {
	rt          *runtime.Runtime
	instanceID  string
	metricsView *runtimev1.MetricsViewSpec
	security    *runtime.ResolvedMetricsViewSecurity
	priority    int

	olap        drivers.OLAPStore
	olapRelease func()
	instanceCfg drivers.InstanceConfig

	watermark time.Time
}

func NewExecutor(ctx context.Context, rt *runtime.Runtime, instanceID string, mv *runtimev1.MetricsViewSpec, sec *runtime.ResolvedMetricsViewSecurity, priority int) (*Executor, error) {
	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connector for metrics view: %w", err)
	}

	instanceCfg, err := rt.InstanceConfig(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	return &Executor{
		rt:          rt,
		instanceID:  instanceID,
		metricsView: mv,
		security:    sec,
		priority:    priority,
		olap:        olap,
		olapRelease: release,
		instanceCfg: instanceCfg,
	}, nil
}

func (e *Executor) Close() {
	e.olapRelease()
}

func (e *Executor) Query(ctx context.Context, qry *Query, executionTime *time.Time) (*drivers.Result, bool, error) {
	if e.security != nil && !e.security.Access {
		return nil, false, runtime.ErrForbidden
	}

	if executionTime != nil {
		e.watermark = *executionTime
	}

	err := e.rewriteQueryTimeRanges(ctx, qry)
	if err != nil {
		return nil, false, err
	}

	ast, err := NewAST(e.metricsView, e.security, qry, e.olap.Dialect())
	if err != nil {
		return nil, false, err
	}

	sql, args, err := ast.SQL()
	if err != nil {
		return nil, false, err
	}

	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Args:     args,
		Priority: e.priority,
	})
	if err != nil {
		return nil, false, err
	}

	// TODO: Get from OLAP instead of hardcoding
	cache := e.olap.Dialect() == drivers.DialectDuckDB

	return res, cache, nil
}

func (e *Executor) Watermark(ctx context.Context) (time.Time, error) {
	return e.resolveWatermark(ctx)
}

func (e *Executor) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	// TODO: Implement it
	panic("not implemented")
}

func (e *Executor) ValidateMetricsView(ctx context.Context) error {
	// TODO: Implement it
	panic("not implemented")
}

func (e *Executor) ValidateQuery(qry *Query) error {
	// TODO: Implement it
	panic("not implemented")
}
