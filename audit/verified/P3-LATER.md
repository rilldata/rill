# P3 — Later backlog

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

Low-impact hardening, cleanup, and product improvements for opportunistic work.

Count: **42**.

| Finding | Source | Original severity | Status | Current evidence |
|---|---|---:|---|---|
| [admin-auth-rbac-05](./parts/admin-auth-rbac.verified.md) — OAuth redirect missing return | `admin-auth-rbac.md` | LOW | PENDING | admin/server/auth/handlers.go:657-665 |
| [admin-auth-rbac-06](./parts/admin-auth-rbac.verified.md) — Non-constant-time token comparison | `admin-auth-rbac.md` | LOW | PENDING | admin/auth_token.go:331,352,370,387 |
| [abp-09](./parts/admin-billing-provision.verified.md) — Orb reads do not paginate | `admin-billing-provision.md` | LOW | PENDING | admin/billing/orb.go:305-317,654-658 |
| [abp-12](./parts/admin-billing-provision.verified.md) — Autoscaler dead code | `admin-billing-provision.md` | LOW | PENDING | admin/jobs/river/run_autoscaler.go:20,39-42 |
| [admin-database-02](./parts/admin-database.verified.md) — Normalized attributes discarded | `admin-database.md` | LOW | PENDING | admin/database/postgres/postgres.go:2099-2136 |
| [admin-database-07](./parts/admin-database.verified.md) — Dead incomplete access-request finder | `admin-database.md` | LOW | PENDING | admin/database/postgres/postgres.go:2772-2794 |
| [admin-database-08](./parts/admin-database.verified.md) — Unbounded virtual-file listing and dead FindUsers | `admin-database.md` | LOW | PENDING | admin/database/postgres/postgres.go:830,2921-2933 |
| [admin-projects-deploy-10](./parts/admin-projects-deploy.verified.md) — GitHub callback missing return | `admin-projects-deploy.md` | LOW | PENDING | admin/server/github.go:885-902 |
| [ai-08](./parts/ai.verified.md) — Title truncation splits UTF-8 | `ai.md` | LOW | PENDING | runtime/ai/router_agent.go:237-240 |
| [ai-10](./parts/ai.verified.md) — Unused Session.RootID contains panic paths | `ai.md` | LOW | PENDING | runtime/ai/ai.go:901-917 |
| [cli-08](./parts/cli.verified.md) — PKCE callback does not validate state | `cli.md` | LOW | PENDING | cli/pkg/local/server.go:835-845 |
| [cli-09](./parts/cli.verified.md) — Dead mismatched device-auth interface | `cli.md` | LOW | PENDING | cli/pkg/deviceauth/authenticator.go:29-32,80+ |
| [cli-11](./parts/cli.verified.md) — PKCE verifier modulo bias | `cli.md` | LOW | PENDING | cli/pkg/pkce/authenticator.go:150-159 |
| [CANV-14](./parts/deep-audit-canvas.verified.md) — Dead grid diff | `deep-audit-canvas.md` | LOW | PENDING | grid.ts:98-115 |
| [drivers-connectors-ai-07](./parts/drivers-connectors-ai.verified.md) — Slack error open link is wrong | `drivers-connectors-ai.md` | LOW | PENDING | runtime/drivers/slack/notifier.go:137-138 |
| [drivers-connectors-ai-09](./parts/drivers-connectors-ai.verified.md) — Information schema table cap | `drivers-connectors-ai.md` | LOW | PENDING | runtime/drivers/information_schema.go:68-99 |
| [drivers-object-store-13](./parts/drivers-object-store.verified.md) — Athena output resolved twice | `drivers-object-store.md` | LOW | PENDING | runtime/drivers/athena/warehouse.go:52,125 |
| [drivers-olap-06](./parts/drivers-olap.verified.md) — Drivers mutate Statement Query | `drivers-olap.md` | LOW | PENDING | runtime/drivers/clickhouse/olap.go:168; runtime/drivers/pinot/olap.go:89 |
| [drivers-olap-13](./parts/drivers-olap.verified.md) — Druid readiness query extra arg | `drivers-olap.md` | LOW | PENDING | runtime/drivers/druid/information_schema.go:92-93 |
| [drivers-olap-14](./parts/drivers-olap.verified.md) — Pinot prints errors to stdout | `drivers-olap.md` | LOW | PENDING | runtime/drivers/pinot/information_schema.go:273-277 |
| [drivers-olap-15](./parts/drivers-olap.verified.md) — Information schema 1000-table cap | `drivers-olap.md` | LOW | PENDING | runtime/drivers/information_schema.go:68-99 |
| [frontend-canvas-08](./parts/frontend-canvas.verified.md) — Redundant grid writes | `frontend-canvas.md` | LOW | PENDING | grid.ts:98-115 |
| [frontend-canvas-09](./parts/frontend-canvas.verified.md) — Dead inconsistent grid methods | `frontend-canvas.md` | LOW | PENDING | row.ts:12-53 |
| [frontend-dashboard-state-02](./parts/frontend-dashboard-state.verified.md) — ASCII-fragile decompression | `frontend-dashboard-state.md` | LOW | PENDING | compression.ts:11,22 |
| [frontend-dashboard-state-06](./parts/frontend-dashboard-state.verified.md) — Milliseconds stored as seconds | `frontend-dashboard-state.md` | LOW | PENDING | toProto.ts:218-222 |
| [frontend-dashboard-state-08](./parts/frontend-dashboard-state.verified.md) — Time-grain cleaning fallthrough | `frontend-dashboard-state.md` | LOW | PENDING | clean-url-params.ts:27-49 |
| [frontend-dashboard-viz-07](./parts/frontend-dashboard-viz.verified.md) — Shared array reverse mutation | `frontend-dashboard-viz.md` | LOW | PENDING | pivot-column-definition.ts:144,540 |
| [frontend-dashboard-viz-09](./parts/frontend-dashboard-viz.verified.md) — Broken empty width estimate | `frontend-dashboard-viz.md` | LOW | PENDING | dimension-table-utils.ts:156-157 |
| [frontend-dashboard-viz-13](./parts/frontend-dashboard-viz.verified.md) — Skeleton column miscount | `frontend-dashboard-viz.md` | LOW | PENDING | Leaderboard.svelte:295-302 |
| [frontend-runtime-client-11](./parts/frontend-runtime-client.verified.md) — Dead columnMap | `frontend-runtime-client.md` | LOW | PENDING | request-queue.ts:25,97-102 |
| [fs-shared-12](./parts/frontend-shared.verified.md) — Dead theme CSS code | `frontend-shared.md` | LOW | PENDING | css-sanitizer.ts |
| [parser-08](./parts/parser.verified.md) — default_measures error names default_dimensions | `parser.md` | LOW | PENDING | runtime/parser/parse_metrics_view.go:624-628 |
| [plg-first-run-10](./parts/plg-first-user-experience.verified.md) — Hardcoded examples | `plg-first-user-experience.md` | LOW | PENDING | welcome/constants.ts:3-28 |
| [queries-10](./parts/queries.verified.md) — Comparison measure indexes are dead | `queries.md` | LOW | PENDING | runtime/queries/metricsview_comparison_toplist.go:47-50,451-474 |
| [reconcilers-10](./parts/reconcilers.verified.md) — Summary truncation can split UTF-8 | `reconcilers.md` | LOW | PENDING | runtime/reconcilers/report.go:787-792 |
| [resolvers-07](./parts/resolvers.verified.md) — Resolver export path is unreachable | `resolvers.md` | LOW | PENDING | runtime/resolver.go:160; runtime/server/query_resolver.go:55 |
| [runtime-pkg-11](./parts/runtime-pkg.verified.md) — Log loss has no client signal | `runtime-pkg.md` | PRODUCT-IMPROVEMENT | PENDING | runtime/pkg/logbuffer/buffer.go:41-48 |
| [runtime-pkg-12](./parts/runtime-pkg.verified.md) — Blob ingestion tuning is hardcoded | `runtime-pkg.md` | PRODUCT-IMPROVEMENT | PENDING | runtime/pkg/blob/download.go:26-29 |
| [web-admin-04](./parts/web-admin.verified.md) — Embed RPC cleanup discarded | `web-admin.md` | LOW | PENDING | embed/+layout.svelte:108-113 |
| [web-admin-05](./parts/web-admin.verified.md) — Dead embed conditional | `web-admin.md` | LOW | PENDING | embed/+layout.svelte:43-53 |
| [web-admin-07](./parts/web-admin.verified.md) — Global flag not restored | `web-admin.md` | LOW | PENDING | embed/+layout.svelte:39 |
| [web-admin-08](./parts/web-admin.verified.md) — Dead billing state | `web-admin.md` | LOW | PENDING | BillingCTAHandler.ts:12-29 |

