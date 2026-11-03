# Verified audit: `resolvers.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **8** numbered findings. Every source ID is classified once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| resolvers-04 | MEDIUM | P1 | `/metrics-sql` ignores documented `additional_where` | Schema advertises it at `runtime/resolvers/builtin_metrics_sql.go:28-36`, but args/forwarding omit it at `:40-43,60-72` | Decode and forward the expression, or remove it from the contract; add endpoint coverage. |
| resolvers-05 | MEDIUM | P1 | SQL cache key omits connector | `runtime/resolvers/sql.go:147-152` keys only on SQL while both DuckDB and ClickHouse are cacheable | Include connector identity/dialect in the key and test same SQL against two connectors. |
| resolvers-06 | LOW | P2 | Union/API resolvers mutate shared args | `runtime/resolvers/union.go:45-60`; `runtime/resolvers/api.go:41-63` | Clone args for every child/proxy before deletion, static-arg merge, or recursion tracking. |
| resolvers-07 | LOW | P3 | Resolver export surface is unreachable | `ForExport` is only false/pass-through (`runtime/resolver.go:160`; `runtime/server/query_resolver.go:55`); no `.ResolveExport(` caller exists | Wire an export dispatch and tests, or delete the dead interface/plumbing. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| resolvers-01 | FIXED | Nil time ranges are allocated before dereference at `runtime/resolvers/ai.go:79-91`, with regression coverage. | PR #9703, `7fd2bf9c7` |
| resolvers-02 | FIXED | Cache hashing now serializes full `SecurityClaims` (including additional rules, permissions, and skip-checks) at `runtime/resolver.go:208-220`. | PR #9704, `4564e65ac` |
| resolvers-03 | FIXED | Annotation initialization now resolves security and rejects `!CanAccess()` at `runtime/resolvers/annotations.go:51-58`. | PR #9705, `9367e47a2` |

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| resolvers-08 | LOW | P3 | The dereference-before-check remains, but successful aggregation resolution guarantees a non-nil result, so the claimed runtime panic is not currently reachable. | `runtime/resolvers/legacy_metrics.go:129-145`; move the dereference inside the guard and add a defensive nil-result test. |

Counts: **4 pending, 3 fixed/remove, 1 partial/needs-test, 0 invalid = 8 total**.

## Machine-readable manifest

```text
FIXED|resolvers.md|resolvers-01|HIGH|-|-|AI resolver omitted time ranges panic|runtime/resolvers/ai.go:79-91; PR#9703
FIXED|resolvers.md|resolvers-02|HIGH|-|-|Query cache omitted security context|runtime/resolver.go:208-220; PR#9704
FIXED|resolvers.md|resolvers-03|HIGH|-|-|Annotations bypassed metrics-view access|runtime/resolvers/annotations.go:51-58; PR#9705
PENDING|resolvers.md|resolvers-04|MEDIUM|P1|-|Metrics SQL API ignores additional_where|runtime/resolvers/builtin_metrics_sql.go:28-43,60-72
PENDING|resolvers.md|resolvers-05|MEDIUM|P1|-|SQL cache key omits connector|runtime/resolvers/sql.go:147-152
PENDING|resolvers.md|resolvers-06|LOW|P2|-|Union and API resolvers mutate shared args|runtime/resolvers/union.go:45-60; runtime/resolvers/api.go:41-63
PENDING|resolvers.md|resolvers-07|LOW|P3|-|Resolver export path is unreachable|runtime/resolver.go:160; runtime/server/query_resolver.go:55
PARTIAL|resolvers.md|resolvers-08|LOW|P3|-|Legacy metrics dereferences before dead nil guard|runtime/resolvers/legacy_metrics.go:129-145
```
