package reconcilers_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/testruntime"
)

func TestModelReconcileScenarios(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantRes  int
		wantErr  int
		wantWarn int
	}{
		{
			name: "multiple SQL tests",
			files: map[string]string{
				"rill.yaml": "",
				"models/m1.yaml": `
type: model
sql: SELECT * FROM range(5)
tests:
  - name: Test Row Count
    sql: SELECT count(*) = 5 as ok FROM range(5)
    description: Should have 5 rows
  - name: Validate 3 is present
    sql: SELECT count(*) = 1 as ok FROM range(5) WHERE column0 = 3
`,
			},
			wantRes:  2,
			wantErr:  0,
			wantWarn: 0,
		},
		{
			name: "failing test detected",
			files: map[string]string{
				"rill.yaml": "",
				"models/m2.yaml": `
type: model
sql: SELECT * FROM range(3)
tests:
  - name: Should Fail
    sql: SELECT count(*) = 5 as ok FROM range(3)
    description: This test should fail because row count is not 5
`,
			},
			wantRes:  2,
			wantErr:  1,
			wantWarn: 0,
		},
		{
			name: "no tests",
			files: map[string]string{
				"rill.yaml": "",
				"models/m3.yaml": `
type: model
sql: SELECT 1 as a
`,
			},
			wantRes:  2,
			wantErr:  0,
			wantWarn: 0,
		},
		{
			name: "multiple models with and without tests",
			files: map[string]string{
				"rill.yaml": "",
				"models/m4.yaml": `
type: model
sql: SELECT 1 as a
tests:
  - name: Always Pass
    sql: SELECT 1 as ok
`,
				"models/m5.yaml": `
type: model
sql: SELECT 2 as b
`,
			},
			wantRes:  3,
			wantErr:  0,
			wantWarn: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rt, id := testruntime.NewInstance(t)
			testruntime.PutFiles(t, rt, id, tc.files)
			testruntime.ReconcileParserAndWait(t, rt, id)
			testruntime.RequireReconcileState(t, rt, id, tc.wantRes, tc.wantErr, tc.wantWarn)
		})
	}
}
