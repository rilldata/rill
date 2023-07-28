package transporter

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
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

	iter, err := s.from.Query(ctx, src.Props, src.SQL)
	if err != nil {
		return err
	}
	defer iter.Close()

	schema, err := iter.Schema(ctx)
	if err != nil {
		if errors.Is(err, drivers.ErrIteratorDone) {
			return fmt.Errorf("no results found for the query")
		}
		return err
	}

	if total, ok := iter.Size(drivers.ProgressUnitRecord); ok {
		s.logger.Info("records to be ingested", zap.Uint64("rows", total))
		p.Target(int64(total), drivers.ProgressUnitRecord)
	}
	// create table
	qry, err := createTableQuery(schema, dbSink.Table)
	if err != nil {
		return err
	}

	if err := s.to.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1}); err != nil {
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

		for num := 0; ; num++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if num == _batchSize {
					p.Observe(_batchSize, drivers.ProgressUnitRecord)
					num = 0
					if err := a.Flush(); err != nil {
						return err
					}
				}

				row, err := iter.Next(ctx)
				if err != nil {
					if errors.Is(err, drivers.ErrIteratorDone) {
						p.Observe(int64(num), drivers.ProgressUnitRecord)
						return nil
					}
					return err
				}

				colValues := make([]driver.Value, len(row))
				for i, col := range row {
					colValues[i] = driver.Value(col)
				}

				if err := a.AppendRowArray(colValues); err != nil {
					return err
				}
			}
		}
	})
}

func createTableQuery(schema *runtimev1.StructType, name string) (string, error) {
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s(", safeName(name))
	for i, s := range schema.Fields {
		i++
		duckDBType, err := pbTypeToDuckDB(s.Type.Code)
		if err != nil {
			return "", err
		}
		query += fmt.Sprintf("%s %s", safeName(s.Name), duckDBType)
		if i != len(schema.Fields) {
			query += ","
		}
	}
	query += ")"
	return query, nil
}

func pbTypeToDuckDB(code runtimev1.Type_Code) (string, error) {
	switch code {
	case runtimev1.Type_CODE_UNSPECIFIED:
		return "", fmt.Errorf("unspecified code")
	case runtimev1.Type_CODE_BOOL:
		return "BOOLEAN", nil
	case runtimev1.Type_CODE_INT8:
		return "TINYINT", nil
	case runtimev1.Type_CODE_INT16:
		return "SMALLINT", nil
	case runtimev1.Type_CODE_INT32:
		return "INTEGER", nil
	case runtimev1.Type_CODE_INT64:
		return "BIGINT", nil
	case runtimev1.Type_CODE_INT128:
		return "HUGEINT", nil
	case runtimev1.Type_CODE_UINT8:
		return "UTINYINT", nil
	case runtimev1.Type_CODE_UINT16:
		return "USMALLINT", nil
	case runtimev1.Type_CODE_UINT32:
		return "UINTEGER", nil
	case runtimev1.Type_CODE_UINT64:
		return "UBIGINT", nil
	case runtimev1.Type_CODE_FLOAT32:
		return "FLOAT", nil
	case runtimev1.Type_CODE_FLOAT64:
		return "DOUBLE", nil
	case runtimev1.Type_CODE_TIMESTAMP:
		return "TIMESTAMP", nil
	case runtimev1.Type_CODE_DATE:
		return "DATE", nil
	case runtimev1.Type_CODE_TIME:
		return "TIME", nil
	case runtimev1.Type_CODE_STRING:
		return "VARCHAR", nil
	case runtimev1.Type_CODE_BYTES:
		return "BLOB", nil
	case runtimev1.Type_CODE_ARRAY:
		return "", fmt.Errorf("array is not supported")
	case runtimev1.Type_CODE_STRUCT:
		return "", fmt.Errorf("struct is not supported")
	case runtimev1.Type_CODE_MAP:
		return "", fmt.Errorf("map is not supported")
	case runtimev1.Type_CODE_DECIMAL:
		return "DECIMAL", nil
	case runtimev1.Type_CODE_JSON:
		return "JSON", nil
	default:
		return "", fmt.Errorf("unknown type_code %s", code)
	}
}

type rawer interface {
	Raw() driver.Conn
}
