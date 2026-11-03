# Deep Audit: Canvas Dashboards

Scope: `web-common/src/features/canvas` in its entirety — components (KPI, charts, pivot, markdown), `layout-util.ts`, stores (`canvas-entity`, `time-state`, `filter-state`, `filter-manager`, `grid`), state-managers, filters, inspector, `selector.ts`, and `variables-store.ts`. Method: a multi-lens find phase (correctness, state, memory, performance) followed by an adversarial verify phase in which every candidate was traced end to end in code — and where feasible reproduced with executable tests — before being accepted; severities were adjusted where verification narrowed or widened the blast radius.

## Summary

| ID | Title | Severity | Category | Verdict |
|----|-------|----------|----------|---------|
| CANV-1 | `timeAndFilterStore` leaks a new inner subscription on every dependency change | high | memory-leak | CONFIRMED |
| CANV-2 | Move-with-`insertRow` flags the wrong row as touched, resetting widths of an untouched row in YAML | high | correctness | CONFIRMED |
| CANV-3 | `metricsViewName` never updates on `metrics_view` change, binding component filters/time to the wrong view | medium | correctness | CONFIRMED |
| CANV-4 | Time-range resolution requests re-issued for every metrics view on every spec emission | medium | performance | CONFIRMED |
| CANV-5 | Module-level `lastKey` shared across all pivot instances resets pagination | medium | state | CONFIRMED |
| CANV-6 | Async `TimeState.interval` has no stale-response guard; older request overwrites newer interval | medium | state | CONFIRMED |
| CANV-7 | In-list flag not removed on conversion to Select: `filter(...)` result discarded | medium | correctness | CONFIRMED |
| CANV-8 | FilterManager UI-filter stores leak inner subscriptions on each rebuild | medium | memory-leak | CONFIRMED |
| CANV-9 | FilterManager leaked subscriptions rerun `convertToUIFilters` per past mount (companion to CANV-8) | medium | memory-leak | CONFIRMED |
| CANV-10 | `clearDefaultFilters` throws when the YAML has no `defaults` key | medium | correctness | CONFIRMED |
| CANV-11 | Exponential O(2^n) metrics-view combination map rebuilt on every spec emission | medium | performance | CONFIRMED |
| CANV-12 | Unconditional `specStore.set` on every spec emission rebuilds every chart's query pipeline | medium | performance | CONFIRMED |
| CANV-13 | Markdown templated-string query store leaks inner subscriptions and fresh `timeAndFilterStore` chains | medium | memory-leak | CONFIRMED |
| CANV-14 | `Grid.updateFromCanvasRows` sets row stores unconditionally, making its own diff dead code | low | performance | CONFIRMED |

---

## CANV-1: `timeAndFilterStore` leaks a new inner subscription on every dependency change

**File:** `web-common/src/features/canvas/components/BaseCanvasComponent.ts:364` · **Severity:** high · **Verdict:** CONFIRMED

In BaseCanvasComponent's `timeAndFilterStore` getter, the outer derived over 17 stores executes `derived([mvFilters.parsed], ...).subscribe(set)` (line 364) inside its callback, discarding the unsubscriber and returning no cleanup function. Every emission of any dependency (time-range change, filter change, spec reconcile) while the store is subscribed permanently adds one subscription to the session-lived `FilterState.parsed` writable; one-shot `get(this.timeAndFilterStore)` calls in BaseChart/ComboChart/custom-chart leak one subscription each as well. Leaked closures capture that run's timeRange/comparisonTimeRange/canvasData, so on each subsequent `parsed` emission all accumulated stale closures fire in order, pushing a synchronous burst of stale time windows through `set` to downstream consumers (pivot, charts, KPIs) before the final correct value. Subscriptions are never released — not on outer-store teardown, component destroy, or unmount — so memory and per-interaction CPU grow without bound until the canvas page is left.

**Manual reproduction:**
1. Run `rill start` on a project with a canvas dashboard containing several components backed by a metrics view, and open the canvas in Chrome.
2. Interact once with a dimension filter (add any filter pill) so that `FilterState` exists for the metrics view and the leaking code path (the `mvFilters` branch) is active.
3. Open Chrome DevTools > Memory and take a heap snapshot. Filter the snapshot for closures; note the baseline count of closures retaining `ParsedFilters` / `V1TimeRange` contexts.
4. Change the global time range (or toggle time comparison / edit a filter) 30-50 times. Each change re-runs the outer derived callback for every subscribed component, adding one undisposed subscription to `FilterState.parsed` per component per change.
5. Take a second heap snapshot and compare: the closure count from step 3 grows roughly by (number of canvas components x number of interactions) and is never reclaimed, even after forcing GC.
6. To observe the stale-value burst: open DevTools > Network (filter on metrics-view query endpoints), then add or remove one dimension filter. Observe a burst of extra query requests/invalidation churn whose size grows with the number of prior time-range interactions, because each accumulated stale closure pushes an outdated time window through the store before the final correct value.
7. DevTools > Performance: record while toggling a filter early in the session and again after many interactions; scripting time per toggle grows steadily, confirming the per-interaction CPU cost increase until page reload.

