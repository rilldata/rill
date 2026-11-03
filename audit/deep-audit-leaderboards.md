# Deep Audit: Leaderboards

Scope: `web-common/src/features/dashboards/leaderboard`, `web-common/src/features/dashboards/dimension-table`, and `web-common/src/features/dashboards/leaderboard-context-column.ts`, plus directly coupled shared code (`dashboard-utils.ts`, state-manager selectors/actions). Method: a multi-lens find phase (correctness, state, security, performance, memory) followed by an adversarial verify phase in which every candidate finding was traced end to end through the code (and, where relevant, installed library sources and git history) before being accepted. Two independent reports of the same percent-of-total bug were merged into LEAD-3.

## Summary

| ID | Title | Severity | Category | Verdict |
|----|-------|----------|----------|---------|
| LEAD-1 | `makeHref` renders unvalidated URL schemes from row data as anchor hrefs | medium | security | CONFIRMED |
| LEAD-2 | Measure value of exactly 0 renders as no-data "-" in leaderboard cells | medium | correctness | CONFIRMED |
| LEAD-3 | Percent-of-total divides by undefined/zero total, rendering "NaN%"/"Infinity%" | medium | correctness | CONFIRMED |
| LEAD-4 | Selected filter values with no data in the current time range vanish from the leaderboard | medium | correctness | CONFIRMED |
| LEAD-5 | Selected value can appear twice in leaderboard: strict `Set.delete` vs coercing selection compare | medium | state | CONFIRMED |
| LEAD-6 | Dimension table selection matching lacks the legacy string-filter compatibility used by leaderboards | medium | state | CONFIRMED |
| LEAD-7 | Global Ctrl/Cmd+A handler hijacks select-all from textareas/contenteditable, bulk-mutating filters | medium | state | CONFIRMED |
| LEAD-8 | Dimension table virtualizers torn down and recreated on every scroll range change | medium | performance | CONFIRMED |
| LEAD-9 | `rowOverscanAmount=120` effectively disables row virtualization for the 250-row dimension table | medium | performance | CONFIRMED |
| LEAD-10 | `selectedDimensionValues` stores recreated inline for every leaderboard on every dashboard-store change | medium | performance | CONFIRMED |
| LEAD-11 | Dimension table column widths estimated once at mount from empty rows and never recomputed | medium | state | CONFIRMED |
| LEAD-12 | `DIMENSION_TABLE_CONFIG` singleton mutated per-table and never reset; header height leaks across dashboards | low | state | CONFIRMED |
| LEAD-13 | Leaderboard IntersectionObserver never disconnected on component destroy | low | memory-leak | CONFIRMED |

## LEAD-1: `makeHref` renders unvalidated URL schemes from row data as anchor hrefs

**File:** `web-common/src/features/dashboards/dashboard-utils.ts:174` · **Severity: medium** · **Verdict: CONFIRMED**

`makeHref` (dashboard-utils.ts:155-181) validates only that a URI has *some* scheme, not that the scheme is safe: any value containing a protocol is returned unchanged and bound directly to `<a href>` in `LeaderboardRow.svelte`, virtualized-table `Cell.svelte`, and the pivot/dimension-table paths (via `makeDimensionHref`), with no sanitization. When a dimension is configured with `uri: true` the href is the raw dimension value, and for the expression form it is the per-row `__rill_uri` measure resolved from the dataset — so `javascript:`, `data:`, `file:`, `vbscript:`, `intent:` etc. reach the DOM unchecked. This is a genuine missing-scheme-allowlist defect. It is medium rather than high: exploitation requires the author to enable `uri` on an attacker-influenced column, a victim click, and (for the `javascript:`/`data:` vectors) a legacy WebView/embedding context, since modern desktop browsers block those schemes for the `target="_blank"` top-level navigations these anchors use.

**Manual reproduction**
1. In Rill Developer, create/edit a model whose data includes a text column (e.g. `landing_page`) and insert a row where that column's value is `javascript:alert(document.domain)` (or `data:text/html,<script>alert(1)</script>`).
2. In the metrics view YAML, define a dimension on that column and set `uri: true` on it.
3. Open an explore dashboard built on that metrics view and view the leaderboard for that dimension (or open the dimension detail table).
4. Observe the external-link icon rendered next to the malicious row value; hover to see the `title`/href is the raw `javascript:`/`data:` string (no `https://` prefix was added because a scheme was detected).
5. Inspect the anchor in DevTools: `<a target="_blank" rel="noopener noreferrer" href="javascript:alert(document.domain)">`. In older WebView/embedding contexts, clicking it executes the scheme; in modern desktop Chrome/Firefox the `javascript:`/`data:` navigation is blocked, but other schemes (`file:`, `intent:`, `vbscript:`) still pass through unvalidated.
6. Confirm the missing guard directly: `leaderboard-utils.spec.ts` has no scheme-validation case, and `makeHref` returns any protocol-bearing string unchanged.

