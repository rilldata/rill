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
	skipNonPrimaryTimeDimension = "non_primary_time_dimension"
	skipExactMeasures           = "exact_measures"
)

// Rollup routing decides whether a metrics query can be served from a
// pre-aggregated rollup table instead of the base table.
//
// Routing decision:
//
//  1. Quick disqualification: raw-row queries, queries referencing a non-primary time dimension, and queries that compute measures exactly are skipped.
//
//  2. Eligibility (per rollup): a rollup is eligible only if all of these hold:
//     a. Query time grain is derivable from the rollup grain (e.g. month from day).
//     b. For day+ grains, the query timezone matches the rollup timezone.
//     c. Both the base and (when present) comparison time range starts are aligned to the rollup grain.
//        When the metrics view sets base_time_grain and a rollup rejects the query's own bounds,
//        the checks here and in step 3 are repeated with the bounds rounded up to that grain,
//        since the base table has no rows between its boundaries.
//     d. All queried dimensions are present in the rollup.
//     e. The base and comparison time range time dimensions are available in the rollup.
//     f. All queried measures are present. Computed measures are rejected except for comparison-derived
//        computes (ComparisonValue / ComparisonDelta / ComparisonRatio / ComparisonTime), which are
//        allowed as long as their referenced base measure is in the rollup.
//     g. All WHERE filter dimensions are present in the rollup. The primary time dimension never counts as present here,
//        since the rollup's time column is truncated to its grain; time bounds in the WHERE are moved into the time range earlier by rewriteQueryWhereTimeRange.
//
//  3. Time coverage: an eligible rollup must cover both the base and (when present) comparison time ranges.
//     An open start or end (including a query without a time range) is checked against the base table's min or max.
//     The query range is clamped to the widest available source: the start to min(baseMin, rollupMin) and
//     the end to max(baseMax, rollupEffEnd), where the base data is treated as ending just after its last row (baseMax),
//     since rollupEffEnd is exclusive. So a rollup that extends further back (or forward) than the
//     base is preserved here, not rejected. Additionally, if the base table has data beyond the query range's end,
//     that end must be aligned to the rollup grain (to prevent the last bucket from pulling in extra data).
//
//  4. Selection: selection is based on coarsest grain, then earlier definition order. The
//     exception is when the query time range goes beyond what the candidate rollups cover — i.e.
//     rollups have mismatched data time range depths in that case the rollup with the wider effective return range wins.
//
// The selected rollup is returned as a synthetic MetricsViewSpec that points to the rollup table.
// The caller uses this spec to build the query AST, so the rest of the query pipeline remains same.
//
// Assumption: a rollup's max timestamp is no later than the base table's max. Coverage and end-alignment checks use
// the base's max as the reference for "is there data at or beyond the query end" — a rollup that gets ahead of the base
// may return incorrect results at the tail of the data.

// rollupCandidate tracks an eligible rollup along with selection metadata.
type rollupCandidate struct {
	rollup      *runtimev1.MetricsViewSpec_Rollup
	grainOrder  int
	index       int       // position in MetricsViewSpec.Rollups; used as the final tiebreaker (earlier wins)
	effRetStart time.Time // earliest timestamp this rollup will actually return for this query
	effRetEnd   time.Time // latest timestamp this rollup will actually return for this query (exclusive)
}

