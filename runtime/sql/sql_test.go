package sql

import (
	"testing"

	"github.com/rilldata/rill/runtime/sql/rpc"
	"github.com/stretchr/testify/require"
)

func TestSanity(t *testing.T) {
	isolate := NewIsolate()

	sql := "select 1 as foo, 'hello' as bar"
	schema := `{ "tables": [] }`
	dialect := "duckdb"

	res := isolate.ConvertSQL(sql, schema, dialect)
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
				Schema:  `{ "tables": [] }`,
			},
		},
	}

	res := isolate.request(&r)

	require.Equal(t, `SELECT 1 AS "FOO", 'hello' AS "BAR"`, (*res).GetTranspileResponse().Sql)

	err := isolate.Close()
	require.NoError(t, err)
}

func TestSanityGetAST(t *testing.T) {
	isolate := NewIsolate()

	sql := "select 1 as foo, 'hello' as bar"
	schema := `{ "tables": [] }`

	res := isolate.getAST(sql, schema)
	println(res)
	println(len(res))

	err := isolate.Close()
	require.NoError(t, err)
}
