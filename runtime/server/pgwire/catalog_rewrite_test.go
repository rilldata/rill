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
		`SELECT '(SELECT json_build_object(1) FROM pg_catalog.pg_sequence) AS identity_options'`,
		`SELECT 1 /* (SELECT json_build_object(1) FROM pg_catalog.pg_sequence) AS identity_options */`,
		`SELECT (SELECT json_build_object('a', 1) FROM pg_catalog.pg_sequence) AS "identity_options"`,
		`SELECT (SELECT json_build_object('a', 1) FROM pg_catalog.pg_class) AS identity_options`,
	} {
		t.Run(query, func(t *testing.T) {
			rewritten, err := rewriteCatalogSQL(query)
			require.NoError(t, err)
			require.Equal(t, query, rewritten)
		})
	}
}

func TestCatalogIdentityOptionsRewrite(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
	}{
		{"sqlalchemy 1.4", `SELECT a.attname, (SELECT json_build_object('always', a.attidentity = 'a', 'start', s.seqstart) FROM pg_catalog.pg_sequence s JOIN pg_catalog.pg_class c ON s.seqrelid = c."oid" WHERE s.seqrelid = pg_catalog.pg_get_serial_sequence(a.attrelid::regclass::text, a.attname)::regclass::oid) AS identity_options FROM pg_catalog.pg_attribute a`},
		{"sqlalchemy 2.x", `SELECT a.attname, (SELECT json_build_object('always', a.attidentity = 'a', 'start', pg_catalog.pg_sequence.seqstart) AS json_build_object_1 FROM pg_catalog.pg_sequence WHERE a.attidentity != '' AND pg_catalog.pg_sequence.seqrelid = CAST(CAST(pg_catalog.pg_get_serial_sequence(CAST(CAST(a.attrelid AS REGCLASS) AS TEXT), a.attname) AS REGCLASS) AS OID)) AS identity_options FROM pg_catalog.pg_attribute a`},
		{"unqualified lowercase", `select a.attname, (select json_build_object('start', s.seqstart) from pg_sequence s where s.seqrelid = a.attrelid::regclass::oid) as identity_options from pg_catalog.pg_attribute a`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rewritten, err := rewriteCatalogSQL(tc.query)
			require.NoError(t, err)
			require.Contains(t, rewritten, "a.attname, NULL AS identity_options")
			require.NotContains(t, rewritten, "json_build_object")
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
