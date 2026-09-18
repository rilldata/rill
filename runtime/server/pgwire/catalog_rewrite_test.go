package pgwire

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCatalogRewritesPreserveQuotedText(t *testing.T) {
	for _, query := range []string{
		`SELECT 'version()', 'pg_backend_pid()', 'pg_get_indexdef(foo)', 'pg_catalog.pg_matviews'`,
		`SELECT $$pg_catalog.format_type(a.atttypid, a.atttypmod)$$, $tag$'pg_class'::regclass$tag$`,
		`SELECT 1 AS "version()", 2 AS "pg_catalog.pg_matviews"`,
		`SELECT 1 /* version() pg_get_indexdef(foo) pg_catalog.pg_matviews */ -- pg_backend_pid()`,
		`SELECT custom_version(), app.version(), app.pg_backend_pid(), app.pg_catalog.pg_matviews`,
		`SELECT my_pg_get_indexdef(foo), my_pg_catalog.pg_matviews`,
	} {
		t.Run(query, func(t *testing.T) {
			rewritten, err := rewriteCatalogSQL(query)
			require.NoError(t, err)
			require.Equal(t, query, rewritten)
		})
	}
}

func TestCatalogFunctionRewrites(t *testing.T) {
	rt, id := testruntime.NewInstanceForProject(t, "ad_bids")
	session, err := NewSession(rt, id, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()
	for _, tc := range []struct {
		query  string
		names  []string
		values []string
	}{
		{`SELECT 'version()' AS label`, []string{"label"}, []string{"version()"}},
		{`SELECT 'pg_backend_pid()' AS label`, []string{"label"}, []string{"pg_backend_pid()"}},
		{`SELECT version(), pg_backend_pid()`, []string{"version", "pg_backend_pid"}, []string{"PostgreSQL 16.3 (Rill pgwire)", "1234"}},
		{`SELECT version() AS server_version, pg_backend_pid() AS pid`, []string{"server_version", "pid"}, []string{"PostgreSQL 16.3 (Rill pgwire)", "1234"}},
		{`SELECT version() server_version, pg_backend_pid() "pid"`, []string{"server_version", "pid"}, []string{"PostgreSQL 16.3 (Rill pgwire)", "1234"}},
		{`SELECT PG_CATALOG.VERSION /* comment */ (), pg_catalog.pg_backend_pid ()`, []string{"version", "pg_backend_pid"}, []string{"PostgreSQL 16.3 (Rill pgwire)", "1234"}},
		{`SELECT coalesce(NULL, version(), 'fallback') AS v, 1 + pg_backend_pid() AS pid`, []string{"v", "pid"}, []string{"PostgreSQL 16.3 (Rill pgwire)", "1235"}},
		{`SELECT (SELECT version()) AS nested, version() || '!' AS combined`, []string{"nested", "combined"}, []string{"PostgreSQL 16.3 (Rill pgwire)", "PostgreSQL 16.3 (Rill pgwire)!"}},
		{`SELECT CASE WHEN pg_backend_pid() > 0 THEN version() END AS v`, []string{"v"}, []string{"PostgreSQL 16.3 (Rill pgwire)"}},
		{`SELECT 1 AS value ORDER BY pg_backend_pid(), version()`, []string{"value"}, []string{"1"}},
	} {
		t.Run(tc.query, func(t *testing.T) {
			description, err := session.Describe(t.Context(), tc.query, nil)
			require.NoError(t, err)
			rows, err := session.Query(t.Context(), tc.query, nil, nil)
			require.NoError(t, err)
			defer rows.Close()
			require.Equal(t, description.Fields, rows.Fields())
			var names []string
			for _, field := range rows.Fields() {
				names = append(names, string(field.Name))
			}
			require.Equal(t, tc.names, names)
			next := rows.Next()
			require.NoError(t, rows.Err())
			require.True(t, next)
			var values []string
			for _, value := range rows.Values() {
				values = append(values, string(value))
			}
			require.Equal(t, tc.values, values)
			require.False(t, rows.Next())
			require.NoError(t, rows.Err())
		})
	}
}
