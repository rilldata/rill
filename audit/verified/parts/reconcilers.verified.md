# Verified audit: `reconcilers.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **11** numbered findings, including `reconcilers-11` under Product Improvements. Every source ID is classified once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| reconcilers-04 | MEDIUM | P1 | Report retry can redeliver successful notifiers | `runtime/reconcilers/report.go:463-475,594-612` overwrites `sent` for each notifier | OR-accumulate dirty delivery state and/or persist per-recipient delivery before retry. |
| reconcilers-05 | MEDIUM | P1 | Pending-partition pagination stops after 1000 | Caller combines `WherePending` and `AfterKey` at `runtime/reconcilers/model.go:1637-1665`; SQLite combines `executed_on IS NULL` with comparisons at `runtime/drivers/sqlite/catalog.go:158-172` | Paginate pending rows by stable key/index independently of nullable `executed_on`; test >1000 rows. |
| reconcilers-06 | MEDIUM | P1 | Migration re-executes its applied top version | `runtime/reconcilers/migration.go:76-97` loops inclusively from persisted `State.Version` | Represent an unapplied sentinel and start subsequent runs at `from+1`; add restart/reconcile tests. |
| reconcilers-07 | LOW | P2 | Migration downgrade underflows unsigned delta | `runtime/reconcilers/migration.go:76-83` computes `to-from` before handling `to < from` | Handle downgrade/equality explicitly before a signed/bounds-safe delta. |
| reconcilers-08 | LOW | P2 | Rename comparison mutates parser model definition | `runtime/reconcilers/project_parser.go:624-635,676-690` copies triggers into `def.ModelSpec` | Clone the definition/spec before comparison; test a failed rename followed by insertion. |
| reconcilers-10 | LOW | P3 | Report summary truncation can split UTF-8 | `runtime/reconcilers/report.go:787-792` slices bytes | Truncate on rune boundaries and test non-ASCII input. |
| reconcilers-11 | MEDIUM | P2 | Project-parser watcher retries without backoff | `runtime/reconcilers/project_parser.go:246-258` returns `Retrigger: time.Now()` for watch failures | Add bounded exponential backoff with jitter, reset after a healthy watch interval. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| reconcilers-01 | FIXED | Non-email alert paths guard nullable admin metadata before URL access at `runtime/reconcilers/alert.go:699-700,892-903,927-932`. | PR #9707, `4726d8450` |
| reconcilers-02 | FIXED | Successful rename replaces `prevResult` with `renameRes` at `runtime/reconcilers/model.go:193-196`, so the later existence check targets the new table. | PR #9709, `07d194ccb` |
| reconcilers-03 | FIXED | Deletion now calls partition cleanup; `clearPartitions` delegates to `DeleteModelPartitions` at `runtime/reconcilers/model.go:1127-1135`, with regression coverage. | PR #9711, `06ecbe7e4` |

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| reconcilers-09 | LOW | P2 | PATCH still persists the new hash before execution, but the audit's “lost automatic retry” scenario is overstated because a patch spec change alone does not trigger execution. | `runtime/reconcilers/model.go:1986-2009`; separate pending/applied hashes and test downstream visibility plus failed manual/scheduled patch runs. |

Counts: **7 pending, 3 fixed/remove, 1 partial/needs-test, 0 invalid = 11 total**.

## Machine-readable manifest

```text
FIXED|reconcilers.md|reconcilers-01|HIGH|-|-|Non-email alert dereferenced nil admin metadata|runtime/reconcilers/alert.go:699-700,892-903,927-932; PR#9707
FIXED|reconcilers.md|reconcilers-02|HIGH|-|-|Rename existence check used old model table|runtime/reconcilers/model.go:193-196; PR#9709
FIXED|reconcilers.md|reconcilers-03|MEDIUM|-|-|Deleting partitioned model leaked catalog rows|runtime/reconcilers/model.go:1127-1135; PR#9711
PENDING|reconcilers.md|reconcilers-04|MEDIUM|P1|-|Report retry can redeliver successful notifiers|runtime/reconcilers/report.go:463-475,594-612
PENDING|reconcilers.md|reconcilers-05|MEDIUM|P1|-|Pending partition pagination stops after first batch|runtime/reconcilers/model.go:1637-1665; runtime/drivers/sqlite/catalog.go:158-172
PENDING|reconcilers.md|reconcilers-06|MEDIUM|P1|-|Migration re-executes current version|runtime/reconcilers/migration.go:76-97
PENDING|reconcilers.md|reconcilers-07|LOW|P2|-|Migration downgrade underflows delta|runtime/reconcilers/migration.go:76-83
PENDING|reconcilers.md|reconcilers-08|LOW|P2|-|Rename probe mutates parser definition|runtime/reconcilers/project_parser.go:624-635,676-690
PARTIAL|reconcilers.md|reconcilers-09|LOW|P2|-|Patch mode persists hash before execution|runtime/reconcilers/model.go:1986-2009
PENDING|reconcilers.md|reconcilers-10|LOW|P3|-|Summary truncation can split UTF-8|runtime/reconcilers/report.go:787-792
PENDING|reconcilers.md|reconcilers-11|MEDIUM|P2|-|Project parser watch retries immediately|runtime/reconcilers/project_parser.go:246-258
```
