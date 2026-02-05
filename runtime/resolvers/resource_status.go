package resolvers

import (
	"context"
	"errors"
	"io"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/typepb"
)

func init() {
	runtime.RegisterResolverInitializer("resource_status", newResourceStatus)
}

// resourceStatusResolver is a resolver that returns an overview of the instance's resources and their reconcile status.
// The output fields are:
//   - type: the resource type
//   - name: the resource name
//   - status: the reconcile status of the resource (one of "pending", "running", "idle")
//   - error: the error message if the resource is in error state
type resourceStatusResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	props      *resourceStatusProps
}

// resourceStatusProps declares the properties for the "resource_status" resolver.
type resourceStatusProps struct {
	// WhereError is a flag to only return resources that are in the error state.
	WhereError bool `mapstructure:"where_error"`
}

// newResourceStatus creates a new resourceStatusResolver.
func newResourceStatus(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &resourceStatusProps{}
	if err := mapstructure.Decode(opts.Properties, props); err != nil {
		return nil, err
	}

	return &resourceStatusResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		props:      props,
	}, nil
}

func (r *resourceStatusResolver) Close() error {
	return nil
}

func (r *resourceStatusResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *resourceStatusResolver) Refs() []*runtimev1.ResourceName {
	return nil
}

func (r *resourceStatusResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *resourceStatusResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	ctrl, err := r.runtime.Controller(ctx, r.instanceID)
	if err != nil {
		return nil, err
	}

	rs, err := ctrl.List(ctx, "", "", false)
	if err != nil {
		return nil, err
	}

	if r.props.WhereError {
		// In-place trim resources that are not in error state
		i := 0
		for i < len(rs) {
			r := rs[i]
			if r.Meta.ReconcileError == "" {
				// Remove from the slice
				rs[i] = rs[len(rs)-1]
				rs[len(rs)-1] = nil
				rs = rs[:len(rs)-1]
				continue
			}
			rs[i] = r
			i++
		}
	}

	slices.SortFunc(rs, func(a, b *runtimev1.Resource) int {
		an := a.Meta.Name
		bn := b.Meta.Name
		if an.Kind < bn.Kind {
			return -1
		}
		if an.Kind > bn.Kind {
			return 1
		}
		return strings.Compare(an.Name, bn.Name)
	})

	rows := make([]map[string]any, 0, len(rs))
	for _, r := range rs {
		rows = append(rows, map[string]any{
			"type":   runtime.PrettifyResourceKind(r.Meta.Name.Kind),
			"name":   r.Meta.Name.Name,
			"status": runtime.PrettifyReconcileStatus(r.Meta.ReconcileStatus),
			"error":  r.Meta.ReconcileError,
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

func (r *resourceStatusResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *resourceStatusResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}