**Suggested fix:** Allowlist schemes in `makeHref`: parse with `new URL(...)` after the `https://` fallback and return `undefined` unless the protocol is `http:` or `https:` (optionally `mailto:`).

## LEAD-2: Measure value of exactly 0 renders as no-data "-" in leaderboard cells

**File:** `web-common/src/features/dashboards/leaderboard/LeaderboardRow.svelte:261` · **Severity: medium** · **Verdict: CONFIRMED**

In the leaderboard, a measure value of exactly 0 renders as the no-data placeholder "-" instead of a formatted 0. `LeaderboardRow.svelte` uses truthiness rather than a null check in three cells: the main value cell (lines 261-263), the time-comparison absolute-delta cell (lines 305-307), and the relative-delta cell (lines 335-337). Because `0` is falsy, the formatted value is replaced with `null`, which `FormattedDataType`/`Base` and `PercentageChange` render as "-" (their explicit no-data state). The zero survives the data pipeline intact (`finiteOrNull` in `leaderboard-utils.ts` preserves 0), so any dimension value whose measure, absolute delta, or relative delta is exactly 0 displays as if data were missing. This is inconsistent with the same row's hover tooltip (line 248, which uses `!= null` and shows "0") and with the expanded dimension table (`dimension-table-utils.ts` line 473, `val !== null`), which both show the formatted 0.

**Manual reproduction**
1. In Rill Developer, create a model with a dimension value whose measure is exactly 0, e.g. a SQL model: `SELECT 'a' AS cat, 0 AS num, TIMESTAMP '2026-07-10' AS ts UNION ALL SELECT 'b', 10, TIMESTAMP '2026-07-10'`.
2. Create a metrics view on the model with dimension `cat`, measure `sum(num)`, and timeseries `ts`, plus an explore dashboard for it.
3. Open the explore dashboard and look at the `cat` leaderboard: the row for 'a' shows '-' in the measure column instead of '0'.
4. Hover over that cell: the tooltip shows '0', contradicting the '-' shown in the cell.
5. Click 'Expand' to open the dimension table for `cat`: the same row shows a formatted '0', confirming the leaderboard and table disagree for identical data.
6. For the delta cells: extend the model so a dimension value has the same measure value in two adjacent periods (e.g. 'b' = 10 on 2026-07-09 and 10 on 2026-07-10), enable time comparison (e.g. 'Previous day'), and enable comparison columns. The absolute-delta column for 'b' shows '-' instead of '0', and the percent-delta column shows '-' instead of '0%'.

**Suggested fix:** Use explicit null checks in all three cells: `values[measureName] != null ? formatters[measureName]?.(values[measureName]) : null`, and likewise for `deltaAbsMap` and `deltaRels`.

## LEAD-3: Percent-of-total divides by undefined/zero total, rendering "NaN%"/"Infinity%"

**File:** `web-common/src/features/dashboards/leaderboard/leaderboard-utils.ts:106` (also line 216) · **Severity: medium** · **Verdict: CONFIRMED** (independently found and verified twice; merged)

Both `cleanUpComparisonValue` (lines 105-106) and `prepareLeaderboardItemData` (lines 215-216) compute percent-of-total with the guard `total !== null && value !== null ? value / total : null`, which misses `total === undefined` and `total === 0`. Because `Leaderboard.svelte` initializes `leaderboardTotals` to `{}` until the totals query resolves (lines 213-220) and gates row rendering only on the sorted-rows query's loading state (line 210), leaderboard percent-of-total cells transiently render "NaN%" whenever the sorted query resolves before the independent totals query (intermittent, since the cheaper ungrouped totals query usually wins the race; below-the-fold rows are not gated by any loading state at all). Separately and permanently, a measure with `valid_percent_of_total: true` whose filtered total is exactly 0 (e.g. a net amount summing to zero) renders "Infinity%"/"-Infinity%" for nonzero rows (and "NaN%" for zero rows), because `PercentageChange.svelte` treats NaN/Infinity as ordinary numbers (`Math.round(100 * value).toString()`). The dimension table guards both cases (`dimension-table-utils.ts` lines 92 and 528 return `PERC_DIFF.CURRENT_VALUE_NO_DATA`), so the leaderboard is inconsistent with it.