**Suggested fix:** Capture the inner unsubscriber and return it from the outer derived callback (Svelte invokes the returned cleanup before each re-run and on teardown), or restructure to a single flat `derived` that includes `mvFilters.parsed`.

---

## CANV-2: Move-with-`insertRow` flags the wrong row as touched, resetting widths of an untouched row in YAML

**File:** `web-common/src/features/canvas/layout-util.ts:239` · **Severity:** high · **Verdict:** CONFIRMED

In `generateArrayRearrangeFunction`'s `move` case, `insertRow` splices a new row into `newArray` but the parallel `touchedRows` array is not spliced, so the subsequent `touchedRows[source.row] = true; touchedRows[rowIndex] = true` (lines 239-240) uses pre-splice indices against post-splice positions whenever the destination row is at or above the source row (any drag of a component into a new row above its current row). The final `cleaned.map(..., cleanedTouched[index])` then applies the touched flag to the wrong rows: the YAML row updater resets item widths (`COLUMN_COUNT / items.length`) on an untouched row, destroying user-configured widths there, while the actual source row keeps its stale widths, which no longer sum to 12; render-time `normalizeSizeArray` silently evens the source row out, so the rendered layout diverges from the persisted YAML. The move of the item itself is unaffected because `sourceRow` is captured by reference before the splice. Verification reproduced the misalignment with an executable vitest spec exercising the exact UI-dispatched operation; the only existing spec covering `insertRow` is an `add` operation, which does not hit the misalignment.

**Manual reproduction:**
1. In Rill Developer, create a canvas dashboard with two rows: row 1 containing two components (e.g. two KPI grids) and row 2 containing two components (e.g. two charts).
2. In the canvas editor, drag the column divider in row 1 so its components have unequal widths (e.g. 3 and 9), and do the same in row 2 (e.g. 4 and 8). Open the YAML and confirm the widths 3/9 and 4/8 are persisted.
3. Drag the first component of row 2 and drop it on the horizontal drop-zone divider at the very top of the canvas (above row 1), creating a new row.
4. Inspect the YAML: the previously-untouched row 1 has had its item widths reset to 6 and 6 (user-set 3/9 lost), while the old row 2's remaining component still has width 8, so that row's widths sum to 8 instead of 12.
5. Observe that the canvas still renders the old row 2's remaining component at full width (`normalizeSizeArray` evens out any row not summing to 12), so the rendered layout no longer matches the persisted YAML; committing the file records the corrupted widths.

**Suggested fix:** Splice `touchedRows` in lockstep with `newArray` whenever `insertRow` inserts a row (`touchedRows.splice(rowIndex, 0, false)`) before assigning the source/destination flags, using the post-splice source index (`source.row + 1` when `source.row >= rowIndex`).

---

## CANV-3: `metricsViewName` never updates on `metrics_view` change, binding component filters/time to the wrong view

**File:** `web-common/src/features/canvas/components/BaseCanvasComponent.ts:97` · **Severity:** medium · **Verdict:** CONFIRMED

`BaseCanvasComponent.metricsViewName` is assigned only in the constructor (line 97) and never refreshed in `update()` (lines 172-181), while `CanvasEntity.processRows` (canvas-entity.ts:756) reuses the existing instance whenever the renderer type is unchanged (`areSameType` compares types only). When a user switches a component's metrics view in the inspector (`MetricSelectorDropdown` → `updateProperty("metrics_view", ...)`), the instance keeps the old name, so: (1) `timeAndFilterStore` fetches `hasTimeSeries` and the global filter state (`mvFilters`) for the OLD metrics view (lines 227-229) while validating them against the NEW view's dimensions/measures (lines 286-289), so the `where` clause sent with the component's queries (consumed by KPIProvider, charts, pivot, etc.) reflects the old view's filter map; (2) `localFilters` (FilterState) and `localTimeControls` (TimeState) remain keyed to the old view, so subsequently added widget-level filters resolve against the old view's dimension/measure maps and `hasTimeSeries` checks the wrong view; (3) `inputParams()` implementations that read `this.metricsViewName` (e.g. kpi-grid) show or hide time-series options incorrectly. Concrete failure: switch a KPI from mv_a to mv_b, then add a global filter on a dimension that exists only in mv_b — the KPI silently ignores it until a full page reload. Scope-limiting factors: filters on dimensions shared by both views coincidentally work (they land in both FilterStates), and the stale binding is confined to the live editing session (a reload reconstructs components correctly), so deployed-dashboard viewers are not affected.

