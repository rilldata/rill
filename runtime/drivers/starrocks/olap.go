package starrocks

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

var _ drivers.OLAPStore = (*connection)(nil)

// StarRocks constants
const (
	// defaultCatalog is the default catalog name in StarRocks
	defaultCatalog = "default_catalog"

	// Database type names
	dbTypeTinyInt   = "TINYINT"
	dbTypeBinary    = "BINARY"
	dbTypeVarBinary = "VARBINARY"
)

// isBinaryType checks if the database type is a binary type
func isBinaryType(dbType string) bool {
	return dbType == dbTypeBinary || dbType == dbTypeVarBinary
}

// Dialect implements drivers.OLAPStore.
func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectStarRocks
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false // StarRocks instances are typically always running
}

// WithConnection implements drivers.OLAPStore.
func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	// StarRocks supports connection affinity for temp tables
	db, err := c.getDB(ctx)
	if err != nil {
		return fmt.Errorf("get db connection: %w", err)
	}

	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	// Set catalog and database context
	if err := c.setCatalogContext(ctx, conn); err != nil {
		return err
	}

	// Create wrapped context with connection
	wrappedCtx := context.WithValue(ctx, connCtxKey{}, conn)
	ensuredCtx := context.WithValue(context.Background(), connCtxKey{}, conn)

	return fn(wrappedCtx, ensuredCtx)
}

// connCtxKey is used to store connection in context for WithConnection.
type connCtxKey struct{}

// Exec implements drivers.OLAPStore.
func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	if stmt.DryRun {
		return nil
	}
	res, err := c.Query(ctx, stmt)
	if err != nil {
		return err
	}
	if res != nil {
		return res.Close()
	}
	return nil
}

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, resErr error) {
	if c.logQueries {
		c.logger.Info("StarRocks query",
			zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)),
			zap.Any("args", stmt.Args),
			observability.ZapCtx(ctx))
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Check if we have a connection from WithConnection (already has catalog context set)
	if conn, ok := ctx.Value(connCtxKey{}).(*sqlx.Conn); ok && conn != nil {
		// Handle dry run with EXPLAIN
		if stmt.DryRun {
			_, err = conn.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
			return nil, err
		}

		rows, err := conn.QueryxContext(ctx, stmt.Query, stmt.Args...)
		if err != nil {
			return nil, err
		}
		defer func() {
			if resErr != nil {
				_ = rows.Close()
			}
		}()

		// Infer schema from query result and get column types
		schema, cts, err := rowsToSchemaWithTypes(rows)
		if err != nil {
			return nil, fmt.Errorf("infer schema from query result: %w", err)
		}

		starrocksRows := &starrocksRows{
			Rows:     rows,
			scanDest: prepareScanDest(schema),
			colTypes: cts,
		}
		res = &drivers.Result{Rows: starrocksRows, Schema: schema}
		return res, nil
	}

	// Always use dedicated connection with context to ensure catalog/database is set
	// This is necessary because:
	// 1. External catalogs don't include database in DSN
	// 2. Connection pool connections may not have context set
	// 3. StarRocks requires explicit database context for queries
	conn, err := db.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	if err := c.setCatalogContext(ctx, conn); err != nil {
		conn.Close()
		return nil, err
	}

	// Handle dry run with EXPLAIN
	if stmt.DryRun {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		conn.Close()
		return nil, err
	}

	rows, err := conn.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		conn.Close()
		return nil, err
	}
	defer func() {
		if resErr != nil {
			_ = rows.Close()
			conn.Close()
		}
	}()

	// Infer schema from query result and get column types
	schema, cts, err := rowsToSchemaWithTypes(rows)
	if err != nil {
		return nil, fmt.Errorf("infer schema from query result: %w", err)
	}

	starrocksRows := &starrocksRows{
		Rows:     rows,
		scanDest: prepareScanDest(schema),
		colTypes: cts,
		conn:     conn, // Store connection to close when rows are closed
	}
	res = &drivers.Result{Rows: starrocksRows, Schema: schema}
	return res, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	// Execute with LIMIT 0 to get schema without returning data
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Always use dedicated connection with context to ensure catalog/database is set
	conn, err := db.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	if err := c.setCatalogContext(ctx, conn); err != nil {
		return nil, err
	}

	// Execute with LIMIT 0 to get schema (only add if not already present)
	finalQuery := query
	if !strings.Contains(strings.ToUpper(query), "LIMIT") {
		finalQuery = query + " LIMIT 0"
	}
	rows, err := conn.QueryxContext(ctx, finalQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToSchema(rows)
}