**Manual reproduction**
1. Deterministic case (Infinity%): create a model whose measure sums to zero, e.g. `SELECT * FROM (VALUES ('a', 5.0), ('b', -5.0)) t(category, net)`.
2. Create a metrics view with dimension `category` and measure `net_total` = `SUM(net)` with `valid_percent_of_total: true`, and an explore dashboard on it.
3. Open the explore dashboard with the leaderboard sorted by `net_total` so the percent-of-total context column is visible.
4. Observe the percent-of-total column shows "Infinity%" for row 'a' (5/0) and "-Infinity%" for row 'b' (-5/0), permanently. The expanded dimension table for the same dimension correctly shows '-' for the same cells.
5. Transient case (NaN%): open any explore dashboard with a `valid_percent_of_total` measure; in Chrome DevTools enable network throttling (e.g. Slow 4G) and repeatedly change a filter or the time range.
6. Whenever the sorted leaderboard row query resolves before the totals query, the percent-of-total column briefly shows "NaN%" on every row until the totals response arrives.

**Suggested fix:** Change the guard to `total != null && total !== 0 && value !== null ? value / total : null` (or use `Number.isFinite(value / total)`) in both `cleanUpComparisonValue` and `prepareLeaderboardItemData`.

## LEAD-4: Selected filter values with no data in the current time range vanish from the leaderboard

**File:** `web-common/src/features/dashboards/leaderboard/Leaderboard.svelte:272` · **Severity: medium** · **Verdict: CONFIRMED**

The below-the-fold fallback in `Leaderboard.svelte` (lines 272-279) is all-or-nothing: if the below-the-fold aggregation query returns any rows, only those rows are rendered; placeholder rows (with null measure values) for selected values are rendered only when the query returns zero rows. Because the query filters `dim IN (belowTheFoldValues)` over the selected time range and a GROUP BY returns no row for a value with no underlying data, a partial result silently drops selected filter values from the leaderboard. Concretely: with values A and B in the include filter and a time range where A has data (but is not top-7) and B has none, only A renders below the fold and B disappears from the leaderboard entirely — even though the same code renders both as null-valued placeholders when neither has data, showing the intended behavior is to always display selected values. The value remains visible and removable in the top filter bar pill, so the impact is confined to the leaderboard surface.

**Manual reproduction**
1. In Rill Developer, open an explore dashboard backed by time-series data where a dimension has values that only occur in part of the overall time range (e.g. a 'customer' value that stopped appearing after a certain date).
2. Set a wide time range (e.g. All Time). In a dimension leaderboard, click to include two values A and B that are both outside the top 7 for the current sort measure; both appear as selected rows below the fold (under the divider).
3. Narrow the time range to a window where A still has data but B has zero rows, keeping both outside the top 7.
4. Observe the leaderboard: only A is rendered below the fold; B has vanished from the leaderboard even though it is still an active include filter (the filter pill at the top still lists B).
5. For contrast, pick a time range where neither A nor B has any data: both now reappear below the fold as rows with null/blank measure values, demonstrating the inconsistent all-or-nothing fallback.

**Suggested fix:** Merge instead of replace: start from the query rows, then append placeholder `{ [dimensionName]: value, ...: null }` entries for any `belowTheFoldValues` not present in `data.data` (matched with `compareLeaderboardValues`).

## LEAD-5: Selected value can appear twice in leaderboard: strict `Set.delete` vs coercing selection compare

**File:** `web-common/src/features/dashboards/leaderboard/leaderboard-utils.ts:189` · **Severity: medium** · **Verdict: CONFIRMED**

In `prepareLeaderboardItemData` (line 189), `belowTheFoldValues` is seeded from `selectedValues` and pruned with `Set.delete(dimensionValue)` (strict equality), but selection highlighting uses the type-coercing `compareLeaderboardValues` (lines 191-193), the helper added in commit `7c58f94f2` specifically because legacy filters stringify all non-null values. When a numeric dimension has a filter value stored as a string (legacy proto-state URL/bookmark, or a current URL like `?f=year IN ('2024')`, since the filter grammar parses quoted values as strings) and the matching row is in the top-`slice` results (returned by the aggregation API as a number), the row renders selected above the fold but the string stays in `belowTheFoldValues`. `Leaderboard.svelte` then fires an extra below-the-fold aggregation query and renders the same dimension value a second time as a below-the-fold row; the duplicate appears even if the extra query returns nothing, because the fallback at lines 272-279 fabricates a null-measure row from `belowTheFoldValues`. This is a regression against the intent of `7c58f94f2`, which pruned by the comparator-found index.

