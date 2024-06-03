package metricsview

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
)

// resolveWatermark resolves the metric view's watermark expression.
// If the resolved time is zero, it defaults to the current time.
func (e *Executor) resolveWatermark(ctx context.Context) (time.Time, error) {
	if !e.watermark.IsZero() {
		return e.watermark, nil
	}

	dialect := e.olap.Dialect()

	var expr string
	if e.metricsView.WatermarkExpression != "" {
		expr = e.metricsView.WatermarkExpression
	} else if e.metricsView.TimeDimension != "" {
		expr = fmt.Sprintf("MAX(%s)", dialect.EscapeIdentifier(e.metricsView.TimeDimension))
	} else {
		return time.Time{}, errors.New("cannot determine time anchor for relative time range")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", expr, dialect.EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table))

	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Priority: e.priority,
	})
	if err != nil {
		return time.Time{}, err
	}
	defer res.Close()

	var t time.Time
	for res.Next() {
		if err := res.Scan(&t); err != nil {
			return time.Time{}, fmt.Errorf("failed to scan time anchor: %w", err)
		}
	}

	if t.IsZero() {
		t = time.Now()
	}

	e.watermark = t
	return t, nil
}
