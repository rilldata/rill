# Verified: admin-database.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **9** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| admin-database-02 | LOW | P3 | `ifNotExists` branch uses unnormalized attributes | `admin/database/postgres/postgres.go:2099-2136`; one branch inserts `attrs`, the conflict branch inserts raw `attributes` | Use the validated/normalized value in both branches. |
| admin-database-04 | LOW | P2 | Dev deployments can be stopped by prod TTL | `admin/database/postgres/postgres.go:647-660`; prod-TTL clause has no environment predicate | Scope prod TTL to non-dev and dev TTL to dev explicitly. |
| admin-database-05 | LOW | P2 | Dynamic variable INSERT lacks empty/case-duplicate guard | `admin/database/postgres/postgres.go:3214-3262`; conflict target is lowercased but input keys are not deduplicated | Reject empty input and deduplicate by normalized lowercase name before SQL construction. |
| admin-database-06 | LOW | P1 | Database-state assertions panic admin process | `admin/database/postgres/postgres.go:1050,2133,2167,3971,3989`; row-count conditions still call `panic` | Return typed corruption/invariant errors and alert without crashing the service. |
| admin-database-07 | LOW | P3 | Dead access-request finder returns incomplete rows | `admin/database/postgres/postgres.go:2772-2794`; implementation selects only user ID and has no caller | Remove the dead interface/method or select full rows and expose a bounded use case. |
| admin-database-08 | LOW | P3 | Virtual-file listing is unbounded; FindUsers is dead | `admin/database/postgres/postgres.go:830,2921-2933`; virtual files have no limit/cursor; `FindUsers` has no caller | Add keyset pagination to the live finder and delete the dead method. |
| admin-database-09 | LOW | P2 | Error mapping and migration detection are fragile | `admin/database/postgres/postgres.go:3264-3284`; raw delete error; `admin/database/postgres/migrate.go:24-32` matches error text | Apply `parseErr`; inspect PostgreSQL SQLSTATE `42P01`. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| — | None | No relevant merged changes are present after the audit snapshot. | — |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| admin-database-01 | MEDIUM | P3 | LATENT API DEFECT | Four billing readers at `admin/database/postgres/postgres.go:3080-3125` still call `c.db` instead of `c.getDB(ctx)`, but no current caller uses them inside a transaction. Correct them and add a transaction-visibility unit test. |
| admin-database-03 | MEDIUM | P1 | NEEDS MULTI-CONNECTION TEST | `admin/database/postgres/migrate.go:38-50` acquires/unlocks a session advisory lock through the pool while migrations open separate transactions. Pin one connection; add a pool-contention migration test. |

Counts: **7 pending, 0 fixed/remove, 2 partial/needs-test, 0 invalid = 9 total**.

## Machine-readable ledger

```text
NEEDS_TEST|admin-database.md|admin-database-01|MEDIUM|P3|-|Billing readers bypass transaction context|admin/database/postgres/postgres.go:3080-3125
PENDING|admin-database.md|admin-database-02|LOW|P3|-|Normalized attributes discarded|admin/database/postgres/postgres.go:2099-2136
NEEDS_TEST|admin-database.md|admin-database-03|MEDIUM|P1|-|Migration advisory lock uses connection pool|admin/database/postgres/migrate.go:38-50
PENDING|admin-database.md|admin-database-04|LOW|P2|-|Prod TTL can hibernate dev deployment|admin/database/postgres/postgres.go:647-660
PENDING|admin-database.md|admin-database-05|LOW|P2|-|Dynamic project-variable insert guards missing|admin/database/postgres/postgres.go:3214-3262
PENDING|admin-database.md|admin-database-06|LOW|P1|-|Persistence layer panics on database state|admin/database/postgres/postgres.go:1050,2133,2167,3971,3989
PENDING|admin-database.md|admin-database-07|LOW|P3|-|Dead incomplete access-request finder|admin/database/postgres/postgres.go:2772-2794
PENDING|admin-database.md|admin-database-08|LOW|P3|-|Unbounded virtual-file listing and dead FindUsers|admin/database/postgres/postgres.go:830,2921-2933
PENDING|admin-database.md|admin-database-09|LOW|P2|-|Fragile database error handling|admin/database/postgres/postgres.go:3264-3284; admin/database/postgres/migrate.go:24-32
```
