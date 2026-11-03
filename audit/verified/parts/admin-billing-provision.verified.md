# Verified: admin-billing-provision.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **12** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| abp-02 | HIGH | P1 | Static runtime slot allocation is a TOCTOU race | `admin/provisioner/static/static.go:94-119`; the capacity read and assignment upsert remain separate and unlocked | Make capacity check plus assignment atomic with a transaction/advisory lock or guarded SQL insert. |
| abp-03 | MEDIUM | P1 | Credit-depletion retry can skip remaining hibernation | `admin/jobs/river/credit_balance_event_handlers.go:110-198`; issue transition commits before hibernation/email, and retry no-ops after `OnCreditTrial` is gone | Make the depleted state resumable and hibernation independently idempotent. |
| abp-04 | MEDIUM | P1 | Org teardown/hibernation processes only 100 projects | `admin/jobs/river/org_jobs.go:210,251`; both flows make one `FindProjectsForOrganization(..., 100)` call | Loop with keyset pagination or fan out per-project jobs. |
| abp-05 | LOW | P2 | Trial-end email runs inside DB transaction | `admin/jobs/river/trial_checks.go:147-200`; `SendTrialEnded` precedes `tx.Commit` | Commit the state transition first, then send best-effort email outside the transaction. |
| abp-06 | LOW | P2 | Used flusher shutdown does not join background loop | `admin/used_flusher.go:113-133`; `Close` waits on a group that `flush` increments, but not on the loop goroutine | Give the loop its own done signal/WaitGroup, join it, then perform a final synchronous flush. |
| abp-09 | LOW | P3 | Orb credit and plan reads do not paginate | `admin/billing/orb.go:305-317,654-658`; both consume one page and plans retain a pagination TODO | Use Orb auto-paging iterators for credits and plans. |
| abp-10 | LOW | P2 | Invoice issue updates race across jobs | `admin/jobs/river/biller_event_handlers.go:36-79,115-162`; both workers read-modify-write whole metadata without locking | Merge under a row lock/transaction or serialize by organization. |
| abp-11 | LOW | P2 | DeleteOrg and PlanChanged jobs lack uniqueness | `admin/jobs/river/river.go:631-657`; both inserts still use bare `river.InsertOpts` | Add `UniqueOpts{ByArgs:true}` with suitable state scoping. |
| abp-12 | LOW | P3 | Autoscaler is permanently disabled dead code | `admin/jobs/river/run_autoscaler.go:20,39-42`; compile-time `disableAutoscaler = true` always exits | Remove the worker until ready, or enable it only through real configuration after hardening/tests. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| abp-01 | FIXED — remove | `admin/provisioner/clickhousestatic/provisioner.go:335-350` now places the normalized unique resource ID before human-readable org/project suffixes; regression tests assert distinct long names | PR #9683, commit `fbcb248470347d7809614a26534d0bf5f17cc916` |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| abp-07 | LOW | P3 | VALID BUT DORMANT | Nil dereferences remain at `admin/jobs/river/run_autoscaler.go:55-64`, but `disableAutoscaler = true` at line 20 makes the branch unreachable. Fix before enabling the worker; do not describe it as a current production panic. |
| abp-13 | LOW | P3 | NEEDS CONCURRENCY TEST | Independent trial-start email senders remain at `admin/billing.go:209-215` and `admin/jobs/river/org_jobs.go:155-164`; duplication requires an operator repair racing a retrying start-trial job. Add a deterministic interleaving test before promoting this to a confirmed bug. |

Counts: **9 pending, 1 fixed/remove, 2 partial/needs-test, 0 invalid = 12 total**.

## Machine-readable ledger

```text
FIXED|admin-billing-provision.md|abp-01|HIGH|-|-|ClickHouse static database-name collision|admin/provisioner/clickhousestatic/provisioner.go:335-350; PR #9683 fbcb248470347d7809614a26534d0bf5f17cc916
PENDING|admin-billing-provision.md|abp-02|HIGH|P1|-|Static runtime slot allocation TOCTOU|admin/provisioner/static/static.go:94-119
PENDING|admin-billing-provision.md|abp-03|MEDIUM|P1|-|Credit-depletion retry skips hibernation|admin/jobs/river/credit_balance_event_handlers.go:110-198
PENDING|admin-billing-provision.md|abp-04|MEDIUM|P1|-|Org jobs stop after first 100 projects|admin/jobs/river/org_jobs.go:210,251
PENDING|admin-billing-provision.md|abp-05|LOW|P2|-|Trial-end email inside DB transaction|admin/jobs/river/trial_checks.go:147-200
PENDING|admin-billing-provision.md|abp-06|LOW|P2|-|Used flusher shutdown does not join loop|admin/used_flusher.go:113-133
NEEDS_TEST|admin-billing-provision.md|abp-07|LOW|P3|-|Dormant autoscaler nil dereferences|admin/jobs/river/run_autoscaler.go:20,55-64
PENDING|admin-billing-provision.md|abp-09|LOW|P3|-|Orb reads do not paginate|admin/billing/orb.go:305-317,654-658
PENDING|admin-billing-provision.md|abp-10|LOW|P2|-|Invoice metadata update race|admin/jobs/river/biller_event_handlers.go:36-79,115-162
PENDING|admin-billing-provision.md|abp-11|LOW|P2|-|Billing/org jobs lack uniqueness|admin/jobs/river/river.go:631-657
PENDING|admin-billing-provision.md|abp-12|LOW|P3|-|Autoscaler dead code|admin/jobs/river/run_autoscaler.go:20,39-42
NEEDS_TEST|admin-billing-provision.md|abp-13|LOW|P3|-|Trial-start email can double-send in repair race|admin/billing.go:209-215; admin/jobs/river/org_jobs.go:155-164
```
