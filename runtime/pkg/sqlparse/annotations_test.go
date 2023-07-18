package sqlparse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAST_ExtractAnnotations(t *testing.T) {
	sqlVariations := []struct {
		title       string
		sql         string
		annotations map[string]string
	}{
		{
			"comments at the top",
			`
-- some random comment
-- @materialise_v1
-- @materialise_v2  :	true  
-- @materialise_v3  :	tr ue  
-- some other comment
select * from AdBids
`,
			map[string]string{
				"materialise_v1": "",
				"materialise_v2": "true",
				"materialise_v3": "tr ue",
			},
		},
		{
			"comments in the middle",
			`
select
-- @measure: avg
-- @measure.format: usd
a,
-- @dimension
b
from AdBids
`,
			map[string]string{
				"measure":        "avg",
				"measure.format": "usd",
				"dimension":      "",
			},
		},
	}

	for _, tt := range sqlVariations {
		t.Run(tt.title, func(t *testing.T) {
			require.EqualValues(t, tt.annotations, ExtractAnnotations(tt.sql))
		})
	}
}
