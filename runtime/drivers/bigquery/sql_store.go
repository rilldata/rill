package bigquery

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

// Query implements drivers.SQLStore
func (c *Connection) Query(ctx context.Context, props map[string]any, sql string) (drivers.RowIterator, error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	client, err := c.createClient(ctx, srcProps)
	if err != nil {
		if strings.Contains(err.Error(), "unable to detect projectID") {
			return nil, fmt.Errorf("projectID not detected in credentials. Please set `project_id` in source yaml")
		}
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	if err := client.EnableStorageReadClient(ctx); err != nil {
		client.Close()
		return nil, err
	}

	q := client.Query(sql)
	it, err := q.Read(ctx)
	if err != nil && !strings.Contains(err.Error(), "Syntax error") {
		c.logger.Info("query failed, retrying without storage api", zap.Error(err))
		// the query results are always cached in a temporary table that storage api can use
		// there are some exceptions when results aren't cached
		// so we also try without storage api
		client, err = c.createClient(ctx, srcProps)
		if err != nil {
			return nil, fmt.Errorf("failed to create bigquery client: %w", err)
		}

		q := client.Query(sql)
		it, err = q.Read(ctx)
	}
	if err != nil {
		client.Close()
		return nil, err
	}

	return &rowIterator{
		client: client,
		bqIter: it,
	}, nil
}

type rowIterator struct {
	client  *bigquery.Client
	next    []any
	nexterr error
	schema  *runtimev1.StructType
	bqIter  *bigquery.RowIterator
}

var _ drivers.RowIterator = &rowIterator{}

func (r *rowIterator) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	if r.schema != nil {
		return r.schema, nil
	}

	// schema is only available after first next call
	r.next, r.nexterr = r.Next(ctx)
	if r.nexterr != nil {
		return nil, r.nexterr
	}

	fields := make([]*runtimev1.StructType_Field, len(r.bqIter.Schema))
	for i, s := range r.bqIter.Schema {
		dbt, err := toPB(s)
		if err != nil {
			return nil, err
		}

		fields[i] = &runtimev1.StructType_Field{Name: s.Name, Type: dbt}
	}
	r.schema = &runtimev1.StructType{Fields: fields}
	return r.schema, nil
}

func (r *rowIterator) Next(ctx context.Context) ([]any, error) {
	if r.next != nil || r.nexterr != nil {
		next, err := r.next, r.nexterr
		r.next = nil
		r.nexterr = nil
		return next, err
	}

	var row row = make([]any, 0)
	if err := r.bqIter.Next(&row); err != nil {
		if errors.Is(err, iterator.Done) {
			return nil, drivers.ErrIteratorDone
		}
		return nil, err
	}

	return row, nil
}

func (r *rowIterator) Close() error {
	return r.client.Close()
}

func (r *rowIterator) Size(unit drivers.ProgressUnit) (uint64, bool) {
	if unit == drivers.ProgressUnitRecord {
		return r.bqIter.TotalRows, true
	}

	return 0, false
}

type row []any

var _ bigquery.ValueLoader = &row{}

func (r *row) Load(v []bigquery.Value, s bigquery.Schema) error {
	m := make([]any, len(v))
	for i := 0; i < len(v); i++ {
		if s[i].Type == bigquery.RecordFieldType || s[i].Repeated {
			return fmt.Errorf("repeated or nested data is not supported")
		}

		m[i] = convert(v[i])
	}
	*r = m
	return nil
}

// type conversion table for time
func toPB(field *bigquery.FieldSchema) (*runtimev1.Type, error) {
	t := &runtimev1.Type{Nullable: !field.Required}
	switch field.Type {
	case bigquery.StringFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.JSONFieldType:
		t.Code = runtimev1.Type_CODE_JSON
	case bigquery.IntervalFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.GeographyFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.FloatFieldType:
		t.Code = runtimev1.Type_CODE_FLOAT64
	case bigquery.NumericFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	// big numeric can have width upto 76 digits which can't be converted to DECIMAL type in duckdb
	// which supports width upto 38 digits only.
	case bigquery.BigNumericFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.TimestampFieldType:
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case bigquery.DateTimeFieldType:
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case bigquery.TimeFieldType:
		t.Code = runtimev1.Type_CODE_TIME
	case bigquery.DateFieldType:
		t.Code = runtimev1.Type_CODE_DATE
	case bigquery.BooleanFieldType:
		t.Code = runtimev1.Type_CODE_BOOL
	case bigquery.IntegerFieldType:
		t.Code = runtimev1.Type_CODE_INT64
	case bigquery.BytesFieldType:
		t.Code = runtimev1.Type_CODE_BYTES
	case bigquery.RecordFieldType:
		return nil, fmt.Errorf("record type not supported")
	default:
		// TODO :: may be just use VARCHAR ?
		return nil, fmt.Errorf("type %s not supported", field.Type)
	}
	return t, nil
}

func convert(v any) any {
	if v == nil {
		return nil
	}
	// refer to documentation on bigquery.RowIterator.Next for the superset of all go types possible
	switch val := v.(type) {
	case civil.Date:
		return val.In(time.UTC)
	case civil.Time:
		t, _ := time.Parse("15:04:05.999999999", val.String())
		return t
	case civil.DateTime:
		return val.In(time.UTC)
	case *big.Rat:
		if val.IsInt() {
			return val.FloatString(0)
		}
		return strings.TrimRight(val.FloatString(38), "0")
	default:
		return val
	}
}
