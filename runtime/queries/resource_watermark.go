package queries

import (
	"context"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ResourceWatermark struct {
	ResourceKind string     `json:"resource_kind,omitempty"`
	ResourceName string     `json:"resource_name,omitempty"`
	Result       *time.Time `json:"-"`
}

var _ runtime.Query = &ResourceWatermark{}

func (q *ResourceWatermark) Key() string {
	return fmt.Sprintf("ResourceWatermark:%s/%s", q.ResourceKind, q.ResourceName)
}

func (q *ResourceWatermark) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: q.ResourceKind, Name: q.ResourceName},
	}
}

func (q *ResourceWatermark) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: 24,
	}
}

func (q *ResourceWatermark) UnmarshalResult(v any) error {
	res, ok := v.(*time.Time)
	if !ok {
		return fmt.Errorf("ResourceWatermark: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ResourceWatermark) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return err
	}

	rs, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: q.ResourceKind, Name: q.ResourceName}, false)
	if err != nil {
		return err
	}

	switch q.ResourceKind {
	case runtime.ResourceKindMetricsView:
		return q.resolveMetricsView(ctx, rt, instanceID, priority, rs)
	default:
		// For resources without watermark support, q.Result will be nil.
		return nil
	}
}

func (q *ResourceWatermark) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}

func (q *ResourceWatermark) resolveMetricsView(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int, rs *runtimev1.Resource) error {
	mv := rs.GetMetricsView()
	if mv == nil {
		return fmt.Errorf("internal: resource %q is not a metrics view", rs.Meta.Name.Name)
	}

	spec := mv.State.ValidSpec
	if spec == nil {
		return fmt.Errorf("metrics view %q is not valid", rs.Meta.Name.Name)
	}

	sql := ""
	if spec.WatermarkExpression != "" {
		sql = fmt.Sprintf("SELECT %s FROM %s", spec.WatermarkExpression, safeName(spec.Table))
	} else if spec.TimeDimension != "" {
		sql = fmt.Sprintf("SELECT MAX(%s) FROM %s", safeName(spec.TimeDimension), safeName(spec.Table))
	} else {
		// No watermark available
		return nil
	}

	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	res, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer res.Close()

	var t time.Time
	for res.Next() {
		if err := res.Scan(&t); err != nil {
			return err
		}
	}

	if !t.IsZero() {
		q.Result = &t
	}

	return nil
}
