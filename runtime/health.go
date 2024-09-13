package runtime

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
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

	// cached
	OLAP       string
	Dashboards map[string]string

	Hash string
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
	olap, release, err := r.OLAP(ctx, ctrl.InstanceID, "")
	if err != nil {
		res.OLAP = err.Error()
		return res, nil
	}
	defer release()

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

	err = olap.(drivers.Handle).Ping(ctx)
	if err != nil {
		res.OLAP = err.Error()
	} else {
		// run queries against dashboards
		if query != nil {
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
		}
	}

	// save to cache
	// populate the versions
	hash, err := r.healthResultHash(ctrl.catalog.version, resources)
	if err != nil {
		return nil, err
	}
	res.Hash = hash

	bytes, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	err = catalog.UpsertInstanceHealth(ctx, &drivers.InstanceHealth{
		InstanceID: instanceID,
		Health:     bytes,
	})
	if err != nil {
		return nil, err
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

	hash, err := r.healthResultHash(ctrlVersion, resources)
	if err != nil || hash != c.Hash {
		return nil, false
	}
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

func (r *Runtime) healthResultHash(ctrlVersion int64, resources []*runtimev1.Resource) (string, error) {
	hash := md5.New()
	err := binary.Write(hash, binary.BigEndian, ctrlVersion)
	if err != nil {
		return "", err
	}

	for _, res := range resources {
		err := binary.Write(hash, binary.BigEndian, res.Meta.StateVersion)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
