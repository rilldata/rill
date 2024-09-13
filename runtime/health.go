package runtime

import (
	"context"
	"encoding/json"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

type DashboardHealthQuery func(ctx context.Context, instanceID, name string) (Query, error)

type Health struct {
	HangingConn     error
	Registry        error
	InstancesHealth map[string]*InstanceHealth
}

type InstanceHealth struct {
	// always recomputed
	Controller string
	Repo       string
	Catalog    string

	// below fields can be cached
	OLAP       string
	Dashboards map[string]string

	// below fields determine if cache should be used
	ControllerVersion int64
	DashboardVersion  map[string]int64
}

func (r *Runtime) Health(ctx context.Context, query DashboardHealthQuery) (*Health, error) {
	instances, err := r.registryCache.list()
	if err != nil {
		return nil, err
	}

	ih := make(map[string]*InstanceHealth, len(instances))
	for _, inst := range instances {
		ih[inst.ID], err = r.InstanceHealth(ctx, inst.ID, query)
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

func (r *Runtime) InstanceHealth(ctx context.Context, instanceID string, query DashboardHealthQuery) (*InstanceHealth, error) {
	res := &InstanceHealth{}
	// check repo error
	repo, rr, err := r.Repo(ctx, instanceID)
	if err != nil {
		res.Repo = err.Error()
	} else {
		err = repo.(drivers.Handle).Ping(ctx)
		if err != nil {
			res.Repo = err.Error()
		}
		rr()
	}

	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		res.Controller = err.Error()
		return res, nil
	}

	// check olap error
	resources, err := ctrl.List(ctx, ResourceKindMetricsView, "", false)
	if err != nil {
		return nil, err
	}

	instanceHealth, ok := r.cachedInstanceHealth(ctx, ctrl.InstanceID, ctrl.catalog.version, resources)
	if ok {
		res.OLAP = instanceHealth.OLAP
		res.Dashboards = instanceHealth.Dashboards
		return res, nil
	}

	for _, resource := range resources {
		res.DashboardVersion[resource.Meta.Name.Name] = resource.Meta.StateVersion
	}
	res.ControllerVersion = ctrl.catalog.version

	olap, release, err := r.OLAP(ctx, ctrl.InstanceID, "")
	if err != nil {
		res.OLAP = err.Error()
		return res, nil
	}
	defer release()

	err = olap.(drivers.Handle).Ping(ctx)
	if err != nil {
		res.OLAP = err.Error()
		return res, nil
	}

	if query == nil {
		return res, nil
	}
	res.Dashboards = make(map[string]string, len(resources))
	for _, mv := range resources {
		q, err := query(ctx, ctrl.InstanceID, mv.Meta.Name.Name)
		if err != nil {
			res.Dashboards[mv.Meta.Name.Name] = err.Error()
			continue
		}
		err = r.Query(ctx, ctrl.InstanceID, q, 1)
		if err != nil {
			res.Dashboards[mv.Meta.Name.Name] = err.Error()
		}
	}
	return res, nil
}

func (r *Runtime) cachedInstanceHealth(ctx context.Context, instanceID string, ctrlVersion int64, resources []*runtimev1.Resource) (*InstanceHealth, bool) {
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, false
	}
	defer release()

	cached, err := catalog.FindInstanceHealth(ctx, instanceID)
	if err != nil {
		return nil, false
	}

	c := &InstanceHealth{}
	err = json.Unmarshal(cached.Health, c)
	if err != nil {
		return nil, false
	}

	if c.ControllerVersion != ctrlVersion {
		return nil, false
	}

	for _, res := range resources {
		v, ok := c.DashboardVersion[res.Meta.Name.Name]
		if !ok || v != res.Meta.StateVersion {
			return nil, false
		}
	}
	// ignores deleted metric views
	return c, true
}

func (h *InstanceHealth) To() *runtimev1.InstanceHealth {
	if h == nil {
		return nil
	}
	r := &runtimev1.InstanceHealth{
		ControllerError: h.Controller,
		RepoError:       h.Repo,
		OlapError:       h.OLAP,
		DashboardErrors: h.Dashboards,
	}
	return r
}
