# Deep Audit: UI State Management

Scope: `web-common/src/features/dashboards/stores`, `web-common/src/features/dashboards/state-managers`, `web-common/src/features/dashboards/url-state`, `web-common/src/features/dashboards/proto-state`, plus `web-common/src/features/canvas/stores` and `web-common/src/features/canvas/state-managers`. Method: a multi-lens find phase (correctness, state lifecycle, memory, performance, security lenses over the state-management surface) followed by an adversarial verify phase in which every candidate finding was re-traced end to end in code, its trigger paths checked for reachability, and its severity re-scored; refuted claims are listed in the appendix.

## Summary

| ID | Title | Severity | Category | Verdict |
|---|---|---|---|---|
| STAT-1 | Malformed or oversized `gzipped_state` URL param crashes dashboard load with no error containment | high | correctness | CONFIRMED |
| STAT-2 | `DashboardStateSync.handleURLChange` leaks the `updating` lock on early returns, disabling URL/state sync | medium | state | CONFIRMED |
| STAT-3 | `DashboardStateSync.handleURLChange` lock also stuck via empty valid spec or storage-quota throw | medium | state | CONFIRMED |
| STAT-4 | `Filters.toggleMultipleDimensionValueSelections` keeps old LIKE expr when converting to IN (duplicate filters) | medium | correctness | CONFIRMED |
| STAT-5 | Leaked inner derived subscriptions in `FilterManager` UI filter stores multiply heavy recomputation | medium | memory-leak | CONFIRMED |
| STAT-6 | `getAllCombinations` builds 2^N subsets of metrics views on every canvas spec update | medium | performance | CONFIRMED |
| STAT-7 | Canvas `FilterState.toggleDimensionValueSelections` discards `Array.filter` result; dimension stuck in in-list mode | medium | correctness | CONFIRMED |
| STAT-8 | Canvas `TimeState.interval` derived store: stale async query result can overwrite the newer interval (race) | medium | state | CONFIRMED |
| STAT-9 | `memoizeMetricsStore` module-level cache is never pruned and pins the first mount's `StateManagers` forever | medium | memory-leak | CONFIRMED |
| STAT-10 | Canvas `lastVisitedState` keyed by canvas name only, leaking filter state across projects and orgs | medium | security | CONFIRMED |
| STAT-11 | `FilterManager` nested derived stores never unsubscribe inner deriveds; subscriptions leak on spec updates/remounts | low | memory-leak | CONFIRMED |
| STAT-12 | `timeRangeSummaryStore` derives from the entire `dashboardStore`, recreating its QueryObserver on every mutation | low | performance | CONFIRMED |
| STAT-13 | Legacy proto-state deserialization throws on crafted/corrupt input and is exported without containment | low | correctness | CONFIRMED |

Note: STAT-2 and STAT-3 describe the same underlying defect (the `updating` lock in `DashboardStateSync.handleURLChange` is never released on early return) surfaced independently by two lenses with different trigger analyses; they are reported separately because the verified trigger sets differ. STAT-5 and STAT-11 likewise cover the same nested-derived pattern in `FilterManager` from the CPU-multiplication and subscription-lifecycle angles respectively. One fix resolves each pair.

---

## STAT-1: Malformed or oversized `gzipped_state` URL param crashes dashboard load with no error containment

**File:** `web-common/src/features/dashboards/url-state/convertURLToExplorePreset.ts:85` · **Severity: high** · **Verdict: CONFIRMED**

`convertURLToExplorePreset` (convertURLToExplorePreset.ts:83-89) inflates the `gzipped_state` URL param with no try/catch. `decompressUrlParams` (compression.ts:19-23) throws on hostile or damaged input at three points, all empirically verified: `protoBase64.dec` throws "invalid base64 string." on non-base64 input, `gunzipSync` throws "invalid gzip data" on corrupt or truncated payloads, and `String.fromCharCode(...bytes)` throws `RangeError: Maximum call stack size exceeded` once the decompressed params exceed the JS engine's argument limit (verified at 500KB in Node; V8's limit is on the order of 100k elements). No caller contains the exception: `convertURLSearchParamsToExploreState` → `DashboardStateDataLoader.getMergedExploreState` (line 330) → the `initExploreState` getter inside `getCompoundQuery`'s derived, subscribed by `DashboardStateSync`'s constructor and `DashboardStateManager.svelte`. The `ErrorPage` fallback in `DashboardStateManager` only renders errors from the query result's `error` field, so a thrown getter bypasses it: the dashboard is stuck on the loading spinner (or crashes during component init if specs are already cached) with only an uncaught console exception. The gap is clearly an oversight: the adjacent legacy `state` param is wrapped in try/catch (`fromLegacyStateUrlParam`, lines 244-260), and the session-storage path through the same converter is also wrapped (explore-web-view-store.ts:127-158). Truncated shared links are a realistic trigger because compression only activates for URLs over 2000 characters, so every real `gzipped_state` link is long enough to be truncated by messaging clients.

**Manual reproduction:**