// setCatalogContext sets the catalog and database context on a connection.
// Handles both external catalogs (Iceberg, Hive) and default_catalog.
// For external catalogs: SET CATALOG → USE database
// For default_catalog: USE database only
func (c *connection) setCatalogContext(ctx context.Context, conn *sqlx.Conn) error {
	return switchCatalogContext(ctx, conn, c.configProp.Catalog, c.configProp.Database)
}

// InformationSchema implements drivers.OLAPStore.
func (c *connection) InformationSchema() drivers.OLAPInformationSchema {
	return c
}

// All implements drivers.OLAPInformationSchema.
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	// StarRocks doesn't easily expose physical size per table in information_schema
	// This could be extended to query system tables if needed
	return nil
}

// Lookup implements drivers.OLAPInformationSchema.
// StarRocks mapping: db = catalog, schema = database, name = table
func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	// db parameter is the catalog (e.g., default_catalog)
	// schema parameter is the database name (e.g., sales, analytics)

	// Use default database if schema is empty
	if schema == "" {
		schema = c.configProp.Database
	}

	// Determine catalog
	catalog := db
	if catalog == "" {
		catalog = c.configProp.Catalog
		if catalog == "" {
			catalog = defaultCatalog
		}
	}

	// GetTable: database param = catalog, databaseSchema param = database
	meta, err := c.GetTable(ctx, catalog, schema, name)
	if err != nil {
		return nil, fmt.Errorf("get table metadata for %s.%s.%s: %w", catalog, schema, name, err)
	}

	rtSchema := &runtimev1.StructType{}
	for colName, colType := range meta.Schema {
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: databaseTypeToPB(colType, true),
		})
	}

	return &drivers.OlapTable{
		Database:          catalog, // Catalog name (e.g., default_catalog, iceberg_catalog)
		DatabaseSchema:    schema,  // Database name (e.g., sales, analytics)
		Name:              name,
		View:              meta.View,
		Schema:            rtSchema,
		UnsupportedCols:   nil,
		PhysicalSizeBytes: 0,
	}, nil
}

// rowsToSchemaWithTypes extracts schema and column types from query result.
// Returns both to avoid calling ColumnTypes() twice.
func rowsToSchemaWithTypes(r *sqlx.Rows) (*runtimev1.StructType, []*sql.ColumnType, error) {
	if r == nil {
		return nil, nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: databaseTypeToPB(ct.DatabaseTypeName(), nullable),
		}
	}

	return &runtimev1.StructType{Fields: fields}, cts, nil
}

// rowsToSchema extracts schema from query result.
// Used by QuerySchema where we don't need column types.
func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	schema, _, err := rowsToSchemaWithTypes(r)
	return schema, err
}

