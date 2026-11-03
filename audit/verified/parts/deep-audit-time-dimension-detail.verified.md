# Verified audit: deep-audit-time-dimension-detail

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **13 numbered findings** (`TIME-1` through `TIME-13`). The embedded NUL byte in the source was ignored while parsing headings. Every heading appears once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| TIME-1 | high | P0 | Stored XSS in TDD row header | `time-dimension-details/TDDTable.svelte:247-262` still interpolates raw dimension text into HTML returned to `RegularTable`; `pivot/RegularTable.svelte:129-138` assigns `innerHTML` | Escape all data-derived text or render DOM/Svelte nodes without `innerHTML`; add malicious-dimension E2E tests. |
| TIME-2 | high | P1 | Dangerous URI schemes reach anchors | `dashboards/dashboard-utils.ts:155-181` still accepts any syntactically present scheme; `TDDTable.svelte:247-254` places it in `href` | Allowlist `https:`, `http:` (and explicitly chosen safe schemes); reject everything else centrally. |
| TIME-3 | high | P1 | Global TDD scroll position can lock another table | `TDDTable.svelte:400-408`; `time-dimension-data-store.ts:469` retain one module-level `lastKnownPosition` with no identity/reset | Scope position by instance/explore/table and clamp/reset on dataset changes/destroy. |
| TIME-4 | medium | P1 | TDD memo cache retains stale runtime context | `time-dimension-data-store.ts:463`; `state-managers/memoize-metrics-store.ts:10-15` retain stores forever and omit instance ID from key | Scope key by runtime instance and evict/dispose with the explore/provider. |
| TIME-5 | medium | P1 | Missing pinned value freezes TDD scrolling | `time-dimension-data-store.ts:120`; `TDDTable.svelte` still waits for selected values that can be absent from the top/search result | Fetch selected values separately or merge placeholders; never gate initialization on exact presence. |
| TIME-6 | medium | P2 | Hover redraws the virtualized viewport repeatedly | `TDDTable.svelte:378` still fans hover state through table-wide redraw paths | Keep hover visual state local/imperative and batch updates. |
| TIME-7 | medium | P2 | Dashboard emissions rebuild the whole TDD query graph | `time-dimension-data-store.ts:343` derives creation from the broad dashboard store | Derive a stable, minimal query config and preserve observers across unrelated state changes. |
| TIME-8 | medium | P1 | Row-query error renders endless skeleton | `time-dimension-data-store.ts:380` still maps error/no-data into the loading path | Propagate error state and render retry/error UI. |
| TIME-9 | medium | P1 | Absolute change is blank when comparison is zero | `time-dimension-data-store.ts:297,314` still uses zero-sensitive truthiness | Check null explicitly; define divide-by-zero separately from absolute delta. |
| TIME-10 | medium | P1 | Null row totals display as zero | `TDDTable.svelte:212` still formats/coerces absent totals | Preserve null through formatting and render the no-data placeholder. |
| TIME-11 | medium | P2 | Table preparation recomputes SVGs/transposition | `time-dimension-data-store.ts:160` still performs full preparation per emission | Memoize by data/config and defer sparkline generation to visible rows. |
| TIME-12 | low | P2 | TDD chart rebuilds provider/query on prop identity | `time-dimension-details/charts/TDDChart.svelte:131` still reconstructs from broad prop identity | Stabilize provider/query inputs and update only semantic changes. |
| TIME-13 | low | P2 | TDD cache key concatenation can collide | `state-managers/memoize-metrics-store.ts:14` remains `metricsName + exploreName` | Use a tuple/length-prefixed key including instance ID. |

## Fixed / remove

None.

## Partial / needs test

None.

Counts: **13 pending (1 P0, 7 P1, 5 P2), 0 fixed, 0 partial = 13**.

## Machine-readable inventory

```text
PENDING|deep-audit-time-dimension-detail.md|TIME-1|high|P0|-|Stored XSS in TDD row header|TDDTable.svelte:247-262
PENDING|deep-audit-time-dimension-detail.md|TIME-2|high|P1|-|Dangerous URI schemes|dashboard-utils.ts:155-181
PENDING|deep-audit-time-dimension-detail.md|TIME-3|high|P1|-|Global TDD scroll lock|TDDTable.svelte:400-408
PENDING|deep-audit-time-dimension-detail.md|TIME-4|medium|P1|-|Stale TDD runtime cache|time-dimension-data-store.ts:463
PENDING|deep-audit-time-dimension-detail.md|TIME-5|medium|P1|-|Missing pinned value freezes table|time-dimension-data-store.ts:120
PENDING|deep-audit-time-dimension-detail.md|TIME-6|medium|P2|-|Hover redraw storm|TDDTable.svelte:378
PENDING|deep-audit-time-dimension-detail.md|TIME-7|medium|P2|-|TDD query graph rebuild|time-dimension-data-store.ts:343
PENDING|deep-audit-time-dimension-detail.md|TIME-8|medium|P1|-|Query error looks loading|time-dimension-data-store.ts:380
PENDING|deep-audit-time-dimension-detail.md|TIME-9|medium|P1|-|Zero comparison blanks absolute change|time-dimension-data-store.ts:297,314
PENDING|deep-audit-time-dimension-detail.md|TIME-10|medium|P1|-|Null total rendered zero|TDDTable.svelte:212
PENDING|deep-audit-time-dimension-detail.md|TIME-11|medium|P2|-|Unmemoized table preparation|time-dimension-data-store.ts:160
PENDING|deep-audit-time-dimension-detail.md|TIME-12|low|P2|-|TDD chart pipeline rebuild|TDDChart.svelte:131
PENDING|deep-audit-time-dimension-detail.md|TIME-13|low|P2|-|Cache-key collision|memoize-metrics-store.ts:14
```
