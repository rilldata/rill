package mysql_test

import (
	"context"
	"sort"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

const (
	database       = ""
	databaseSchema = "mysql"
)

var knownTestTables = []string{"all_datatypes", "bar", "baz", "foo", "foz", "model"}

const numKnown = 6

func TestInformationSchema(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestMySQL(t)
	ctx := t.Context()
	infoSchema := olap.InformationSchema()
	t.Run("testListDatabaseSchemas", func(t *testing.T) {
		testListDatabaseSchemas(t, ctx, infoSchema)
	})
	t.Run("testListTables", func(t *testing.T) {
		testListTables(t, ctx, infoSchema)
	})
	t.Run("testListTablesPagination", func(t *testing.T) {
		testListTablesPagination(t, ctx, infoSchema)
	})
	t.Run("testAll", func(t *testing.T) {
		testAll(t, ctx, infoSchema)
	})
	t.Run("testLookup", func(t *testing.T) {
		testLookup(t, ctx, infoSchema)
	})
	t.Run("testLoadDDL", func(t *testing.T) {
		testLoadDDL(t, ctx, infoSchema)
	})
}

func testListDatabaseSchemas(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	schemas, _, err := infoSchema.ListDatabaseSchemas(ctx, 0, "")
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

func testListTables(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	all, _, err := infoSchema.ListTables(ctx, database, databaseSchema, "", 0, "")
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
		require.Equal(t, database, tbl.Database, "table %s: expected Database=%s", tbl.Name, database)
		require.Equal(t, databaseSchema, tbl.DatabaseSchema, "table %s: expected DatabaseSchema=%s", tbl.Name, databaseSchema)
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}

	// like filter: %ba% should match bar and baz only
	liked, _, err := infoSchema.ListTables(ctx, database, databaseSchema, "%ba%", 0, "")
	require.NoError(t, err)
	liked = filterTableInfos(liked)
	require.Equal(t, 2, len(liked))
	require.Equal(t, "bar", liked[0].Name)
	require.Equal(t, "baz", liked[1].Name)
}

func testLookup(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	bar, err := infoSchema.Lookup(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	require.Equal(t, "bar", bar.Name)
	require.Equal(t, database, bar.Database)
	require.Equal(t, databaseSchema, bar.DatabaseSchema)
	require.True(t, bar.IsDefaultDatabase)
	require.True(t, bar.IsDefaultDatabaseSchema)
	require.Equal(t, 2, len(bar.Schema.Fields))
	fieldNames := make(map[string]bool)
	for _, f := range bar.Schema.Fields {
		fieldNames[f.Name] = true
	}
	require.True(t, fieldNames["bar"])
	require.True(t, fieldNames["baz"])
	require.False(t, bar.View)

	model, err := infoSchema.Lookup(ctx, database, databaseSchema, "model")
	require.NoError(t, err)
	require.True(t, model.View)

	_, err = infoSchema.Lookup(ctx, database, databaseSchema, "nonexistent_table")
	require.Error(t, err)

	allDatatypes, err := infoSchema.Lookup(ctx, database, databaseSchema, "all_datatypes")
	require.NoError(t, err)
	require.False(t, allDatatypes.View)

	type fieldSpec struct {
		name     string
		code     runtimev1.Type_Code
		rawType  string
		nullable bool
	}
	expectedFields := []fieldSpec{
		{"tinyint_col", runtimev1.Type_CODE_INT8, "tinyint", true},
		{"smallint_col", runtimev1.Type_CODE_INT16, "smallint", true},
		{"mediumint_col", runtimev1.Type_CODE_INT32, "mediumint", true},
		{"int_col", runtimev1.Type_CODE_INT32, "int", true},
		{"bigint_col", runtimev1.Type_CODE_INT64, "bigint", true},
		{"float_col", runtimev1.Type_CODE_FLOAT64, "float", true},
		{"double_col", runtimev1.Type_CODE_FLOAT64, "double", true},
		{"decimal_col", runtimev1.Type_CODE_STRING, "decimal", true},
		{"char_col", runtimev1.Type_CODE_STRING, "char", true},
		{"varchar_col", runtimev1.Type_CODE_STRING, "varchar", true},
		{"tinytext_col", runtimev1.Type_CODE_STRING, "tinytext", true},
		{"text_col", runtimev1.Type_CODE_STRING, "text", true},
		{"mediumtext_col", runtimev1.Type_CODE_STRING, "mediumtext", true},
		{"longtext_col", runtimev1.Type_CODE_STRING, "longtext", true},
		{"binary_col", runtimev1.Type_CODE_STRING, "binary", true},
		{"varbinary_col", runtimev1.Type_CODE_STRING, "varbinary", true},
		{"tinyblob_col", runtimev1.Type_CODE_STRING, "tinyblob", true},
		{"blob_col", runtimev1.Type_CODE_STRING, "blob", true},
		{"mediumblob_col", runtimev1.Type_CODE_STRING, "mediumblob", true},
		{"longblob_col", runtimev1.Type_CODE_STRING, "longblob", true},
		{"enum_col", runtimev1.Type_CODE_STRING, "enum", true},
		{"set_col", runtimev1.Type_CODE_STRING, "set", true},
		{"date_col", runtimev1.Type_CODE_DATE, "date", true},
		{"datetime_col", runtimev1.Type_CODE_TIMESTAMP, "datetime", true},
		{"timestamp_col", runtimev1.Type_CODE_TIMESTAMP, "timestamp", true},
		{"time_col", runtimev1.Type_CODE_TIME, "time", true},
		{"year_col", runtimev1.Type_CODE_INT16, "year", true},
		{"boolean_col", runtimev1.Type_CODE_INT8, "tinyint", true}, // BOOLEAN is alias for TINYINT(1) in MySQL info_schema
		{"bit_col", runtimev1.Type_CODE_STRING, "bit", true},
		{"json_col", runtimev1.Type_CODE_JSON, "json", true},
	}
	actualFields := make([]fieldSpec, len(allDatatypes.Schema.Fields))
	for i, f := range allDatatypes.Schema.Fields {
		actualFields[i] = fieldSpec{f.Name, f.Type.Code, f.Type.RawType, f.Type.Nullable}
	}
	require.Equal(t, expectedFields, actualFields)
}

func testListTablesPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 2

	var collectedAll []*drivers.TableInfo
	var nextToken string
	for {
		tables, token, err := infoSchema.ListTables(ctx, database, databaseSchema, "", uint32(pageSize), nextToken)
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
	tables, nextToken, err := infoSchema.ListTables(ctx, database, databaseSchema, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, numKnown, len(filterTableInfos(tables)))
	require.Empty(t, nextToken)

	// Paginate with like=%ba% (matches bar, baz): collect all pages
	var liked []*drivers.TableInfo
	var likedToken string
	for {
		page, tok, err := infoSchema.ListTables(ctx, database, databaseSchema, "%ba%", 1, likedToken)
		require.NoError(t, err)
		liked = append(liked, page...)
		likedToken = tok
		if tok == "" {
			break
		}
	}
	liked = filterTableInfos(liked)
	require.Equal(t, 2, len(liked))
	require.Equal(t, "bar", liked[0].Name)
	require.Equal(t, "baz", liked[1].Name)
}

func testLoadDDL(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	table, err := infoSchema.Lookup(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, table)
	require.NoError(t, err)
	require.NotEmpty(t, table.DDL)

	view, err := infoSchema.Lookup(ctx, database, databaseSchema, "model")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, view)
	require.NoError(t, err)
	require.NotEmpty(t, view.DDL)
}

func testAll(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	all, _, err := infoSchema.All(ctx, "", 0, "")
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
		require.Equal(t, database, tbl.Database, "table %s: expected Database=%s", tbl.Name, database)
		require.Equal(t, databaseSchema, tbl.DatabaseSchema, "table %s: expected DatabaseSchema=%s", tbl.Name, databaseSchema)
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}

	// like filter: %ba% should match bar and baz only
	liked, _, err := infoSchema.All(ctx, "%ba%", 0, "")
	require.NoError(t, err)
	liked = filterTableInfos(liked)
	require.Equal(t, 2, len(liked))
	require.Equal(t, "bar", liked[0].Name)
	require.Equal(t, "baz", liked[1].Name)
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
