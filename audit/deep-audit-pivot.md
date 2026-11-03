# Deep Audit: Pivot Table

Scope: the pivot table feature under `web-common/src/features/dashboards/pivot` (with its canvas copy and direct consumers where the defect crosses the boundary). Method: a multi-lens find phase (correctness, state management, performance, memory, security) followed by an adversarial verify phase in which every candidate finding was traced end to end through the code before being accepted, corrected, or refuted.

## Summary

| ID | Title | Severity | Category | Verdict |
|----|-------|----------|----------|---------|
| PIVO-1 | Dropped rows in `mergeRowTotalsInOrder` misalign skeleton vs `rowDimensionValues`, assigning cells to wrong rows | high | correctness | CONFIRMED |
| PIVO-2 | `expandedTableMap` cache short-circuit discards newly fetched pages and fresh cell data for the lifetime of a config | high | state | CONFIRMED |
| PIVO-3 | Module-level shared `lastKey` across pivot config store instances spuriously resets `rowPage`, breaking pagination | high | state | CONFIRMED |
| PIVO-4 | Full pipeline teardown/rebuild on every dashboard-store emission, including no-op changes like `activeCell` | medium | performance | CONFIRMED |
| PIVO-5 | Expanded-row assembly re-merges all expanded subtables on every emission of any expansion's query chain | medium | performance | CONFIRMED |
| PIVO-6 | `getPivotConfigKey` omits `searchText`, so cache invalidation and rowPage reset miss search changes | medium | state | CONFIRMED |
| PIVO-7 | Row virtualizer destroyed and recreated on every scroll range change via the `rowScrollOffset` feedback loop | medium | performance | CONFIRMED |
| PIVO-8 | Leaf column filter rebuilds and linearly scans `Object.keys(totals)` once per candidate leaf column | medium | performance | CONFIRMED |
| PIVO-9 | Memoized pivot data stores and their caches are retained forever per metrics-view/explore | medium | memory-leak | CONFIRMED |
| PIVO-10 | Hovering a column header recomputes state for every rendered cell in the non-column-virtualized body | medium | performance | CONFIRMED |
| PIVO-11 | Sort accessor for a missing column dimension throws and breaks the whole pivot | medium | correctness | CONFIRMED |
| PIVO-12 | `RegularTable` writes untrusted strings via `innerHTML`, enabling stored XSS from dimension values | medium | security | CONFIRMED |
| PIVO-13 | `getValuesForExpandedKey` forEach skips missing levels instead of stopping, reading values from the wrong level | low | correctness | CONFIRMED |
| PIVO-14 | Expansion entries for rows hidden under collapsed parents keep spawning queries and assembly work | low | performance | CONFIRMED |

---

## PIVO-1: Dropped rows in `mergeRowTotalsInOrder` misalign skeleton vs `rowDimensionValues`, assigning cells to wrong rows

**File:** `web-common/src/features/dashboards/pivot/pivot-table-transformations.ts:249` · **Severity: high** · **Verdict: CONFIRMED**

When a pivot table has a column dimension, time comparison enabled, and is sorted by a delta measure cell under a specific column value, the sorted row-axes query uses a FULL OUTER comparison join (returns dimension values present in either period) while the follow-up unsorted measure-totals query defaults to sorting by the base measure, which the backend rewrites to a LEFT join returning only base-period values. `mergeRowTotalsInOrder` drops the comparison-only values from the merged totals, so the table skeleton has fewer rows than the unfiltered `rowDimensionValues` list that is still passed to `reduceTableCellDataIntoRows` (both in the base table via `pivot-data-store.ts:446` and in the expansion path via `pivot-expansion.ts:478`/`515`). Because cells are indexed by position in the unfiltered list but written into the filtered skeleton (`tableData[rowOffset + rowIndex]`), every cell after the dropped value is written into the row labeled with the following dimension value, and the last row's cells are silently discarded — the table displays measure values under the wrong dimension rows.

**Manual reproduction**

