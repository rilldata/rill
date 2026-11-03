# P1 — High priority next

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

Major correctness and reliability defects that should follow the P0 queue.

Count: **142**.

| Finding | Source | Original severity | Status | Current evidence |
|---|---|---:|---|---|
| [admin-auth-rbac-10](./parts/admin-auth-rbac.verified.md) — Group membership indirectly grants project admin | `admin-auth-rbac.md` | LOW | PENDING | admin/server/usergroups.go:583-613 |
| [abp-02](./parts/admin-billing-provision.verified.md) — Static runtime slot allocation TOCTOU | `admin-billing-provision.md` | HIGH | PENDING | admin/provisioner/static/static.go:94-119 |
| [abp-03](./parts/admin-billing-provision.verified.md) — Credit-depletion retry skips hibernation | `admin-billing-provision.md` | MEDIUM | PENDING | admin/jobs/river/credit_balance_event_handlers.go:110-198 |
| [abp-04](./parts/admin-billing-provision.verified.md) — Org jobs stop after first 100 projects | `admin-billing-provision.md` | MEDIUM | PENDING | admin/jobs/river/org_jobs.go:210,251 |
| [admin-database-06](./parts/admin-database.verified.md) — Persistence layer panics on database state | `admin-database.md` | LOW | PENDING | admin/database/postgres/postgres.go:1050,2133,2167,3971,3989 |
| [admin-projects-deploy-04](./parts/admin-projects-deploy.verified.md) — Nil primary deployment panics handlers | `admin-projects-deploy.md` | MEDIUM | PENDING | admin/server/reports.go:320; admin/server/virtual_file.go:77,177,236 |
| [admin-projects-deploy-05](./parts/admin-projects-deploy.verified.md) — GetRepoMeta nil deployment panic | `admin-projects-deploy.md` | LOW | PENDING | admin/server/repos.go:58-65,92-93 |
| [admin-projects-deploy-06](./parts/admin-projects-deploy.verified.md) — LookupCanvas inverted state check | `admin-projects-deploy.md` | MEDIUM | PENDING | admin/canvases.go:14-36 |
| [admin-projects-deploy-08](./parts/admin-projects-deploy.verified.md) — CreateProject partial-failure orphan | `admin-projects-deploy.md` | MEDIUM | PENDING | admin/projects.go:24-124 |
| [admin-projects-deploy-09](./parts/admin-projects-deploy.verified.md) — Private bookmark ownership bypass | `admin-projects-deploy.md` | LOW | PENDING | admin/server/bookmarks.go:39-69 |
| [ai-02](./parts/ai.verified.md) — Subscriber send can hang under session lock | `ai.md` | HIGH | PENDING | runtime/ai/ai.go:653,887-896; runtime/server/chat.go:387 |
| [ai-03](./parts/ai.verified.md) — Read-only SQL tool can execute mutations | `ai.md` | MEDIUM | PENDING | runtime/ai/query_sql.go:38-41; runtime/resolvers/sql.go:163-165 |
| [cli-02](./parts/cli.verified.md) — PKCE authenticator map race | `cli.md` | MEDIUM | PENDING | cli/pkg/local/app.go:59; cli/pkg/local/server.go:825,845,852 |
| [cli-03](./parts/cli.verified.md) — Helper client cache races | `cli.md` | MEDIUM | PENDING | cli/pkg/cmdutil/helper.go:170-277 |
| [cli-05](./parts/cli.verified.md) — World-readable CLI credentials | `cli.md` | MEDIUM | PENDING | cli/pkg/dotrill/dotrill.go:113,362 |
| [deep-ai-performance-02](./parts/deep-ai-performance.verified.md) — No prompt caching controls | `deep-ai-performance.md` | HIGH | PENDING | runtime/drivers/claude/claude.go:256-342; runtime/ai/ai.go:1608-1664 |
| [deep-ai-performance-06](./parts/deep-ai-performance.verified.md) — File write waits for global reconcile | `deep-ai-performance.md` | MEDIUM | PENDING | runtime/ai/file_write.go:118-156 |
| [deep-ai-performance-08](./parts/deep-ai-performance.verified.md) — Agent progress couples to client drain | `deep-ai-performance.md` | LOW | PENDING | runtime/ai/ai.go:653,887-896; runtime/server/chat.go:387 |
| [deep-ai-performance-10](./parts/deep-ai-performance.verified.md) — Develop-file seed filter cannot match | `deep-ai-performance.md` | MEDIUM | PENDING | runtime/ai/develop_file.go:96-119 |
| [CANV-1](./parts/deep-audit-canvas.verified.md) — timeAndFilterStore subscription leak | `deep-audit-canvas.md` | HIGH | PENDING | BaseCanvasComponent.ts:285-364 |
| [CANV-2](./parts/deep-audit-canvas.verified.md) — Insert-row corrupts touched widths | `deep-audit-canvas.md` | HIGH | PENDING | layout-util.ts:239 |
| [CANV-3](./parts/deep-audit-canvas.verified.md) — Stale metrics-view binding | `deep-audit-canvas.md` | MEDIUM | PENDING | BaseCanvasComponent.ts:97,172-181 |
| [CANV-5](./parts/deep-audit-canvas.verified.md) — Shared pivot lastKey | `deep-audit-canvas.md` | MEDIUM | PENDING | canvas/components/pivot/util.ts:88 |
| [CANV-6](./parts/deep-audit-canvas.verified.md) — Stale interval race | `deep-audit-canvas.md` | MEDIUM | PENDING | time-state.ts:162,199-206 |
| [CANV-8](./parts/deep-audit-canvas.verified.md) — FilterManager subscription leak | `deep-audit-canvas.md` | MEDIUM | PENDING | filter-manager.ts:145-180 |
| [CANV-10](./parts/deep-audit-canvas.verified.md) — Missing defaults throws | `deep-audit-canvas.md` | MEDIUM | PENDING | canvas-entity.ts:498 |
| [CANV-13](./parts/deep-audit-canvas.verified.md) — Markdown store leak | `deep-audit-canvas.md` | MEDIUM | PENDING | markdown/util.ts:135-196 |
| [LEAD-1](./parts/deep-audit-leaderboards.verified.md) — Unsafe URI schemes | `deep-audit-leaderboards.md` | MEDIUM | PENDING | dashboard-utils.ts:155-181 |
| [LEAD-2](./parts/deep-audit-leaderboards.verified.md) — Zero shown as no-data | `deep-audit-leaderboards.md` | MEDIUM | PENDING | LeaderboardRow.svelte:261,305,335 |
| [LEAD-3](./parts/deep-audit-leaderboards.verified.md) — NaN/Infinity percent total | `deep-audit-leaderboards.md` | MEDIUM | PENDING | leaderboard-utils.ts:105-106,215-216 |
| [LEAD-4](./parts/deep-audit-leaderboards.verified.md) — Selected no-data values vanish | `deep-audit-leaderboards.md` | MEDIUM | PENDING | Leaderboard.svelte:272-279 |
| [LEAD-5](./parts/deep-audit-leaderboards.verified.md) — Selected row duplicates by type | `deep-audit-leaderboards.md` | MEDIUM | PENDING | leaderboard-utils.ts:189-193 |
| [LEAD-6](./parts/deep-audit-leaderboards.verified.md) — Selection compatibility mismatch | `deep-audit-leaderboards.md` | MEDIUM | PENDING | DimensionTable.svelte:66 |
| [PIVO-1](./parts/deep-audit-pivot.verified.md) — Rows receive wrong cells after axis drop | `deep-audit-pivot.md` | HIGH | PENDING | pivot-table-transformations.ts:249 |
| [PIVO-2](./parts/deep-audit-pivot.verified.md) — Expanded cache discards fresh data | `deep-audit-pivot.md` | HIGH | PENDING | pivot-data-assembly.ts:89-95 |
| [PIVO-3](./parts/deep-audit-pivot.verified.md) — Shared lastKey breaks pagination | `deep-audit-pivot.md` | HIGH | PENDING | pivot-data-config.ts:25 |
| [PIVO-6](./parts/deep-audit-pivot.verified.md) — Config key omits search | `deep-audit-pivot.md` | MEDIUM | PENDING | pivot-utils.ts:100 |
| [PIVO-9](./parts/deep-audit-pivot.verified.md) — Pivot caches never evict | `deep-audit-pivot.md` | MEDIUM | PENDING | pivot-data-store.ts:699 |
| [PIVO-11](./parts/deep-audit-pivot.verified.md) — Missing sort dimension throws | `deep-audit-pivot.md` | MEDIUM | PENDING | pivot-utils.ts:361 |
| [STAT-1](./parts/deep-audit-state-management.verified.md) — Malformed compressed state crash | `deep-audit-state-management.md` | HIGH | PENDING | convertURLToExplorePreset.ts:85 |
| [STAT-4](./parts/deep-audit-state-management.verified.md) — LIKE remains after IN conversion | `deep-audit-state-management.md` | MEDIUM | PENDING | Filters.ts:312 |
| [STAT-5](./parts/deep-audit-state-management.verified.md) — Filter recomputation leak | `deep-audit-state-management.md` | MEDIUM | PENDING | filter-manager.ts:114-143 |
| [STAT-8](./parts/deep-audit-state-management.verified.md) — Stale interval race | `deep-audit-state-management.md` | MEDIUM | PENDING | time-state.ts:162,199-206 |
| [STAT-9](./parts/deep-audit-state-management.verified.md) — Unbounded stale cache | `deep-audit-state-management.md` | MEDIUM | PENDING | memoize-metrics-store.ts:10-15 |
| [STAT-10](./parts/deep-audit-state-management.verified.md) — Cross-project canvas state | `deep-audit-state-management.md` | MEDIUM | PENDING | canvas-entity.ts:54,599,660 |
| [TIME-2](./parts/deep-audit-time-dimension-detail.verified.md) — Dangerous URI schemes | `deep-audit-time-dimension-detail.md` | HIGH | PENDING | dashboard-utils.ts:155-181 |
| [TIME-3](./parts/deep-audit-time-dimension-detail.verified.md) — Global TDD scroll lock | `deep-audit-time-dimension-detail.md` | HIGH | PENDING | TDDTable.svelte:400-408 |
| [TIME-4](./parts/deep-audit-time-dimension-detail.verified.md) — Stale TDD runtime cache | `deep-audit-time-dimension-detail.md` | MEDIUM | PENDING | time-dimension-data-store.ts:463 |
| [TIME-5](./parts/deep-audit-time-dimension-detail.verified.md) — Missing pinned value freezes table | `deep-audit-time-dimension-detail.md` | MEDIUM | PENDING | time-dimension-data-store.ts:120 |
| [TIME-8](./parts/deep-audit-time-dimension-detail.verified.md) — Query error looks loading | `deep-audit-time-dimension-detail.md` | MEDIUM | PENDING | time-dimension-data-store.ts:380 |
| [TIME-9](./parts/deep-audit-time-dimension-detail.verified.md) — Zero comparison blanks absolute change | `deep-audit-time-dimension-detail.md` | MEDIUM | PENDING | time-dimension-data-store.ts:297,314 |
| [TIME-10](./parts/deep-audit-time-dimension-detail.verified.md) — Null total rendered zero | `deep-audit-time-dimension-detail.md` | MEDIUM | PENDING | TDDTable.svelte:212 |
| [deep-cloud-deploy-01](./parts/deep-cloud-deploy.verified.md) — Repeated first-deploy pulls parses and archive download | `deep-cloud-deploy.md` | HIGH | PENDING | runtime/registry.go:339-478; runtime/config_reloader.go:132-163 |
| [deep-cloud-deploy-02](./parts/deep-cloud-deploy.verified.md) — Global config-reloader serialization | `deep-cloud-deploy.md` | HIGH | PENDING | runtime/config_reloader.go:27-68,188-217 |
| [deep-cloud-deploy-03](./parts/deep-cloud-deploy.verified.md) — Deployment reconcile dedup drops state change | `deep-cloud-deploy.md` | MEDIUM | PENDING | admin/jobs/river/river.go:342-350; admin/jobs/river/reconcile_deployment.go:50-51,86-89,130-146 |
| [deep-cloud-deploy-04](./parts/deep-cloud-deploy.verified.md) — Stale desired status resurrects deployment | `deep-cloud-deploy.md` | MEDIUM | PENDING | admin/deployments.go:97-119,139-175 |
| [deep-cloud-deploy-05](./parts/deep-cloud-deploy.verified.md) — Project deployment mutation race | `deep-cloud-deploy.md` | MEDIUM | PENDING | admin/projects.go:336-406,423-449 |
| [deep-cloud-deploy-06](./parts/deep-cloud-deploy.verified.md) — Fatal controller is never restarted | `deep-cloud-deploy.md` | MEDIUM | PENDING | runtime/registry.go:296-303,500-513 |
| [deep-cloud-deploy-07](./parts/deep-cloud-deploy.verified.md) — Connector transient failure never retries | `deep-cloud-deploy.md` | MEDIUM | PENDING | runtime/reconcilers/connector.go:83-103 |
| [deep-cloud-deploy-09](./parts/deep-cloud-deploy.verified.md) — Instance record lost updates | `deep-cloud-deploy.md` | LOW | PENDING | runtime/config_reloader.go:73-100; runtime/runtime.go:140-255 |
| [deep-cloud-deploy-11](./parts/deep-cloud-deploy.verified.md) — Quota and branch check-then-act | `deep-cloud-deploy.md` | LOW | PENDING | admin/server/projects.go:501-520; admin/server/deployment.go:309-338 |
| [deep-cloud-deploy-14](./parts/deep-cloud-deploy.verified.md) — Reconcile terminal failure strands deployment | `deep-cloud-deploy.md` | LOW | PENDING | admin/jobs/river/river.go:874-889 |
| [deep-connectors-01](./parts/deep-connectors.verified.md) — rduckdb remote-delete panic and wrong metadata key | `deep-connectors.md` | HIGH | PENDING | runtime/pkg/rduckdb/remote.go:276,285-298 |
| [deep-connectors-05](./parts/deep-connectors.verified.md) — rduckdb write-handle metric callback leak | `deep-connectors.md` | HIGH | PENDING | runtime/pkg/rduckdb/db.go:956; runtime/pkg/rduckdb/generic.go:189 |
| [deep-connectors-07](./parts/deep-connectors.verified.md) — Postgres OLAP single-connection default | `deep-connectors.md` | MEDIUM | PENDING | runtime/drivers/postgres/postgres.go:211-216,425-431 |
| [deep-connectors-08](./parts/deep-connectors.verified.md) — MySQL and Snowflake pools unbounded | `deep-connectors.md` | MEDIUM | PENDING | runtime/drivers/mysql/mysql.go:384-390; runtime/drivers/snowflake/snowflake.go:381-386 |
| [deep-connectors-10](./parts/deep-connectors.verified.md) — Druid arrayLines truncation undetected | `deep-connectors.md` | MEDIUM | PENDING | runtime/drivers/druid/druidsqldriver/druid_api_sql_driver.go:319-335,483 |
| [deep-connectors-11](./parts/deep-connectors.verified.md) — BigQuery partial write returns success | `deep-connectors.md` | MEDIUM | PENDING | runtime/drivers/bigquery/warehouse.go:308-310,354-360 |
| [deep-connectors-12](./parts/deep-connectors.verified.md) — rduckdb generic connection leak | `deep-connectors.md` | MEDIUM | PENDING | runtime/pkg/rduckdb/generic.go:538-556 |
| [deep-connectors-13](./parts/deep-connectors.verified.md) — rduckdb dirty protocol serves stale data | `deep-connectors.md` | MEDIUM | PENDING | runtime/pkg/rduckdb/db.go:470-477,562-569,715-724,741-751 |
| [deep-connectors-15](./parts/deep-connectors.verified.md) — ClickHouse file ingest double buffers | `deep-connectors.md` | MEDIUM | PENDING | runtime/drivers/clickhouse/model_executor_filestore_self.go:151-162 |
| [deep-connectors-20](./parts/deep-connectors.verified.md) — Remote discovery returns handled not-found | `deep-connectors.md` | MEDIUM | PENDING | runtime/pkg/rduckdb/remote.go:165-172 |
| [deep-data-security-02](./parts/deep-data-security.verified.md) — Raw-SQL row-filter attribute injection | `deep-data-security.md` | MEDIUM | PENDING | runtime/security.go:759-767; runtime/metricsview/ast.go:366-395 |
| [deep-metrics-sql-06](./parts/deep-metrics-sql.verified.md) — Pipes concat compiles as logical OR | `deep-metrics-sql.md` | MEDIUM | PENDING | runtime/metricsview/metricssql/parser.go:45-53,487-511 |
| [deep-metrics-sql-09](./parts/deep-metrics-sql.verified.md) — Canvas ref discovery swallows parse errors | `deep-metrics-sql.md` | MEDIUM | PENDING | runtime/canvases.go:138-144 |
| [deep-sql-aggregation-01](./parts/deep-sql-aggregation.verified.md) — Two-phase percent-total comparison emits invalid SQL | `deep-sql-aggregation.md` | HIGH | PENDING | runtime/metricsview/executor/executor_rewrite_two_phase_comparisons.go:28-45,100-107 |
| [deep-sql-aggregation-02](./parts/deep-sql-aggregation.verified.md) — Comparison top-N runs before HAVING | `deep-sql-aggregation.md` | HIGH | PENDING | runtime/metricsview/executor/executor_rewrite_two_phase_comparisons.go:47-66; executor_rewrite_approx_comparisons.go:108-170 |
| [deep-sql-aggregation-03](./parts/deep-sql-aggregation.verified.md) — Half-open ranges bypass time caps | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/metricsview/executor/executor_enforce_query_limits.go:41-58 |
| [deep-sql-aggregation-04](./parts/deep-sql-aggregation.verified.md) — Half-open ranges can select incomplete rollup | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/metricsview/executor/executor_rewrite_rollup.go:132-133,220-230,503-537 |
| [deep-sql-aggregation-05](./parts/deep-sql-aggregation.verified.md) — Missing comparison range panics error path | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/metricsview/ast.go:603-604 |
| [deep-sql-aggregation-06](./parts/deep-sql-aggregation.verified.md) — Spine without time range panics | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/metricsview/executor/executor_rewrite_time.go:45-54 |
| [deep-sql-aggregation-07](./parts/deep-sql-aggregation.verified.md) — DST comparisons mis-join buckets | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/metricsview/ast.go:1525-1576 |
| [deep-sql-aggregation-08](./parts/deep-sql-aggregation.verified.md) — Druid search uses logical dimension names | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/metricsview/executor/executor.go:830-851 |
| [deep-sql-aggregation-09](./parts/deep-sql-aggregation.verified.md) — ClickHouse wrap corrupts CTE metadata | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/metricsview/executor/executor_wrap_clickhouse_computed_time_dim.go:75-87; runtime/metricsview/ast.go:709-728 |
| [deep-sql-aggregation-10](./parts/deep-sql-aggregation.verified.md) — Pivot staging is unbounded | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/metricsview/executor/executor_pivot.go:98-103,143-165 |
| [deep-sql-aggregation-16](./parts/deep-sql-aggregation.verified.md) — Pre-bound timestamps omit Now | `deep-sql-aggregation.md` | MEDIUM | PENDING | runtime/resolvers/metricsview_time_range.go:191-253; runtime/queries/metricsview_aggregation.go:231-287 |
| [drivers-connectors-ai-03](./parts/drivers-connectors-ai.verified.md) — Archive admin repo nil dereference | `drivers-connectors-ai.md` | MEDIUM | PENDING | runtime/drivers/admin/repo.go:257-260,298-301,328-331,373-376,412-416 |
| [drivers-connectors-ai-04](./parts/drivers-connectors-ai.verified.md) — Connector update retains stale handle | `drivers-connectors-ai.md` | MEDIUM | PENDING | runtime/runtime.go:211-253 |
| [drivers-object-store-01](./parts/drivers-object-store.verified.md) — Redshift UNLOAD quoted SQL breakage | `drivers-object-store.md` | HIGH | PENDING | runtime/drivers/redshift/warehouse.go:120 |
| [drivers-olap-01](./parts/drivers-olap.verified.md) — Pinot All nil-table panic | `drivers-olap.md` | HIGH | PENDING | runtime/drivers/pinot/information_schema.go:266-288 |
| [drivers-olap-02](./parts/drivers-olap.verified.md) — DuckDB connection metric race | `drivers-olap.md` | MEDIUM | PENDING | runtime/drivers/duckdb/duckdb.go:214,320,730,744 |
| [drivers-olap-03](./parts/drivers-olap.verified.md) — ClickHouse Open resource leaks | `drivers-olap.md` | MEDIUM | PENDING | runtime/drivers/clickhouse/clickhouse.go:308-368 |
| [drivers-olap-05](./parts/drivers-olap.verified.md) — ClickHouse context goroutine leak | `drivers-olap.md` | MEDIUM | PENDING | runtime/drivers/clickhouse/olap.go:513-540 |
| [drivers-olap-07](./parts/drivers-olap.verified.md) — StarRocks INT128 scan overflow | `drivers-olap.md` | MEDIUM | PENDING | runtime/drivers/starrocks/olap.go:204-205,322-329 |
| [drivers-olap-08](./parts/drivers-olap.verified.md) — Pinot cell conversion error and panic | `drivers-olap.md` | MEDIUM | PENDING | runtime/drivers/pinot/sqldriver/driver.go:229-235 |
| [frontend-canvas-03](./parts/frontend-canvas.verified.md) — Stale async interval overwrites latest | `frontend-canvas.md` | MEDIUM | PENDING | time-state.ts:127-209 |
| [frontend-canvas-04](./parts/frontend-canvas.verified.md) — Canvas state key leaks across instances | `frontend-canvas.md` | MEDIUM | PENDING | canvas-entity.ts:54,599,660 |
| [frontend-canvas-05](./parts/frontend-canvas.verified.md) — Nested derived subscriptions leak | `frontend-canvas.md` | MEDIUM | PENDING | BaseCanvasComponent.ts:285-364 |
| [frontend-dashboard-state-03](./parts/frontend-dashboard-state.verified.md) — Boolean sync lock drops changes | `frontend-dashboard-state.md` | MEDIUM | PENDING | DashboardStateSync.ts:208-299 |
| [frontend-dashboard-viz-01](./parts/frontend-dashboard-viz.verified.md) — Big-number zero comparison hidden | `frontend-dashboard-viz.md` | MEDIUM | PENDING | MeasureBigNumber.svelte:150,254 |
| [frontend-dashboard-viz-02](./parts/frontend-dashboard-viz.verified.md) — Dimension export ignores sort | `frontend-dashboard-viz.md` | MEDIUM | PENDING | dimension-table-export.ts:103-148 |
| [frontend-dashboard-viz-03](./parts/frontend-dashboard-viz.verified.md) — Pivot export drops measure sort | `frontend-dashboard-viz.md` | MEDIUM | PENDING | pivot-export.ts:155-181 |
| [frontend-dashboard-viz-06](./parts/frontend-dashboard-viz.verified.md) — Percent total divides by zero | `frontend-dashboard-viz.md` | LOW | PENDING | leaderboard-utils.ts:105-106,215-216 |
| [frontend-runtime-client-01](./parts/frontend-runtime-client.verified.md) — Failed chat send loses draft | `frontend-runtime-client.md` | HIGH | PENDING | conversation.ts:249,378,705 |
| [frontend-runtime-client-02](./parts/frontend-runtime-client.verified.md) — Tables query not invalidated | `frontend-runtime-client.md` | MEDIUM | PENDING | invalidation.ts:218 |
| [frontend-runtime-client-03](./parts/frontend-runtime-client.verified.md) — Queue abort listener survives removal | `frontend-runtime-client.md` | MEDIUM | PENDING | request-queue.ts:129,147,158 |
| [frontend-runtime-client-05](./parts/frontend-runtime-client.verified.md) — JWT wait ignores abort | `frontend-runtime-client.md` | MEDIUM | PENDING | runtime-client.ts:119 |
| [fs-shared-01](./parts/frontend-shared.verified.md) — Rename/delete swallow failures | `frontend-shared.md` | HIGH | PENDING | actions.ts:206-232,258-276 |
| [fs-shared-02](./parts/frontend-shared.verified.md) — String version comparison | `frontend-shared.md` | MEDIUM | PENDING | actions.ts:95-111 |
| [fs-shared-04](./parts/frontend-shared.verified.md) — Secrets persist after failed connector flow | `frontend-shared.md` | MEDIUM | PENDING | connector.ts:112-166 |
| [fs-shared-06](./parts/frontend-shared.verified.md) — Duplicate overwrites copy | `frontend-shared.md` | MEDIUM | PENDING | actions.ts:234-256 |
| [metricsview-02](./parts/metricsview.verified.md) — Malformed cast condition panics | `metricsview.md` | MEDIUM | PENDING | runtime/metricsview/astexpr.go:146-181 |
| [metricsview-04](./parts/metricsview.verified.md) — Invalid time grain can panic pre-validation | `metricsview.md` | MEDIUM | PENDING | runtime/metricsview/ast.go:194-214; runtime/metricsview/query.go:321-344 |
| [metricsview-05](./parts/metricsview.verified.md) — LIKE regex conversion does not escape metacharacters | `metricsview.md` | MEDIUM | PENDING | runtime/metricsview/astexpr.go:787-796 |
| [metricsview-06](./parts/metricsview.verified.md) — Percent of total divides by rounded raw literal | `metricsview.md` | MEDIUM | PENDING | runtime/metricsview/ast.go:558-569 |
| [metricsview-08](./parts/metricsview.verified.md) — Primary time dimension bypasses field access | `metricsview.md` | LOW | PENDING | runtime/metricsview/ast.go:647-669 |
| [metricsview-09](./parts/metricsview.verified.md) — NULL percent total hard-errors | `metricsview.md` | MEDIUM | PENDING | runtime/metricsview/executor/executor_rewrite_percent_of_totals.go:82-98 |
| [parser-03](./parts/parser.verified.md) — Multi-env reparse can erase change flag | `parser.md` | MEDIUM | PENDING | runtime/parser/parser.go:433-480,572-575 |
| [plg-first-run-01](./parts/plg-first-user-experience.verified.md) — Sample-data silent dead-end | `plg-first-user-experience.md` | HIGH | PENDING | generate-sample-data.ts:59-70 |
| [plg-first-run-02](./parts/plg-first-user-experience.verified.md) — AI generation infinite poll | `plg-first-user-experience.md` | MEDIUM | PENDING | actions.ts:162-179 |
| [plg-first-run-03](./parts/plg-first-user-experience.verified.md) — Deploy org route 404 | `plg-first-user-experience.md` | MEDIUM | PENDING | deploy/+page.ts:4-9 |
| [plg-first-run-08](./parts/plg-first-user-experience.verified.md) — String reconcile versions | `plg-first-user-experience.md` | LOW | PENDING | actions.ts:95-111 |
| [plg-first-run-13](./parts/plg-first-user-experience.verified.md) — Partial example unpack | `plg-first-user-experience.md` | LOW | PENDING | runtime/server/examples.go:97-110 |
| [queries-02](./parts/queries.verified.md) — Millisecond grain panics null fill | `queries.md` | MEDIUM | PENDING | runtime/queries/timeutil.go:12-35; runtime/queries/metricsview_timeseries.go:348-350,379-381 |
| [reconcilers-04](./parts/reconcilers.verified.md) — Report retry can redeliver successful notifiers | `reconcilers.md` | MEDIUM | PENDING | runtime/reconcilers/report.go:463-475,594-612 |
| [reconcilers-05](./parts/reconcilers.verified.md) — Pending partition pagination stops after first batch | `reconcilers.md` | MEDIUM | PENDING | runtime/reconcilers/model.go:1637-1665; runtime/drivers/sqlite/catalog.go:158-172 |
| [reconcilers-06](./parts/reconcilers.verified.md) — Migration re-executes current version | `reconcilers.md` | MEDIUM | PENDING | runtime/reconcilers/migration.go:76-97 |
| [resolvers-04](./parts/resolvers.verified.md) — Metrics SQL API ignores additional_where | `resolvers.md` | MEDIUM | PENDING | runtime/resolvers/builtin_metrics_sql.go:28-43,60-72 |
| [resolvers-05](./parts/resolvers.verified.md) — SQL cache key omits connector | `resolvers.md` | MEDIUM | PENDING | runtime/resolvers/sql.go:147-152 |
| [controller-late-completion](./REVIEW-FINDINGS-DELTA.md) — Late reconcile completion can block after controller shutdown | `REVIEW_FINDINGS_TODO.md` | HIGH | PENDING | runtime/controller.go:320-350 stops draining after 30s; invoke sends unconditionally at :1335 |
| [runtime-core-02](./parts/runtime-core.verified.md) — Failed instance delete waits on nil channel | `runtime-core.md` | HIGH | PENDING | runtime/registry.go:114-120,398-415 |
| [runtime-pkg-02](./parts/runtime-pkg.verified.md) — Slow log subscriber blocks under mutex | `runtime-pkg.md` | HIGH | PENDING | runtime/pkg/logbuffer/buffer.go:38-48 |
| [runtime-pkg-03](./parts/runtime-pkg.verified.md) — Telemetry HTTP can hang shutdown | `runtime-pkg.md` | MEDIUM | PENDING | runtime/pkg/activity/sink_intake.go:57-68,121-131 |
| [runtime-server-03](./parts/runtime-server.verified.md) — JWKS bootstrap leaks refreshers | `runtime-server.md` | MEDIUM | PENDING | runtime/server/auth/jwt.go:202-224 |
| [runtime-server-04](./parts/runtime-server.verified.md) — SSE producer sends ignore cancellation | `runtime-server.md` | MEDIUM | PENDING | runtime/server/sse.go:137,171,218,265; runtime/server/chat.go:525,545 |
| [runtime-server-05](./parts/runtime-server.verified.md) — SSE validates types after goroutine launch | `runtime-server.md` | MEDIUM | PENDING | runtime/server/sse.go:84,111-229 |
| [runtime-server-10](./parts/runtime-server.verified.md) — Export appends error after streaming | `runtime-server.md` | LOW | PENDING | runtime/server/downloads.go:328-356 |
| [runtime-server-11](./parts/runtime-server.verified.md) — Raw query limit rewrite is DuckDB-only | `runtime-server.md` | MEDIUM | PENDING | runtime/server/queries.go:39-47,110-160 |
| [web-admin-01](./parts/web-admin.verified.md) — Open redirect | `web-admin.md` | HIGH | PENDING | upgrade-callback/+page.svelte:34,97-100 |
| [web-admin-02](./parts/web-admin.verified.md) — View-As prior-user cache | `web-admin.md` | MEDIUM | PENDING | [project]/+layout.svelte:196-207 |
| [web-admin-03](./parts/web-admin.verified.md) — View-As admin permission fallback | `web-admin.md` | MEDIUM | PENDING | project-runtime.ts:38-39 |
| [web-admin-06](./parts/web-admin.verified.md) — Embed RPC origin not pinned | `web-admin.md` | LOW | PENDING | rpc.ts:64-104 |

