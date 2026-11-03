# Verified audit: frontend-canvas

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **10 numbered findings** (`frontend-canvas-01` through `-10`). Every numbered heading is represented exactly once below. The six unnumbered feature-idea bullets are backlog suggestions, not audit findings, and are not counted.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| frontend-canvas-02 | MEDIUM | P2 | In-list marker survives conversion to Select | `web-common/src/features/canvas/stores/filter-state.ts:308-326` still discards the result of `dimensionsWithInListFilter.filter(...)` | Reassign/splice the filtered array and add a URL round-trip test. |
| frontend-canvas-03 | MEDIUM | P1 | Stale async interval can overwrite the latest range | `web-common/src/features/canvas/stores/time-state.ts:127-209` still resolves `Promise.all(...).then(set)` without an epoch/cancellation guard | Add a generation token and ignore stale interval/min-max responses. |
| frontend-canvas-04 | MEDIUM | P1 | Last canvas state is keyed only by canvas name | `web-common/src/features/canvas/stores/canvas-entity.ts:54,599,660` still uses module-level `lastVisitedState` with `this.name` | Key by instance/project plus canvas name and clear on disposal. |
| frontend-canvas-05 | MEDIUM | P1 | Nested derived stores leak subscriptions | `web-common/src/features/canvas/components/BaseCanvasComponent.ts:285-364`; `stores/filter-manager.ts:114-198`; `components/markdown/util.ts:135-196` still contain unreturned inner `.subscribe(set)` calls | Return inner unsubscribers or flatten dependencies; cache `timeAndFilterStore`. |
| frontend-canvas-06 | LOW | P2 | Save-default subscription can remain live | `web-common/src/features/canvas/stores/canvas-entity.ts:476-484` unsubscribes only after a second emission | Navigate from the computed target directly or guarantee cleanup/timeout. |
| frontend-canvas-07 | LOW | P2 | Image object URLs are never revoked | `web-common/src/features/canvas/components/image/Image.svelte:33-60` creates object URLs without `URL.revokeObjectURL` | Revoke on replacement/destroy and guard stale fetch completions. |
| frontend-canvas-08 | LOW | P3 | Grid update emits redundant writes | `web-common/src/features/canvas/stores/grid.ts:98-115` still writes items/widths twice | Compute final values once, compare, then emit at most once. |
| frontend-canvas-09 | LOW | P3 | Dead grid methods have inconsistent indexing | `web-common/src/features/canvas/stores/row.ts:12-53`; `stores/grid.ts:20-74` retain unused 0/1-based mutation helpers | Remove them or normalize indexing and add tests before reuse. |
| frontend-canvas-10 | LOW | P2 | Raw image asset path is interpolated into URL | `web-common/src/features/canvas/components/image/Image.svelte:47` still interpolates `url` without segment encoding/validation | Encode the asset key as a path segment and reject traversal/query fragments. |

## Fixed / remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| frontend-canvas-01 | FIXED; remove from fix queue | `CanvasPdfExportView.svelte:21-61` now iterates `$layout`; `CanvasPdfExportTabGroup.svelte` and `CanvasPdfExportTab.svelte` render tab contents | PR #9722, commit `0b6b50df3b145238309c673ddd957968709dd62d` |

## Partial / needs test

None.

Counts: **9 pending, 1 fixed/remove, 0 partial, 0 invalid = 10**.

## Machine-readable inventory

```text
FIXED|frontend-canvas.md|frontend-canvas-01|HIGH|-|-|PDF export omitted tab-group content|CanvasPdfExportView.svelte:21-61; PR#9722
PENDING|frontend-canvas.md|frontend-canvas-02|MEDIUM|P2|-|In-list marker survives Select conversion|filter-state.ts:308-326
PENDING|frontend-canvas.md|frontend-canvas-03|MEDIUM|P1|-|Stale async interval overwrites latest|time-state.ts:127-209
PENDING|frontend-canvas.md|frontend-canvas-04|MEDIUM|P1|-|Canvas state key leaks across instances|canvas-entity.ts:54,599,660
PENDING|frontend-canvas.md|frontend-canvas-05|MEDIUM|P1|-|Nested derived subscriptions leak|BaseCanvasComponent.ts:285-364
PENDING|frontend-canvas.md|frontend-canvas-06|LOW|P2|-|Save-default subscription can leak|canvas-entity.ts:476-484
PENDING|frontend-canvas.md|frontend-canvas-07|LOW|P2|-|Image object URLs not revoked|Image.svelte:33-60
PENDING|frontend-canvas.md|frontend-canvas-08|LOW|P3|-|Redundant grid writes|grid.ts:98-115
PENDING|frontend-canvas.md|frontend-canvas-09|LOW|P3|-|Dead inconsistent grid methods|row.ts:12-53
PENDING|frontend-canvas.md|frontend-canvas-10|LOW|P2|-|Raw image asset URL path|Image.svelte:47
```
