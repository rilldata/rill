# Legacy Queries — Code Review
**Scope:** `runtime/queries/*.go` (metrics-view aggregation/toplist/comparison/timeseries/totals/rows/search/schema, column profiling: cardinality, null-count, descriptive stats, numeric/rug histograms, time-grain, time-range, top-k, table head/columns/cardinality, column timeseries, export helpers, proto adapters, watermark). Cross-referenced `runtime/server/queries_*.go`, `runtime/server/downloads.go`, `runtime/metricsview/executor`, and the relevant driver dialects.  |  **Health:** fair

## Overview
This package holds Rill's pre-resolver "predefined query" implementations. The metrics-view queries are now thin adapters that rewrite the legacy request shape into a `metricsview.Query` and delegate to the `metricsview/executor`; the column-profiling queries still build dialect-specific SQL by hand and run it directly against the raw table. The delegation refactor is mostly clean, but the adapters and the older profiling code carry several latent panics from unguarded pointer dereferences and unchecked type assertions, all reachable from authenticated APIs (exports, reports, profiling). The biggest concrete risk is a nil-pointer panic in the metrics-view aggregation export path when no time range is supplied. Secondary risks are dialect edge cases (single-row StarRocks columns, millisecond grain) that panic instead of erroring, plus a fair amount of dead code left over from the pre-executor implementation. SQL construction is generally parameterized or identifier-escaped, with the notable exception of the profiling APIs that intentionally accept raw SQL expressions (gated by `ReadProfiling`).

## Findings

### queries-01 [HIGH] `MetricsViewAggregation.Export` panics when the request has no time range

