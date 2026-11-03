# Current must-fix and high-priority backlog

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

This is the clean, deduplicated successor to the historical 42-item list. It contains only current high-priority work in the active tables. **37 remain P0/P1; 3 are still pending but were reclassified to P2/P3; 2 were verified fixed/invalid and removed.**

Priority is based on current impact: P0 is an active security/data-loss/corruption/crash boundary; P1 is major correctness or reliability work.

## P0 (11)

| ID | Finding | Verified source | Current code evidence |
|---|---|---|---|
| MF-05 | `logbuffer.GetLogs(limit=0)` panics and can crash the runtime | [runtime-pkg.md](./parts/runtime-pkg.verified.md) | runtime/pkg/logbuffer/buffer.go:73-95 |
| MF-13 | Magic tokens replay the creator's admin flag and privileged attributes | [admin-auth-rbac.md](./parts/admin-auth-rbac.verified.md) | admin/server/magic_tokens.go:90-100; admin/server/runtime_jwt.go:113-120 |
| MF-14 | Editable deployment deletion continues after checkpoint push failure | [admin-projects-deploy.md](./parts/admin-projects-deploy.verified.md) | admin/deployments.go:413-476 |
| MF-15 | Object-store `path_prefixes` are bypassed by glob listing and downloads | [drivers-object-store.md](./parts/drivers-object-store.verified.md) | runtime/drivers/s3/object_store.go:70-101 |
| MF-16 | Object-store prefix checks are not boundary-aware | [drivers-object-store.md](./parts/drivers-object-store.verified.md) | runtime/drivers/object_store.go:177-180,253-287 |
| MF-20 | Failed delete reconciliation still hard-deletes the catalog resource | [REVIEW_FINDINGS_TODO.md delta](./REVIEW-FINDINGS-DELTA.md) | runtime/controller.go:1459-1470 logs the delete error and still calls catalog.delete |
| MF-27 | Embed singleton state can retain the previous token, host, instance, or external user | [REVIEW_FINDINGS_TODO.md delta](./REVIEW-FINDINGS-DELTA.md) | embed-store.ts:30-53 is process-global; embed/+layout.ts:7-10 initializes only when absent |
| MF-29 | rduckdb deletes the old remote version even when new metadata upload fails | [deep-connectors.md](./parts/deep-connectors.verified.md) | runtime/pkg/rduckdb/remote.go:244-255 |
| MF-30 | ClickHouse object-store credentials leak into SQL traces/logs | [deep-connectors.md](./parts/deep-connectors.verified.md) | runtime/drivers/clickhouse/model_executor_objectstore_self.go:126-163 |
| MF-31 | Salesforce batch-result fetch errors become empty files | [deep-connectors.md](./parts/deep-connectors.verified.md) | runtime/drivers/salesforce/bulk.go:201-231 |
| MF-39 | TDD row headers render data-controlled dimension values through `innerHTML` | [deep-audit-time-dimension-detail.md](./parts/deep-audit-time-dimension-detail.verified.md) | TDDTable.svelte:247-262 |

## P1 (26)

