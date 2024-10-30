package metricssqlparser_test

import (
	reflect "reflect"
	"testing"

	"github.com/rilldata/rill/runtime/metricsview"
	metricssqlparser "github.com/rilldata/rill/runtime/pkg/metricssql"
	"github.com/stretchr/testify/require"
)

func TestParseSQLFilter(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    *metricsview.Expression
		wantErr bool
	}{
		{
			"equal expression",
			"dim = 'val'",
			&metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorEq,
					Expressions: []*metricsview.Expression{
						{
							Name: "dim",
						},
						{
							Value: "val",
						},
					},
				},
			},
			false,
		},
		{
			"in expression",
			"dim IN ('helllo', 'world')",
			&metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorIn,
					Expressions: []*metricsview.Expression{
						{
							Name: "dim",
						},
						{
							Value: []any{"helllo", "world"},
						},
					},
				},
			},
			false,
		},
		{
			"date_add expression",
			"time >= '2021-01-01' + INTERVAL 1 DAY",
			&metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorGte,
					Expressions: []*metricsview.Expression{
						{
							Name: "time",
						},
						{
							Value: "2021-01-02T00:00:00Z",
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metricssqlparser.ParseSQLFilter(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSQLFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSQLFilter() = %v, want %v", must(t, got), must(t, tt.want))
			}
		})
	}
}

func must(t *testing.T, e *metricsview.Expression) string {
	str, err := metricsview.ExpressionToString(e)
	require.NoError(t, err)
	return str
}