// StarRocks type mapping to Rill type codes
// Reference: https://docs.starrocks.io/docs/sql-reference/data-types/
var starRocksTypeMap = map[string]runtimev1.Type_Code{
	// Boolean
	"BOOLEAN": runtimev1.Type_CODE_BOOL,
	"BOOL":    runtimev1.Type_CODE_BOOL,

	// Integer types
	"TINYINT":  runtimev1.Type_CODE_INT8,
	"SMALLINT": runtimev1.Type_CODE_INT16,
	"INT":      runtimev1.Type_CODE_INT32,
	"INTEGER":  runtimev1.Type_CODE_INT32,
	"BIGINT":   runtimev1.Type_CODE_INT64,

	// LARGEINT: 16-byte signed integer, range [-2^127+1, 2^127-1]
	// NOTE: StarRocks information_schema reports LARGEINT columns as "BIGINT UNSIGNED"
	// This is a known behavior - both strings map to the same INT128 type
	"BIGINT UNSIGNED": runtimev1.Type_CODE_INT128,
	"LARGEINT":        runtimev1.Type_CODE_INT128,

	// Floating point types
	"FLOAT":  runtimev1.Type_CODE_FLOAT32,
	"DOUBLE": runtimev1.Type_CODE_FLOAT64,

	// Decimal types
	// Fast DECIMAL (DECIMAL64/DECIMAL128): Default in v4.0+, variable-width integers
	//   - DECIMAL64:  Precision [1-18], stored as int64
	//   - DECIMAL128: Precision [19-38], stored as int128
	// DECIMAL256: v4.0+, Precision (38-76], stored as int256
	// DECIMALV2: Legacy implementation (pre-v4.0)
	"DECIMAL":    runtimev1.Type_CODE_DECIMAL,
	"DECIMALV2":  runtimev1.Type_CODE_DECIMAL,
	"DECIMAL32":  runtimev1.Type_CODE_DECIMAL,
	"DECIMAL64":  runtimev1.Type_CODE_DECIMAL,
	"DECIMAL128": runtimev1.Type_CODE_DECIMAL,
	"DECIMAL256": runtimev1.Type_CODE_DECIMAL,

	// String types
	"CHAR":    runtimev1.Type_CODE_STRING,
	"VARCHAR": runtimev1.Type_CODE_STRING,
	"STRING":  runtimev1.Type_CODE_STRING,
	"TEXT":    runtimev1.Type_CODE_STRING,

	// Binary types
	"BINARY":    runtimev1.Type_CODE_BYTES,
	"VARBINARY": runtimev1.Type_CODE_BYTES,

	// Date/Time types (StarRocks does NOT support TIME type)
	"DATE":     runtimev1.Type_CODE_DATE,
	"DATETIME": runtimev1.Type_CODE_TIMESTAMP,

	// Semi-structured types
	"JSON":   runtimev1.Type_CODE_JSON,
	"JSONB":  runtimev1.Type_CODE_JSON,
	"ARRAY":  runtimev1.Type_CODE_ARRAY,
	"MAP":    runtimev1.Type_CODE_MAP,
	"STRUCT": runtimev1.Type_CODE_STRUCT,

	// Special types (stored as strings)
	"HLL":        runtimev1.Type_CODE_STRING, // HyperLogLog
	"BITMAP":     runtimev1.Type_CODE_STRING, // Bitmap
	"PERCENTILE": runtimev1.Type_CODE_STRING, // Percentile

	// NULL type
	"NULL": runtimev1.Type_CODE_UNSPECIFIED,
}

// databaseTypeToPB converts StarRocks database types to Rill's generic schema type.
func databaseTypeToPB(dbt string, nullable bool) *runtimev1.Type {
	// Normalize type name
	upperDbt := strings.ToUpper(dbt)

	// Handle parameterized types (e.g., DECIMAL(10,2), VARCHAR(255))
	baseDbt := upperDbt
	if idx := strings.Index(upperDbt, "("); idx != -1 {
		baseDbt = upperDbt[:idx]
	}

	// Lookup type code
	code, ok := starRocksTypeMap[baseDbt]
	if !ok {
		code = runtimev1.Type_CODE_UNSPECIFIED
	}

	return &runtimev1.Type{
		Code:     code,
		Nullable: nullable,
	}
}

// starrocksRows wraps sqlx.Rows to provide MapScan method with proper type handling.
// This is required because sqlx scans into any, and MySQL driver returns []byte for DECIMAL
// and other types when not given a specific type to scan into.
type starrocksRows struct {
	*sqlx.Rows
	scanDest []any
	colTypes []*sql.ColumnType
	conn     *sqlx.Conn // Optional connection to close when rows are closed (for external catalogs)
}

// Close closes the rows and releases any associated resources.
// For external catalog queries, this also closes the dedicated connection.
func (r *starrocksRows) Close() error {
	err := r.Rows.Close()
	if r.conn != nil {
		connErr := r.conn.Close()
		if err == nil {
			err = connErr
		}
	}
	return err
}