| ID | Finding | Verified source | Current code evidence |
|---|---|---|---|
| MF-01 | Unsafe web-admin redirects after billing/device flows | [web-admin.md](./parts/web-admin.verified.md) | upgrade-callback/+page.svelte:34,97-100 |
| MF-04 | Pinot schema enumeration can panic on a transient table-schema failure | [drivers-olap.md](./parts/drivers-olap.verified.md) | runtime/drivers/pinot/information_schema.go:266-288 |
| MF-06 | A slow log subscriber can deadlock instance logging | [runtime-pkg.md](./parts/runtime-pkg.verified.md) | runtime/pkg/logbuffer/buffer.go:38-48 |
| MF-07 | `DeleteInstance` can block forever after registry-store deletion fails | [runtime-core.md](./parts/runtime-core.verified.md) | runtime/registry.go:114-120,398-415 |
| MF-08 | AI session notification can strand the producer under the session lock | [ai.md](./parts/ai.verified.md) | runtime/ai/ai.go:653,887-896; runtime/server/chat.go:387 |
| MF-09 | Static runtime allocation can overcommit hosts | [admin-billing-provision.md](./parts/admin-billing-provision.verified.md) | admin/provisioner/static/static.go:94-119 |
| MF-10 | File rename/delete helpers swallow errors | [frontend-shared.md](./parts/frontend-shared.verified.md) | actions.ts:206-232,258-276 |
| MF-11 | Failed chat sends lose the draft and retain a false optimistic message | [frontend-runtime-client.md](./parts/frontend-runtime-client.verified.md) | conversation.ts:249,378,705 |
| MF-12 | Redshift `UNLOAD` breaks queries containing single quotes | [drivers-object-store.md](./parts/drivers-object-store.verified.md) | runtime/drivers/redshift/warehouse.go:120 |
| MF-17 | AI `query_sql` executes mutations despite advertising read-only behavior | [ai.md](./parts/ai.verified.md) | runtime/ai/query_sql.go:38-41; runtime/resolvers/sql.go:163-165 |
| MF-18 | Local credentials are written with world-readable permissions | [cli.md](./parts/cli.verified.md) | cli/pkg/dotrill/dotrill.go:113,362 |
| MF-21 | Late reconcile completions can block forever after controller shutdown | [REVIEW_FINDINGS_TODO.md delta](./REVIEW-FINDINGS-DELTA.md) | runtime/controller.go:320-350 stops draining after 30s; invoke sends unconditionally at :1335 |
| MF-22 | DuckDB connection-count metrics have a data race | [drivers-olap.md](./parts/drivers-olap.verified.md) | runtime/drivers/duckdb/duckdb.go:214,320,730,744 |
| MF-23 | ClickHouse partial-open failures leak initialized resources | [drivers-olap.md](./parts/drivers-olap.verified.md) | runtime/drivers/clickhouse/clickhouse.go:308-368 |
| MF-25 | Exhausted deployment reconcile jobs do not persist a terminal error state | [deep-cloud-deploy.md](./parts/deep-cloud-deploy.verified.md) | admin/jobs/river/river.go:874-889 |
| MF-26 | Migration reconciliation re-executes the current version | [reconcilers.md](./parts/reconcilers.verified.md) | runtime/reconcilers/migration.go:76-97 |
| MF-28 | rduckdb `deleteRemote` panics on listing errors and deletes the wrong `meta.json` | [deep-connectors.md](./parts/deep-connectors.verified.md) | runtime/pkg/rduckdb/remote.go:276,285-298 |
| MF-32 | rduckdb registers a non-removable DB-stats callback per write-handle open | [deep-connectors.md](./parts/deep-connectors.verified.md) | runtime/pkg/rduckdb/db.go:956; runtime/pkg/rduckdb/generic.go:189 |
| MF-33 | Canvas `timeAndFilterStore` leaks inner subscriptions | [deep-audit-canvas.md](./parts/deep-audit-canvas.verified.md) | BaseCanvasComponent.ts:285-364 |
| MF-34 | Canvas row insertion marks the wrong rows touched | [deep-audit-canvas.md](./parts/deep-audit-canvas.verified.md) | layout-util.ts:239 |
| MF-35 | Pivot row-total merging can assign cells to the wrong dimension rows | [deep-audit-pivot.md](./parts/deep-audit-pivot.verified.md) | pivot-table-transformations.ts:249 |
| MF-36 | Pivot expanded-table cache discards new pages and fresh data | [deep-audit-pivot.md](./parts/deep-audit-pivot.verified.md) | pivot-data-assembly.ts:89-95 |
| MF-37 | A module-level pivot `lastKey` breaks pagination across instances | [deep-audit-pivot.md](./parts/deep-audit-pivot.verified.md) | pivot-data-config.ts:25 |
| MF-38 | Malformed or oversized `gzipped_state` crashes dashboard load | [deep-audit-state-management.md](./parts/deep-audit-state-management.verified.md) | convertURLToExplorePreset.ts:85 |
| MF-40 | Data-controlled `javascript:`/`data:` URIs reach dashboard links | [deep-audit-time-dimension-detail.md](./parts/deep-audit-time-dimension-detail.verified.md) | dashboard-utils.ts:155-181 |
| MF-41 | TDD global scroll position is never reset or scoped | [deep-audit-time-dimension-detail.md](./parts/deep-audit-time-dimension-detail.verified.md) | TDDTable.svelte:400-408 |

## Deprioritized from the high-priority rollup (3)

These findings remain valid and appear in the planned/later source queues, but current-code review no longer supports treating them as must-fix blockers.

| ID | Current priority | Finding | Verified source | Current code evidence |
|---|---:|---|---|---|
| MF-19 | P2 | Local PKCE auth/logout handlers accept external redirects | [cli.md](./parts/cli.verified.md) | cli/pkg/local/server.go:815-820,879,897-901 |
| MF-24 | P2 | Runtime-client JWT can become `Bearer undefined` after waiting for refresh | [frontend-runtime-client.md](./parts/frontend-runtime-client.verified.md) | runtime-client.ts:63 |
| MF-42 | P2 | Metrics SQL accepts unsupported operators into invalid IR | [deep-metrics-sql.md](./parts/deep-metrics-sql.verified.md) | runtime/metricsview/metricssql/parser.go:487-511 |

## Removed since the prior 42-item snapshot (2)

| ID | Finding | Disposition | Evidence |
|---|---|---|---|
| MF-02 | AI file tools can escape the project root | FIXED | runtime/ai/util.go:66-80; runtime/drivers/repo.go:70-82; PR#9723 |
| MF-03 | Runtime repo file APIs can escape the project root | FIXED | runtime/drivers/repo.go:150-160; PR #9723 6793dbb0265ce4d126cc7206529dfef54660ea82 |

The original IDs are retained so links and remediation history remain stable. Full source-level inventories, including lower priorities and duplicates, are linked above and from [README.md](./README.md).