**Manual reproduction**
1. In Rill Developer, create a project whose model has an integer column, e.g. `SELECT 2020 + (range % 6) AS year, range AS id FROM range(0, 1000)`, and a metrics view exposing `year` as a dimension with a `count(*)` measure, plus an explore dashboard for it.
2. Open the explore dashboard so the `year` leaderboard is visible; note the top values (e.g. 2024 is within the top 7 rows).
3. Edit the browser URL to add a dimension filter with the numeric value quoted as a string: append `?f=year IN ('2024')`. Alternatively, load a legacy `state=` proto URL or an old bookmark that stored the numeric filter value as a string.
4. Observe the `year` leaderboard: 2024 renders selected in the above-the-fold list AND appears a second time as an extra below-the-fold row under the divider (duplicated entry, both marked selected).
5. Optionally open DevTools > Network and observe an additional `metrics-view-aggregation` request whose where clause is `year = '2024'` — a wasted query issued solely for the value already shown above the fold. Even if that request returns no rows, the duplicate row still renders (with a blank measure) via the client-side fallback.

**Suggested fix:** Prune with the same comparator used for selection: iterate `selectedValues` and remove entries where `compareLeaderboardValues(selected, dimensionValue)` is true, instead of `Set.delete(dimensionValue)`.

## LEAD-6: Dimension table selection matching lacks the legacy string-filter compatibility used by leaderboards

**File:** `web-common/src/features/dashboards/dimension-table/DimensionTable.svelte:66` · **Severity: medium** · **Verdict: CONFIRMED**

