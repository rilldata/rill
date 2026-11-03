# Verified audit: deep-audit-pivot

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **14 numbered findings** (`PIVO-1` through `PIVO-14`). Every heading appears once.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| PIVO-1 | high | P1 | Dropped axes misassign cell data to rows | `pivot/pivot-table-transformations.ts:249`; assembly still keeps unfiltered `rowDimensionValues` after totals drop values | Return and use the filtered axes list for both skeleton and cell placement; regression-test comparison-only rows. |
| PIVO-2 | high | P1 | Expanded-table cache discards fresh pages/data | `pivot/pivot-data-assembly.ts:89-95` still returns a cached table by config key regardless of fresh query/page | Cache expansion state separately, include pagination, and invalidate on fresh cell data/collapse. |
| PIVO-3 | high | P1 | Module-global lastKey breaks multi-pivot pagination | `pivot/pivot-data-config.ts:25` and `canvas/components/pivot/util.ts:88` still share module-level reset keys | Move `lastKey` into each store instance. |
| PIVO-4 | medium | P2 | Unrelated dashboard changes rebuild pivot pipeline | `pivot/pivot-data-store.ts:142` still derives from broad state and recreates downstream stores | Select a stable minimal config and preserve query observers. |
| PIVO-5 | medium | P2 | Any expansion emission re-merges every subtable | `pivot/pivot-data-store.ts:488` still assembles all expanded data together | Incrementally update the changed expansion and memoize unchanged subtrees. |
| PIVO-6 | medium | P1 | Pivot config key omits search text | `pivot/pivot-utils.ts:100` still excludes `searchText` | Include normalized search text in cache/reset keys and add pagination tests. |
| PIVO-7 | medium | P2 | Scroll feedback recreates row virtualizer | `pivot/PivotTable.svelte:164` still includes scroll offset in the virtualizer creation chain | Create virtualizer once; update offset/options imperatively. |
| PIVO-8 | medium | P2 | Leaf-column construction repeatedly scans totals | `pivot/pivot-column-definition.ts:90` still rebuilds keys/filters inside candidate loops | Pre-index totals/axes once per response. |
| PIVO-9 | medium | P1 | Memoized pivot stores/caches never evict | `pivot/pivot-data-store.ts:699` still retains per-explore stores for process lifetime | Scope by provider/instance and dispose maps/query subscriptions on teardown. |
| PIVO-10 | medium | P2 | Header hover recomputes every body cell | `pivot/NestedTable.svelte:592` still propagates hover through non-virtualized cell state | Apply hover using a column class/CSS variable without per-cell state recomputation. |
| PIVO-11 | medium | P1 | Missing column dimension crashes sort accessor | `pivot/pivot-utils.ts:361` still assumes a matching dimension exists | Validate accessor axes and ignore/reset invalid persisted sort state. |
| PIVO-12 | medium | P0 | Stored XSS through RegularTable innerHTML | `pivot/RegularTable.svelte:129-174,197-231` still assigns data/formatter strings to `innerHTML` | Replace HTML string renderers with safe text/nodes; tightly sanitize unavoidable markup and add XSS tests. |
| PIVO-13 | low | P2 | Expansion key traversal skips missing levels | `pivot/pivot-expansion.ts:75` still uses `forEach`, which cannot break on a missing level | Use a loop with early return/failure and validate path shape. |
| PIVO-14 | low | P2 | Hidden descendant expansions keep querying | `pivot/pivot-expansion.ts:203` still builds work from entries under collapsed ancestors | Prune/disable descendants when an ancestor collapses. |

## Fixed / remove

None. PR #9763 fixes global-filter scope in pivot click filtering and does not touch these findings.

## Partial / needs test

None.

Counts: **14 pending (1 P0, 6 P1, 7 P2), 0 fixed, 0 partial = 14**.

## Machine-readable inventory

```text
PENDING|deep-audit-pivot.md|PIVO-1|high|P1|-|Rows receive wrong cells after axis drop|pivot-table-transformations.ts:249
PENDING|deep-audit-pivot.md|PIVO-2|high|P1|-|Expanded cache discards fresh data|pivot-data-assembly.ts:89-95
PENDING|deep-audit-pivot.md|PIVO-3|high|P1|-|Shared lastKey breaks pagination|pivot-data-config.ts:25
PENDING|deep-audit-pivot.md|PIVO-4|medium|P2|-|Full pivot pipeline rebuild|pivot-data-store.ts:142
PENDING|deep-audit-pivot.md|PIVO-5|medium|P2|-|All expansions re-merged|pivot-data-store.ts:488
PENDING|deep-audit-pivot.md|PIVO-6|medium|P1|-|Config key omits search|pivot-utils.ts:100
PENDING|deep-audit-pivot.md|PIVO-7|medium|P2|-|Virtualizer recreated on scroll|PivotTable.svelte:164
PENDING|deep-audit-pivot.md|PIVO-8|medium|P2|-|Repeated totals scans|pivot-column-definition.ts:90
PENDING|deep-audit-pivot.md|PIVO-9|medium|P1|-|Pivot caches never evict|pivot-data-store.ts:699
PENDING|deep-audit-pivot.md|PIVO-10|medium|P2|-|Hover recomputes body cells|NestedTable.svelte:592
PENDING|deep-audit-pivot.md|PIVO-11|medium|P1|-|Missing sort dimension throws|pivot-utils.ts:361
PENDING|deep-audit-pivot.md|PIVO-12|medium|P0|-|Stored XSS via innerHTML|RegularTable.svelte:129-231
PENDING|deep-audit-pivot.md|PIVO-13|low|P2|-|Expansion traversal skips gaps|pivot-expansion.ts:75
PENDING|deep-audit-pivot.md|PIVO-14|low|P2|-|Collapsed descendants keep querying|pivot-expansion.ts:203
```
