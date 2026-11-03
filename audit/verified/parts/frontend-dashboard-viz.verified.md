# Verified audit: frontend-dashboard-viz

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **12 numbered findings** (the source intentionally has no `-10`). All appear once.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| frontend-dashboard-viz-01 | MEDIUM | P1 | Big number hides comparison when baseline is zero | `big-number/MeasureBigNumber.svelte:150,254` still uses truthiness for `comparisonValue` | Use explicit null/undefined checks and define zero-baseline delta behavior. |
| frontend-dashboard-viz-02 | MEDIUM | P1 | Dimension export ignores dimension sort | `dimension-table/dimension-table-export.ts:103-148` still builds only measure/delta sorts | Reuse the on-screen `getSort` mapping in export requests. |
| frontend-dashboard-viz-03 | MEDIUM | P1 | Pivot export drops measure-column sort | `pivot/pivot-export.ts:155-181` only matches sorting IDs to row dimensions | Translate measure/column accessors using the same resolver as the table. |
| frontend-dashboard-viz-04 | LOW | P2 | Leaderboard observer is not disconnected | `leaderboard/Leaderboard.svelte:93-110` creates an observer without destroy cleanup | Disconnect/unobserve in `onDestroy`. |
| frontend-dashboard-viz-05 | LOW | P2 | Leaderboard max drops zero / yields negative infinity | `leaderboard/leaderboard-utils.ts:350-351` still filters by truthiness and spreads possibly empty arrays | Preserve zero and return a finite fallback for no values. |
| frontend-dashboard-viz-06 | LOW | P1 | Percent-of-total divides by zero | `leaderboard/leaderboard-utils.ts:105-106,215-216` checks null but not zero/undefined | Require a finite non-zero total; render no-data otherwise. |
| frontend-dashboard-viz-07 | LOW | P3 | Pivot helper mutates shared dimensions with reverse | `pivot/pivot-column-definition.ts:144,540` still calls mutating `.reverse()` | Use `toReversed()` or `[...array].reverse()`. |
| frontend-dashboard-viz-08 | LOW | P2 | Empty time-series rows share one array | `time-series/utils.ts:145-147` still fills with a shared array reference | Allocate one array per row and add mutation-isolation tests. |
| frontend-dashboard-viz-09 | LOW | P3 | Width estimator has dead sort and empty infinity | `dimension-table/dimension-table-utils.ts:156-157` retains the dead sort/empty max path | Remove dead work and return a bounded default width. |
| frontend-dashboard-viz-11 | LOW | P2 | Filter merge can duplicate an expression | `pivot/pivot-merge-filters.ts:93-114` retains the falsy-expression push paths | Normalize expressions before merging and dedupe structurally. |
| frontend-dashboard-viz-12 | LOW | P2 | Zero-dimension search reports NaN progress | `dimension-search/useDimensionSearchResults.ts:80-81` divides by dimension count | Return 100%/idle explicitly for an empty workload. |
| frontend-dashboard-viz-13 | LOW | P3 | Leaderboard skeleton column count is wrong | `leaderboard/Leaderboard.svelte:295-302` still miscounts optional context columns | Derive skeleton columns from the same visible-column list. |

## Fixed / remove

None.

## Partial / needs test

None. PR #9763 changed pivot click-filter scope, but it does not address any finding in this source.

Counts: **12 pending, 0 fixed, 0 partial = 12**.

## Machine-readable inventory

```text
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-01|MEDIUM|P1|-|Big-number zero comparison hidden|MeasureBigNumber.svelte:150,254
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-02|MEDIUM|P1|-|Dimension export ignores sort|dimension-table-export.ts:103-148
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-03|MEDIUM|P1|-|Pivot export drops measure sort|pivot-export.ts:155-181
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-04|LOW|P2|-|Observer not disconnected|Leaderboard.svelte:93-110
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-05|LOW|P2|-|Max drops zero|leaderboard-utils.ts:350-351
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-06|LOW|P1|-|Percent total divides by zero|leaderboard-utils.ts:105-106,215-216
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-07|LOW|P3|-|Shared array reverse mutation|pivot-column-definition.ts:144,540
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-08|LOW|P2|-|Shared empty-data arrays|time-series/utils.ts:145-147
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-09|LOW|P3|-|Broken empty width estimate|dimension-table-utils.ts:156-157
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-11|LOW|P2|-|Duplicate filter merge|pivot-merge-filters.ts:93-114
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-12|LOW|P2|-|NaN search progress|useDimensionSearchResults.ts:80-81
PENDING|frontend-dashboard-viz.md|frontend-dashboard-viz-13|LOW|P3|-|Skeleton column miscount|Leaderboard.svelte:295-302
```