1. Start Rill Developer (`rill start`) on any project containing an explore dashboard, and open the dashboard at `http://localhost:9009/explore/<explore_name>` to confirm it loads normally.
2. Edit the URL to append the parameter `?gzipped_state=abc` (i.e. `http://localhost:9009/explore/<explore_name>?gzipped_state=abc`) and press Enter.
3. Observe the dashboard never renders: the page is stuck on the loading spinner or goes blank, with no error page and no fallback to default dashboard state. Open DevTools Console and observe the uncaught exception `invalid gzip data` (use `?gzipped_state=%21%21%21` to instead get `invalid base64 string.`).
4. Contrast with the guarded legacy param: load `http://localhost:9009/explore/<explore_name>?state=abc` — the dashboard loads with default state despite the garbage value, because that code path has a try/catch.
5. To reproduce the realistic truncated-link variant: apply enough dimension filters (e.g. a large in-list filter) that the URL exceeds 2000 characters and Rill rewrites it to `?gzipped_state=<long-token>`; copy that URL, delete the last 5 characters of the token, and open it in a new tab — same crash instead of a graceful error.

**Suggested fix:** Wrap the decompress call in try/catch and push a `getSingleFieldError`-style error falling back to defaults (mirroring `fromLegacyStateUrlParam`). In `decompressUrlParams`, replace the `String.fromCharCode` spread with `new TextDecoder().decode(...)` (no argument-count limit, and correct UTF-8 symmetry with the `TextEncoder` used in `compressUrlParams`); optionally cap the decompressed size.

---

## STAT-2: `DashboardStateSync.handleURLChange` leaks the `updating` lock on early returns, disabling URL/state sync

**File:** `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:225` · **Severity: medium** · **Verdict: CONFIRMED**

`handleURLChange` sets the re-entrancy lock `this.updating = true` (line 215) and then has two early returns that never release it: line 225 when `rillDefaultExploreURLParams.data` is undefined and line 235 when `getExploreStateFromURLParams` returns undefined. Both become reachable after initialization when the explore or its metrics view loses its validSpec: in Rill Developer, `rill.stage_changes` is "false" (cli/pkg/local/app.go:217), so a YAML validation error clears `State.ValidSpec` in the runtime, and the frontend's GetExplore refetch then yields undefined specs. A single afterNavigate (e.g. browser Back on the same dashboard) during that window leaks the lock: thereafter `handleURLChange` bails at line 214 and `gotoNewState` bails at line 311, so filter/time-range/sort changes never update the URL, session storage, or most-recent state, and back/forward no longer updates the dashboard, until the component remounts (navigate away and back, or reload). Note: contrary to the original report, a rejected `resolveTimeRanges` call cannot leak the lock (it catches errors internally and always resolves), and ordinary query refetches do not trigger the bug because TanStack Query retains previous data; the validSpec must actually be cleared.

**Manual reproduction:**

1. Start Rill Developer (`rill start`) on a project that has a working explore dashboard (stage_changes is forced to false locally, which is required for this repro).
2. Open the explore dashboard in the browser at `http://localhost:9009/explore/<name>` and wait for it to load.
3. Change a filter or time range so the URL gains query parameters (this creates a browser history entry for the same dashboard route).
4. In an external editor (not the Rill UI, so the dashboard page stays mounted), edit the dashboard's metrics view YAML to introduce a validation error (e.g. reference a nonexistent column in a measure) and save. Wait a few seconds for reconcile; the runtime clears the valid spec and the frontend refetches it.
5. Without leaving the dashboard page, press the browser Back button. This fires afterNavigate → handleURLChange, which sets the updating lock and returns early at the `if (!rillDefaultExploreURLParams) return;` guard because the valid spec is gone, leaking the lock.
6. Revert the YAML error and save; wait for the dashboard to become valid and render normally again.
7. Now change any filter, time range, or sort in the dashboard: observe the browser URL never updates (normally every such change rewrites the query string). Press browser Back/Forward: the dashboard state no longer follows the URL. New filter state is also no longer persisted to session storage, so switching views loses it.
8. Navigate to another page and back (or hard-reload): URL sync works again, confirming the leaked in-memory lock was the cause. To observe the lock directly, set a breakpoint or log in `DashboardStateSync.gotoNewState` and confirm it returns at the `if (this.updating) return;` guard on every state change after step 5.

**Suggested fix:** Wrap the body in try/finally releasing the lock, or reset `this.updating = false` before each early return.

---

## STAT-3: `DashboardStateSync.handleURLChange` lock also stuck via empty valid spec or storage-quota throw

**File:** `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:215` · **Severity: medium** · **Verdict: CONFIRMED**

Same defect as STAT-2, verified with a distinct trigger analysis. `DashboardStateSync.handleURLChange` sets `this.updating = true` (line 215) but the early returns at line 225 (`!rillDefaultExploreURLParams`) and line 235 (`!partialExplore`) never reset it, and the caller invokes it fire-and-forget (`void stateSync.handleURLChange(...)` in DashboardStateManager.svelte:95). These returns are reachable after initialization when the explore's valid spec becomes empty — e.g. the referenced metrics view file is deleted/renamed (the runtime then clears the explore's ValidSpec even with the default `stage_changes: true`; with `stage_changes: false` any YAML error does it) — and a back/forward navigation occurs during that window. Unwrapped `sessionStorage.setItem`/`localStorage.setItem` calls inside the locked region (explore-web-view-store.ts:181, most-recent-explore-state.ts:101) can also throw quota errors that leave the flag stuck. Once stuck, every future `handleURLChange` (line 214) and `gotoNewState` (line 311) bails, so for the rest of the session the URL stops following filter/time-range changes, session and most-recent state stop being saved, and back/forward stops updating the dashboard, with no error surfaced; only a page reload recovers. Two mechanisms in the original report do not exist: `resolveTimeRanges` cannot reject (it catches internally and returns), and `await goto(newUrl)` in `gotoNewState` resolves rather than rejects on aborted same-origin navigations in SvelteKit 2.55.

