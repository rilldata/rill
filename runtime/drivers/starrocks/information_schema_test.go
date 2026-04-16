package starrocks

import (
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers/starrocks/teststarrocks"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInformationSchemaListTables(t *testing.T) {
	testmode.Expensive(t)

	dsn := teststarrocks.StartWithData(t)
	// Connect with test_db as the database so DATABASE() = 'test_db' returns true
	dsn = strings.Replace(dsn, "/?", "/test_db?", 1)

	conn, err := driver{}.Open("", "default", map[string]any{
		"dsn": dsn,
	}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()

	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	// "default_catalog" is the default catalog in StarRocks; "test_db" is the current database
	tables, _, err := infoSchema.ListTables(t.Context(), "default_catalog", "test_db", 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, tables)

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}
}
