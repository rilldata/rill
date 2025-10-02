package bigquery

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOLAP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Load .env file at the repo root (if any)
	_, currentFile, _, _ := goruntime.Caller(0)
	fmt.Println(currentFile)
	envPath := filepath.Join(currentFile, "..", "..", "..", "..", ".env")
	_, err := os.Stat(envPath)
	if err == nil {
		require.NoError(t, godotenv.Load(envPath))
	}

	gac := os.Getenv("RILL_RUNTIME_BIGQUERY_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON")
	require.NotEmpty(t, gac, "Bigquery RILL_RUNTIME_BIGQUERY_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON not configured")

	h, err := driver{}.Open("default", map[string]any{"google_application_credentials": gac}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	olap, ok := h.AsOLAP("default")
	require.True(t, ok)

	tests := []struct {
		query  string
		result map[string]any
	}{
		{
			"SELECT [true, false, true] AS booleans;",
			map[string]any{"booleans": "[true,false,true]"},
		},
		{
			"SELECT GENERATE_ARRAY(21, 14, -1) AS countdown;",
			map[string]any{"countdown": "[21,20,19,18,17,16,15,14]"},
		},
		{
			"SELECT true AS bool",
			map[string]any{"bool": true},
		},
		{
			"SELECT CAST('2021-01-01' AS DATE) AS date",
			map[string]any{"date": "2021-01-01"},
		},
		{
			"SELECT CAST('2025-01-31 23:59:59.999999' AS DATETIME) AS datetime;",
			map[string]any{"datetime": "2025-01-31T23:59:59.999999000"},
		},
		{
			`select JSON '{  "id": 10,  "type": "fruit",  "name": "apple",  "on_menu": true,  "recipes":    {      "salads":      [        { "id": 2001, "type": "Walnut Apple Salad" },        { "id": 2002, "type": "Apple Spinach Salad" }      ],      "desserts":      [        { "id": 3001, "type": "Apple Pie" },        { "id": 3002, "type": "Apple Scones" },        { "id": 3003, "type": "Apple Crumble" }      ]    }}' AS json`,
			map[string]any{"json": `{"id":10,"name":"apple","on_menu":true,"recipes":{"desserts":[{"id":3001,"type":"Apple Pie"},{"id":3002,"type":"Apple Scones"},{"id":3003,"type":"Apple Crumble"}],"salads":[{"id":2001,"type":"Walnut Apple Salad"},{"id":2002,"type":"Apple Spinach Salad"}]},"type":"fruit"}`},
		},
		{
			"SELECT 9223372036854775807 AS integer",
			map[string]any{"integer": int64(9223372036854775807)},
		},
		{
			"SELECT cast(9.9999999999999999999999999999999999999E+28 as NUMERIC) as number",
			map[string]any{"number": "99999999999999999999999999999.999999999"},
		},
		{
			"SELECT cast(0.1 as NUMERIC) as number",
			map[string]any{"number": "0.1"},
		},
		{
			"SELECT cast(5.7896044618658097711785492504343953926634992332820282019728792003956564819967E+38 as BIGNUMERIC) as number",
			map[string]any{"number": "578960446186580977117854925043439539266.34992332820282019728792003956564819967"},
		},
		{
			"SELECT cast(3.14 as FLOAT64) as number",
			map[string]any{"number": 3.14},
		},
		{
			"SELECT RANGE(Date'2020-01-01', Date'2025-01-01') AS date_range",
			map[string]any{"date_range": "[2020-01-01, 2025-01-01)"},
		},
		{
			"SELECT STRUCT(1 AS a, 'abc' AS b) as str",
			map[string]any{"str": `{"a":1,"b":"abc"}`},
		},
		{
			"SELECT TIME'23:59:59.999999' AS t",
			map[string]any{"t": "23:59:59.999999000"},
		},
		{
			"SELECT TIMESTAMP'2025-01-01 23:59:59.999999 UTC' AS t",
			map[string]any{"t": time.Date(2025, 1, 1, 23, 59, 59, 999999000, time.UTC)},
		},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			rows, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query})
			require.NoError(t, err)
			defer rows.Close()
			require.True(t, rows.Next())
			res := make(map[string]any)
			err = rows.MapScan(res)
			require.NoError(t, err)
			require.Equal(t, test.result, res)
		})
	}
}
