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

func (e *Executor) resolveDuckDB(ctx context.Context, timeExpr string) (TimestampsResult, error) {
	filter := e.security.RowFilter()
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}
	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)

	var watermarkExpr string
	if e.metricsView.WatermarkExpression != "" {
		watermarkExpr = e.metricsView.WatermarkExpression
	} else {
		watermarkExpr = fmt.Sprintf("max(%s)", timeExpr)
	}

	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", %[2]s as \"watermark\" FROM %[3]s %[4]s",
		timeExpr,
		watermarkExpr,
		escapedTableName,
		filter,
	)

	rows, err := e.olap.Query(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return TimestampsResult{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var minTime, maxTime, watermark *time.Time
		err = rows.Scan(&minTime, &maxTime, &watermark)
		if err != nil {
			return TimestampsResult{}, err
		}
		return TimestampsResult{
			Min:       safeTime(minTime),
			Max:       safeTime(maxTime),
			Watermark: safeTime(watermark),
		}, nil
	}

	err = rows.Err()
	if err != nil {
		return TimestampsResult{}, err
	}

	return TimestampsResult{}, errors.New("no rows returned")
}

func (e *Executor) resolveClickHouse(ctx context.Context, timeExpr string) (TimestampsResult, error) {
	filter := e.security.RowFilter()
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}
	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)

	var watermarkExpr string
	if e.metricsView.WatermarkExpression != "" {
		watermarkExpr = e.metricsView.WatermarkExpression
	} else {
		watermarkExpr = fmt.Sprintf("max(%s)", timeExpr)
	}

	// if datetime column is not nullable then ch returns 0 value instead of NULL when there are no rows so set aggregate_functions_null_for_empty
	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", %[2]s as \"watermark\" FROM %[3]s %[4]s SETTINGS aggregate_functions_null_for_empty = 1",
		timeExpr,
		watermarkExpr,
		escapedTableName,
		filter,
	)

	rows, err := e.olap.Query(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return TimestampsResult{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var minTime, maxTime, watermark *time.Time
		err = rows.Scan(&minTime, &maxTime, &watermark)
		if err != nil {
			return TimestampsResult{}, err
		}
		return TimestampsResult{
			Min:       safeTime(minTime),
			Max:       safeTime(maxTime),
			Watermark: safeTime(watermark),
		}, nil
	}

	err = rows.Err()
	if err != nil {
		return TimestampsResult{}, err
	}

	return TimestampsResult{}, errors.New("no rows returned")
}

func (e *Executor) resolvePinot(ctx context.Context, timeExpr string) (TimestampsResult, error) {
	filter := e.security.RowFilter()
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}
	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)

	var watermarkExpr string
	if e.metricsView.WatermarkExpression != "" {
		watermarkExpr = e.metricsView.WatermarkExpression
	} else {
		watermarkExpr = fmt.Sprintf("max(%s)", timeExpr)
	}

	rangeSQL := fmt.Sprintf(
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", %[2]s as \"watermark\" FROM %[3]s %[4]s",
		timeExpr,
		watermarkExpr,
		escapedTableName,
		filter,
	)

	rows, err := e.olap.Query(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return TimestampsResult{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var minTime, maxTime, watermark *time.Time
		err = rows.Scan(&minTime, &maxTime, &watermark)
		if err != nil {
			// retry again with long type as pinot supports timestamp columns with long type
			var minTime, maxTime, watermark int64
			innerErr := rows.Scan(&minTime, &maxTime, &watermark)
			if innerErr != nil {
				return TimestampsResult{}, err
			}
			return TimestampsResult{
				Min:       time.UnixMilli(minTime),
				Max:       time.UnixMilli(maxTime),
				Watermark: time.UnixMilli(watermark),
			}, nil
		}
		return TimestampsResult{
			Min:       safeTime(minTime),
			Max:       safeTime(maxTime),
			Watermark: safeTime(watermark),
		}, nil
	}

	err = rows.Err()
	if err != nil {
		return TimestampsResult{}, err
	}

	return TimestampsResult{}, errors.New("no rows returned")
}

func (e *Executor) resolveDruid(ctx context.Context, timeExpr string) (TimestampsResult, error) {
	filter := e.security.RowFilter()
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}
	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)

	var ts TimestampsResult
	group, ctx := errgroup.WithContext(ctx)

	// don't populate the cache, but use it if it's there as druid timeboundary query will create a cache entry for each segment
	useCache := true
	populateCache := false

	group.Go(func() error {
		minSQL := fmt.Sprintf(
			"SELECT min(%[1]s) as \"min\" FROM %[2]s %[3]s",
			timeExpr,
			escapedTableName,
			filter,
		)

		rows, err := e.olap.Query(ctx, &drivers.Statement{
			Query:            minSQL,
			Priority:         e.priority,
			ExecutionTimeout: defaultExecutionTimeout,
			UseCache:         &useCache,
			PopulateCache:    &populateCache,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&ts.Min)
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
			timeExpr,
			escapedTableName,
			filter,
		)

		rows, err := e.olap.Query(ctx, &drivers.Statement{
			Query:            maxSQL,
			Priority:         e.priority,
			ExecutionTimeout: defaultExecutionTimeout,
			UseCache:         &useCache,
			PopulateCache:    &populateCache,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&ts.Max)
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

	if e.metricsView.WatermarkExpression != "" {
		group.Go(func() error {
			maxSQL := fmt.Sprintf(
				"SELECT %[1]s as \"watermark\" FROM %[2]s %[3]s",
				e.metricsView.WatermarkExpression,
				escapedTableName,
				filter,
			)

			rows, err := e.olap.Query(ctx, &drivers.Statement{
				Query:            maxSQL,
				Priority:         e.priority,
				ExecutionTimeout: defaultExecutionTimeout,
				UseCache:         &useCache,
				PopulateCache:    &populateCache,
			})
			if err != nil {
				return err
			}
			defer rows.Close()

			if rows.Next() {
				err = rows.Scan(&ts.Watermark)
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
	}

	err := group.Wait()
	if err != nil {
		return TimestampsResult{}, err
	}

	// If there's no custom watermark expression, the watermark defaults to the max time.
	if e.metricsView.WatermarkExpression == "" {
		ts.Watermark = ts.Max
	}

	return ts, nil
}

func safeTime(tm *time.Time) time.Time {
	if tm == nil {
		return time.Time{}
	}
	return *tm
}
