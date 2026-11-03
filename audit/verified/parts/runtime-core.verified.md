# Verified audit: `runtime-core.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **9** numbered findings (the source has no `runtime-core-09`). Every finding is classified once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| runtime-core-02 | HIGH | P1 | Failed instance deletion waits on a nil channel | `runtime/registry.go:114-120,398-415` | Return on delete error; wait only on a non-nil channel with context cancellation. |
| runtime-core-03 | MEDIUM | P2 | One mutex covers all security template evaluation | `runtime/security.go:275-342` | Use double-checked cache access and compute outside the global lock, or shard/singleflight by key. |
| runtime-core-06 | LOW | P2 | `markPending` clears ancestor instead of descendant | `runtime/controller.go:1161-1175`, especially `:1169` uses `n` while status uses `dn` | Pass `dn`; add a descendant-with-old-error test. |
| runtime-core-07 | LOW | P2 | `Repo` returns an already-called release on error | `runtime/connections.go:62-79` | Return a nil release after releasing, matching sibling accessors. |
| runtime-core-10 | LOW | P2 | Transitive access expands before every cache lookup | `runtime/security.go:255-283,783-819` | Cache expanded rules or include their stable expansion key in the security cache/singleflight. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| runtime-core-01 | FIXED | Resolver caching now hashes full `SecurityClaims` at `runtime/resolver.go:208-220`, covering `AdditionalRules`, permissions, and `SkipChecks`. | PR #9704, `4564e65ac` |

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| runtime-core-04 | LOW | P1 | Empty user IDs still match a missing owner annotation, but exploitation needs an unusual annotation plus empty-ID principal. | `runtime/security.go:443-452,519-528`; require non-empty `userID` and test absent owner. |
| runtime-core-05 | LOW | P3 | Resource kind is still absent from the security cache key, but collision also needs same name and identical nanosecond state timestamp. | `runtime/security.go:874-900`; hash kind and add a constructed collision test. |
| runtime-core-08 | LOW | P2 | Reconciler completion can leak only when work remains hung past the 30-second shutdown drain. | Unbuffered channel at `runtime/controller.go:148`, timed drain `:320-350`, send `:1335`; make completion delivery non-blocking after shutdown and test a hung reconciler. |

Counts: **5 pending, 1 fixed/remove, 3 partial/needs-test, 0 invalid = 9 total**.

## Machine-readable manifest

```text
FIXED|runtime-core.md|runtime-core-01|HIGH|-|-|Resolver cache omitted AdditionalRules|runtime/resolver.go:208-220; PR#9704
PENDING|runtime-core.md|runtime-core-02|HIGH|P1|-|Failed instance delete waits on nil channel|runtime/registry.go:114-120,398-415
PENDING|runtime-core.md|runtime-core-03|MEDIUM|P2|-|Global security mutex covers template evaluation|runtime/security.go:275-342
PARTIAL|runtime-core.md|runtime-core-04|LOW|P1|-|Empty user ID can match missing owner annotation|runtime/security.go:443-452,519-528
PARTIAL|runtime-core.md|runtime-core-05|LOW|P3|-|Security cache key omits resource kind|runtime/security.go:874-900
PENDING|runtime-core.md|runtime-core-06|LOW|P2|-|markPending clears ancestor error|runtime/controller.go:1161-1175
PENDING|runtime-core.md|runtime-core-07|LOW|P2|-|Repo returns already-called release on error|runtime/connections.go:62-79
PARTIAL|runtime-core.md|runtime-core-08|LOW|P2|-|Hung reconciler can leak on completion send|runtime/controller.go:148,320-350,1335
PENDING|runtime-core.md|runtime-core-10|LOW|P2|-|Transitive access re-expands outside cache|runtime/security.go:255-283,783-819
```
