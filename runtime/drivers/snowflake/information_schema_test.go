package snowflake_test

import (
	"context"
	"sort"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

var knownTestTables = []string{"ALL_DATATYPES", "BAR", "BAZ", "FOO", "FOZ", "MODEL"}

const numKnown = 6

func filterOLAP(tables []*drivers.OlapTable) []*drivers.OlapTable {
	known := make(map[string]bool, numKnown)
	for _, n := range knownTestTables {
		known[n] = true
	}
	var out []*drivers.OlapTable
	for _, t := range tables {
		if known[t.Name] {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func filterTableInfos(tables []*drivers.TableInfo) []*drivers.TableInfo {
	// Snowflake returns identifiers in upper-case by default.
	known := make(map[string]bool, numKnown)
	for _, n := range knownTestTables {
		known[n] = true
	}
	var out []*drivers.TableInfo
	for _, t := range tables {
		if known[t.Name] {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func TestInformationSchemaSnowflake(t *testing.T) {
	testmode.Expensive(t)

	conn, olap := acquireTestSnowflake(t)
	ctx := t.Context()
	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	// Resolve current database and schema from the session.
	rows, err := olap.Query(ctx, &drivers.Statement{Query: "SELECT CURRENT_DATABASE(), CURRENT_SCHEMA()"})
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var database, databaseSchema string
	require.NoError(t, rows.Scan(&database, &databaseSchema))
	require.NoError(t, rows.Close())
	require.NotEmpty(t, database)
	require.NotEmpty(t, databaseSchema)

	t.Run("testAll", func(t *testing.T) {
		testAll(t, olap)
	})
	t.Run("testLookup", func(t *testing.T) {
		testLookup(t, olap, database, databaseSchema)
	})
	t.Run("testListDatabaseSchemas", func(t *testing.T) {
		testListDatabaseSchemas(t, infoSchema, database, databaseSchema)
	})
	t.Run("testListTables", func(t *testing.T) {
		testListTables(t, infoSchema, database, databaseSchema)
	})
	t.Run("testGetTable", func(t *testing.T) {
		testGetTable(t, infoSchema, database, databaseSchema)
	})
	t.Run("testListTablesPagination", func(t *testing.T) {
		testListTablesPagination(t, infoSchema, database, databaseSchema)
	})
	t.Run("testLoadDDL", func(t *testing.T) {
		testLoadDDL(t, olap, database, databaseSchema)
	})
}

func testAll(t *testing.T, olap drivers.OLAPStore) {
	all, _, err := olap.InformationSchema().All(context.Background(), "", 0, "")
	require.NoError(t, err)
	tables := filterOLAP(all)
	require.Equal(t, numKnown, len(tables))

	// Tables are sorted alphabetically (Snowflake uppercases identifiers)
	require.Equal(t, "ALL_DATATYPES", tables[0].Name)
	require.Equal(t, "BAR", tables[1].Name)
	require.Equal(t, "BAZ", tables[2].Name)
	require.Equal(t, "FOO", tables[3].Name)
	require.Equal(t, "FOZ", tables[4].Name)
	require.Equal(t, "MODEL", tables[5].Name)

	require.True(t, tables[5].View, "MODEL should be a view")
	for _, tbl := range tables[:5] {
		require.False(t, tbl.View, "table %s should not be a view", tbl.Name)
	}
}

func testLookup(t *testing.T, olap drivers.OLAPStore, database, databaseSchema string) {
	ctx := context.Background()

	bar, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, "BAR")
	require.NoError(t, err)
	require.Equal(t, "BAR", bar.Name)
	require.Equal(t, 2, len(bar.Schema.Fields))
	fieldNames := make(map[string]bool)
	for _, f := range bar.Schema.Fields {
		fieldNames[f.Name] = true
	}
	require.True(t, fieldNames["BAR"], "expected column BAR")
	require.True(t, fieldNames["BAZ"], "expected column BAZ")
	require.False(t, bar.View)

	model, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, "MODEL")
	require.NoError(t, err)
	require.True(t, model.View)
}

func testListDatabaseSchemas(t *testing.T, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	schemas, _, err := infoSchema.ListDatabaseSchemas(context.Background(), 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, schemas)

	found := false
	for _, s := range schemas {
		if s.Database == database && s.DatabaseSchema == databaseSchema {
			found = true
			break
		}
	}
	require.True(t, found, "expected schema %s.%s in ListDatabaseSchemas", database, databaseSchema)
}

func testListTables(t *testing.T, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	all, _, err := infoSchema.ListTables(context.Background(), database, databaseSchema, 0, "")
	require.NoError(t, err)
	tables := filterTableInfos(all)
	require.Equal(t, numKnown, len(tables))

	require.Equal(t, "ALL_DATATYPES", tables[0].Name)
	require.Equal(t, "BAR", tables[1].Name)
	require.Equal(t, "BAZ", tables[2].Name)
	require.Equal(t, "FOO", tables[3].Name)
	require.Equal(t, "FOZ", tables[4].Name)
	require.Equal(t, "MODEL", tables[5].Name)
	require.True(t, tables[5].View)

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}
}

func testGetTable(t *testing.T, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	ctx := context.Background()

	bar, err := infoSchema.GetTable(ctx, database, databaseSchema, "BAR")
	require.NoError(t, err)
	require.Equal(t, 2, len(bar.Schema))
	require.Contains(t, bar.Schema, "BAR")
	require.Contains(t, bar.Schema, "BAZ")
	require.False(t, bar.View)

	noTable, err := infoSchema.GetTable(ctx, database, databaseSchema, "nonexistent_table")
	require.NoError(t, err)
	require.Equal(t, 0, len(noTable.Schema))

	model, err := infoSchema.GetTable(ctx, database, databaseSchema, "MODEL")
	require.NoError(t, err)
	require.True(t, model.View)
}

func testListTablesPagination(t *testing.T, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	ctx := context.Background()
	pageSize := 2

	var collectedAll []*drivers.TableInfo
	var nextToken string
	for {
		tables, token, err := infoSchema.ListTables(ctx, database, databaseSchema, uint32(pageSize), nextToken)
		require.NoError(t, err)
		require.LessOrEqual(t, len(tables), pageSize)
		collectedAll = append(collectedAll, tables...)
		nextToken = token
		if token == "" {
			break
		}
	}
	require.Equal(t, numKnown, len(filterTableInfos(collectedAll)))

	// All at once
	tables, nextToken, err := infoSchema.ListTables(ctx, database, databaseSchema, 0, "")
	require.NoError(t, err)
	require.Equal(t, numKnown, len(filterTableInfos(tables)))
	require.Empty(t, nextToken)
}

func testLoadDDL(t *testing.T, olap drivers.OLAPStore, database, databaseSchema string) {
	ctx := context.Background()

	table, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, "BAR")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(ctx, table)
	require.NoError(t, err)
	require.NotEmpty(t, table.DDL)

	view, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, "MODEL")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(ctx, view)
	require.NoError(t, err)
	require.NotEmpty(t, view.DDL)
}
