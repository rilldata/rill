# Verified audit: deep-audit-canvas

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **14 numbered findings** (`CANV-1` through `CANV-14`). Every heading appears exactly once.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| CANV-1 | high | P1 | Component time/filter store leaks subscriptions | `canvas/components/BaseCanvasComponent.ts:285-364` still creates an inner derived subscription without returning cleanup | Flatten the derivation or return the inner unsubscribe; instantiate the store once. |
| CANV-2 | high | P1 | Insert-row move corrupts touched-row width updates | `canvas/layout-util.ts:239` still mutates layout rows without splicing/reindexing parallel touched state | Keep rows and touched flags structurally coupled; regression-test moving upward into a new row. |
| CANV-3 | medium | P1 | Component keeps old metrics-view binding after edit | `BaseCanvasComponent.ts:97,172-181` still initializes `metricsViewName` only in the constructor while component instances are reused | Update/recreate metrics-view-dependent state when `metrics_view` changes. |
| CANV-4 | medium | P2 | Time-range calls repeat for every spec emission/view | `canvas/stores/time-state.ts:158` still derives network resolution from broad spec emissions | Deduplicate/cache by semantic range, zone, and metrics-view timeseries config. |
| CANV-5 | medium | P1 | Canvas pivots share pagination reset key | `canvas/components/pivot/util.ts:88` remains module-global | Keep the key inside each pivot store instance. |
| CANV-6 | medium | P1 | Stale interval response overwrites current range | `canvas/stores/time-state.ts:162,199-206` still resolves asynchronously with no epoch guard | Add generation/cancellation and gate both interval and min/max writes. |
| CANV-7 | medium | P2 | In-list mode survives Select conversion | `canvas/stores/filter-state.ts:309` still discards `.filter()` result | Reassign/splice and cover URL serialization round-trip. |
| CANV-8 | medium | P1 | FilterManager derived stores leak subscriptions | `canvas/stores/filter-manager.ts:145-180` still has unreturned nested subscriptions | Return cleanup or flatten dependencies. |
| CANV-9 | medium | P2 | Leaked filter subscriptions multiply conversion work | `canvas/stores/filter-manager.ts:114-143` still recreates/retains conversion subscribers | Same lifecycle fix as CANV-8 plus one stable conversion store. |
| CANV-10 | medium | P1 | Clearing defaults can dereference missing defaults | `canvas/stores/canvas-entity.ts:498` still assumes the YAML defaults object exists | Initialize/guard defaults and test a canvas with no defaults block. |
| CANV-11 | medium | P2 | Metrics-view combinations grow exponentially | `canvas/stores/filter-manager.ts:1063` still builds all subsets with `getAllCombinations` | Replace with direct indexing by actual component/filter dependencies. |
| CANV-12 | medium | P2 | No-op spec emissions rebuild chart queries | `canvas/components/BaseCanvasComponent.ts:180` still unconditionally sets spec state | Compare semantic spec changes before emitting; preserve providers. |
| CANV-13 | medium | P1 | Markdown template stores leak inner subscriptions | `canvas/components/markdown/util.ts:135-196` still nests subscriptions and repeatedly obtains `timeAndFilterStore` | Flatten/cached derived graph and return cleanup. |
| CANV-14 | low | P3 | Grid diff is defeated by unconditional writes | `canvas/stores/grid.ts:98-115` still writes before/after its comparisons | Compute final row state and emit only changed stores. |

## Fixed / remove

None.

## Partial / needs test

None.

Counts: **14 pending (8 P1, 5 P2, 1 P3), 0 fixed, 0 partial = 14**.

## Machine-readable inventory

```text
PENDING|deep-audit-canvas.md|CANV-1|high|P1|-|timeAndFilterStore subscription leak|BaseCanvasComponent.ts:285-364
PENDING|deep-audit-canvas.md|CANV-2|high|P1|-|Insert-row corrupts touched widths|layout-util.ts:239
PENDING|deep-audit-canvas.md|CANV-3|medium|P1|-|Stale metrics-view binding|BaseCanvasComponent.ts:97,172-181
PENDING|deep-audit-canvas.md|CANV-4|medium|P2|-|Repeated time-range requests|time-state.ts:158
PENDING|deep-audit-canvas.md|CANV-5|medium|P1|-|Shared pivot lastKey|canvas/components/pivot/util.ts:88
PENDING|deep-audit-canvas.md|CANV-6|medium|P1|-|Stale interval race|time-state.ts:162,199-206
PENDING|deep-audit-canvas.md|CANV-7|medium|P2|-|In-list marker retained|filter-state.ts:309
PENDING|deep-audit-canvas.md|CANV-8|medium|P1|-|FilterManager subscription leak|filter-manager.ts:145-180
PENDING|deep-audit-canvas.md|CANV-9|medium|P2|-|Leaked conversions multiply work|filter-manager.ts:114-143
PENDING|deep-audit-canvas.md|CANV-10|medium|P1|-|Missing defaults throws|canvas-entity.ts:498
PENDING|deep-audit-canvas.md|CANV-11|medium|P2|-|Exponential combinations|filter-manager.ts:1063
PENDING|deep-audit-canvas.md|CANV-12|medium|P2|-|No-op spec rebuild|BaseCanvasComponent.ts:180
PENDING|deep-audit-canvas.md|CANV-13|medium|P1|-|Markdown store leak|markdown/util.ts:135-196
PENDING|deep-audit-canvas.md|CANV-14|low|P3|-|Dead grid diff|grid.ts:98-115
```