// rewriteQueryForRollup checks if a rollup table can satisfy the query.
// It returns a synthetic MetricsViewSpec pointing to the rollup table, or nil if no rollup matches.
func (e *Executor) rewriteQueryForRollup(ctx context.Context, qry *metricsview.Query) (*runtimev1.MetricsViewSpec, error) {
	e.latestQueryTable = e.metricsView.Table

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

	// Disqualify: queries using a non-primary time dimension (rollups are built on the primary).
	// The comparison time range should have the same TimeDimension as the base, so checking TimeRange is sufficient.
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

	// Disqualify: queries that compute measures with their exact_expression.
	// Exact aggregates (e.g. exact distinct counts) cannot be computed from pre-aggregated rows.
	// This is keyed on the exact measures resolved by rewriteQueryExactMeasures, so an `exact` flag on a measure without an exact_expression keeps the rollup eligible.
	if len(qry.ExactMeasures) > 0 {
		_, span := tracer.Start(ctx, "rollup.selection")
		span.SetAttributes(
			attribute.Int("rollup.candidate_count", len(e.metricsView.Rollups)),
			attribute.String("rollup.result", "skipped"),
			attribute.String("rollup.skip_reason", skipExactMeasures),
		)
		span.End()
		return nil, nil
	}

	// Extract the time grain from the query (from time floor dimensions)
	queryGrain := extractQueryTimeGrain(qry)

	// Extract dimension names from the WHERE clause
	whereDims := collectWhereDimensions(qry.Where)

	// Determine whether the query has a non-zero time range / comparison time range using start and end to make sure they are resolved.
	// At this point e.RewriteQueryTimeRanges is called earlier, and thus other fields would be unset, and only start and end will be set.
	hasTimeRange := qry.TimeRange != nil && (!qry.TimeRange.Start.IsZero() || !qry.TimeRange.End.IsZero())
	hasComparisonTimeRange := qry.ComparisonTimeRange != nil && (!qry.ComparisonTimeRange.Start.IsZero() || !qry.ComparisonTimeRange.End.IsZero())

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
	fetchTimestamps := func() error {
		if tsFetched {
			return nil
		}
		var err error
		ts, err = e.Timestamps(ctx, "")
		if err != nil {
			return err
		}
		tsFetched = true
		baseMin, baseMax = ts.Min, ts.Max
		return nil
	}

	// When the base table stores time at base_time_grain, it has no rows between two boundaries of that grain,
	// so a time bound between two boundaries selects the same rows as the next boundary.
	// A rollup that rejects the query's own bounds is checked again with the bounds rounded up to the next boundary.
	// The query keeps its own bounds, which select the same rows from the base table and from the selected rollup.
	rounded := qry
	baseLoc := time.UTC
	if e.metricsView.BaseTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		if e.metricsView.BaseTimeZone != "" {
			loc, err := time.LoadLocation(e.metricsView.BaseTimeZone)
			if err != nil {
				return nil, fmt.Errorf("invalid base time zone %q: %w", e.metricsView.BaseTimeZone, err)
			}
			baseLoc = loc
		}
		rounded = roundQueryTimeRanges(qry, e.metricsView.BaseTimeGrain, baseLoc, e.metricsView.FirstDayOfWeek)
	}

	// Rounding returns wrong results if the base table has rows between boundaries.
	// Checking the base table's min and max timestamps catches some wrong declarations for free.
	// With a data_time_range, the timestamps are the declared range rather than the table's actual bounds, so there is nothing to check.
	// It must be called after the timestamps are fetched.
	baseGrainChecked, baseGrainOK := false, false
	checkBaseGrain := func() bool {
		if !baseGrainChecked {
			baseGrainChecked = true
			baseGrainOK = e.metricsView.DataTimeRange != "" ||
				(ceilTime(baseMin, e.metricsView.BaseTimeGrain, baseLoc, e.metricsView.FirstDayOfWeek).Equal(baseMin) &&
					ceilTime(baseMax, e.metricsView.BaseTimeGrain, baseLoc, e.metricsView.FirstDayOfWeek).Equal(baseMax))
			if !baseGrainOK {
				selectionSpan.SetAttributes(attribute.Bool("rollup.base_time_grain_mismatch", true))
			}
		}
		return baseGrainOK
	}

	var best *rollupCandidate
	for i, rollup := range e.metricsView.Rollups {
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

		failCandidate := func(err error) error {
			candidateSpan.SetStatus(codes.Error, err.Error())
			candidateSpan.End()
			return err
		}

		rollupLoc := time.UTC
		if rollup.TimeZone != "" {
			loc, err := time.LoadLocation(rollup.TimeZone)
			if err != nil {
				return nil, failCandidate(fmt.Errorf("invalid timezone %q for rollup %q: %w", rollup.TimeZone, rollup.Table, err))
			}
			rollupLoc = loc
		}

		// checkQuery runs the eligibility, coverage and end alignment checks against q's time ranges.
		// It returns the reject reason and trace attributes, or "" if the rollup can serve q.
		var rollupMin, rollupEffEnd time.Time
		checkQuery := func(q *metricsview.Query) (string, []attribute.KeyValue, error) {
			eligible, reason, err := rollupEligible(rollup, q, queryGrain, whereDims, e.metricsView.TimeDimension, e.metricsView.FirstDayOfWeek)
			if err != nil {
				return "", nil, err
			}
			if !eligible {
				return reason, nil, nil
			}

			if err := fetchTimestamps(); err != nil {
				return "", nil, err
			}
			rts, ok := ts.Rollups[rollup.Table]
			if !ok {
				return rejectTimestampsUnavailable, nil, nil
			}
			// The rollup's effective end is its max time + 1 grain period (the max bucket covers up to the next grain boundary)
			rollupMin = rts.Min
			rollupEffEnd = timeutil.OffsetTime(rts.Max, timeutil.TimeGrainFromAPI(rollup.TimeGrain), 1, rollupLoc)

			// Check coverage for the base time range and the comparison time range when present.
			// An open start or end (including a query without a time range) must be covered up to the edge of the data;
			// otherwise a partial rollup would silently truncate the result.
			if reason, attrs := checkRangeCoverage(q.TimeRange, baseMin, baseMax, rollupMin, rollupEffEnd); reason != "" {
				return reason, attrs, nil
			}
			if hasComparisonTimeRange {
				if reason, attrs := checkRangeCoverage(q.ComparisonTimeRange, baseMin, baseMax, rollupMin, rollupEffEnd); reason != "" {
					return reason, attrs, nil
				}
			}

			// Check end alignment now: if data extends beyond the query end and the end is not aligned to the rollup grain,
			// the last rollup bucket would include data beyond the requested range. rollupEligible before only checked start alignment.
			if hasTimeRange {
				if reason, attrs := checkEndAlignment(q.TimeRange.End, baseMax, rollup.TimeGrain, rollupLoc, e.metricsView.FirstDayOfWeek); reason != "" {
					return reason, attrs, nil
				}
			}
			if hasComparisonTimeRange {
				if reason, attrs := checkEndAlignment(q.ComparisonTimeRange.End, baseMax, rollup.TimeGrain, rollupLoc, e.metricsView.FirstDayOfWeek); reason != "" {
					return reason, attrs, nil
				}
			}
			return "", nil, nil
		}

		reason, attrs, err := checkQuery(qry)
		if err != nil {
			return nil, failCandidate(err)
		}
		// Rounded bounds are only tried after the query's own bounds are rejected, so rounding never rejects a rollup that would otherwise be used.
		// The base table check runs last, since checkQuery only fetches the timestamps it needs once the rollup is otherwise eligible.
		if reason != "" && rounded != qry {
			roundedReason, _, err := checkQuery(rounded)
			if err != nil {
				return nil, failCandidate(err)
			}
			if roundedReason == "" && checkBaseGrain() {
				reason, attrs = "", nil
				candidateSpan.SetAttributes(attribute.Bool("rollup.base_time_grain_rounded", true))
			}
		}
		if reason != "" {
			candidateSpan.SetAttributes(
				attribute.String("rollup.eligible", "false"),
				attribute.String("rollup.reject_reason", reason),
			)
			candidateSpan.SetAttributes(attrs...)
			candidateSpan.End()
			continue
		}

		candidateSpan.SetAttributes(attribute.String("rollup.eligible", "true"))
		candidateSpan.End()

		// Compute the candidate's effective return range: the slice of time this rollup will actually emit for this query.
		// The rollup is filtered to [max(query.start, rollupMin), min(query.end, rollupEffEnd)].
		// For two candidates with identical effective return ranges — selection falls to coarsest grain and then index order.
		// It uses the rounded bounds for every candidate, so candidates checked with and without rounding compare equally.
		effRetStart := rollupMin
		effRetEnd := rollupEffEnd
		if hasTimeRange {
			if !rounded.TimeRange.Start.IsZero() && rounded.TimeRange.Start.After(effRetStart) {
				effRetStart = rounded.TimeRange.Start
			}
			if !rounded.TimeRange.End.IsZero() && rounded.TimeRange.End.Before(effRetEnd) {
				effRetEnd = rounded.TimeRange.End
			}
		}

		c := &rollupCandidate{
			rollup:      rollup,
			grainOrder:  grainOrder[rollup.TimeGrain],
			index:       i,
			effRetStart: effRetStart,
			effRetEnd:   effRetEnd,
		}

		// Selection: coarsest grain wins, then earlier definition order. The exception is when the query time range goes beyond what the candidate rollups cover;
		// in that case the wider effective return range wins (earliest start, then latest end) as it will give the most complete data.
		// When multiple rollups cover the query time range, they will all have the same effRetStart and effRetEnd so selection is base on grain and index order.
		switch {
		case best == nil:
			best = c
		case c.effRetStart.Before(best.effRetStart):
			best = c
		case c.effRetStart.Equal(best.effRetStart) && c.effRetEnd.After(best.effRetEnd):
			best = c
		case c.effRetStart.Equal(best.effRetStart) && c.effRetEnd.Equal(best.effRetEnd) && c.grainOrder > best.grainOrder:
			best = c
		case c.effRetStart.Equal(best.effRetStart) && c.effRetEnd.Equal(best.effRetEnd) && c.grainOrder == best.grainOrder && c.index < best.index:
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
	e.latestQueryTable = best.rollup.Table
	return buildSyntheticSpec(e.metricsView, best.rollup), nil
}

// rollupEligible checks whether a rollup table can satisfy the given query.
// It returns (eligible, reason, error) where reason is non-empty only when eligible is false.
// primaryTimeDim is the metrics view's default time dimension name (used when query fields omit it).
func rollupEligible(rollup *runtimev1.MetricsViewSpec_Rollup, qry *metricsview.Query, queryGrain runtimev1.TimeGrain, whereDims map[string]bool, primaryTimeDim string, firstDayOfWeek uint32) (bool, string, error) {
	// 1. Grain derivable: if query has a time grain, it must be derivable from rollup grain
	if queryGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		if !grainDerivableFrom(queryGrain, rollup.TimeGrain) {
			return false, rejectGrainNotDerivable, nil
		}
	}

	// 2. For day+ rollup grains, the query timezone must match the rollup's timezone.
	// Sub-day grains are timezone-agnostic (hour boundaries are the same everywhere).
	if timeutil.TimeGrainFromAPI(rollup.TimeGrain).Compare(timeutil.TimeGrainDay) >= 0 {
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

	// 3. Start time aligned to rollup grain (use rollup timezone for alignment) for both the base
	// and comparison ranges. End alignment is checked conditionally in the coverage phase: only when
	// the base table has data beyond the query end (to prevent the last rollup bucket from pulling in extra data).
	rollupLoc := time.UTC
	if rollup.TimeZone != "" {
		loc, err := time.LoadLocation(rollup.TimeZone)
		if err != nil {
			return false, "", fmt.Errorf("invalid timezone %q for rollup %q: %w", rollup.TimeZone, rollup.Table, err)
		}
		rollupLoc = loc
	}
	if qry.TimeRange != nil && !qry.TimeRange.Start.IsZero() {
		if !timeAligned(qry.TimeRange.Start, rollup.TimeGrain, rollupLoc, firstDayOfWeek) {
			return false, rejectStartNotAligned, nil
		}
	}
	if qry.ComparisonTimeRange != nil && !qry.ComparisonTimeRange.Start.IsZero() {
		if !timeAligned(qry.ComparisonTimeRange.Start, rollup.TimeGrain, rollupLoc, firstDayOfWeek) {
			return false, rejectStartNotAligned, nil
		}
	}

	// 4. All query dimensions present in rollup.
	// The primary time dimension is left out of rollupDims even when the rollup lists it (e.g. with default or '*' dimensions):
	// the rollup's time column is truncated to its grain, so it can't serve the raw time dimension or filters on it.
	rollupDims := make(map[string]bool, len(rollup.Dimensions))
	for _, d := range rollup.Dimensions {
		if strings.EqualFold(d, primaryTimeDim) {
			continue
		}
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
	// since they produce incorrect results on pre-aggregated rollup tables. Comparison-derived
	// computed measures (ComparisonValue / ComparisonDelta / ComparisonRatio / ComparisonTime) and
	// expression measures are allowed because they are pure SQL wrappers over referenced base measures:
	// the actual aggregation is done by those measures, which must themselves be in the rollup.
	rollupMeasures := make(map[string]bool, len(rollup.Measures))
	for _, m := range rollup.Measures {
		rollupMeasures[strings.ToLower(m)] = true
	}
	for _, m := range qry.Measures {
		if m.Compute != nil {
			refNames, ok := rollupSafeReferencedMeasures(m.Compute)
			if !ok {
				// Compute that cannot be re-aggregated from a rollup (count, count_distinct, percent_of_total, etc.)
				return false, rejectComputedMeasure, nil
			}
			for _, refName := range refNames {
				// A comparison compute may reference an expression measure defined in the same query.
				// That measure is not in the rollup, but the measures it references may be, so check those instead.
				// Its own rollup safety is checked when the loop reaches it.
				if _, ok := queryExpressionMeasure(qry, refName); ok {
					continue
				}
				if !rollupMeasures[strings.ToLower(refName)] {
					return false, rejectMeasureMissing, nil
				}
			}
			continue
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
		if smallest == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED || timeutil.TimeGrainFromAPI(g).Compare(timeutil.TimeGrainFromAPI(smallest)) < 0 {
			smallest = g
		}
	}
	return smallest
}

// roundQueryTimeRanges returns a shallow copy of qry with the start and end of its time ranges rounded up to boundaries of the given grain.
// It returns qry itself when every bound is already on a boundary.
func roundQueryTimeRanges(qry *metricsview.Query, grain runtimev1.TimeGrain, tz *time.Location, firstDayOfWeek uint32) *metricsview.Query {
	round := func(tr *metricsview.TimeRange) *metricsview.TimeRange {
		if tr == nil {
			return nil
		}
		start := ceilTime(tr.Start, grain, tz, firstDayOfWeek)
		end := ceilTime(tr.End, grain, tz, firstDayOfWeek)
		if start.Equal(tr.Start) && end.Equal(tr.End) {
			return tr
		}
		res := *tr
		res.Start, res.End = start, end
		return &res
	}
	tr := round(qry.TimeRange)
	ctr := round(qry.ComparisonTimeRange)
	if tr == qry.TimeRange && ctr == qry.ComparisonTimeRange {
		return qry
	}
	res := *qry
	res.TimeRange, res.ComparisonTimeRange = tr, ctr
	return &res
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

// checkRangeCoverage verifies that the rollup covers the given time range. The query range is clamped to the widest
// available data range — i.e. min(baseMin, rollupMin) on the start and max(baseEnd, rollupEffEnd) on the end
// so a rollup isn't rejected when the query asks for data outside what either source can serve.
// A nil range or an open (zero) start or end extends to the edge of that data range.
// Returns the reject reason and trace attributes, or "" if covered.
func checkRangeCoverage(tr *metricsview.TimeRange, baseMin, baseMax, rollupMin, rollupEffEnd time.Time) (string, []attribute.KeyValue) {
	if tr == nil {
		tr = &metricsview.TimeRange{}
	}
	deepestStart := baseMin
	if rollupMin.Before(baseMin) {
		deepestStart = rollupMin
	}
	effectiveStart := tr.Start
	if effectiveStart.IsZero() || deepestStart.After(effectiveStart) {
		effectiveStart = deepestStart
	}
	// baseMax is the timestamp of the last row, so it is inclusive, while rollupEffEnd is exclusive.
	// Treat the base data as ending just after baseMax, so a rollup must end strictly after the last base row to cover it.
	baseEnd := baseMax.Add(time.Nanosecond)
	furthestEnd := baseEnd
	if rollupEffEnd.After(baseEnd) {
		furthestEnd = rollupEffEnd
	}
	effectiveEnd := tr.End
	if effectiveEnd.IsZero() || effectiveEnd.After(furthestEnd) {
		effectiveEnd = furthestEnd
	}
	if !effectiveStart.IsZero() && rollupMin.After(effectiveStart) {
		return rejectStartNotCovered, []attribute.KeyValue{
			attribute.String("rollup.effective_start", effectiveStart.Format(time.RFC3339)),
			attribute.String("rollup.rollup_min", rollupMin.Format(time.RFC3339)),
		}
	}
	if !effectiveEnd.IsZero() && rollupEffEnd.Before(effectiveEnd) {
		return rejectEndNotCovered, []attribute.KeyValue{
			attribute.String("rollup.effective_end", effectiveEnd.Format(time.RFC3339)),
			attribute.String("rollup.rollup_eff_end", rollupEffEnd.Format(time.RFC3339)),
		}
	}
	return "", nil
}

// checkEndAlignment enforces grain alignment on a range's end when the base table has data at or
// beyond that end (otherwise the last rollup bucket would pull in data past the requested range).
// Returns the reject reason and trace attributes, or "" if no enforcement is needed or it passes.
func checkEndAlignment(end, baseMax time.Time, grain runtimev1.TimeGrain, loc *time.Location, firstDayOfWeek uint32) (string, []attribute.KeyValue) {
	if end.IsZero() || baseMax.Before(end) {
		return "", nil
	}
	if timeAligned(end, grain, loc, firstDayOfWeek) {
		return "", nil
	}
	return rejectEndNotAligned, []attribute.KeyValue{
		attribute.String("rollup.query_end", end.Format(time.RFC3339)),
		attribute.String("rollup.base_max", baseMax.Format(time.RFC3339)),
		attribute.String("rollup.rollup_grain", grain.String()),
	}
}

// rollupSafeReferencedMeasures returns the base measures referenced by a computed measure,
// along with ok=true if the compute is a type that's safe to evaluate against a rollup.
// Returns (nil, true) for ComparisonTime and URI since they have no measure dependency.
// Returns (nil, false) for computes that cannot be correctly re-aggregated from a pre-aggregated rollup table
// (count, count_distinct, percent_of_total).
func rollupSafeReferencedMeasures(c *metricsview.MeasureCompute) ([]string, bool) {
	switch {
	case c.ComparisonValue != nil:
		return []string{c.ComparisonValue.Measure}, true
	case c.ComparisonDelta != nil:
		return []string{c.ComparisonDelta.Measure}, true
	case c.ComparisonRatio != nil:
		return []string{c.ComparisonRatio.Measure}, true
	case c.ComparisonTime != nil:
		return nil, true
	case c.URI != nil:
		return nil, true
	case c.Expression != nil:
		parsed, err := metricsview.ParseMeasureExpression(c.Expression.Expression)
		if err != nil {
			// The expression is validated again when the AST is built, which reports the error properly.
			// Here we only need to keep an invalid expression from taking the rollup path.
			return nil, false
		}
		return parsed.Refs(), true
	}
	return nil, false
}

// queryExpressionMeasure returns the expression measure with the given name defined in the query, if any.
func queryExpressionMeasure(qry *metricsview.Query, name string) (metricsview.Measure, bool) {
	for _, qm := range qry.Measures {
		if qm.Name == name && qm.Compute != nil && qm.Compute.Expression != nil {
			return qm, true
		}
	}
	return metricsview.Measure{}, false
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
