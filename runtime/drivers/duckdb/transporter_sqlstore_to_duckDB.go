package duckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const (
	_sqlStoreIteratorBatchSize = 32
	_batchSize                 = 10000
)

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

func (s *sqlStoreToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) (transferErr error) {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	s.logger = s.logger.With(zap.String("source", sinkCfg.Table))

	rowIter, err := s.from.Query(ctx, srcProps)
	if err != nil {
		if !errors.Is(err, drivers.ErrNotImplemented) {
			return err
		}
	} else { // no error consume rowIterator
		defer rowIter.Close()
		return s.transferFromRowIterator(ctx, rowIter, sinkCfg.Table, opts.Progress)
	}
	limitInBytes, _ := s.to.(drivers.Handle).Config()["storage_limit_bytes"].(int64)
	if limitInBytes == 0 {
		limitInBytes = math.MaxInt64
	}
	iter, err := s.from.QueryAsFiles(ctx, srcProps, &drivers.QueryOption{TotalLimitInBytes: limitInBytes}, opts.Progress)
	if err != nil {
		return err
	}
	defer iter.Close()

	start := time.Now()
	s.logger.Debug("started transfer from local file to duckdb", zap.String("sink_table", sinkCfg.Table), observability.ZapCtx(ctx))
	defer func() {
		s.logger.Debug("transfer finished",
			zap.Duration("duration", time.Since(start)),
			zap.Bool("success", transferErr == nil),
			observability.ZapCtx(ctx))
	}()
	create := true
	// TODO :: iteration over fileiterator is similar(apart from no schema changes possible here)
	// to consuming fileIterator in objectStore_to_duckDB
	// both can be refactored to follow same path
	for {
		files, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		format := fileutil.FullExt(files[0])
		if iter.Format() != "" {
			format += "." + iter.Format()
		}

		from, err := sourceReader(files, format, make(map[string]any))
		if err != nil {
			return err
		}

		if create {
			err = s.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM %s", from))
			create = false
		} else {
			err = s.to.InsertTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM %s", from))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sqlStoreToDuckDB) transferFromRowIterator(ctx context.Context, iter drivers.RowIterator, table string, p drivers.Progress) error {
	schema, err := iter.Schema(ctx)
	if err != nil {
		if errors.Is(err, drivers.ErrIteratorDone) {
			return fmt.Errorf("no results found for the query")
		}
		return err
	}

	if total, ok := iter.Size(drivers.ProgressUnitRecord); ok {
		s.logger.Debug("records to be ingested", zap.Uint64("rows", total))
		p.Target(int64(total), drivers.ProgressUnitRecord)
	}
	// we first ingest data in a temporary table in the main db
	// and then copy it to the final table to ensure that the final table is always created using CRUD APIs which takes care
	// whether table goes in main db or in separate table specific db
	tmpTable := fmt.Sprintf("__%s_tmp_sqlstore", table)
	// generate create table query
	qry, err := CreateTableQuery(schema, tmpTable)
	if err != nil {
		return err
	}

	// create table
	err = s.to.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1, LongRunning: true})
	if err != nil {
		return err
	}

	defer func() {
		// ensure temporary table is cleaned
		err := s.to.Exec(context.Background(), &drivers.Statement{
			Query:       fmt.Sprintf("DROP TABLE IF EXISTS %s", tmpTable),
			Priority:    100,
			LongRunning: true,
		})
		if err != nil {
			s.logger.Error("failed to drop temp table", zap.String("table", tmpTable), zap.Error(err))
		}
	}()

	err = s.to.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, conn *sql.Conn) error {
		// append data using appender API
		return rawConn(conn, func(conn driver.Conn) error {
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

					if err := a.AppendRow(row...); err != nil {
						return err
					}
				}
			}
		})
	})
	if err != nil {
		return err
	}

	// copy data from temp table to target table
	return s.to.CreateTableAsSelect(ctx, table, false, fmt.Sprintf("SELECT * FROM %s", tmpTable))
}

func CreateTableQuery(schema *runtimev1.StructType, name string) (string, error) {
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
