package sql

import (
	"testing"

	"github.com/rilldata/rill/runtime/sql/rpc"
	"github.com/stretchr/testify/require"
)

func TestSanity(t *testing.T) {
	isolate := NewIsolate()

	sql := "select 1 as foo, 'hello' as bar"
	catalog := `{ "schemas": [], "artifacts": [] }`
	dialect := "duckdb"

	res := isolate.ConvertSQL(sql, catalog, dialect)
	require.Equal(t, `SELECT 1 AS "FOO", 'hello' AS "BAR"`, res)

	err := isolate.Close()
	require.NoError(t, err)
}

func TestTranspile(t *testing.T) {
	isolate := NewIsolate()

	sql := "select 1 as foo, 'hello' as bar"

	r := rpc.Request{
		Request: &rpc.Request_TranspileRequest{
			TranspileRequest: &rpc.TranspileRequest{
				Sql:     sql,
				Dialect: rpc.Dialect_DUCKDB,
				Catalog: `{ "schemas": [], "artifacts": [] }`,
			},
		},
	}

	res := isolate.request(&r)

	require.Equal(t, `SELECT 1 AS "FOO", 'hello' AS "BAR"`, (*res).GetTranspileResponse().Sql)

	err := isolate.Close()
	require.NoError(t, err)
}

func TestTranspileNoBase64(t *testing.T) {
	isolate := NewIsolate()

	sql := "select 1 as foo, 'hello' as bar"

	r := rpc.Request{
		Request: &rpc.Request_TranspileRequest{
			TranspileRequest: &rpc.TranspileRequest{
				Sql:     sql,
				Dialect: rpc.Dialect_DUCKDB,
				Catalog: `{ "schemas": [], "artifacts": [] }`,
			},
		},
	}

	res := isolate.requestNoBase64(&r)

	require.Equal(t, `SELECT 1 AS "FOO", 'hello' AS "BAR"`, (*res).GetTranspileResponse().Sql)

	err := isolate.Close()
	require.NoError(t, err)
}

func TestSanityGetAST(t *testing.T) {
	isolate := NewIsolate()

	sql := "select 1 as foo, 'hello' as bar"
	catalog := `{ "schemas": [], "artifacts": [] }`

	res := isolate.getAST(sql, catalog)
	println(res)
	println(len(res))

	err := isolate.Close()
	require.NoError(t, err)
}
