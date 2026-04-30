package postgres_test

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
	database       = "postgres"
	databaseSchema = "public"
)

var knownTestTables = []string{"all_datatypes", "bar", "baz", "foo", "foz", "model"}

const numKnown = 6

func TestInformationSchema(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestPostgres(t)
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
	t.Run("testListTablesForAll", func(t *testing.T) {
		testListTablesForAll(t, ctx, infoSchema)
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
}

func testListTablesForAll(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	all, _, err := infoSchema.ListTables(ctx, "", "", "", 0, "")
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

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}
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
		{"id", runtimev1.Type_CODE_INT64, "integer", true},
		{"uuid", runtimev1.Type_CODE_UUID, "uuid", true},
		{"name", runtimev1.Type_CODE_STRING, "text", true},
		{"age", runtimev1.Type_CODE_INT64, "integer", true},
		{"is_married", runtimev1.Type_CODE_BOOL, "boolean", true},
		{"date_of_birth", runtimev1.Type_CODE_DATE, "date", true},
		{"time_of_day", runtimev1.Type_CODE_STRING, "time without time zone", true},
		{"created_at", runtimev1.Type_CODE_TIMESTAMP, "timestamp without time zone", true},
		{"personal_info", runtimev1.Type_CODE_JSON, "json", true},
		{"personal_info2", runtimev1.Type_CODE_JSON, "jsonb", true},
		{"is_alive", runtimev1.Type_CODE_STRING, "bit", true},
		{"binary_data", runtimev1.Type_CODE_STRING, "bit varying", true},
		{"gender", runtimev1.Type_CODE_STRING, "character", true},
		{"gender_full", runtimev1.Type_CODE_STRING, "character varying", true},
		{"nickname", runtimev1.Type_CODE_STRING, "character", true},
		{"num_of_dependents", runtimev1.Type_CODE_INT64, "smallint", true},
		{"biography", runtimev1.Type_CODE_STRING, "text", true},
		{"last_login", runtimev1.Type_CODE_TIMESTAMP, "timestamp with time zone", true},
		{"weight", runtimev1.Type_CODE_FLOAT64, "real", true},
		{"height", runtimev1.Type_CODE_FLOAT64, "double precision", true},
		{"sibling_rank", runtimev1.Type_CODE_INT64, "smallint", true},
		{"credit_score", runtimev1.Type_CODE_INT64, "integer", true},
		{"net_worth", runtimev1.Type_CODE_INT64, "bigint", true},
		{"salary_history", runtimev1.Type_CODE_ARRAY, "ARRAY", true},
		{"login_history", runtimev1.Type_CODE_ARRAY, "ARRAY", true},
		{"emp_salary", runtimev1.Type_CODE_DECIMAL, "numeric", true},
		{"country", runtimev1.Type_CODE_UNSPECIFIED, "USER-DEFINED", true},
	}
	actualFields := make([]fieldSpec, len(allDatatypes.Schema.Fields))
	for i, f := range allDatatypes.Schema.Fields {
		actualFields[i] = fieldSpec{f.Name, f.Type.Code, f.Type.RawType, f.Type.Nullable}
	}
	require.Equal(t, expectedFields, actualFields)
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

func filterOLAP(tables []*drivers.TableInfo) []*drivers.TableInfo {
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
