package mysql

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"errors"
	"fmt"

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
		return nil, fmt.Errorf("the property 'dsn' is required for MySQL. Provide 'dsn' in the YAML properties or pass '--var connector.mysql.dsn=...' to 'rill start'")
	}

	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	conf.ParseTime = true // if set to false, time is scanned as an array rather than as time.Time

	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
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

	if err := iter.setSchema(); err != nil {
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
	fieldDests   []any // Destinations are used while scanning rows
	columnTypes  []*sql.ColumnType
}

// Close implements drivers.RowIterator.
func (r *rowIterator) Close() error {
	r.rows.Close()
	r.db.Close()
	return nil
}

// Next implements drivers.RowIterator.
func (r *rowIterator) Next(ctx context.Context) ([]sqldriver.Value, error) {
	var err error
	if !r.rows.Next() {
		err := r.rows.Err()
		if err == nil {
			return nil, drivers.ErrIteratorDone
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, drivers.ErrNoRows
		}
		return nil, err
	}

	// Scan expects destinations to be pointers
	for i := range r.fieldDests {
		r.fieldDests[i], err = r.fieldMappers[i].dest(r.columnTypes[i].ScanType())
		if err != nil {
			return nil, err
		}
	}

	if err := r.rows.Scan(r.fieldDests...); err != nil {
		return nil, err
	}

	for i := range r.schema.Fields {
		// Dereference destinations and fill the row
		r.row[i], err = r.fieldMappers[i].value(r.fieldDests[i])
		if err != nil {
			return nil, err
		}
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

func (r *rowIterator) setSchema() error {
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
	r.row = make([]sqldriver.Value, len(r.schema.Fields))
	r.fieldMappers = mappers
	r.fieldDests = make([]any, len(r.schema.Fields))
	r.columnTypes, err = r.rows.ColumnTypes()
	if err != nil {
		return err
	}

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
