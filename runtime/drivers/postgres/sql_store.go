package postgres

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	if srcProps.DatabaseURL != "" { // get from src properties
		dsn = srcProps.DatabaseURL
	} else if url, ok := c.config["database_url"].(string); ok && url != "" { // get from driver configs
		dsn = url
	} else {
		return nil, fmt.Errorf("the property 'database_url' is required for Postgres. Provide 'database_url' in the YAML properties or pass '--env connectors.postgres.database_url=...' to 'rill start'")
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	res, err := conn.Query(ctx, srcProps.SQL)
	if err != nil {
		conn.Close(ctx)
		return nil, err
	}

	schema, mappers, err := rowsToSchema(res)
	if err != nil {
		res.Close()
		conn.Close(ctx)
		return nil, err
	}

	return &rowIterator{
		conn:         conn,
		rows:         res,
		schema:       schema,
		fieldMappers: mappers,
		row:          make([]sqldriver.Value, len(schema.Fields)),
	}, nil
}

// QueryAsFiles implements drivers.SQLStore
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any, opt *drivers.QueryOption, p drivers.Progress) (drivers.FileIterator, error) {
	return nil, drivers.ErrNotImplemented
}

type rowIterator struct {
	conn   *pgx.Conn
	rows   pgx.Rows
	schema *runtimev1.StructType

	row          []sqldriver.Value
	fieldMappers []mapper
}

// Close implements drivers.RowIterator.
func (r *rowIterator) Close() error {
	r.rows.Close()
	return r.conn.Close(context.Background())
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

	vals, err := r.rows.Values()
	if err != nil {
		return nil, err
	}

	for i := range r.schema.Fields {
		mapper := r.fieldMappers[i]
		if vals[i] == nil {
			r.row[i] = nil
			continue
		}
		r.row[i], err = mapper.value(vals[i])
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

func rowsToSchema(r pgx.Rows) (*runtimev1.StructType, []mapper, error) {
	fds := r.FieldDescriptions()
	conn := r.Conn()
	if conn == nil {
		// not possible but keeping it for graceful failures
		return nil, nil, fmt.Errorf("nil pgx conn")
	}

	mappers := make([]mapper, len(fds))
	fields := make([]*runtimev1.StructType_Field, len(fds))
	typeMap := conn.TypeMap()
	for i, fd := range fds {
		dt, err := columnTypeDatabaseTypeName(typeMap, fds[i].DataTypeOID)
		if err != nil {
			return nil, nil, err
		}
		mapper, ok := oidToMapperMap[dt]
		if !ok {
			return nil, nil, fmt.Errorf("datatype %q is not supported", dt)
		}
		mappers[i] = mapper
		fields[i] = &runtimev1.StructType_Field{
			Name: fd.Name,
			Type: mapper.runtimeType(),
		}
	}

	return &runtimev1.StructType{Fields: fields}, mappers, nil
}

// columnTypeDatabaseTypeName returns the database system type name. If the name is unknown the OID is returned.
func columnTypeDatabaseTypeName(typeMap *pgtype.Map, datatypeOID uint32) (string, error) {
	if dt, ok := typeMap.TypeForOID(datatypeOID); ok {
		return strings.ToLower(dt.Name), nil
	}
	return "", fmt.Errorf("custom datatypes are not supported")
}

type sourceProperties struct {
	SQL         string `mapstructure:"sql"`
	DatabaseURL string `mapstructure:"database_url"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory for connector \"postgres\"")
	}
	return conf, err
}