**Manual reproduction:**
1. In Rill Developer, create a project with two metrics views, mv_a and mv_b, where mv_b has a dimension (e.g. `country`) that does NOT exist in mv_a.
2. Create a canvas dashboard with two components: a KPI (or chart) bound to mv_a, and another component bound to mv_b.
3. Open the canvas, select the mv_a component, and in the inspector sidebar switch its "Metrics view" dropdown from mv_a to mv_b. The component re-renders with mv_b measures (spec-driven parts update, masking the bug).
4. In the canvas filter bar, add a global dimension filter on `country` (the dimension that exists only in mv_b) and select a value that excludes most rows.
5. Observe: the other mv_b component's numbers change to reflect the filter, but the switched component's numbers do NOT change. In DevTools Network tab, its metrics-view aggregation request carries no `where` condition for `country`.
6. Reload the page: the switched component now respects the filter and its numbers change — confirming the stale binding existed only for the pre-reload session.
7. Alternate observation (hasTimeSeries path): make mv_a have no time dimension and mv_b have one, then switch a KPI from mv_a to mv_b; the Sparkline / Time range display options in the inspector remain hidden (stale hasTimeSeries) until reload.

**Suggested fix:** Recompute `this.metricsViewName` (and recreate/re-key `localFilters` and `localTimeControls`) in `update()` and in the spec-store subscription when `spec.metrics_view` differs from the cached name, or read the metrics view name from the current spec inside `timeAndFilterStore` instead of caching it.

---

## CANV-4: Time-range resolution requests re-issued for every metrics view on every spec emission

**File:** `web-common/src/features/canvas/stores/time-state.ts:158` · **Severity:** medium · **Verdict:** CONFIRMED

The canvas global (and per-component local) TimeState `interval` derived store re-issues `queryServiceMetricsViewTimeRanges` network requests for every referenced metrics view whenever the canvas spec store emits, even when nothing relevant changed. Both of its upstream stores are unconditionally reset on every `processSpec`: canvas-entity.ts:332 sets `_metricsViews` without an equality guard (Svelte writables always notify for object values), and time-manager.ts `checkIfHasTimeSeries` builds and sets fresh Maps each call — so `processSpec` triggers the interval derivation twice per run. `processSpec` itself runs on every spec-store emission (including isFetching flips with unchanged data), and every Canvas resource reconcile event refetches the `resolveCanvas` query (resource-invalidators.ts:188-194), so each YAML-edit reconcile or upstream refresh produces a burst of duplicate MetricsViewTimeRanges requests (one-plus per metrics view). `fetchTimeRanges` uses `staleTime: 60` (60 ms, likely meant 60 s) and `deriveInterval` sets `cacheBust` whenever the range contains "now" (invalidating the cache before fetching), so these re-runs are real network/OLAP round-trips rather than cache hits. The new Interval object identity also propagates through `timeAndFilterStore`, causing Svelte recompute/render churn in all components — though downstream chart data queries are value-keyed and deduped by TanStack, so there is no chart-data refetch storm and no correctness impact.

**Manual reproduction:**
1. In Rill Developer, open a project with a canvas dashboard that references 2 or more metrics views with time dimensions (e.g. add a chart from each metrics view), with a default time range set (e.g. `defaults: time_range: P7D`).
2. Open the canvas file in the editor so the split code/preview view is showing.
3. Open browser DevTools > Network and filter requests by "time-ranges" (the MetricsViewTimeRanges query endpoint); note the baseline request count after the page settles.
4. Make a trivial, time-unrelated edit to the canvas YAML (e.g. change `display_name` by one character) and wait for the save/reconcile to complete.
5. Observe a burst of new MetricsViewTimeRanges requests, at least one per referenced metrics view, even though the time range, timezone, and metrics view list are unchanged.
6. Repeat the edit several times: each reconcile produces the same duplicate burst. With a time range containing "now", the cacheBust path invalidates the query cache first, so duplicates fire even when re-runs land within the 60 ms staleTime window.
7. Optional: trigger a refresh of an upstream source/model the canvas depends on and observe the same duplicate burst without touching the canvas file, confirming it fires on any spec emission, not just edits.

**Suggested fix:** Guard the store writes with equality checks (skip `_metricsViews.set`/`hasTimeSeriesMap.set` when contents are unchanged, as the neighboring `checkAndSet*` helpers already do), and/or short-circuit the interval derived when (range, metricsViewNames, timeZone) are unchanged from the previous run. Revisit `staleTime: 60` (likely meant 60_000).

---

## CANV-5: Module-level `lastKey` shared across all pivot instances resets pagination

**File:** `web-common/src/features/canvas/components/pivot/util.ts:88` · **Severity:** medium · **Verdict:** CONFIRMED

