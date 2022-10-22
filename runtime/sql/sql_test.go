package sql

import (
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/sql/rpc"
	"github.com/stretchr/testify/require"
)

func TestTranspileSelect(t *testing.T) {
	sql := fmt.Sprintf(`select %d as foo, 'hello' as bar, h1.id, h1."power", h2.name from main.heroes h1 join main.heroes h2 on h1.id = h2.id`, 10)
	catalog := map[string]any{
		"artifacts": []map[string]any{
			{
				"name":    "MV",
				"type":    "METRICS_VIEW",
				"payload": "Create Metrics View MV DIMENSIONS \"power\" Measures count(DISTINCT name) AS names FROM main.heroes",
			},
		},
		"schemas": []map[string]any{
			{
				"name": "main",
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
			},
		},
	}

	res, err := Transpile(sql, rpc.Dialect_DUCKDB, catalog)
	require.NoError(t, err)
	require.Equal(t, res, res)
}

func TestParseSelect(t *testing.T) {
	sql := `select 10 as foo, 'hello' as bar, h1.id, h1."power", h2.name from main.heroes h1 join main.heroes h2 on h1.id = h2.id`
	catalog := map[string]any{
		"artifacts": []any{},
		"schemas": []map[string]any{
			{
				"name": "main",
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
			},
		},
	}

	res, err := Parse(sql, catalog)
	require.NoError(t, err)
	require.Equal(t, 5, len(res.GetSqlSelectProto().GetSelectList().List))
}

func TestParseMetricsView(t *testing.T) {
	sql := `
		CREATE METRICS VIEW FOO_BAR
		DIMENSIONS
			ID,
			BIRTHDAY,
			"POWER",
			NAME
		MEASURES
			COUNT(*) AS "COUNT",
			COUNT(DISTINCT "POWER") AS POWERS,
			COUNT(DISTINCT NAME) AS NAMES,
			SUM(LEVEL) AS LEVES
		FROM MAIN.HEROES
	`
	catalog := map[string]any{
		"artifacts": []any{},
		"schemas": []map[string]any{
			{
				"name": "main",
				"tables": []map[string]any{
					{
						"name": "heroes",
						"columns": []map[string]any{
							{"name": "id", "type": "varchar"},
							{"name": "birthday", "type": "timestamp"},
							{"name": "power", "type": "varchar"},
							{"name": "name", "type": "varchar"},
							{"name": "level", "type": "integer"},
						},
					},
				},
			},
		},
	}

	res, err := Parse(sql, catalog)
	require.NoError(t, err)
	require.Equal(t, "SUM", res.GetSqlCreateMetricsViewProto().Measures.List[3].GetSqlBasicCallProto().OperandList[0].GetSqlBasicCallProto().Operator.Name)
}
