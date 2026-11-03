# `REVIEW_FINDINGS_TODO.md` current-code delta

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

`REVIEW_FINDINGS_TODO.md` is a supplemental rollup outside `audit/`, not an independent area report. Its findings overlap heavily with the 37 source reports, so they are not counted again in the source-level totals. This file records the current status of its 17-item High Priority section and the latest merged-PR comparison.

## High Priority status

Of the original 17 High Priority findings, **12 remain unresolved and 5 are fixed**. Current-code reprioritization leaves 11 of those unresolved items at P0/P1 and moves the bounded JWT-header race to P2. The only new closure since the 2026-07-25 remediation snapshot is the repo-path confinement fix in PR #9723.

| Current priority | Canonical ID | Finding | Current evidence |
|---:|---|---|---|
| P0 | MF-14 | Editable deployment deletion continues after checkpoint push failure | `admin/deployments.go:413-421,462-473`; verified in `admin-projects-deploy.verified.md` |
| P1 | MF-09 | Static runtime allocation is not serialized | `admin/provisioner/static/static.go:94-119` |
| P1 | MF-25 | Exhausted deployment jobs leave transitional state | `admin/jobs/river/river.go:874-889` logs terminal failure but does not update the deployment |
| P0 | MF-20 | Failed delete reconciliation hard-deletes the catalog resource | `runtime/controller.go:1459-1470` logs `inv.result.Err`, then still calls `c.catalog.delete` |
| P1 | MF-21 | Late reconcile completion can block after shutdown | `runtime/controller.go:320-350` stops draining after 30s; `invoke` sends unconditionally at `:1335` |
| P1 | MF-26 | Migration reconciliation re-executes the current version | `runtime/reconcilers/migration.go:76-97` still loops from `from` through `to`, inclusive |
| P1 | MF-22 | DuckDB connection-count callback races | `runtime/drivers/duckdb/duckdb.go:214,320,730,744` |
| P1 | MF-23 | ClickHouse partial-open failures leak resources | `runtime/drivers/clickhouse/clickhouse.go:308-368` |
| P0 | MF-15 | Object-store glob/download entry points bypass `path_prefixes` | `runtime/drivers/s3/object_store.go:70-101` |
| P2 | MF-24 | JWT wait can resume with `Bearer undefined` | `runtime-client.ts:63-67` checks the token only before `await waitForFreshJwt()` |
| P1 | MF-01 | Billing/device redirects accept unsafe destinations | `upgrade-callback/+page.svelte:34,97-100` |
| P0 | MF-27 | Embed singleton retains a previous token/host/instance | `embed-store.ts:30-53` is global; `embed/+layout.ts:7-10` initializes only when absent |

## Fixed / removed

| Finding | Evidence |
|---|---|
| Static ClickHouse database-name collision | PR #9683 |
| Partition rows leaked on model deletion | PR #9711 |
| Dashboard state-sync lock leak | PR #9701 |
| Canvas PDF omitted tab groups | PR #9722 |
| Runtime repo APIs escaped the project root | PR #9723; shared `ResolveRepoPath` plus repo-driver regression tests |

PR #9723 also fixes the duplicate AI-tool witness (MF-02), so the clean must-fix rollup removes two IDs for one root-cause PR.

## Merged-PR review after 2026-07-25

| PR | Audit disposition |
|---|---|
| [#9723](https://github.com/rilldata/rill/pull/9723) | Exact match: fixes MF-02 and MF-03. |
| [#9750](https://github.com/rilldata/rill/pull/9750) | Valuable adjacent fix for model-output existence checks; it does **not** fix the supplemental metrics-view reconciler lookup finding in `runtime/reconcilers/metrics_view.go:82-99`. |
| [#9763](https://github.com/rilldata/rill/pull/9763) | Fixes pivot filter scoping, but no existing audit finding describes that defect. |
| [#9753](https://github.com/rilldata/rill/pull/9753) | Deletes old stopped deployments on redeploy; it does not fix any deployment audit finding and does not address editable-branch loss. |

Focused verification passed:

- `go test ./runtime/drivers -run '^TestResolveRepoPath$' -count=1`
- `go test ./runtime/ai -run '^Test(Read|Write)File$' -count=1`
- `go test ./runtime/drivers/file -run '^TestListGlob$/traversal_globs_cannot_escape_the_root$' -count=1`
