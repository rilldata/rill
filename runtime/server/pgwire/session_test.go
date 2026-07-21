package pgwire

import (
	"context"
	"testing"

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
	query, err := interpolateParameters(
		`SELECT value FROM metrics WHERE id = $1 AND label = $2 AND literal = '$1'`,
		[]base.Parameter{
			{OID: pgtype.Int4OID, Format: pgtype.BinaryFormatCode, Value: binaryInt},
			{OID: pgtype.TextOID, Format: pgtype.TextFormatCode, Value: []byte("O'Reilly")},
		},
		types,
	)
	require.NoError(t, err)
	require.Equal(t, `SELECT value FROM metrics WHERE id = 42 AND label = 'O''Reilly'::text AND literal = '$1'`, query)
}

func TestNormalizeNumericForBinaryEncoding(t *testing.T) {
	value, err := normalizeValue("123.45", pgtype.NumericOID)
	require.NoError(t, err)
	_, err = pgtype.NewMap().Encode(pgtype.NumericOID, pgtype.BinaryFormatCode, value, nil)
	require.NoError(t, err)
}

func TestInferParameterOIDs(t *testing.T) {
	require.Equal(t,
		[]uint32{pgtype.Int4OID, pgtype.TimestamptzOID, pgtype.TextOID},
		inferParameterOIDs("SELECT $1::int4, $2::timestamp with time zone, $3", nil),
	)
}

func TestCatalogClassification(t *testing.T) {
	require.True(t, isCatalogQuery("SHOW timezone"))
	require.True(t, isCatalogQuery("SELECT version()"))
	require.True(t, isCatalogQuery("SELECT * FROM pg_catalog.pg_type"))
	require.False(t, isCatalogQuery("SELECT country FROM sales"))
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

	// Representative Superset/SQLAlchemy pg_catalog probe.
	rows, err = session.Query(context.Background(), `SELECT c.relname FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relname = 'ad_bids_metrics'`, nil, nil)
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.Equal(t, "ad_bids_metrics", string(rows.Values()[0]))
	require.NoError(t, rows.Close())

	_, err = session.Query(context.Background(), "SELECT * FROM missing_metrics_view", nil, nil)
	require.Error(t, err)
}