The module-level singleton `let lastKey` (line 88) is shared by all canvas pivot/table component instances, but `processPivot`/`processFlat` use it to detect "config changed, reset pagination" per component. Since `getPivotConfigKey` excludes `rowPage` and each `PivotCanvasComponent` has its own config-derived store, any recompute of component B stores keyB in the shared variable; the next recompute of component A (including the recompute triggered by A's own infinite-scroll `rowPage` increment) sees `lastKey !== keyA` and resets A's `rowPage` to 1, discarding loaded pages and re-issuing page-1 aggregation queries. On any canvas with two or more table/pivot components, alternating scrolls between them (or any shared specStore/timeAndFilterStore emission) makes it impossible to keep both components paged past page 1. The explore counterpart (dashboards/pivot/pivot-data-config.ts) uses the same module-level pattern but is safe in practice because explore mounts one pivot at a time; the pattern became a bug when copied to canvas where many pivots coexist.

**Manual reproduction:**
1. In Rill Developer, open or create a canvas dashboard containing two table (or pivot) components, A and B, with different column sets, each backed by a metrics view whose dimension has more distinct rows than one page (more than ~50 rows so infinite scroll triggers additional pages).
2. Open the browser DevTools network tab and filter for metrics-view aggregation requests.
3. Scroll table A to the bottom so it fetches its next page: additional rows append and a new aggregation request is issued.
4. Scroll table B to the bottom so it fetches its next page (this stores B's config key in the shared module-level variable).
5. Return to table A and scroll to the bottom again to load its next page.
6. Observe the bug: instead of appending the next page, table A snaps back to page 1 — its scroll position and previously loaded rows reset, and the network tab shows the page-1 aggregation queries for A re-issued. Alternating between the two tables reproduces the reset every time.

**Suggested fix:** Move the key tracking into per-component state: store `lastKey` on `PivotCanvasComponent` (or close over it inside `createPivotConfig`) instead of at module scope.

---

## CANV-6: Async `TimeState.interval` has no stale-response guard; older request overwrites newer interval

**File:** `web-common/src/features/canvas/stores/time-state.ts:199` · **Severity:** medium · **Verdict:** CONFIRMED

The async `interval` derived (lines 142-209) launches `Promise.all(metricsViewsWithTimeSeries.map((mvName) => deriveInterval(range, client, mvName, timeZone)))` and calls `set(latestInterval)` plus `this.minMaxTimeStamps.set(...)` in `.then()` with no cancellation, sequence token, or cleanup. `deriveInterval` performs a real network fetch via `queryClient.fetchQuery` with `staleTime: 60` ms, so consecutive invocations (user changes time range or timezone) race, and TanStack's in-flight request deduplication makes out-of-order resolution deterministic when switching A → B → A quickly (the third invocation joins A's earlier in-flight request, so B's request resolves last and its stale result overwrites the store). The result: `rangeStore` shows the new range in the time pill while `interval` — which `BaseCanvasComponent.timeAndFilterStore` uses to build the `timeRange` for every component data query — holds the superseded window, and `canPanStore`/`minMaxTimeStamps` are computed against stale values. The wrong state persists until the next input change re-triggers the derived.

**Manual reproduction:**
1. In Rill Developer, open a project with a canvas dashboard whose metrics view has a time dimension, and open the canvas in preview mode.
2. Open browser DevTools, go to the Network tab, and set throttling to "Slow 3G" so that MetricsViewTimeRanges requests take multiple seconds.
3. Using the global time range pill, quickly select three ranges in succession, ending on the first one, e.g. "Last 24 hours" -> "Last 7 days" -> "Last 24 hours", completing all three clicks before the first request finishes (TanStack dedup makes the third selection reuse the first in-flight request, so the "Last 7 days" response resolves last and wins).
4. Wait for all MetricsViewTimeRanges requests to settle. The time pill shows "Last 24 hours", but inspect the subsequent component data requests in the Network tab: their timeRange start/end span 7 days, and charts render the 7-day window.
5. Optionally verify the pan arrows: with the mismatched state, the left/right pan availability (`canPanStore`) is computed against the stale interval rather than the displayed range.
6. Change any time input (range, grain, or timezone) once more and let it settle: the state self-corrects, confirming the stale write was from the superseded async batch.

**Suggested fix:** Guard with a monotonically increasing token (`const myToken = ++this.intervalToken; ... if (myToken !== this.intervalToken) return;`) or return a cleanup from the derived callback that sets a cancelled flag.

---

## CANV-7: In-list flag not removed on conversion to Select: `filter(...)` result discarded

**File:** `web-common/src/features/canvas/stores/filter-state.ts:309` · **Severity:** medium · **Verdict:** CONFIRMED

In `toggleDimensionValueSelections` (line 309), when an existing in-list filter is converted to a Select filter, the code calls `dimensionsWithInListFilter.filter((d) => d !== dimensionName)` and discards the result, so the dimension is never removed from the in-list flag array (the sibling action `applyDimensionInListMode` correctly mutates the same array with `push`). The immediately following `getFilterParam` call therefore serializes the newly created IN expression back into the URL with `IN LIST` syntax (converters.ts lines 113-117), and the URL round-trip re-parse (`convertFilterParamToExpression` → `getDimensionFilterItemsMap` line 593) marks the pill as `DimensionFilterMode.InList` again. The user sees a "Converted filter type to Select" toast, and the filter's values are replaced with the clicked value(s), but the pill silently stays in In-List mode; query results are still correct, so the impact is a persistent UI/URL state mismatch rather than wrong data.

**Manual reproduction:**
1. Start Rill Developer with a project that has a canvas dashboard containing a leaderboard component backed by a metrics view with a string dimension.
2. Open the canvas dashboard.
3. In the filter bar, add a filter pill for that dimension, open the pill dropdown, switch its mode to "In list", and paste two or more values, then apply. Confirm the URL filter param now reads `<dimension> IN LIST ('a','b')`.
4. In the leaderboard component, click one of the dimension's values to select it.
5. Observe the toast "Converted filter type to Select" appears.
6. Check the browser URL: the filter param still uses `IN LIST (...)` syntax instead of `IN (...)`.
7. Reopen the filter pill dropdown: the pill is still in "In list" mode, not "Select", contradicting the toast; the value list has been replaced with just the clicked value.

**Suggested fix:** Mutate in place: `const idx = dimensionsWithInListFilter.indexOf(dimensionName); if (idx !== -1) dimensionsWithInListFilter.splice(idx, 1);`.

---

## CANV-8: FilterManager UI-filter stores leak inner subscriptions on each rebuild

**File:** `web-common/src/features/canvas/stores/filter-manager.ts:178` · **Severity:** medium · **Verdict:** CONFIRMED

`activeUIFiltersStore` (line ~178), `defaultUIFiltersStore` (line ~141), and `filterMapStore` (line ~196) all follow the pattern `derived([this.metricsViewFilters], ([...], set) => { derived([...stores]).subscribe(set); })` with the inner unsubscriber discarded and no cleanup function returned, so Svelte cannot dispose the inner subscription when the outer callback re-runs or when the outer store loses its last subscriber. A new inner subscription chain is leaked (a) each time `updateConfig` registers a metrics view not seen before (`StoreOfStores.set` replaces the Map, re-running the outer callback) and (b) each time a component mounts and takes the outer store from 0 to 1 subscribers (CanvasFilters.svelte subscribes to both `activeUIFiltersStore` and `filterMapStore`). The inner deriveds subscribe to always-live writables on the FilterManager, which is owned by the registry-cached CanvasEntity whose `dispose()` never tears down the FilterManager, so chains accumulate for the session lifetime of the cached canvas entity. Each leaked chain re-runs the nontrivial `convertToUIFilters` on every pinned/temporary/filter change (CPU cost grows with each mount cycle), and after a spec update adds a metrics view, the older chain (missing the new view's `parsed` store) fires first and synchronously emits a `UIFilters` value computed without the new view's filters before the newer chain corrects it, which synchronous downstream deriveds (canvas pivot `whereFilterStore`, chat context) can observe. The leak is bounded per FilterManager (the cycle is GC-able once the entity is evicted and unreferenced) but unbounded across mount cycles while the entity remains cached. See CANV-9 for the recomputation-cost profile of the same defect.

**Manual reproduction:**
1. Run Rill Developer on a project with a canvas dashboard that has filters enabled (filter bar visible) and at least one pinned or active filter.
2. Open the canvas dashboard, then navigate away (e.g. to a model or the project home) and back to the canvas 10-15 times; each visit mounts CanvasFilters.svelte and leaks one inner subscription chain per store.
3. In browser DevTools, set a breakpoint (or a logpoint counting invocations) inside `FilterManager.convertToUIFilters` in filter-manager.ts, then toggle a filter pin or add a temporary filter once.
4. Observe `convertToUIFilters` executes once per leaked chain plus one (roughly N+1 times for N revisits) for a single user action; the count grows by one with every additional visit to the canvas.
5. For the stale-emission variant: with the filter bar mounted, edit the canvas YAML to add a component referencing a metrics view not previously used on the canvas; after reconcile, add a subscriber log on `activeUIFiltersStore` and change a pinned filter: the store synchronously emits an intermediate `UIFilters` value missing the new metrics view's filters before emitting the corrected value.
6. Memory observation: take a heap snapshot before and after 20 navigation cycles on the same canvas and compare retained closure counts referencing FilterManager/`convertToUIFilters`; they grow linearly with visits while the canvas entity stays cached.

**Suggested fix:** Return the inner unsubscriber from the outer derived callback so Svelte disposes it on re-run and on teardown.

---

## CANV-9: FilterManager leaked subscriptions rerun `convertToUIFilters` per past mount (companion to CANV-8)

**File:** `web-common/src/features/canvas/stores/filter-manager.ts:141` · **Severity:** medium · **Verdict:** CONFIRMED

Companion finding to CANV-8, verified independently with a focus on the recomputation cost. The three UI-filter stores (`defaultUIFiltersStore` line 141, `activeUIFiltersStore` line 178, `filterMapStore` line 196) end their outer derived callback with `).subscribe(set);` and return nothing. Because Svelte derived stores are lazy, the outer callback re-runs on every first-subscriber transition — i.e. every SPA navigation into the canvas page — and also whenever `StoreOfStores.set` notifies (new metrics view added via `updateConfig`/`processSpec`). Each re-run creates a fresh inner subscription while the previous one stays permanently attached to the FilterManager's writable stores (pinned/temporary/required keys, allMeasures/allDimensions, each `FilterState.parsed`). The FilterManager lives for the whole browser session: CanvasEntity is cached in `canvasRegistry` (`removeCanvasStore` has no callers) and `CanvasEntity.dispose()` does not tear down FilterManager stores. After N visits to a canvas, a single filter interaction runs `convertToUIFilters` N+1 times — each a full pass over all merged measures and dimensions plus two sorts whose comparator rebuilds `Array.from(temporaryFilterKeys.keys())` per comparison — and each leaked closure calls `set`, triggering N+1 downstream notifications and filter-bar re-renders. Notification order means the final value is still correct (newest subscription fires last), so this is a growing CPU/memory leak rather than a correctness bug.

**Manual reproduction:**
1. Start Rill Developer (`rill start`) on a project containing a canvas dashboard backed by at least one metrics view with filters enabled (a pinned filter and a few dimension filters make the effect easier to see).
2. Open the canvas dashboard in the browser, then open Chrome DevTools > Sources and set a logpoint with `console.count('convertToUIFilters')` at the top of `convertToUIFilters` in filter-manager.ts (line ~402).
3. Click a filter pill's pin toggle once and note the number of hits (baseline, typically 1-2 per triggering store).
4. Using in-app SPA navigation (not a full page reload), navigate away from the canvas and back. Repeat 10-20 times.
5. Toggle the same filter pin again and count hits: it now fires roughly once per prior visit for a single click (N+1 executions), confirming the accumulated leaked inner subscriptions.
6. Memory side: in DevTools > Memory, take a heap snapshot, navigate in/out of the canvas 10 more times, take a second snapshot, and compare — the retained-closure count for subscribers attached to `pinnedFilterKeysStore`/`temporaryFilterKeysStore` grows linearly with visits and is never released.
7. Optional user-visible symptom: with many accumulated visits on a canvas with many measures/dimensions, filter pill interactions become measurably slower (record a Performance profile while clicking a pill and observe repeated `convertToUIFilters` frames).

**Suggested fix:** Return the inner unsubscriber from each outer derived callback (`return derived(...).subscribe(set);`) so Svelte tears the previous inner subscription down on re-run and on last-unsubscribe.

---

## CANV-10: `clearDefaultFilters` throws when the YAML has no `defaults` key

**File:** `web-common/src/features/canvas/stores/canvas-entity.ts:498` · **Severity:** medium · **Verdict:** CONFIRMED

In `clearDefaultFilters` (line 498), `yaml.get("defaults").items.length` is evaluated outside any try/catch. When the canvas YAML has no `defaults:` mapping (the state of every newly created canvas) or `defaults:` is empty/null, `yaml.get("defaults")` returns undefined/null and the property access throws an uncaught `TypeError: Cannot read properties of undefined (reading 'items')`, aborting the function before the trailing `yaml.delete("filters")` and `updateEditorContent` run. The preceding `deleteIn` calls are inside a try/catch (and do throw "Expected YAML collection at defaults" when the key is missing, which is swallowed), but that catch does not cover line 498. The button is one click away: the inspector's always-visible "Defaults" tab (CanvasTabs.svelte via PageEditor.svelte:268) unconditionally renders DefaultFilterDisplay.svelte, whose "Clear default filters" button calls `clearDefaultFilters` directly. Net effect: an uncaught exception in the click handler, and if the file has a top-level `filters:` block (pinned/required) but no `defaults:`, that block is silently never removed. Verified empirically against the repo's installed yaml v2.8.3.

**Manual reproduction:**
1. Start Rill Developer (`rill start`) on a project and create a new canvas dashboard (or open any canvas whose YAML has no `defaults:` block).
2. Open the canvas in the visual editor and select the canvas page settings so the right-hand "Canvas configurations" inspector is visible.
3. Click the "Defaults" tab in the inspector header.
4. Open the browser DevTools console.
5. Click the "Clear default filters" button at the bottom of the panel.
6. Observe an uncaught `TypeError: Cannot read properties of undefined (reading 'items')` in the console originating from `clearDefaultFilters` in canvas-entity.ts; the YAML file content is not updated.
7. To observe the silent data-loss variant: hand-edit the canvas YAML to add a top-level `filters:` block (e.g. `filters:\n  pinned:\n    - some_dimension`) while keeping no `defaults:` block, then repeat steps 3-5 and observe the `filters:` block remains in the file even though the clear action should have removed it.

**Suggested fix:** Guard the access: `const defaults = yaml.get("defaults"); if (isMap(defaults) && defaults.items.length === 0) { ... }`, mirroring the `isMap`/`instanceof YAMLMap` check already used in `saveDefaultFilters` (lines 440-443).

---

## CANV-11: Exponential O(2^n) metrics-view combination map rebuilt on every spec emission

**File:** `web-common/src/features/canvas/stores/filter-manager.ts:1063` · **Severity:** medium · **Verdict:** CONFIRMED

`FilterManager.updateConfig` ends by calling `createDimensionsForUniqueCombinationsOfMetricsViews`, which invokes `getAllCombinations` (line 1063) to enumerate every non-empty subset of the canvas's referenced metrics views (2^n − 1 subsets, each `structuredClone`'d) and builds a merged dimension Map per subset. `updateConfig` is invoked by `CanvasEntity.processSpec` on every specStore emission (canvas-entity.ts:551-559 subscribes with no equality guard beyond a first-load skip), and canvas resource write events force a ResolveCanvas refetch (resource-invalidators.ts:189-195), so the full exponential rebuild reruns on every canvas reconcile/save even when metrics views are unchanged. Only keys corresponding to the distinct metrics-view groups produced by `mergeFilters` are ever looked up in `convertToUIFilters` (lines 453/472), so all but a linear number of the 2^n entries are dead weight. Cost is negligible for typical canvases (1-5 metrics views) but grows to hundreds of thousands of Map insertions at ~15 metrics views and multi-second main-thread freezes plus large memory growth at ~20.

**Manual reproduction:**
1. In a Rill Developer project, create one model and 16 metrics views over it (copy one metrics view YAML 16 times with distinct names, each with ~10 dimensions).
2. Create a canvas dashboard with one KPI or chart component per metrics view so all 16 metrics views are referenced by the canvas.
3. Open the canvas preview in the browser and open Chrome DevTools > Performance; start recording.
4. Edit the canvas YAML (e.g., change `display_name`) and save; each save reconciles the Canvas resource, which refetches ResolveCanvas and reruns `FilterManager.updateConfig`.
5. Stop the recording and inspect the flame chart: observe a long main-thread task inside `createDimensionsForUniqueCombinationsOfMetricsViews` / `getAllCombinations` / `structuredClone` (65,535 subsets at 16 metrics views), repeated on every save and on query status flips.
6. Optionally scale to 20 metrics views (~1M subsets): the tab visibly freezes for seconds on every reconcile, and the DevTools Memory panel shows a large heap jump from the ~1M-entry Map of dimension arrays.

**Suggested fix:** Compute dimension arrays only for the metrics-view groupings that actually occur (the distinct `mvNames` sets produced by `mergeFilters`), instead of all 2^n subsets, and skip the rebuild entirely when the metrics-view specs are unchanged since the last `updateConfig`.

---

## CANV-12: Unconditional `specStore.set` on every spec emission rebuilds every chart's query pipeline

**File:** `web-common/src/features/canvas/components/BaseCanvasComponent.ts:180` · **Severity:** medium · **Verdict:** CONFIRMED

`CanvasEntity.acquire()` (canvas-entity.ts:551-559) calls `processSpec(data)` on every emission of the canvas-level TanStack query store with no data-changed guard, and query result objects are new on every observer notification (refetch, invalidation, isFetching toggles). `processRows` then calls `existingClass.update(newResource, path)` for every component on the canvas (canvas-entity.ts:757), and `BaseCanvasComponent.update()` (lines 172-181) unconditionally calls `this.resource.set(resource)` and `this.specStore.set(yamlSpec)`. Svelte's `writable.set` uses `safe_not_equal`, which returns true for any object value even when the reference is identical, so every subscriber of every component's `specStore` is notified on every canvas spec emission regardless of whether that component's spec changed. This has two per-component costs: (1) the constructor subscription (lines 157-165) re-runs `onFilterStringChange` and `onUrlChange`; (2) in CanvasChart.svelte, `$: chartSpec = $specStore` dirties `$: chartData = getChartDataForCanvas(...)`, which calls `component.createChartDataQuery(...)` — each provider constructs fresh derived stores and multiple new `createQuery` observers (topN X, topN color, main aggregation), and the `$chartData` auto-subscription swap destroys the old QueryObservers and mounts new ones for every chart on every emission. Mitigations that bound the impact: the TanStack cache dedupes actual network fetches for identical query keys, and emissions occur per debounced file-save/reconcile/refetch rather than literally per keystroke — so the cost is main-thread observer/derived-chain churn and re-renders proportional to chart count, visible as jank on large canvases during editing or frequent reconciles, not redundant network traffic.

**Manual reproduction:**
1. In Rill Developer, open or create a project with a canvas dashboard containing many chart components (10 or more charts across several rows) backed by a metrics view.
2. Open the canvas file so the visual canvas preview is rendered, and open the browser DevTools Performance panel.
3. (Optional, clearest signal) Add a temporary `console.count(component.id)` at the top of `getChartDataForCanvas` in web-common/src/features/canvas/components/charts/selector.ts and run the dev frontend.
4. Start a Performance recording, then make a trivial edit that only affects one component — e.g. change one chart's `title` in the canvas YAML — and let autosave/reconcile complete. Repeat the edit a few times.
5. Stop the recording. Observe that on every save/reconcile, `getChartDataForCanvas` / `createChartDataQuery` runs once for every chart on the canvas (the console.count increments for all N charts, not just the edited one), and the flame chart shows a long task per emission spent constructing derived stores and TanStack QueryObservers.
6. On a canvas with many charts, this is user-visible as main-thread jank (dropped frames, sluggish typing in the split-view YAML editor) during rapid edits. Refocusing the browser window (which triggers a query refetch returning identical data) reproduces the same full rebuild without any edit at all.

**Suggested fix:** In `update()`, skip `specStore.set` when the new `yamlSpec` reference (or a cheap deep-equal / JSON string) matches the current value, and likewise for `resource`. Alternatively make CanvasChart derive its query from the store directly instead of re-invoking `getChartDataForCanvas` on each spec object emission.

---

## CANV-13: Markdown templated-string query store leaks inner subscriptions and fresh `timeAndFilterStore` chains

**File:** `web-common/src/features/canvas/components/markdown/util.ts:194` · **Severity:** medium · **Verdict:** CONFIRMED

`getResolveTemplatedStringQueryOptions` (lines 135-196) nests a derived inside the async-form callback of an outer derived and calls `.subscribe(set)` while discarding the unsubscriber and returning nothing, so Svelte never cleans up the inner subscription. Every run of the outer callback — once per Markdown component mount (the outer store is created and subscribed per mount in Markdown.svelte) and once per emission of `filterManager.metricsViewFilters` (which emits when a new metrics view's filter store is registered) — permanently leaks a live pipeline. Each run also invokes the `component.timeAndFilterStore` getter (BaseCanvasComponent.ts:184), which constructs and subscribes an entirely new 17-source derived chain whose sources live as long as the canvas, so leaked pipelines are retained indefinitely and keep recomputing `buildRequestBody` (serializing all parsed metrics-view filters) and building query options on every subsequent time or filter change. In a session that repeatedly remounts templated markdown tiles (dashboard navigation, edit/preview toggles), CPU work per filter/time change grows linearly and old spec/filter objects are retained until page reload. (Note: the dominant leak vector is component remounts, not steady-state reconciles — `metricsViewFilters` emits only when a new metrics view's FilterState is first registered.)

**Manual reproduction:**
1. In Rill Developer, create a canvas dashboard containing a markdown component whose content uses templating syntax (e.g. `{{ ... }}` referencing a metrics view measure) alongside at least one metrics-view-backed component so filters and time controls are active.
2. Open the canvas dashboard in the browser and open Chrome DevTools. In the Sources tab, add a logpoint (or `console.count('buildRequestBody')`) inside `buildRequestBody` in web-common/src/features/canvas/components/markdown/util.ts.
3. Change the global time range once and note the number of `buildRequestBody` invocations (baseline: 1 per markdown tile).
4. Remount the markdown component about 10 times by navigating away from the canvas and back (or toggling between the canvas edit and preview views).
5. Change the global time range or a dimension filter once more: observe `buildRequestBody` now fires roughly once per past mount (10+ times) for a single user action, and the count keeps growing with every additional remount.
6. Memory retention: in DevTools > Memory, take a heap snapshot, perform 10 more remount cycles, take a second snapshot, and compare — closures referencing the markdown util derived callback and TimeAndFilterStore chains accumulate and are never released until a full page reload.

**Suggested fix:** Return the inner unsubscriber from the outer callback (`return derived(...).subscribe(set);`) and hoist `component.timeAndFilterStore` to a single memoized store created once per component.

---

## CANV-14: `Grid.updateFromCanvasRows` sets row stores unconditionally, making its own diff dead code

**File:** `web-common/src/features/canvas/stores/grid.ts:101` · **Severity:** low · **Verdict:** CONFIRMED

In `Grid.updateFromCanvasRows` (grid.ts:98-115), `row.items` and `row.widths` are set unconditionally at lines 101-102 before the changed-check at lines 104-109, making the diff guard dead code: since Svelte writables always notify on `set` with a fresh array, every row store in the top-level grid and in every tab group emits new array identities on every `processSpec` run (which fires on every specStore emission, including reconcile-driven refetches with unchanged data). Additionally, `widths` is set twice per call with inconsistent fallback defaults (25 at line 102 vs `COLUMN_COUNT / length` at line 112). Each emission causes a wasted invalidate-and-flush of every `StaticCanvasRow`/`EditableCanvasRow` component (re-running `normalizeSizeArray`, the grid-template string rebuild, and keyed-each diffing). Because the each blocks are keyed by column index and Svelte batches the synchronous sets into one flush, no components remount and no visual glitch occurs — the cost is redundant reactive work per row per spec emission, exactly what the (dead) diff was written to avoid.

**Manual reproduction:**
Not directly user-visible (no remounts or visual glitches occur because the each blocks are keyed and Svelte batches the flush); observe via instrumentation:
1. In a Rill project with a canvas dashboard containing several rows, add a temporary `console.count("row-flush")` inside the reactive statement `$: widths = normalizeSizeArray($itemWidths)` in web-common/src/features/canvas/StaticCanvasRow.svelte (line 24) and EditableCanvasRow.svelte (line 68).
2. Run Rill Developer (`rill start`) and open the canvas dashboard in edit mode.
3. Trigger a reconcile that does not change the canvas layout: edit an unrelated field in the canvas YAML (e.g. `display_name`) or save an unrelated file in the project.
4. Observe in the browser console that the counter increments once per row (and per tab-group row) on every spec emission, even though row items and widths are unchanged — the diff at grid.ts:104-109 never suppresses a notification.
5. Optionally, set a breakpoint on grid.ts line 101 and confirm it executes on every emission before the guarded set, and that the guarded branch at line 108 is redundant when it runs.

**Suggested fix:** Delete the unconditional `row.items.set`/`row.widths.set` at lines 101-102 and keep only the guarded sets; add the same changed-check for widths before setting.
