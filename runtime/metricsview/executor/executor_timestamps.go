package executor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"golang.org/x/sync/errgroup"
)

const (
	defaultExecutionTimeout = time.Minute * 3
)

func (e *Executor) resolveDuckDB(ctx context.Context, timeExpr string) (metricsview.TimestampsResult, error) {
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
		QueryAttributes:  e.queryAttributes,
	})
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var minTime, maxTime, watermark *time.Time
		err = rows.Scan(&minTime, &maxTime, &watermark)
		if err != nil {
			return metricsview.TimestampsResult{}, err
		}
		return metricsview.TimestampsResult{
			Min:       safeTime(minTime),
			Max:       safeTime(maxTime),
			Watermark: safeTime(watermark),
		}, nil
	}

	err = rows.Err()
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}

	return metricsview.TimestampsResult{}, errors.New("no rows returned")
}

func (e *Executor) resolveClickHouse(ctx context.Context, timeExpr string) (metricsview.TimestampsResult, error) {
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
		"SELECT min(%[1]s) as \"min\", max(%[1]s) as \"max\", %[2]s as \"watermark\", count(1) > 0 AS count FROM %[3]s %[4]s",
		timeExpr,
		watermarkExpr,
		escapedTableName,
		filter,
	)

	rows, err := e.olap.Query(ctx, &drivers.Statement{
		Query:            rangeSQL,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
		QueryAttributes:  e.queryAttributes,
	})
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var minTime, maxTime, watermark *time.Time
		var count int
		err = rows.Scan(&minTime, &maxTime, &watermark, &count)
		if err != nil {
			return metricsview.TimestampsResult{}, err
		}
		if count == 0 {
			// if datetime column is not nullable then ch returns 0 value instead of NULL when there are no rows
			// The 0 value thus returned does not return true for IsZero() check
			return metricsview.TimestampsResult{}, nil
		}
		return metricsview.TimestampsResult{
			Min:       safeTime(minTime),
			Max:       safeTime(maxTime),
			Watermark: safeTime(watermark),
		}, nil
	}

	err = rows.Err()
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}

	return metricsview.TimestampsResult{}, errors.New("no rows returned")
}

func (e *Executor) resolvePinot(ctx context.Context, timeExpr string) (metricsview.TimestampsResult, error) {
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
		QueryAttributes:  e.queryAttributes,
	})
	if err != nil {
		return metricsview.TimestampsResult{}, err
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
				return metricsview.TimestampsResult{}, err
			}
			return metricsview.TimestampsResult{
				Min:       time.UnixMilli(minTime),
				Max:       time.UnixMilli(maxTime),
				Watermark: time.UnixMilli(watermark),
			}, nil
		}
		return metricsview.TimestampsResult{
			Min:       safeTime(minTime),
			Max:       safeTime(maxTime),
			Watermark: safeTime(watermark),
		}, nil
	}

	err = rows.Err()
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}

	return metricsview.TimestampsResult{}, errors.New("no rows returned")
}

func (e *Executor) resolveDruid(ctx context.Context, timeExpr string) (metricsview.TimestampsResult, error) {
	filter := e.security.RowFilter()
	if filter != "" {
		filter = fmt.Sprintf(" WHERE %s", filter)
	}
	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)

	var ts metricsview.TimestampsResult
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
			QueryAttributes:  e.queryAttributes,
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
			// don't return error if there are no rows as druid does not return any rows when where clause does not match
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
			QueryAttributes:  e.queryAttributes,
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
			// don't return error if there are no rows as druid does not return any rows when where clause does not match
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
				QueryAttributes:  e.queryAttributes,
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
				// don't return error if there are no rows as druid does not return any rows when where clause does not match
			}
			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}

	// If there's no custom watermark expression, the watermark defaults to the max time.
	if e.metricsView.WatermarkExpression == "" {
		ts.Watermark = ts.Max
	}

	return ts, nil
}

