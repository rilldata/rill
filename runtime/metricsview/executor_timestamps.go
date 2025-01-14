package metricsview

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/sync/errgroup"
)

const (
	defaultExecutionTimeout = time.Minute * 3
)

func (e *Executor) resolveDuckDB(ctx context.Context) (time.Time, time.Time, time.Time, error) {
	timeDim := e.olap.Dialect().EscapeIdentifier(e.metricsView.TimeDimension)
	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)

	var watermarkExpr string
	if e.metricsView.WatermarkExpression != "" {
		watermarkExpr = e.metricsView.WatermarkExpression
	} else {
		watermarkExpr = fmt.Sprintf("MAX(%s)", timeDim)
	}

	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", %[2]s as \"watermark\", max(%[1]s) - min(%[1]s) as \"interval\" FROM %[3]s %[4]s",
		timeDim,
		watermarkExpr,
		escapedTableName,
		e.security.RowFilter(),
	)

	rows, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}
	defer rows.Close()

	if rows.Next() {
		rowMap := make(map[string]any)
		err = rows.MapScan(rowMap)
		if err != nil {
			return time.Time{}, time.Time{}, time.Time{}, err
		}
		if v := rowMap["min"]; v != nil {
			minTime, ok := v.(time.Time)
			if !ok {
				return time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("not a timestamp column")
			}
			return minTime, rowMap["max"].(time.Time), rowMap["watermark"].(time.Time), nil
		}
	}

	err = rows.Err()
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}

	return time.Time{}, time.Time{}, time.Time{}, errors.New("no rows returned")
}

func (e *Executor) resolveDruid(ctx context.Context) (time.Time, time.Time, time.Time, error) {
	filter := e.security.RowFilter()
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}
	timeDim := e.olap.Dialect().EscapeIdentifier(e.metricsView.TimeDimension)
	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)

	var minTime, maxTime, watermark time.Time
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		minSQL := fmt.Sprintf(
			"SELECT min(%[1]s) as \"min\" FROM %[2]s %[3]s",
			timeDim,
			escapedTableName,
			filter,
		)

		rows, err := e.olap.Execute(ctx, &drivers.Statement{
			Query:            minSQL,
			Priority:         e.priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&minTime)
			if err != nil {
				return err
			}
		} else {
			err = rows.Err()
			if err != nil {
				return err
			}
			return errors.New("no rows returned for min time")
		}

		return nil
	})

	group.Go(func() error {
		maxSQL := fmt.Sprintf(
			"SELECT max(%[1]s) as \"max\" FROM %[2]s %[3]s",
			timeDim,
			escapedTableName,
			filter,
		)

		rows, err := e.olap.Execute(ctx, &drivers.Statement{
			Query:            maxSQL,
			Priority:         e.priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&maxTime)
			if err != nil {
				return err
			}
		} else {
			err = rows.Err()
			if err != nil {
				return err
			}
			return errors.New("no rows returned for max time")
		}
		return nil
	})

	group.Go(func() error {
		var watermarkExpr string
		if e.metricsView.WatermarkExpression != "" {
			watermarkExpr = e.metricsView.WatermarkExpression
		} else {
			watermarkExpr = fmt.Sprintf("MAX(%s)", timeDim)
		}

		maxSQL := fmt.Sprintf(
			"SELECT %[1]s as \"watermark\" FROM %[2]s %[3]s",
			watermarkExpr,
			escapedTableName,
			filter,
		)

		rows, err := e.olap.Execute(ctx, &drivers.Statement{
			Query:            maxSQL,
			Priority:         e.priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&watermark)
			if err != nil {
				return err
			}
		} else {
			err = rows.Err()
			if err != nil {
				return err
			}
			return errors.New("no rows returned for max time")
		}
		return nil
	})

	err := group.Wait()
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}

	return minTime, maxTime, watermark, nil
}

func (e *Executor) resolveClickHouseAndPinot(ctx context.Context) (time.Time, time.Time, time.Time, error) {
	filter := e.security.RowFilter()
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}
	timeDim := e.olap.Dialect().EscapeIdentifier(e.metricsView.TimeDimension)
	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)

	var watermarkExpr string
	if e.metricsView.WatermarkExpression != "" {
		watermarkExpr = e.metricsView.WatermarkExpression
	} else {
		watermarkExpr = fmt.Sprintf("MAX(%s)", timeDim)
	}

	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) AS \"min\", max(%[1]s) AS \"max\", %[2]s as \"watermark\" FROM %[3]s %[4]s",
		timeDim,
		watermarkExpr,
		escapedTableName,
		filter,
	)

	rows, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var minVal, maxVal, watermark *time.Time
		err = rows.Scan(&minVal, &maxVal, &watermark)
		if err != nil {
			return time.Time{}, time.Time{}, time.Time{}, err
		}
		return *minVal, *maxVal, *watermark, nil
	}

	err = rows.Err()
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}

	return time.Time{}, time.Time{}, time.Time{}, errors.New("no rows returned")
}
