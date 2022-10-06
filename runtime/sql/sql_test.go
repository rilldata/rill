package sql

import (
	"testing"

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
