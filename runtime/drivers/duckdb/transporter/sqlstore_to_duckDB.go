package transporter

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

const _batchSize = 10000

type sqlStoreToDuckDB struct {
	to     drivers.OLAPStore
	from   drivers.SQLStore
	logger *zap.Logger
}

var _ drivers.Transporter = &sqlStoreToDuckDB{}

func NewSQLStoreToDuckDB(from drivers.SQLStore, to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &sqlStoreToDuckDB{
		to:     to,
		from:   from,
		logger: logger,
	}
}

func (s *sqlStoreToDuckDB) Transfer(ctx context.Context, source drivers.Source, sink drivers.Sink, opts *drivers.TransferOpts, p drivers.Progress) error {
	src, ok := source.DatabaseSource()
	if !ok {
		return fmt.Errorf("type of source should `drivers.DatabaseSource`")
	}
	dbSink, ok := sink.DatabaseSink()
	if !ok {
		return fmt.Errorf("type of source should `drivers.DatabaseSink`")
	}

	iter, err := s.from.Exec(ctx, src)
	if err != nil {
		return err
	}
	defer iter.Close()

	schema, err := iter.ResultSchema(ctx)
	if err != nil {
		return err
	}

	if total, ok := iter.Size(drivers.ProgressUnitRecord); ok {
		s.logger.Info("records to be ingested", zap.Uint64("rows", total))
		p.Target(int64(total), drivers.ProgressUnitRecord)
	}
	// create table
	if err := s.to.Exec(ctx, &drivers.Statement{Query: createTableQuery(schema, dbSink.Table), Priority: 1}); err != nil {
		return err
	}

	return s.to.WithRaw(ctx, 1, func(driverConn any) error {
		var conn driver.Conn
		// we are wrapping connections with otel connections
		// appender need duckdb driver connection
		if c, ok := driverConn.(rawer); ok {
			conn = c.Raw()
		} else {
			conn = driverConn.(driver.Conn)
		}

		if err != nil {
			return err
		}

		a, err := duckdb.NewAppenderFromConn(conn, "", dbSink.Table)
		if err != nil {
			return err
		}
		defer a.Close()

		// TODO :: may be add metric for length of buffer to determine an optimal capacity?
		ch := make(chan result, 1000)
		go func() {
			for {
				select {
				case <-ctx.Done():
					ch <- result{val: nil, err: ctx.Err()}
					close(ch)
					return
				default:
					row, err := iter.Next(ctx)
					ch <- result{val: row, err: err}
					if err != nil {
						close(ch)
						// received an error return
						return
					}
				}
			}
		}()

		for num := 0; ; num++ {
			if num == _batchSize {
				p.Observe(_batchSize, drivers.ProgressUnitRecord)
				num = 0
				if err := a.Flush(); err != nil {
					return err
				}
			}

			res := <-ch
			if res.err != nil {
				if errors.Is(res.err, iterator.Done) {
					return nil
				}
				return res.err
			}

			colValues := make([]driver.Value, len(res.val))
			for i, col := range res.val {
				colValues[i] = driver.Value(col)
			}

			if err := a.AppendRowArray(colValues); err != nil {
				return err
			}
		}
	})
}

func createTableQuery(schema drivers.Schema, name string) string {
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s(", safeName(name))
	for i, s := range schema {
		i++
		query += fmt.Sprintf("%s %s", safeName(s.Name), s.Type)
		if i != len(schema) {
			query += ","
		}
	}
	query += ")"
	return query
}

type rawer interface {
	Raw() driver.Conn
}

type result struct {
	val []any
	err error
}
