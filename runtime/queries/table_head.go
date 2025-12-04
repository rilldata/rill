package queries

import (
	"context"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
)

type TableHead struct {
	Connector      string
	Database       string
	DatabaseSchema string
	TableName      string
	Limit          int
	Result         []*structpb.Struct
	Schema         *runtimev1.StructType
}

var _ runtime.Query = &TableHead{}

var supportedTableHeadDialects = map[drivers.Dialect]bool{
	drivers.DialectDuckDB:     true,
	drivers.DialectClickHouse: true,
	drivers.DialectDruid:      true,
	drivers.DialectPinot:      true,
	drivers.DialectBigQuery:   true,
	drivers.DialectSnowflake:  true,
	drivers.DialectAthena:     true,
	drivers.DialectRedshift:   true,
	drivers.DialectMySQL:      true,
	drivers.DialectPostgres:   true,
}

func (q *TableHead) Key() string {
	return fmt.Sprintf("TableHead:%s:%d", q.TableName, q.Limit)
}

func (q *TableHead) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *TableHead) MarshalResult() *runtime.QueryResult {
	var size int64
	if len(q.Result) > 0 {
		// approx
		size = sizeProtoMessage(q.Result[0]) * int64(len(q.Result))
	}

	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: size,
	}
}

func (q *TableHead) UnmarshalResult(v any) error {
	res, ok := v.([]*structpb.Struct)
	if !ok {
		return fmt.Errorf("TableHead: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *TableHead) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	if !supportedTableHeadDialects[olap.Dialect()] {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	query, err := q.buildTableHeadSQL(ctx, olap)
	if err != nil {
		return err
	}

	rows, err := olap.Query(ctx, &drivers.Statement{
		Query:            query,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	data, err := rowsToData(rows)
	if err != nil {
		return err
	}

	q.Result = data
	q.Schema = rows.Schema
	return nil
}

func (q *TableHead) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		if opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_CSV || opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET {
			filename := q.TableName
			sql, err := q.buildTableHeadSQL(ctx, olap)
			if err != nil {
				return err
			}
			args := []interface{}{}
			if err := DuckDBCopyExport(ctx, w, opts, sql, args, filename, olap, opts.Format); err != nil {
				return err
			}
		} else {
			if err := q.generalExport(ctx, rt, instanceID, w, opts); err != nil {
				return err
			}
		}
	case drivers.DialectDruid:
		if err := q.generalExport(ctx, rt, instanceID, w, opts); err != nil {
			return err
		}
	case drivers.DialectClickHouse:
		if err := q.generalExport(ctx, rt, instanceID, w, opts); err != nil {
			return err
		}
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	return nil
}

func (q *TableHead) generalExport(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(q.TableName)
		if err != nil {
			return err
		}
	}

	meta := structTypeToMetricsViewColumn(q.Schema)

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return WriteCSV(meta, q.Result, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return WriteXLSX(meta, q.Result, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return WriteParquet(meta, q.Result, w)
	}

	return nil
}

func (q *TableHead) buildTableHeadSQL(ctx context.Context, olap drivers.OLAPStore) (string, error) {
	columns, err := supportedColumns(ctx, olap, q.Database, q.DatabaseSchema, q.TableName)
	if err != nil {
		return "", err
	}

	limitClause := ""
	if q.Limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	sql := fmt.Sprintf(
		`SELECT %s FROM %s%s`,
		strings.Join(columns, ","),
		olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
		limitClause,
	)
	return sql, nil
}

func supportedColumns(ctx context.Context, olap drivers.OLAPStore, db, schema, tblName string) ([]string, error) {
	tbl, err := olap.InformationSchema().Lookup(ctx, db, schema, tblName)
	if err != nil {
		return nil, err
	}
	var columns []string
	for _, field := range tbl.Schema.Fields {
		columns = append(columns, olap.Dialect().EscapeIdentifier(field.Name))
	}
	return columns, nil
}
