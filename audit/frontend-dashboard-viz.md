# Dashboard Visualizations — Code Review
**Scope:** `web-common/src/features/dashboards/{pivot,time-series,leaderboard,dimension-table,filters,big-number,dimension-search}`  |  **Health:** fair

## Overview
This area renders every interactive surface of an explore/canvas dashboard: the pivot table (by far the largest and most intricate subsystem, with a multi-stage query pipeline, infinite scroll, click-to-filter selection, and column virtualization), the time-series measure charts, leaderboards, the dimension detail table, filter pills, and big numbers. The code is generally mature, thoughtfully factored (e.g. the staged `derivedSwitch` pipeline in `pivot-data-store.ts`, the shared `buildPivotFilter` core), and reasonably well tested at the util level. Most defects found are localized correctness/robustness edge cases rather than structural problems: comparison rendering when a comparison value is exactly `0`, export order not matching on-screen sort, a missing observer teardown, and several `Math.max(...[])`/divide-by-zero edges that are mostly (but not always) guarded downstream. The biggest risks are subtle user-visible correctness gaps in the comparison/export paths and a small but real memory leak in the leaderboard. No security or data-loss issues were found in scope.

## Findings

### frontend-dashboard-viz-01 [MEDIUM] Big number hides all comparison deltas when the comparison-period value is exactly 0
- **Category:** correctness  **Confidence:** high  **Location:** `web-common/src/features/dashboards/big-number/MeasureBigNumber.svelte:254` (and `:150`)
- **Problem:** The entire comparison block is gated on `{#if showComparison && comparisonValue}`. Because `comparisonValue` is a raw number, a legitimate comparison value of `0` is falsy and suppresses the whole block — both the absolute delta and the percentage. `diff = value - comparisonValue` is computed correctly (`value - 0`) but never rendered. Relatedly, `noChange = !comparisonValue` makes the `{:else}` "no change" branch (line 278) unreachable inside that block (comparisonValue is always truthy there), so it is dead code.
- **Failure scenario:** A measure whose comparison period summed to exactly `0` (e.g. a net metric, or an all-zero prior period) and whose current value is `1000` shows the big number with no delta at all, implying there is no comparison data when in fact the change is +1000.
- **Recommendation:** Gate on data presence, not truthiness: use `comparisonValue != null` (and `showComparison`) to decide whether to render the absolute delta. Keep the percentage suppressed only when the denominator is `0` (division undefined), and show the absolute change regardless. Remove or fix the dead "no change" branch.

**Verification:** CONFIRMED — `MeasureBigNumber.svelte:254` gates the whole comparison block on `{#if showComparison && comparisonValue}`, so `comparisonValue === 0` (falsy) suppresses it; `noChange = !comparisonValue` at line 150 makes the `{:else}` "no change" branch at lines 278-281 unreachable inside that block (the block only renders when `comparisonValue` is truthy, so `noChange` is always false there).

**Manual Reproduction Steps (for QA):**
1. Create/open an explore dashboard with a measure that can legitimately sum to 0 (e.g. `SUM` over a signed column, or use a dataset where the prior period has no rows so the sum is 0 but not null).
2. Pick a time range whose comparison period totals exactly 0 for that measure, and turn on time comparison via the comparison toggle in the time controls.
3. Look at the big number for that measure. Expected: an absolute delta equal to the current value (e.g. "+1,000") below the big number. Actual: no delta or percentage is rendered at all, as if comparison data were missing.
4. Cross-check: pick a period where the comparison value is nonzero and confirm the delta/percentage appear, proving comparison is otherwise working.

