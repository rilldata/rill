package executor

import (
	"context"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"google.golang.org/protobuf/proto"
)

// Rollup routing decides whether a metrics query can be served from a
// pre-aggregated rollup table instead of the base table.
//
// Routing decision:
//
//  1. Quick disqualification: raw-row queries are not routed to rollups for now, and comparison time ranges
//     queries are also avoided for simplification.
//
//  2. Eligibility (per rollup): a rollup is eligible only if all of these hold:
//     a. Query time grain is derivable from the rollup grain (e.g. month from day).
//     b. For day+ grains, the query timezone matches the rollup timezone.
//     c. Query time range boundaries are aligned to the rollup grain.
//     d. All queried dimensions (including WHERE filter dimensions) are present in the rollup.
//     e. All queried measures are present; computed measures (count, count_distinct) are rejected.
//
//  3. Time coverage: an eligible rollup must cover the requested time range.
//     For explicit time ranges, the query range is clamped to the base table's actual data range first.
//     For no-time-range queries ("all data"), the rollup must cover the base table's full min/max range.
//
//  4. Selection: among eligible rollups, prefer the coarsest grain (fewer rows to scan).
//     On ties, prefer the rollup with the smallest data range (tighter coverage).
//
// The selected rollup is returned as a synthetic MetricsViewSpec that points to the rollup table.
// The caller uses this spec to build the query AST, so the rest of the query pipeline remains same.

// rollupRewrite holds the result of rewriting a query for a rollup.
// spec is set to the synthetic MetricsViewSpec pointing to the rollup table.
type rollupRewrite struct {
	spec *runtimev1.MetricsViewSpec
}

// rollupCandidate tracks an eligible rollup along with selection metadata.
type rollupCandidate struct {
	rollup     *runtimev1.MetricsViewSpec_RollupTable
	grainOrder int
	dataRange  time.Duration // max - min; 0 if no time dimension
}

