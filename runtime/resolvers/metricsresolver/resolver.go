package metricsresolver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
)

type Resolver struct {
	query               *Query
	metricsView         *runtimev1.MetricsViewSpec
	security            *runtime.ResolvedMetricsViewSecurity
	olap                drivers.OLAPStore
	olapRelease         func()
	exporting           bool
	priority            int
	executionTime       *time.Time
	interactiveRowLimit int64
}

type resolverProps struct {
	*Query
}

type resolverArgs struct {
	Priority      int        `mapstructure:"priority"`
	ExecutionTime *time.Time `mapstructure:"execution_time"`
}

func New(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &resolverProps{}
	if err := mapstructure.Decode(opts.Properties, props); err != nil {
		return nil, err
	}

	args := &resolverArgs{}
	if err := mapstructure.Decode(opts.Args, args); err != nil {
		return nil, err
	}

	inst, err := opts.Runtime.Instance(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	cfg, err := inst.Config()
	if err != nil {
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

	security, err := opts.Runtime.ResolveMetricsViewSecurity(opts.UserAttributes, opts.InstanceID, mv, res.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return nil, err
	}

	if security != nil && !security.Access {
		return nil, queries.ErrForbidden
	}

	olap, release, err := opts.Runtime.OLAP(ctx, opts.InstanceID, mv.Connector)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		query:               props.Query,
		metricsView:         mv,
		security:            security,
		olap:                olap,
		olapRelease:         release,
		exporting:           opts.ForExport,
		priority:            args.Priority,
		executionTime:       args.ExecutionTime,
		interactiveRowLimit: cfg.InteractiveSQLRowLimit,
	}, nil
}

func (r *Resolver) Close() error {
	r.olapRelease()
	return nil
}

func (r *Resolver) Key() string {
	hash, err := hashstructure.Hash(r.query, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return strconv.FormatUint(hash, 16)
}

func (r *Resolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.query.MetricsView}}
}

func (r *Resolver) Validate(ctx context.Context) error {
	// TODO: E.g. check dims/measures exist?
	return nil
}

func (r *Resolver) ResolveInteractive(ctx context.Context) (*runtime.ResolverResult, error) {
	return nil, errors.New("not implemented")
}

func (r *Resolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}
