package drivers

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type AIService interface {
	GenerateMetricsViewYAML(ctx context.Context, baseTable string, sqlDialect string, schema *runtimev1.StructType) (string, error)
}
