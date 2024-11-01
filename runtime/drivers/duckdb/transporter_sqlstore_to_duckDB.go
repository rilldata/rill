package duckdb

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

type sqlStoreToDuckDB struct {
	to     *connection
	from   drivers.SQLStore
	logger *zap.Logger
}

var _ drivers.Transporter = &sqlStoreToDuckDB{}

func newSQLStoreToDuckDB(from drivers.SQLStore, c *connection, logger *zap.Logger) drivers.Transporter {
	return &sqlStoreToDuckDB{
		to:     c,
		from:   from,
		logger: logger,
	}
}

func (s *sqlStoreToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) (transferErr error) {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	s.logger = s.logger.With(zap.String("source", sinkCfg.Table))

	rowIter, err := s.from.Query(ctx, srcProps)
	if err != nil {
		return err
	}
	defer func() {
		err := rowIter.Close()
		if err != nil && !errors.Is(err, ctx.Err()) {
			s.logger.Error("error in closing row iterator", zap.Error(err))
		}
	}()
	return s.transferFromRowIterator(ctx, rowIter, sinkCfg.Table)
}

func (s *sqlStoreToDuckDB) transferFromRowIterator(ctx context.Context, iter drivers.RowIterator, table string) error {
	schema, err := iter.Schema(ctx)
	if err != nil {
		if errors.Is(err, drivers.ErrIteratorDone) {
			return drivers.ErrNoRows
		}
		return err
	}

	if total, ok := iter.Size(drivers.ProgressUnitRecord); ok {
		s.logger.Debug("records to be ingested", zap.Uint64("rows", total))
	}
	// we first ingest data in a temporary table in the main db
	// and then copy it to the final table to ensure that the final table is always created using CRUD APIs
	tmpTable := fmt.Sprintf("__%s_tmp_sqlstore", table)
	// generate create table query
	qry, err := createTableQuery(schema, tmpTable)
	if err != nil {
		return err
	}

	rwConn, release, err := s.to.acquireConn(ctx, false)
	if err != nil {
		return err
	}
	defer release()
	conn := rwConn.Connx()

	// create table
	_, err = conn.ExecContext(ctx, qry, nil)
	if err != nil {
		return err
	}

	defer func() {
		// ensure temporary table is cleaned
		_, err = conn.ExecContext(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", tmpTable))
		if err != nil {
			s.logger.Error("failed to drop temp table", zap.String("table", tmpTable), zap.Error(err))
		}
	}()

	// append data using appender API
	err = rawConn(conn.Conn, func(conn driver.Conn) error {
		a, err := duckdb.NewAppenderFromConn(conn, "", tmpTable)
		if err != nil {
			return err
		}
		defer func() {
			err = a.Close()
			if err != nil {
				s.logger.Error("appender closed failed", zap.Error(err))
			}
		}()

		for num := 0; ; num++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if num == 10000 {
					num = 0
					if err := a.Flush(); err != nil {
						return err
					}
				}

				row, err := iter.Next(ctx)
				if err != nil {
					if errors.Is(err, drivers.ErrIteratorDone) {
						return nil
					}
					return err
				}
				if err := convert(row, schema); err != nil { // duckdb specific datatype conversion
					return err
				}

				if err := a.AppendRow(row...); err != nil {
					return err
				}
			}
		}
	})
	if err != nil {
		return err
	}

	// copy data from temp table to target table
	return s.to.CreateTableAsSelect(ctx, table, false, fmt.Sprintf("SELECT * FROM %s", tmpTable), nil)
}

func createTableQuery(schema *runtimev1.StructType, name string) (string, error) {
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s(", safeName(name))
	for i, s := range schema.Fields {
		i++
		duckDBType, err := pbTypeToDuckDB(s.Type)
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

func convert(row []driver.Value, schema *runtimev1.StructType) error {
	for i, v := range row {
		if v == nil {
			continue
		}
		if schema.Fields[i].Type.Code == runtimev1.Type_CODE_UUID {
			val, ok := v.([16]byte)
			if !ok {
				return fmt.Errorf("unknown type for UUID field %s: %T", schema.Fields[i].Name, v)
			}
			var uuid duckdb.UUID
			copy(uuid[:], val[:])
			row[i] = uuid
		}
	}
	return nil
}

func pbTypeToDuckDB(t *runtimev1.Type) (string, error) {
	code := t.Code
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
		// keeping type as json but appending varchar using the appender API causes duckdb invalid vector error intermittently
		return "VARCHAR", nil
	case runtimev1.Type_CODE_UUID:
		return "UUID", nil
	default:
		return "", fmt.Errorf("unknown type_code %s", code)
	}
}
