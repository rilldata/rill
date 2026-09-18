package pgwire

import (
	"context"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rilldata/rill/runtime"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestInterpolateParameters(t *testing.T) {
	types := pgtype.NewMap()
	binaryInt, err := types.Encode(pgtype.Int4OID, pgtype.BinaryFormatCode, int32(42), nil)
	require.NoError(t, err)
	parsed, err := parseSQL(`SELECT value FROM metrics WHERE id = $1 AND label = $2 AND literal = '$1'`)
	require.NoError(t, err)
	query, err := parsed.interpolateMetricsParameters(
		[]base.Parameter{
			{OID: pgtype.Int4OID, Format: pgtype.BinaryFormatCode, Value: binaryInt},
			{OID: pgtype.TextOID, Format: pgtype.TextFormatCode, Value: []byte("O'Reilly")},
		},
		types,
		false,
	)
	require.NoError(t, err)
	require.Equal(t, `SELECT value FROM metrics WHERE id = 42 AND label = 'O''Reilly' AND literal = '$1'`, query)
}

func TestEncodeTimestamptzTextUsesNumericOffset(t *testing.T) {
	value, err := normalizeValue("2022-01-01T00:00:00Z", pgtype.TimestamptzOID)
	require.NoError(t, err)
	encoded, err := encodeValue(pgtype.NewMap(), pgtype.TimestamptzOID, pgtype.TextFormatCode, value)
	require.NoError(t, err)
	require.Equal(t, "2022-01-01 00:00:00+00:00", string(encoded))
}

func TestNormalizeNumericForBinaryEncoding(t *testing.T) {
	value, err := normalizeValue("123.45", pgtype.NumericOID)
	require.NoError(t, err)
	_, err = pgtype.NewMap().Encode(pgtype.NumericOID, pgtype.BinaryFormatCode, value, nil)
	require.NoError(t, err)
}

func TestInferParameterOIDs(t *testing.T) {
	parsed, err := parseSQL("SELECT $1::int4, $2::timestamp with time zone, $3")
	require.NoError(t, err)
	require.Equal(t, []uint32{pgtype.Int4OID, pgtype.TimestamptzOID, pgtype.TextOID}, parsed.parameterOIDs(nil))
}