**Status (2026-07-25):** ✅ Fixed and merged in [#9715](https://github.com/rilldata/rill/pull/9715).

- **Category:** bug  **Confidence:** high  **Location:** `runtime/queries/metricsview_aggregation.go:166` (and `:523`)
- **Problem:** `Export` dereferences `q.TimeRange.TimeDimension` guarded only by `if mv.ValidSpec.TimeDimension != ""`, with no nil check on `q.TimeRange`. The sibling `Resolve` path was fixed to guard this (lines 102–105 build `qryTimeDim` defensively), but `Export` was not. A second unguarded deref exists in `generateExportHeaders` at line 523 (`qry.TimeRange.Start`); `qry.TimeRange` is only populated when `q.TimeRange != nil` (rewrite at line 369), and `RewriteQueryTimeRanges` returns early without allocating it when the metrics view has no time dimension (`executor_rewrite_time.go:16`).
- **Failure scenario:** `runtime/server/downloads.go:192` sets `tr := r.TimeRange`, which is `nil` when the export request carries no `TimeRange`/`TimeStart`/`TimeEnd`. `downloadHandler` then calls `q.Export`. If the metrics view has a time dimension (the common case), line 166 dereferences a nil `*runtimev1.TimeRange` and panics, aborting the download. If the metrics view has *no* time dimension but `IncludeHeader` is set (the report-export path), line 166 is skipped but line 523 panics instead. Between the two sites, an aggregation export with a nil `TimeRange` crashes regardless of whether the view is time-dimensioned.
- **Recommendation:** Mirror the `Resolve` guard: compute `qryTimeDim := ""` then `if q.TimeRange != nil { qryTimeDim = q.TimeRange.TimeDimension }`, and pass `qryTimeDim`. Also nil-check `qry.TimeRange` before reading `.Start/.End` in `generateExportHeaders`.

**Verification:** CONFIRMED — `metricsview_aggregation.go:165-166` dereferences `q.TimeRange.TimeDimension` guarded only by `if mv.ValidSpec.TimeDimension != ""`, whereas the sibling `Resolve` path (lines 102-105) defensively computes `qryTimeDim` with a `q.TimeRange != nil` check. `downloads.go:192` sets `tr := r.TimeRange` and only reassigns it when `r.TimeStart`/`r.TimeEnd` are set, so an aggregation export request with none of the three leaves `TimeRange` nil. The second site at `metricsview_aggregation.go:523` reads `qry.TimeRange.Start` after `RewriteQueryTimeRanges`; `executor_rewrite_time.go:16` returns early without allocating `qry.TimeRange` when the view has no time dimension, so with `IncludeHeader` set on a non-time-dimensioned view it panics there instead.

**Manual Reproduction Steps (for QA):**
1. Open a Rill dashboard backed by a metrics view that has a time dimension.
2. Trigger an aggregation export (the dashboard's "Export" -> CSV/XLSX/Parquet on a pivot/table view) constructed so the request carries no time range (e.g. a view/report with no time filter applied, or an API call to `POST /v1/instances/{instance_id}/queries/export` with a `MetricsViewAggregationRequest` and no `TimeRange`/`TimeStart`/`TimeEnd`).
3. Expected: a downloaded file. Actual: the download fails / the runtime returns a 500 because `Export` panics on the nil `q.TimeRange` at `metricsview_aggregation.go:166`.
4. For the second site: configure a scheduled report export with "include header" enabled on a metrics view that has NO time dimension, then export it. Expected: file with header rows. Actual: panic at `metricsview_aggregation.go:523` reading `qry.TimeRange.Start`.

### queries-02 [MEDIUM] Millisecond time grain panics the timeseries null-fill via an unchecked type assertion
- **Category:** bug  **Confidence:** high  **Location:** `runtime/queries/metricsview_timeseries.go:346,377` + `runtime/queries/timeutil.go:34`
- **Problem:** `timeGrainToDuration` returns `duration.InfDuration{}` for `TIME_GRAIN_MILLISECOND` (millisecond is explicitly commented out as "not supported") and for `TIME_GRAIN_UNSPECIFIED`. `addTo`/`addToNew` then do `sd := d.(duration.StandardDuration)` — a concrete type assertion without the `,ok` form — which panics for `InfDuration`. `addTo` is called unconditionally for every result row (`start = addTo(t, dur, tz)`, line 281), not only during null filling.
- **Failure scenario:** A `MetricsViewTimeSeries` request with `TimeGranularity = TIME_GRAIN_MILLISECOND`. The handler at `runtime/server/queries_metrics.go` performs no grain validation, and the executor accepts millisecond: `TimeGrain.Valid()` returns true (`metricsview/query.go:313`), the AST does not reject it, and DuckDB's `DateTruncExpr` emits valid `date_trunc('MILLISECOND', …)` SQL. Rows come back, the first row hits line 281, `d.(duration.StandardDuration)` panics on the `InfDuration`.
- **Refuted sub-claim:** The prior report also flagged `TIME_GRAIN_UNSPECIFIED`. That case does *not* reach the panic: the executor rejects an unspecified-grain time floor up front (`metricsview/ast.go:383`, `"grain" must be specified for time floor`), so `e.Query` returns a clean error before any row is produced. Only millisecond survives to the panic.
- **Recommendation:** Validate the grain up front (reject/clamp millisecond), or make `addTo`/`addToNew` use the comma-ok form and fall back gracefully. Ideally `timeGrainToDuration` should return an error for unsupported grains rather than a sentinel that later panics.

**Verification:** CONFIRMED — `timeutil.go:12-35` explicitly comments out `TIME_GRAIN_MILLISECOND` (line 15) so `timeGrainToDuration` falls through to `return duration.InfDuration{}` (line 34) for it. `metricsview_timeseries.go:281` calls `addTo(t, dur, tz)` unconditionally for every result row, and `addTo` (line 346) / `addToNew` (line 377) do `sd := d.(duration.StandardDuration)`, a concrete type assertion that panics on `InfDuration`. `TimeGrain.Valid()` returns true for millisecond (`metricsview/query.go:315`) and the DuckDB dialect maps it to `"MILLISECOND"` (`drivers/dialect.go:128-129`), so the request reaches row processing. The server handler (`queries_metrics.go:251`) passes `TimeGranularity` through with no grain validation. The `TIME_GRAIN_UNSPECIFIED` sub-claim is correctly noted as refuted.

**Manual Reproduction Steps (for QA):**
1. Issue a metrics-view timeseries query with millisecond granularity, e.g. `POST /v1/instances/{instance_id}/queries/metrics-views/{name}/timeseries` with `"time_granularity": "TIME_GRAIN_MILLISECOND"`, a valid measure, and a small time range that returns at least one row.
2. Expected: either results bucketed by millisecond or a clean 4xx error rejecting the grain. Actual: the runtime panics at `metricsview_timeseries.go:346` (`d.(duration.StandardDuration)` on an `InfDuration`), surfacing as a 500 / dropped connection.
3. If no UI surface offers millisecond grain, confirm in code: `runtime/queries/timeutil.go:34` returns `InfDuration{}` for millisecond and `metricsview_timeseries.go:346` asserts to `StandardDuration` without the comma-ok form.

### queries-03 [MEDIUM] Descriptive statistics dereferences `*sd` guarded only on `min`, panicking on single-row StarRocks columns
- **Category:** bug  **Confidence:** medium  **Location:** `runtime/queries/column_desc_stats.go:88-96`
- **Problem:** The code checks `if minVal != nil` and then dereferences `maxVal`, `q25`, `q50`, `q75`, `mean`, and `sd`. On StarRocks, `stddev_samp` (sample standard deviation) is undefined for n=1 and returns SQL `NULL` under MySQL semantics, which StarRocks follows (`runtime/drivers/starrocks/dialect.go`: `CAST(stddev_samp(...) AS DOUBLE) as sd`). For a column where exactly one row survives the `NOT NULL` filter, `min`/`max`/`avg`/quantiles are non-null but `sd` is nil, so `stats.Sd = *sd` (line 95) panics.
- **Refuted sub-claim:** The prior report also implicated ClickHouse. ClickHouse's `stddevSamp` returns `nan` (a non-null Float64), not `NULL`, for a single row, so `sd` scans as a non-nil `*float64` and does not panic — the same reason DuckDB is safe (it hardcodes `'NaN'::DOUBLE as sd`). The bug is StarRocks-specific.
- **Failure scenario:** Profiling a StarRocks-backed model column that has a single non-null numeric value (e.g., a mostly-empty column) → `minVal != nil`, `sd == nil` → nil-pointer panic at line 95.
- **Recommendation:** Guard each optional pointer independently (leave the corresponding proto field at its zero value when nil), or coalesce NULL stddev to `0`/`NaN` in the StarRocks dialect SQL as DuckDB already does.

**Verification:** PLAUSIBLE — the code pattern is exactly as described: `column_desc_stats.go:88-97` checks only `if minVal != nil` and then unconditionally dereferences `maxVal`, `q25`, `q50`, `q75`, `mean`, and `sd` (line 95: `stats.Sd = *sd`). The StarRocks dialect emits `CAST(stddev_samp(%[1]s) AS DOUBLE) as sd` (`drivers/starrocks/dialect.go:114`), which yields SQL `NULL` for a single-row column under MySQL semantics, unlike DuckDB (`'NaN'::DOUBLE as sd`, `dialect.go:203`) and ClickHouse (`stddevSamp(...)::DOUBLE`, `dialect.go:254`, returns `nan`). The panic is real but conditional on a StarRocks connector plus a column with exactly one non-null value, so it is latent rather than always-triggered; the ClickHouse refutation is correct.

**Manual Reproduction Steps (for QA):**
1. Connect a Rill project to a StarRocks OLAP connector with a model/table containing a numeric column where exactly one row is non-null (all other rows NULL).
2. Open the column profiling / descriptive statistics for that column (hover the column in the model preview, or `GET /v1/instances/{instance_id}/queries/descriptive-statistics/tables/{table}?column_name=...`).
3. Expected: statistics with `sd` = 0 or NaN. Actual: nil-pointer panic at `column_desc_stats.go:95` because `stddev_samp` of one row returns SQL NULL while `min`/`max`/`avg` are non-null.
4. Not reproducible on DuckDB or ClickHouse (they return NaN, not NULL). Confirm in code: `runtime/queries/column_desc_stats.go:88-96` vs `runtime/drivers/starrocks/dialect.go:114`.

### queries-04 [MEDIUM] Profiling APIs inline caller-supplied SQL expressions (`ColumnTopK.Agg`, column-timeseries measures)
- **Category:** security  **Confidence:** high  **Location:** `runtime/queries/column_topk.go:72`, `runtime/queries/column_timeseries.go:641`
- **Problem:** `ColumnTopK` interpolates `q.Agg` raw into `SELECT %s AS value, %s AS count ...` (the column and table are escaped, `Agg` is not), and `getExpressionColumnsFromMeasures` inlines each `measure.Expression` verbatim. Both values arrive over the wire: `ColumnTopKRequest.Agg` (`runtime/server/queries_columns.go`, copied into `q.Agg`) and `ColumnTimeSeriesRequest.Measures[].Expression` (`runtime/server/queries_timeseries.go`, copied into `q.Measures`). These run directly against the underlying table, bypassing the metrics-view security policy (row filters / field access). The only gate is the `ReadProfiling` claim (verified: both handlers return `ErrForbidden` unless `Can(runtime.ReadProfiling)`).
- **Failure scenario:** A principal holding `ReadProfiling` sends `Agg: "(SELECT secret FROM other_table LIMIT 1)"` or a subquery in a measure expression and reads arbitrary data from the OLAP connection — data the metrics-view policy layer would otherwise restrict. Any future path that reaches these queries without `ReadProfiling`, or any deployment that grants `ReadProfiling` more broadly than intended, becomes arbitrary SQL execution. (Impact is bounded today because `ReadProfiling` is a developer/edit-level permission, not a dashboard-viewer permission.)
- **Recommendation:** This is legacy "expressions are SQL" behavior, so treat it as defense-in-depth: keep these queries strictly behind `ReadProfiling`, document that the expression fields are trusted SQL, and (longer term) constrain `Agg` to an allowlist of aggregation functions over the escaped column rather than a free-form string.

**Verification:** CONFIRMED — `column_topk.go:72-78` interpolates `q.Agg` raw into `SELECT %s AS value, %s AS count FROM %s ...` (only the column via `EscapeIdentifier` and the table via `EscapeTable` are escaped; `q.Agg` is not). `column_timeseries.go:638-641` (`getExpressionColumnsFromMeasures`) inlines `measure.Expression` verbatim into the SELECT, used at lines 289, 342, and 807. Both handlers gate on `ReadProfiling` (`queries_columns.go` and `queries_timeseries.go` return `ErrForbidden` unless `Can(runtime.ReadProfiling)`), so impact is bounded to that developer-level permission today. The characterization as defense-in-depth is accurate.

**Manual Reproduction Steps (for QA):**
1. As a principal with `ReadProfiling` (developer/edit access), call the top-K API `POST /v1/instances/{instance_id}/queries/topk/tables/{table}` with `agg` set to a SQL subquery, e.g. `"agg": "(SELECT max(secret) FROM other_table)"`, or the profiling timeseries API with a `measures[].expression` containing a subquery.
2. Expected (if hardened): the request is rejected or the expression is treated as an opaque column reference. Actual: the raw SQL executes directly against the OLAP table, returning data outside the metrics-view security policy (row filters / field access are bypassed).
3. Confirm in code: `runtime/queries/column_topk.go:74` (`q.Agg` un-escaped in the SELECT) and `runtime/queries/column_timeseries.go:641` (`measure.Expression` inlined).

### queries-05 [MEDIUM] Parquet export of a metrics-view timeseries builds a schema with an untyped time column

**Status (2026-07-25):** ✅ Fixed and merged in [#9725](https://github.com/rilldata/rill/pull/9725).

- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/queries/metricsview_timeseries.go:178-183` + `runtime/queries/metricsview.go:310-355`
- **Problem:** `Export` prepends a meta column `{Name: spec.TimeDimension}` with an empty `Type`, then stores the timestamp as a string in each row's `Records`. In `WriteParquet`, `runtimev1.Type_Code_value[""]` → `CODE_UNSPECIFIED`, which matches no `case` in the schema-building switch (lines 316–349), so `arrowField.Type` stays a nil `arrow.DataType`. `array.NewRecordBuilder(mem, schema)` (line 355) then constructs a builder for a field whose type is nil, which panics (arrow's `NewBuilder` dereferences the nil `DataType`). CSV and XLSX are unaffected because they read struct fields by name and ignore the meta type. (`MetricsViewRows` parquet export is safe — it routes through the executor's own parquet writer, not this helper.)
- **Failure scenario:** Download a `MetricsViewTimeSeries` with `format = EXPORT_FORMAT_PARQUET` (wired at `runtime/server/downloads.go:272`, no format guard in `Export`) → panic / failed export.
- **Recommendation:** Set the prepended time column's `Type` to `TIMESTAMP` (and store an actual timestamp value), or exclude Parquet for this query type the way `Totals`/profiling queries return `ErrExportNotSupported`.

**Verification:** CONFIRMED — `metricsview_timeseries.go:178-180` prepends `{Name: spec.TimeDimension}` with no `Type` field (empty string), and line 182 stores the timestamp as a string via `structpb.NewStringValue(...Format(time.RFC3339Nano))`. In `WriteParquet` (`metricsview.go:310-355`), `runtimev1.Type_Code_value[""]` resolves to `CODE_UNSPECIFIED` (value 0), which matches no `case` in the switch (lines 316-349), leaving `arrowField.Type` as a nil `arrow.DataType`. `array.NewRecordBuilder(mem, schema)` at line 355 then builds a field with a nil type and panics. CSV/XLSX paths (lines 190-192) read fields by name and are unaffected. Parquet is reachable with no format guard in this `Export`.

**Manual Reproduction Steps (for QA):**
1. Open a dashboard/report backed by a metrics view with a time dimension.
2. Export a timeseries query as Parquet, e.g. `POST /v1/instances/{instance_id}/queries/export` with a `MetricsViewTimeSeriesRequest` and `"format": "EXPORT_FORMAT_PARQUET"`.
3. Expected: a valid `.parquet` file. Actual: the export fails / 500 because `array.NewRecordBuilder` panics on the nil arrow type for the untyped time column (`metricsview.go:355`).
4. Confirm the same export as CSV or XLSX succeeds (those ignore the meta type).

### queries-06 [LOW] Unexpected timestamp column type panics instead of returning an error
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/queries/metricsview_timeseries.go:256`, `runtime/queries/column_timeseries.go:196`
- **Problem:** When scanning the time column, both files `panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))` for any type outside a small known set. A driver returning a timestamp as a string (some OLAP dialects do) turns a recoverable query error into a panic in the query goroutine. Reachability is low in practice — the truncated time column is produced by `date_trunc`/`time_bucket` and normally scans as `time.Time` — but the inconsistency is real, since `resolveStarRocks` handles the identical situation correctly.
- **Failure scenario:** A dialect/column whose timestamp scans as `string` reaches `populateResult`/`Resolve` and panics rather than surfacing "unexpected type" as an error.
- **Recommendation:** Return `fmt.Errorf(...)` instead of `panic` (as `resolveStarRocks` already does at `column_timeseries.go:850`).

**Verification:** PLAUSIBLE — both panics exist as described: `metricsview_timeseries.go:256` (`panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))` in the `default` of the type switch, reached only when `v != nil`) and `column_timeseries.go:196` (same panic, `default` case). The StarRocks path handles the identical situation cleanly with `return fmt.Errorf("unexpected type for timestamp column: %T", v)` at `column_timeseries.go:850`, confirming the inconsistency. Reachability is low: the time column is produced by `date_trunc`/`time_bucket` and normally scans as `time.Time`, so a driver would have to return the truncated timestamp as an unexpected Go type (e.g. string) to trigger it.

**Manual Reproduction Steps (for QA):**
1. Not readily user-reproducible via the UI — it requires an OLAP driver/dialect that scans the truncated time-bucket column as a type other than `time.Time`/`*time.Time`/`int64` (e.g. a string), which the supported dialects do not do.
2. Confirm in code: `runtime/queries/metricsview_timeseries.go:256` and `runtime/queries/column_timeseries.go:196` `panic` in the `default` type-switch case, versus the graceful `return fmt.Errorf(...)` at `column_timeseries.go:850`.

### queries-07 [LOW] `ColumnTimeseries.Resolve` leaks the result cursor on the `rows.Err()` error path
- **Category:** correctness  **Confidence:** high  **Location:** `runtime/queries/column_timeseries.go:217-219`
- **Problem:** Every in-loop error path calls `rows.Close()` before returning (lines 186, 195, 202, 208), but the post-loop `if err := rows.Err(); err != nil { return err }` (lines 217–219) returns without closing `rows` (the manual `rows.Close()` is on line 222, after the check). There is no `defer rows.Close()` for this query.
- **Failure scenario:** A streaming/scan error surfaced via `rows.Err()` leaves the result (and any connection resources it pins) unclosed until the surrounding `WithConnection` tears the connection down.
- **Recommendation:** Add `defer rows.Close()` right after the `olap.Query` call (and drop the manual close), or close before returning in the `rows.Err()` branch.

**Verification:** CONFIRMED — the `olap.Query` at `column_timeseries.go:161` has no `defer rows.Close()`. Every in-loop error path closes explicitly (lines 186, 195, 202, 208), but the post-loop `if err := rows.Err(); err != nil { return err }` at lines 217-219 returns before the manual `rows.Close()` on line 222, leaking the cursor on that path.

**Manual Reproduction Steps (for QA):**
1. Not directly user-reproducible — leaking the cursor requires `rows.Err()` to return a non-nil error after the scan loop completes (a streaming/scan failure surfaced only at end-of-iteration), which is hard to induce deterministically from the UI.
2. Confirm in code: `runtime/queries/column_timeseries.go:161` has no `defer rows.Close()`, and the `rows.Err()` branch at lines 217-219 returns while `rows.Close()` sits at line 222.

### queries-08 [LOW] Unchecked `rowMap["max"].(time.Time)` assertion in `ColumnTimeRange`
- **Category:** bug  **Confidence:** low  **Location:** `runtime/queries/column_time_range.go:113,166`
- **Problem:** The `min` value is type-checked with `,ok` and a clean error, but `max` is asserted as `rowMap["max"].(time.Time)` with no check. This is an internal inconsistency rather than a practically reachable crash: `min` and `max` are `min(col)`/`max(col)` over the same column, so when `min` scans as a non-nil `time.Time`, `max` does too. The assertion would only panic if a driver returned the two aggregates of one column as different Go types, which does not happen for the supported dialects.
- **Failure scenario:** Largely theoretical — a driver returning `min` and `max` of the same timestamp column with differing Go types.
- **Recommendation:** Use the comma-ok form for `max` as well for symmetry and defensiveness.

**Verification:** PLAUSIBLE — the asymmetry is real: `column_time_range.go:107-113` type-checks `min` with `,ok` and a clean error (`"not a timestamp column"`) but then asserts `summary.Max = timestamppb.New(rowMap["max"].(time.Time))` with no check (line 113); the StarRocks path repeats the same pattern at lines 160-166. Because `min` and `max` are `min(col)`/`max(col)` over the same column, they share a Go type in practice, so the unchecked assertion is essentially unreachable — a defensive-symmetry nit rather than a live crash, matching the finding's own low rating.

**Manual Reproduction Steps (for QA):**
1. Not user-reproducible — would require a driver returning `min` and `max` of the same timestamp column as different Go types, which the supported dialects do not do.
2. Confirm in code: `runtime/queries/column_time_range.go:108-113` (checked `min` vs unchecked `max` assertion) and the mirror at lines 161-166.

### queries-09 [LOW] Diagnostic histogram computes a garbage bucket count for constant columns
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/queries/column_numeric_histogram.go:289-302`
- **Problem:** `calculateDiagnosticMethod` only bails on nil min/max/range, not on `range == 0` (line 289). For a constant column (`min == max`), `rng` is a non-nil `0.0`, so it proceeds; `ticks` becomes `0`, and `NiceAndStep(min, min, 0)` returns `gap == 0` (the loop exits via the `IsNaN(step)` branch with `prestep` still 0). Then `bucketCount := int(math.Ceil((endTick-startTick)/gap))` evaluates `int(math.Ceil(0.0/0.0)) = int(NaN)` — an implementation-defined value (on amd64/arm64, `math.MinInt64`) that is interpolated into `range(0, %d)`. `calculateFDMethod` guards this case (`bucketSize == 0` → return) and the rug histogram uses a fixed 500-bucket count, but the diagnostic path does not.
- **Failure scenario:** Profiling a numeric column whose finite values are all identical yields an empty or malformed diagnostic histogram (the negative/garbage bucket count produces an empty `generate_series`) instead of a single-bucket result.
- **Recommendation:** Add an early `if *rng == 0 { return nil }` (or emit a single bucket), matching the FD path.

**Verification:** CONFIRMED — `column_numeric_histogram.go:289-290` bails only on `minVal == nil || maxVal == nil || rng == nil`, not on `*rng == 0`. For a constant column (`min == max`, `rng` a non-nil `0.0`) it proceeds: `ticks` becomes `0` (line 294-296, `*rng < 40`), then `NiceAndStep(min, min, 0)` returns `gap == 0` (in `column_numeric_histogram_nice.go`, `tickIncrement` yields NaN which hits the `math.IsNaN(step)` branch at line 47, returning `prestep == 0`). Line 299 then computes `int(math.Ceil((endTick-startTick)/gap)) = int(math.Ceil(0/0)) = int(NaN)`, an implementation-defined garbage value interpolated into the bucket range. `calculateFDMethod` guards this via `*rangeVal == 0.0` (line ~108, `return 0, nil`), and the rug path uses a fixed count, but the diagnostic path does not.

**Manual Reproduction Steps (for QA):**
1. Create a model/table with a numeric column whose non-null finite values are all identical (e.g. every row = 5).
2. Request the diagnostic numeric histogram for that column: `GET /v1/instances/{instance_id}/queries/numeric-histogram/tables/{table}?column_name=...&histogram_method=HISTOGRAM_METHOD_DIAGNOSTIC` (or open the column profile that renders the diagnostic histogram).
3. Expected: a single-bucket histogram (or empty). Actual: an empty/malformed histogram driven by a garbage bucket count from `int(NaN)` at `column_numeric_histogram.go:299`.
4. Confirm in code: missing `*rng == 0` early-return at `column_numeric_histogram.go:289`.

### queries-10 [LOW] Dead per-measure index bookkeeping in the comparison query
- **Category:** tech-debt  **Confidence:** high  **Location:** `runtime/queries/metricsview_comparison_toplist.go:453-474`
- **Problem:** `calculateMeasuresMeta` computes `baseSubqueryIndex` and `outerIndex` (with the `inner`/`outer += 4` stepping) for every measure, but nothing reads those fields — verified via grep, only `.expand` is consumed (by `validateSort` at line 527 and `validateMeasureAliases` at line 542). This is leftover from the hand-written nested-subquery SQL that was replaced by the `metricsview` executor.
- **Failure scenario:** None functional; it is misleading dead code in an area explicitly being retired, and the `+= 4` magic invites incorrect edits.
- **Recommendation:** Drop the two fields and the `inner`/`outer` counters; keep only the `expand` map (or replace `metricsViewMeasureMeta` with a `map[string]bool`).

**Verification:** CONFIRMED — `metricsview_comparison_toplist.go:453-474` computes `inner`/`outer` counters (with `outer += 4` for expanded measures) and writes them into `baseSubqueryIndex`/`outerIndex` (lines 464-465). A repo-wide grep for `baseSubqueryIndex`/`outerIndex` returns only the struct field declarations (lines 48-49) and these two write sites — no reader anywhere. Only `.expand` is consumed downstream. This is dead bookkeeping left over from the retired hand-written nested-subquery SQL.

**Manual Reproduction Steps (for QA):**
1. Not user-reproducible — no functional effect (dead code).
2. Confirm in code: `grep -rn "baseSubqueryIndex\|outerIndex" runtime/` returns only `runtime/queries/metricsview_comparison_toplist.go:48-49` (declarations) and `:464-465` (writes); no reads.

### queries-11 [LOW] Aggregation export always issues an extra time-range resolution query
- **Category:** performance  **Confidence:** medium  **Location:** `runtime/queries/metricsview_aggregation.go:165-175`
- **Problem:** `Export` calls `ResolveTimestampResult` (a full `metrics_time_range` resolver round-trip) plus `BindQuery` whenever `mv.ValidSpec.TimeDimension != ""`, whereas `Resolve` was tightened to only do so when `len(mv.ValidSpec.Rollups) > 0` and the query targets the primary time dimension (lines 108–116). Because `Resolve` runs correctly without binding for non-rollup views, the executor already resolves time ranges lazily; the binding is only needed for rollup routing, so the export's unconditional bind is redundant for every non-rollup view (and is issued even when the query has no time range at all).
- **Failure scenario:** Large/slow OLAP sources incur an unnecessary min/max/watermark scan on each aggregation export; also couples export to rollup plumbing it does not need.
- **Recommendation:** Apply the same `len(mv.ValidSpec.Rollups) > 0` (and primary-time-dimension) guard used in `Resolve` before binding timestamps in `Export`. (Fixing queries-01 by computing `qryTimeDim` defensively is a prerequisite, since both touch the same block.)

**Verification:** CONFIRMED — `Export` (`metricsview_aggregation.go:165-175`) calls `ResolveTimestampResult` + `e.BindQuery` whenever `mv.ValidSpec.TimeDimension != ""`, with no rollup guard. `Resolve` (lines 108-116) gates the identical binding on `len(mv.ValidSpec.Rollups) > 0 && mv.ValidSpec.TimeDimension != "" && (qryTimeDim == "" || qryTimeDim == mv.ValidSpec.TimeDimension)`, confirming the binding is only needed for rollup routing. `ResolveTimestampResult` runs the `metrics_time_range` resolver (lines 230-246), a full min/max/watermark round-trip. So every aggregation export on a time-dimensioned but non-rollup view issues a redundant scan.

**Manual Reproduction Steps (for QA):**
1. On a metrics view that has a time dimension but no rollups defined, export an aggregation (CSV/XLSX/Parquet).
2. Observe the OLAP query log: an extra `metrics_time_range` (min/max/watermark) query is issued before the export query, even though the export does not need rollup routing. The equivalent non-export `Resolve` path does not issue it.
3. Confirm in code: `runtime/queries/metricsview_aggregation.go:165` (unconditional bind in `Export`) versus lines 108-116 (rollup-guarded bind in `Resolve`).

### queries-12 [LOW] Null-filled timeseries is silently truncated at 5000 buckets
- **Category:** product-improvement  **Confidence:** high  **Location:** `runtime/queries/metricsview_timeseries.go:332,356`
- **Problem:** Both `addNulls` and `addNullsNew` stop after 5000 iterations via a hard `break` with no signal to the caller. A wide time range at a fine grain (e.g., minutes over several months) silently drops the tail of the series instead of returning an error or a downsample hint.
- **Failure scenario:** A user charts minute-granularity data over a multi-month range and sees the line abruptly end partway through, with no indication the data was truncated.
- **Recommendation:** Compute the expected bucket count up front and either reject the request with an actionable error ("range too large for this grain") or surface a truncation flag in the response so the UI can warn.

**Verification:** CONFIRMED — `addNulls` (`metricsview_timeseries.go:330-343`) and `addNullsNew` (lines 353-374) both stop after `i > 5000` via a hard `break // safety break` (lines 332-334 and 356-358) with no signal to the caller. A wide range at fine grain silently drops the tail of the null-filled series.

**Manual Reproduction Steps (for QA):**
1. Chart a measure over a metrics-view timeseries at minute granularity across a multi-month range with null filling enabled (a range that spans well over 5000 minute buckets).
2. Expected: either the full series, a downsample, or an explicit "range too large" error/warning. Actual: the line/series abruptly ends after ~5000 buckets with no indication the tail was truncated.
3. Confirm in code: the `if i > 5000 { break }` guards at `metricsview_timeseries.go:332` and `:356`.

## Product Improvements
- **queries-12** (above): silent 5000-bucket truncation of null-filled series should be an explicit error or a UI-visible truncation signal.
- **Consistent nil/edge handling:** queries-01, -03, -06, -08 all stem from the same anti-pattern — dereferencing OLAP/driver outputs (or request pointers) that can be nil, and `panic`-ing on unexpected shapes. A small shared helper set (safe timestamp scan, "all-or-nothing" numeric-stats decode) would remove a whole class of user-visible 500s/crashes.
- **Grain validation at the edge:** the metrics-view timeseries handler accepts any `TimeGrain` enum without validation, pushing an unsupported millisecond grain deep into null-fill/duration code (queries-02). Validating supported grains in the server handler yields a clear 4xx error instead of a panic.

## New Feature Ideas
- **Streaming / chunked in-memory query results (M):** Non-export `Resolve` paths (`rowsToData`) buffer the entire result set into `[]*structpb.Struct` before returning. A capped/streamed decode with an explicit "result too large" error would harden toplist/rows/aggregation against pathological requests.
- **Export truncation metadata (S):** Return row-count and "truncated" flags from `Export`/`Resolve` so the UI can tell users when a CSV/XLSX hit the export limit or a chart hit the 5000-bucket cap.
- **Golden SQL snapshot tests per dialect (M):** The profiling SQL builders (histograms, time-grain, rug, timeseries) branch heavily on DuckDB/ClickHouse/StarRocks with subtle differences (reserved words, `range` vs `generate_series`, cast syntax). Snapshot the generated SQL per dialect to catch drift as dialects are added.
- **Deprecate-and-delete tracker for legacy queries (M–L):** Several files are now thin executor adapters or contain dead code (queries-10). A tracked plan that routes the remaining call sites through resolvers and deletes the adapters would shrink this package materially.
- **Aggregation-function allowlist for profiling (S–M):** Replace free-form `ColumnTopK.Agg` / measure `Expression` with a typed {function, column} model for the common cases, reserving raw SQL for an explicit, separately-gated path (see queries-04).

## Notes & Missing Test Coverage
- **queries-01** (aggregation export with nil `TimeRange`) has no test; `metricsview_aggregation_test.go` is large but the export-without-time-range case is not exercised. A regression test that exports an aggregation on a time-dimensioned metrics view with `TimeRange == nil` (and separately with `IncludeHeader` on a non-time-dimensioned view) would lock in the fix for both deref sites.
- **queries-02/03** are dialect/edge-data specific (millisecond grain; single-row StarRocks column). Neither is covered; the StarRocks stddev case in particular needs an integration test since it depends on MySQL NULL semantics.
- **queries-05** (Parquet timeseries export) is not covered; existing timeseries export tests appear to use CSV.
- SQL construction outside the profiling expression fields (queries-04) is consistently identifier/table-escaped or parameterized; I found no additional injection via table/column/timezone/limit inputs. `resource_watermark.go` expressions are author-controlled (metrics-view spec); the column-timeseries measure expressions are `ReadProfiling`-gated.

## Verification Notes
All twelve prior findings were independently re-traced against the code; none were entirely removed (removed_count = 0), but two carried false sub-claims that were corrected, and several severities/confidences/locations were adjusted:

- **queries-02:** The `TIME_GRAIN_UNSPECIFIED` half of the claim is a false positive and was removed. The executor rejects an unspecified-grain time floor at `metricsview/ast.go:383` before returning any rows, so `addTo` is never reached for that grain. The `TIME_GRAIN_MILLISECOND` half is confirmed and fully traced (no handler validation → `TimeGrain.Valid()` true → DuckDB `DateTruncExpr` emits valid SQL → `InfDuration` → panic), so the finding is retained at high confidence, narrowed to millisecond. Corrected the second location line (`377`, was `376`).
- **queries-03:** The ClickHouse half is a false positive and was removed — ClickHouse `stddevSamp` returns `nan` (non-null), not `NULL`, for n=1, so no panic there (same as DuckDB's hardcoded `NaN`). Narrowed to StarRocks (MySQL `NULL` semantics), which is confirmed. Kept medium/medium.
- **queries-05:** Confirmed, but the failure mechanism in the original write-up ("unequal column lengths at `NewRecordBatch`") was corrected: the empty `Type` maps to `CODE_UNSPECIFIED`, leaving a nil `arrow.DataType`, which panics earlier at `array.NewRecordBuilder` (metricsview.go:355). Conclusion (broken parquet export) unchanged; kept medium.
- **queries-08:** Downgraded confidence high → low. The unchecked `max` assertion is a real inconsistency but is essentially unreachable (min/max of the same column always share a Go type), so it is now framed as a defensive-symmetry nit.
- **queries-06:** Confidence high → medium, with an explicit note that reachability is low (the column is produced by `date_trunc`/`time_bucket` and normally scans as `time.Time`).
- **queries-01, -04, -07, -09, -10, -11, -12:** Confirmed as written; locations/line references verified and lightly tightened. queries-01's second deref site (`:523`) was re-traced through `RewriteQueryTimeRanges` (`executor_rewrite_time.go:16`), which confirms `qry.TimeRange` stays nil for a non-time-dimensioned view. queries-10's "no reader" claim was confirmed by grep (`baseSubqueryIndex`/`outerIndex` are write-only). queries-11 was confirmed by contrasting `Resolve` (binds only with rollups) against `Export` (binds unconditionally).
