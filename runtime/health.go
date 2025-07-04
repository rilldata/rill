package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// Health reports a combined health status for the runtime and all instances.
type Health struct {
	HangingConn     error
	Registry        error
	InstancesHealth map[string]*InstanceHealth
}

// InstanceHealth contains health information for a single instance.
// The information about OLAP and metrics views is cached in the catalog.
// We want to avoid hitting the underlying OLAP engine when OLAP engine can scale to zero when no queries are generated within TTL.
// We do not want to keep it running just to check health. In such cases, we use the cached health information.
type InstanceHealth struct {
	// Controller error. The controller is considered healthy if this is empty.
	Controller string `json:"controller"`
	// ControllerVersion is the version of the controller that cached this health information.
	// It is used for health cache checks.
	ControllerVersion int64 `json:"controller_version"`
	// OLAP error. May be cached for OLAPs that scale to zero. The OLAP is considered healthy if this is empty.
	OLAP string `json:"olap"`
	// Repo error. The repo is considered healthy if this is empty.
	Repo string `json:"repo"`
	// MetricsViews contains health checks for metrics views.
	MetricsViews map[string]InstanceHealthMetricsViewError `json:"metrics_views"`
	// ParseErrCount is the number of parse errors in the project parser.
	ParseErrCount int `json:"parse_error_count"`
	// ReconcileErrCount is the number of resources with reconcile errors.
	ReconcileErrCount int `json:"reconcile_error_count"`
}

// InstanceHealthMetricsViewError contains health information for a single metrics view.
type InstanceHealthMetricsViewError struct {
	Err     string `json:"err"`
	Version int64  `json:"version"`
}

// Proto converts InstanceHealth to the proto representation.
func (h *InstanceHealth) Proto() *runtimev1.InstanceHealth {
	if h == nil {
		return nil
	}
	r := &runtimev1.InstanceHealth{
		ControllerError:     h.Controller,
		RepoError:           h.Repo,
		OlapError:           h.OLAP,
		ParseErrorCount:     int32(h.ParseErrCount),
		ReconcileErrorCount: int32(h.ReconcileErrCount),
	}
	r.MetricsViewErrors = make(map[string]string, len(h.MetricsViews))
	for k, v := range h.MetricsViews {
		if v.Err != "" {
			r.MetricsViewErrors[k] = v.Err
		}
	}
	return r
}

func (r *Runtime) Health(ctx context.Context, fullStatus bool) (*Health, error) {
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
		// if there is a single instance hosted on this runtime then instead of returning error msgs throw error if OLAP/repo/controller are in error state
		if len(instances) == 1 && !fullStatus {
			h := ih[inst.ID]
			if h.OLAP != "" {
				return nil, errors.New(h.OLAP)
			}
			if h.Repo != "" {
				return nil, errors.New(h.Repo)
			}
			if h.Controller != "" {
				return nil, errors.New(h.Controller)
			}
		}
	}
	return &Health{
		HangingConn:     r.connCache.HangingErr(),
		Registry:        r.registryCache.store.(drivers.Handle).Ping(ctx),
		InstancesHealth: ih,
	}, nil
}