func TestRuntimePGWireSupersetQueriesEndToEnd(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	server, err := base.NewServer(base.Options{
		NewSession: func(context.Context, map[string]string, string) (base.Session, error) {
			return NewSession(rt, instanceID, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
		},
	})
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	serveCtx, stopServer := context.WithCancel(t.Context())
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(serveCtx, listener)
	}()
	t.Cleanup(func() {
		stopServer()
		require.NoError(t, <-serveDone)
	})

	config, err := pgx.ParseConfig("postgresql://rill@" + listener.Addr().String() + "/default?sslmode=disable")
	require.NoError(t, err)
	conn, err := pgx.ConnectConfig(t.Context(), config)
	require.NoError(t, err)
	defer conn.Close(t.Context())

	// Parameter values must remain Metrics SQL values, even when they contain
	// catalog names, quotes, backslashes or an empty string.
	for _, value := range []string{"msn.com", "", "pg_catalog", "O'Reilly", `a\' OR 1=1 --`} {
		var count int64
		err := conn.QueryRow(t.Context(), "SELECT total_records FROM ad_bids_metrics_view WHERE domain = $1", value).Scan(&count)
		require.NoError(t, err, value)
		if value == "msn.com" {
			require.Positive(t, count)
		} else {
			require.Zero(t, count)
		}
	}

	// A quoted placeholder is a literal and must not require an argument.
	var count int64
	err = conn.QueryRow(t.Context(), "SELECT total_records FROM ad_bids_metrics_view WHERE domain = '$1'").Scan(&count)
	require.NoError(t, err)
	require.Zero(t, count)

	for _, value := range []string{"", `\`, "O'Reilly"} {
		var result string
		err := conn.QueryRow(t.Context(), "SELECT $1", value).Scan(&result)
		require.NoError(t, err)
		require.Equal(t, value, result)
	}

	// Pagination placeholders are integers even when Parse does not supply OIDs.
	for _, query := range []string{
		"SELECT domain FROM ad_bids_metrics_view ORDER BY domain LIMIT $1 OFFSET $2",
		"SELECT domain FROM ad_bids_metrics_view ORDER BY domain LIMIT $2, $1",
	} {
		rows, err := conn.Query(t.Context(), query, int64(2), int64(1))
		require.NoError(t, err)
		var domains []string
		for rows.Next() {
			var domain string
			require.NoError(t, rows.Scan(&domain))
			domains = append(domains, domain)
		}
		require.NoError(t, rows.Err())
		rows.Close()
		require.Len(t, domains, 2)
		var expected []string
		rows, err = conn.Query(t.Context(), "SELECT domain FROM ad_bids_metrics_view ORDER BY domain LIMIT 2 OFFSET 1")
		require.NoError(t, err)
		for rows.Next() {
			var domain string
			require.NoError(t, rows.Scan(&domain))
			expected = append(expected, domain)
		}
		require.NoError(t, rows.Err())
		rows.Close()
		require.Equal(t, expected, domains)
	}

	// SQLAlchemy resolves the table OID and columns in separate queries. This
	// exercises separate catalog rebuilds through the actual extended protocol.
	var tableOID int64
	err = conn.QueryRow(t.Context(), `SELECT c.oid FROM pg_catalog.pg_class c LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind IN ('r', 'v', 'm', 'f', 'p')`, "public", "ad_bids_metrics_view").Scan(&tableOID)
	require.NoError(t, err)
	rows, err := conn.Query(t.Context(), `SELECT a.attname, pg_catalog.format_type(a.atttypid, a.atttypmod) FROM pg_catalog.pg_attribute a WHERE a.attrelid = $1 AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum`, strconv.FormatInt(tableOID, 10))
	require.NoError(t, err)
	var columns [][2]string
	for rows.Next() {
		var name, typ string
		require.NoError(t, rows.Scan(&name, &typ))
		columns = append(columns, [2]string{name, typ})
	}
	require.NoError(t, rows.Err())
	rows.Close()
	require.Equal(t, [][2]string{
		{"timestamp", "timestamp with time zone"},
		{"publisher", "character varying"},
		{"domain", "character varying"},
		{"total_records", "bigint"},
		{"bid_price_sum", "double precision"},
	}, columns)

	// This is the time-grain query generated by Superset's chart API.
	rows, err = conn.Query(t.Context(), `SELECT DATE_TRUNC('day', timestamp) AS timestamp, total_records AS total_records FROM public.ad_bids_metrics_view GROUP BY DATE_TRUNC('day', timestamp) ORDER BY timestamp ASC LIMIT 5`)
	require.NoError(t, err)
	var timestamps []time.Time
	for rows.Next() {
		var timestamp time.Time
		var totalRecords int64
		require.NoError(t, rows.Scan(&timestamp, &totalRecords))
		timestamps = append(timestamps, timestamp)
		require.Positive(t, totalRecords)
	}
	require.NoError(t, rows.Err())
	rows.Close()
	require.Len(t, timestamps, 5)
	require.Equal(t, time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC), timestamps[0].UTC())
}

func TestRuntimeSessionMetricsSQLAndCatalog(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	session, err := NewSession(rt, instanceID, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()

	rows, err := session.Query(context.Background(), "SELECT pub, dom FROM ad_bids_metrics WHERE dom = 'msn.com' LIMIT 1", nil, nil)
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.Len(t, rows.Values(), 2)
	require.NoError(t, rows.Close())

	// Representative Metabase information_schema probe.
	rows, err = session.Query(context.Background(), "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'ad_bids_metrics'", nil, nil)
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.Equal(t, "ad_bids_metrics", string(rows.Values()[0]))
	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())

	// Representative Superset/SQLAlchemy pg_catalog probes. SQLAlchemy resolves
	// a table OID first and looks up its columns in a separate query, so catalog
	// OIDs and column order must remain stable across catalog database rebuilds.
	rows, err = session.Query(context.Background(), `SELECT c.oid FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relname = 'ad_bids_metrics_view'`, nil, nil)
	require.NoError(t, err)
	require.True(t, rows.Next())
	tableOID := string(rows.Values()[0])
	require.NoError(t, rows.Close())

	rows, err = session.Query(context.Background(), `SELECT a.attname, pg_catalog.format_type(a.atttypid, a.atttypmod) FROM pg_catalog.pg_attribute a WHERE a.attrelid = `+tableOID+` AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum`, nil, nil)
	require.NoError(t, err)
	var columns [][2]string
	for rows.Next() {
		columns = append(columns, [2]string{string(rows.Values()[0]), string(rows.Values()[1])})
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	require.Equal(t, [][2]string{
		{"timestamp", "timestamp with time zone"},
		{"publisher", "character varying"},
		{"domain", "character varying"},
		{"total_records", "bigint"},
		{"bid_price_sum", "double precision"},
	}, columns)

	_, err = session.Query(context.Background(), "SELECT * FROM missing_metrics_view", nil, nil)
	require.Error(t, err)
}

func TestCatalogRejectsExternalAccessAndMultipleStatements(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	session, err := NewSession(rt, instanceID, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()
	path := filepath.Join(t.TempDir(), "private.txt")
	require.NoError(t, os.WriteFile(path, []byte("private fixture"), 0o600))
	for _, query := range []string{
		"SELECT content FROM read_text(" + quoteLiteral(path) + ") /* pg_catalog */",
		"SELECT content FROM read_text(" + quoteLiteral(path) + "), pg_catalog.pg_type LIMIT 1",
		"SELECT * FROM pg_catalog.pg_type; SET enable_external_access=true",
		"COPY (SELECT * FROM pg_catalog.pg_type) TO " + quoteLiteral(filepath.Join(t.TempDir(), "out.csv")),
	} {
		rows, err := session.Query(t.Context(), query, nil, nil)
		if rows != nil {
			defer rows.Close()
		}
		require.Error(t, err, query)
	}
}

func TestDescribeMatchesMetricsResults(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	session, err := NewSession(rt, instanceID, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()
	for _, query := range []string{
		"SELECT publisher, total_records FROM ad_bids_metrics_view LIMIT 1",
		"SELECT DATE_TRUNC('day', timestamp) AS day, total_records FROM ad_bids_metrics_view LIMIT 1",
		"SELECT total_records, bid_price_sum FROM ad_bids_metrics_view",
	} {
		description, err := session.Describe(t.Context(), query, nil)
		require.NoError(t, err, query)
		rows, err := session.Query(t.Context(), query, nil, nil)
		require.NoError(t, err, query)
		require.Equal(t, rows.Fields(), description.Fields, query)
		require.NoError(t, rows.Close())
	}
}

func TestDescribeDoesNotEvaluateRows(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: map[string]string{
		"rill.yaml": "",
		"model.sql": "SELECT 'not-a-number' AS value, DATE '2025-01-01' AS day, true AS active",
		"metrics.yaml": `type: metrics_view
model: model
dimensions:
  - column: day
  - column: active
measures:
  - name: invalid_sum
    expression: sum(CAST(value AS BIGINT))
`,
	}})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)
	session, err := NewSession(rt, instanceID, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()
	description, err := session.Describe(t.Context(), "SELECT invalid_sum FROM metrics", nil)
	require.NoError(t, err)
	require.Len(t, description.Fields, 1)

	rows, err := session.Query(t.Context(), "SELECT invalid_sum FROM metrics", nil, nil)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
		}
		err = rows.Err()
	}
	require.Error(t, err, "execution must evaluate the invalid cast, but description must not")

	rows, err = session.Query(t.Context(), "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'metrics' ORDER BY ordinal_position", nil, nil)
	require.NoError(t, err)
	defer rows.Close()
	var columns [][2]string
	for rows.Next() {
		columns = append(columns, [2]string{string(rows.Values()[0]), string(rows.Values()[1])})
	}
	require.NoError(t, rows.Err())
	require.Equal(t, [][2]string{{"day", "DATE"}, {"active", "BOOLEAN"}, {"invalid_sum", "DECIMAL(38,9)"}}, columns)
}

func TestCatalogParameterTypesMatchDescription(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	session, err := NewSession(rt, instanceID, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()
	for _, tc := range []struct {
		name  string
		oid   uint32
		value any
	}{
		{"smallint", pgtype.Int2OID, int16(42)},
		{"integer", pgtype.Int4OID, int32(42)},
		{"bigint", pgtype.Int8OID, int64(42)},
		{"real", pgtype.Float4OID, float32(1.25)},
		{"double", pgtype.Float8OID, float64(1.25)},
		{"numeric", pgtype.NumericOID, pgtype.Numeric{Int: big.NewInt(123456), Exp: -5, Valid: true}},
		{"boolean", pgtype.BoolOID, true},
		{"text", pgtype.TextOID, "O'Reilly \\ pg_catalog $1"},
		{"empty_text", pgtype.TextOID, ""},
		{"empty_bytes", pgtype.ByteaOID, []byte{}},
		{"null", pgtype.Int8OID, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query := "SELECT $2 AS second, $1 AS first, $2 AS repeated; -- trailing comment"
			description, err := session.Describe(t.Context(), query, []uint32{pgtype.TextOID, tc.oid})
			require.NoError(t, err)
			for _, format := range []int16{pgtype.TextFormatCode, pgtype.BinaryFormatCode} {
				encoded, err := session.types.Encode(tc.oid, format, tc.value, []byte{})
				require.NoError(t, err)
				rows, err := session.Query(t.Context(), query, []base.Parameter{
					{OID: pgtype.TextOID, Value: []byte("first")},
					{OID: tc.oid, Format: format, Value: encoded},
				}, []int16{pgtype.BinaryFormatCode})
				require.NoError(t, err)
				for i, field := range rows.Fields() {
					require.Equal(t, description.Fields[i].DataTypeOID, field.DataTypeOID)
					require.Equal(t, description.Fields[i].DataTypeSize, field.DataTypeSize)
				}
				require.Equal(t, tc.oid, rows.Fields()[0].DataTypeOID)
				next := rows.Next()
				require.NoError(t, rows.Err())
				require.True(t, next)
				expected, err := session.types.Encode(tc.oid, pgtype.BinaryFormatCode, tc.value, []byte{})
				require.NoError(t, err)
				require.Equal(t, expected, rows.Values()[0])
				require.Equal(t, []byte("first"), rows.Values()[1])
				require.Equal(t, expected, rows.Values()[2])
				require.False(t, rows.Next())
				require.NoError(t, rows.Err())
				require.NoError(t, rows.Close())
			}
		})
	}
}

func TestMetricsColumnNamedLikeCatalog(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: map[string]string{
		"rill.yaml":    "",
		"model.sql":    "SELECT 'real-row' AS pg_type",
		"metrics.yaml": "type: metrics_view\nmodel: model\ndimensions:\n  - column: pg_type\nmeasures:\n  - name: total\n    expression: count(*)\n",
	}})
	session, err := NewSession(rt, id, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()
	query := "SELECT pg_type FROM metrics"
	description, err := session.Describe(t.Context(), query, nil)
	require.NoError(t, err)
	rows, err := session.Query(t.Context(), query, nil, nil)
	require.NoError(t, err)
	defer rows.Close()
	require.Equal(t, description.Fields, rows.Fields())
	require.True(t, rows.Next())
	require.Equal(t, "real-row", string(rows.Values()[0]))
	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func TestRuntimeSessionSecurityPolicies(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: map[string]string{
		"rill.yaml": "",
		"model.sql": "SELECT 'a' AS tenant, 10 AS amount UNION ALL SELECT 'b', 20",
		"metrics.yaml": `type: metrics_view
model: model
dimensions:
  - column: tenant
measures:
  - name: total
    expression: count(*)
  - name: secret_total
    expression: sum(amount)
security:
  access: "'{{ .user.role }}' = 'reader'"
  row_filter: "tenant = '{{ .user.tenant }}'"
  exclude:
    - names: [secret_total]
      if: "true"
`,
	}})
	for _, role := range []string{"reader", "blocked"} {
		t.Run(role, func(t *testing.T) {
			session, err := NewSession(rt, id, &runtime.SecurityClaims{
				Permissions:    []runtime.Permission{runtime.ReadMetrics},
				UserAttributes: map[string]any{"role": role, "tenant": "a"},
			})
			require.NoError(t, err)
			defer session.Close()

			rows, err := session.Query(t.Context(), "SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'metrics' ORDER BY ordinal_position", nil, nil)
			require.NoError(t, err)
			var columns []string
			for rows.Next() {
				columns = append(columns, string(rows.Values()[0]))
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			if role == "blocked" {
				require.Empty(t, columns)
			} else {
				require.Equal(t, []string{"tenant", "total"}, columns)
			}

			for _, query := range []string{"SELECT tenant, total FROM metrics", "SELECT secret_total FROM metrics"} {
				description, describeErr := session.Describe(t.Context(), query, nil)
				rows, queryErr := session.Query(t.Context(), query, nil, nil)
				if rows != nil {
					defer rows.Close()
				}
				if role == "blocked" || query == "SELECT secret_total FROM metrics" {
					require.Error(t, describeErr)
					require.Error(t, queryErr)
					continue
				}
				require.NoError(t, describeErr)
				require.NoError(t, queryErr)
				require.Equal(t, description.Fields, rows.Fields())
				require.True(t, rows.Next())
				require.Equal(t, [][]byte{[]byte("a"), []byte("1")}, rows.Values())
				require.False(t, rows.Next())
				require.NoError(t, rows.Err())
			}
		})
	}
}

func TestCatalogDatabaseReuseAndInvalidation(t *testing.T) {
	files := map[string]string{
		"rill.yaml":    "",
		"model.sql":    "SELECT 'a' AS x, 1 AS y",
		"metrics.yaml": "type: metrics_view\nmodel: model\ndimensions:\n  - column: x\nmeasures:\n  - name: total\n    expression: count(*)\n",
	}
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: files})
	session, err := NewSession(rt, id, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()
	columnsQuery := "SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'metrics' ORDER BY ordinal_position"
	columns := func(rows base.Rows) []string {
		var names []string
		for rows.Next() {
			names = append(names, string(rows.Values()[0]))
		}
		require.NoError(t, rows.Err())
		require.NoError(t, rows.Close())
		return names
	}

	// An open cursor must not block or corrupt a second catalog query on the same session.
	first, err := session.Query(t.Context(), columnsQuery, nil, nil)
	require.NoError(t, err)
	catalog := session.catalog
	require.NotNil(t, catalog)
	second, err := session.Query(t.Context(), "SELECT current_schema(), (SELECT count(*) FROM pg_matviews)", nil, nil)
	require.NoError(t, err)
	require.True(t, second.Next())
	require.Equal(t, [][]byte{[]byte("public"), []byte("0")}, second.Values())
	require.NoError(t, second.Close())
	require.Equal(t, []string{"x", "total"}, columns(first))
	require.Same(t, catalog, session.catalog, "unchanged metrics views must reuse the catalog database")

	files["metrics.yaml"] = "type: metrics_view\nmodel: model\ndimensions:\n  - column: x\n  - column: y\nmeasures:\n  - name: total\n    expression: count(*)\n"
	testruntime.PutFiles(t, rt, id, files)
	testruntime.ReconcileParserAndWait(t, rt, id)
	rows, err := session.Query(t.Context(), columnsQuery, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"x", "y", "total"}, columns(rows))
	require.NotSame(t, catalog, session.catalog, "changed metrics views must rebuild the catalog database")
}

func TestUnmappedTypesEncodeAsText(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: map[string]string{
		"rill.yaml":    "",
		"model.sql":    "SELECT TIMESTAMP '2020-01-01' AS ts, INTERVAL 1 DAY AS span, [1, 2] AS arr, 3::UINTEGER AS u, {'a': 1} AS st",
		"metrics.yaml": "type: metrics_view\nmodel: model\ntimeseries: ts\ndimensions:\n  - column: span\n  - column: arr\n  - column: u\n  - column: st\nmeasures:\n  - name: total\n    expression: count(*)\n",
	}})
	session, err := NewSession(rt, id, &runtime.SecurityClaims{Permissions: runtime.AllPermissions, SkipChecks: true})
	require.NoError(t, err)
	defer session.Close()

	for _, tc := range []struct {
		query    string
		expected []string
	}{
		{
			// DuckDB catalog values, including types without a direct PostgreSQL equivalent.
			query:    "SELECT 3::UINTEGER, 4::UBIGINT, 65535::USMALLINT, INTERVAL 1 DAY, [1, 2]::INTEGER[], {'a': 1}, [[1]]::INTEGER[][], '5a1a0d5a-0000-4000-8000-000000000000'::UUID",
			expected: []string{"3", "4", "65535", "1 day 00:00:00", "{1,2}", `{"a":1}`, "[[1]]", "5a1a0d5a-0000-4000-8000-000000000000"},
		},
		{
			// Resolver values arrive JSON-encoded, e.g. intervals as microseconds.
			query:    "SELECT span, arr, u, st, total FROM metrics",
			expected: []string{"24:00:00", "{1,2}", "3", `{"a":1}`, "1"},
		},
	} {
		description, err := session.Describe(t.Context(), tc.query, nil)
		require.NoError(t, err, tc.query)
		rows, err := session.Query(t.Context(), tc.query, nil, nil)
		require.NoError(t, err, tc.query)
		require.Equal(t, description.Fields, rows.Fields(), tc.query)
		require.True(t, rows.Next(), tc.query)
		require.NoError(t, rows.Err(), tc.query)
		var values []string
		for _, value := range rows.Values() {
			values = append(values, string(value))
		}
		require.Equal(t, tc.expected, values, tc.query)
		require.False(t, rows.Next())
		require.NoError(t, rows.Err())
		require.NoError(t, rows.Close())
	}
}