// rewriteQueryForRollup checks if a rollup table can satisfy the query.
// It returns a rollupRewrite with a synthetic spec, or nil if no rollup matches.
func (e *Executor) rewriteQueryForRollup(ctx context.Context, qry *metricsview.Query) *rollupRewrite {
	if len(e.metricsView.Rollups) == 0 {
		return nil
	}

	// Disqualify: raw rows queries
	if qry.Rows {
		return nil
	}

	// Disqualify: queries with comparison time ranges (future improvement)
	if qry.ComparisonTimeRange != nil {
		return nil
	}

	// Extract the time grain from the query (from time floor dimensions)
	queryGrain := extractQueryTimeGrain(qry)

	// Extract dimension names from the WHERE clause
	whereDims := collectWhereDimensions(qry.Where)

	// Determine whether the query has a non-zero time range
	hasTimeRange := qry.TimeRange != nil && (!qry.TimeRange.Start.IsZero() || !qry.TimeRange.End.IsZero())

	grainOrderMap := map[runtimev1.TimeGrain]int{
		runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND: 0,
		runtimev1.TimeGrain_TIME_GRAIN_SECOND:      1,
		runtimev1.TimeGrain_TIME_GRAIN_MINUTE:      2,
		runtimev1.TimeGrain_TIME_GRAIN_HOUR:        3,
		runtimev1.TimeGrain_TIME_GRAIN_DAY:         4,
		runtimev1.TimeGrain_TIME_GRAIN_WEEK:        5,
		runtimev1.TimeGrain_TIME_GRAIN_MONTH:       6,
		runtimev1.TimeGrain_TIME_GRAIN_QUARTER:     7,
		runtimev1.TimeGrain_TIME_GRAIN_YEAR:        8,
	}

	// Fetch base table timestamps for clamping the query range.
	// Uses the shared watermark cache (same cache as rollup watermarks).
	// Also needed for no-time-range queries: we treat them as "all data" and
	// verify the rollup covers the base table's full range.
	var baseTS watermarkEntry
	var hasBaseTS bool
	if e.metricsView.TimeDimension != "" {
		if baseMin, baseMax, ok := e.fetchBaseWatermark(ctx); ok {
			baseTS = watermarkEntry{min: baseMin, max: baseMax}
			hasBaseTS = true
		}
	}

	// For no-time-range queries on time-dimensioned views, we need base watermarks
	// to verify rollup coverage. If unavailable, fall back to the base table.
	if !hasTimeRange && e.metricsView.TimeDimension != "" && !hasBaseTS {
		return nil
	}

	var best *rollupCandidate
	for _, rollup := range e.metricsView.Rollups {
		if rollup.Table == "" {
			continue // not yet resolved?
		}
		if !rollupEligible(rollup, qry, queryGrain, whereDims, e.metricsView.FirstDayOfWeek) {
			continue
		}

		var dataRange time.Duration
		if e.metricsView.TimeDimension != "" {
			rollupMin, rollupMax, ok := e.fetchRollupWatermark(ctx, rollup)
			if !ok {
				continue // could not fetch watermarks; skip this rollup
			}

			// Compute rollup effective end: max time + 1 grain period (the max bucket covers up to the next grain boundary)
			rollupLoc := time.UTC
			if rollup.Timezone != "" {
				if loc, err := time.LoadLocation(rollup.Timezone); err == nil {
					rollupLoc = loc
				}
			}
			rollupEffEnd := timeutil.OffsetTime(rollupMax, timeutil.TimeGrainFromAPI(rollup.TimeGrain), 1, rollupLoc)

			if hasTimeRange {
				// Clamp query range to the base table's actual data range.
				// This ensures a rollup isn't rejected when the query extends beyond both the base table and rollup.
				effectiveStart := qry.TimeRange.Start
				if hasBaseTS && !effectiveStart.IsZero() && !baseTS.min.IsZero() && baseTS.min.After(effectiveStart) {
					effectiveStart = baseTS.min
				}
				effectiveEnd := qry.TimeRange.End
				if hasBaseTS && !effectiveEnd.IsZero() && !baseTS.max.IsZero() && baseTS.max.Before(effectiveEnd) {
					effectiveEnd = baseTS.max
				}

				// Check coverage: rollup must cover the effective (clamped) range
				if !effectiveStart.IsZero() && rollupMin.After(effectiveStart) {
					continue
				}
				if !effectiveEnd.IsZero() && rollupEffEnd.Before(effectiveEnd) {
					continue
				}
			} else {
				// No time range: rollup must cover the base table's full range
				if !baseTS.min.IsZero() && rollupMin.After(baseTS.min) {
					continue
				}
				if !baseTS.max.IsZero() && rollupEffEnd.Before(baseTS.max) {
					continue
				}
			}

			dataRange = rollupMax.Sub(rollupMin)
		}

		c := &rollupCandidate{
			rollup:     rollup,
			grainOrder: grainOrderMap[rollup.TimeGrain],
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
		return nil
	}

	return &rollupRewrite{spec: buildSyntheticSpec(e.metricsView, best.rollup)}
}

// rollupEligible checks whether a rollup table can satisfy the given query.
func rollupEligible(rollup *runtimev1.MetricsViewSpec_RollupTable, qry *metricsview.Query, queryGrain runtimev1.TimeGrain, whereDims map[string]bool, firstDayOfWeek uint32) bool {
	// 1. Grain derivable: if query has a time grain, it must be derivable from rollup grain
	if queryGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		if !metricsview.GrainDerivableFrom(queryGrain, rollup.TimeGrain) {
			return false
		}
	}

	// 2. For day+ rollup grains, the query timezone must match the rollup's timezone.
	// Sub-day grains are timezone-agnostic (hour boundaries are the same everywhere).
	if rollup.TimeGrain >= runtimev1.TimeGrain_TIME_GRAIN_DAY {
		rollupTZ := normalizeTimezone(rollup.Timezone)
		queryTZ := normalizeTimezone(qry.TimeZone)
		if rollupTZ != queryTZ {
			return false
		}
	}

	// 3. Time range aligned to rollup grain (use rollup timezone for alignment)
	if qry.TimeRange != nil {
		rollupLoc := time.UTC
		if rollup.Timezone != "" {
			if loc, err := time.LoadLocation(rollup.Timezone); err == nil {
				rollupLoc = loc
			}
		}
		if !metricsview.TimeRangeAligned(qry.TimeRange.Start, qry.TimeRange.End, rollup.TimeGrain, rollupLoc, firstDayOfWeek) {
			return false
		}
	}

	// 4. All query dimensions present in rollup
	rollupDims := make(map[string]bool, len(rollup.Dimensions))
	for _, d := range rollup.Dimensions {
		rollupDims[strings.ToLower(d)] = true
	}
	for _, d := range qry.Dimensions {
		name := d.Name
		if d.Compute != nil && d.Compute.TimeFloor != nil {
			// Time floor dimensions reference the underlying time dimension; skip for dimension check
			// (the time dimension column exists in the rollup table as the time column)
			continue
		}
		if !rollupDims[strings.ToLower(name)] {
			return false
		}
	}

	// 5. All queried measures present in rollup; reject computed measures (count, count_distinct, etc.)
	// since they produce incorrect results on pre-aggregated rollup tables.
	rollupMeasures := make(map[string]bool, len(rollup.Measures))
	for _, m := range rollup.Measures {
		rollupMeasures[strings.ToLower(m)] = true
	}
	for _, m := range qry.Measures {
		if m.Compute != nil {
			return false // computed measures are invalid on rollup tables
		}
		if !rollupMeasures[strings.ToLower(m.Name)] {
			return false
		}
	}

	// 6. All WHERE dimensions present in rollup
	for dim := range whereDims {
		if !rollupDims[strings.ToLower(dim)] {
			return false
		}
	}

	return true
}

// extractQueryTimeGrain finds the time grain from the query's dimensions.
// It returns the grain from the first time floor dimension found, or UNSPECIFIED.
func extractQueryTimeGrain(qry *metricsview.Query) runtimev1.TimeGrain {
	for _, d := range qry.Dimensions {
		if d.Compute != nil && d.Compute.TimeFloor != nil {
			return d.Compute.TimeFloor.Grain.ToProto()
		}
	}
	return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
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

// buildSyntheticSpec creates a MetricsViewSpec that points to the rollup table.
// Since rollup tables have the same column names as the base table, the base measure expressions work directly against the rollup table.
func buildSyntheticSpec(original *runtimev1.MetricsViewSpec, rollup *runtimev1.MetricsViewSpec_RollupTable) *runtimev1.MetricsViewSpec {
	synth := proto.Clone(original).(*runtimev1.MetricsViewSpec)

	// Point to rollup table
	synth.Table = rollup.Table
	synth.Model = ""
	if rollup.Connector != "" {
		synth.Connector = rollup.Connector
	}
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

// normalizeTimezone returns a canonical timezone string for comparison. Empty, "UTC", and "Etc/UTC" are all treated as equivalent.
func normalizeTimezone(tz string) string {
	if tz == "" || strings.EqualFold(tz, "UTC") || strings.EqualFold(tz, "Etc/UTC") {
		return "UTC"
	}
	return tz
}