**Manual reproduction:**

1. Start Rill Developer (`rill start`) on a project with a working explore dashboard backed by a metrics view file (e.g. `metrics/my_metrics.yaml` referenced by `dashboards/my_explore.yaml`).
2. Open the explore dashboard in the browser. Apply a filter or change the time range and confirm the URL search params update (URL sync working).
3. Navigate within the dashboard at least once so there is browser history (e.g. change a filter, or switch to Pivot and back).
4. While the dashboard tab stays open, delete or rename the referenced metrics view YAML file in the project directory (or set `stage_changes: false` in rill.yaml and introduce a syntax error into the metrics view file). Wait a few seconds for reconcile; the explore's valid spec is now empty.
5. In the browser, press the Back button (or Forward) once. This fires afterNavigate → handleURLChange, which sets the `updating` lock and early-returns without resetting it. Nothing visible happens and no error is shown.
6. Restore the metrics view file (undo the delete/rename or fix the error) and wait for the dashboard to become healthy again; data renders normally.
7. Now change any filter, dimension, or time range in the dashboard. Observe the URL no longer updates to reflect the change, pressing Back/Forward no longer changes the dashboard state, and reopening the dashboard does not restore the latest state (session/most-recent state saving is dead). No error appears in the UI or console.
8. Reload the page and observe URL sync works again, confirming the stuck in-memory `updating` flag was the cause. (To observe directly: in DevTools, set a breakpoint at DashboardStateSync.ts line 214 and inspect `this.updating === true` on subsequent navigations.)

**Suggested fix:** Wrap the bodies of `handleURLChange` and `gotoNewState` in try/finally that resets `this.updating`.

---

## STAT-4: `Filters.toggleMultipleDimensionValueSelections` keeps old LIKE expr when converting to IN (duplicate filters)

**File:** `web-common/src/features/dashboards/stores/Filters.ts:312` · **Severity: medium** · **Verdict: CONFIRMED**

In the `Filters` class used by the alert and scheduled-report filter forms (Filters.ts lines 309-316), `toggleMultipleDimensionValueSelections` handles an existing LIKE/NLIKE (Contains) filter by pushing a brand-new IN expression onto `whereFilter` instead of replacing the LIKE expression in place, as the equivalent explore action does (state-managers/actions/dimension-filters.ts lines 88-93). The stale LIKE expression stays in the filter, so the query built for the alert/report ANDs both predicates; the filter-pill selector (`getDimensionFiltersMap`) dedupes by identifier and renders only the first (LIKE) expression, so the selection never appears in the UI; and because `getWhereFilterExpressionIndex` keeps finding the LIKE first, every further value toggle pushes another duplicate IN expression rather than toggling values, so values can never be deselected. The over-constrained where clause is silently persisted into unattended alerts/reports.

**Manual reproduction:**

1. In Rill Cloud, open an explore dashboard and choose Create Alert (or Home > Reports > Create Report) to open the alert/report dialog that contains the filters form (FiltersForm.svelte).
2. In the dialog's filter bar, add a dimension filter (e.g. country), switch the filter pill's mode to Contains, type a search string such as "us", and press Apply. The pill now shows a Contains filter (`LIKE '%us%'`).
3. Re-open the same filter pill and click one of the listed dimension values (e.g. "United States") while still in Contains mode.
4. Observe the pill still shows the Contains filter and never reflects the clicked value; clicking the same value repeatedly neither selects nor deselects it visibly.
5. Inspect the request in DevTools > Network: the preview/aggregation query (and the alert/report spec on save) contains BOTH `country LIKE '%us%'` AND `country IN ('United States')` under the same AND, and each additional click in step 3 appends another duplicate IN expression for the dimension.
6. Compare with the same interaction on the explore dashboard itself (outside the dialog): there the Contains pill is correctly converted to a Select pill with the clicked value and a "Converted filter type to Select" notification appears.

**Suggested fix:** When `wasLikeFilter` is true, replace the expression in place (`wf.cond!.exprs![exprIndex] = expr`) instead of pushing a new one, mirroring the actions/dimension-filters.ts implementation.

---

## STAT-5: Leaked inner derived subscriptions in `FilterManager` UI filter stores multiply heavy recomputation

**File:** `web-common/src/features/canvas/stores/filter-manager.ts:141` · **Severity: medium** · **Verdict: CONFIRMED**

