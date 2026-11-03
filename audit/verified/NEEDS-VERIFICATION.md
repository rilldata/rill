# Partial, dormant, or test-dependent findings

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

These findings are not cleanly actionable until the named reachability or regression test is completed.

Count: **32**.

| Finding | Source | Original severity | Status | Current evidence |
|---|---|---:|---|---|
| [queries-03](./parts/queries.verified.md) — StarRocks single-row stats can dereference nil sd | `queries.md` | MEDIUM | PARTIAL | runtime/queries/column_desc_stats.go:88-97 |
| [runtime-core-04](./parts/runtime-core.verified.md) — Empty user ID can match missing owner annotation | `runtime-core.md` | LOW | PARTIAL | runtime/security.go:443-452,519-528 |
| [runtime-pkg-06](./parts/runtime-pkg.verified.md) — Blob object key can escape download temp directory | `runtime-pkg.md` | MEDIUM | PARTIAL | runtime/pkg/blob/download.go:230-242 |
| [admin-auth-rbac-04](./parts/admin-auth-rbac.verified.md) — Runtime proxy uses production permission for branches | `admin-auth-rbac.md` | LOW | NEEDS_TEST | admin/server/runtime_proxy.go:40-93 |
| [admin-database-03](./parts/admin-database.verified.md) — Migration advisory lock uses connection pool | `admin-database.md` | MEDIUM | NEEDS_TEST | admin/database/postgres/migrate.go:38-50 |
| [drivers-object-store-04](./parts/drivers-object-store.verified.md) — Object-key zip slip | `drivers-object-store.md` | MEDIUM | NEEDS_TEST | runtime/pkg/blob/download.go:230+ |
| [deep-metrics-sql-14](./parts/deep-metrics-sql.verified.md) — Template resolution precedes parsing | `deep-metrics-sql.md` | LOW | PARTIAL | runtime/resolvers/metrics_sql.go:62-68,82-119 |
| [metricsview-01](./parts/metricsview.verified.md) — Nested limit pushdown can lose OFFSET | `metricsview.md` | MEDIUM | PARTIAL | runtime/metricsview/executor/executor_rewrite_limit_pushdown.go:10-74 |
| [queries-06](./parts/queries.verified.md) — Unexpected timestamp scan type panics | `queries.md` | LOW | PARTIAL | runtime/queries/metricsview_timeseries.go:247-260; runtime/queries/column_timeseries.go:190-197 |
| [queries-08](./parts/queries.verified.md) — Column time range asserts max type under min guard | `queries.md` | LOW | PARTIAL | runtime/queries/column_time_range.go:100-114 |
| [reconcilers-09](./parts/reconcilers.verified.md) — Patch mode persists hash before execution | `reconcilers.md` | LOW | PARTIAL | runtime/reconcilers/model.go:1986-2009 |
| [runtime-core-08](./parts/runtime-core.verified.md) — Hung reconciler can leak on completion send | `runtime-core.md` | LOW | PARTIAL | runtime/controller.go:148,320-350,1335 |
| [runtime-server-06](./parts/runtime-server.verified.md) — JWT claims mutate and alias attributes | `runtime-server.md` | LOW | PARTIAL | runtime/server/auth/claims.go:29-36,66-70 |
| [admin-auth-rbac-11](./parts/admin-auth-rbac.verified.md) — First-user superuser race | `admin-auth-rbac.md` | LOW | NEEDS_TEST | admin/user.go:168-189 |
| [drivers-connectors-ai-06](./parts/drivers-connectors-ai.verified.md) — Gemini usage metadata nil dereference | `drivers-connectors-ai.md` | LOW | NEEDS_TEST | runtime/drivers/gemini/gemini.go:332-335 |
| [drivers-object-store-10](./parts/drivers-object-store.verified.md) — Redshift fractional timestamps rejected | `drivers-object-store.md` | MEDIUM | NEEDS_TEST | runtime/drivers/redshift/olap.go:390-401 |
| [ai-05](./parts/ai.verified.md) — Shared-session boundary fails open only if an invariant breaks | `ai.md` | LOW | PARTIAL | runtime/ai/ai.go:135-152 |
| [ai-09](./parts/ai.verified.md) — Session locking is inconsistent but currently latent | `ai.md` | LOW | PARTIAL | runtime/ai/ai.go:535-599,682-738,880-896 |
| [ai-13](./parts/ai.verified.md) — Message dirty flag is never reset | `ai.md` | LOW | PARTIAL | runtime/ai/ai.go:558-594,886-893 |
| [deep-metrics-sql-13](./parts/deep-metrics-sql.verified.md) — Alias and parser-maintenance bundle partly improved | `deep-metrics-sql.md` | LOW | PARTIAL | runtime/metricsview/metricssql/parser.go:45-53,239-273; PR#9716 |
| [parser-07](./parts/parser.verified.md) — Metrics-view explore misses project theme fallback | `parser.md` | LOW | PARTIAL | runtime/parser/parse_explore.go:227-238; runtime/parser/parse_metrics_view.go:946-970; PR#9680 |
| [parser-09](./parts/parser.verified.md) — Post-insert failures panic | `parser.md` | LOW | PARTIAL | runtime/parser/parse_metrics_view.go:850-861; runtime/parser/parse_canvas.go:244-251 |
| [resolvers-08](./parts/resolvers.verified.md) — Legacy metrics dereferences before dead nil guard | `resolvers.md` | LOW | PARTIAL | runtime/resolvers/legacy_metrics.go:129-145 |
| [runtime-core-05](./parts/runtime-core.verified.md) — Security cache key omits resource kind | `runtime-core.md` | LOW | PARTIAL | runtime/security.go:874-900 |
| [runtime-server-07](./parts/runtime-server.verified.md) — Download-token decompression has no output bound | `runtime-server.md` | LOW | PARTIAL | runtime/server/downloads.go:30-33,468-475 |
| [admin-auth-rbac-01](./parts/admin-auth-rbac.verified.md) — Cached permission pointer mutation | `admin-auth-rbac.md` | LOW | NEEDS_TEST | admin/server/auth/claims.go:201-241; admin/server/runtime_proxy.go:81-84 |
| [abp-07](./parts/admin-billing-provision.verified.md) — Dormant autoscaler nil dereferences | `admin-billing-provision.md` | LOW | NEEDS_TEST | admin/jobs/river/run_autoscaler.go:20,55-64 |
| [abp-13](./parts/admin-billing-provision.verified.md) — Trial-start email can double-send in repair race | `admin-billing-provision.md` | LOW | NEEDS_TEST | admin/billing.go:209-215; admin/jobs/river/org_jobs.go:155-164 |
| [admin-database-01](./parts/admin-database.verified.md) — Billing readers bypass transaction context | `admin-database.md` | MEDIUM | NEEDS_TEST | admin/database/postgres/postgres.go:3080-3125 |
| [cli-07](./parts/cli.verified.md) — GitHelper uses struct org instead of argument | `cli.md` | LOW | NEEDS_TEST | cli/pkg/cmdutil/helper.go:515-516 |
| [cli-10](./parts/cli.verified.md) — Device redirect double escaping | `cli.md` | LOW | NEEDS_TEST | cli/pkg/deviceauth/authenticator.go:84 |
| [drivers-connectors-ai-08](./parts/drivers-connectors-ai.verified.md) — AI Config methods retain API keys | `drivers-connectors-ai.md` | LOW | NEEDS_TEST | runtime/drivers/openai/openai.go:268-272; runtime/drivers/claude/claude.go:229-233; runtime/drivers/gemini/gemini.go:254-258 |

