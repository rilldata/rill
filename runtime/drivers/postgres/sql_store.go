package postgres

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
		return nil, fmt.Errorf("the property 'database_url' is required for Postgres. Provide 'database_url' in the YAML properties or pass '--env connector.postgres.database_url=...' to 'rill start'")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}

	res, err := conn.Query(ctx, srcProps.SQL)
	if err != nil {
		conn.Release()
		pool.Close()
		return nil, err
	}

	iter := &rowIterator{
		conn: conn,
		rows: res,
		pool: pool,
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
	conn   *pgxpool.Conn
	rows   pgx.Rows
	pool   *pgxpool.Pool
	schema *runtimev1.StructType

	row          []sqldriver.Value
	fieldMappers []mapper
}

// Close implements drivers.RowIterator.
func (r *rowIterator) Close() error {
	r.rows.Close()
	r.conn.Release()
	r.pool.Close()
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

func (r *rowIterator) setSchema(ctx context.Context) error {
	fds := r.rows.FieldDescriptions()
	conn := r.rows.Conn()
	if conn == nil {
		// not possible but keeping it for graceful failures
		return fmt.Errorf("nil pgx conn")
	}

	mappers := make([]mapper, len(fds))
	fields := make([]*runtimev1.StructType_Field, len(fds))
	typeMap := conn.TypeMap()
	oidToMapperMap := getOidToMapperMap()

	var newConn *pgxpool.Conn
	defer func() {
		if newConn != nil {
			newConn.Release()
		}
	}()
	for i, fd := range fds {
		dt := columnTypeDatabaseTypeName(typeMap, fds[i].DataTypeOID)
		if dt == "" {
			var err error
			if newConn == nil {
				newConn, err = r.acquireConn(ctx)
				if err != nil {
					return err
				}
			}
			dt, err = r.registerIfEnum(ctx, newConn.Conn(), oidToMapperMap, fds[i].DataTypeOID)
			if err != nil {
				return err
			}
		}
		mapper, ok := oidToMapperMap[dt]
		if !ok {
			return fmt.Errorf("datatype %q is not supported", dt)
		}
		mappers[i] = mapper
		fields[i] = &runtimev1.StructType_Field{
			Name: fd.Name,
			Type: mapper.runtimeType(),
		}
	}

	r.schema = &runtimev1.StructType{Fields: fields}
	r.fieldMappers = mappers
	r.row = make([]sqldriver.Value, len(r.schema.Fields))
	return nil
}

func (r *rowIterator) registerIfEnum(ctx context.Context, conn *pgx.Conn, oidToMapperMap map[string]mapper, oid uint32) (string, error) {
	// custom datatypes are not supported
	// but it is possible to support enum with this approach
	var isEnum bool
	var typName string
	err := conn.QueryRow(ctx, "SELECT typtype = 'e' AS isEnum, typname FROM pg_type WHERE oid = $1", oid).Scan(&isEnum, &typName)
	if err != nil {
		return "", err
	}

	if !isEnum {
		return "", fmt.Errorf("custom datatypes are not supported")
	}

	dataType, err := conn.LoadType(ctx, typName)
	if err != nil {
		return "", err
	}

	r.rows.Conn().TypeMap().RegisterType(dataType)
	oidToMapperMap[typName] = &charMapper{}
	register(oidToMapperMap, typName, &charMapper{})
	return typName, nil
}

func (r *rowIterator) acquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	// acquire another connection
	ctxWithTimeOut, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	conn, err := r.pool.Acquire(ctxWithTimeOut)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("postgres connector require 2 connections. Set `max_connections` to atleast 2")
		}
		return nil, err
	}
	return conn, nil
}

// columnTypeDatabaseTypeName returns the database system type name. If the name is unknown the OID is returned.
func columnTypeDatabaseTypeName(typeMap *pgtype.Map, datatypeOID uint32) string {
	if dt, ok := typeMap.TypeForOID(datatypeOID); ok {
		return strings.ToLower(dt.Name)
	}
	return ""
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
