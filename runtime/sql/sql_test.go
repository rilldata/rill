package sql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// func TestTranspileSelect(t *testing.T) {
// 	sql := fmt.Sprintf(`select %d as foo, 'hello' as bar, h1.id, h1."power", h2.name from heroes h1 join heroes h2 on h1.id = h2.id`, 10)
// 	schema := map[string]any{
// 		"tables": []map[string]any{
// 			{
// 				"name": "heroes",
// 				"columns": []map[string]any{
// 					{"name": "id", "type": "varchar"},
// 					{"name": "power", "type": "varchar"},
// 					{"name": "name", "type": "varchar"},
// 				},
// 			},
// 		},
// 	}

// 	res, err := Transpile(sql, rpc.Dialect_DUCKDB, schema)
// 	require.NoError(t, err)
// 	require.Equal(t, res, res)
// }

func TestParseSelect(t *testing.T) {
	sql := fmt.Sprintf(`select 10 as foo, 'hello' as bar, h1.id, h1."power", h2.name from heroes h1 join heroes h2 on h1.id = h2.id`)
	schema := map[string]any{
		"tables": []map[string]any{
			{
				"name": "heroes",
				"columns": []map[string]any{
					{"name": "id", "type": "varchar"},
					{"name": "power", "type": "varchar"},
					{"name": "name", "type": "varchar"},
				},
			},
		},
	}

	res, err := Parse(sql, schema)
	require.NoError(t, err)
	require.Equal(t, 5, len(res.GetSqlSelectProto().GetSelectList().List))
}
