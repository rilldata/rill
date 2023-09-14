package server

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestInlineMeasureValidation(t *testing.T) {
	require.NoError(t, validateInlineMeasures([]*runtimev1.InlineMeasure{{Expression: "COUNT(*)"}}))
	require.NoError(t, validateInlineMeasures([]*runtimev1.InlineMeasure{{Expression: "COUNT(my_dim)"}}))
	require.NoError(t, validateInlineMeasures([]*runtimev1.InlineMeasure{{Expression: "COUNT(DISTINCT my_dim)"}}))
	require.NoError(t, validateInlineMeasures([]*runtimev1.InlineMeasure{{Expression: "count(distinct my_dim)"}}))
	require.Error(t, validateInlineMeasures([]*runtimev1.InlineMeasure{{Expression: "SUM(my_dim)"}}))
}
