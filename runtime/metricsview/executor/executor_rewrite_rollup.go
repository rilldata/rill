package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/proto"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/metricsview/executor")

// Rollup rejection reasons: eligibility phase
const (
	rejectGrainNotDerivable     = "grain_not_derivable"
	rejectTimezoneMismatch      = "timezone_mismatch"
	rejectStartNotAligned       = "start_not_aligned"
	rejectDimensionMissing      = "dimension_missing"
	rejectTimeDimensionMissing  = "time_dimension_missing"
	rejectComputedMeasure       = "computed_measure"
	rejectMeasureMissing        = "measure_missing"
	rejectWhereDimensionMissing = "where_dimension_missing"
	rejectTimestampsUnavailable = "timestamps_unavailable"
)

// Rollup rejection reasons: coverage phase
const (
	rejectStartNotCovered = "start_not_covered"
	rejectEndNotCovered   = "end_not_covered"
	rejectEndNotAligned   = "end_not_aligned"
)

// Rollup skip reasons: early disqualification
const (
	skipRawRows                 = "raw_rows"
	skipComparisonTimeRange     = "comparison_time_range"
	skipNonPrimaryTimeDimension = "non_primary_time_dimension"
)

// Rollup routing decides whether a metrics query can be served from a
// pre-aggregated rollup table instead of the base table.
//
// Routing decision:
//
//  1. Quick disqualification: raw-row queries and comparison time range queries are skipped.
//
//  2. Eligibility (per rollup): a rollup is eligible only if all of these hold:
//     a. Query time grain is derivable from the rollup grain (e.g. month from day).
//     b. For day+ grains, the query timezone matches the rollup timezone.
//     c. Query time range start is aligned to the rollup grain.
//     d. All queried dimensions are present in the rollup.
//     e. The time range's time dimension is available in the rollup.
//     f. All queried measures are present; computed measures (count, count_distinct) are rejected.
//     g. All WHERE filter dimensions are present in the rollup.
//
//  3. Time coverage: an eligible rollup must cover the requested time range.
//     For explicit time ranges, the query range is clamped to the base table's actual data range first.
//     For no-time-range queries ("all data"), the rollup must cover the base table's full min/max range.
//     Additionally, if the base table has data beyond the query end, the query end must be aligned
//     to the rollup grain (to prevent the last bucket from pulling in extra data).
//
//  4. Selection: among eligible rollups, prefer the coarsest grain (fewer rows to scan).
//     On ties, prefer the rollup with the smallest data range (tighter coverage).
//
// The selected rollup is returned as a synthetic MetricsViewSpec that points to the rollup table.
// The caller uses this spec to build the query AST, so the rest of the query pipeline remains same.

// rollupCandidate tracks an eligible rollup along with selection metadata.
type rollupCandidate struct {
	rollup     *runtimev1.MetricsViewSpec_Rollup
	grainOrder int
	dataRange  time.Duration // max - min; 0 if no time dimension
}

