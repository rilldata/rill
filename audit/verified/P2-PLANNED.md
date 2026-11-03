# P2 — Planned fixes

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

Confirmed lower-impact bugs, performance problems, and bounded robustness work.

Count: **160**.

| Finding | Source | Original severity | Status | Current evidence |
|---|---|---:|---|---|
| [admin-auth-rbac-02](./parts/admin-auth-rbac.verified.md) — Revoked token remains cached | `admin-auth-rbac.md` | LOW | PENDING | admin/auth_token.go:249-312 |
| [admin-auth-rbac-07](./parts/admin-auth-rbac.verified.md) — Transient token errors negatively cached | `admin-auth-rbac.md` | LOW | PENDING | admin/auth_token.go:298-311 |
| [admin-auth-rbac-08](./parts/admin-auth-rbac.verified.md) — Unvalidated device redirect | `admin-auth-rbac.md` | LOW | PENDING | admin/server/auth/device_code.go:80-90 |
| [admin-auth-rbac-09](./parts/admin-auth-rbac.verified.md) — DeleteUser email enumeration | `admin-auth-rbac.md` | LOW | PENDING | admin/server/users.go:525-537 |
| [admin-auth-rbac-12](./parts/admin-auth-rbac.verified.md) — Unbounded page size | `admin-auth-rbac.md` | LOW | PENDING | admin/server/page_token.go:70-75 |
| [admin-auth-rbac-13](./parts/admin-auth-rbac.verified.md) — Transient validation logs browser out | `admin-auth-rbac.md` | LOW | PENDING | admin/server/auth/middleware.go:98-112 |
| [abp-05](./parts/admin-billing-provision.verified.md) — Trial-end email inside DB transaction | `admin-billing-provision.md` | LOW | PENDING | admin/jobs/river/trial_checks.go:147-200 |
| [abp-06](./parts/admin-billing-provision.verified.md) — Used flusher shutdown does not join loop | `admin-billing-provision.md` | LOW | PENDING | admin/used_flusher.go:113-133 |
| [abp-10](./parts/admin-billing-provision.verified.md) — Invoice metadata update race | `admin-billing-provision.md` | LOW | PENDING | admin/jobs/river/biller_event_handlers.go:36-79,115-162 |
| [abp-11](./parts/admin-billing-provision.verified.md) — Billing/org jobs lack uniqueness | `admin-billing-provision.md` | LOW | PENDING | admin/jobs/river/river.go:631-657 |
| [admin-database-04](./parts/admin-database.verified.md) — Prod TTL can hibernate dev deployment | `admin-database.md` | LOW | PENDING | admin/database/postgres/postgres.go:647-660 |
| [admin-database-05](./parts/admin-database.verified.md) — Dynamic project-variable insert guards missing | `admin-database.md` | LOW | PENDING | admin/database/postgres/postgres.go:3214-3262 |
| [admin-database-09](./parts/admin-database.verified.md) — Fragile database error handling | `admin-database.md` | LOW | PENDING | admin/database/postgres/postgres.go:3264-3284; admin/database/postgres/migrate.go:24-32 |
| [admin-projects-deploy-03](./parts/admin-projects-deploy.verified.md) — Custom-domain propagation dead branch | `admin-projects-deploy.md` | MEDIUM | PENDING | admin/server/organizations.go:1021-1050 |
| [admin-projects-deploy-07](./parts/admin-projects-deploy.verified.md) — GitHub webhook processed synchronously | `admin-projects-deploy.md` | MEDIUM | PENDING | admin/github.go:344-411 |
| [admin-projects-deploy-11](./parts/admin-projects-deploy.verified.md) — Managed GitHub repository orphan | `admin-projects-deploy.md` | LOW | PENDING | admin/github.go:250-274 |
| [admin-projects-deploy-12](./parts/admin-projects-deploy.verified.md) — Asset handler commits success before read | `admin-projects-deploy.md` | LOW | PENDING | admin/server/assets.go:227-230 |
| [ai-04](./parts/ai.verified.md) — AI metrics limit parsing and lower bound are unsafe | `ai.md` | MEDIUM | PENDING | runtime/ai/metrics_view_query.go:241-246 |
| [ai-06](./parts/ai.verified.md) — Stack traces enter persisted LLM messages | `ai.md` | LOW | PENDING | runtime/ai/ai.go:991-995,1337 |
| [ai-07](./parts/ai.verified.md) — Final response retains only last text block | `ai.md` | LOW | PENDING | runtime/ai/ai.go:1405-1414 |
| [ai-11](./parts/ai.verified.md) — Nested agents lack aggregate budget | `ai.md` | MEDIUM | PENDING | runtime/ai/ai.go:1132-1151 |
| [ai-12](./parts/ai.verified.md) — Analyst context becomes stale after navigation | `ai.md` | MEDIUM | PENDING | runtime/ai/analyst_agent.go:106-161 |
| [cli-04](./parts/cli.verified.md) — Local PKCE/logout open redirect | `cli.md` | MEDIUM | PENDING | cli/pkg/local/server.go:815-820,879,897-901 |
| [cli-06](./parts/cli.verified.md) — ListMatchingProjects swallows errors | `cli.md` | MEDIUM | PENDING | cli/pkg/local/server.go:727-745 |
| [cli-12](./parts/cli.verified.md) — Invisible env upload | `cli.md` | LOW | PENDING | cli/cmd/project/deploy.go:309,437-451 |
| [cli-13](./parts/cli.verified.md) — PollServer health-check race | `cli.md` | LOW | PENDING | cli/pkg/local/app.go:480-529 |
| [cli-14](./parts/cli.verified.md) — App-start telemetry silently dropped | `cli.md` | LOW | PENDING | cli/pkg/local/app.go:536-556 |
| [deep-ai-performance-01](./parts/deep-ai-performance.verified.md) — No token-level LLM streaming | `deep-ai-performance.md` | HIGH | PENDING | runtime/drivers/ai.go:10-12; runtime/ai/ai.go:1282-1351 |
| [deep-ai-performance-03](./parts/deep-ai-performance.verified.md) — Tool calls execute serially | `deep-ai-performance.md` | MEDIUM | PENDING | runtime/ai/ai.go:1353-1397 |
| [deep-ai-performance-04](./parts/deep-ai-performance.verified.md) — Prompt mandates excessive queries and charts | `deep-ai-performance.md` | MEDIUM | PENDING | runtime/ai/instructions/data/analysis.md:25-40 |
| [deep-ai-performance-05](./parts/deep-ai-performance.verified.md) — Managed AI adds proxy hop and lacks tuning | `deep-ai-performance.md` | MEDIUM | PENDING | runtime/drivers/admin/ai.go:11-36 |
| [deep-ai-performance-07](./parts/deep-ai-performance.verified.md) — Develop-file subagent cold starts per file | `deep-ai-performance.md` | LOW | PENDING | runtime/ai/develop_file.go:67-138 |
| [CANV-4](./parts/deep-audit-canvas.verified.md) — Repeated time-range requests | `deep-audit-canvas.md` | MEDIUM | PENDING | time-state.ts:158 |
| [CANV-7](./parts/deep-audit-canvas.verified.md) — In-list marker retained | `deep-audit-canvas.md` | MEDIUM | PENDING | filter-state.ts:309 |
| [CANV-9](./parts/deep-audit-canvas.verified.md) — Leaked conversions multiply work | `deep-audit-canvas.md` | MEDIUM | PENDING | filter-manager.ts:114-143 |
| [CANV-11](./parts/deep-audit-canvas.verified.md) — Exponential combinations | `deep-audit-canvas.md` | MEDIUM | PENDING | filter-manager.ts:1063 |
| [CANV-12](./parts/deep-audit-canvas.verified.md) — No-op spec rebuild | `deep-audit-canvas.md` | MEDIUM | PENDING | BaseCanvasComponent.ts:180 |
| [LEAD-7](./parts/deep-audit-leaderboards.verified.md) — Global select-all hijack | `deep-audit-leaderboards.md` | MEDIUM | PENDING | DimensionDisplay.svelte:233 |
| [LEAD-8](./parts/deep-audit-leaderboards.verified.md) — Virtualizers recreated on scroll | `deep-audit-leaderboards.md` | MEDIUM | PENDING | DimensionTable.svelte:103 |
| [LEAD-9](./parts/deep-audit-leaderboards.verified.md) — Excessive overscan | `deep-audit-leaderboards.md` | MEDIUM | PENDING | DimensionTable.svelte:54,113 |
| [LEAD-10](./parts/deep-audit-leaderboards.verified.md) — Selection stores recreated | `deep-audit-leaderboards.md` | MEDIUM | PENDING | LeaderboardDisplay.svelte:104 |
| [LEAD-11](./parts/deep-audit-leaderboards.verified.md) — Widths estimated before data | `deep-audit-leaderboards.md` | MEDIUM | PARTIAL (#9789, 2026-08-30: delta/percent columns now sized reactively from row content; dimension and default measure columns still one-shot estimated at mount) | DimensionTable.svelte:80-83,118-125 |
| [LEAD-12](./parts/deep-audit-leaderboards.verified.md) — Singleton config leaks height | `deep-audit-leaderboards.md` | LOW | PENDING | DimensionTable.svelte:85 |
| [LEAD-13](./parts/deep-audit-leaderboards.verified.md) — Observer not disconnected | `deep-audit-leaderboards.md` | LOW | PENDING | Leaderboard.svelte:96-110 |
| [PIVO-4](./parts/deep-audit-pivot.verified.md) — Full pivot pipeline rebuild | `deep-audit-pivot.md` | MEDIUM | PENDING | pivot-data-store.ts:142 |
| [PIVO-5](./parts/deep-audit-pivot.verified.md) — All expansions re-merged | `deep-audit-pivot.md` | MEDIUM | PENDING | pivot-data-store.ts:488 |
| [PIVO-7](./parts/deep-audit-pivot.verified.md) — Virtualizer recreated on scroll | `deep-audit-pivot.md` | MEDIUM | PENDING | PivotTable.svelte:164 |
| [PIVO-8](./parts/deep-audit-pivot.verified.md) — Repeated totals scans | `deep-audit-pivot.md` | MEDIUM | PENDING | pivot-column-definition.ts:90 |
| [PIVO-10](./parts/deep-audit-pivot.verified.md) — Hover recomputes body cells | `deep-audit-pivot.md` | MEDIUM | PENDING | NestedTable.svelte:592 |
| [PIVO-13](./parts/deep-audit-pivot.verified.md) — Expansion traversal skips gaps | `deep-audit-pivot.md` | LOW | PENDING | pivot-expansion.ts:75 |
| [PIVO-14](./parts/deep-audit-pivot.verified.md) — Collapsed descendants keep querying | `deep-audit-pivot.md` | LOW | PENDING | pivot-expansion.ts:203 |
| [STAT-6](./parts/deep-audit-state-management.verified.md) — Exponential combinations | `deep-audit-state-management.md` | MEDIUM | PENDING | filter-manager.ts:1063 |
| [STAT-7](./parts/deep-audit-state-management.verified.md) — In-list marker retained | `deep-audit-state-management.md` | MEDIUM | PENDING | filter-state.ts:309 |
| [STAT-11](./parts/deep-audit-state-management.verified.md) — Nested store leak | `deep-audit-state-management.md` | LOW | PENDING | filter-manager.ts:145-198 |
| [STAT-12](./parts/deep-audit-state-management.verified.md) — Broad query observer rebuild | `deep-audit-state-management.md` | LOW | PENDING | state-managers.ts:119 |
| [STAT-13](./parts/deep-audit-state-management.verified.md) — Unsafe legacy state deserialize | `deep-audit-state-management.md` | LOW | PENDING | fromProto.ts:365 |
| [TIME-6](./parts/deep-audit-time-dimension-detail.verified.md) — Hover redraw storm | `deep-audit-time-dimension-detail.md` | MEDIUM | PENDING | TDDTable.svelte:378 |
| [TIME-7](./parts/deep-audit-time-dimension-detail.verified.md) — TDD query graph rebuild | `deep-audit-time-dimension-detail.md` | MEDIUM | PENDING | time-dimension-data-store.ts:343 |
| [TIME-11](./parts/deep-audit-time-dimension-detail.verified.md) — Unmemoized table preparation | `deep-audit-time-dimension-detail.md` | MEDIUM | PENDING | time-dimension-data-store.ts:160 |
| [TIME-12](./parts/deep-audit-time-dimension-detail.verified.md) — TDD chart pipeline rebuild | `deep-audit-time-dimension-detail.md` | LOW | PENDING | TDDChart.svelte:131 |
| [TIME-13](./parts/deep-audit-time-dimension-detail.verified.md) — Cache-key collision | `deep-audit-time-dimension-detail.md` | LOW | PENDING | memoize-metrics-store.ts:14 |
| [deep-cloud-deploy-08](./parts/deep-cloud-deploy.verified.md) — Unbounded controller cold-start revalidation | `deep-cloud-deploy.md` | MEDIUM | PENDING | runtime/controller.go:174-182; runtime/registry.go:200-217 |
| [deep-cloud-deploy-10](./parts/deep-cloud-deploy.verified.md) — Migration lookup non-model panic | `deep-cloud-deploy.md` | LOW | PENDING | runtime/reconcilers/migration.go:139-152 |
| [deep-cloud-deploy-12](./parts/deep-cloud-deploy.verified.md) — Pending deployment push dropped | `deep-cloud-deploy.md` | LOW | PENDING | admin/github.go:396-401 |
| [deep-connectors-06](./parts/deep-connectors.verified.md) — Object-store page rebuilds client and credentials | `deep-connectors.md` | MEDIUM | PENDING | runtime/drivers/s3/object_store.go:61-82 |
| [deep-connectors-09](./parts/deep-connectors.verified.md) — Druid cancel hook not stopped | `deep-connectors.md` | MEDIUM | PENDING | runtime/drivers/druid/druidsqldriver/druid_api_sql_driver.go:116-129 |
| [deep-connectors-14](./parts/deep-connectors.verified.md) — Incremental DuckDB full copy and upload | `deep-connectors.md` | MEDIUM | PENDING | runtime/pkg/rduckdb/db.go:512-518; runtime/pkg/rduckdb/remote.go:204-234 |
| [deep-connectors-16](./parts/deep-connectors.verified.md) — HTTPS temp file leak | `deep-connectors.md` | MEDIUM | PENDING | runtime/drivers/https/https.go:277 |
| [deep-connectors-17](./parts/deep-connectors.verified.md) — rduckdb RenameTable unescaped identifiers | `deep-connectors.md` | LOW | PENDING | runtime/pkg/rduckdb/generic.go:416 |
| [deep-connectors-18](./parts/deep-connectors.verified.md) — Transient retries too narrow | `deep-connectors.md` | LOW | PENDING | runtime/pkg/blob/download.go:323-340; runtime/pkg/rduckdb/remote.go:302-320 |
| [deep-connectors-19](./parts/deep-connectors.verified.md) — rduckdb read path serialization | `deep-connectors.md` | LOW | PENDING | runtime/pkg/rduckdb/db.go:329-349,1235-1274 |
| [deep-metrics-sql-07](./parts/deep-metrics-sql.verified.md) — Time expressions evaluate during parsing | `deep-metrics-sql.md` | MEDIUM | PENDING | runtime/metricsview/metricssql/time_range_parser.go:16-81 |
| [deep-metrics-sql-08](./parts/deep-metrics-sql.verified.md) — Unsupported operators fail late | `deep-metrics-sql.md` | MEDIUM | PENDING | runtime/metricsview/metricssql/parser.go:487-511 |
| [deep-metrics-sql-10](./parts/deep-metrics-sql.verified.md) — Filter and sort names lack parse validation | `deep-metrics-sql.md` | MEDIUM | PENDING | runtime/metricsview/metricssql/filter_parser.go:39-92; runtime/metricsview/metricssql/parser.go:357-367 |
| [deep-metrics-sql-11](./parts/deep-metrics-sql.verified.md) — LIMIT conversion can overflow | `deep-metrics-sql.md` | LOW | PENDING | runtime/metricsview/metricssql/filter_parser.go:425-448; runtime/metricsview/metricssql/parser.go:330-353 |
| [deep-metrics-sql-12](./parts/deep-metrics-sql.verified.md) — LIKE is always ILIKE and dates are ambiguous | `deep-metrics-sql.md` | LOW | PENDING | runtime/metricsview/metricssql/parser.go:503-504 |
| [deep-sql-aggregation-11](./parts/deep-sql-aggregation.verified.md) — Druid exactify omits HAVING measures | `deep-sql-aggregation.md` | LOW | PENDING | runtime/metricsview/executor/executor_rewrite_druid_exactify.go:33-62 |
| [deep-sql-aggregation-12](./parts/deep-sql-aggregation.verified.md) — Two-phase comparison double-applies OFFSET | `deep-sql-aggregation.md` | LOW | PENDING | runtime/metricsview/executor/executor_rewrite_two_phase_comparisons.go:47-66 |
| [deep-sql-aggregation-13](./parts/deep-sql-aggregation.verified.md) — Sub-day rollup ignores fractional timezone | `deep-sql-aggregation.md` | LOW | PENDING | runtime/metricsview/executor/executor_rewrite_rollup.go:312-356; runtime/metricsview/executor/rollup_grain.go:86-105 |
| [deep-sql-aggregation-14](./parts/deep-sql-aggregation.verified.md) — Multi-grain comparison ignores physical time expression | `deep-sql-aggregation.md` | LOW | PENDING | runtime/metricsview/ast.go:1550-1566 |
| [deep-sql-aggregation-15](./parts/deep-sql-aggregation.verified.md) — Array NIN excludes NULL arrays | `deep-sql-aggregation.md` | LOW | PENDING | runtime/metricsview/astexpr.go:662-702 |
| [drivers-connectors-ai-02](./parts/drivers-connectors-ai.verified.md) — Cancelled open leaks handle | `drivers-connectors-ai.md` | MEDIUM | PENDING | runtime/connection_cache.go:169-181 |
| [drivers-connectors-ai-05](./parts/drivers-connectors-ai.verified.md) — Slack fanout aborts on first failure | `drivers-connectors-ai.md` | MEDIUM | PENDING | runtime/drivers/slack/notifier.go:68-76,100-107,187-236 |
| [drivers-connectors-ai-10](./parts/drivers-connectors-ai.verified.md) — Admin repo close config race | `drivers-connectors-ai.md` | LOW | PENDING | runtime/drivers/admin/repo.go:697-706,1027-1033 |
| [drivers-connectors-ai-11](./parts/drivers-connectors-ai.verified.md) — Virtual repo path traversal | `drivers-connectors-ai.md` | LOW | PENDING | runtime/drivers/admin/repo_virtual.go:65-84 |
| [drivers-object-store-07](./parts/drivers-object-store.verified.md) — Redshift pagination loses context | `drivers-object-store.md` | MEDIUM | PENDING | runtime/drivers/redshift/olap.go:242 |
| [drivers-object-store-08](./parts/drivers-object-store.verified.md) — Iterator close skips S3 cleanup | `drivers-object-store.md` | MEDIUM | PENDING | runtime/drivers/redshift/warehouse.go:238-245; runtime/drivers/athena/warehouse.go:253-260 |
| [drivers-object-store-09](./parts/drivers-object-store.verified.md) — HTTPS connector config ignored | `drivers-object-store.md` | MEDIUM | PENDING | runtime/drivers/https/https.go:105-115,238-283 |
| [drivers-object-store-11](./parts/drivers-object-store.verified.md) — Blob listing fully materialized | `drivers-object-store.md` | MEDIUM | PENDING | runtime/pkg/blob/download.go:76-89 |
| [drivers-object-store-12](./parts/drivers-object-store.verified.md) — Salesforce empty batch panic | `drivers-object-store.md` | LOW | PENDING | runtime/drivers/salesforce/bulk.go:160-178 |
| [drivers-object-store-14](./parts/drivers-object-store.verified.md) — Azure getter mutates config | `drivers-object-store.md` | LOW | PENDING | runtime/drivers/azure/azure.go:108-113 |
| [drivers-olap-04](./parts/drivers-olap.verified.md) — DuckDB failed-open metric leak | `drivers-olap.md` | MEDIUM | PENDING | runtime/drivers/duckdb/duckdb.go:213-229 |
| [drivers-olap-09](./parts/drivers-olap.verified.md) — ClickHouse openHandle DB leaks | `drivers-olap.md` | LOW | PENDING | runtime/drivers/clickhouse/clickhouse.go:940-978 |
| [drivers-olap-10](./parts/drivers-olap.verified.md) — ClickHouse single-row context bypass | `drivers-olap.md` | LOW | PENDING | runtime/drivers/clickhouse/olap.go:504-531 |
| [drivers-olap-11](./parts/drivers-olap.verified.md) — Pinot request ignores context | `drivers-olap.md` | LOW | PENDING | runtime/drivers/pinot/sqldriver/driver.go:63-77 |
| [drivers-olap-12](./parts/drivers-olap.verified.md) — Unescaped connector config in SQL | `drivers-olap.md` | LOW | PENDING | runtime/drivers/starrocks/starrocks.go:362-380; runtime/drivers/duckdb/duckdb.go:513 |
| [drivers-olap-16](./parts/drivers-olap.verified.md) — Pinot silently drops tables | `drivers-olap.md` | LOW | PENDING | runtime/drivers/pinot/information_schema.go:273-281 |
| [drivers-olap-17](./parts/drivers-olap.verified.md) — ClickHouse metric callback leak | `drivers-olap.md` | LOW | PENDING | runtime/drivers/clickhouse/clickhouse.go:978 |
| [frontend-canvas-02](./parts/frontend-canvas.verified.md) — In-list marker survives Select conversion | `frontend-canvas.md` | MEDIUM | PENDING | filter-state.ts:308-326 |
| [frontend-canvas-06](./parts/frontend-canvas.verified.md) — Save-default subscription can leak | `frontend-canvas.md` | LOW | PENDING | canvas-entity.ts:476-484 |
| [frontend-canvas-07](./parts/frontend-canvas.verified.md) — Image object URLs not revoked | `frontend-canvas.md` | LOW | PENDING | Image.svelte:33-60 |
| [frontend-canvas-10](./parts/frontend-canvas.verified.md) — Raw image asset URL path | `frontend-canvas.md` | LOW | PENDING | Image.svelte:47 |
| [frontend-dashboard-state-04](./parts/frontend-dashboard-state.verified.md) — Double initialization race | `frontend-dashboard-state.md` | LOW | PENDING | DashboardStateSync.ts:129-201 |
| [frontend-dashboard-state-05](./parts/frontend-dashboard-state.verified.md) — Cache collision and retention | `frontend-dashboard-state.md` | LOW | PENDING | memoize-metrics-store.ts:10-15 |
| [frontend-dashboard-state-07](./parts/frontend-dashboard-state.verified.md) — Nullable TDD measure restore | `frontend-dashboard-state.md` | LOW | PENDING | explore-web-view-store.ts:142-152 |
| [frontend-dashboard-state-09](./parts/frontend-dashboard-state.verified.md) — TDD total semantic mismatch | `frontend-dashboard-state.md` | LOW | PENDING | time-dimension-data-store.ts:227-277 |
| [frontend-dashboard-state-10](./parts/frontend-dashboard-state.verified.md) — Comparison range unresolved | `frontend-dashboard-state.md` | LOW | PENDING | DashboardStateSync.ts:148-160 |
| [frontend-dashboard-state-11](./parts/frontend-dashboard-state.verified.md) — URL-state errors hidden | `frontend-dashboard-state.md` | LOW | PENDING | DashboardStateDataLoader.ts:329-335 |
| [frontend-dashboard-state-12](./parts/frontend-dashboard-state.verified.md) — Replace-state re-entry | `frontend-dashboard-state.md` | LOW | PENDING | DashboardStateSync.ts:278-299 |
| [frontend-dashboard-viz-04](./parts/frontend-dashboard-viz.verified.md) — Observer not disconnected | `frontend-dashboard-viz.md` | LOW | PENDING | Leaderboard.svelte:93-110 |
| [frontend-dashboard-viz-05](./parts/frontend-dashboard-viz.verified.md) — Max drops zero | `frontend-dashboard-viz.md` | LOW | PENDING | leaderboard-utils.ts:350-351 |
| [frontend-dashboard-viz-08](./parts/frontend-dashboard-viz.verified.md) — Shared empty-data arrays | `frontend-dashboard-viz.md` | LOW | PENDING | time-series/utils.ts:145-147 |
| [frontend-dashboard-viz-11](./parts/frontend-dashboard-viz.verified.md) — Duplicate filter merge | `frontend-dashboard-viz.md` | LOW | PENDING | pivot-merge-filters.ts:93-114 |
| [frontend-dashboard-viz-12](./parts/frontend-dashboard-viz.verified.md) — NaN search progress | `frontend-dashboard-viz.md` | LOW | PENDING | useDimensionSearchResults.ts:80-81 |
| [frontend-runtime-client-04](./parts/frontend-runtime-client.verified.md) — Priority keys mismatch RPC | `frontend-runtime-client.md` | LOW | PENDING | request-priorities.ts:9 |
| [frontend-runtime-client-06](./parts/frontend-runtime-client.verified.md) — SSE resume consumes retry | `frontend-runtime-client.md` | LOW | PENDING | sse-connection.ts:104,162,178 |
| [frontend-runtime-client-07](./parts/frontend-runtime-client.verified.md) — Cross-resource priority ignored | `frontend-runtime-client.md` | LOW | PENDING | request-queue.ts:31,176,210 |
| [frontend-runtime-client-08](./parts/frontend-runtime-client.verified.md) — Bearer undefined race | `frontend-runtime-client.md` | LOW | PENDING | runtime-client.ts:63 |
| [frontend-runtime-client-09](./parts/frontend-runtime-client.verified.md) — Date stripped in arrays | `frontend-runtime-client.md` | LOW | PENDING | strip-undefined.ts:15 |
| [frontend-runtime-client-10](./parts/frontend-runtime-client.verified.md) — Provider clobbers global client | `frontend-runtime-client.md` | LOW | PENDING | RuntimeProvider.svelte:25,37 |
| [fs-shared-05](./parts/frontend-shared.verified.md) — Raw upload filename URL | `frontend-shared.md` | LOW | PENDING | file-upload.ts:18-27 |
| [fs-shared-07](./parts/frontend-shared.verified.md) — YAML strings unescaped | `frontend-shared.md` | LOW | PENDING | code-utils.ts:306-311 |
| [fs-shared-08](./parts/frontend-shared.verified.md) — File dialog timing race | `frontend-shared.md` | LOW | PENDING | file-upload.ts:36-63 |
| [fs-shared-09](./parts/frontend-shared.verified.md) — Delete dialog closes early | `frontend-shared.md` | LOW | PENDING | ForceDeleteConfirmationDialog.svelte:36-40 |
| [fs-shared-10](./parts/frontend-shared.verified.md) — Degenerate names accepted | `frontend-shared.md` | LOW | PENDING | name-utils.ts:1 |
| [fs-shared-11](./parts/frontend-shared.verified.md) — Unescaped regex/replacement | `frontend-shared.md` | LOW | PENDING | code-utils.ts:342-397 |
| [metricsview-07](./parts/metricsview.verified.md) — Annotation validation does not cache OLAP handles | `metricsview.md` | LOW | PENDING | runtime/metricsview/executor/executor_validate.go:649-665 |
| [parser-04](./parts/parser.verified.md) — InitEmpty panics on empty display name | `parser.md` | LOW | PENDING | runtime/parser/init.go:18-21 |
| [parser-05](./parts/parser.verified.md) — Environment override bypasses known fields | `parser.md` | LOW | PENDING | runtime/parser/parse_node.go:346-356,370-375 |
| [parser-06](./parts/parser.verified.md) — Comparison dimension accepts measure | `parser.md` | LOW | PENDING | runtime/parser/parse_metrics_view.go:665-668 |
| [parser-10](./parts/parser.verified.md) — Invalid SQL position wraps to huge line | `parser.md` | LOW | PENDING | runtime/parser/parse_model.go:340-348,410-425 |
| [plg-first-run-04](./parts/plg-first-user-experience.verified.md) — Unpack failure hidden | `plg-first-user-experience.md` | MEDIUM | PENDING | ProjectCards.svelte:76-78 |
| [plg-first-run-05](./parts/plg-first-user-experience.verified.md) — Fixed five-second reload | `plg-first-user-experience.md` | MEDIUM | PENDING | ProjectCards.svelte:62-67 |
| [plg-first-run-06](./parts/plg-first-user-experience.verified.md) — Four resources per source | `plg-first-user-experience.md` | MEDIUM | PENDING | import.ts:60-65,90-114 |
| [plg-first-run-07](./parts/plg-first-user-experience.verified.md) — AppStart dropped | `plg-first-user-experience.md` | LOW | PENDING | cli/pkg/local/app.go:544-549 |
| [plg-first-run-11](./parts/plg-first-user-experience.verified.md) — No connector error state | `plg-first-user-experience.md` | LOW | PENDING | ConnectYourDataWidget.svelte:55-82 |
| [plg-first-run-12](./parts/plg-first-user-experience.verified.md) — Dimensionless dashboard unexplained | `plg-first-user-experience.md` | LOW | PENDING | generate_metrics_view.go:272-292 |
| [queries-04](./parts/queries.verified.md) — Profiling APIs inline SQL expressions | `queries.md` | MEDIUM | PENDING | runtime/queries/column_topk.go:70-78; runtime/queries/column_timeseries.go:638-642 |
| [queries-07](./parts/queries.verified.md) — Column timeseries leaks cursor on rows error | `queries.md` | LOW | PENDING | runtime/queries/column_timeseries.go:181-222 |
| [queries-09](./parts/queries.verified.md) — Constant histogram computes invalid bucket count | `queries.md` | LOW | PENDING | runtime/queries/column_numeric_histogram.go:285-302 |
| [queries-11](./parts/queries.verified.md) — Aggregation export performs redundant timestamp probe | `queries.md` | LOW | PENDING | runtime/queries/metricsview_aggregation.go:165-179 |
| [queries-12](./parts/queries.verified.md) — Null-filled timeseries silently truncates | `queries.md` | LOW | PENDING | runtime/queries/metricsview_timeseries.go:328-376 |
| [reconcilers-07](./parts/reconcilers.verified.md) — Migration downgrade underflows delta | `reconcilers.md` | LOW | PENDING | runtime/reconcilers/migration.go:76-83 |
| [reconcilers-08](./parts/reconcilers.verified.md) — Rename probe mutates parser definition | `reconcilers.md` | LOW | PENDING | runtime/reconcilers/project_parser.go:624-635,676-690 |
| [reconcilers-11](./parts/reconcilers.verified.md) — Project parser watch retries immediately | `reconcilers.md` | MEDIUM | PENDING | runtime/reconcilers/project_parser.go:246-258 |
| [resolvers-06](./parts/resolvers.verified.md) — Union and API resolvers mutate shared args | `resolvers.md` | LOW | PENDING | runtime/resolvers/union.go:45-60; runtime/resolvers/api.go:41-63 |
| [runtime-core-03](./parts/runtime-core.verified.md) — Global security mutex covers template evaluation | `runtime-core.md` | MEDIUM | PENDING | runtime/security.go:275-342 |
| [runtime-core-06](./parts/runtime-core.verified.md) — markPending clears ancestor error | `runtime-core.md` | LOW | PENDING | runtime/controller.go:1161-1175 |
| [runtime-core-07](./parts/runtime-core.verified.md) — Repo returns already-called release on error | `runtime-core.md` | LOW | PENDING | runtime/connections.go:62-79 |
| [runtime-core-10](./parts/runtime-core.verified.md) — Transitive access re-expands outside cache | `runtime-core.md` | LOW | PENDING | runtime/security.go:255-283,783-819 |
| [runtime-pkg-04](./parts/runtime-pkg.verified.md) — DuckDB write handles leak OTel callbacks | `runtime-pkg.md` | MEDIUM | PENDING | runtime/pkg/rduckdb/db.go:955-957 |
| [runtime-pkg-05](./parts/runtime-pkg.verified.md) — DuckDB AST JSON uses unsafe interpolation | `runtime-pkg.md` | LOW | PENDING | runtime/pkg/duckdbsql/ast_rewrite.go:97-130 |
| [runtime-pkg-07](./parts/runtime-pkg.verified.md) — File watcher error field races | `runtime-pkg.md` | LOW | PENDING | runtime/pkg/filewatcher/filewatcher.go:84-104 |
| [runtime-pkg-09](./parts/runtime-pkg.verified.md) — Blob download eagerly collects all objects | `runtime-pkg.md` | LOW | PENDING | runtime/pkg/blob/download.go:76-83 |
| [runtime-pkg-10](./parts/runtime-pkg.verified.md) — Downloaded and copied files use broad modes | `runtime-pkg.md` | LOW | PENDING | runtime/pkg/blob/download.go:230-242; runtime/pkg/fileutil/fileutil.go:67,102 |
| [runtime-server-08](./parts/runtime-server.verified.md) — Public health leaks component errors | `runtime-server.md` | LOW | PENDING | runtime/server/health.go:88-97 |
| [runtime-server-09](./parts/runtime-server.verified.md) — Global mutex serializes SQL JSON rewrite | `runtime-server.md` | LOW | PENDING | runtime/server/queries.go:19-26 |
| [web-admin-09](./parts/web-admin.verified.md) — Org form false dirty state | `web-admin.md` | LOW | PENDING | OrgNameSettings.svelte:91-103 |
| [web-admin-10](./parts/web-admin.verified.md) — Negative public URL TTL | `web-admin.md` | LOW | PENDING | form-utils.ts:58-63 |
| [web-admin-11](./parts/web-admin.verified.md) — View-As capped at 1000 | `web-admin.md` | LOW | PENDING | ViewAsUserPopover.svelte:14-19 |

