package sql

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/rilldata/rill/runtime/sql/ast"
	"github.com/rilldata/rill/runtime/sql/rpc"
)

// Transpile transpiles a Rill SQL statement to a target dialect
func Transpile(sql string, dialect rpc.Dialect, catalog map[string]any) (string, error) {
	res, err := getIsolate().Request(&rpc.Request{
		Request: &rpc.Request_TranspileRequest{
			TranspileRequest: &rpc.TranspileRequest{
				Sql:     sql,
				Dialect: dialect,
				Catalog: marshalCatalog(catalog),
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
func Parse(sql string, catalog map[string]any) (*ast.SqlNodeProto, error) {
	res, err := getIsolate().Request(&rpc.Request{
		Request: &rpc.Request_ParseRequest{
			ParseRequest: &rpc.ParseRequest{
				Sql:     sql,
				Catalog: marshalCatalog(catalog),
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

// marshalCatalog serializes a runtime catalog to the catalog format expected by the SQL library
func marshalCatalog(catalog map[string]any) string {
	data, err := json.Marshal(catalog)
	if err != nil {
		panic(err)
	}
	return string(data)
}