1. Prepare a project whose data has a dimension value with events only in the comparison period (e.g. a customer that churned: rows exist for last month but none for the current month).
2. Open the explore dashboard in Rill Developer and switch to the Pivot tab.
3. Add a row dimension (e.g. customer), a column dimension (e.g. country), and a measure (e.g. revenue).
4. Select a time range covering only the current period and enable time comparison (previous period), so each measure shows current, delta, and delta % sub-columns.
5. Click the delta sub-column header under one specific column-dimension value to sort by it, and toggle to ascending so churned values (large negative delta) rank near the top of the row order.
6. Observe the rows below the churned dimension value: their per-column measure cells belong to the row above (each subsequent row shows the previous row's numbers), and the bottom-most affected row shows empty or skeleton-total values.
7. To confirm a specific cell is wrong, apply the same dimension value + column value as filters in the explore view and compare the number against the pivot cell; the pivot cell matches the neighboring row's value instead.

**Suggested fix:** When `mergeRowTotalsInOrder` drops a value, also drop it from the `rowDimensionValues` used for `reduceTableCellDataIntoRows` and the cell query filter — e.g. return both the ordered totals and the corresponding filtered value list, and thread that pair through `assembleBasePivotData`/`addExpandedDataToPivot`.

---

## PIVO-2: `expandedTableMap` cache short-circuit discards newly fetched pages and fresh cell data for the lifetime of a config

**File:** `web-common/src/features/dashboards/pivot/pivot-data-assembly.ts:89` · **Severity: high** · **Verdict: CONFIRMED**

Once any pivot row is expanded, `cacheExpandedPivotData` stores the assembled table under a config key that excludes `rowPage` and expanded state, and both stage 3 of `pivot-data-store.ts` (lines 422-436) and `assembleBasePivotData` (`pivot-data-assembly.ts:89-95`) short-circuit to that cached snapshot whenever the key matches. Consequences: (a) infinite scroll silently breaks — incrementing `rowPage` fetches page-2+ row axes and appends them to the skeleton, but the skeleton is discarded in favor of the cached page-1 table, while `getReachedEndForRowData` (computed from the fresh axes) keeps returning false, so further pages are repeatedly fetched and discarded; (b) any refetch of `tableCellQuery` (e.g. query invalidation after a model rebuild in Rill Developer) is ignored for the lifetime of the config key, so top-level body cells show the stale snapshot while the totals row and expanded subRows (which are re-queried each derivation) update, producing visibly inconsistent data; (c) collapsing all rows never clears the entry (`cacheExpandedPivotData` only runs when `expandedRowMeasureValues` is non-empty and nothing else deletes it), so recovery requires changing sort/filter/time/measures or reloading. The comment at `pivot-data-store.ts:420` shows the intent was only to avoid a loading flash during refetch, but the check is key presence rather than `isFetching`, so completed fresh responses are also discarded. The behavior predates the #9538 refactor.

**Manual reproduction**

1. Start Rill Developer (`rill start`) on a project whose metrics view has at least two dimensions and one measure, where the outer dimension has more than 50 distinct values; open the explore dashboard and switch to the Pivot tab.
2. Add two dimensions as Rows (the high-cardinality one outermost) and one measure as a column; leave the table in nested (non-flat) mode.
3. Baseline: without expanding anything, scroll to the bottom of the table and confirm rows beyond the first 50 load via infinite scroll.
4. Change the sort (click a column header) to reset state, scroll back to the top, then expand any outer row with its chevron.
5. Scroll to the bottom again: the table now stops at 50 outer rows and no new rows ever render; in DevTools Network tab, observe repeated metrics-view aggregation requests with offset=50 (and higher on further scrolling) whose results are never displayed.
6. Stale-data variant: with the row still expanded, edit the underlying model SQL (e.g. change the measure's source values) and save so the model rebuilds and queries invalidate; observe that the totals row and the expanded subRows update while the top-level body cells keep the old values.
7. Collapse the expanded row: the body remains stale and pagination remains broken; changing sort, filters, or time range (which changes the config key) restores correct behavior.

**Suggested fix:** Key the expanded cache by configKey + rowPage (or merge fresh pivotSkeleton/cellData with the cached expanded subRows instead of returning the cached table wholesale), and delete the cache entry when the expanded state becomes empty or when the cell query returns fresh data.

---

## PIVO-3: Module-level shared `lastKey` across pivot config store instances spuriously resets `rowPage`, breaking pagination

**File:** `web-common/src/features/dashboards/pivot/pivot-data-config.ts:25` · **Severity: high** · **Verdict: CONFIRMED**

Both `web-common/src/features/canvas/components/pivot/util.ts` (line 88) and `pivot-data-config.ts` (line 25) keep the pagination-reset guard `lastKey` in a module-level variable shared across all pivot config store instances. On canvas dashboards, every table/pivot component creates its own `createPivotConfig` derived store (one per `PivotCanvasComponent`), all live at once, and `getPivotConfigKey` excludes `rowPage` while infinite scroll updates `rowPage` through the derived store's own `pivotState` input. With two or more table/pivot components on one canvas, a scroll-triggered recomputation of table A sees `lastKey` holding table B's key, misreads its own unchanged config as changed, and resets A's `rowPage` to 1 — so infinite scroll cannot advance past the first page (50 rows) on any table that is not the most recent writer of `lastKey`, and interleaved scrolling breaks all of them. The explore-side copy in `pivot-data-config.ts` is effectively benign within a single explore (all live instances compute identical keys) but causes a spurious rowPage reset when navigating between different explore dashboards. The reset block also mutates the store-owned `config.pivot` object in place and dispatches a store update synchronously inside the derived computation; these are sloppy but self-converging.

**Manual reproduction**

1. In Rill Developer, open a project with a metrics view whose leaf-level row count exceeds 50 rows (e.g. a dimension with many distinct values), and create a canvas dashboard containing two table components (or a table and a pivot) with different column sets, e.g. table A on dimension X and table B on dimension Y.
2. Open the canvas preview so both tables render at the same time, and open browser DevTools > Network filtered to "aggregation" requests.
3. Scroll the FIRST table (the one added/rendered first) to the bottom to trigger infinite scroll.
4. Observe the bug: an aggregation request for offset 50 (page 2) is immediately followed by a refetch at offset 0, and the table remains stuck at the first ~50 rows; rows beyond page 1 never render no matter how many times you scroll.
5. Now scroll the SECOND table to the bottom; after step 4 has run once, it exhibits the same reset, so neither table can paginate past the first page.
6. Control: remove one of the two tables from the canvas and repeat step 3; infinite scroll now works, confirming the cross-component interference.

**Suggested fix:** Make `lastKey` per-instance (closure of `createPivotConfig`/`getPivotConfig`) or include a per-component identity in the key; avoid mutating `config.pivot` in place — perform the reset via an explicit subscription or store action outside the derived body.

---

## PIVO-4: Full pipeline teardown/rebuild on every dashboard-store emission, including no-op changes like `activeCell`

**File:** `web-common/src/features/dashboards/pivot/pivot-data-store.ts:142` · **Severity: medium** · **Verdict: CONFIRMED**

Every emission of the explore dashboard store — including query-irrelevant changes such as `setPivotActiveCell` fired by a plain cell click in Rill Developer, where the data viewer is always enabled — produces a fresh `PivotDataStoreConfig` from `getPivotConfig` (no equality or memo guard), which causes `derivedSwitch` in `createPivotDataStore` to tear down and synchronously rebuild the entire four-stage pivot pipeline: all TanStack query observers are recreated, the full column-def tree is rebuilt (`sliceColumnAxesDataForDef` + `getColumnDefForPivot`), cell data is re-reduced into rows and the whole table is re-cloned via `structuredClone` (`pivot-data-assembly.ts:113`), and expansion/final-assembly re-runs. Because `refetchOnMount` is false the recreated queries hit the cache (no network refetches), so the damage is bounded to main-thread CPU work proportional to table size plus a full table re-render from new data/columnDef references — per-click jank that grows to multi-frame stalls only on large pivots. `getPivotConfigKey` already identifies query-relevant config but is only used for the expanded-data cache and rowPage reset, never to skip rebuilds.

**Manual reproduction**

1. Start Rill Developer (`rill start`) on a project with a metrics view that has several dimensions and measures, and open an explore dashboard (in Rill Developer, `canShowDataViewer` is always true, so cell clicks set the active cell).
2. Open the Pivot tab and build a moderately large pivot: 1-2 row dimensions, 1-2 column dimensions, 2-3 measures; enable time comparison for more accessor columns; expand a few rows.
3. Open Chrome DevTools > Performance and start recording; also open the Network tab filtered to metrics-view aggregation requests.
4. Click any measure cell in the table body (this only sets `pivot.activeCell` and opens the row viewer; no queried data changes).
5. In the Performance recording, observe a long synchronous task on each click containing `getColumnDefForPivot`, `reduceTableCellDataIntoRows`, `structuredClone`, `prepareNestedPivotData`, and `buildFinalPivotStateDetails` frames, plus a full table re-render; note the Network tab shows no new aggregation requests, proving the entire rebuild is redundant CPU work.
6. To feel it as a user: build a large pivot (hundreds of visible rows, 2 column dimensions x 3 measures with comparison) and click cells repeatedly; each click lags noticeably even though the displayed data is unchanged. Alternatively, add a temporary `console.count` in `createColumnAxesStage` (`pivot-data-store.ts:169`) and watch it increment on every cell click and on any unrelated explore-state mutation.

**Suggested fix:** Dedupe config emissions before they reach `derivedSwitch` (e.g. compare `getPivotConfigKey(config)` plus rowPage/expanded and skip emission when unchanged), or split query-irrelevant fields (activeCell) out of the config store; memoize `getColumnDefForPivot` and the base-assembly result on `configKey` + input identities.

---

## PIVO-5: Expanded-row assembly re-merges all expanded subtables on every emission of any expansion's query chain

**File:** `web-common/src/features/dashboards/pivot/pivot-data-store.ts:488` · **Severity: medium** · **Verdict: CONFIRMED**

Stage 4 of the pivot pipeline (`pivot-data-store.ts:488`) re-assembles all expanded rows on every emission of any expansion's query chain. `queryExpandedRowMeasureValues` creates one nested query chain per expanded key, and each chain emits 3-4+ times as its axis, totals, and subtable queries resolve; every emission re-runs `addExpandedDataToPivot`, which loops over ALL E expansions and re-runs `reduceTableCellDataIntoRows` (with per-cell `getAccessorForCell` string building) on each expansion's subtable — up to 5000 rows per expansion when column dimensions are present — with no per-expansion memoization (`cacheExpandedPivotData` only caches the whole assembled table for stage-3 reuse). Toggling any expansion also rebuilds the entire pipeline via the config store, so each new expansion click costs O(E x cells x emissions) of redundant reduction work plus a tanstack-table model rebuild and re-render per emission; lag grows roughly linearly with the number of expanded rows per click (O(E^2) cumulatively), and is most noticeable with column dimensions, several measures, and high-cardinality nested dimensions.

**Manual reproduction**

1. Start Rill Developer with a project whose metrics view has at least 2 dimensions usable as row dimensions, 2 dimensions usable as column dimensions, 2-3 measures, and high cardinality (thousands of distinct values in the nested row dimension).
2. Open the explore dashboard and switch to the Pivot tab.
3. Configure the pivot: 2 row dimensions (high-cardinality dimension as the second/nested one), 2 column dimensions, and 2-3 measures. Column dimensions are required to hit the 5000-row subtable path.
4. Open Chrome DevTools > Performance and start recording.
5. Expand outer rows one at a time (click the expand chevron on 10-15 different rows), pausing until each finishes loading before expanding the next.
6. Observe that each successive expansion click takes longer to settle than the previous one: the Performance flame chart shows, for a single click, multiple repeated long tasks containing `addExpandedDataToPivot` / `reduceTableCellDataIntoRows` / `getAccessorForCell` frames (one full pass per query-chain emission, each pass re-reducing every already-expanded subtable), followed by repeated tanstack table model rebuilds.
7. As a control, collapse all rows and expand a single row on a fresh pivot: the same profile shows only one small reduction pass, confirming the growth is caused by re-reducing all prior expansions on every emission.

**Suggested fix:** Memoize the reduced subtable per `expandIndex` (keyed on the query data identity) so unchanged expansions are not re-reduced, and/or debounce/batch emissions so one assembly pass runs per settled state instead of per chain emission.

---

## PIVO-6: `getPivotConfigKey` omits `searchText`, so cache invalidation and rowPage reset miss search changes

**File:** `web-common/src/features/dashboards/pivot/pivot-utils.ts:100` · **Severity: medium** · **Verdict: CONFIRMED**

`getPivotConfigKey` (`pivot-utils.ts:72-101`) omits `config.searchText` even though the search text changes the pivot's row-axis queries (`createPivotBaseQueryPlan` applies `getDimensionFilterWithSearch` to the anchor dimension). `searchText` flows into the pivot config from the global `dimensionSearchText` writable, which the TDD search box writes and never resets on unmount, so it can change between config emissions while every field that IS in the key stays identical. Two consequences: (a) the expanded-row cache (`cache.expandedTableMap`, keyed by this config key, persistent across tab switches via the memoized `usePivotForExplore` store) is served by `assembleBasePivotData` for the same key, so after a search change the pivot displays the previously cached, differently-filtered row set instead of rows matching the current search; (b) the `lastKey !== currentKey` rowPage reset in `getPivotConfig` does not fire, so a user paginated past page 1 keeps their old offset against a now-smaller filtered result set and `getPivotSkeletonForPage` serves stale `cache.lastPivotData`. Editing the search requires being on the TDD view, so the bug manifests on TDD-search + Pivot-tab round trips rather than while the pivot is on screen, which limits but does not eliminate exposure.

**Manual reproduction**

1. Start Rill Developer with an explore dashboard whose metrics view has at least 2 dimensions and 1 measure, and enough cardinality that a search string matches only a subset of values of the first dimension.
2. In the Explore tab, click a measure to open the Time Dimension Detail (TDD) view, and select a comparison dimension.
3. Click the search icon in the TDD header and type a search string that matches a strict subset of the dimension's values (e.g. "goog"). Leave the search bar open; do NOT click the X (the X is the only thing that clears the shared search store).
4. Switch to the Pivot tab. Add two row dimensions, making the searched dimension the first (outermost) row dimension, and add a measure to columns. Observe the outer rows are filtered by the TDD search text; this confirms the search is active in the pivot.
5. Expand one outer row with the chevron and wait for the nested data to load (this populates the expansion cache under a config key that omits the search text).
6. Switch back to the Explore tab (it returns to the TDD view) and change the search text, e.g. clear it entirely or type a different string that matches other values. Keep the search bar open.
7. Switch to the Pivot tab again. Observed bug: the pivot still shows exactly the outer row set from step 5 (old search) instead of rows matching the new search text; DevTools network tab shows fresh metrics-view-aggregation requests carrying the new search filter whose results are not reflected in the displayed outer rows.
8. Pagination variant: repeat steps 1-4 but skip expanding; instead scroll the pivot until more than one page of rows has loaded (rowPage > 1, 50 rows per page). Then do steps 6-7 with a new search matching fewer than 50 values. Observed: the pivot does not reset to page 1; it keeps serving the previous unfiltered rows (or an empty tail page) because the row offset points past the end of the filtered result set.

**Suggested fix:** Include `searchText` (and, for the expansion cache, any other query-affecting fields such as `nestedRowLimits`) in `getPivotConfigKey`, or maintain a separate cache key for `expandedTableMap` that covers everything the queries depend on.

---

## PIVO-7: Row virtualizer destroyed and recreated on every scroll range change via the `rowScrollOffset` feedback loop

**File:** `web-common/src/features/dashboards/pivot/PivotTable.svelte:164` · **Severity: medium** · **Verdict: CONFIRMED**

In `PivotTable.svelte`, `$: virtualizer = createVirtualizer({ count: rows.length, ..., initialOffset: rowScrollOffset, rangeExtractor: ... })` (line 148) has `rowScrollOffset` and `stickyRows` as reactive dependencies, and line 164 feeds the virtualizer's own store output back into the first: `$: rowScrollOffset = $virtualizer?.scrollOffset || 0`. @tanstack/svelte-virtual's store notifies whenever the visible range or `isScrolling` changes (scroll start, every ~24px of scroll at ROW_HEIGHT=24, and scroll end), and each notification carries a changed `scrollOffset`, so each one triggers exactly one full virtualizer recreation: `new Virtualizer(...)`, unsubscribe/cleanup of the old instance (scroll listener removal, ResizeObserver disconnect), re-attachment of observers, a forced `getBoundingClientRect` read, and loss of all per-instance memoization. The loop is bounded (it converges after one recreation per notification because `getScrollOffset()` materializes `scrollOffset = initialOffset` during mount, making the fed-back value equal), so there is no infinite loop or scroll-position jump — but a single scroll gesture through a large pivot recreates the virtualizer dozens of times, and when a totals row is present, `$: stickyRows = totalsRow ? [0] : []` additionally produces a new array identity on every pivotDataStore emission, recreating the virtualizer during data fetch/assembly churn as well.

**Manual reproduction**

1. Start Rill Developer (`rill start`) on a project with a metrics view, open an Explore dashboard, and switch to the Pivot tab.
2. Configure a pivot with at least one row dimension and a couple of measures; expand several nested rows and scroll down to trigger infinite-scroll page loads until the table has on the order of 1000+ rows.
3. Open Chrome DevTools > Sources, find `@tanstack/svelte-virtual/dist/index.js` in the bundle (or the dev-served module), and set a logpoint (e.g. `console.count('virtualizer created')`) on the `new Virtualizer(initialOptions)` line inside `createVirtualizerBase`.
4. Scroll the pivot table continuously for a few seconds. Observe the counter incrementing roughly once per ~24px of scroll (dozens of constructions per gesture) instead of the expected once per mount.
5. For the impact profile: open DevTools > Performance, record while scrolling, and observe repeated ResizeObserver setup/teardown, forced `getBoundingClientRect` reflow reads, and `Virtualizer` constructor work interleaved with scroll handling on the main thread.
6. To observe the secondary trigger: with a totals row visible, change a filter or expand a row so `pivotDataStore` re-emits during fetching, and note the constructor logpoint fires on each store emission even without scrolling.

**Suggested fix:** Create the virtualizer once and push changes via `$virtualizer.setOptions({ count: rows.length, ... })` (the svelte adapter exposes a merging `setOptions` on the store value); with no recreation, the `rowScrollOffset` feedback becomes unnecessary and can be removed.

---

## PIVO-8: Leaf column filter rebuilds and linearly scans `Object.keys(totals)` once per candidate leaf column

**File:** `web-common/src/features/dashboards/pivot/pivot-column-definition.ts:90` · **Severity: medium** · **Verdict: CONFIRMED**

In `createNestedColumns`'s base case (`pivot-column-definition.ts:90-92`), `Object.keys(totals)` is re-evaluated inside the `filter` callback, so every candidate leaf column allocates a fresh K-element key array and scans it linearly with `includes`. Candidate leaves are the cross product of the sliced column-dimension axis values times the measure-name count (base + delta + percent per measure with comparison), and K is the totals-row key count (totals query limit 300 combos x up to 3 names per measure, e.g. ~2700 with 3 measures and comparison). With 2 column dimensions whose sliced axes reach ~50 values each, ~22,500 candidate leaves each allocate and scan a ~2700-entry array — tens of millions of string comparisons — synchronously inside `getColumnDefForPivot`, which re-runs on every stage-3 pipeline emission (`createCellDataStage`), causing main-thread jank on every pivot interaction when column cardinality is high.

**Manual reproduction**

1. In Rill Developer, open a project with a metrics view that has at least two high-cardinality dimensions (50+ distinct values each) and at least 3 measures.
2. Open the explore dashboard, enable time comparison (e.g. select a time range and turn on "Comparing"), then switch to the Pivot tab.
3. Add both high-cardinality dimensions as column dimensions, add all 3 measures as columns, and add one dimension as a row dimension.
4. Open Chrome DevTools > Performance and start recording, then change any pivot input that rebuilds the table (e.g. toggle a measure, change sort, or page columns with the horizontal scroll / next-page control).
5. In the recorded profile, look for a long synchronous task attributed to `createNestedColumns` / `getColumnDefForPivot` (via `createCellDataStage`); the self-time is dominated by `Array.prototype.includes` and `Object.keys`, and the table visibly stutters or freezes while the column definitions rebuild.
6. To confirm the fix, hoist `Object.keys(totals)` into a Set outside the recursion and repeat step 4; the long task shrinks dramatically.

**Suggested fix:** Hoist `const totalKeys = new Set(Object.keys(totals));` out of the recursion and use `totalKeys.has(leaf.accessorKey)`.

---

## PIVO-9: Memoized pivot data stores and their caches are retained forever per metrics-view/explore

**File:** `web-common/src/features/dashboards/pivot/pivot-data-store.ts:699` · **Severity: medium** · **Verdict: CONFIRMED**

`usePivotForExplore` (`pivot-data-store.ts:699`) is wrapped in `memoizeMetricsStore`, whose module-level `Map<string, Store>` is populated on first subscription per `metricsViewName + exploreName` key and never evicted; the derived teardown only unsubscribes the inner store. Each cached `PivotDataStore` closes over the `PivotDataCache` created at `pivot-data-store.ts:140`, and `updatePivotDataCache` writes the fully assembled table (`lastPivotData`), the full `lastPivotColumnDef`, and the most recent expanded-row snapshot (`expandedTableMap`, which holds exactly one config key at a time) into it after every render. The cached store also permanently retains the first StateManagers context (runtimeClient, queryClient, config store chain) it was created with. Net effect: in a long-lived SPA session (web-admin especially), every dashboard whose pivot view or rows viewer was opened leaves its last assembled pivot table pinned in memory for the lifetime of the tab, even after navigating away; memory grows with the number of distinct dashboards visited and GC can never reclaim it.

**Manual reproduction**

1. Start Rill Developer (or open Rill Cloud) with a project containing several explore dashboards over reasonably large data.
2. Open an explore dashboard, switch to the Pivot tab, and build a large pivot (e.g. a high-cardinality row dimension, a column dimension, several measures); expand a few rows and scroll to load more row pages.
3. Navigate away from the pivot (back to Explore view, then to a different dashboard entirely).
4. In Chrome DevTools > Memory, take a heap snapshot and search for `lastPivotData` or `PivotDataCache`: the full array of assembled PivotDataRow objects for the abandoned pivot is still retained, reachable via the module-level Map inside `memoizeMetricsStore`.
5. Repeat steps 2-3 across N different dashboards (distinct metrics views), taking a heap snapshot after each; observe one retained `PivotDataCache` per dashboard with its `lastPivotData` array, and total retained size growing monotonically with N — it never drops after navigation, only on full page reload.

**Suggested fix:** Evict cache entries when the last subscriber unsubscribes (return a cleanup from the derived start function that removes the map entry), or at minimum clear `cache.lastPivotData`/`expandedTableMap` when the store has no subscribers. The same pattern applies to `useTimeSeriesDataStore`, `useTimeControlStore`, and `useDimensionTableData`, though pivot payloads are the largest.

---

## PIVO-10: Hovering a column header recomputes state for every rendered cell in the non-column-virtualized body

**File:** `web-common/src/features/dashboards/pivot/NestedTable.svelte:592` · **Severity: medium** · **Verdict: CONFIRMED**

In `NestedTable.svelte`, every rendered `<td>` computes `nestedCellState({..., inHoveredCol: isHeaderInHoveredRange(i, 1, hoveredColRange), ...})` as an `{@const}` feeding 13 class bindings, making the header-hover state `hoveredColRange` a reactive dependency of every cell. `hoveredColRange` is reassigned (fresh object) on mouseenter and nulled on mouseleave of each column-dimension `<th>`, so each header-to-header cursor transition triggers two full passes that re-run `nestedCellState` and re-diff the class bindings for every rendered cell. Rows are virtualized with `overscan={60}` (~120-160 DOM rows) but columns are not: a page renders up to 150 leaf columns with comparison enabled (`getColumnsPerPage` = 50 x 3), and column pages accumulate as the user scrolls right (`sliceColumnAxesDataForDef` slices to `columnsPerPage * columnPage`), so a wide pivot holds 20,000+ `<td>` nodes. Sweeping the cursor along the header row therefore causes repeated multi-thousand-node update passes and frame drops. (The `O(H^2)` `colStart` reduce in the header row is a hover-independent derived and only recomputes when headerGroups changes; it is negligible.)

**Manual reproduction**

1. Start Rill Developer (`rill start`) on a project with a metrics view that has at least two moderately high-cardinality dimensions and a time dimension.
2. Open an explore dashboard for that metrics view and switch to the Pivot tab.
3. Configure a wide pivot: drag two dimensions into Columns, one dimension into Rows, add one or more measures, and enable a time comparison (e.g. select "Last 7 days" with comparison on) so each measure renders 3 context columns (150 leaf columns per page).
4. Scroll the table horizontally to the right until additional column pages load (columns accumulate), and expand a few rows so the body is dense.
5. Open Chrome DevTools > Performance and start recording.
6. Sweep the mouse cursor left-to-right along the column dimension header row (the second-to-last header row), crossing several headers per second.
7. Stop recording and observe: each header mouseenter/mouseleave produces long scripting tasks (Svelte template effect re-runs across all rendered cells), dropped frames, and visibly laggy hover highlighting; compare with sweeping over the measure header row of a small pivot, which stays at 60fps.

**Suggested fix:** Drive hover highlighting via CSS (a data attribute on the hovered column range plus stylesheet rules) or scope the reactive dependency so only cells in the affected column indices update.

---

## PIVO-11: Sort accessor for a missing column dimension throws and breaks the whole pivot

**File:** `web-common/src/features/dashboards/pivot/pivot-utils.ts:361` · **Severity: medium** · **Verdict: CONFIRMED**

`getColumnDimEntriesFromAccessor` (`pivot-utils.ts:358-361`) decodes a minimized sort accessor without validating the decoded column-dimension index: `colDimensionNames[c]` yields undefined for an out-of-range or NaN `c`, so `columnDimensionAxes[undefined][v]` throws TypeError. It is reached via `getSortForAccessor` -> `getColumnFiltersFromMinimizedAccessor` for any `config.pivot.sorting[0].id` that is neither the first row dimension nor a measure name. The `sort_by` URL param is deserialized without validation (`convertURLToExplorePreset.ts:629-632` — unlike pivot rows/cols, which are filtered against current dimensions in the same function), and `convertPresetToExploreState.ts:491-497` deliberately passes unrecognized sort IDs through on the incorrect assumption that TanStack Table will ignore them. Because column-cell sorts serialize accessors like `"c1v0m0"` verbatim into shared URLs, a bookmarked/shared pivot URL whose column dimension was later removed from the metrics view (or any hand-edited/garbage `sort_by` value) crashes uncaught inside the `derivedSwitch` callback in `pivot-data-store.ts:197`, leaving the pivot view permanently broken instead of degrading gracefully. Out-of-range `v` with valid `c` does not crash; the trigger is specifically an invalid `c` (or non-accessor garbage that parses to NaN).

**Manual reproduction**

1. Start Rill Developer (`rill start`) on any project with an explore dashboard that has at least one dimension and one measure.
2. Open the explore dashboard and switch to the Pivot view; add one row dimension and one measure (column dimensions optional).
3. In the browser address bar, the URL now contains pivot params (view=pivot, rows=..., cols=..., sort_by=...). Edit the URL to set `sort_by=c9v0m0` (an accessor referencing a non-existent 10th column dimension) and press Enter to load it. (`sort_by=foo` also triggers it via the NaN path.)
4. Observe the pivot table fails to render (stuck in loading/blank state) and the DevTools console shows `TypeError: Cannot read properties of undefined` originating from `getColumnDimEntriesFromAccessor` in `pivot-utils.ts` (line 361).
5. Natural stale-state variant (no URL crafting): add two column dimensions to the pivot, click a nested column header cell to sort by it (URL gets `sort_by=c1v0m0`), copy the URL, remove the second column dimension from the metrics view YAML, then load the copied URL — pivot cols are filtered to the surviving dimension on load but `sort_by` is kept, producing the same crash.

**Suggested fix:** After decoding, bail out (return `[]`) when `name` is undefined or `columnDimensionAxes[name]` is missing, mirroring the existing `accessorParts[0] === ""` guard; validate `config.pivot.sorting[0].id` against current dimensions on state load.

---

## PIVO-12: `RegularTable` writes untrusted strings via `innerHTML`, enabling stored XSS from dimension values

**File:** `web-common/src/features/dashboards/pivot/RegularTable.svelte:167` · **Severity: medium** · **Verdict: CONFIRMED**

`TDDTable.svelte`'s `renderRowHeader` interpolates the raw dimension value (`value.value`, from OLAP source data) into an HTML string without escaping (line 262, `${total}`), and `RegularTable.svelte` assigns that string to `th.innerHTML` (line 138). The `escapeHtml` helper present in the file is applied only to the link href/title, not to the dimension text. A stored dimension value such as `<img src=x onerror=...>` is therefore written verbatim into the DOM. In Rill Developer (web-local) there is no CSP (the csp block is commented out), so the injected inline handler executes arbitrary JavaScript in the app origin, giving access to the local runtime API. In Rill Cloud (web-admin) the hash-mode `script-src` CSP without `unsafe-inline` blocks inline event-handler execution, reducing the impact there to HTML injection rather than script execution. The secondary claim about `value.spark` is not exploitable (spark is internally generated SVG, not user data); the dimension-value vector alone confirms the issue. Severity remains medium: exploitation requires control over source dimension data (frequently trusted in BI contexts) and the primary JS-execution impact is limited to Rill Developer, while Cloud is CSP-mitigated.

**Manual reproduction**

1. In Rill Developer (web-local, which has no CSP), create or edit a model/source so that a dimension column contains a row whose value is the string `<img src=x onerror=alert(document.cookie)>` (any dimension with attacker-controlled or externally-ingested text values works).
2. Create a metrics view over that source with the dimension exposed, and open an explore dashboard on it.
3. Click a measure to expand the Time Dimension Detail (TDD) table, then set the comparison to the dimension that contains the malicious value so its values render as left-column row headers.
4. When the row header for that value is drawn, `RegularTable` assigns the returned HTML to `th.innerHTML`; the injected `<img>` triggers its onerror handler and the script executes in the Rill Developer origin (observe the alert, or replace with a cookie/localStorage exfiltration payload).
5. To observe the Cloud (web-admin) behavior instead: repeat on Rill Cloud and note the inline onerror is blocked by the hash-mode CSP `script-src`, but the raw HTML is still injected (verify in DevTools Elements that the unescaped `<img>` node is present in the th), demonstrating HTML injection even where JS execution is mitigated.

**Suggested fix:** Escape all data-derived text before assigning to `innerHTML` inside `RegularTable` (or build DOM nodes with `textContent`), and HTML-escape interpolated dimension values in every `renderRowHeader`/`renderCell` caller.

---

## PIVO-13: `getValuesForExpandedKey` forEach skips missing levels instead of stopping, reading values from the wrong level

**File:** `web-common/src/features/dashboards/pivot/pivot-expansion.ts:75` · **Severity: low** · **Verdict: CONFIRMED**

`getValuesForExpandedKey` (`pivot-expansion.ts:75-81`) uses `forEach` with an early `return` to handle a missing index, which skips only that level while leaving `currentValue` pointing at the stale ancestor array; subsequent levels then index into the wrong nesting level. In practice the misread value is always `undefined` (rows do not carry other levels' dimension keys), and the `values.some((v) => v === undefined ...)` guard in `queryExpandedRowMeasureValues` short-circuits to the same `isFetching: true` no-op state that correct traversal (empty result) would produce, so the cited scenario (stale expanded key after `setPivotOutermostRowLimit`, which indeed does not clear `expanded`) has no observable difference from correct behavior. The `getFiltersForCell`/data-viewer impact does not affect visible cells, since visible row ids resolve fully and all row-shrinking reducers clear `activeCell`; it requires an external data refresh shrinking top-level rows with 3+ row dimensions and an active sub-row cell. The defect is a genuine latent traversal bug worth fixing for robustness, but a fix that breaks and returns partial values would be worse: partial values bypass the undefined guard and fire sub-table queries missing intermediate dimension filters, and `addExpandedDataToPivot` could attach the results to the wrong parent row. The fix should return an empty array (or a sentinel) as soon as a level fails to resolve.

**Manual reproduction**

1. The defect has no user-visible effect in the reachable scenarios (downstream guards produce the same isFetching no-op as correct traversal would). To observe the wrong-level read itself in Rill Developer:
2. Start a project with a metrics view that has at least 3 dimensions, and open its explore dashboard's Pivot tab.
3. Add 3 dimensions as pivot rows and at least one measure as a column; keep the totals row enabled.
4. Expand the outermost row at visible index 7, then expand its first sub-row, so `pivot.expanded` contains a key like "7.0".
5. Reduce the outermost row limit below 7 (the outermost Show more/fewer control, which calls `setPivotOutermostRowLimit` and does not clear `expanded`).
6. In browser DevTools, set a breakpoint at the end of `getValuesForExpandedKey` (`pivot-expansion.ts` line 82) and observe that for key "7.0" it returns `[undefined]` produced by reading `tableData[0][rowDimensions[1]]` from the top-level array (wrong level), instead of an empty array.
7. Observe in the UI that nothing visibly breaks: the stale expansion row is not rendered, the table finishes loading, and no query with a wrong filter is issued (Network tab shows no aggregation request scoped to the stale key).

**Suggested fix:** Use a for loop and abort with an empty array (or sentinel) as soon as `currentValue?.[index]` is missing, so later levels never index into an ancestor's array; do not return partial values.

---

## PIVO-14: Expansion entries for rows hidden under collapsed parents keep spawning queries and assembly work

**File:** `web-common/src/features/dashboards/pivot/pivot-expansion.ts:203` · **Severity: low** · **Verdict: CONFIRMED**

TanStack's `toggleExpanded` removes only the collapsed row's own id from the expanded map, and `setPivotExpanded` (`dashboard-stores.ts:323`) stores the result without pruning descendants, so after collapsing a parent its expanded descendants (e.g. "1.0", "1.1") remain in `config.pivot.expanded` indefinitely. Because `getPivotConfigKey` excludes `expanded`, stage 3 keeps serving `cache.expandedTableMap[configKey]` whose collapsed parent still has populated `subRows`, so `getValuesForExpandedKey` resolves real dimension values for the hidden keys and `queryExpandedRowMeasureValues` (`pivot-expansion.ts:202`) rebuilds a chain of up to 3 aggregation-query observers per hidden key on every store re-derivation (new expansion, cell click, page fetch), and `addExpandedDataToPivot` re-merges each hidden key's subtable (up to 5000 cells) into rows the table never renders. With `refetchOnMount: false` in the global QueryClient the recreated observers do not refire network requests while the cache is warm, so the steady-state cost is CPU (observer/derived-store churn and invisible-subtable reassembly on every assembly pass) and retained query-cache memory; hidden-key queries do hit the network again whenever the cache is invalidated (e.g. a model or source refresh in Rill Developer), since their observers are kept alive forever.

**Manual reproduction**

1. In Rill Developer, open an explore dashboard, switch to the Pivot tab, and add at least 2 row dimensions (e.g. dim_a, dim_b) and one measure.
2. Expand a top-level row (e.g. index "1"), then expand several of its children (creating expanded keys like "1.0", "1.1", ...).
3. Collapse the parent row. Its children are no longer visible.
4. Proof the stale keys persist: re-expand the same parent; its children reappear already expanded without any click, because their ids were never removed from `pivot.expanded` (also visible in the URL/session state for the explore).
5. Proof of recurring hidden work: with the parent still collapsed, open DevTools > Performance and record while clicking a cell or expanding a different top-level row. The flame chart shows `queryExpandedRowMeasureValues` / `addExpandedDataToPivot` / `reduceTableCellDataIntoRows` frames processing an entry for each hidden child key on every interaction, even though those rows are not rendered.
6. Proof of recurring network cost on invalidation: keep the pivot open with the parent collapsed, then edit and save the underlying model file so Rill reconciles it. In DevTools > Network (filter on "aggregation"), observe fresh aggregation requests fired for the hidden children's subtables (axis, totals, and, if column dimensions are present, sub-table cell queries) despite those rows being invisible.

**Suggested fix:** When a row is collapsed, prune descendant keys from the expanded map in `setPivotExpanded` (drop any key prefixed by the collapsed id), or skip keys in `queryExpandedRowMeasureValues` whose ancestor chain is not fully expanded.
