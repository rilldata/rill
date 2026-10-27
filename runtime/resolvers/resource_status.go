package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/pkg/typepb"
	"golang.org/x/exp/maps"
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
	// WhereKind restricts results to resources of the given kinds; accepts shorthands like "model" or "metrics_view".
	WhereKind []string `mapstructure:"where_kind"`
	// WhereName restricts results to resources with the given names (case-insensitive).
	WhereName    []string       `mapstructure:"where_name"`
	UnusedFields map[string]any `mapstructure:",remain"`
}

// newResourceStatus creates a new resourceStatusResolver.
func newResourceStatus(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &resourceStatusProps{}
	if err := mapstructureutil.WeakDecode(opts.Properties, props); err != nil {
		return nil, err
	}
	for i, k := range props.WhereKind {
		props.WhereKind[i] = runtime.ResourceKindFromShorthand(k)
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
	if len(r.props.UnusedFields) > 0 {
		return &runtime.ResolverUnusedFieldsError{
			Name:   "resource_status",
			Fields: maps.Keys(r.props.UnusedFields),
		}
	}
	for _, k := range r.props.WhereKind {
		if !runtime.IsKnownResourceKind(k) {
			return fmt.Errorf("resource_status: unknown resource kind %q", k)
		}
	}
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

	rs = slices.DeleteFunc(rs, func(res *runtimev1.Resource) bool {
		if r.props.WhereError && res.Meta.ReconcileError == "" {
			return true
		}
		if len(r.props.WhereKind) > 0 && !slices.Contains(r.props.WhereKind, res.Meta.Name.Kind) {
			return true
		}
		if len(r.props.WhereName) > 0 && !slices.ContainsFunc(r.props.WhereName, func(n string) bool { return strings.EqualFold(n, res.Meta.Name.Name) }) {
			return true
		}
		return false
	})

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
