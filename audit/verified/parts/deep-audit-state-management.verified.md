# Verified audit: deep-audit-state-management

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **13 numbered findings** (`STAT-1` through `STAT-13`). Every heading appears exactly once. STAT-2 and STAT-3 are both fixed by one merged change and remain separate inventory rows.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| STAT-1 | high | P1 | Malformed compressed state can crash dashboard load | `url-state/convertURLToExplorePreset.ts:85` still decompresses/parses without containment/size guard | Bound decoded size and catch/degrade to defaults with a visible warning. |
| STAT-4 | medium | P1 | LIKE expression survives conversion to IN | `dashboards/stores/Filters.ts:312` still appends the IN expression without reliably replacing old LIKE state | Replace the dimension's prior expression atomically; add Contains→Select tests. |
| STAT-5 | medium | P1 | FilterManager leaks multiply recomputation | `canvas/stores/filter-manager.ts:114-143` still retains nested derived subscriptions | Flatten derivations or return inner cleanup. |
| STAT-6 | medium | P2 | Canvas builds every metrics-view subset | `canvas/stores/filter-manager.ts:1063` still runs exponential `getAllCombinations` | Index only combinations actually referenced by components/filters. |
| STAT-7 | medium | P2 | Canvas in-list marker is not removed | `canvas/stores/filter-state.ts:309` still discards `.filter()` result | Reassign/splice and test URL round trip. |
| STAT-8 | medium | P1 | Stale time interval response can win | `canvas/stores/time-state.ts:162,199-206` still has no request generation | Add epoch/cancellation guards. |
| STAT-9 | medium | P1 | Metrics-store cache pins first managers/runtime | `state-managers/memoize-metrics-store.ts:10-15` remains unscoped and unbounded | Key by instance and dispose/evict on provider teardown. |
| STAT-10 | medium | P1 | Canvas snapshot state crosses project boundary | `canvas/stores/canvas-entity.ts:54,599,660` still keys module map by canvas name | Include runtime instance/project in key and clear on disposal. |
| STAT-11 | low | P2 | FilterManager nested stores never unsubscribe | `canvas/stores/filter-manager.ts:145-198` still discards inner unsubscribers | Same lifecycle fix as STAT-5. |
| STAT-12 | low | P2 | Time summary recreates query observer broadly | `state-managers/state-managers.ts:119` still derives from the whole dashboard store | Select stable time/metrics inputs only. |
| STAT-13 | low | P2 | Legacy proto deserialization can throw | `proto-state/fromProto.ts:365` still trusts crafted enum/grain/state combinations at exported call sites | Validate and catch at the public boundary; skip invalid fields. |

## Fixed / remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| STAT-2 | FIXED; remove | `DashboardStateSync.ts:215-279` now evaluates guards before locking and releases the lock in `finally` | PR #9701, commit `7aa5ebd88` |
| STAT-3 | FIXED; remove | Same fix contains storage/other throws inside the guarded `try/finally`; missing-spec returns occur before lock acquisition | PR #9701, commit `7aa5ebd88` |

## Partial / needs test

None.

Counts: **11 pending, 2 fixed/remove, 0 partial = 13**.

## Machine-readable inventory

```text
PENDING|deep-audit-state-management.md|STAT-1|high|P1|-|Malformed compressed state crash|convertURLToExplorePreset.ts:85
FIXED|deep-audit-state-management.md|STAT-2|medium|-|-|Early-return lock leak|DashboardStateSync.ts:215-279; PR#9701
FIXED|deep-audit-state-management.md|STAT-3|medium|-|-|Spec/storage lock leak|DashboardStateSync.ts:215-279; PR#9701
PENDING|deep-audit-state-management.md|STAT-4|medium|P1|-|LIKE remains after IN conversion|Filters.ts:312
PENDING|deep-audit-state-management.md|STAT-5|medium|P1|-|Filter recomputation leak|filter-manager.ts:114-143
PENDING|deep-audit-state-management.md|STAT-6|medium|P2|-|Exponential combinations|filter-manager.ts:1063
PENDING|deep-audit-state-management.md|STAT-7|medium|P2|-|In-list marker retained|filter-state.ts:309
PENDING|deep-audit-state-management.md|STAT-8|medium|P1|-|Stale interval race|time-state.ts:162,199-206
PENDING|deep-audit-state-management.md|STAT-9|medium|P1|-|Unbounded stale cache|memoize-metrics-store.ts:10-15
PENDING|deep-audit-state-management.md|STAT-10|medium|P1|-|Cross-project canvas state|canvas-entity.ts:54,599,660
PENDING|deep-audit-state-management.md|STAT-11|low|P2|-|Nested store leak|filter-manager.ts:145-198
PENDING|deep-audit-state-management.md|STAT-12|low|P2|-|Broad query observer rebuild|state-managers.ts:119
PENDING|deep-audit-state-management.md|STAT-13|low|P2|-|Unsafe legacy state deserialize|fromProto.ts:365
```