func (r *starrocksRows) MapScan(dest map[string]any) error {
	err := r.Rows.Scan(r.scanDest...)
	if err != nil {
		return err
	}
	for i, ct := range r.colTypes {
		fieldName := ct.Name()
		valPtr := r.scanDest[i]
		if valPtr == nil {
			dest[fieldName] = nil
			continue
		}

		// Handle BINARY/VARBINARY - base64 encode binary data
		dbType := strings.ToUpper(ct.DatabaseTypeName())
		if isBinaryType(dbType) {
			if ns, ok := valPtr.(*sql.NullString); ok {
				if ns.Valid {
					// Convert string back to bytes and base64 encode
					dest[fieldName] = base64.StdEncoding.EncodeToString([]byte(ns.String))
				} else {
					dest[fieldName] = nil
				}
				continue
			}
		}

		switch valPtr := valPtr.(type) {
		case *sql.NullBool:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Bool
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullInt16:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Int16
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullInt32:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Int32
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullInt64:
			if valPtr.Valid {
				// TINYINT is signed and scanned as int64, convert to int8
				if dbType == dbTypeTinyInt {
					dest[fieldName] = int8(valPtr.Int64)
				} else {
					dest[fieldName] = valPtr.Int64
				}
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullFloat64:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Float64
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullString:
			if valPtr.Valid {
				dest[fieldName] = valPtr.String
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullTime:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Time
			} else {
				dest[fieldName] = nil
			}
		default:
			// Safe type assertion for unknown types
			if ptr, ok := valPtr.(*any); ok && ptr != nil {
				dest[fieldName] = *ptr
			} else {
				dest[fieldName] = nil
			}
		}
	}
	return nil
}

// scanDestFactory maps Rill type codes to SQL null type constructors
var scanDestFactory = map[runtimev1.Type_Code]func() any{
	runtimev1.Type_CODE_BOOL:      func() any { return &sql.NullBool{} },
	runtimev1.Type_CODE_INT8:      func() any { return &sql.NullInt64{} }, // TINYINT is signed, MySQL returns int64
	runtimev1.Type_CODE_INT16:     func() any { return &sql.NullInt16{} },
	runtimev1.Type_CODE_INT32:     func() any { return &sql.NullInt32{} },
	runtimev1.Type_CODE_INT64:     func() any { return &sql.NullInt64{} },
	runtimev1.Type_CODE_FLOAT32:   func() any { return &sql.NullFloat64{} }, // MySQL doesn't have NullFloat32
	runtimev1.Type_CODE_FLOAT64:   func() any { return &sql.NullFloat64{} },
	runtimev1.Type_CODE_DECIMAL:   func() any { return &sql.NullString{} }, // Scan as string to avoid []byte
	runtimev1.Type_CODE_STRING:    func() any { return &sql.NullString{} },
	runtimev1.Type_CODE_BYTES:     func() any { return &sql.NullString{} },
	runtimev1.Type_CODE_DATE:      func() any { return &sql.NullTime{} },
	runtimev1.Type_CODE_TIMESTAMP: func() any { return &sql.NullTime{} },
	runtimev1.Type_CODE_JSON:      func() any { return &sql.NullString{} },
	runtimev1.Type_CODE_INT128:    func() any { return &sql.NullString{} }, // LARGEINT as string
	runtimev1.Type_CODE_ARRAY:     func() any { return &sql.NullString{} }, // ARRAY as JSON string
	runtimev1.Type_CODE_MAP:       func() any { return &sql.NullString{} }, // MAP as JSON string
	runtimev1.Type_CODE_STRUCT:    func() any { return &sql.NullString{} }, // STRUCT as JSON string
}

// prepareScanDest creates scan destinations for each field in the schema
func prepareScanDest(schema *runtimev1.StructType) []any {
	scanList := make([]any, len(schema.Fields))
	for i, field := range schema.Fields {
		if factory, ok := scanDestFactory[field.Type.Code]; ok {
			scanList[i] = factory()
		} else {
			scanList[i] = new(any)
		}
	}
	return scanList
}