// rewriteQueryForRollup checks if a rollup table can satisfy the query.
// It returns a synthetic MetricsViewSpec pointing to the rollup table, or nil if no rollup matches.
func (e *Executor) rewriteQueryForRollup(ctx context.Context, qry *metricsview.Query) (*runtimev1.MetricsViewSpec, error) {
	if len(e.metricsView.Rollups) == 0 {
		return nil, nil
	}

	// Disqualify: raw rows queries
	if qry.Rows {
		_, span := tracer.Start(ctx, "rollup.selection")
		span.SetAttributes(
			attribute.Int("rollup.candidate_count", len(e.metricsView.Rollups)),
			attribute.String("rollup.result", "skipped"),
			attribute.String("rollup.skip_reason", skipRawRows),
		)
		span.End()
		return nil, nil
	}

	// Disqualify: queries with comparison time ranges
	if qry.ComparisonTimeRange != nil {
		_, span := tracer.Start(ctx, "rollup.selection")
		span.SetAttributes(
			attribute.Int("rollup.candidate_count", len(e.metricsView.Rollups)),
			attribute.String("rollup.result", "skipped"),
			attribute.String("rollup.skip_reason", skipComparisonTimeRange),
		)
		span.End()
		return nil, nil
	}

	// Disqualify: queries using a non-primary time dimension (rollups are built on the primary)
	if qry.TimeRange != nil && qry.TimeRange.TimeDimension != "" && qry.TimeRange.TimeDimension != e.metricsView.TimeDimension {
		_, span := tracer.Start(ctx, "rollup.selection")
		span.SetAttributes(
			attribute.Int("rollup.candidate_count", len(e.metricsView.Rollups)),
			attribute.String("rollup.result", "skipped"),
			attribute.String("rollup.skip_reason", skipNonPrimaryTimeDimension),
		)
		span.End()
		return nil, nil
	}

	// Extract the time grain from the query (from time floor dimensions)
	queryGrain := extractQueryTimeGrain(qry)

	// Extract dimension names from the WHERE clause
	whereDims := collectWhereDimensions(qry.Where)

	// Determine whether the query has a non-zero time range using start and end to make sure they are resolved. At this point
	// e.RewriteQueryTimeRanges is called earlier, and thus other fields would be unset, and only start and end will be set.
	hasTimeRange := qry.TimeRange != nil && (!qry.TimeRange.Start.IsZero() || !qry.TimeRange.End.IsZero())

	// Parent span for rollup selection
	selectionCtx, selectionSpan := tracer.Start(ctx, "rollup.selection")
	selectionSpan.SetAttributes(attribute.Int("rollup.candidate_count", len(e.metricsView.Rollups)))
	defer selectionSpan.End()

	// Timestamps are fetched lazily on the first eligible rollup. Typically already cached
	// via BindQuery (called by resolvers/metrics.go and queries/metricsview_aggregation.go
	// after resolving through the metrics_time_range resolver). Falls back to querying
	// OLAP directly if not pre-bound.
	var ts metricsview.TimestampsResult
	var baseMin, baseMax time.Time
	tsFetched := false

	var best *rollupCandidate
	for _, rollup := range e.metricsView.Rollups {
		if rollup.Table == "" {
			return nil, fmt.Errorf("rollup for model %q has no resolved table", rollup.Model)
		}

		// Child span per candidate
		_, candidateSpan := tracer.Start(selectionCtx, "rollup.candidate")
		candidateSpan.SetAttributes(
			attribute.String("rollup.table", rollup.Table),
			attribute.String("rollup.grain", rollup.TimeGrain.String()),
			attribute.String("rollup.timezone", rollup.TimeZone),
		)

		rejectCandidate := func(reason string) {
			candidateSpan.SetAttributes(
				attribute.String("rollup.eligible", "false"),
				attribute.String("rollup.reject_reason", reason),
			)
			candidateSpan.End()
		}

		eligible, reason, err := rollupEligible(rollup, qry, queryGrain, whereDims, e.metricsView.TimeDimension, e.metricsView.FirstDayOfWeek)
		if err != nil {
			candidateSpan.SetStatus(codes.Error, err.Error())
			candidateSpan.End()
			return nil, err
		}
		if !eligible {
			rejectCandidate(reason)
			continue
		}

		// Fetch timestamps once, when the first eligible rollup is found
		if !tsFetched {
			tsFetched = true
			ts, err = e.Timestamps(ctx, "")
			if err != nil {
				candidateSpan.SetStatus(codes.Error, err.Error())
				candidateSpan.End()
				return nil, err
			}
			baseMin, baseMax = ts.Min, ts.Max
		}

		rts, ok := ts.Rollups[rollup.Table]
		if !ok {
			rejectCandidate(rejectTimestampsUnavailable)
			continue
		}
		rollupMin, rollupMax := rts.Min, rts.Max

		// Compute rollup effective end: max time + 1 grain period (the max bucket covers up to the next grain boundary)
		rollupLoc := time.UTC
		if rollup.TimeZone != "" {
			loc, err := time.LoadLocation(rollup.TimeZone)
			if err != nil {
				err = fmt.Errorf("invalid timezone %q for rollup %q: %w", rollup.TimeZone, rollup.Table, err)
				candidateSpan.SetStatus(codes.Error, err.Error())
				candidateSpan.End()
				return nil, err
			}
			rollupLoc = loc
		}
		rollupEffEnd := timeutil.OffsetTime(rollupMax, timeutil.TimeGrainFromAPI(rollup.TimeGrain), 1, rollupLoc)

		if hasTimeRange {
			// Clamp query range to the base table's actual data range.
			// This ensures a rollup isn't rejected when the query extends beyond both the base table and rollup.
			effectiveStart := qry.TimeRange.Start
			if !effectiveStart.IsZero() && baseMin.After(effectiveStart) {
				effectiveStart = baseMin
			}
			effectiveEnd := qry.TimeRange.End
			if !effectiveEnd.IsZero() && baseMax.Before(effectiveEnd) {
				effectiveEnd = baseMax
			}

			// Check coverage: rollup must cover the effective (clamped) range
			if !effectiveStart.IsZero() && rollupMin.After(effectiveStart) {
				rejectCandidate(rejectStartNotCovered)
				continue
			}
			if !effectiveEnd.IsZero() && rollupEffEnd.Before(effectiveEnd) {
				rejectCandidate(rejectEndNotCovered)
				continue
			}
		} else {
			// No time range: rollup must cover the base table's full range
			if rollupMin.After(baseMin) {
				rejectCandidate(rejectStartNotCovered)
				continue
			}
			if rollupEffEnd.Before(baseMax) {
				rejectCandidate(rejectEndNotCovered)
				continue
			}
		}

		// End alignment: if data extends beyond the query end and the end is not aligned to the rollup grain,
		// the last rollup bucket would include data beyond the requested range.
		// Essentially it just check if base has data >= query end time, then makes sure the query end time is rollup grain aligned
		if hasTimeRange && !qry.TimeRange.End.IsZero() && !baseMax.Before(qry.TimeRange.End) &&
			!metricsview.TimeAligned(qry.TimeRange.End, rollup.TimeGrain, rollupLoc, e.metricsView.FirstDayOfWeek) {
			rejectCandidate(rejectEndNotAligned)
			continue
		}

		candidateSpan.SetAttributes(attribute.String("rollup.eligible", "true"))
		candidateSpan.End()

		dataRange := rollupMax.Sub(rollupMin)
		c := &rollupCandidate{
			rollup:     rollup,
			grainOrder: metricsview.GrainOrder[rollup.TimeGrain],
			dataRange:  dataRange,
		}

		// Selection priority: coarsest grain (primary); smallest data range (secondary tiebreaker)
		if best == nil || c.grainOrder > best.grainOrder {
			best = c
		} else if c.grainOrder == best.grainOrder && c.dataRange > 0 && (best.dataRange == 0 || c.dataRange < best.dataRange) {
			best = c
		}
	}

	if best == nil {
		selectionSpan.SetAttributes(attribute.String("rollup.result", "none"))
		return nil, nil
	}

	selectionSpan.SetAttributes(
		attribute.String("rollup.result", "selected"),
		attribute.String("rollup.selected_table", best.rollup.Table),
	)
	return buildSyntheticSpec(e.metricsView, best.rollup), nil
}

