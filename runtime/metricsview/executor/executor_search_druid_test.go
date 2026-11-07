package executor

import (
	"context"
	"errors"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

// errSQLCaptured is returned by sqlCaptureOLAP.Query so that a test can stop the executor right after the SQL is generated.
var errSQLCaptured = errors.New("sql captured")

// sqlCaptureOLAP is a stub OLAP store that reports the Druid dialect and records the statement passed to Query instead of executing it.
type sqlCaptureOLAP struct {
	drivers.OLAPStore
	stmt *drivers.Statement
}

func (o *sqlCaptureOLAP) Dialect() drivers.Dialect { return druid.DialectDruid }

func (o *sqlCaptureOLAP) Query(_ context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	o.stmt = stmt
	return nil, errSQLCaptured
}

// TestSearchDruidFallbackSQL checks the full UNION ALL query that Search issues to Druid when native search is unavailable (here because the query has no time range).
// The outer aliases must be quoted since "value" is a reserved keyword in Druid (Calcite) SQL.
// The per-dimension inner queries are covered separately by TestSearchDruidSQLCastsNonStringDimensions.
func TestSearchDruidFallbackSQL(t *testing.T) {
	olap := &sqlCaptureOLAP{}
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "events",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "publisher", Column: "publisher"},
				{Name: "domain", Column: "domain"},
			},
		},
		security: runtime.ResolvedSecurityOpen,
		olap:     olap,
	}

	_, err := e.Search(context.Background(), &metricsview.SearchQuery{
		MetricsView: "mv",
		Dimensions:  []string{"publisher", "domain"},
		Search:      "oo",
	}, nil)
	require.ErrorIs(t, err, errSQLCaptured)
	require.NotNil(t, olap.stmt)
	require.Equal(t, `SELECT 'publisher' AS "dimension", "publisher" AS "value" FROM (SELECT ("publisher") AS "publisher" FROM "events" WHERE (REGEXP_LIKE(("publisher"), ?)) GROUP BY 1) UNION ALL SELECT 'domain' AS "dimension", "domain" AS "value" FROM (SELECT ("domain") AS "domain" FROM "events" WHERE (REGEXP_LIKE(("domain"), ?)) GROUP BY 1)`, olap.stmt.Query)
	require.Equal(t, []any{"^(?i).*oo.*$", "^(?i).*oo.*$"}, olap.stmt.Args)
}
