# Verified product and PLG proposal backlog

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

These are roadmap proposals rather than live defects. Their status records whether the underlying capability is absent or partially delivered on the snapshot.

Count: **84**.

| Finding | Source | Original severity | Status | Current evidence |
|---|---|---:|---|---|
| [PI-07](./parts/product-ideas.verified.md) — Object-store policy boundary | `product-ideas.md` | M-STANDOUT | PENDING | drivers/s3/object_store.go |
| [PI-31](./parts/product-ideas.verified.md) — Scoped tokens pinned embeds | `product-ideas.md` | M-STANDOUT | PENDING | rpc.ts:64-104 |
| [PI-24](./parts/product-ideas.verified.md) — Sandboxed AI | `product-ideas.md` | M-STANDOUT | PARTIAL | PR#9723; PR#9726; query_sql.go |
| [PI-45](./parts/product-ideas.verified.md) — Unified SecureJoin | `product-ideas.md` | S-STANDOUT | PARTIAL | PR#9723; PR#9726 |
| [PLG-1](./parts/plg_improvement_ideas.verified.md) — Resilient sample generation | `plg_improvement_ideas.md` | M-NOW | PENDING | generate-sample-data.ts:59-70 |
| [PLG-3](./parts/plg_improvement_ideas.verified.md) — Surface onboarding failures | `plg_improvement_ideas.md` | M-NOW | PENDING | actions.ts:162-179 |
| [PLG-5](./parts/plg_improvement_ideas.verified.md) — Fix deploy org route | `plg_improvement_ideas.md` | L-NOW | PENDING | deploy/+page.ts:4-9 |
| [NFI-1](./parts/plg-first-user-experience.verified.md) — Offline sample generator | `plg-first-user-experience.md` | M-IDEA | PENDING | generate-sample-data.ts:59-70 |
| [NFI-8](./parts/plg-first-user-experience.verified.md) — Standard error recovery | `plg-first-user-experience.md` | M-IDEA | PENDING | actions.ts:162-179 |
| [PI-01](./parts/product-ideas.verified.md) — Seeded offline sample generator | `product-ideas.md` | M-STANDOUT | PENDING | generate-sample-data.ts:59-70 |
| [PI-08](./parts/product-ideas.verified.md) — Bucket preflight and SSRF guards | `product-ideas.md` | S | PENDING | drivers/https/https.go:259 |
| [PI-13](./parts/product-ideas.verified.md) — Structured query validation | `product-ideas.md` | M-STANDOUT | PENDING | executor/executor.go:165-168 |
| [PI-14](./parts/product-ideas.verified.md) — Window percent total | `product-ideas.md` | M | PENDING | leaderboard-utils.ts:105-106 |
| [PI-33](./parts/product-ideas.verified.md) — Reliable notification delivery | `product-ideas.md` | M | PENDING | reconcilers/report.go |
| [PI-35](./parts/product-ideas.verified.md) — Streaming chunked exports | `product-ideas.md` | L-STANDOUT | PENDING | runtime/server/downloads.go |
| [PI-37](./parts/product-ideas.verified.md) — Capped streamed query results | `product-ideas.md` | M | PENDING | rowsToData |
| [PI-40](./parts/product-ideas.verified.md) — Reversible file operations | `product-ideas.md` | M-STANDOUT | PENDING | actions.ts:206-276 |
| [PI-41](./parts/product-ideas.verified.md) — Deploy dry-run profiles CI | `product-ideas.md` | M-STANDOUT | PENDING | cli/cmd/deploy |
| [PI-44](./parts/product-ideas.verified.md) — Keychain deterministic port | `product-ideas.md` | M-S | PENDING | dotrill.go |
| [PI-48](./parts/product-ideas.verified.md) — Superuser audit safe redirects | `product-ideas.md` | M | PENDING | web-admin upgrade-callback |
| [PI-49](./parts/product-ideas.verified.md) — Data-access audit and policy tester | `product-ideas.md` | L-M | PENDING | project-runtime.ts:38-39 |
| [PI-06](./parts/product-ideas.verified.md) — Clean connector test drive | `product-ideas.md` | M-STANDOUT | PARTIAL | connector.ts:97-166 |
| [PI-18](./parts/product-ideas.verified.md) — Deterministic state sync | `product-ideas.md` | M-STANDOUT | PARTIAL | DashboardStateSync.ts; PR#9701 |
| [PI-28](./parts/product-ideas.verified.md) — AI approval audit | `product-ideas.md` | M | PARTIAL | runtime/ai tool annotations |
| [PI-32](./parts/product-ideas.verified.md) — Public URL management revocation | `product-ideas.md` | M | PARTIAL | features/public-urls |
| [PI-42](./parts/product-ideas.verified.md) — Parser linter isolation | `product-ideas.md` | S-M | PARTIAL | runtime/parser |
| [PI-55](./parts/product-ideas.verified.md) — Capacity metrics safe autoscaler | `product-ideas.md` | M-L | PARTIAL | admin/provisioner |
| [PLG-2](./parts/plg_improvement_ideas.verified.md) — Remove five-second delay | `plg_improvement_ideas.md` | L-NOW | PENDING | ProjectCards.svelte:62-67 |
| [PLG-4](./parts/plg_improvement_ideas.verified.md) — One dashboard per source | `plg_improvement_ideas.md` | L-NOW | PENDING | import.ts:60-65,90-114 |
| [PLG-6](./parts/plg_improvement_ideas.verified.md) — Funnel telemetry | `plg_improvement_ideas.md` | L-NOW | PENDING | cli/pkg/local/app.go:544-556 |
| [PLG-9](./parts/plg_improvement_ideas.verified.md) — Cloud zero-org empty state | `plg_improvement_ideas.md` | M-NOW | PENDING | web-admin/src/routes/+page.svelte:27 |
| [PLG-12](./parts/plg_improvement_ideas.verified.md) — Dimension-less explainer | `plg_improvement_ideas.md` | L-NEXT | PENDING | generate_metrics_view.go:272-292 |
| [NFI-2](./parts/plg-first-user-experience.verified.md) — Seeded datasets | `plg-first-user-experience.md` | M-IDEA | PENDING | welcome/constants.ts:3-28 |
| [NFI-3](./parts/plg-first-user-experience.verified.md) — Reconcile progress | `plg-first-user-experience.md` | M-IDEA | PENDING | ProjectCards.svelte:62-67 |
| [NFI-5](./parts/plg-first-user-experience.verified.md) — Connection preview | `plg-first-user-experience.md` | M-IDEA | PENDING | steps/connector.ts:112-166 |
| [NFI-7](./parts/plg-first-user-experience.verified.md) — Backend example gallery | `plg-first-user-experience.md` | M-IDEA | PENDING | runtime/server/examples.go:19-37 |
| [NFI-9](./parts/plg-first-user-experience.verified.md) — Activation telemetry | `plg-first-user-experience.md` | S-IDEA | PENDING | cli/pkg/local/app.go:544-556 |
| [PI-02](./parts/product-ideas.verified.md) — Reconcile first-dashboard progress | `product-ideas.md` | M | PENDING | ProjectCards.svelte:62-67 |
| [PI-03](./parts/product-ideas.verified.md) — Checklist and funnel telemetry | `product-ideas.md` | M | PENDING | cli/pkg/local/app.go:544-556 |
| [PI-09](./parts/product-ideas.verified.md) — Glob caps progress resume | `product-ideas.md` | M-L | PENDING | pkg/blob/download.go |
| [PI-10](./parts/product-ideas.verified.md) — Schema pagination and sizes | `product-ideas.md` | M | PENDING | drivers/information_schema.go |
| [PI-11](./parts/product-ideas.verified.md) — Env secret manager | `product-ideas.md` | M | PENDING | env-management |
| [PI-12](./parts/product-ideas.verified.md) — Ratio per measures | `product-ideas.md` | L-STANDOUT | PENDING | metricsview/ast.go |
| [PI-15](./parts/product-ideas.verified.md) — Regex filter operator | `product-ideas.md` | M | PENDING | DimensionFilterMode |
| [PI-16](./parts/product-ideas.verified.md) — Partition recovery diagnostics | `product-ideas.md` | M | PENDING | reconcilers/model.go |
| [PI-17](./parts/product-ideas.verified.md) — Semantic impact preview | `product-ideas.md` | M | PENDING | runtime/parser |
| [PI-20](./parts/product-ideas.verified.md) — Persistent widths clean links | `product-ideas.md` | M | PENDING | dimension-table-export.ts:103-148 |
| [PI-21](./parts/product-ideas.verified.md) — Promote export pivot selection | `product-ideas.md` | M | PENDING | pivot-click-selection.ts |
| [PI-25](./parts/product-ideas.verified.md) — Fix-with-AI recovery | `product-ideas.md` | M-STANDOUT | PENDING | actions.ts:162-179 |
| [PI-27](./parts/product-ideas.verified.md) — Conversation cost budget | `product-ideas.md` | M | PENDING | runtime/ai/ai.go |
| [PI-43](./parts/product-ideas.verified.md) — Project verify migration plan | `product-ideas.md` | M | PENDING | runtime/parser |
| [PLG-7](./parts/plg_improvement_ideas.verified.md) — Connection preview before commit | `plg_improvement_ideas.md` | M-NOW | PARTIAL | connector.ts:97-166 |
| [PLG-11](./parts/plg_improvement_ideas.verified.md) — Legible build progress | `plg_improvement_ideas.md` | M-NEXT | PARTIAL | DashboardBuilding.svelte |
| [PLG-14](./parts/plg_improvement_ideas.verified.md) — Backend-driven example gallery | `plg_improvement_ideas.md` | M-NEXT | PARTIAL | runtime/server/examples.go:19-37 |
| [PI-04](./parts/product-ideas.verified.md) — Backend example gallery | `product-ideas.md` | M | PARTIAL | runtime/server/examples.go:19-37 |
| [PI-19](./parts/product-ideas.verified.md) — Tab-aware complete export | `product-ideas.md` | M-STANDOUT | PARTIAL | CanvasPdfExportView.svelte; PR#9722 |
| [PI-26](./parts/product-ideas.verified.md) — Scheduled AI narratives | `product-ideas.md` | M-STANDOUT | PARTIAL | runtime/resolvers/ai.go |
| [PI-30](./parts/product-ideas.verified.md) — Short state-token links | `product-ideas.md` | XL-S-STANDOUT | PARTIAL | BookmarksFormDialog.svelte:184-243 |
| [PI-36](./parts/product-ideas.verified.md) — Security-aware cache keys | `product-ideas.md` | S-M-STANDOUT | PARTIAL | runtime/resolver.go:208-220; PR#9704 |
| [PI-38](./parts/product-ideas.verified.md) — Extended preaggregation | `product-ideas.md` | M-L | PARTIAL | executor_rewrite_rollup.go |
| [PI-39](./parts/product-ideas.verified.md) — Connection provisioner observability | `product-ideas.md` | S-M | PARTIAL | conncache |
| [PI-46](./parts/product-ideas.verified.md) — Service authz harness | `product-ideas.md` | S-STANDOUT | PARTIAL | connector_service.go; PR#9700 |
| [PI-47](./parts/product-ideas.verified.md) — Field-access consistency | `product-ideas.md` | S-M | PARTIAL | resolvers/annotations.go; PR#9705 |
| [PI-51](./parts/product-ideas.verified.md) — AI cost budget metering | `product-ideas.md` | M-STANDOUT | PARTIAL | runtime/ai |
| [PI-52](./parts/product-ideas.verified.md) — Usage consumption dashboard | `product-ideas.md` | M-STANDOUT | PARTIAL | resolvers/project_storage.go |
| [PI-53](./parts/product-ideas.verified.md) — Billing issue console | `product-ideas.md` | M-L | PARTIAL | admin/billing |
| [PLG-8](./parts/plg_improvement_ideas.verified.md) — Onboarding checklist | `plg_improvement_ideas.md` | M-NOW | PENDING | features/welcome |
| [PLG-10](./parts/plg_improvement_ideas.verified.md) — Dashboard-first sharing | `plg_improvement_ideas.md` | L-NEXT | PENDING | web-local deploy routes |
| [PLG-13](./parts/plg_improvement_ideas.verified.md) — Intent-aware splash | `plg_improvement_ideas.md` | L-NEXT | PENDING | welcome/+page.svelte |
| [PLG-15](./parts/plg_improvement_ideas.verified.md) — Activation lifecycle messaging | `plg_improvement_ideas.md` | M-NEXT | PENDING | web-admin/src/features/billing |
| [PLG-16](./parts/plg_improvement_ideas.verified.md) — CLI deploy nudge | `plg_improvement_ideas.md` | L-NEXT | PENDING | cli/pkg/local/app.go |
| [PLG-17](./parts/plg_improvement_ideas.verified.md) — Dashboard narrative | `plg_improvement_ideas.md` | M-LATER | PENDING | generate_canvas_dashboard.go:818-927 |
| [PLG-18](./parts/plg_improvement_ideas.verified.md) — Progressive trial expiry | `plg_improvement_ideas.md` | M-LATER | PENDING | web-admin/src/features/billing |
| [NFI-4](./parts/plg-first-user-experience.verified.md) — Onboarding checklist | `plg-first-user-experience.md` | M-IDEA | PENDING | features/welcome |
| [NFI-6](./parts/plg-first-user-experience.verified.md) — Post-deploy share funnel | `plg-first-user-experience.md` | S-IDEA | PENDING | web-local deploy routes |
| [NFI-10](./parts/plg-first-user-experience.verified.md) — CLI deploy nudge | `plg-first-user-experience.md` | S-IDEA | PENDING | cli/pkg/local/app.go |
| [PI-05](./parts/product-ideas.verified.md) — Dashboard narrative | `product-ideas.md` | M | PENDING | generate_canvas_dashboard.go:818-927 |
| [PI-22](./parts/product-ideas.verified.md) — Rill-time editor | `product-ideas.md` | M | PENDING | time-controls |
| [PI-23](./parts/product-ideas.verified.md) — State debugger hover sync | `product-ideas.md` | S | PENDING | DashboardStateDataLoader.ts |
| [PI-29](./parts/product-ideas.verified.md) — Configurable summarized agents | `product-ideas.md` | L | PENDING | runtime/ai |
| [PI-34](./parts/product-ideas.verified.md) — Dashboard comments | `product-ideas.md` | L | PENDING | features/bookmarks |
| [PI-50](./parts/product-ideas.verified.md) — Post-deploy share loop | `product-ideas.md` | S-STANDOUT | PENDING | web-local deploy routes |
| [PI-54](./parts/product-ideas.verified.md) — Embed telemetry | `product-ideas.md` | M | PENDING | EmbedStore |
| [PLG-19](./parts/plg_improvement_ideas.verified.md) — Viewer creator on-ramp | `plg_improvement_ideas.md` | M-LATER | PARTIAL | BookmarksFormDialog.svelte:184-243 |