All three composite stores in `FilterManager` (`defaultUIFiltersStore`, `activeUIFiltersStore`, `filterMapStore`) create an inner derived and call `.subscribe(set)` without returning the unsubscriber from the outer derived callback, so Svelte's cleanup is a no-op. Each time a new metrics view is added to the canvas while the filter UI is mounted (spec reconcile → `processSpec` → `updateConfig` → `StoreOfStores.set`, which only emits for newly-appearing metrics views), the outer callback re-runs and leaks the previous inner subscription. Stacked stale subscriptions then re-run `convertToUIFilters` (iterating and sorting all dimensions/measures across all metrics views) and emit duplicate values downstream on every filter interaction and every spec reconcile, and the last inner subscription survives component unmount. The final store value is always correct (stale copies fire before the fresh one, whose dependencies are a superset), and the marginal memory impact is small because the `FilterManager` is already retained by the global canvasRegistry; the practical cost is N-fold redundant recomputation and extra downstream re-renders during canvas editing sessions, where N grows by one per metrics view added while mounted.

**Manual reproduction:**

1. Start Rill Developer (`rill start`) on a project with at least three metrics views (A, B, C) and a canvas dashboard whose components initially reference only metrics view A.
2. Open the canvas dashboard preview so the filter bar renders (this subscribes CanvasFilters.svelte to `activeUIFiltersStore` and `filterMapStore`).
3. Open browser DevTools. To observe call counts, patch the running code: in the Sources panel, add a logpoint or breakpoint inside `FilterManager.convertToUIFilters` (filter-manager.ts, the `convertToUIFilters` method body) — or run a Performance recording.
4. Baseline: click a dimension filter pill and toggle a value. Observe `convertToUIFilters` runs once per composite store per state change.
5. In the canvas editor, add a component that references metrics view B and save; then add another component referencing metrics view C and save. Each newly-referenced metrics view triggers one `metricsViewFilters` emission and stacks one leaked inner subscription per composite store.
6. Repeat the same single filter interaction (e.g. toggle one dimension value, or click "+ Add filter" to add a temporary filter). Observe `convertToUIFilters` now fires 3x per store per interaction (once per stacked subscription) instead of once; the temporary-filter toggle fires the whole stack again 200ms later when the flag auto-clears.
7. To observe the unmount leak: navigate away from the canvas page, then take a heap snapshot and search for `FilterManager` / retained subscriber closures on its writable stores — the inner derived subscriptions created before unmount are still registered on `pinnedFilterKeysStore`, `temporaryFilterKeysStore`, `allMeasuresStore`, `allDimensionsStore`, and the `FilterState.parsed` stores (note the FilterManager itself is retained by the canvasRegistry regardless, so compare closure counts rather than object presence).

**Suggested fix:** Return the inner `.subscribe(set)` unsubscriber from the outer derived callback in all three stores so Svelte tears it down on each re-run and on final unsubscribe: `([metricsViewFilters], set) => { ...; return derived([...], ...).subscribe(set); }`.

---

## STAT-6: `getAllCombinations` builds 2^N subsets of metrics views on every canvas spec update

**File:** `web-common/src/features/canvas/stores/filter-manager.ts:1063` · **Severity: medium** · **Verdict: CONFIRMED**

`createDimensionsForUniqueCombinationsOfMetricsViews` (filter-manager.ts line 1084) calls `getAllCombinations(metricsViewNames)` (line 1093), which enumerates every non-empty subset of the canvas's referenced metrics views via DFS, cloning each path with `structuredClone` (line 1069). This is 2^N - 1 subsets for N distinct metrics views, and for each subset the caller merges all member views' dimension maps (lines 1097-1111), giving O(2^N × N × D) work and a result map with 2^N - 1 entries. It runs inside `FilterManager.updateConfig` (lines 373-375), which `CanvasEntity.processSpec` invokes on every canvas spec emission (canvas-entity.ts lines 352, 551-559) — i.e. on every reconcile, and in the canvas editor effectively after every YAML edit. Only the groupings that actually occur (one per distinct dimension/measure name, derived from `dimensionNameToMetricsViewNames`/`measureNameToMetricsViewNames`) are ever looked up in `convertToUIFilters` (lines 453, 472), so nearly the entire power set is wasted work. A canvas referencing ~20 metrics views performs ~1M subset iterations per spec update (multi-second freezes); ~25 views (~33M subsets) hangs or OOMs the tab. Canvases with the typical handful of metrics views are unaffected, which bounds the practical impact.

**Manual reproduction:**

1. In Rill Developer, create a project with 20 small metrics views (e.g. generate 20 copies of a simple metrics view over the same model, each with a few dimensions).
2. Create a canvas dashboard containing one KPI or table component per metrics view, so the canvas references all 20 distinct metrics views.
3. Open the canvas in the browser. Observe the tab freeze for seconds while the page loads (the constructor call to `updateConfig` builds the ~1,048,575-entry power set).
4. Open Chrome DevTools > Performance, start recording, then edit the canvas YAML (e.g. change the title) and save so a reconcile emits a new spec. Observe a long scripting block attributed to `getAllCombinations` / `createDimensionsForUniqueCombinationsOfMetricsViews` on every spec emission.
5. In DevTools > Memory, take a heap snapshot and observe the Map retained by `uniqueMetricsViewGroupingToDimensionArrayStore` holding ~1M entries.
6. Increase to 24-25 metrics views on the canvas and reload: the tab hangs indefinitely or crashes with an out-of-memory error.
7. As a control, repeat with 5 metrics views: no measurable delay, confirming the exponential dependence on the number of referenced metrics views.

