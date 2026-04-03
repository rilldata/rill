package queries

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

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
	drivers.DialectStarRocks:  true,
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
	case drivers.DialectStarRocks:
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
	tbl, err := olap.InformationSchema().Lookup(ctx, q.Database, q.DatabaseSchema, q.TableName)
	if err != nil {
		return "", err
	}

	var columns []string
	for _, field := range tbl.Schema.Fields {
		columns = append(columns, olap.Dialect().EscapeIdentifier(field.Name))
	}

	whereClause := ""
	if olap.Dialect() == drivers.DialectBigQuery && tbl.PartitionColumn != "" {
		latest, err := q.latestBigQueryPartition(ctx, olap)
		if err == nil && !latest.IsZero() {
			whereClause = fmt.Sprintf(
				" WHERE %s >= TIMESTAMP_SUB(CAST('%s' AS TIMESTAMP), INTERVAL 3 DAY)",
				olap.Dialect().EscapeIdentifier(tbl.PartitionColumn),
				latest.UTC().Format(time.RFC3339),
			)
		}
		// If fetching the latest partition fails or returns empty, proceed without a filter.
		// For tables with require_partition_filter=true this will error at query time, which is acceptable.
	} else if olap.Dialect() == drivers.DialectBigQuery && tbl.RangePartitionColumn != "" {
		maxID, ok, err := q.latestBigQueryRangePartition(ctx, olap)
		if err == nil && ok {
			whereClause = fmt.Sprintf(
				" WHERE %s >= %d",
				olap.Dialect().EscapeIdentifier(tbl.RangePartitionColumn),
				maxID,
			)
		}
		// Same as the time-partition case: proceed without a filter if the lookup fails.
	}

	limitClause := ""
	if q.Limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	return fmt.Sprintf(
		"SELECT %s FROM %s%s%s",
		strings.Join(columns, ", "),
		olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
		whereClause,
		limitClause,
	), nil
}

// latestBigQueryPartition queries INFORMATION_SCHEMA.PARTITIONS to find the most recent
// partition time for a table without scanning its data. This is safe to run even on tables
// with require_partition_filter=true because INFORMATION_SCHEMA is metadata-only.
func (q *TableHead) latestBigQueryPartition(ctx context.Context, olap drivers.OLAPStore) (time.Time, error) {
	var infoSchemaTable string
	if q.Database != "" {
		infoSchemaTable = fmt.Sprintf("`%s.%s.INFORMATION_SCHEMA.PARTITIONS`", q.Database, q.DatabaseSchema)
	} else {
		infoSchemaTable = fmt.Sprintf("`%s.INFORMATION_SCHEMA.PARTITIONS`", q.DatabaseSchema)
	}

	sql := fmt.Sprintf(
		"SELECT MAX(partition_id) FROM %s WHERE table_name = ? AND partition_id NOT IN ('__NULL__', '__UNPARTITIONED__')",
		infoSchemaTable,
	)

	rows, err := olap.Query(ctx, &drivers.Statement{
		Query:            sql,
		ExecutionTimeout: defaultExecutionTimeout,
		Args:             []any{q.TableName},
	})
	if err != nil {
		return time.Time{}, err
	}
	defer rows.Close()

	var partitionID *string
	if rows.Next() {
		if err := rows.Scan(&partitionID); err != nil {
			return time.Time{}, err
		}
	}
	if err := rows.Err(); err != nil {
		return time.Time{}, err
	}
	if partitionID == nil || *partitionID == "" {
		return time.Time{}, nil
	}

	return parseBigQueryPartitionID(*partitionID)
}

// latestBigQueryRangePartition queries INFORMATION_SCHEMA.PARTITIONS to find the lower bound of the
// most recent integer-range partition for the table. This is safe to run even on tables with
// require_partition_filter=true because INFORMATION_SCHEMA is metadata-only.
func (q *TableHead) latestBigQueryRangePartition(ctx context.Context, olap drivers.OLAPStore) (int64, bool, error) {
	var infoSchemaTable string
	if q.Database != "" {
		infoSchemaTable = fmt.Sprintf("`%s.%s.INFORMATION_SCHEMA.PARTITIONS`", q.Database, q.DatabaseSchema)
	} else {
		infoSchemaTable = fmt.Sprintf("`%s.INFORMATION_SCHEMA.PARTITIONS`", q.DatabaseSchema)
	}

	sql := fmt.Sprintf(
		"SELECT MAX(SAFE_CAST(partition_id AS INT64)) FROM %s WHERE table_name = ? AND partition_id NOT IN ('__NULL__', '__UNPARTITIONED__')",
		infoSchemaTable,
	)

	rows, err := olap.Query(ctx, &drivers.Statement{
		Query:            sql,
		ExecutionTimeout: defaultExecutionTimeout,
		Args:             []any{q.TableName},
	})
	if err != nil {
		return 0, false, err
	}
	defer rows.Close()

	var maxID *int64
	if rows.Next() {
		if err := rows.Scan(&maxID); err != nil {
			return 0, false, err
		}
	}
	if err := rows.Err(); err != nil {
		return 0, false, err
	}
	if maxID == nil {
		return 0, false, nil
	}
	return *maxID, true, nil
}

// parseBigQueryPartitionID converts a BigQuery partition_id string (e.g. "20240315") to a time.Time.
// BigQuery partition IDs use a fixed-width date format with no separators.
func parseBigQueryPartitionID(id string) (time.Time, error) {
	var format string
	switch len(id) {
	case 10:
		format = "2006010215"
	case 8:
		format = "20060102"
	case 6:
		format = "200601"
	case 4:
		format = "2006"
	default:
		return time.Time{}, fmt.Errorf("unrecognized partition_id format: %q", id)
	}

	t, err := time.Parse(format, id)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse partition_id %q with format %q: %w", id, format, err)
	}
	return t, nil
}