### frontend-dashboard-viz-02 [MEDIUM] Dimension-table export always sorts by a measure, ignoring "sort by dimension"
- **Category:** correctness  **Confidence:** high  **Location:** `web-common/src/features/dashboards/dimension-table/dimension-table-export.ts:103-107,143-148`
- **Problem:** The export request always sets `sort: [{ name: apiSortName, ... }]` where `apiSortName = exploreState.leaderboardSortByMeasureName` (optionally with a delta suffix). The `switch` only handles `DELTA_ABSOLUTE`/`DELTA_PERCENT`; it never inspects the `DIMENSION` sort type. The on-screen table (`DimensionDisplay.svelte:156` via `getSort`) sorts by the dimension name when `sortType === SortType.DIMENSION`. `SortType` is a direct re-export of `DashboardState_LeaderboardSortType` (`proto-state/derived-types.ts`), so the two enums are numerically identical and the export genuinely omits the dimension-sort case — the exported CSV/row order diverges from what the user sees.
- **Failure scenario:** User sorts a dimension table alphabetically by dimension value and clicks Export; the downloaded file is sorted by the measure instead, so rows are in a different order than the screen.
- **Recommendation:** Mirror `getSort`: when `dashboardSortType === DIMENSION`, sort by `selectedDimensionName`; otherwise apply the measure/delta sort as today. Reuse the leaderboard `getSort` helper so the two code paths cannot drift.

**Verification:** CONFIRMED — `dimension-table-export.ts:103` sets `apiSortName = exploreState.leaderboardSortByMeasureName`, the `switch` at lines 115-122 only handles `DELTA_ABSOLUTE`/`DELTA_PERCENT`, and the request at lines 143-148 always sorts by `apiSortName`; the on-screen table uses `getSort` (`DimensionDisplay.svelte:156`, `leaderboard-utils.ts:142-160`), which sorts by `dimensionName` when `type === SortType.DIMENSION`; `SortType` is re-exported from `DashboardState_LeaderboardSortType` in `proto-state/derived-types.ts:12`.

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard and click a dimension name in a leaderboard header (or "Expand Table") to open the dimension detail table.
2. Click the dimension column header so the table sorts alphabetically by dimension value (the sort arrow moves to the dimension column).
3. Note the on-screen row order, then open the export menu and download a CSV.
4. Open the downloaded file. Expected: rows in the same alphabetical dimension order as the screen. Actual: rows are sorted by the measure (highest/lowest values first).

### frontend-dashboard-viz-03 [MEDIUM] Pivot export does not reflect measure-column sort order
- **Category:** product-improvement  **Confidence:** high  **Location:** `web-common/src/features/dashboards/pivot/pivot-export.ts:155-181`
- **Problem:** For non-flat pivots the export sort is derived only from row dimensions (`sort = rowDimensions.map(...)`), taking `desc` from any `pivotState.sorting` entry whose `id` matches a row-dimension name. When the user has sorted the on-screen pivot by a measure column, `pivotState.sorting[0].id` is a measure name or a minimized `m{n}` cell accessor, not a row-dimension name, so the `find` misses, `desc` defaults to `false`, and the sort silently falls back to dimension order. The interactive query handles exactly this case via `getSortForAccessor` (`pivot-utils.ts:466` for the row-totals measure name, `:486` for the `m{n}` accessor), which the export path does not call.
- **Failure scenario:** User sorts a pivot descending by a revenue measure column and exports; the exported rows come back in dimension order, not revenue order.
- **Recommendation:** Detect when the active pivot sort targets a measure accessor and translate it into an aggregation `sort` on that measure (reusing `getSortForAccessor`), or at minimum document that pivot exports are dimension-ordered.

**Verification:** CONFIRMED — `pivot-export.ts:173-181`: for non-flat pivots, `sort = rowDimensions.map(...)` with `desc` taken from `pivotState.sorting.find((s) => s.id === d.name)` — a measure-column sort id (measure name or `m{n}` accessor) never matches a row-dimension name, so the measure sort is silently dropped; the interactive path handles it via `getSortForAccessor` (`pivot-utils.ts`).

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard and switch to the Pivot tab.
2. Add at least one row dimension and one measure as a column (keep the default nested layout, not flat).
3. Click the measure column header to sort descending by that measure; verify the on-screen rows reorder by measure value.
4. Export the pivot via the export menu and open the downloaded file. Expected: rows ordered by the measure (matching the screen). Actual: rows come back in dimension order (ascending by the row dimension).

