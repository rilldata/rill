package resolvers

import (
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/resolvers/metricsresolver"
)

func init() {
	runtime.RegisterResolverInitializer("metrics", metricsresolver.New)
}