func (r *Runtime) InstanceHealth(ctx context.Context, instanceID string) (*InstanceHealth, error) {
	res := &InstanceHealth{}

	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		res.Controller = err.Error()
		return res, nil
	}

	parser, err := ctrl.Get(ctx, GlobalProjectParserName, false)
	if err != nil {
		return nil, err
	}
	res.ParseErrCount = len(parser.GetProjectParser().State.ParseErrors)

	cachedHealth, _ := r.cachedInstanceHealth(ctx, ctrl.InstanceID, ctrl.catalog.version)

	// check repo error
	err = r.checkRepo(ctx, inst)
	if err == nil {
		res.Repo = ""
	} else {
		res.Repo = err.Error()
	}

	// check OLAP error
	err = r.checkOLAP(ctx, inst, cachedHealth)
	if err == nil {
		res.OLAP = ""
	} else {
		res.OLAP = err.Error()
	}

	// check resources with reconcile errors
	resources, err := ctrl.List(ctx, "", "", false)
	if err != nil {
		return nil, err
	}
	for _, r := range resources {
		if r.Meta.ReconcileError != "" {
			res.ReconcileErrCount++
		}
	}

	// run queries against metrics views
	res.MetricsViews = make(map[string]InstanceHealthMetricsViewError, len(resources))
	for _, mv := range resources {
		if mv.GetMetricsView() == nil || mv.GetMetricsView().State.ValidSpec == nil {
			continue
		}
		if mv.GetMetricsView().State.ValidSpec.TimeDimension == "" {
			// no time dimension so metrics_time_range is guranateed to fail
			continue
		}
		olap, release, err := r.OLAP(ctx, instanceID, mv.GetMetricsView().State.ValidSpec.Connector)
		if err != nil {
			res.MetricsViews[mv.Meta.Name.Name] = InstanceHealthMetricsViewError{Err: err.Error()}
			continue
		}
		mayBeScaledToZero := olap.MayBeScaledToZero(ctx)
		release()

		// only use cached health if the OLAP can be scaled to zero
		if cachedHealth != nil && mayBeScaledToZero {
			mvHealth, ok := cachedHealth.MetricsViews[mv.Meta.Name.Name]
			if ok && mvHealth.Version == mv.Meta.StateVersion {
				res.MetricsViews[mv.Meta.Name.Name] = mvHealth
				continue
			}
		}
		resolverRes, err := r.Resolve(ctx, &ResolveOptions{
			InstanceID:         ctrl.InstanceID,
			Resolver:           "metrics_time_range",
			ResolverProperties: map[string]any{"metrics_view": mv.Meta.Name.Name},
			Args:               map[string]any{"priority": 10},
			Claims:             &SecurityClaims{SkipChecks: true},
		})

		mvHealth := InstanceHealthMetricsViewError{
			Version: mv.Meta.StateVersion,
		}
		if err != nil {
			mvHealth.Err = err.Error()
		} else {
			_ = resolverRes.Close()
		}
		res.MetricsViews[mv.Meta.Name.Name] = mvHealth
	}

	// save to cache
	res.ControllerVersion = ctrl.catalog.version
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
		HealthJSON: bytes,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Runtime) cachedInstanceHealth(ctx context.Context, instanceID string, ctrlVersion int64) (*InstanceHealth, bool) {
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
	err = json.Unmarshal(cached.HealthJSON, c)
	if err != nil || ctrlVersion != c.ControllerVersion {
		return nil, false
	}
	return c, true
}

func (r *Runtime) checkRepo(ctx context.Context, inst *drivers.Instance) error {
	h, release, err := r.AcquireHandle(ctx, inst.ID, inst.RepoConnector)
	if err != nil {
		return err
	}
	defer release()

	_, ok := h.AsRepoStore(inst.ID)
	if !ok {
		return fmt.Errorf("connector %q is not a repo connector", inst.RepoConnector)
	}

	return h.Ping(ctx)
}

func (r *Runtime) checkOLAP(ctx context.Context, inst *drivers.Instance, cachedHealth *InstanceHealth) error {
	h, release, err := r.AcquireHandle(ctx, inst.ID, inst.ResolveOLAPConnector())
	if err != nil {
		return err
	}
	defer release()

	olap, ok := h.AsOLAP(inst.ID)
	if !ok {
		return fmt.Errorf("connector %q is not an OLAP connector", inst.ResolveOLAPConnector())
	}

	mayBeScaledToZero := olap.MayBeScaledToZero(ctx)
	if cachedHealth != nil && mayBeScaledToZero {
		if cachedHealth.OLAP != "" {
			return errors.New(cachedHealth.OLAP)
		}
		return nil
	}

	return h.Ping(ctx)
}