### frontend-dashboard-viz-04 [LOW] Leaderboard `IntersectionObserver` is never disconnected on destroy
- **Category:** concurrency  **Confidence:** high  **Location:** `web-common/src/features/dashboards/leaderboard/Leaderboard.svelte:93-110`
- **Problem:** `onMount` creates an `IntersectionObserver` and only calls `unobserve` once the element becomes intersecting. The `onMount` callback returns nothing, and there is no `onDestroy`, so a leaderboard destroyed while still off-screen leaves its observer (with an active observation of `container`, plus the callback closure) registered and non-collectable. The codebase already has the correct pattern in `time-series/measure-chart/interactions.ts` (`createVisibilityObserver`, which returns `() => observer.disconnect()` at line 28 and is torn down in `MeasureChart.svelte`'s `onDestroy` at line 81).
- **Failure scenario:** A dashboard with many dimensions where several leaderboards below the fold are never scrolled into view; navigating away or swapping the metrics view unmounts them while their observers remain registered, retaining DOM nodes.
- **Recommendation:** Reuse `createVisibilityObserver`, or return the disconnect from `onMount` / add `onDestroy(() => observer.disconnect())`.

**Verification:** CONFIRMED — `Leaderboard.svelte:93-110`: `onMount` creates the `IntersectionObserver` and only calls `observer.unobserve(container)` inside the intersection callback; the `onMount` callback returns nothing and there is no `onDestroy`, so a leaderboard unmounted before scrolling into view never disconnects its observer.

**Manual Reproduction Steps (for QA):**
Not user-reproducible through normal UI observation (no visual symptom) — confirm in code: `web-common/src/features/dashboards/leaderboard/Leaderboard.svelte:93-110`, note the missing cleanup return / `onDestroy`. Optional heap check for QA with devtools:
1. Open an explore dashboard with many dimensions so several leaderboards are below the fold; do not scroll.
2. In Chrome DevTools > Memory, take a heap snapshot.
3. Navigate away from the dashboard (e.g. to another dashboard) without ever scrolling, then take another snapshot and search for detached `IntersectionObserver`/detached leaderboard DOM nodes. Expected: none retained. Actual: observers for the never-visible leaderboards remain registered and retain their container nodes.

### frontend-dashboard-viz-05 [LOW] `getLeaderboardMaxValues` drops legitimate `0` values and can yield `-Infinity`
- **Category:** correctness  **Confidence:** high  **Location:** `web-common/src/features/dashboards/leaderboard/leaderboard-utils.ts:350-351`
- **Problem:** `.filter(Boolean)` removes any `0` from the numeric list (after `Math.abs`), and `Math.max(...numericValues)` returns `-Infinity` when the list is empty (measure entirely null/zero). Dropping `0` is harmless for a max of absolute values except in the all-zero case, but the empty-list `-Infinity` is stored in `maxValues[measureName]`. `LeaderboardRow` guards with `maxValue <= 0 → 0` (`LeaderboardRow.svelte:117`), so the current UI is safe, but the stored `-Infinity` is a latent trap for any future consumer and makes the "max" incorrect for all-zero measures.
- **Failure scenario:** A measure whose values are all `0`/null across the leaderboard yields `maxValues[name] === -Infinity`; only the downstream `<= 0` guard prevents `NaN` bar widths.
- **Recommendation:** Filter with `Number.isFinite` (not `Boolean`), and fall back to `0` when the list is empty: `numericValues.length ? Math.max(...numericValues) : 0`.

**Verification:** PLAUSIBLE — `leaderboard-utils.ts:343-351`: `.filter(Boolean)` drops `0` (a valid `Math.abs` result) and `Math.max(...[])` returns `-Infinity` for an all-zero/all-null measure; latent because the only consumer clamps it (`LeaderboardRow.svelte` treats `maxValue <= 0` as `0`), so no user-visible symptom today.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — the downstream `maxValue <= 0` guard hides it. Confirm in code: `web-common/src/features/dashboards/leaderboard/leaderboard-utils.ts:350-351` (`.filter(Boolean)` followed by `Math.max(...numericValues)`); optionally set a breakpoint or log `maxValues` in `Leaderboard.svelte` (`getLeaderboardMaxValues` call) on a dashboard whose measure is all zeros and observe the stored value is `-Infinity` instead of `0`.

### frontend-dashboard-viz-06 [LOW] Leaderboard percent-of-total divides by zero → `Infinity`
- **Category:** correctness  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/leaderboard/leaderboard-utils.ts:215-216` (same pattern at `:105-106`)
- **Problem:** `pctOfTotals[measureName] = total !== null && value !== null ? value / total : null`. When `total === 0` (a valid summable total of zero), this evaluates to `±Infinity` (or `NaN` if the value is also 0), which then flows into `formatMeasurePercentageDifference`/`PercentageChange` for display (`LeaderboardRow.svelte:277,283`). Reachability requires a `valid_percent_of_total` measure whose grand total across all values is exactly `0` (e.g. a signed measure whose positives and negatives cancel), so it is an edge but a real one.
- **Failure scenario:** A summable measure whose period total is `0` renders per-row percent-of-total as `Infinity%`.
- **Recommendation:** Treat `total === 0` like a missing total: `total && value !== null ? value / total : null`.

**Verification:** CONFIRMED — `leaderboard-utils.ts:215-216` (and the same pattern at `:105-106`): `total !== null && measureValue !== null ? measureValue / total : null` yields `Infinity`/`NaN` when `total === 0`; the result flows into the percent-of-total context column rendering in `LeaderboardRow.svelte`. Data-conditional (requires a `valid_percent_of_total` measure whose grand total is exactly 0) but directly reachable through the UI with such data.

**Manual Reproduction Steps (for QA):**
1. Create a metrics view with a measure defined as a signed sum (e.g. `SUM(net_amount)`) with `valid_percent_of_total: true`, over data where positives and negatives cancel to exactly 0 for some period.
2. Open the explore dashboard, select a time range where that measure's grand total is 0.
3. In the leaderboard, ensure the percent-of-total context column is shown for that measure (sort by it, or enable context columns for all measures).
4. Look at the percent-of-total cell for any row with a nonzero value. Expected: blank/null. Actual: `Infinity%` (or `NaN` for zero-value rows).

### frontend-dashboard-viz-07 [LOW] `pivot-column-definition` mutates a shared `colDimensions` array via `.reverse()` twice
- **Category:** tech-debt  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/pivot/pivot-column-definition.ts:540` and `:144`
- **Problem:** `getNestedColumnDef` calls `colDimensions.reverse()` (in place) at line 540 to build the row-header column group, then passes the *same* array to `createColumnDefinitionForDimensions`, which calls `colDimensions.reverse()` again at line 144 to build the row-totals group. Within the function the two groups therefore end up nested in opposite dimension orders (the double-reverse also happens to restore the caller's array to its original order, so no external corruption). It is currently invisible only because the totals-group headers are empty strings (line 149: `header: ""`); any change to those headers, or a guard-condition mismatch between the two blocks, would expose inconsistent nesting.
- **Failure scenario:** With a 2-level column dimension `[A, B]`, the row-header group nests `A > B` while the row-totals group nests `B > A`; a future non-empty totals header would render the reversed hierarchy.
- **Recommendation:** Do not mutate the shared array. Use `[...colDimensions].reverse()` (or `toReversed()`) locally in each function.

**Verification:** PLAUSIBLE — `pivot-column-definition.ts:540` (`getNestedColumnDef`) reverses `colDimensions` in place, then line 607-609 passes the same array to `createColumnDefinitionForDimensions`, which reverses it again at line 144 for the row-totals group; the two groups nest in opposite orders, currently masked because the totals-group headers are empty strings (`header: ""` at line 149). No user-visible symptom today.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — the reversed nesting is hidden by empty totals-group headers. Confirm in code: `web-common/src/features/dashboards/pivot/pivot-column-definition.ts:540` and `:144` (both call `colDimensions.reverse()` on the same array instance; compare the resulting group nesting order in a debugger with a 2-level column-dimension pivot).

### frontend-dashboard-viz-08 [LOW] `transformAggregateDimensionData` builds `emptyData` with one shared array reference
- **Category:** correctness  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/time-series/utils.ts:145-147`
- **Problem:** `new Array(dimensionValues.length).fill([])` fills every slot with the *same* `[]` reference, and it is returned directly on the `!headers.length` early-out (line 152). The main path (lines 165-170) correctly builds distinct arrays, so this early-out is inconsistent. The consumer (`prepareTimeSeries`, `utils.ts:61`) only maps/reads and does not mutate the array (verified), so the aliasing is latent, but any per-row mutation would corrupt all rows at once.
- **Failure scenario:** A refactor that mutates `emptyData[i]` (e.g. to push a null-filled point) would mutate every dimension row simultaneously.
- **Recommendation:** Build distinct arrays: `Array.from({ length: dimensionValues.length }, () => [])`.

**Verification:** PLAUSIBLE — `time-series/utils.ts:145-147`: `new Array(dimensionValues.length).fill([])` shares one `[]` across all slots, and it is returned as-is on the `!headers.length` early-out at line 152; the main path (lines 165-170) builds distinct deep-copied rows. Latent: current consumers only read.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — no current code mutates the shared arrays. Confirm in code: `web-common/src/features/dashboards/time-series/utils.ts:145-152` (`.fill([])` plus the early return); note the contrast with the per-row copies at lines 165-170.

### frontend-dashboard-viz-09 [LOW] `estimateColumnCharacterWidths` has a dead `sort()` and `-Infinity` on empty columns
- **Category:** readability  **Confidence:** high  **Location:** `web-common/src/features/dashboards/dimension-table/dimension-table-utils.ts:156-157`
- **Problem:** `values` is an array of character-length *numbers*; `values.sort()` sorts them lexicographically ("10" < "2") and the result is discarded (the next line takes `Math.max(...values)`, which ignores order), so the sort is pure dead code. `Math.max(...values)` also yields `-Infinity` when a column has only null rows. The `-Infinity` is neutralized downstream in `estimateColumnSizes` (the `Math.max(..., config.minColumnWidth)` guard, and it is only consulted for column index 0), so it is not a live bug, but the code is misleading.
- **Failure scenario:** None user-visible today; the misleading `sort()` invites incorrect edits.
- **Recommendation:** Delete the `sort()`; guard the empty case (`values.length ? Math.max(...values) : 0`).

**Verification:** CONFIRMED — `dimension-table-utils.ts:156-157`: `values.sort()` (lexicographic on numbers, result discarded) immediately followed by `Math.max(...values)`, which is `-Infinity` for an all-null column; `-Infinity` is neutralized downstream in `estimateColumnSizes` by the `Math.max(..., config.minColumnWidth)` clamp. Dead code confirmed; no user-visible impact.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — purely internal dead code with downstream clamping. Confirm in code: `web-common/src/features/dashboards/dimension-table/dimension-table-utils.ts:156-157`.

### frontend-dashboard-viz-11 [LOW] `mergeFilters` can push a duplicate expression when `filter2.cond.exprs` is falsy
- **Category:** correctness  **Confidence:** low  **Location:** `web-common/src/features/dashboards/pivot/pivot-merge-filters.ts:93-98` (same pattern at `:100-105`, `:107-114`)
- **Problem:** `if (!filter2.cond?.exprs) { filter2.cond!.exprs = [copy(ie)] }` then unconditionally `filter2.cond?.exprs?.push(copy(ie))` — when `exprs` was falsy this both seeds the array *and* pushes a second copy, duplicating the expression. Not currently reachable because `filter2` is `copyFilterExpression(filter2) ?? createAndExpression([])` (line 57) and `wrapNonJoinerExpression` always produces a defined `exprs` array, so the seed branch never runs. It remains a latent double-insert.
- **Failure scenario:** Any future caller passing an AND/OR expression with `exprs: undefined` would get every carried-over IN/NIN/other expression duplicated.
- **Recommendation:** Make the seed-vs-push mutually exclusive (`else`), or initialize `exprs` once at the top and always push.

**Verification:** PLAUSIBLE — `pivot-merge-filters.ts:93-98` (and the same pattern at `:100-105`, `:107-114`): `if (!filter2.cond?.exprs) { filter2.cond!.exprs = [copyFilterExpression(ie)] }` is followed by an unconditional `push(copyFilterExpression(ie))`, so a falsy `exprs` would get the expression twice. Currently unreachable: `filter2 = copyFilterExpression(filter2) ?? createAndExpression([])` at line 57 guarantees a defined `exprs` array.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — the seed branch is dead with current callers. Confirm in code: `web-common/src/features/dashboards/pivot/pivot-merge-filters.ts:93-114` (seed-then-push pattern) and line 57 (the guard that keeps it unreachable).

### frontend-dashboard-viz-12 [LOW] `useDimensionSearchResults` computes `NaN` progress with zero dimensions
- **Category:** correctness  **Confidence:** low  **Location:** `web-common/src/features/dashboards/dimension-search/useDimensionSearchResults.ts:80-81`
- **Problem:** `progress = Math.round((completedCount * 100) / dimensions.length)` divides by `dimensions.length`; when a metrics view has no dimensions this is `0/0 → NaN`, and `completed = completedCount === dimensions.length` is `0 === 0 → true`. **Correction to the prior claim:** the `NaN` does *not* surface in the UI. With zero dimensions the derived store runs once with `completed: true`, which routes `GlobalDimensionSearchResults.svelte` to the `{:else if $results.completed && responses.length === 0}` "no results" branch (line 77); and even on the progress branch, `{#if $results.progress < 100}` is `NaN < 100 → false`, so the bar is never shown. This is therefore a defensive/robustness nit in the returned object rather than a user-visible bug.
- **Failure scenario:** None user-visible; `progress: NaN` exists transiently in the derived value for a dimensionless metrics view but is never rendered.
- **Recommendation:** Short-circuit when `dimensions.length === 0` (return `completed: true, progress: 100`), or guard the division.

**Verification:** PLAUSIBLE — `useDimensionSearchResults.ts:80-81`: `progress = Math.round((completedCount * 100) / dimensions.length)` is `NaN` for zero dimensions while `completed` is correctly `true`; as the finding itself notes, the `NaN` never renders (the "no results" branch wins and `NaN < 100` is false), so this is a robustness nit, not a user-visible bug.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — the `NaN` never reaches the UI. Confirm in code: `web-common/src/features/dashboards/dimension-search/useDimensionSearchResults.ts:80-81` (division by `dimensions.length` with no zero guard); optionally log the derived store value on a project whose metrics view defines measures but no dimensions and observe `progress: NaN`.

### frontend-dashboard-viz-13 [LOW] Leaderboard `columnCount` miscounts context columns (loading skeleton only)
- **Category:** correctness  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/leaderboard/Leaderboard.svelte:295-302`
- **Problem:** `columnCount` only adds the percent-of-total term inside the `isTimeComparisonActive ? ... : 0` branch, but the percent-of-total column renders whenever `isValidPercentOfTotal(measureName) && shouldShowContextColumns(measureName)` regardless of time comparison (`Leaderboard.svelte:333`, `LeaderboardRow.svelte:272`) — so with no comparison it under-counts by one per percent-of-total measure. When comparing, it multiplies the context-column term by `leaderboardMeasureNames.length` and uses the *sort* measure's `isValidPercentOfTotal`, even though only the sort measure shows context columns when `leaderboardShowContextForAllMeasures` is false — over-counting for multi-measure leaderboards. The only consumer is `columnCount + 1` passed to `DelayedLoadingRows` (`:380`), which feeds `LoadingRows columns={columnCount}`, so the impact is a transient off-by-N in the loading-skeleton width, not the rendered table.
- **Failure scenario:** A leaderboard with a `valid_percent_of_total` measure and no time comparison shows a loading skeleton with one fewer column than the eventual table.
- **Recommendation:** Derive `columnCount` from the same per-measure conditions used to render the `<col>`/cells (`isValidPercentOfTotal(measureName) && shouldShowContextColumns(measureName)` for percent-of-total, plus two delta columns per shown context measure).

**Verification:** CONFIRMED — `Leaderboard.svelte:295-302`: the percent-of-total term is only counted inside the `isTimeComparisonActive` branch and uses the sort measure's `isValidPercentOfTotal` multiplied across all measures, while the actual `<colgroup>` (lines 333-348) renders the percent-of-total column per measure whenever `isValidPercentOfTotal(measureName) && shouldShowContextColumns(measureName)`, regardless of time comparison; the only consumer is `DelayedLoadingRows` at line 380, so the impact is confined to the loading skeleton.

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard with a leaderboard measure that has `valid_percent_of_total: true`, sort by that measure so its percent-of-total context column is shown, and keep time comparison OFF.
2. Throttle the network in browser devtools (e.g. "Slow 3G") and reload the dashboard, or change the time range to force a refetch.
3. While the leaderboard shows its loading skeleton rows, count the skeleton columns, then compare with the loaded table. Expected: same column count. Actual: the skeleton has one fewer column than the final table (and with time comparison ON plus multiple measures where context columns show only for the sort measure, the skeleton has extra columns).

## Product Improvements
- **frontend-dashboard-viz-02** (dimension-table export ignores dimension sort) and **frontend-dashboard-viz-03** (pivot export ignores measure-column sort) are the export-parity product-facing items: in both cases the downloaded file is ordered differently from the on-screen table, which users cannot visually verify before download.
- **Comparison value of 0** (frontend-dashboard-viz-01) is also a product-perceived correctness gap: users lose the delta on any metric that legitimately hits 0 in the prior period.
- **Pivot column widths are not persisted.** `measureLengths`/`rowDimensionLengths` in `NestedTable.svelte` are component-local `writable` maps recomputed on every mount, so manual column resizes are lost on navigation or reload. Persisting them (per dashboard/metrics view) would make the pivot feel much more stable.

## New Feature Ideas
- **Export current selection from the pivot** (S–M): the pivot already tracks click-selected cells/columns/rows (`pivot-click-selection.ts`); let users export just the selected region instead of the whole table.
- **Persisted, resizable column widths** (M): store pivot/dimension-table column widths in explore state keyed by metrics view so resizes survive reloads and are shareable via URL.
- **Regex / "matches" dimension filter mode** (M): `DimensionFilterMode` already has Select/Contains/InList; a regex mode (backed by a `like`/regex expression) would cover power-user filtering that comma-lists cannot.
- **Cross-measure annotation hover sync** (M): when hovering an annotation or timestamp on one measure chart, highlight the same x-position across all stacked `MeasureChart`s for easier visual correlation.
- **Leaderboard "show N" control** (S): expose the above-the-fold `slice` count (currently a fixed prop of 7) as a per-dashboard setting, with inline "show more" rather than only "Expand Table".
- **Sticky selected rows in the dimension table** (M): pin currently-filtered dimension values to the top of `DimensionTable` so active selections stay visible when scrolling large result sets.

## Notes & Missing Test Coverage
- **Comparison-value-0 rendering** (finding 01) has no test; `MeasureBigNumber` has no spec covering `comparisonValue === 0` vs `null`/`undefined`. Worth a component test asserting the absolute delta renders for a zero prior period.
- **Export sort parity** (findings 02/03): `dimension-table-export.ts` and `pivot-export.ts` have no tests asserting the exported `sort` matches the on-screen sort for each `sortType`/measure-column case. These are exactly the paths users cannot visually verify before download.
- **`getLeaderboardMaxValues` / percent-of-total** (findings 05/06): no spec exercises all-zero or zero-total measures; add cases asserting finite `maxValues` and `null` (not `Infinity`) percent-of-total.
- The pivot infinite-scroll cache logic (`assembleFlatPivotRows`, `getPivotSkeletonForPage`, `getReachedEndForRowData` in `pivot-data-assembly.ts`) is well covered by `tests/pivot-data-refactor.test.ts`, but the interaction between `keepPreviousData` refetches and `cache.lastProcessedRowPage` (a same-page refetch is ignored rather than replacing cached page data) is untested and hard to verify statically; flagged as low-risk but unverified.
- `getColumnFiltersForPage` (`pivot-infinite-scroll.ts:148-165`) assumes `slicedAxesData[timeDimension]` is populated; it would throw on `.map` of `undefined` if a `colDimensionName` were ever absent from `minIndexVisible` (e.g. an empty `columnKeysForPage`) while a time column dimension is present. In the normal flow this is blocked by the empty-totals guard in `createRowAxesStage`, but column-pagination beyond the loaded set was not fully traced.

## Verification Notes
Each finding above was re-derived from the cited code; call sites, downstream guards, and reachability were checked before keeping it.

- **Removed 1 false positive.** The prior `frontend-dashboard-viz-10` ("dimension-table select-all Cmd+A hijacks text selection in the multiline search box") is not a real bug. The dimension search is rendered by `Search.svelte`, whose `multiline` prop **defaults to `false`** (`Search.svelte:13`); `SearchButton.svelte` renders `<Search bind:value {onSubmit} />` without overriding it, so the element is an `<input>`, not a `<textarea>`. The keydown handler's `e.target.tagName === "INPUT"` guard (`DimensionDisplay.svelte:234`) therefore correctly bails, and Cmd+A inside the search box is *not* intercepted. The finding's core premise (a multiline `<textarea>` search) is factually wrong, so it was dropped.
- **Downgraded `frontend-dashboard-viz-12`** from medium → low confidence and corrected its failure scenario: the `0/0 → NaN` progress is real in the returned object, but it never surfaces in the UI (a dimensionless metrics view routes to the "no results" branch, and `NaN < 100` is false so the progress bar is hidden). Retitled from "premature completion" — completion is actually correct when there are no queries to await.
- **Raised confidence on `frontend-dashboard-viz-02` and `-03`** from medium → high: both export-vs-on-screen sort divergences were fully traced. For 02, `SortType` is a verified direct re-export of `DashboardState_LeaderboardSortType`, and the export's `switch` demonstrably omits the `DIMENSION` case. For 03, the interactive path's `getSortForAccessor` (`pivot-utils.ts:466,486`) confirms measure-column sorting is a real, supported interaction that the export drops.
- **Corrected file:line anchors** for findings 05 (`:350-351`, the `.filter(Boolean)`/`Math.max`, not `:342`) and 06 (primary at `:215-216`, secondary at `:105-106`).
- **Confirmed as-written (latent, downstream-guarded):** findings 05 (`LeaderboardRow.svelte:117` `maxValue <= 0 → 0`), 07 (empty totals headers mask the reversed nesting; double-reverse restores the caller's array), 08 (`prepareTimeSeries` only reads the shared array), 09 (`estimateColumnSizes` `Math.max(..., minColumnWidth)` guard, column 0 only), and 11 (unreachable given the `createAndExpression([])`/`wrapNonJoinerExpression` guards). These are kept at LOW because they are real inconsistencies but not currently live bugs.
- **Added 1 new finding** (`frontend-dashboard-viz-13`): the leaderboard `columnCount` reactive computation does not match the per-measure conditions that actually render context columns. Verified in code; impact is confined to the `DelayedLoadingRows` skeleton, hence LOW.
- Findings 01 and 04 were confirmed at their stated severity/confidence without change.
