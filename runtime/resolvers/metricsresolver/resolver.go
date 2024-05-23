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
	dialect             drivers.Dialect
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
		dialect:             olap.Dialect(),
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

func (r *Resolver) BuildAST(ctx context.Context) (*AST, error) {
	err := r.rewriteQueryTimeRanges(ctx)
	if err != nil {
		return nil, err
	}

	ast, err := buildAST(r.metricsView, r.security, r.query, r.dialect)
	if err != nil {
		return nil, fmt.Errorf("failed to build AST: %w", err)
	}

	return ast, nil
}

func (r *Resolver) rewriteQueryTimeRanges(_ context.Context) error {
	// TODO: Use qry.time_zone, r.executionTime, if necessary, resolve the watermark or start/end time of the MV
	// TODO: If resolving watermark, cache it to avoid repeat for comparison time range
	panic("not implemented")
}

// func ResolveTimeRange(tr *runtimev1.TimeRange, mv *runtimev1.MetricsViewSpec) (time.Time, time.Time, error) {
// 	tz := time.UTC

// 	if tr.TimeZone != "" {
// 		var err error
// 		tz, err = time.LoadLocation(tr.TimeZone)
// 		if err != nil {
// 			return time.Time{}, time.Time{}, fmt.Errorf("invalid time_range.time_zone %q: %w", tr.TimeZone, err)
// 		}
// 	}

// 	var start, end time.Time
// 	if tr.Start != nil {
// 		start = tr.Start.AsTime().In(tz)
// 	}
// 	if tr.End != nil {
// 		end = tr.End.AsTime().In(tz)
// 	}

// 	isISO := false

// 	if tr.IsoDuration != "" {
// 		if !start.IsZero() && !end.IsZero() {
// 			return time.Time{}, time.Time{}, fmt.Errorf("only two of time_range.{start,end,iso_duration} can be specified")
// 		}

// 		d, err := duration.ParseISO8601(tr.IsoDuration)
// 		if err != nil {
// 			return time.Time{}, time.Time{}, fmt.Errorf("invalid iso_duration %q: %w", tr.IsoDuration, err)
// 		}

// 		if !start.IsZero() {
// 			end = d.Add(start)
// 		} else if !end.IsZero() {
// 			start = d.Sub(end)
// 		} else {
// 			return time.Time{}, time.Time{}, fmt.Errorf("one of time_range.{start,end} must be specified with time_range.iso_duration")
// 		}

// 		isISO = true
// 	}

// 	if tr.IsoOffset != "" {
// 		d, err := duration.ParseISO8601(tr.IsoOffset)
// 		if err != nil {
// 			return time.Time{}, time.Time{}, fmt.Errorf("invalid iso_offset %q: %w", tr.IsoOffset, err)
// 		}

// 		if !start.IsZero() {
// 			start = d.Sub(start)
// 		}
// 		if !end.IsZero() {
// 			end = d.Sub(end)
// 		}

// 		isISO = true
// 	}

// 	// Only modify the start and end if ISO duration or offset was sent.
// 	// This is to maintain backwards compatibility for calls from the UI.
// 	if isISO {
// 		fdow := int(mv.FirstDayOfWeek)
// 		if mv.FirstDayOfWeek > 7 || mv.FirstDayOfWeek <= 0 {
// 			fdow = 1
// 		}
// 		fmoy := int(mv.FirstMonthOfYear)
// 		if mv.FirstMonthOfYear > 12 || mv.FirstMonthOfYear <= 0 {
// 			fmoy = 1
// 		}
// 		if !start.IsZero() {
// 			start = timeutil.TruncateTime(start, convTimeGrain(tr.RoundToGrain), tz, fdow, fmoy)
// 		}
// 		if !end.IsZero() {
// 			end = timeutil.TruncateTime(end, convTimeGrain(tr.RoundToGrain), tz, fdow, fmoy)
// 		}
// 	}

// 	return start, end, nil
// }
