package runtime

import (
	"context"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

type Health struct {
	HangingConn     error
	Registry        error
	InstancesHealth map[string]*InstanceHealth
}

type InstanceHealth struct {
	Controller error
	Repo       error
	OLAP       error
}

func (h *InstanceHealth) To() *runtimev1.InstanceHealth {
	if h == nil {
		return nil
	}
	r := &runtimev1.InstanceHealth{}
	if h.Controller != nil {
		r.ControllerError = h.Controller.Error()
	}
	if h.Repo != nil {
		r.RepoError = h.Repo.Error()
	}
	if h.OLAP != nil {
		r.OlapError = h.OLAP.Error()
	}
	return r
}

func (r *Runtime) Health(ctx context.Context) (*Health, error) {
	instances, err := r.registryCache.list()
	if err != nil {
		return nil, err
	}

	ih := make(map[string]*InstanceHealth, len(instances))
	for _, inst := range instances {
		ih[inst.ID], err = r.InstanceHealth(ctx, inst.ID)
		if err != nil && !errors.Is(err, drivers.ErrNotFound) {
			return nil, err
		}
	}
	return &Health{
		HangingConn:     r.connCache.HangingErr(),
		Registry:        r.registryCache.store.(drivers.Handle).Ping(ctx),
		InstancesHealth: ih,
	}, nil
}

func (r *Runtime) InstanceHealth(ctx context.Context, instanceID string) (*InstanceHealth, error) {
	return r.registryCache.instanceHealth(ctx, instanceID)
}
