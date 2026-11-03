# Verified: deep-cloud-deploy.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **14** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| deep-cloud-deploy-01 | HIGH | P1 | First deploy repeats pulls/parses/archive download | `runtime/registry.go:339-478`; `runtime/config_reloader.go:132-163`; archive restart remains forced by `cfg.UsesArchive` at line 149 | Key archive sync by archive ID, skip duplicate create-time reconcile/pull, and avoid restarting a fresh archive controller. |
| deep-cloud-deploy-02 | HIGH | P1 | Global config-reloader mutex serializes every instance | `runtime/config_reloader.go:27-68,188-217`; one lock spans network/config/repo/controller work | Lock only the map or shard per instance; reload instances with bounded parallelism. |
| deep-cloud-deploy-03 | MEDIUM | P1 | Reconcile-job dedup can drop a desired-state edge | `admin/jobs/river/river.go:342-350`; `reconcile_deployment.go:50-51,86-89,130-146`; running jobs participate in uniqueness and an early branch omits final recheck | Exclude running from uniqueness or complete/recheck transactionally; observe duplicate insert results. |
| deep-cloud-deploy-04 | MEDIUM | P1 | Project update can resurrect stopping/deleting deployment | `admin/deployments.go:97-119,139-175`; stale `DesiredStatus` is written back unconditionally | Retrigger without rewriting desired status, or use expected-version/CAS updates. |
| deep-cloud-deploy-05 | MEDIUM | P1 | Project deployment mutations have no serialization | `admin/projects.go:336-406,423-449`; redeploy/hibernate remain independent multi-step writes | Serialize per project with advisory lock/job and add optimistic project updates. |
| deep-cloud-deploy-06 | MEDIUM | P1 | Fatal controller exit permanently downs instance | `runtime/registry.go:296-303,500-513`; controller error is stored with no restart loop; health check only reports it | Restart with bounded exponential backoff and surface/remediate persistent failure. |
| deep-cloud-deploy-07 | MEDIUM | P1 | Transient connector ping can strand dependencies | `runtime/reconcilers/connector.go:83-103`; test failure returns no `Retrigger` after persisting spec hash | Retrigger transient failures with backoff or provide controller-level retry policy. |
| deep-cloud-deploy-08 | MEDIUM | P2 | Restart revalidates all resources with unbounded cold starts | `runtime/controller.go:174-182`; `runtime/reconcilers/metrics_view.go:135-182`; `runtime/registry.go:200-217` | Skip unchanged validation, rate-limit controller starts, and stagger sweeps. |
| deep-cloud-deploy-09 | LOW | P1 | Instance writers can clobber one another | `runtime/config_reloader.go:73-100`; `runtime/runtime.go:140-255`; full-row registry writes lack version CAS | Funnel edits through a per-instance owner or implement versioned field-level updates and retry. |
| deep-cloud-deploy-10 | LOW | P2 | Migration lookup can panic for non-model resource | `runtime/reconcilers/migration.go:139-152`; single-value `*Resource_Model` assertions remain | Use checked type switches and return an actionable unsupported-kind error. |
| deep-cloud-deploy-11 | LOW | P1 | Quota and branch uniqueness are check-then-act | `admin/server/projects.go:501-520`; `admin/server/deployment.go:309-338`; no enforcing DB constraint | Add partial unique branch index and transactional quota locking/enforcement. |
| deep-cloud-deploy-12 | LOW | P2 | Push during Pending deployment is dropped | `admin/github.go:396-401`; only Running/Updating deployments receive trigger, with hourly catch-up | Trigger Pending deployments idempotently or queue a delayed retry. |
| deep-cloud-deploy-13 | MEDIUM | P0 | Editable redeploy deletes successor’s live branch | `admin/projects.go:336-406`; `admin/deployments.go:461-473`; latest PR #9753 only cleans old stopped deployments and does not preserve the shared editable branch | Preserve a branch used by the successor or checkpoint/delete/create in a lossless sequence; exclude edit deployments from bulk reset. |
| deep-cloud-deploy-14 | LOW | P1 | Failed reconcile job strands transitional status | `admin/jobs/river/river.go:874-889`; final error/panic only logs/cancels; `admin/deployments.go:817` sets Errored only when enqueue fails | Mark deployment Errored on final discard/panic or run a janitor for orphaned transitions. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| — | None | PR #9753 (`22c323fce`) changes cleanup of old stopped deployments but does not close any of these fourteen findings. | — |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| — | — | — | None | The concurrency findings need deterministic tests, but the unsafe state transitions/locking contracts are still directly present. |

Counts: **14 pending, 0 fixed/remove, 0 partial/needs-test, 0 invalid = 14 total**.

## Machine-readable ledger

```text
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-01|HIGH|P1|-|Repeated first-deploy pulls parses and archive download|runtime/registry.go:339-478; runtime/config_reloader.go:132-163
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-02|HIGH|P1|-|Global config-reloader serialization|runtime/config_reloader.go:27-68,188-217
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-03|MEDIUM|P1|-|Deployment reconcile dedup drops state change|admin/jobs/river/river.go:342-350; admin/jobs/river/reconcile_deployment.go:50-51,86-89,130-146
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-04|MEDIUM|P1|-|Stale desired status resurrects deployment|admin/deployments.go:97-119,139-175
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-05|MEDIUM|P1|-|Project deployment mutation race|admin/projects.go:336-406,423-449
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-06|MEDIUM|P1|-|Fatal controller is never restarted|runtime/registry.go:296-303,500-513
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-07|MEDIUM|P1|-|Connector transient failure never retries|runtime/reconcilers/connector.go:83-103
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-08|MEDIUM|P2|-|Unbounded controller cold-start revalidation|runtime/controller.go:174-182; runtime/registry.go:200-217
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-09|LOW|P1|-|Instance record lost updates|runtime/config_reloader.go:73-100; runtime/runtime.go:140-255
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-10|LOW|P2|-|Migration lookup non-model panic|runtime/reconcilers/migration.go:139-152
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-11|LOW|P1|-|Quota and branch check-then-act|admin/server/projects.go:501-520; admin/server/deployment.go:309-338
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-12|LOW|P2|-|Pending deployment push dropped|admin/github.go:396-401
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-13|MEDIUM|P0|-|Editable redeploy deletes active branch|admin/projects.go:336-406; admin/deployments.go:461-473
PENDING|deep-cloud-deploy.md|deep-cloud-deploy-14|LOW|P1|-|Reconcile terminal failure strands deployment|admin/jobs/river/river.go:874-889
```
