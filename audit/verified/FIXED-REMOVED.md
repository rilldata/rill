# Fixed findings removed from the active queues

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

These source findings are verified fixed, deleted with the affected code, or refuted on the snapshot and do not appear in any active priority file.

Count: **33**.

| Finding | Source | Original severity | Status | Current evidence |
|---|---|---:|---|---|
| [abp-01](./parts/admin-billing-provision.verified.md) — ClickHouse static database-name collision | `admin-billing-provision.md` | HIGH | FIXED | admin/provisioner/clickhousestatic/provisioner.go:335-350; PR #9683 fbcb248470347d7809614a26534d0bf5f17cc916 |
| [ai-01](./parts/ai.verified.md) — File tools permit project-root traversal | `ai.md` | HIGH | FIXED | runtime/ai/util.go:66-80; runtime/drivers/repo.go:70-82; PR#9723 |
| [cli-01](./parts/cli.verified.md) — DeployProject pushed env to wrong project | `cli.md` | HIGH | FIXED | cli/pkg/local/server.go:454-457; PR #9699 89761750be0efe09c98e7fdbb0c5b802fdc6aaea |
| [STAT-2](./parts/deep-audit-state-management.verified.md) — Early-return lock leak | `deep-audit-state-management.md` | MEDIUM | FIXED | DashboardStateSync.ts:215-279; PR#9701 |
| [STAT-3](./parts/deep-audit-state-management.verified.md) — Spec/storage lock leak | `deep-audit-state-management.md` | MEDIUM | FIXED | DashboardStateSync.ts:215-279; PR#9701 |
| [deep-data-security-01](./parts/deep-data-security.verified.md) — Metrics summary leaked values across row security | `deep-data-security.md` | HIGH | FIXED | runtime/metricsview/executor/executor_summary.go:46-49,158-190; PR #9717 d40f49708adb7a73633f7b7ab2f43f8672201896 |
| [deep-metrics-sql-01](./parts/deep-metrics-sql.verified.md) — IS FALSE compiled as IS TRUE | `deep-metrics-sql.md` | HIGH | FIXED | runtime/metricsview/metricssql/filter_parser.go:138-186; PR#9712 |
| [deep-metrics-sql-02](./parts/deep-metrics-sql.verified.md) — Top-level JOIN was silently discarded | `deep-metrics-sql.md` | HIGH | FIXED | runtime/metricsview/metricssql/parser.go:190-203; PR#9721 |
| [deep-metrics-sql-03](./parts/deep-metrics-sql.verified.md) — NOT BETWEEN compiled as BETWEEN | `deep-metrics-sql.md` | HIGH | FIXED | runtime/metricsview/metricssql/filter_parser.go:115-135; PR#9719 |
| [deep-metrics-sql-04](./parts/deep-metrics-sql.verified.md) — Decimal literals were lost | `deep-metrics-sql.md` | HIGH | FIXED | runtime/metricsview/metricssql/filter_parser.go:425-448; PR#9720 |
| [deep-metrics-sql-05](./parts/deep-metrics-sql.verified.md) — Top-level SELECT silently dropped clauses | `deep-metrics-sql.md` | HIGH | FIXED | runtime/metricsview/metricssql/parser.go:76-101,135-167,239-327; PR#9716 |
| [drivers-connectors-ai-01](./parts/drivers-connectors-ai.verified.md) — RepoStore path traversal | `drivers-connectors-ai.md` | HIGH | FIXED | runtime/drivers/repo.go:150-160; PR #9723 6793dbb0265ce4d126cc7206529dfef54660ea82 |
| [drivers-object-store-03](./parts/drivers-object-store.verified.md) — BigQuery client leak | `drivers-object-store.md` | HIGH | FIXED | runtime/drivers/bigquery/warehouse.go:212-215; PR #9706 6e1615722e60ea902260bac6afad586a23c66e93 |
| [drivers-object-store-06](./parts/drivers-object-store.verified.md) — ResolveLocalPath sibling-prefix escape | `drivers-object-store.md` | MEDIUM | FIXED | runtime/pkg/fileutil/fileutil.go:317-340; PR #9726 ec882259df62e6d1d9a08987ac37945e8a2a3a89 |
| [frontend-canvas-01](./parts/frontend-canvas.verified.md) — PDF export omitted tab-group content | `frontend-canvas.md` | HIGH | FIXED | CanvasPdfExportView.svelte:21-61; PR#9722 |
| [frontend-dashboard-state-01](./parts/frontend-dashboard-state.verified.md) — Updating lock leaked on early return | `frontend-dashboard-state.md` | HIGH | FIXED | DashboardStateSync.ts:215-279; PR#9701 |
| [parser-01](./parts/parser.verified.md) — Component variable value could panic project parse | `parser.md` | HIGH | FIXED | runtime/parser/parse_component.go:29-41; PR#9702 |
| [parser-02](./parts/parser.verified.md) — Annotation validation checked parent fields | `parser.md` | MEDIUM | FIXED | runtime/parser/parse_metrics_view.go:735-756; PR#9724 |
| [queries-01](./parts/queries.verified.md) — Aggregation export panicked without time range | `queries.md` | HIGH | FIXED | runtime/queries/metricsview_aggregation.go:165-168; PR#9715 |
| [queries-05](./parts/queries.verified.md) — Timeseries Parquet schema omitted time type | `queries.md` | MEDIUM | FIXED | runtime/queries/metricsview_timeseries.go:177-186; runtime/queries/metricsview.go:343-351; PR#9725 |
| [reconcilers-01](./parts/reconcilers.verified.md) — Non-email alert dereferenced nil admin metadata | `reconcilers.md` | HIGH | FIXED | runtime/reconcilers/alert.go:699-700,892-903,927-932; PR#9707 |
| [reconcilers-02](./parts/reconcilers.verified.md) — Rename existence check used old model table | `reconcilers.md` | HIGH | FIXED | runtime/reconcilers/model.go:193-196; PR#9709 |
| [reconcilers-03](./parts/reconcilers.verified.md) — Deleting partitioned model leaked catalog rows | `reconcilers.md` | MEDIUM | FIXED | runtime/reconcilers/model.go:1127-1135; PR#9711 |
| [resolvers-01](./parts/resolvers.verified.md) — AI resolver omitted time ranges panic | `resolvers.md` | HIGH | FIXED | runtime/resolvers/ai.go:79-91; PR#9703 |
| [resolvers-02](./parts/resolvers.verified.md) — Query cache omitted security context | `resolvers.md` | HIGH | FIXED | runtime/resolver.go:208-220; PR#9704 |
| [resolvers-03](./parts/resolvers.verified.md) — Annotations bypassed metrics-view access | `resolvers.md` | HIGH | FIXED | runtime/resolvers/annotations.go:51-58; PR#9705 |
| [runtime-core-01](./parts/runtime-core.verified.md) — Resolver cache omitted AdditionalRules | `runtime-core.md` | HIGH | FIXED | runtime/resolver.go:208-220; PR#9704 |
| [runtime-pkg-01](./parts/runtime-pkg.verified.md) — ResolveLocalPath sandbox traversal | `runtime-pkg.md` | HIGH | FIXED | runtime/pkg/fileutil/fileutil.go:317-340; PR#9726 |
| [runtime-server-01](./parts/runtime-server.verified.md) — ConnectorService handlers lacked authorization | `runtime-server.md` | HIGH | FIXED | runtime/server/connector_service.go:15,42,85,121,146,179,212; PR#9700 |
| [runtime-server-02](./parts/runtime-server.verified.md) — QueryBatch sent concurrently on one stream | `runtime-server.md` | HIGH | REMOVE | RPC and runtime/server/batch_query.go removed by PR#9714 |
| [deep-ai-performance-09](./parts/deep-ai-performance.verified.md) — Refutation: no extra router or feedback call on normal web turn | `deep-ai-performance.md` | LOW | INVALID | runtime/ai/router_agent.go:102-142; runtime/ai/feedback_agent.go:68-94 |
| [deep-data-security-03](./parts/deep-data-security.verified.md) — Global timestamp bounds are intentional semantics | `deep-data-security.md` | LOW | INVALID | runtime/metricsview/executor/executor.go:170-177; PR #9717 d40f49708adb7a73633f7b7ab2f43f8672201896 |
| [fs-shared-03](./parts/frontend-shared.verified.md) — Unreachable sanitizer injection | `frontend-shared.md` | REMOVED-REFUTED | INVALID | css-sanitizer.ts:87-136 |

