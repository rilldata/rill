package metricssql_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/stretchr/testify/require"
)

func TestParseFilter(t *testing.T) {
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
			"in expression",
			"dim NOT IN ('helllo', 'world')",
			&metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorNin,
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
			"subquery expression with having",
			"dim IN (SELECT dim FROM mv HAVING count > 10)",
			&metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorIn,
					Expressions: []*metricsview.Expression{
						{
							Name: "dim",
						},
						{
							Subquery: &metricsview.Subquery{
								Dimension: metricsview.Dimension{Name: "dim"},
								Measures:  []metricsview.Measure{{Name: "count"}},
								Having: &metricsview.Expression{
									Condition: &metricsview.Condition{
										Operator: metricsview.OperatorGt,
										Expressions: []*metricsview.Expression{
											{
												Name: "count",
											},
											{
												Value: 10,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"subquery expression without filters",
			"dim IN (SELECT dim FROM mv)",
			&metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorIn,
					Expressions: []*metricsview.Expression{
						{
							Name: "dim",
						},
						{
							Subquery: &metricsview.Subquery{
								Dimension: metricsview.Dimension{Name: "dim"},
							},
						},
					},
				},
			},
			false,
		},
		{
			"subquery expression with where and having",
			"dim IN (SELECT dim FROM mv WHERE country = 'US' HAVING count > 10 AND sales > 100)",
			&metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorIn,
					Expressions: []*metricsview.Expression{
						{
							Name: "dim",
						},
						{
							Subquery: &metricsview.Subquery{
								Dimension: metricsview.Dimension{Name: "dim"},
								Measures:  []metricsview.Measure{{Name: "count"}, {Name: "sales"}},
								Where: &metricsview.Expression{
									Condition: &metricsview.Condition{
										Operator: metricsview.OperatorEq,
										Expressions: []*metricsview.Expression{
											{
												Name: "country",
											},
											{
												Value: "US",
											},
										},
									},
								},
								Having: &metricsview.Expression{
									Condition: &metricsview.Condition{
										Operator: metricsview.OperatorAnd,
										Expressions: []*metricsview.Expression{
											{
												Condition: &metricsview.Condition{
													Operator: metricsview.OperatorGt,
													Expressions: []*metricsview.Expression{
														{
															Name: "count",
														},
														{
															Value: 10,
														},
													},
												},
											},
											{
												Condition: &metricsview.Condition{
													Operator: metricsview.OperatorGt,
													Expressions: []*metricsview.Expression{
														{
															Name: "sales",
														},
														{
															Value: 100,
														},
													},
												},
											},
										},
									},
								},
							},
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
			got, err := metricssql.ParseFilter(tt.sql)
			if tt.wantErr {
				require.Equal(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, got, tt.want)
			}
		})
	}
}
