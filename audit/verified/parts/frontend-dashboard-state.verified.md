# Verified audit: frontend-dashboard-state

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **12 numbered findings**. Every `frontend-dashboard-state-01` through `-12` heading appears exactly once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| frontend-dashboard-state-02 | low | P3 | Compression decoder is ASCII-fragile | `web-common/src/features/dashboards/url-state/compression.ts:11,22` still pairs `TextEncoder` with `String.fromCharCode` | Use `TextDecoder`; add Unicode/large-state round-trip tests. |
| frontend-dashboard-state-03 | medium | P1 | Boolean lock drops a competing state change | `DashboardStateSync.ts:208-299` still returns while `updating` is true and holds the flag across `resolveTimeRanges` | Queue/replay the newest generation after the current sync finishes. |
| frontend-dashboard-state-04 | low | P2 | Explore initialization can race twice | `DashboardStateSync.ts:129-201` checks `initialized` before an await and sets it afterward | Add an initializing promise/token and make init idempotent. |
| frontend-dashboard-state-05 | low | P2 | Metrics-store cache collides and never evicts | `state-managers/memoize-metrics-store.ts:10-15` remains a permanent map keyed by string concatenation | Scope by runtime instance, use a tuple-safe key, and evict on teardown. |
| frontend-dashboard-state-06 | low | P3 | Timestamp seconds contain milliseconds | `proto-state/toProto.ts:218-222` and `fromProto.ts:484-486` retain the internally consistent non-protobuf unit | Serialize real seconds/nanos with a migration/compatibility test. |
| frontend-dashboard-state-07 | low | P2 | TDD session restore admits null measure | `loaders/explore-web-view-store.ts:142-152` still casts `URLSearchParams.get(...)` to string | Validate the measure and fall back to the default view. |
| frontend-dashboard-state-08 | low | P3 | Time-grain cleaning falls through | `url-state/clean-url-params.ts:27-49` retains the special-case fallthrough | Return after the special case or consolidate the rules. |
| frontend-dashboard-state-09 | low | P2 | TDD total header/body use different totals | `time-dimension-details/time-dimension-data-store.ts:227-277` still uses current total in the header and current+comparison in cells | Define one semantic and apply it to header, body, tooltip, and tests. |
| frontend-dashboard-state-10 | low | P2 | Comparison range is not server-resolved | `DashboardStateSync.ts:148-160,241-254` still comments out comparison range resolution | Resolve both ranges together and test relative comparison presets. |
| frontend-dashboard-state-11 | low | P2 | Invalid URL-state errors are discarded | `DashboardStateDataLoader.ts:329-335` still destructures state and drops converter errors | Surface a recoverable banner and telemetry; retain valid fields. |
| frontend-dashboard-state-12 | low | P2 | Replace-state navigation can re-enter sync | `DashboardStateSync.ts:278-299` deliberately releases before `goto`, permitting a redundant pass | Add a navigation generation/expected-URL marker rather than one boolean. |

## Fixed / remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| frontend-dashboard-state-01 | FIXED; remove from fix queue | Guards now run before taking the lock and the locked body is enclosed in `try/finally` at `DashboardStateSync.ts:215-279` | PR #9701, commit `7aa5ebd88` |

## Partial / needs test

None.

Counts: **11 pending, 1 fixed/remove, 0 partial = 12**.

## Machine-readable inventory

```text
FIXED|frontend-dashboard-state.md|frontend-dashboard-state-01|high|-|-|Updating lock leaked on early return|DashboardStateSync.ts:215-279; PR#9701
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-02|low|P3|-|ASCII-fragile decompression|compression.ts:11,22
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-03|medium|P1|-|Boolean sync lock drops changes|DashboardStateSync.ts:208-299
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-04|low|P2|-|Double initialization race|DashboardStateSync.ts:129-201
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-05|low|P2|-|Cache collision and retention|memoize-metrics-store.ts:10-15
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-06|low|P3|-|Milliseconds stored as seconds|toProto.ts:218-222
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-07|low|P2|-|Nullable TDD measure restore|explore-web-view-store.ts:142-152
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-08|low|P3|-|Time-grain cleaning fallthrough|clean-url-params.ts:27-49
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-09|low|P2|-|TDD total semantic mismatch|time-dimension-data-store.ts:227-277
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-10|low|P2|-|Comparison range unresolved|DashboardStateSync.ts:148-160
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-11|low|P2|-|URL-state errors hidden|DashboardStateDataLoader.ts:329-335
PENDING|frontend-dashboard-state.md|frontend-dashboard-state-12|low|P2|-|Replace-state re-entry|DashboardStateSync.ts:278-299
```