// rollupEligible checks whether a rollup table can satisfy the given query.
// It returns (eligible, reason, error) where reason is non-empty only when eligible is false.
// primaryTimeDim is the metrics view's default time dimension name (used when query fields omit it).
func rollupEligible(rollup *runtimev1.MetricsViewSpec_Rollup, qry *metricsview.Query, queryGrain runtimev1.TimeGrain, whereDims map[string]bool, primaryTimeDim string, firstDayOfWeek uint32) (bool, string, error) {
	// 1. Grain derivable: if query has a time grain, it must be derivable from rollup grain
	if queryGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		if !metricsview.GrainDerivableFrom(queryGrain, rollup.TimeGrain) {
			return false, rejectGrainNotDerivable, nil
		}
	}

	// 2. For day+ rollup grains, the query timezone must match the rollup's timezone.
	// Sub-day grains are timezone-agnostic (hour boundaries are the same everywhere).
	if rollup.TimeGrain >= runtimev1.TimeGrain_TIME_GRAIN_DAY {
		rollupTZ, err := normalizeTimezone(rollup.TimeZone)
		if err != nil {
			return false, "", err
		}
		queryTZ, err := normalizeTimezone(qry.TimeZone)
		if err != nil {
			return false, "", err
		}
		if rollupTZ != queryTZ {
			return false, rejectTimezoneMismatch, nil
		}
	}

	// 3. Start time aligned to rollup grain (use rollup timezone for alignment).
	// End alignment is checked conditionally in the coverage phase: only when the base table
	// has data beyond the query end (to prevent the last rollup bucket from pulling in extra data).
	if qry.TimeRange != nil && !qry.TimeRange.Start.IsZero() {
		rollupLoc := time.UTC
		if rollup.TimeZone != "" {
			loc, err := time.LoadLocation(rollup.TimeZone)
			if err != nil {
				return false, "", fmt.Errorf("invalid timezone %q for rollup %q: %w", rollup.TimeZone, rollup.Table, err)
			}
			rollupLoc = loc
		}
		if !metricsview.TimeAligned(qry.TimeRange.Start, rollup.TimeGrain, rollupLoc, firstDayOfWeek) {
			return false, rejectStartNotAligned, nil
		}
	}

	// 4. All query dimensions present in rollup
	rollupDims := make(map[string]bool, len(rollup.Dimensions))
	for _, d := range rollup.Dimensions {
		rollupDims[strings.ToLower(d)] = true
	}
	// dimInRollup checks if a dimension is available in the rollup: either it's the primary
	// time dimension (always present as the rollup's time column) or it's in the dimensions list.
	dimInRollup := func(dim string) bool {
		if strings.EqualFold(dim, primaryTimeDim) {
			return true
		}
		return rollupDims[strings.ToLower(dim)]
	}

	for _, d := range qry.Dimensions {
		if d.Compute != nil && d.Compute.TimeFloor != nil {
			// TimeFloor references an underlying time dimension; it must be available in the rollup
			if !dimInRollup(d.Compute.TimeFloor.Dimension) {
				return false, rejectTimeDimensionMissing, nil
			}
			continue
		}
		if !rollupDims[strings.ToLower(d.Name)] {
			return false, rejectDimensionMissing, nil
		}
	}

	// 5. Time range's time dimension must be available in the rollup
	trTimeDim := primaryTimeDim
	if qry.TimeRange != nil && qry.TimeRange.TimeDimension != "" {
		trTimeDim = qry.TimeRange.TimeDimension
	}
	if trTimeDim != "" && !dimInRollup(trTimeDim) {
		return false, rejectTimeDimensionMissing, nil
	}

	// 6. All queried measures present in rollup; reject computed measures (count, count_distinct, etc.)
	// since they produce incorrect results on pre-aggregated rollup tables.
	rollupMeasures := make(map[string]bool, len(rollup.Measures))
	for _, m := range rollup.Measures {
		rollupMeasures[strings.ToLower(m)] = true
	}
	for _, m := range qry.Measures {
		if m.Compute != nil {
			return false, rejectComputedMeasure, nil
		}
		if !rollupMeasures[strings.ToLower(m.Name)] {
			return false, rejectMeasureMissing, nil
		}
	}

	// 7. All WHERE dimensions present in rollup
	for dim := range whereDims {
		if !rollupDims[strings.ToLower(dim)] {
			return false, rejectWhereDimensionMissing, nil
		}
	}

	return true, "", nil
}

