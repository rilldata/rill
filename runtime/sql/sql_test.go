package sql

import (
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/sql/rpc"
	"github.com/stretchr/testify/require"
)

func TestTranspileSelect(t *testing.T) {
	sql := fmt.Sprintf(`select %d as foo, 'hello' as bar, h1.id, h1."power", h2.name from main.heroes h1 join main.heroes h2 on h1.id = h2.id`, 10)
	catalog := []*drivers.CatalogEntry{
		{
			Name: "heroes",
			Type: drivers.ObjectTypeTable,
			Object: &runtimev1.Table{
				Name:    "heroes",
				Managed: false,
				Schema: &runtimev1.StructType{
					Fields: []*runtimev1.StructType_Field{
						{Name: "id", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "birthday", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}},
						{Name: "power", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "name", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "level", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}},
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
	catalog := []*drivers.CatalogEntry{
		{
			Name: "heroes",
			Type: drivers.ObjectTypeTable,
			Object: &runtimev1.Table{
				Name:    "heroes",
				Managed: false,
				Schema: &runtimev1.StructType{
					Fields: []*runtimev1.StructType_Field{
						{Name: "id", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "birthday", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}},
						{Name: "power", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "name", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "level", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}},
					},
				},
			},
		},
	}

	res, err := Parse(sql, rpc.Dialect_DUCKDB, catalog)
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

	catalog := []*drivers.CatalogEntry{
		{
			Name: "heroes",
			Type: drivers.ObjectTypeTable,
			Object: &runtimev1.Table{
				Name:    "heroes",
				Managed: false,
				Schema: &runtimev1.StructType{
					Fields: []*runtimev1.StructType_Field{
						{Name: "id", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "birthday", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}},
						{Name: "power", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "name", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
						{Name: "level", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}},
					},
				},
			},
		},
	}

	res, err := Parse(sql, rpc.Dialect_DUCKDB, catalog)
	require.NoError(t, err)
	require.Equal(t, "SUM", res.GetSqlCreateMetricsViewProto().Measures.List[3].GetSqlBasicCallProto().OperandList[0].GetSqlBasicCallProto().Operator.Name)
}
