package bigquery

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

// Exec implements drivers.SQLStore
func (c *Connection) Exec(ctx context.Context, src *drivers.DatabaseSource) (drivers.RowIterator, error) {
	props, err := parseSourceProperties(src.Props)
	if err != nil {
		return nil, err
	}

	client, err := c.createClient(ctx, props)
	if err != nil {
		if strings.Contains(err.Error(), "unable to detect projectID") {
			return nil, fmt.Errorf("projectID not detected in credentials. Please set project ID")
		}
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	if err := client.EnableStorageReadClient(ctx); err != nil {
		client.Close()
		return nil, err
	}

	q := client.Query(src.Query)
	it, err := q.Read(ctx)
	if err != nil && !strings.Contains(err.Error(), "Syntax error") {
		c.logger.Info("query failed, retyring without storage api", zap.Error(err))
		// the query results are always cached in a temporary table that storage api can use
		// there are some exceptions when results aren't cached
		// so we also try without storage api
		client, err = c.createClient(ctx, props)
		if err != nil {
			return nil, fmt.Errorf("failed to create bigquery client: %w", err)
		}

		q := client.Query(src.Query)
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
	schema  drivers.Schema
	bqIter  *bigquery.RowIterator
}

var _ drivers.RowIterator = &rowIterator{}

func (r *rowIterator) ResultSchema(ctx context.Context) (drivers.Schema, error) {
	if r.schema != nil {
		return r.schema, nil
	}

	// schema is only available after first next call
	r.next, r.nexterr = r.Next(ctx)
	if r.nexterr != nil {
		return nil, r.nexterr
	}

	r.schema = make([]drivers.Field, len(r.bqIter.Schema))
	for i, s := range r.bqIter.Schema {
		dbt, err := bqToDuckDB(string(s.Type))
		if err != nil {
			return nil, err
		}

		r.schema[i] = drivers.Field{Name: s.Name, Type: dbt}
	}
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

func bqToDuckDB(dbt string) (string, error) {
	switch dbt {
	case "STRING":
		return "VARCHAR", nil
	case "JSON":
		return "VARCHAR", nil
	case "INTERVAL":
		return "VARCHAR", nil
	case "GEOGRAPHY":
		return "VARCHAR", nil
	case "FLOAT":
		return "DOUBLE", nil
	// TODO :: NUMERIC and BIGNUMERIC are represented as *big.Rat type.
	// There is no support for these types in go-duckdb driver.
	// Users can cast these to duckdb types in model.
	case "NUMERIC":
		return "VARCHAR", nil
	case "BIGNUMERIC":
		return "VARCHAR", nil
	case "TIMESTAMP":
		return "TIMESTAMP", nil
	// TODO :: DATETIME, TIME, DATE doesn't have equivalent constructs in go and not supported in go-duckdb driver.
	// Users can cast these to duckdb types in model.
	case "DATETIME":
		return "VARCHAR", nil
	case "TIME":
		return "VARCHAR", nil
	case "DATE":
		return "VARCHAR", nil
	case "BOOLEAN":
		return "BOOLEAN", nil
	case "INTEGER":
		return "INTEGER", nil
	case "BYTES":
		return "BLOB", nil
	case "RECORD":
		return "", fmt.Errorf("record type not supported")
	default:
		// TODO :: may be just use VARCHAR ?
		return "", fmt.Errorf("type %s not supported", dbt)
	}
}

func convert(v any) any {
	if v == nil {
		return nil
	}
	// refer to documentation on bigquery.RowIterator.Next for the superset of all go types possible
	switch val := v.(type) {
	case civil.Date:
		return val.String()
	case civil.Time:
		return val.String()
	case civil.DateTime:
		return val.String()
	case *big.Rat:
		f, _ := val.Float64()
		if math.IsInf(f, 0) {
			// big.Rat can't always be represented in float64
			// we use string notation(in the form num/denom) in such cases which is not ideal
			return val.String()
		}
		return f
	default:
		return val
	}
}