// buildSyntheticSpec creates a MetricsViewSpec that points to the rollup table.
// Since rollup tables have the same column names as the base table, the base measure expressions work directly against the rollup table.
// Note - This function does not rewrite dimensions/measures in the spec even though Rollups can have less dimensions/measures than the base
// i.e. because rollupEligible check before this will skip rollup if the query references dims/measure not in rollup so keeping it simple here.
// Also, this is an internal function, if we ever export it then it would make sense to do full rewrite.
func buildSyntheticSpec(original *runtimev1.MetricsViewSpec, rollup *runtimev1.MetricsViewSpec_Rollup) *runtimev1.MetricsViewSpec {
	synth := proto.Clone(original).(*runtimev1.MetricsViewSpec)

	// Point to rollup table (connector stays the same as base)
	synth.Table = rollup.Table
	synth.Model = ""
	if rollup.Database != "" {
		synth.Database = rollup.Database
	}
	if rollup.DatabaseSchema != "" {
		synth.DatabaseSchema = rollup.DatabaseSchema
	}

	// Clear rollups to prevent recursion
	synth.Rollups = nil

	return synth
}

// extractQueryTimeGrain finds the smallest time grain across all time floor dimensions in the query.
// When multiple TimeFloor dimensions exist (e.g. pivot tables with multiple time levels),
// the smallest grain determines the rollup grain requirement.
func extractQueryTimeGrain(qry *metricsview.Query) runtimev1.TimeGrain {
	smallest := runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
	for _, d := range qry.Dimensions {
		if d.Compute == nil || d.Compute.TimeFloor == nil {
			continue
		}
		g := d.Compute.TimeFloor.Grain.ToProto()
		if g == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			continue
		}
		if smallest == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED || metricsview.GrainOrder[g] < metricsview.GrainOrder[smallest] {
			smallest = g
		}
	}
	return smallest
}

// collectWhereDimensions recursively collects dimension names referenced in a WHERE expression.
func collectWhereDimensions(expr *metricsview.Expression) map[string]bool {
	dims := make(map[string]bool)
	collectWhereDimensionsRec(expr, dims)
	return dims
}

func collectWhereDimensionsRec(expr *metricsview.Expression, dims map[string]bool) {
	if expr == nil {
		return
	}
	if expr.Name != "" {
		dims[expr.Name] = true
	}
	if expr.Condition != nil {
		for _, sub := range expr.Condition.Expressions {
			collectWhereDimensionsRec(sub, dims)
		}
	}
	if expr.Subquery != nil {
		if expr.Subquery.Dimension.Name != "" {
			dims[expr.Subquery.Dimension.Name] = true
		}
		collectWhereDimensionsRec(expr.Subquery.Where, dims)
		collectWhereDimensionsRec(expr.Subquery.Having, dims)
	}
}

// normalizeTimezone validates and normalizes a timezone string for comparison.
// It normalizes UTC variants (empty, "UTC", "Etc/UTC") to "UTC".
// Note: Go's time.LoadLocation preserves the input name, so aliases like "US/Eastern"
// are not resolved to "America/New_York". Users should use canonical IANA names.
func normalizeTimezone(tz string) (string, error) {
	if tz == "" || strings.EqualFold(tz, "UTC") {
		return "UTC", nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	name := loc.String()
	if name == "Etc/UTC" || name == "Etc/GMT" {
		return "UTC", nil
	}
	return name, nil
}