`DimensionTable.svelte` (lines 64-67) computes `selectedIndex` and `DimensionDisplay.svelte` (lines 191-193) computes `areAllTableRowsSelected` using strict equality between filter values and aggregation-row values, while the leaderboard uses `compareLeaderboardValues` (`leaderboard-utils.ts` line 317, added in #3769) precisely because older filters stored all non-null values as strings. `selectedDimensionValues` returns raw filter-expression values and `prepareDimensionTableRows` keeps raw JSON row values (its `as string` is a type cast only), so for a non-string dimension with a string-typed filter value (legacy bookmark, legacy URL, or any URL of the form `dim IN ('2024')`) the leaderboard shows the value as selected but the expanded dimension table shows no checkmark in the gutter, `areAllTableRowsSelected` is false, and select-all / Meta+A (via `selectItemsInFilter`, which also compares strictly) appends duplicate typed values (e.g. `IN ('2024', 2024)`) for values already in the filter.

**Manual reproduction**
1. In Rill Developer, open a project whose metrics view has a dimension on a non-string column (e.g. an INTEGER column such as `year` or `status_code`).
2. Open the explore dashboard for that metrics view.
3. Edit the URL to add a string-typed filter for the numeric dimension, e.g. append `?f=year IN ('2024')` (URL-encoded: `?f=year%20IN%20('2024')`), and load the page.
4. Look at the leaderboard for that dimension: the row for 2024 shows a check icon (selected), because the leaderboard matches via `compareLeaderboardValues`.
5. Click the dimension name to expand it into the dimension table.
6. Observe the bug: the 2024 row has no checkmark in the filter gutter, and the select-all checkbox in the dimension header is unchecked even though every visible row is in the filter.
7. Press Meta+A (or click the select-all control): a notification reports items were added, and the filter pill now contains duplicate entries for the same value (string '2024' plus number 2024) instead of recognizing the rows as already selected.

**Suggested fix:** Use `compareLeaderboardValues(label, row[dimensionName])` in both `DimensionTable.svelte` and `DimensionDisplay.svelte` when matching selected filter values against query rows.

## LEAD-7: Global Ctrl/Cmd+A handler hijacks select-all from textareas/contenteditable, bulk-mutating filters

**File:** `web-common/src/features/dashboards/dimension-table/DimensionDisplay.svelte:233` · **Severity: medium** · **Verdict: CONFIRMED**

`DimensionDisplay.svelte` attaches a window-level keydown handler for Ctrl/Cmd+A whose only exemption is `e.target.tagName === "INPUT"`. While a dimension table is expanded, pressing Cmd/Ctrl+A with focus in any other editable element — the TipTap contenteditable chat input (`ChatInput.svelte`; ProseMirror never stops propagation, so the event reaches window even though TipTap handles Mod-a itself) or any `<textarea>` such as the chat FeedbackModal's — triggers `toggleAllSearchItems()`, bulk-adding up to 250 dimension values (queryLimit) to the dashboard filter as a side effect, announced only by a toast. In the textarea case the handler's `preventDefault()` additionally blocks the user's native select-all.

**Manual reproduction**
1. Run `rill start` on a project with an explore dashboard and open the explore in the browser.
2. Click a dimension name in a leaderboard to expand it into the dimension table view (this mounts DimensionDisplay and its window keydown handler).
3. Note the current filter chips for that dimension (ideally none).
4. Open the AI chat sidebar (chat toggle in the top bar), click into the chat input, and type a sentence.
5. Press Cmd+A (macOS) or Ctrl+A (Windows/Linux).
6. Observe: a toast appears ('Added N items to filter', N up to 250) and the dimension's filter chip is now populated with every visible table value — the filter was bulk-mutated even though focus was in the chat input.
7. Textarea variant: send a chat message, hover the assistant response and click thumbs-down to open the feedback modal, type into the textarea, press Cmd/Ctrl+A. Observe that the text is NOT selected (preventDefault) and the same filter-mutation toast fires behind the modal.

**Suggested fix:** Exempt all text-entry targets: return early when `e.target` is an `HTMLInputElement`, `HTMLTextAreaElement`, or matches `e.target.closest('[contenteditable="true"]')` (or `isContentEditable`).

## LEAD-8: Dimension table virtualizers torn down and recreated on every scroll range change

**File:** `web-common/src/features/dashboards/dimension-table/DimensionTable.svelte:103` · **Severity: medium** · **Verdict: CONFIRMED**

There is a reactive feedback loop between each virtualizer store and its own creation: `$: rowScrollOffset = $rowVirtualizer?.scrollOffset || 0` (line 70) feeds `initialOffset: rowScrollOffset` inside `$: rowVirtualizer = createVirtualizer({...})` (lines 103-114), and the same pattern applies to `colScrollOffset`/`columnVirtualizer` (lines 72, 129-140). The @tanstack/svelte-virtual store emits on every virtual-core `onChange`, which fires whenever `isScrolling` flips or the visible range's start/end index changes (every ~24px of vertical scroll at rowHeight 24). Each emission updates `rowScrollOffset`, which re-runs the `createVirtualizer` statement, causing Svelte to unsubscribe the old store (full virtualizer cleanup: scroll/rect listeners removed, observer disconnected) and subscribe a new one (new Virtualizer construction, `setOptions`, `_didMount`, listener re-registration, recomputation of all row measurements). The loop settles after one recreation per emission because the new instance seeds its scrollOffset from `initialOffset`, so scroll position is visually preserved — which masks the waste. Introduced by PR #3296, which changed the original `$: rowScrollOffset = 0` to reactively track the store. Impact is bounded: the dimension table caps rows at 250, no code path uses `measureElement`, and the old instance is fully cleaned up (no leak), but the component still performs a full virtualizer teardown/re-mount plus an extra reactive flush on every row-boundary crossing and at every scroll gesture start/end.

**Manual reproduction**
1. Start Rill Developer and open an explore dashboard whose dimension has more than ~100 distinct values.
2. Click a dimension name in the leaderboard to open the full dimension table (it loads up to 250 rows).
3. Open browser DevTools > Sources, locate `DimensionTable.svelte` in the dev bundle, and add a logpoint on the `$: rowVirtualizer = createVirtualizer({...})` statement (line ~103); alternatively add a logpoint in `node_modules/@tanstack/svelte-virtual/dist/index.js` inside `createVirtualizerBase`.
4. Scroll the dimension table vertically by a few hundred pixels in one continuous gesture.
5. Observe the logpoint firing once per ~24px of scroll (each row-boundary crossing) plus at scroll start and end — each hit is a full Virtualizer teardown and reconstruction, not a single creation at mount. Scrolling horizontally across measure columns shows the same behavior for the column virtualizer.
6. For contrast, apply the suggested fix and repeat: the logpoint fires only once at mount.

**Suggested fix:** Create each virtualizer once (e.g. in `onMount`) and push option changes with `$rowVirtualizer.setOptions({ count: rows.length, ... })` in a reactive block; remove the `$: rowScrollOffset = $rowVirtualizer?.scrollOffset || 0` lines so scroll position is only captured imperatively before deliberate recreations.

## LEAD-9: `rowOverscanAmount=120` effectively disables row virtualization for the 250-row dimension table

**File:** `web-common/src/features/dashboards/dimension-table/DimensionTable.svelte:52` · **Severity: medium** · **Verdict: CONFIRMED**

`DimensionTable.svelte` defaults `rowOverscanAmount` to 120 and passes it as `overscan` to the TanStack row virtualizer. TanStack's overscan is measured in items (virtual-core computes `startIndex - overscan` to `endIndex + overscan`; its own default is 1), so with `rowHeight: 24` and a typical ~1000px viewport (~40 visible rows), the rendered range spans firstVisible-120 through lastVisible+120 — always covering all rows, since `DimensionDisplay.svelte` caps the query at 250 rows (`queryLimit = 250`) and passes no override. Row virtualization is therefore a no-op: all 250 rows are always mounted, and `TableCells.svelte` renders renderedRows × renderedColumns `Cell` components (~2,500 with 10 measure/context columns within the 12-per-side column overscan, more with comparisons enabled), plus 250 dimension-value rows and 250 filter-gutter rows. This inflates initial mount time and makes every update that touches row-level props (selection changes, and each LEAD-8 virtualizer recreation) re-render the full table instead of just the visible window.

**Manual reproduction**
1. In Rill Developer, open an explore dashboard backed by a metrics view with many measures (10+, ideally with time comparison enabled) and a high-cardinality dimension (250+ distinct values).
2. Click the dimension in the left panel to expand it into the dimension table, which fetches up to 250 rows.
3. Open browser DevTools > Elements and count rendered row nodes (e.g. inspect the absolutely-positioned row divs in Console): all 250 rows and their cells are present in the DOM even though only ~40 fit in the viewport, and scrolling to the bottom never removes nodes at the top.
4. Open DevTools > Performance, start recording, then click a dimension value to toggle a filter selection: the flame chart shows a long scripting/render block proportional to the full 250-row × rendered-column cell count rather than just the visible window.
5. For comparison, edit `rowOverscanAmount` to 10 and repeat steps 3-4: DOM node count drops to roughly the visible window plus 20 rows, and the selection-update flame chart shrinks correspondingly.

**Suggested fix:** Reduce the default row overscan to a small item count (TanStack default is 1-4; even 10-20 keeps scrolling smooth), keeping the prop override for consumers that need more.

## LEAD-10: `selectedDimensionValues` stores recreated inline for every leaderboard on every dashboard-store change

**File:** `web-common/src/features/dashboards/leaderboard/LeaderboardDisplay.svelte:104` · **Severity: medium** · **Verdict: CONFIRMED**

In `LeaderboardDisplay.svelte` (line ~104), `selectedValues` is computed inline inside the `{#each $visibleDimensions}` block via `selectedDimensionValues(client, [metricsViewName], $dashboardStore.whereFilter, ...)`. Because `$dashboardStore` re-emits on every explore-state mutation (it derives from the global `metricsExplorerStore` with object-typed values, which Svelte never dedupes), the expression re-runs for all N visible leaderboards on every dashboard change — sort direction, leaderboard measure selection, comparison toggles, and even per-mousemove chart scrubbing (`setSelectedScrubRange`) — not just filter changes. Each evaluation returns a brand-new store, so every Leaderboard resubscribes and re-runs everything depending on `$selectedValues`: `atLeastOneActive`, the full `prepareLeaderboardItemData` row rebuild, and `belowTheFoldRows`' per-row `findIndex`. When a dimension has a Contains filter, each evaluation additionally constructs fresh TanStack query observers via `useDimensionSearch`; because the global query client sets `refetchOnMount: false`, this causes no extra network requests or loading flash — the cost is observer construction/teardown churn only. The impact is wasted CPU across all leaderboards on unrelated state changes, worst during chart scrubbing on dashboards with many dimensions. `useSelectedValuesForCompareDimension` in the same file demonstrates the intended memoized `derived` pattern.

**Manual reproduction**
1. Start Rill Developer with a project whose explore dashboard has many dimensions (10+ visible leaderboards); open the explore dashboard.
2. Add a Contains filter: click a dimension filter pill, switch the mode to Contains, type some search text, and apply it.
3. Open browser DevTools > Performance and start recording (optionally add a `console.count` or breakpoint in `selectedDimensionValues` in `state-managers/selectors/dimension-filters.ts`).
4. Perform state changes unrelated to filters: click a leaderboard sort header a few times, toggle the leaderboard measure selection, then click-drag (scrub) across the timeseries chart.
5. In the flame chart (or console counts), observe `selectedDimensionValues` / `useDimensionSearch` / `prepareLeaderboardItemData` executing once per visible leaderboard for every interaction — including per mousemove during the scrub drag — even though dimension selections never changed.
6. Confirm in the Network tab that no new aggregation requests fire during these interactions (the waste is CPU and query-observer churn, not network traffic).

**Suggested fix:** Hoist selection stores out of the template: build a derived map keyed by dimension name that only recreates a dimension's store when `whereFilter`/`timeRange` actually change (like `useSelectedValuesForCompareDimension`), and pass the memoized store as the prop.

## LEAD-11: Dimension table column widths estimated once at mount from empty rows and never recomputed

**File:** `web-common/src/features/dashboards/dimension-table/DimensionTable.svelte:78` · **Severity: medium** · **Verdict: CONFIRMED**

In `DimensionTable.svelte`, `estimateColumnCharacterWidths(columns, rows)` is a plain `const` (lines 78-81) that runs once at mount. DimensionDisplay mounts the table as soon as `columns` (derived from the metrics view spec, not row data) is non-empty, which on an uncached first open is before the sorted aggregation query resolves, i.e. with `rows = []`. `Math.max(...[])` yields `-Infinity` per column, and because `-Infinity` is truthy, the dimension value column is thereafter sized purely from its header via `Math.max(-Infinity, effectiveHeaderWidth, minColumnWidth)`, clamping it to roughly 104-160px instead of up to the 320px cap. The reactive `estimateColumnSizes` block (line 116) re-runs on row/search changes but reuses the stale mount-time `columnWidths` map, so long dimension values stay truncated until the user manually resizes the column or the component remounts with cached (non-empty) rows. Separately, lines 84-86 mutate the module-level singleton `DIMENSION_TABLE_CONFIG` (`config.columnHeaderHeight = 46`) with no reset path; note this trigger depends on measure column labels (any label over 9 characters), not on the empty rows — see LEAD-12.

**Manual reproduction**
1. In Rill Developer, open a project with an explore dashboard whose dimension has long values (e.g. URLs or descriptions well over ~25 characters).
2. Hard-refresh the browser (to empty the TanStack Query cache), open the explore dashboard, and click the dimension's name in a leaderboard to expand the dimension table. The table mounts showing 'Loading...' before rows arrive.
3. After data loads, observe the dimension value column is only ~104-160px wide and long values are truncated with ellipses. Typing in the dimension search box to surface even longer values never widens the column either.
4. Collapse back to the leaderboard view and immediately re-open the same dimension: this remounts DimensionTable with cached rows present, and the column now sizes to its content (up to 320px), confirming the width is only estimated at mount time.
5. For the shared-config mutation: hard-refresh, open a dashboard whose visible measure labels are all 9 characters or fewer and expand a dimension table; note the column header row is 28px tall. Then open any dimension table (same or different dashboard) that has a measure label longer than 9 characters (headers become 46px). Return to the first dimension table: its headers now also render 46px tall for the rest of the session.

**Suggested fix:** Make the estimation reactive (`$: ({ columnWidths, largestColumnLength } = estimateColumnCharacterWidths(columns, rows))`, guarded to run when rows first become non-empty), and derive `columnHeaderHeight` per-instance (e.g. a local copy of the config via spread) instead of mutating the shared `DIMENSION_TABLE_CONFIG`.

## LEAD-12: `DIMENSION_TABLE_CONFIG` singleton mutated per-table and never reset; header height leaks across dashboards

**File:** `web-common/src/features/dashboards/dimension-table/DimensionTable.svelte:85` · **Severity: low** · **Verdict: CONFIRMED**

`DIMENSION_TABLE_CONFIG` is a module-level const object (`columnHeaderHeight: 28`) that both `DimensionTable.svelte` (line 85) and `AlertPreviewTable.svelte` (line 41) mutate in place at component init, setting `columnHeaderHeight` to 46 whenever any measure column label exceeds 9 characters; nothing ever writes 28 back. Since ES modules are singletons per browser session, the mutation is a one-way ratchet shared across all dimension tables and alert preview tables: once any table with a long measure name mounts, every subsequently mounted table renders 46px-tall headers even when all its column labels are short. Because the header render and the virtualizer `paddingStart` read the same object at mount, geometry within one table stays self-consistent, so the user-visible effect is unnecessarily tall (double-height) headers and inconsistent header sizing depending on navigation order, rather than broken row alignment; a true header/body offset mismatch is only reachable if a mounted short-label table's virtualizer re-initializes (e.g. on sort) after another component (such as an alert preview dialog) has mutated the singleton.

**Manual reproduction**
1. Create a project with one metrics view/explore where every measure has a short display name of 9 characters or fewer (e.g. label "Count"), and a second metrics view/explore with a measure whose display name is longer than 9 characters (e.g. "Total Records").
2. Hard-refresh the browser and open the short-label explore first. Click a dimension to expand the dimension table. In DevTools, inspect a measure column header button (StickyHeader) and note its height: 28px.
3. Without reloading, navigate to the second explore (long measure name) and expand any dimension table. This executes `config.columnHeaderHeight = 46` on the shared module singleton.
4. Navigate back to the first (short-label) explore and expand the same dimension table as in step 2. Inspect the same header element: its height is now 46px, and the first data row starts 18px lower than in step 2, despite identical data.
5. Confirm the ratchet: hard-refresh the page while on the short-label explore and expand the dimension table again; the header is back to 28px, proving the 46px rendering in step 4 was leaked state from the other dashboard. The same leak can be triggered via an alert creation preview table (`AlertPreviewTable`) instead of a second dashboard.

**Suggested fix:** Make the config immutable and compute the header height as local component state (e.g. `const columnHeaderHeight = largestColumnLength > LIMIT ? 46 : config.columnHeaderHeight`), passing it through context instead of mutating the module constant.

## LEAD-13: Leaderboard IntersectionObserver never disconnected on component destroy

**File:** `web-common/src/features/dashboards/leaderboard/Leaderboard.svelte:96` · **Severity: low** · **Verdict: CONFIRMED**

In `Leaderboard.svelte`, `onMount` (lines 93-110) creates an IntersectionObserver with root set to the explore dashboard's long-lived scroll container and observes the leaderboard container, but never returns a cleanup function and has no `onDestroy`; the only `unobserve` happens inside the callback when the leaderboard first intersects. In the explore dashboard's `LeaderboardDisplay.svelte`, one Leaderboard per visible dimension is rendered in a keyed `{#each}`, so leaderboards below the scroll fold that never intersect are destroyed with their observer registration still active whenever the user hides or reorders dimensions. This accumulates stale IntersectionObserver registrations against the persistent scroll container for as long as the page lives. The impact is bounded in practice: per the IntersectionObserver design, unreachable observers/targets remain garbage-collectable without an explicit disconnect (Chromium holds observation records weakly from the observer), so this leaks small observer objects and transient intersection bookkeeping rather than guaranteed detached-DOM subtrees, though exact retention is engine-dependent. The canvas leaderboard surface is unaffected (it passes `visible` directly and no `parentElement`).

**Manual reproduction**
1. Run `rill start` on a project with an explore dashboard that has enough dimensions that several leaderboards render below the scroll fold (e.g. 10+ dimensions).
2. Open the explore dashboard in Chrome and do NOT scroll the leaderboard panel, so below-fold leaderboards never intersect and keep their observers registered.
3. Open DevTools > Memory, click the garbage-collect (trash) icon, take a heap snapshot, and search the class filter for 'IntersectionObserver'; note the instance count.
4. Using the leaderboard's dimension visibility control, hide most dimensions, then re-show them; repeat this hide/show cycle about 10 times without scrolling the leaderboard panel.
5. Click the garbage-collect icon again and take a second heap snapshot; search for 'IntersectionObserver' and compare counts. Surviving observers from destroyed Leaderboard components indicate the leak; optionally search 'Detached' to check whether any leaderboard `div[role=table]` subtrees are also retained in your browser/engine.
6. To confirm the fix, apply `return () => observer.disconnect();` at the end of the `onMount` callback, rebuild, and repeat steps 2-5: the IntersectionObserver count should return to baseline after GC.

**Suggested fix:** Return `() => observer.disconnect()` from `onMount` so the observer is released on destroy (and use `observer.disconnect()` instead of `unobserve` in the callback).
