package runtime

import (
	"context"
	"encoding/json"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

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
	Controller string `json:"controller"`
	// OLAP error can be cached
	OLAP string `json:"olap"`
	Repo string `json:"repo"`
	// MetricsViews errors can be cached
	MetricsViews      map[string]InstanceHealthMetricsViewError `json:"metrics_views"`
	ParseErrCount     int                                       `json:"parse_error_count"`
	ReconcileErrCount int                                       `json:"reconcile_error_count"`

	// cached health check information can be used if controller version is same and metrics view spec is same
	ControllerVersion int64 `json:"controller_version"`
}

type InstanceHealthMetricsViewError struct {
	Err     string `json:"err"`
	Version int64  `json:"version"`
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
	res := &InstanceHealth{}
	// check repo error
	err := r.pingRepo(ctx, instanceID)
	if err != nil {
		res.Repo = err.Error()
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
	// set to true if any of the olap engines can be scaled to zero
	var canScaleToZero bool

	// check OLAP error
	olap, release, err := r.OLAP(ctx, instanceID, "")
	if err != nil {
		res.OLAP = err.Error()
	} else {
		mayBeScaledToZero := olap.MayBeScaledToZero(ctx)
		canScaleToZero = canScaleToZero || mayBeScaledToZero
		if cachedHealth != nil && mayBeScaledToZero {
			res.OLAP = cachedHealth.OLAP
		} else {
			err = r.pingOLAP(ctx, olap)
			if err != nil {
				res.OLAP = err.Error()
			}
		}
		release()
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
		olap, release, err := r.OLAP(ctx, instanceID, mv.GetMetricsView().State.ValidSpec.Connector)
		if err != nil {
			res.MetricsViews[mv.Meta.Name.Name] = InstanceHealthMetricsViewError{Err: err.Error()}
			release()
			continue
		}
		mayBeScaledToZero := olap.MayBeScaledToZero(ctx)
		canScaleToZero = canScaleToZero || mayBeScaledToZero
		release()

		// only use cached health if the OLAP can be scaled to zero
		if cachedHealth != nil && mayBeScaledToZero {
			mvHealth, ok := cachedHealth.MetricsViews[mv.Meta.Name.Name]
			if ok && mvHealth.Version == mv.Meta.StateVersion {
				res.MetricsViews[mv.Meta.Name.Name] = mvHealth
				continue
			}
		}
		_, err = r.Resolve(ctx, &ResolveOptions{
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

func (r *Runtime) pingRepo(ctx context.Context, instanceID string) error {
	repo, rr, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer rr()
	h, ok := repo.(drivers.Handle)
	if !ok {
		return errors.New("unable to ping repo")
	}
	return h.Ping(ctx)
}

func (r *Runtime) pingOLAP(ctx context.Context, olap drivers.OLAPStore) error {
	h, ok := olap.(drivers.Handle)
	if !ok {
		return errors.New("unable to ping olap")
	}
	return h.Ping(ctx)
}

func (h *InstanceHealth) To() *runtimev1.InstanceHealth {
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