**Suggested fix:** Compute the dimension arrays only for the metrics-view groupings that actually occur (the distinct `mvNames` sets from `dimensionNameToMetricsViewNames` / `measureNameToMetricsViewNames`), or compute entries lazily on first lookup, instead of enumerating the power set. Both sides already sort names, so restricting to occurring groupings preserves correctness.

---

## STAT-7: Canvas `FilterState.toggleDimensionValueSelections` discards `Array.filter` result; dimension stuck in in-list mode

**File:** `web-common/src/features/canvas/stores/filter-state.ts:309` · **Severity: medium** · **Verdict: CONFIRMED**

In canvas dashboards, `FilterState.toggleDimensionValueSelections` (filter-state.ts line 309) calls `dimensionsWithInListFilter.filter((d) => d !== dimensionName)` and discards the result, so when a user toggles a value on a dimension that has an in-list filter, the dimension is never removed from the in-list set. The returned filter string is therefore serialized with `IN LIST` syntax (converters.ts lines 112-117), and after the URL round-trip (`applyFiltersToUrl` → `onUrlChange` → `getFiltersFromText`) the pill reverts to In-list mode despite the "Converted filter type to Select" notification. Because the state remains stuck in in-list mode, every subsequent value click re-enters the conversion branch and replaces the entire selection with only the clicked values, so incremental multi-select on that dimension is broken (each click wipes the previous selection and re-fires the notification). The action is reachable from the canvas filter pill, leaderboard row clicks, and pivot click-to-filter.

**Manual reproduction:**

1. Start Rill Developer with a project that has a canvas dashboard whose filter bar is enabled and backed by a metrics view with a string dimension (e.g. run `rill devtool start local` or `rill start` on a demo project and open the canvas).
2. In the canvas filter bar, add a filter pill for a dimension, open the pill, switch the mode to "In list", paste two or more values (e.g. from the clipboard), and apply. Confirm the URL contains `filters.<metrics_view>=<dim> IN LIST (...)`.
3. Open the same pill's dropdown and click a checkbox to toggle one value in the value list.
4. Observe the "Converted filter type to Select" notification appears, but the URL is still written as `<dim> IN LIST (...)` instead of `<dim> IN (...)`, and after the URL updates the pill still shows In-list mode.
5. Click another value in the dropdown. Observe the notification fires again and the previous selection is replaced by only the newly clicked value, instead of both values being selected; the filter never actually behaves as a Select filter.

**Suggested fix:** Reassign the filtered array (e.g. `dimensionsWithInListFilter = dimensionsWithInListFilter.filter((d) => d !== dimensionName)` and use that variable in the subsequent `getFilterParam` call), or splice the entry in place since the array is part of the shared `parsed` store object.

---

## STAT-8: Canvas `TimeState.interval` derived store: stale async query result can overwrite the newer interval (race)

**File:** `web-common/src/features/canvas/stores/time-state.ts:162` · **Severity: medium** · **Verdict: CONFIRMED**

The canvas `TimeState.interval` derived store (time-state.ts, lines 127-210) launches network-backed `deriveInterval` calls via `Promise.all` inside the derivation callback and writes the result with `set(latestInterval)` and `this.minMaxTimeStamps.set(...)` in the `.then`, with no cancellation, no cleanup return, and no check that the inputs that produced the promise are still current. Because each range string maps to a distinct TanStack query key (rill-time-ranges.ts `fetchTimeRanges`) and hence an independent HTTP request — with cache hits (60 ms staleTime), `cacheBust` for "now"-based ranges, and multi-metrics-view `Promise.all` all adding latency variance — two rapid range (or timezone) changes can resolve out of order. When the older run resolves last, `interval` and `minMaxTimeStamps` are overwritten with the stale window and stay wrong until the next dependency change, so every canvas component queries and renders a time window that contradicts the URL and the range selector, and `canPanStore` enables/disables pan arrows incorrectly.

**Manual reproduction:**

1. In Rill Developer (`rill start`), open a project with a canvas dashboard whose metrics view has a time dimension (time series enabled).
2. Open the canvas dashboard in Chrome and open DevTools; in the Network tab filter for `time-ranges` (the MetricsViewTimeRanges query), and enable network throttling (e.g. Slow 3G) to widen the race window.
3. First select "Last 30 days" in the global time range dropdown and let it load, then switch to "Last 7 days", and immediately (before the 7-day time-ranges request resolves) switch back to "Last 30 days". The second selection may be served faster (cached/near-identical key timing) while the first is still in flight.
4. Watch the two in-flight `time-ranges` requests: if the request for the older selection (P7D) completes after the one for the newer selection (P30D), the stale `.then` fires last.
5. Observe the mismatch: the range pill and URL say "Last 30 days" while every chart/table on the canvas renders only 7 days of data (compare the x-axis extent of a time series chart against the selected range). The pan arrows (left/right chevrons next to the time range) may also be enabled/disabled incorrectly.
6. The wrong window persists until you change the range, timezone, or reload; changing any of them re-derives and self-heals. Because the bug is a network-timing race, repeat step 3 a few times under throttling if the first attempt resolves in order.

**Suggested fix:** Capture a generation token per derivation run (e.g. `const token = ++this.latestRun`) and ignore the promise result if `token !== this.latestRun` when it resolves; alternatively return a cleanup function from the derived callback that flips a `cancelled` flag checked in the `.then`.

