package mysql

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// Query implements drivers.SQLStore
func (c *connection) Query(ctx context.Context, props map[string]any) (drivers.RowIterator, error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	var dsn string
	if srcProps.DSN != "" { // get from src properties
		dsn = srcProps.DSN
	} else if url, ok := c.config["dsn"].(string); ok && url != "" { // get from driver configs
		dsn = url
	} else {
		return nil, fmt.Errorf("the property 'dsn' is required for MySQL. Provide 'dsn' in the YAML properties or pass '--env connector.mysql.dsn=...' to 'rill start'")
	}

	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	conf.ParseTime = true // if set to false, time is scanned as an array rather than as time.Time

	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		db.Close()
		return nil, err
	}

	// Validate DSN data:
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	rows, err := db.QueryContext(ctx, srcProps.SQL)
	if err != nil {
		return nil, err
	}

	iter := &rowIterator{
		db:   db,
		rows: rows,
	}

	if err := iter.setSchema(ctx); err != nil {
		iter.Close()
		return nil, err
	}
	return iter, nil
}

// QueryAsFiles implements drivers.SQLStore
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any, opt *drivers.QueryOption, p drivers.Progress) (drivers.FileIterator, error) {
	return nil, drivers.ErrNotImplemented
}

type rowIterator struct {
	db   *sql.DB
	rows *sql.Rows

	schema       *runtimev1.StructType
	row          []sqldriver.Value
	fieldMappers []mapper
}

// Close implements drivers.RowIterator.
func (r *rowIterator) Close() error {
	r.rows.Close()
	r.db.Close()
	return nil
}

// Next implements drivers.RowIterator.
func (r *rowIterator) Next(ctx context.Context) ([]sqldriver.Value, error) {
	if !r.rows.Next() {
		err := r.rows.Err()
		if err == nil {
			return nil, drivers.ErrIteratorDone
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no results found for the query")
		}
		return nil, err
	}

	cts, err := r.rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	// Scan expects destinations to be pointers
	dests := make([]any, len(cts))
	for i := range dests {
		dests[i], err = r.fieldMappers[i].dest(cts[i].ScanType())
		if err != nil {
			return nil, err
		}
	}

	if err := r.rows.Scan(dests...); err != nil {
		return nil, err
	}

	for i := range r.schema.Fields {
		v := dests[i]
		if v == nil {
			r.row[i] = nil
			continue
		}
		// BIT type can be scanned as bytes only and hence needs to be converted to a bit string for the sake of usability
		// This is the only type that requires a conversion so no need for a generalization for now
		if bm, ok := r.fieldMappers[i].(*bitMapper); ok {
			r.row[i], err = bm.value(dests[i])
			if err != nil {
				return nil, err
			}
			continue
		}
		// Destinations are pointers that need to be dereferenced before passing further
		ptr := reflect.ValueOf(v)
		value := reflect.Indirect(ptr)
		// Binary destinations are slices that might be nil
		if value.Kind() == reflect.Slice && value.IsNil() {
			r.row[i] = nil
			continue
		}
		// Nullable columns are Valuers that need to be unwrapped
		if valuer, ok := value.Interface().(sqldriver.Valuer); ok {
			r.row[i], err = valuer.Value()
			if err != nil {
				return nil, err
			}
			continue
		}
		r.row[i] = value.Interface()
	}
	return r.row, nil
}

// Schema implements drivers.RowIterator.
func (r *rowIterator) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	return r.schema, nil
}

// Size implements drivers.RowIterator.
func (r *rowIterator) Size(unit drivers.ProgressUnit) (uint64, bool) {
	return 0, false
}

var _ drivers.RowIterator = &rowIterator{}

func (r *rowIterator) setSchema(ctx context.Context) error {
	cts, err := r.rows.ColumnTypes()
	if err != nil {
		return err
	}

	mappers := make([]mapper, len(cts))
	fields := make([]*runtimev1.StructType_Field, len(cts))
	dbTypeNameToMapperMap := getDBTypeNameToMapperMap()

	for i, ct := range cts {
		mapper, ok := dbTypeNameToMapperMap[ct.DatabaseTypeName()]
		if !ok {
			return fmt.Errorf("datatype %q is not supported", ct.DatabaseTypeName())
		}
		mappers[i] = mapper
		runtimeType, err := mapper.runtimeType(ct.ScanType())
		if err != nil {
			return err
		}
		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: runtimeType,
		}
	}

	r.schema = &runtimev1.StructType{Fields: fields}
	r.fieldMappers = mappers
	r.row = make([]sqldriver.Value, len(r.schema.Fields))
	return nil
}

type sourceProperties struct {
	SQL string `mapstructure:"sql"`
	DSN string `mapstructure:"dsn"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory for connector \"mysql\"")
	}
	return conf, err
}
