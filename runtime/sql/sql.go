package sql

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/sql/ast"
	"github.com/rilldata/rill/runtime/sql/rpc"
)

// Transpile transpiles a Rill SQL statement to a target dialect
func Transpile(sql string, dialect rpc.Dialect, catalog []*drivers.CatalogEntry) (string, error) {
	res, err := getIsolate().Request(&rpc.Request{
		Request: &rpc.Request_TranspileRequest{
			TranspileRequest: &rpc.TranspileRequest{
				Sql:     sql,
				Dialect: dialect,
				Catalog: marshalCatalog(dialect, catalog),
			},
		},
	})
	if err != nil {
		// not a user error
		panic(err)
	}

	if res.Error != nil {
		return "", errors.New(res.Error.StackTrace)
	}

	tr := res.GetTranspileResponse()
	if tr == nil {
		panic(fmt.Errorf("expected TranspileRequest to return TranspileResponse"))
	}

	return tr.Sql, nil
}

// Parse parses and validates a Rill SQL statement
func Parse(sql string, dialect rpc.Dialect, catalog []*drivers.CatalogEntry) (*ast.SqlNodeProto, error) {
	res, err := getIsolate().Request(&rpc.Request{
		Request: &rpc.Request_ParseRequest{
			ParseRequest: &rpc.ParseRequest{
				Sql:     sql,
				Catalog: marshalCatalog(dialect, catalog),
			},
		},
	})
	if err != nil {
		// not a user error
		panic(err)
	}

	if res.Error != nil {
		return nil, errors.New(res.Error.Message)
	}

	pr := res.GetParseResponse()
	if pr == nil {
		panic(fmt.Errorf("expected ParseRequest to return ParseResponse"))
	}

	return pr.Ast, nil
}

// See getIsolate
var isolate *Isolate
var isolateOnce sync.Once

// getIsolate returns a lazily-loaded Isolate, which will never be closed.
// If the performance of using a single thread-bound isolate suffers, we should
// consider instead using a pool of isolates.
func getIsolate() *Isolate {
	isolateOnce.Do(func() {
		isolate = OpenIsolate()
	})
	return isolate
}

// marshalCatalog serializes runtime catalog objects to the catalog format expected by the SQL library.
// See sql/src/test/resources for schema example.
func marshalCatalog(dialect rpc.Dialect, objs []*drivers.CatalogEntry) string {
	var artifacts []map[string]any
	var tables []map[string]any
	for _, obj := range objs {
		switch obj.Type {
		case drivers.ObjectTypeMetricsView:
			panic(fmt.Errorf("not implemented"))
		case drivers.ObjectTypeTable, drivers.ObjectTypeSource:
			var schema *runtimev1.StructType
			if obj.Type == drivers.ObjectTypeTable {
				schema = obj.GetTable().Schema
			} else if obj.Type == drivers.ObjectTypeSource {
				schema = obj.GetSource().Schema
			} else {
				panic(fmt.Errorf("not reachable"))
			}
			columns := make([]map[string]any, len(schema.Fields))
			for i, f := range schema.Fields {
				columns[i] = map[string]any{
					"name": f.Name,
					"type": typeCodeToSQLType(f.Type.Code),
				}
			}
			tables = append(tables, map[string]any{
				"name":    obj.Name,
				"columns": columns,
			})
		default:
			panic(fmt.Errorf("unhandled catalog type '%v'", obj.Type))
		}
	}

	var schema string
	switch dialect {
	case rpc.Dialect_DRUID:
		schema = "druid"
	case rpc.Dialect_DUCKDB:
		schema = "main"
	default:
		panic(fmt.Errorf("unhandled dialect: %s", dialect.String()))
	}

	catalog := map[string]any{
		"artifacts": artifacts,
		"schemas": []map[string]any{
			{
				"name":   schema,
				"tables": tables,
			},
		},
	}

	data, err := json.Marshal(catalog)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func typeCodeToSQLType(t runtimev1.Type_Code) string {
	switch t {
	case runtimev1.Type_CODE_BOOL:
		return "BOOLEAN"
	case runtimev1.Type_CODE_INT8:
		return "TINYINT"
	case runtimev1.Type_CODE_INT16:
		return "SMALLINT"
	case runtimev1.Type_CODE_INT32:
		return "INTEGER"
	case runtimev1.Type_CODE_INT64:
		return "BIGINT"
	case runtimev1.Type_CODE_INT128:
		return "HUGEINT"
	case runtimev1.Type_CODE_UINT8:
		return "UTINYINT"
	case runtimev1.Type_CODE_UINT16:
		return "USMALLINT"
	case runtimev1.Type_CODE_UINT32:
		return "UINTEGER"
	case runtimev1.Type_CODE_UINT64:
		return "UBIGINT"
	case runtimev1.Type_CODE_UINT128:
		return "HUGEINT"
	case runtimev1.Type_CODE_FLOAT32:
		return "FLOAT"
	case runtimev1.Type_CODE_FLOAT64:
		return "DOUBLE"
	case runtimev1.Type_CODE_TIMESTAMP:
		return "TIMESTAMP"
	case runtimev1.Type_CODE_DATE:
		return "DATE"
	case runtimev1.Type_CODE_TIME:
		return "TIME"
	case runtimev1.Type_CODE_STRING:
		return "VARCHAR"
	case runtimev1.Type_CODE_BYTES:
		return "BLOB"
	case runtimev1.Type_CODE_ARRAY:
		return "LIST"
	case runtimev1.Type_CODE_STRUCT:
		return "STRUCT"
	case runtimev1.Type_CODE_MAP:
		return "MAP"
	case runtimev1.Type_CODE_DECIMAL:
		return "DECIMAL"
	case runtimev1.Type_CODE_JSON:
		return "VARCHAR"
	case runtimev1.Type_CODE_UUID:
		return "VARCHAR"
	default:
		return "ANY" // TODO: verify
	}
}