---

## STAT-9: `memoizeMetricsStore` module-level cache is never pruned and pins the first mount's `StateManagers` forever

**File:** `web-common/src/features/dashboards/state-managers/memoize-metrics-store.ts:10` · **Severity: medium** · **Verdict: CONFIRMED**

The closure-scoped `cache` in `memoizeMetricsStore` (memoize-metrics-store.ts:10) is created once per module-scope call site (`useTimeControlStore`, `usePivotForExplore`, `useTimeSeriesDataStore`, `useDimensionTableData`, and the TDD timeseries variant) and lives for the whole JS session; entries are only ever added (`cache.set`, line 19), never evicted, and the Map is closure-local so no code can clear it. Because `StateManagersProvider.svelte` creates a fresh `StateManagers` on every dashboard mount (routes remount it via `{#key exploreName}`), the cached store permanently pins the FIRST mount's `StateManagers` graph (validSpecStore, timeRangeSummaryStore, selectors, actions, runtimeClient) for every distinct metricsView+explore pair visited — memory grows monotonically per session. Two hazards beyond pure retention: (1) in web-admin, `RuntimeProvider` disposes and evicts its `RuntimeClient` on destroy, so cached stores keep a disposed client whose JWT is never refreshed; its interceptor throws AbortError once the stale JWT nears expiry, so refetches triggered by the pinned observers can fail; (2) the key `metricsName + exploreName` omits instanceId (and a delimiter), so in a web-admin session two projects with identically named metrics view and explore collide, returning a store wired to the first project's runtime client and instanceId. The originally reported visualEditing stale-metrics-name scenario was refuted during verification: the outer derived recomputes the key from the current ctx's live `metricsViewName` writable, and key equality guarantees the pinned ctx's writable holds the same name.

**Manual reproduction:**

1. Observing the leak (Chrome DevTools): start Rill Developer (`rill start`) on a project with several explore dashboards (or open Rill Cloud web-admin for a project with several explores).
2. Open Chrome DevTools > Memory, force a garbage collection, and take a heap snapshot.
3. Visit 5-10 distinct explore dashboards one after another, then navigate back to the home/project page so no dashboard is mounted.
4. Force GC again and take a second heap snapshot. Search the snapshot for closures retaining `createTimeControlStore`, `createStateManagers`, or objects with `validSpecStore`/`timeRangeSummaryStore` properties: one full StateManagers graph is retained per distinct dashboard visited, reachable only from the module-level memoize caches, and the count grows by one for each new dashboard, never shrinking.
5. Observing the disposed-client hazard (Rill Cloud): in web-admin, open an explore dashboard in project A, then navigate to a different project and back after ~15 minutes (past the JWT TTL); watch the Network tab for aborted/failed runtime queries originating from the pinned first-mount observers when a refetch is triggered.
6. Observing the cross-project key collision (Rill Cloud): create two projects that both contain a metrics view and explore with identical names (e.g. a template deployed twice). In one browser session, open the explore in project A, then navigate (client-side) to the same-named explore in project B; inspect Network requests for time-range/spec queries still addressed to project A's instanceId/host.

**Suggested fix:** Key the cache with a delimiter and scope it per-StateManagers (e.g. a WeakMap keyed by ctx, or attach the memoized stores to the object created in `createStateManagers`) so entries die with the dashboard, or add explicit eviction when the dashboard is removed. Include instanceId in the key if the cache remains global.

---

## STAT-10: Canvas `lastVisitedState` keyed by canvas name only, leaking filter state across projects and orgs

**File:** `web-common/src/features/canvas/stores/canvas-entity.ts:599` · **Severity: medium** · **Verdict: CONFIRMED**

The module-level cache `export const lastVisitedState = new Map<string, string>()` (canvas-entity.ts:54) stores each canvas's last-visited URL query string (filters, time range) keyed only by canvas resource name: `saveSnapshot` writes `lastVisitedState.set(this.name, filterState)` on every URL change with params (lines 537, 659-661), and `handleCanvasRedirect` reads it as first priority when a canvas is opened with an empty query string, redirecting via relative `` goto(`?${snapshot}`, { replaceState: true }) `` (lines 597-605). Because web-admin is a SPA with no full-page reload between orgs/projects (only auth flows assign `window.location.href`), the Map survives cross-project navigation, so opening `orgB/projB/canvas/sales` after filtering `orgA/projA/canvas/sales` silently rewrites orgB's URL with orgA's saved params — embedding orgA dimension values (e.g. customer names) in a copyable/shareable URL, and applying them to orgB's data when metrics-view and dimension names coincide (common for templated staging/prod projects). This is the one canvas store missing the namespacing that its siblings have: the canvas registry keys by `${instanceId}::${canvasName}` (state-managers.ts `makeCanvasId`) and explore dashboards namespace state with `${orgName}__${projectName}__`. Scope caveat: the Map is per-tab, in-memory, same-user state, so it is not direct cross-user data exposure; the disclosure path is via URLs the user shares onward, plus a correctness bug of wrong filters silently applied.

**Manual reproduction:**

