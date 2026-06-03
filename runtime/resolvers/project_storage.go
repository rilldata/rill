package resolvers

import (
	"context"
	"errors"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/typepb"
)

// _projectStorageDefaultCacheTTL is the default TTL for caching project storage results.
// Caching ensures we don't hit the drivers too frequently with heavy metadata queries.
// A hard-coded TTL without invalidation should be sufficient for our current use cases (billing UI and storage usage alerts).
// It is a best effort TTL that uses fixed-size buckets.
const _projectStorageDefaultCacheTTL = 60 * time.Second

func init() {
	runtime.RegisterResolverInitializer("project_storage", newProjectStorage)
}

type projectStorageResolver struct {
	runtime    *runtime.Runtime
	instanceID string
}

func newProjectStorage(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	return &projectStorageResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
	}, nil
}

func (r *projectStorageResolver) Close() error {
	return nil
}

func (r *projectStorageResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	// Simple TTL. See _projectStorageDefaultCacheTTL for details.
	key := time.Now().Truncate(_projectStorageDefaultCacheTTL).Format(time.RFC3339)
	return []byte(key), true, nil
}

func (r *projectStorageResolver) Refs() []*runtimev1.ResourceName {
	return nil
}

func (r *projectStorageResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *projectStorageResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	// Get the instance to determine the default OLAP connector.
	inst, err := r.runtime.Instance(ctx, r.instanceID)
	if err != nil {
		return nil, err
	}
	defaultOLAP := inst.ResolveOLAPConnector()

	// Build a map of connector specs
	ctrl, err := r.runtime.Controller(ctx, r.instanceID)
	if err != nil {
		return nil, err
	}
	rs, err := ctrl.List(ctx, runtime.ResourceKindConnector, "", false)
	if err != nil {
		return nil, err
	}
	connectors := make(map[string]*runtimev1.ConnectorSpec, len(rs))
	for _, res := range rs {
		connectors[res.Meta.Name.Name] = res.GetConnector().Spec
	}

	// Build a set of relevant connector names
	relevant := make(map[string]bool)
	// 1. The default OLAP
	relevant[defaultOLAP] = true
	// 2. Managed connectors
	for name, spec := range connectors {
		if spec.Provision {
			relevant[name] = true
		}
	}
	// 3. Connectors used by metrics views
	rs, err = ctrl.List(ctx, runtime.ResourceKindMetricsView, "", false)
	if err != nil {
		return nil, err
	}
	for _, res := range rs {
		mv := res.GetMetricsView()
		if mv == nil || mv.State == nil || mv.State.ValidSpec == nil {
			continue
		}
		c := mv.State.ValidSpec.Connector
		if c != "" {
			relevant[c] = true
		}
	}

	// For each relevant connector, open a handle to get the driver name and estimate size.
	rows := make([]map[string]any, 0, len(relevant))
	for name := range relevant {
		isDefault := name == defaultOLAP
		isManaged := false
		if spec, ok := connectors[name]; ok {
			isManaged = spec.Provision
		} else if name == "duckdb" {
			// Backwards compatibility: some projects don't have an explicit connector file for managed DuckDB.
			isManaged = true
		}

		sizeBytes, driver, err := r.resolveForConnector(ctx, name)
		errMsg := ""
		if err != nil {
			if errors.Is(err, ctx.Err()) {
				return nil, err
			}
			errMsg = err.Error()
		}
		rows = append(rows, map[string]any{
			"connector":       name,
			"driver":          driver,
			"is_default_olap": isDefault,
			"managed":         isManaged,
			"size_bytes":      sizeBytes,
			"error":           errMsg,
		})
	}

	var schema *runtimev1.StructType
	if len(rows) > 0 {
		schema = typepb.InferFromValue(rows[0]).StructType
	} else {
		schema = &runtimev1.StructType{}
	}

	return runtime.NewMapsResolverResult(rows, schema), nil
}

func (r *projectStorageResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *projectStorageResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}

func (r *projectStorageResolver) resolveForConnector(ctx context.Context, name string) (size int64, driver string, err error) {
	handle, release, err := r.runtime.AcquireHandle(ctx, r.instanceID, name)
	if err != nil {
		return -1, "unknown", err
	}
	defer release()

	olap, ok := handle.AsOLAP(r.instanceID)
	if !ok {
		return -1, handle.Driver(), errors.New("not an OLAP connector")
	}

	sizeBytes, err := olap.EstimateSize(ctx)
	if err != nil {
		return -1, handle.Driver(), err
	}

	return sizeBytes, handle.Driver(), nil
}
