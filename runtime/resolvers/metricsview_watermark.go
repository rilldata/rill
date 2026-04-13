package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

const defaultWatermarkCacheTTL = 5 * time.Minute

func init() {
	runtime.RegisterResolverInitializer("metrics_timestamps", newMetricsWatermarkResolver)
}

type metricsWatermarkResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	mvName     string
	mv         *runtimev1.MetricsViewSpec
	executor   *executor.Executor
	args       *metricsWatermarkResolverArgs
}

type metricsWatermarkResolverProps struct {
	MetricsView string `mapstructure:"metrics_view"`
}

type metricsWatermarkResolverArgs struct {
	Table          string `mapstructure:"table"`
	Database       string `mapstructure:"database"`
	DatabaseSchema string `mapstructure:"database_schema"`
	Priority       int    `mapstructure:"priority"`
}

// newMetricsWatermarkResolver returns physical data coverage (min/max timestamps) with open security.
// It bypasses row-level security because watermarks reflect physical data boundaries, not user-visible data.
// Access is restricted to internal callers and admins.
func newMetricsWatermarkResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	if !opts.Claims.SkipChecks && !opts.Claims.Admin() {
		return nil, errors.New("must be an admin to query metrics watermarks")
	}

	props := &metricsWatermarkResolverProps{}
	if err := mapstructureutil.WeakDecode(opts.Properties, props); err != nil {
		return nil, err
	}

	args := &metricsWatermarkResolverArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
		return nil, err
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: props.MetricsView}, false)
	if err != nil {
		return nil, err
	}

	mv := res.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil, fmt.Errorf("metrics view %q is invalid", res.Meta.Name.Name)
	}

	if mv.TimeDimension == "" {
		return nil, fmt.Errorf("metrics view %q has no time dimension", props.MetricsView)
	}

	// If a table override is provided and it differs from the base table, build a synthetic spec
	spec := mv
	if args.Table != "" && !sameTable(mv, args.Table, args.Database, args.DatabaseSchema) {
		spec = executor.BuildSyntheticSpec(mv, &runtimev1.MetricsViewSpec_Rollup{
			Table:          args.Table,
			Database:       args.Database,
			DatabaseSchema: args.DatabaseSchema,
		})
	}

	// Use open security (no row filter); watermark is about physical data coverage
	ex, err := executor.New(ctx, opts.Runtime, opts.InstanceID, props.MetricsView, spec, false, runtime.ResolvedSecurityOpen, args.Priority, nil)
	if err != nil {
		return nil, err
	}

	return &metricsWatermarkResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		mvName:     props.MetricsView,
		mv:         mv,
		executor:   ex,
		args:       args,
	}, nil
}

func (r *metricsWatermarkResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *metricsWatermarkResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	key, ok, err := cacheKeyForMetricsView(ctx, r.runtime, r.instanceID, r.mvName, r.args.Priority)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		// MV-level caching disabled (streaming/external OLAP); use time-bucketed key for TTL caching
		ttl := defaultWatermarkCacheTTL
		if r.mv.WatermarkCacheTtlSeconds > 0 {
			ttl = time.Duration(r.mv.WatermarkCacheTtlSeconds) * time.Second
		}
		bucket := time.Now().Truncate(ttl).Unix()
		key = []byte(fmt.Sprintf("wm:%d", bucket))
	}
	// Append table/database/schema to differentiate rollup entries from the base table entry
	if r.args.Table != "" && !sameTable(r.mv, r.args.Table, r.args.Database, r.args.DatabaseSchema) {
		key = append(key, []byte(":"+r.args.Table+":"+r.args.Database+":"+r.args.DatabaseSchema)...)
	}
	return key, true, nil // always cacheable
}

func (r *metricsWatermarkResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.mvName}}
}

func (r *metricsWatermarkResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsWatermarkResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	ts, err := r.executor.Timestamps(ctx, "")
	if err != nil {
		return nil, err
	}

	row := map[string]any{}
	if !ts.Min.IsZero() {
		row["min"] = ts.Min
		row["max"] = ts.Max
	} else {
		row["min"] = nil
		row["max"] = nil
	}
	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "min", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "max", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
		},
	}
	return runtime.NewMapsResolverResult([]map[string]any{row}, schema), nil
}

func (r *metricsWatermarkResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *metricsWatermarkResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}

// sameTable returns true if the given table/database/schema match the metrics view's base table.
func sameTable(mv *runtimev1.MetricsViewSpec, table, database, databaseSchema string) bool {
	return table == mv.Table && database == mv.Database && databaseSchema == mv.DatabaseSchema
}