1. Set up two Rill Cloud projects (e.g. via `rill devtool start cloud`) in the same or different orgs — projA and projB — each containing a canvas dashboard with the same resource name, e.g. `sales.yaml` defining canvas `sales`. For maximum effect, make them copies of the same project (templated staging/prod pattern) so metrics-view and dimension names match.
2. In web-admin, open `/<orgA>/<projA>/canvas/sales` and apply a dimension filter with an identifiable value (e.g. customer = "Acme Corp") and a non-default time range. Confirm the URL now carries search params encoding those values.
3. Without refreshing the page, navigate through the app UI (breadcrumbs, org switcher, project list — any client-side link) to `/<orgB>/<projB>/canvas/sales`, arriving at the canvas URL with no search params.
4. Observe the URL is immediately and silently rewritten (replaceState) to include projA's query string: orgA's filter values ("Acme Corp") now appear in orgB's address bar, ready to be copied/shared with people who have access only to orgB.
5. If the metrics-view and dimension names match across the projects, also observe projB's canvas renders filtered to orgA's values instead of its defaults.
6. Control 1: repeat step 3 but hard-refresh (F5) on the projB canvas URL instead of client-side navigating — the redirect does not happen, proving the in-memory module Map is the carrier.
7. Control 2: repeat the same flow with two explore dashboards of the same name — no cross-project bleed occurs, because explore state is namespaced with `${orgName}__${projectName}__`.

**Suggested fix:** Key `lastVisitedState` by `${instanceId}::${canvasName}` (the same `makeCanvasId` scheme used by the canvas registry), or store the snapshot on the `CanvasEntity` instance instead of a module-level map.

---

## STAT-11: `FilterManager` nested derived stores never unsubscribe inner deriveds; subscriptions leak on spec updates/remounts

**File:** `web-common/src/features/canvas/stores/filter-manager.ts:145` · **Severity: low** · **Verdict: CONFIRMED**

