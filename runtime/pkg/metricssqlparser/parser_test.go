package metricssqlparser

import (
	"context"
	"testing"

	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCompiler_Compile(t *testing.T) {
	compiler := New()
	runtime, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	ctrl, err := runtime.Controller(context.Background(), instanceID)
	require.NoError(t, err)

	passTests := map[string]string{
		"select pub, dom from ad_bids_metrics LIMIT 5": "SELECT publisher, domain FROM ad_bids LIMIT 5",
	}
	for inSQL, outSQL := range passTests {
		t.Run(t.Name(), func(t *testing.T) {
			got, _, _, err := compiler.Compile(ctrl, instanceID, inSQL, nil)
			require.NoError(t, err)
			if got != outSQL {
				t.Errorf("Compiler.Compile() input = %v, got = %v, want %v", inSQL, got, outSQL)
			}
		})
	}
}
