package bigquery_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestListDatabaseSchemas(t *testing.T) {
	testmode.Expensive(t)
	conn, _ := acquireTestBigQuery(t)
	is, ok := conn.AsInformationSchema()
	require.True(t, ok)

	all, next, err := is.ListDatabaseSchemas(t.Context(), 1000, "")
	require.NoError(t, err)
	require.Empty(t, next)
	require.Contains(t, all, &drivers.DatabaseSchemaInfo{Database: "rilldata", DatabaseSchema: "integration_test"})
	require.Contains(t, all, &drivers.DatabaseSchemaInfo{Database: "rilldata", DatabaseSchema: "integration_test_2"})

	// Paginating one at a time returns the same schemas.
	var paged []*drivers.DatabaseSchemaInfo
	for {
		res, tok, err := is.ListDatabaseSchemas(t.Context(), 1, next)
		require.NoError(t, err)
		require.Len(t, res, 1)
		paged = append(paged, res...)
		if tok == "" {
			break
		}
		next = tok
	}
	require.Equal(t, all, paged)
}

func TestListTables(t *testing.T) {
	testmode.Expensive(t)
	conn, _ := acquireTestBigQuery(t)
	is, ok := conn.AsInformationSchema()
	require.True(t, ok)

	all, next, err := is.ListTables(t.Context(), "rilldata", "integration_test", 1000, "")
	require.NoError(t, err)
	require.Empty(t, next)
	require.Contains(t, all, &drivers.TableInfo{Name: "all_datatypes", View: false})
	require.Contains(t, all, &drivers.TableInfo{Name: "ad_bids", View: false})
	require.Contains(t, all, &drivers.TableInfo{Name: "ad_bids_view", View: true})
	require.Contains(t, all, &drivers.TableInfo{Name: "ad_bids_mv", View: true})

	// Paginating one at a time returns the same tables.
	var paged []*drivers.TableInfo
	for {
		res, tok, err := is.ListTables(t.Context(), "rilldata", "integration_test", 1, next)
		require.NoError(t, err)
		require.LessOrEqual(t, len(res), 1)
		paged = append(paged, res...)
		if tok == "" {
			break
		}
		next = tok
	}
	require.Equal(t, all, paged)
}

func TestLookupView(t *testing.T) {
	testmode.Expensive(t)
	conn, _ := acquireTestBigQuery(t)
	is, ok := conn.AsInformationSchema()
	require.True(t, ok)

	for name, view := range map[string]bool{"ad_bids": false, "ad_bids_view": true, "ad_bids_mv": true} {
		tbl, err := is.Lookup(t.Context(), "rilldata", "integration_test", name)
		require.NoError(t, err)
		require.Equal(t, view, tbl.View, name)
	}
}