Subscription-lifecycle view of the same pattern as STAT-5. The three nested derived stores (`defaultUIFiltersStore` lines 114-143, `activeUIFiltersStore` 145-180, `filterMapStore` 182-198) create an inner derived and call `.subscribe(set)` without returning the unsubscriber from the outer derived callback, so Svelte never tears the inner subscription down. The outer callback re-runs whenever `metricsViewFilters` emits (`StoreOfStores.set` always creates a new Map, triggered when `updateConfig` registers a `FilterState` for a newly referenced metrics view) and on every inactive-to-active cycle of the outer store (canvas unmount/remount, with the `FilterManager` kept alive by the module-level canvasRegistry in state-managers.ts). Each such event permanently adds one inner subscription per store to the long-lived writables (`pinnedFilterKeysStore`, `allMeasuresStore`, `allDimensionsStore`, each `FilterState.parsed`, etc.). Consequences: every filter mutation runs `convertToUIFilters` N+1 times (one per leaked subscription), growing over an edit/browse session; after a new metrics view is added, stale subscriptions also emit a transient UIFilters/filterMap value computed from the old store list (missing the new view's filters) before the fresh subscription overwrites it in the same synchronous flush — the final value is always correct, so there is no persistent UI incorrectness. Memory is reclaimed only when the `CanvasEntity` is evicted from the registry.

**Manual reproduction:**

1. Not directly user-visible; observe via CPU/computation counts. Setup: start Rill Developer (`rill start`) on a project that has a canvas dashboard backed by at least one metrics view.
2. In the browser DevTools, add a logpoint or `console.count("convertToUIFilters")` breakpoint at the top of `convertToUIFilters` in filter-manager.ts (or temporarily add `console.count` in source).
3. Open the canvas dashboard and toggle any dimension filter value once; note the number of `convertToUIFilters` invocations for that single mutation (baseline, roughly 2-3: one per active nested store).
4. Navigate away from the canvas (e.g. to the project file list or another dashboard) and back to the canvas. Repeat this away/back cycle 5 times; the `CanvasEntity` stays cached in canvasRegistry, and each remount re-runs the outer derived callbacks without cleaning up the previous inner subscriptions.
5. Toggle a single dimension filter value again and observe `convertToUIFilters` now fires roughly 3x(cycles+1) times for one mutation, growing linearly with every remount cycle.
6. Alternative trigger without navigation: in canvas edit mode, add a component that references a metrics view not previously used on this canvas; `updateConfig` registers a new `FilterState`, `metricsViewFilters` emits, and each outer callback stacks another inner subscription. Subsequent filter mutations show the same multiplying `convertToUIFilters` counts, and a breakpoint on the outer store's subscribers shows a transient UIFilters emission that omits the newly added metrics view's filters immediately before the correct emission in the same flush.

**Suggested fix:** Return the inner `.subscribe(set)` unsubscriber from each outer derived callback (same fix as STAT-5).

---

## STAT-12: `timeRangeSummaryStore` derives from the entire `dashboardStore`, recreating its QueryObserver on every mutation

**File:** `web-common/src/features/dashboards/state-managers/state-managers.ts:119` · **Severity: low** · **Verdict: CONFIRMED**

`timeRangeSummaryStore` (state-managers.ts lines 119-139) derives from the entire `dashboardStore` but uses only `selectedTimeDimension` from it. Because metricsExplorerStore reducers mutate the entity in place and re-set the same object (dashboard-stores.ts lines 43-55), every discrete dashboard interaction (filter toggle, leaderboard sort change, pivot interaction, time range change) notifies `dashboardStore`, which tears down and recreates the TanStack QueryObserver for the metrics-view time-range query and emits a fresh QueryObserverResult object even when `selectedTimeDimension` is unchanged. The fresh object identity redundantly re-runs `getDefaultExplorePreset` via `defaultExploreState` (lines 152-164) and adds an extra re-run of each selector readable per interaction. Mitigating factors established during verification: scrub drags only write to a local store during mousemove (the global store is updated once on mouseup), the selector readables depend on `dashboardStore` directly and re-run per mutation anyway, and the unchanged query key with `staleTime: Infinity` means no network refetches occur. The waste is therefore bounded per user interaction: unnecessary observer construction/subscription churn and redundant default-preset recomputation, not continuous per-frame work.

**Manual reproduction:**

1. Run `rill start` on any project with an explore dashboard whose metrics view has a time dimension, and open the explore dashboard in the browser.
2. Open browser DevTools > Sources, find state-managers.ts (via source maps), and add a logpoint (e.g. `console.count('timeRange observer recreated')`) inside the derived callback at the `createQueryServiceMetricsViewTimeRange(` call (~line 124).
3. Interact with the dashboard without touching anything time-related: toggle a dimension filter value on and off several times, change the leaderboard sort, expand/collapse a leaderboard.
4. Observe the logpoint counter increment once per interaction even though `selectedTimeDimension` never changed — each hit is a full teardown and reconstruction of the TanStack QueryObserver plus a fresh result emission that re-runs `getDefaultExplorePreset`.
5. Optionally add a similar logpoint inside the `defaultExploreState` derived callback (~line 154) to confirm `getDefaultExplorePreset` re-executes on each of those interactions.
6. Note in the Network tab that no new MetricsViewTimeRange requests are issued (staleTime: Infinity, unchanged query key), confirming the issue is client-side churn only and not user-visible latency in typical use.

**Suggested fix:** Derive an intermediate `selectedTimeDimension` store (with an equality guard so unchanged values do not emit) and make `timeRangeSummaryStore` depend on that instead of the whole `dashboardStore`.

---

## STAT-13: Legacy proto-state deserialization throws on crafted/corrupt input and is exported without containment

**File:** `web-common/src/features/dashboards/proto-state/fromProto.ts:365` · **Severity: low** · **Verdict: CONFIRMED**

`getDashboardStateFromUrl` (fromProto.ts:82) deserializes legacy proto state with uncontained throw paths (`decodeURIComponent` URIError, invalid base64 in `protoBase64.dec`, malformed protobuf in `DashboardState.fromBinary`), and two live web-admin bookmark code paths invoke it without any containment: the legacy bookmark data transformer used by `parseBookmarks` in ExploreBookmarks.svelte, and `getHomeBookmarkExploreState` used on the explore dashboard page. A single bookmark record whose legacy `data` field is corrupt (only producible via direct admin API calls by an authenticated project member, or storage corruption — the `?state=` URL path is already guarded by `fromLegacyStateUrlParam`'s try/catch, and legacy clients always wrote valid data) throws inside a Svelte reactive/derived computation, breaking the entire bookmarks dropdown or the explore page's home-bookmark load instead of skipping the one bad bookmark. The originally reported `TIME_GRAIN[...].label` TypeError for grain 0 is incorrect — `TIME_GRAIN` has an explicit `TIME_GRAIN_UNSPECIFIED` fallback entry (config.ts:678); only out-of-range enum values in a hand-crafted binary proto hit that TypeError.

**Manual reproduction:**

1. In Rill Cloud, open a project that has an explore dashboard and note the project ID and explore name.
2. As an authenticated project member, create a legacy-format bookmark via the admin API with corrupt data, e.g.: `POST /v1/projects/{projectId}/bookmarks` with body `{ "displayName": "corrupt", "data": "%E0%A4%A", "resourceKind": "rill.runtime.v1.Explore", "resourceName": "<exploreName>", "projectId": "<projectId>" }` — note `data` is set (invalid URI-escape/base64) and `urlSearch` is left empty so the legacy transformer path is taken.
3. Open the explore dashboard in Rill Cloud and click the Bookmarks dropdown in the toolbar.
4. Observe in DevTools console an uncaught URIError ("URI malformed") or base64 decode error thrown from `parseBookmarks` → `exploreBookmarkDataTransformer` → `getDashboardStateFromUrl`, and observe that the bookmarks list fails to render (all bookmarks are affected, not just the corrupt one).
5. Variant: additionally set `"default": true` on the corrupt bookmark to make it the home bookmark, then reload the explore dashboard page; the throw now occurs inside `getHomeBookmarkExploreState`'s derived store during page load.

**Suggested fix:** Wrap the two bookmark call sites (or the exported function itself) in try/catch so a bad bookmark degrades gracefully; guard the `TIME_GRAIN[...]` lookups with a fallback label and skip chips whose grain does not map.

---

## Refuted findings

- **Full state serialization and storage writes on every store mutation, including per-mousemove scrub updates** (`DashboardStateSync.ts:307`) — refuted: scrub drags keep state in `ScrubController`'s private component-level writable and only commit to the dashboard store once on pointer-up (`isScrubbing: false` hardcoded at both call sites), so the serialization/storage pipeline runs at most once per completed gesture, not at mousemove frequency.
