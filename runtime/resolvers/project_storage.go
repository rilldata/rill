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
	key := time.Now().Truncate(60 * time.Second).Format(time.RFC3339)
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

	// Get the controller and list connector and metrics view resources separately.
	ctrl, err := r.runtime.Controller(ctx, r.instanceID)
	if err != nil {
		return nil, err
	}

	connectorResources, err := ctrl.List(ctx, runtime.ResourceKindConnector, "", false)
	if err != nil {
		return nil, err
	}

	mvResources, err := ctrl.List(ctx, runtime.ResourceKindMetricsView, "", false)
	if err != nil {
		return nil, err
	}

	// Build a map of connector name to spec (for Provision/driver info).
	connectorSpecs := make(map[string]*runtimev1.ConnectorSpec, len(connectorResources))
	for _, res := range connectorResources {
		connectorSpecs[res.Meta.Name.Name] = res.GetConnector().Spec
	}

	// Collect connector names used by metrics views.
	mvConnectors := make(map[string]bool, len(mvResources))
	for _, res := range mvResources {
		mv := res.GetMetricsView()
		if mv == nil || mv.State == nil || mv.State.ValidSpec == nil {
			continue
		}
		c := mv.State.ValidSpec.Connector
		if c == "" {
			c = defaultOLAP
		}
		mvConnectors[c] = true
	}

	// Build the set of relevant connectors: default OLAP, managed (Provision), or used by a metrics view.
	relevant := make(map[string]bool)
	relevant[defaultOLAP] = true
	for name, spec := range connectorSpecs {
		if spec.Provision {
			relevant[name] = true
		}
	}
	for name := range mvConnectors {
		relevant[name] = true
	}

	// For each relevant connector, open a handle to get the driver name and estimate size.
	rows := make([]map[string]any, 0, len(relevant))
	for name := range relevant {
		handle, release, err := r.runtime.AcquireHandle(ctx, r.instanceID, name)
		if err != nil {
			continue
		}
		driver := handle.Driver()

		olap, ok := handle.AsOLAP(r.instanceID)
		if !ok {
			release()
			continue
		}

		sizeBytes, err := olap.EstimateSize(ctx)
		release()
		if err != nil {
			return nil, err
		}

		// Determine managed status from connector spec; default to true for default OLAP if no spec exists.
		managed := false
		if spec, ok := connectorSpecs[name]; ok {
			managed = spec.Provision
		} else if name == defaultOLAP {
			managed = true
		}

		rows = append(rows, map[string]any{
			"connector":       name,
			"driver":          driver,
			"is_default_olap": name == defaultOLAP,
			"managed":         managed,
			"size_bytes":      sizeBytes,
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
