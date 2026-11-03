# Verified: admin-auth-rbac.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **13** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| admin-auth-rbac-03 | MEDIUM | P0 | Magic link inherits creator’s privileged attributes | `admin/server/magic_tokens.go:90-100`; `admin/server/server.go:364-390`; `admin/server/runtime_jwt.go:113-120`; creator attributes including `admin` remain stored and replayed | Mint an explicit non-admin attribute set for share tokens; do not inherit privilege-bearing custom attributes. |
| admin-auth-rbac-02 | LOW | P2 | Revoked tokens remain cached for 10 seconds | `admin/auth_token.go:249-312`; `PurgeAuthTokenCache` has only two test call sites and no production revoke path | Evict the hashed token on every revoke/delete/rotation, or version the cache. |
| admin-auth-rbac-13 | LOW | P2 | Transient validation error logs browser out | `admin/server/auth/middleware.go:98-112`; any validation error still deletes the access-token cookie | Clear only on definitive invalid/expired/not-found errors; preserve session on infrastructure errors. |
| admin-auth-rbac-07 | LOW | P2 | Transient token errors are negatively cached | `admin/auth_token.go:298-311`; all non-context errors enter the ten-second cache | Cache only definitive invalid-token outcomes. |
| admin-auth-rbac-05 | LOW | P3 | OAuth anonymous redirect falls through | `admin/server/auth/handlers.go:657-665`; redirect at line 661 still has no return | Return immediately after redirect. |
| admin-auth-rbac-09 | LOW | P2 | DeleteUser permits account enumeration | `admin/server/users.go:525-537`; user lookup/not-found precedes authorization | Perform coarse authorization first or normalize missing/forbidden responses. |
| admin-auth-rbac-06 | LOW | P3 | Token hash equality is not constant-time | `admin/auth_token.go:331,352,370,387`; all secret hashes use `bytes.Equal` | Use `subtle.ConstantTimeCompare`/`hmac.Equal`. |
| admin-auth-rbac-12 | LOW | P2 | API page size is unbounded | `admin/server/page_token.go:70-75`; every nonzero uint32 is converted directly to int | Clamp to a service-wide maximum. |
| admin-auth-rbac-08 | LOW | P2 | Device flow embeds an unvalidated redirect | `admin/server/auth/device_code.go:80-90`; request value is copied to the verification URL; frontend ultimately navigates to it | Validate with the canonical safe-redirect policy before returning the URL and again before navigation. |
| admin-auth-rbac-10 | LOW | P1 | Member manager can grant project admin through group | `admin/server/usergroups.go:583-613`; only org-member/org-admin authority is checked and project roles are explicitly not checked | Require authority over every effective project role before adding a group member. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| — | None | A similarly named hardening commit exists on another local ref, but it is **not an ancestor of `origin/main`**; current main retains these paths. | — |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| admin-auth-rbac-01 | LOW | P3 | LATENT FOOTGUN | `admin/server/auth/claims.go:201-241` returns cached permission pointers and callers mutate them (`runtime_proxy.go:81-84`), but claims are request-scoped today. Return clones and add a race/alias test. |
| admin-auth-rbac-04 | LOW | P1 | NEEDS POLICY DECISION | `admin/server/runtime_proxy.go:40-93` resolves branch deployments but checks/forces only `ReadProd`; public branches are therefore anonymous. Decide whether public means primary-only; then enforce environment-appropriate permission. |
| admin-auth-rbac-11 | LOW | P2 | NEEDS CONCURRENCY TEST | `admin/user.go:168-189` does `CheckUsersEmpty` then insert in an ordinary transaction. Add a two-transaction first-signup test and serialize/bootstrap explicitly. |

Counts: **10 pending, 0 fixed/remove, 3 partial/needs-test, 0 invalid = 13 total**.

## Machine-readable ledger

```text
PENDING|admin-auth-rbac.md|admin-auth-rbac-03|MEDIUM|P0|-|Magic token replays creator admin attributes|admin/server/magic_tokens.go:90-100; admin/server/runtime_jwt.go:113-120
NEEDS_TEST|admin-auth-rbac.md|admin-auth-rbac-01|LOW|P3|-|Cached permission pointer mutation|admin/server/auth/claims.go:201-241; admin/server/runtime_proxy.go:81-84
PENDING|admin-auth-rbac.md|admin-auth-rbac-02|LOW|P2|-|Revoked token remains cached|admin/auth_token.go:249-312
NEEDS_TEST|admin-auth-rbac.md|admin-auth-rbac-04|LOW|P1|-|Runtime proxy uses production permission for branches|admin/server/runtime_proxy.go:40-93
PENDING|admin-auth-rbac.md|admin-auth-rbac-13|LOW|P2|-|Transient validation logs browser out|admin/server/auth/middleware.go:98-112
PENDING|admin-auth-rbac.md|admin-auth-rbac-07|LOW|P2|-|Transient token errors negatively cached|admin/auth_token.go:298-311
PENDING|admin-auth-rbac.md|admin-auth-rbac-05|LOW|P3|-|OAuth redirect missing return|admin/server/auth/handlers.go:657-665
PENDING|admin-auth-rbac.md|admin-auth-rbac-09|LOW|P2|-|DeleteUser email enumeration|admin/server/users.go:525-537
PENDING|admin-auth-rbac.md|admin-auth-rbac-06|LOW|P3|-|Non-constant-time token comparison|admin/auth_token.go:331,352,370,387
PENDING|admin-auth-rbac.md|admin-auth-rbac-12|LOW|P2|-|Unbounded page size|admin/server/page_token.go:70-75
PENDING|admin-auth-rbac.md|admin-auth-rbac-08|LOW|P2|-|Unvalidated device redirect|admin/server/auth/device_code.go:80-90
PENDING|admin-auth-rbac.md|admin-auth-rbac-10|LOW|P1|-|Group membership indirectly grants project admin|admin/server/usergroups.go:583-613
NEEDS_TEST|admin-auth-rbac.md|admin-auth-rbac-11|LOW|P2|-|First-user superuser race|admin/user.go:168-189
```
