package runtime

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type Health struct {
	HungConn  bool
	Registry  error
	Instances map[string]*InstanceHealth
}

type InstanceHealth struct {
	Controller   error
	RepoSync     error
	AdminConnect error
}

func (h *InstanceHealth) To() *runtimev1.InstanceHealth {
	return &runtimev1.InstanceHealth{
		Controller:   h.Controller.Error(),
		RepoSync:     h.RepoSync.Error(),
		AdminConnect: h.AdminConnect.Error(),
	}
}

func (r *Runtime) Health(ctx context.Context) *Health {
	s := &Health{}

	s.HungConn = r.connCache.hasHungConn

	if err := r.registryCache.store.Ping(ctx); err != nil {
		s.Registry = err
	}

	s.Instances = r.registryCache.health(ctx)
	return s
}
