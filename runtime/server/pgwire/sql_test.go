package pgwire

import (
	"testing"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"
	"github.com/stretchr/testify/require"
)

func TestParameterScanning(t *testing.T) {
	for _, query := range []string{
		`SELECT '$1', "column$2", ` + "`column$3`" + ` FROM metrics -- $4`,
		`SELECT '$1''$2' FROM metrics /* $3 /* $4 */ $5 */`,
		`SELECT $$ $1 $$, $tag$ $2 $tag$ FROM metrics`,
		`SELECT E'escaped\'$1' FROM metrics`,
		`SELECT '\' FROM metrics -- $1`,
	} {
		t.Run(query, func(t *testing.T) {
			parsed, err := parseSQL(query)
			require.NoError(t, err)
			require.Empty(t, parsed.parameterOIDs(nil))
			result, err := parsed.interpolateMetricsParameters(nil, pgtype.NewMap(), false)
			require.NoError(t, err)
			require.Equal(t, query, result)
		})
	}
	query := "SELECT '$2' FROM metrics WHERE publisher = $1 /* $3 */ -- $4"
	parsed, err := parseSQL(query)
	require.NoError(t, err)
	require.Equal(t, []uint32{pgtype.TextOID}, parsed.parameterOIDs(nil))
	result, err := parsed.interpolateMetricsParameters([]base.Parameter{{OID: pgtype.TextOID, Value: []byte("publisher")}}, pgtype.NewMap(), false)
	require.NoError(t, err)
	require.Equal(t, "SELECT '$2' FROM metrics WHERE publisher = 'publisher' /* $3 */ -- $4", result)

	for _, query := range []string{"SELECT $0", "SELECT $65536", "SELECT $99999999999999999999", "SELECT 'unterminated", "SELECT /* unterminated", "SELECT $$unterminated", "SELECT 1; SELECT 2"} {
		_, err := parseSQL(query)
		require.Error(t, err, query)
	}
}

func TestCatalogRoutingIgnoresValuesAndComments(t *testing.T) {
	for _, query := range []string{
		`SELECT publisher FROM metrics WHERE publisher = 'pg_catalog'`,
		`SELECT publisher FROM metrics /* pg_catalog */`,
		"SELECT publisher FROM metrics -- information_schema",
		`SELECT pg_type FROM metrics`,
		`SELECT "pg_class", pg_catalog.value FROM metrics pg_catalog`,
		`SELECT * FROM public.pg_type`,
		`SELECT extract(day FROM timestamp) AS pg_type FROM metrics`,
		`SELECT publisher FROM metrics ORDER BY publisher, pg_type`,
		`COPY (SELECT * FROM pg_catalog.pg_type) TO 'file'`,
	} {
		parsed, err := parseSQL(query)
		require.NoError(t, err)
		require.False(t, parsed.isCatalog(), query)
	}
	for _, query := range []string{
		`SELECT * FROM "pg_catalog"."pg_type"`,
		`SELECT 'FROM is a string'`,
		`SELECT * FROM pg_type`,
		`SELECT * FROM public.metrics, pg_catalog.pg_type`,
		`SELECT * FROM (SELECT * FROM information_schema.tables) t`,
		`WITH t AS (SELECT * FROM pg_class) SELECT * FROM t`,
	} {
		parsed, err := parseSQL(query)
		require.NoError(t, err)
		require.True(t, parsed.isCatalog(), query)
	}
}

func TestEncodeEmptyAndNullValues(t *testing.T) {
	for _, format := range []int16{pgtype.TextFormatCode, pgtype.BinaryFormatCode} {
		for _, oid := range []uint32{pgtype.TextOID, pgtype.ByteaOID} {
			var value any = ""
			if oid == pgtype.ByteaOID {
				value = []byte{}
			}
			field := pgproto3.FieldDescription{DataTypeOID: oid, Format: format}
			encoded, _, err := encodeRow(pgtype.NewMap(), []pgproto3.FieldDescription{field, field}, []any{value, nil}, nil, nil)
			require.NoError(t, err)
			require.NotNil(t, encoded[0])
			require.Nil(t, encoded[1])
		}
	}
}

func TestLeadingZeroParameterIndex(t *testing.T) {
	parsed, err := parseSQL("SELECT $01")
	require.NoError(t, err)
	require.Equal(t, []uint32{pgtype.TextOID}, parsed.parameterOIDs(nil))
}
