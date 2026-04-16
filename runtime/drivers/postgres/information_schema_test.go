package postgres_test

import (
	"context"
	"sort"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

var knownTestTables = []string{"all_datatypes", "bar", "baz", "foo", "foz", "model"}

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

func TestInformationSchemaPostgres(t *testing.T) {
	testmode.Expensive(t)

	conn, olap := acquireTestPostgres(t)
	ctx := t.Context()
	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	// Resolve current database and schema from the session.
	rows, err := olap.Query(ctx, &drivers.Statement{Query: "SELECT current_database(), current_schema()"})
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
		testListTables(t, ctx, infoSchema, database, databaseSchema)
	})
	t.Run("testGetTable", func(t *testing.T) {
		testGetTable(t, ctx, infoSchema, database, databaseSchema)
	})
	t.Run("testListTablesPagination", func(t *testing.T) {
		testListTablesPagination(t, ctx, infoSchema, database, databaseSchema)
	})
	t.Run("testLoadDDL", func(t *testing.T) {
		testLoadDDLIS(t, olap, database, databaseSchema)
	})
}

func testAll(t *testing.T, olap drivers.OLAPStore) {
	all, _, err := olap.InformationSchema().All(context.Background(), "", 0, "")
	require.NoError(t, err)
	tables := filterOLAP(all)
	require.Equal(t, numKnown, len(tables))

	require.Equal(t, "all_datatypes", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)
	require.True(t, tables[5].View)
	for _, tbl := range tables[:5] {
		require.False(t, tbl.View, "table %s should not be a view", tbl.Name)
	}
}

func testLookup(t *testing.T, olap drivers.OLAPStore, database, databaseSchema string) {
	ctx := context.Background()

	bar, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	require.Equal(t, "bar", bar.Name)
	require.Equal(t, 2, len(bar.Schema.Fields))
	fieldNames := make(map[string]bool)
	for _, f := range bar.Schema.Fields {
		fieldNames[f.Name] = true
	}
	require.True(t, fieldNames["bar"])
	require.True(t, fieldNames["baz"])
	require.False(t, bar.View)

	model, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, "model")
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

func testListTables(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	all, _, err := infoSchema.ListTables(ctx, database, databaseSchema, 0, "")
	require.NoError(t, err)
	tables := filterTableInfos(all)
	require.Equal(t, numKnown, len(tables))

	require.Equal(t, "all_datatypes", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)
	require.True(t, tables[5].View)

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}
}

func testGetTable(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	bar, err := infoSchema.GetTable(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	require.Equal(t, 2, len(bar.Schema))
	require.Contains(t, bar.Schema, "bar")
	require.Contains(t, bar.Schema, "baz")
	require.False(t, bar.View)

	noTable, err := infoSchema.GetTable(ctx, database, databaseSchema, "nonexistent_table")
	require.NoError(t, err)
	require.Equal(t, 0, len(noTable.Schema))

	model, err := infoSchema.GetTable(ctx, database, databaseSchema, "model")
	require.NoError(t, err)
	require.True(t, model.View)
}

func testListTablesPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema, database, databaseSchema string) {
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

func testLoadDDLIS(t *testing.T, olap drivers.OLAPStore, database, databaseSchema string) {
	ctx := context.Background()

	table, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(ctx, table)
	require.NoError(t, err)
	require.NotEmpty(t, table.DDL)

	view, err := olap.InformationSchema().Lookup(ctx, database, databaseSchema, "model")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(ctx, view)
	require.NoError(t, err)
	require.NotEmpty(t, view.DDL)
}
