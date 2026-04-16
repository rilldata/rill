package starrocks

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/starrocks/teststarrocks"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	srDatabase       = "default_catalog"
	srDatabaseSchema = "test_db"
)

var srKnownTestTables = []string{"all_datatypes", "bar", "baz", "foo", "foz", "model"}

const srNumKnown = 6

func srFilterOLAP(tables []*drivers.OlapTable) []*drivers.OlapTable {
	known := make(map[string]bool, srNumKnown)
	for _, n := range srKnownTestTables {
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

func srFilterTableInfos(tables []*drivers.TableInfo) []*drivers.TableInfo {
	known := make(map[string]bool, srNumKnown)
	for _, n := range srKnownTestTables {
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

func TestInformationSchemaStarRocks(t *testing.T) {
	testmode.Expensive(t)

	dsn := teststarrocks.StartWithData(t)
	// Connect with test_db as the current database so DATABASE() = 'test_db' returns true
	dsn = strings.Replace(dsn, "/?", "/test_db?", 1)

	conn, err := driver{}.Open("", "default", map[string]any{
		"dsn": dsn,
	}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)
	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	ctx := t.Context()

	t.Run("testAll", func(t *testing.T) {
		testSRAll(t, olap)
	})
	t.Run("testAllLike", func(t *testing.T) {
		testSRAllLike(t, olap)
	})
	t.Run("testLookup", func(t *testing.T) {
		testSRLookup(t, olap)
	})
	t.Run("testAllPagination", func(t *testing.T) {
		testSRAllPagination(t, olap)
	})
	t.Run("testAllPaginationWithLike", func(t *testing.T) {
		testSRAllPaginationWithLike(t, olap)
	})
	t.Run("testListDatabaseSchemas", func(t *testing.T) {
		testSRListDatabaseSchemas(t, infoSchema)
	})
	t.Run("testListTables", func(t *testing.T) {
		testSRListTables(t, ctx, infoSchema)
	})
	t.Run("testGetTable", func(t *testing.T) {
		testSRGetTable(t, ctx, infoSchema)
	})
	t.Run("testListTablesPagination", func(t *testing.T) {
		testSRListTablesPagination(t, ctx, infoSchema)
	})
	t.Run("testLoadDDL", func(t *testing.T) {
		testSRLoadDDL(t, olap)
	})
}

func testSRAll(t *testing.T, olap drivers.OLAPStore) {
	all, _, err := olap.InformationSchema().All(context.Background(), "", 0, "")
	require.NoError(t, err)
	tables := srFilterOLAP(all)
	require.Equal(t, srNumKnown, len(tables))

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

func testSRAllLike(t *testing.T, olap drivers.OLAPStore) {
	all, _, err := olap.InformationSchema().All(context.Background(), "%odel", 0, "")
	require.NoError(t, err)
	tables := srFilterOLAP(all)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "model", tables[0].Name)

	all, _, err = olap.InformationSchema().All(context.Background(), "%ba%", 0, "")
	require.NoError(t, err)
	tables = srFilterOLAP(all)
	require.Equal(t, 2, len(tables))
	require.Equal(t, "bar", tables[0].Name)
	require.Equal(t, "baz", tables[1].Name)
}

func testSRLookup(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	bar, err := olap.InformationSchema().Lookup(ctx, srDatabase, srDatabaseSchema, "bar")
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

	model, err := olap.InformationSchema().Lookup(ctx, srDatabase, srDatabaseSchema, "model")
	require.NoError(t, err)
	require.True(t, model.View)
}

func testSRAllPagination(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	pageSize := 2

	var collectedAll []*drivers.OlapTable
	var nextToken string
	for {
		tables, token, err := olap.InformationSchema().All(ctx, "", uint32(pageSize), nextToken)
		require.NoError(t, err)
		require.LessOrEqual(t, len(tables), pageSize)
		collectedAll = append(collectedAll, tables...)
		nextToken = token
		if token == "" {
			break
		}
	}
	require.Equal(t, srNumKnown, len(srFilterOLAP(collectedAll)))

	// All at once
	tables, nextToken, err := olap.InformationSchema().All(ctx, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, srNumKnown, len(srFilterOLAP(tables)))
	require.Empty(t, nextToken)
}

func testSRAllPaginationWithLike(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	pageSize := 1

	var collectedAll []*drivers.OlapTable
	var nextToken string
	for {
		tables, token, err := olap.InformationSchema().All(ctx, "%ba%", uint32(pageSize), nextToken)
		require.NoError(t, err)
		collectedAll = append(collectedAll, tables...)
		nextToken = token
		if token == "" {
			break
		}
	}
	filtered := srFilterOLAP(collectedAll)
	require.Equal(t, 2, len(filtered))
	require.Equal(t, "bar", filtered[0].Name)
	require.Equal(t, "baz", filtered[1].Name)
}

func testSRListDatabaseSchemas(t *testing.T, infoSchema drivers.InformationSchema) {
	schemas, _, err := infoSchema.ListDatabaseSchemas(context.Background(), 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, schemas)

	found := false
	for _, s := range schemas {
		if s.Database == srDatabase && s.DatabaseSchema == srDatabaseSchema {
			found = true
			break
		}
	}
	require.True(t, found, "expected schema %s.%s in ListDatabaseSchemas", srDatabase, srDatabaseSchema)
}

func testSRListTables(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	all, _, err := infoSchema.ListTables(ctx, srDatabase, srDatabaseSchema, 0, "")
	require.NoError(t, err)
	tables := srFilterTableInfos(all)
	require.Equal(t, srNumKnown, len(tables))

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

func testSRGetTable(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	bar, err := infoSchema.GetTable(ctx, srDatabase, srDatabaseSchema, "bar")
	require.NoError(t, err)
	require.Equal(t, 2, len(bar.Schema))
	require.Contains(t, bar.Schema, "bar")
	require.Contains(t, bar.Schema, "baz")
	require.False(t, bar.View)

	noTable, err := infoSchema.GetTable(ctx, srDatabase, srDatabaseSchema, "nonexistent_table")
	require.NoError(t, err)
	require.Equal(t, 0, len(noTable.Schema))

	model, err := infoSchema.GetTable(ctx, srDatabase, srDatabaseSchema, "model")
	require.NoError(t, err)
	require.True(t, model.View)
}

func testSRListTablesPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 2

	var collectedAll []*drivers.TableInfo
	var nextToken string
	for {
		tables, token, err := infoSchema.ListTables(ctx, srDatabase, srDatabaseSchema, uint32(pageSize), nextToken)
		require.NoError(t, err)
		require.LessOrEqual(t, len(tables), pageSize)
		collectedAll = append(collectedAll, tables...)
		nextToken = token
		if token == "" {
			break
		}
	}
	require.Equal(t, srNumKnown, len(srFilterTableInfos(collectedAll)))

	// All at once
	tables, nextToken, err := infoSchema.ListTables(ctx, srDatabase, srDatabaseSchema, 0, "")
	require.NoError(t, err)
	require.Equal(t, srNumKnown, len(srFilterTableInfos(tables)))
	require.Empty(t, nextToken)
}

func testSRLoadDDL(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	table, err := olap.InformationSchema().Lookup(ctx, srDatabase, srDatabaseSchema, "bar")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(ctx, table)
	require.NoError(t, err)
	require.NotEmpty(t, table.DDL)

	view, err := olap.InformationSchema().Lookup(ctx, srDatabase, srDatabaseSchema, "model")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(ctx, view)
	require.NoError(t, err)
	require.NotEmpty(t, view.DDL)
}