func (e *Executor) resolveStarRocks(ctx context.Context, timeExpr string) (metricsview.TimestampsResult, error) {
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
		"SELECT min(%[1]s) as `min`, max(%[1]s) as `max`, %[2]s as `watermark` FROM %[3]s %[4]s",
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
		return metricsview.TimestampsResult{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var minTime, maxTime, watermark *time.Time
		err = rows.Scan(&minTime, &maxTime, &watermark)
		if err != nil {
			return metricsview.TimestampsResult{}, err
		}
		return metricsview.TimestampsResult{
			Min:       safeTime(minTime),
			Max:       safeTime(maxTime),
			Watermark: safeTime(watermark),
		}, nil
	}

	err = rows.Err()
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}

	return metricsview.TimestampsResult{}, errors.New("no rows returned")
}

func safeTime(tm *time.Time) time.Time {
	if tm == nil {
		return time.Time{}
	}
	return *tm
}

// fetchTimestamps queries min/max timestamps from an arbitrary table using dialect-specific optimizations.
// For Druid, it runs separate parallel queries for min and max to use metadata lookups.
func (e *Executor) fetchTimestamps(ctx context.Context, db, schema, table string) (time.Time, time.Time, error) {
	timeExpr, err := e.timeColumnOrExpr(e.metricsView.TimeDimension)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	escapedTable := e.olap.Dialect().EscapeTable(db, schema, table)

	if e.olap.Dialect() == drivers.DialectDruid {
		return e.fetchTimestampsDruid(ctx, timeExpr, escapedTable)
	}
	return e.fetchTimestampsDefault(ctx, timeExpr, escapedTable)
}

// fetchTimestampsDefault runs a single MIN/MAX query (works for DuckDB, ClickHouse, Pinot, StarRocks).
func (e *Executor) fetchTimestampsDefault(ctx context.Context, timeExpr, escapedTable string) (time.Time, time.Time, error) {
	sql := fmt.Sprintf("SELECT MIN(%s), MAX(%s) FROM %s", timeExpr, timeExpr, escapedTable)
	rows, err := e.olap.Query(ctx, &drivers.Statement{
		Query:            sql,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
		QueryAttributes:  e.queryAttributes,
	})
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var minTime, maxTime *time.Time
		if err := rows.Scan(&minTime, &maxTime); err != nil {
			return time.Time{}, time.Time{}, err
		}
		return safeTime(minTime), safeTime(maxTime), nil
	}

	return time.Time{}, time.Time{}, errors.New("no rows returned")
}

// fetchTimestampsDruid runs parallel min/max queries to leverage Druid's metadata-based lookups.
func (e *Executor) fetchTimestampsDruid(ctx context.Context, timeExpr, escapedTable string) (time.Time, time.Time, error) {
	useCache := true
	populateCache := false

	var minTime, maxTime time.Time
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		minSQL := fmt.Sprintf("SELECT MIN(%s) as \"min\" FROM %s", timeExpr, escapedTable)
		rows, err := e.olap.Query(ctx, &drivers.Statement{
			Query:            minSQL,
			Priority:         e.priority,
			ExecutionTimeout: defaultExecutionTimeout,
			UseCache:         &useCache,
			PopulateCache:    &populateCache,
			QueryAttributes:  e.queryAttributes,
		})
		if err != nil {
			return err
		}
		defer rows.Close()
		if rows.Next() {
			return rows.Scan(&minTime)
		}
		return nil
	})

	group.Go(func() error {
		maxSQL := fmt.Sprintf("SELECT MAX(%s) as \"max\" FROM %s", timeExpr, escapedTable)
		rows, err := e.olap.Query(ctx, &drivers.Statement{
			Query:            maxSQL,
			Priority:         e.priority,
			ExecutionTimeout: defaultExecutionTimeout,
			UseCache:         &useCache,
			PopulateCache:    &populateCache,
			QueryAttributes:  e.queryAttributes,
		})
		if err != nil {
			return err
		}
		defer rows.Close()
		if rows.Next() {
			return rows.Scan(&maxTime)
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		return time.Time{}, time.Time{}, err
	}

	return minTime, maxTime, nil
}
