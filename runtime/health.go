package runtime

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

type Health struct {
	HungConn error
	Registry error
}

type InstanceHealth struct {
	Controller error
	Repo       error
	OLAP       error
}

func (r *Runtime) Health(ctx context.Context) Health {
	s := Health{}
	s.HungConn = r.connCache.HangingErr()
	s.Registry = r.registryCache.store.(drivers.Handle).Ping(ctx)
	return s
}

func (r *Runtime) InstanceHealth(ctx context.Context, instanceID string) (InstanceHealth, error) {
	return r.registryCache.health(ctx, instanceID)
}
