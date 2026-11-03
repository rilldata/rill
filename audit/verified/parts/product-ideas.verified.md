# Verified audit: product-ideas

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

The source contains **55 actionable but unnumbered `###` idea headings**. To make coverage auditable, this file assigns stable IDs `PI-01` through `PI-55` in source order. Every idea heading is represented exactly once. These are product proposals, so “pending” means the proposed end-state is not implemented; “partial” means a material primitive exists but the proposal remains incomplete.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| PI-01 | M standout | P1 | Seeded datasets and offline sample generator | `features/sample-data/generate-sample-data.ts:59-70` remains chat-only | Bundle deterministic DuckDB/Parquet paths and fail over automatically. |
| PI-02 | M | P2 | Reconcile-driven first-dashboard progress | `welcome/ProjectCards.svelte:62-67` still waits 5s and reloads | Drive a bounded progress surface from resource state. |
| PI-03 | M | P2 | Onboarding checklist and funnel telemetry | `cli/pkg/local/app.go:544-556` still drops AppStart on connector errors; no checklist exists | Fix event emission and add resource-backed activation steps/events. |
| PI-05 | M | P3 | First-run dashboard narrative | Generated-dashboard flow has no summary surface; `generate_canvas_dashboard.go:818-927` only emits components | Add a non-blocking narrative with deterministic fallback. |
| PI-07 | M standout | P0 | Unified object-store policy boundary | Access checks remain split across per-driver list/glob/download/model paths (`runtime/drivers/{s3,gcs,azure}/object_store.go`) | Introduce one delimiter-aware authorization function and adversarial tests. |
| PI-08 | S | P1 | Bucket preflight and SSRF guardrails | `runtime/drivers/https/https.go:259` still uses general HTTP fetching for configured URLs | Validate bucket access and deny private/metadata destinations with explicit opt-in. |
| PI-09 | M/L | P2 | Glob caps, progress, resumable downloads | `runtime/pkg/blob/download.go` still materializes listings before download | Add object/byte caps, streaming manifest, progress and resume. |
| PI-10 | M | P2 | Cross-schema pagination and uniform size | `runtime/drivers/information_schema.go` and per-driver estimators still expose uneven limits/`-1` sizes | Cursor through schemas/tables and normalize estimates. |
| PI-11 | M | P2 | `.env` secret manager | `.env` is edited through file/env actions; no reference/orphan management view exists | Build reference graph, masked list, and safe removal. |
| PI-12 | L standout | P2 | Ratio/per-dimension measures | `runtime/metricsview/ast.go` still reports `support for "per" not implemented` | Implement semantic ratio planning and dialect tests. |
| PI-13 | M standout | P1 | Structured query validation API | `runtime/metricsview/executor/executor.go:165-168` still panics in `ValidateQuery` | Return typed field/grain/filter/limit violations before execution. |
| PI-14 | M | P1 | Percent-of-total in one window query | Existing rewrite and frontend still have zero/NULL edge paths (`leaderboard-utils.ts:105-106,215-216`) | Use dialect window support with guarded fallback. |
| PI-15 | M | P2 | Native regex filter operator | Frontend modes still omit regex and dialect conversions remain heterogeneous | Add typed operator with per-dialect escaping/capability checks. |
| PI-16 | M | P2 | Partition recovery and rebuild reasons | Reconciler tracks partition states but no complete per-partition recovery/trigger explanation UI exists | Expose failed partitions, targeted rerun, and persisted trigger reason. |
| PI-17 | M | P2 | Semantic diff and impact preview | Parser DAG exists, while edit flows do not preview removed-field consumers | Compute resource/field impact before save and link affected resources. |
| PI-20 | M | P2 | Persistent widths and clean links | Widths remain local and exports still have sort-parity findings (`dimension-table-export.ts:103-148`; `pivot-export.ts:155-181`) | Persist width state and reuse canonical sort/link serialization. |
| PI-21 | M | P2 | Globalize/export pivot selection | Pivot selection primitives exist but expose neither promotion nor selection-only export | Convert selected axes to global filters/export bounds. |
| PI-22 | M | P3 | Human-readable Rill-time editor | Parser/time controls exist but no editable expression surface is wired | Add validated expression editor with named chips. |
| PI-23 | S | P3 | State debugger and cross-measure hover | Current loader/chart state has no source-diff overlay or shared hover bus | Add dev overlay and chart hover context. |
| PI-25 | M standout | P2 | Standard Fix-with-AI recovery | Error surfaces still diverge (`ProjectCards.svelte:76-78`; `actions.ts:162-179`) | Reuse one Retry/no-AI/edit/fix action set. |
| PI-27 | M | P2 | Conversation cost meter and budgets | Token metrics exist internally but no displayed/enforced session budget | Aggregate usage, enforce budget, and end gracefully. |
| PI-29 | L | P3 | Configurable agents, summaries, citations | Agent router remains fixed and long-context/citation handling is not a stable structured contract | Add registry, rolling summary and citation objects. |
| PI-31 | M standout | P0 | Scoped magic tokens and origin-pinned embeds | `web-common/src/lib/rpc.ts:64-104` still accepts any origin/replies `*`; token attribute scope remains broad | Add least-privilege token profiles and parent-origin handshake. |
| PI-33 | M | P1 | Idempotent per-recipient notification delivery | Report/notifier delivery still lacks an execution idempotency contract and isolated structured results | Persist delivery keys/status and retry failed recipients only. |
| PI-34 | L | P3 | Dashboard comments and annotations | Bookmark storage exists, but no comment/annotation resource or UI exists | Add scoped threads/anchors, permissions and notifications. |
| PI-35 | L standout | P1 | Streaming/chunked exports | Download path remains request/response streaming without resumable staged artifacts | Stage chunks in object storage and return signed resumable downloads. |
| PI-37 | M | P1 | Bound in-memory query results | Non-export resolve paths still buffer complete `Struct` slices and AI limits need lower-bound enforcement | Stream/cap decode and return explicit too-large errors. |
| PI-40 | M standout | P1 | Reversible/bulk/conflict-aware file operations | `entity-management/actions/actions.ts:206-276` still swallows failures/overwrites copies; no trash/bulk UI | Add trash/undo, multi-select and conflict dialog with transactional results. |
| PI-41 | M standout | P1 | Deploy dry-run, profiles and CI login | CLI has no unified non-mutating deploy-plan/profile/CI token workflow | Add dry-run plan and explicit credential/profile selection. |
| PI-43 | M | P2 | Project verification and migration dry-run | Parser/reconciler expose errors but no one-shot DAG/unused/deprecation/migration plan command | Build read-only verify/plan command with machine output. |
| PI-44 | M/S | P1 | Keychain credentials and deterministic port | CLI credential/local-start paths still rely on plaintext config and timing-based port behavior | Integrate OS keyring and reserve/print a verified port. |
| PI-48 | M | P1 | Superuser audit, revocation, safe redirects | `web-admin` still has open redirect sinks and no single complete support-action audit surface | Persist actions, evict revoked sessions, centralize redirect validation. |
| PI-49 | L/M | P1 | Data-access audit log and policy tester | View-As remains fallible (`project-runtime.ts:38-39`) and there is no query-level user/token/resource audit surface | Persist access events and provide a fail-closed arbitrary-attributes policy simulator. |
| PI-50 | S standout | P3 | Post-deploy share growth loop | Deploy/local completion has no one consolidated share success/nudge | Offer copy/invite/open plus a one-time CLI deploy hint. |
| PI-54 | M | P3 | Per-embed analytics | Embed API has no origin handshake or per-parent usage telemetry | Complete origin identity then emit privacy-scoped embed events. |

## Fixed / remove

None of the 55 complete proposals is fully delivered at this snapshot.

## Partial / needs test

| ID | Original severity | Priority | Existing implementation / evidence | Remaining work |
|---|---|---|---|---|
| PI-04 | M | P2 | Backend `ListExamples` exists at `runtime/server/examples.go:19-37` | Frontend remains `welcome/constants.ts:3-28`; add metadata, preflight, badges, errors. |
| PI-06 | M standout | P1 | Connector validation/reconcile and explorer primitives exist (`add-data/manager/steps/connector.ts:97-166`) | Validate against temporary spec before writing files/secrets; implement real AI connector Ping. |
| PI-18 | M standout | P1 | PR #9701 (`7aa5ebd88`) fixed the permanent lock leak with guard ordering/finally | Replace remaining boolean race/drop/double-init behavior with generations. |
| PI-19 | M standout | P2 | PR #9722 (`0b6b50df3`) added active/all tab rendering for canvas PDF | Complete PNG parity and partial-failure summary; verify all scheduled/export paths. |
| PI-24 | M standout | P0 | PRs #9723 (`6793dbb02`) and #9726 (`ec882259d`) jail AI/repo/local paths | AI SQL remains mutable despite read-only hint; add read-only role/connection and centralize all path sites. |
| PI-26 | M standout | P2 | AI/report/notifier infrastructure exists | Deliver a reliable narrative report end-to-end, including missing-time-range handling. |
| PI-28 | M | P1 | Tool annotations identify destructive operations | UI confirmation, diff/statement preview, durable tool audit are still absent. |
| PI-30 | XL/S standout | P2 | Bookmarks persist explore/canvas state (`BookmarksFormDialog.svelte:184-243`) | Add short server tokens and universal clean-link UX. |
| PI-32 | M | P1 | Public URL creation/list/revoke UI exists under `web-admin/src/features/public-urls` | Add past-date validation, audit columns/countdown, and cache-evicting immediate revoke. |
| PI-36 | S/M standout | P2 | PR #9704 (`4564e65ac`) hashes full `SecurityClaims`, including rules/permissions/skip checks, in resolver cache keys | Add the proposal's exhaustive per-resolver/legacy-query correctness harness. |
| PI-38 | M/L | P2 | Rollup and approximate-comparison rewriters exist | Extend eligible time dimensions/measures and delta/ratio sort support. |
| PI-39 | S/M | P2 | Conncache/provisioner metrics exist internally | Expose actionable handles/slot health and reclaim orphan resources. |
| PI-42 | S/M | P1 | Parser has schemas and resource-specific validation | Add uniform schema linting, deprecation warnings, and per-node panic isolation. |
| PI-45 | S standout | P0 | PRs #9723/#9726 added containment checks and regression tests for repo/AI/local path sites | Consolidate to one symlink-aware helper and cover blob/remaining callers. |
| PI-46 | S standout | P2 | PR #9700 (`d4553f382`) added authorization to all ConnectorService handlers and tests | Generalize this into a generated/table-driven harness covering every service and RPC. |
| PI-47 | S/M | P2 | PR #9705 (`9367e47a2`) added the missing annotations `CanAccess` gate | Decide/test primary-time field semantics and enforce a uniform resolver invariant. |
| PI-51 | M standout | P2 | AI providers already record token metrics | Enforce/display per-session budgets and connect them to billing usage. |
| PI-52 | M standout | P2 | `runtime/resolvers/project_storage.go` and status usage UI expose project/connector storage | Add full org usage dashboard, parallel sizing, query/AI/seat/credit metrics. |
| PI-53 | M/L | P2 | Billing jobs/issues/reconcile primitives exist under `admin/billing` | Build operator console and one convergent idempotent repair path. |
| PI-55 | M/L | P1 | Provisioner/autoscaler primitives and flags exist | Add capacity gauges/alerts, safe policies/dry run, and fix collision/TOCTOU risks. |

Counts: **35 pending, 0 fixed/remove, 20 partial = 55**.

## Machine-readable inventory

```text
PENDING|product-ideas.md|PI-01|M-standout|P1|-|Seeded offline sample generator|generate-sample-data.ts:59-70
PENDING|product-ideas.md|PI-02|M|P2|-|Reconcile first-dashboard progress|ProjectCards.svelte:62-67
PENDING|product-ideas.md|PI-03|M|P2|-|Checklist and funnel telemetry|cli/pkg/local/app.go:544-556
PARTIAL|product-ideas.md|PI-04|M|P2|-|Backend example gallery|runtime/server/examples.go:19-37
PENDING|product-ideas.md|PI-05|M|P3|-|Dashboard narrative|generate_canvas_dashboard.go:818-927
PARTIAL|product-ideas.md|PI-06|M-standout|P1|-|Clean connector test drive|connector.ts:97-166
PENDING|product-ideas.md|PI-07|M-standout|P0|-|Object-store policy boundary|drivers/s3/object_store.go
PENDING|product-ideas.md|PI-08|S|P1|-|Bucket preflight and SSRF guards|drivers/https/https.go:259
PENDING|product-ideas.md|PI-09|M-L|P2|-|Glob caps progress resume|pkg/blob/download.go
PENDING|product-ideas.md|PI-10|M|P2|-|Schema pagination and sizes|drivers/information_schema.go
PENDING|product-ideas.md|PI-11|M|P2|-|Env secret manager|env-management
PENDING|product-ideas.md|PI-12|L-standout|P2|-|Ratio per measures|metricsview/ast.go
PENDING|product-ideas.md|PI-13|M-standout|P1|-|Structured query validation|executor/executor.go:165-168
PENDING|product-ideas.md|PI-14|M|P1|-|Window percent total|leaderboard-utils.ts:105-106
PENDING|product-ideas.md|PI-15|M|P2|-|Regex filter operator|DimensionFilterMode
PENDING|product-ideas.md|PI-16|M|P2|-|Partition recovery diagnostics|reconcilers/model.go
PENDING|product-ideas.md|PI-17|M|P2|-|Semantic impact preview|runtime/parser
PARTIAL|product-ideas.md|PI-18|M-standout|P1|-|Deterministic state sync|DashboardStateSync.ts; PR#9701
PARTIAL|product-ideas.md|PI-19|M-standout|P2|-|Tab-aware complete export|CanvasPdfExportView.svelte; PR#9722
PENDING|product-ideas.md|PI-20|M|P2|-|Persistent widths clean links|dimension-table-export.ts:103-148
PENDING|product-ideas.md|PI-21|M|P2|-|Promote export pivot selection|pivot-click-selection.ts
PENDING|product-ideas.md|PI-22|M|P3|-|Rill-time editor|time-controls
PENDING|product-ideas.md|PI-23|S|P3|-|State debugger hover sync|DashboardStateDataLoader.ts
PARTIAL|product-ideas.md|PI-24|M-standout|P0|-|Sandboxed AI|PR#9723; PR#9726; query_sql.go
PENDING|product-ideas.md|PI-25|M-standout|P2|-|Fix-with-AI recovery|actions.ts:162-179
PARTIAL|product-ideas.md|PI-26|M-standout|P2|-|Scheduled AI narratives|runtime/resolvers/ai.go
PENDING|product-ideas.md|PI-27|M|P2|-|Conversation cost budget|runtime/ai/ai.go
PARTIAL|product-ideas.md|PI-28|M|P1|-|AI approval audit|runtime/ai tool annotations
PENDING|product-ideas.md|PI-29|L|P3|-|Configurable summarized agents|runtime/ai
PARTIAL|product-ideas.md|PI-30|XL-S-standout|P2|-|Short state-token links|BookmarksFormDialog.svelte:184-243
PENDING|product-ideas.md|PI-31|M-standout|P0|-|Scoped tokens pinned embeds|rpc.ts:64-104
PARTIAL|product-ideas.md|PI-32|M|P1|-|Public URL management revocation|features/public-urls
PENDING|product-ideas.md|PI-33|M|P1|-|Reliable notification delivery|reconcilers/report.go
PENDING|product-ideas.md|PI-34|L|P3|-|Dashboard comments|features/bookmarks
PENDING|product-ideas.md|PI-35|L-standout|P1|-|Streaming chunked exports|runtime/server/downloads.go
PARTIAL|product-ideas.md|PI-36|S-M-standout|P2|-|Security-aware cache keys|runtime/resolver.go:208-220; PR#9704
PENDING|product-ideas.md|PI-37|M|P1|-|Capped streamed query results|rowsToData
PARTIAL|product-ideas.md|PI-38|M-L|P2|-|Extended preaggregation|executor_rewrite_rollup.go
PARTIAL|product-ideas.md|PI-39|S-M|P2|-|Connection provisioner observability|conncache
PENDING|product-ideas.md|PI-40|M-standout|P1|-|Reversible file operations|actions.ts:206-276
PENDING|product-ideas.md|PI-41|M-standout|P1|-|Deploy dry-run profiles CI|cli/cmd/deploy
PARTIAL|product-ideas.md|PI-42|S-M|P1|-|Parser linter isolation|runtime/parser
PENDING|product-ideas.md|PI-43|M|P2|-|Project verify migration plan|runtime/parser
PENDING|product-ideas.md|PI-44|M-S|P1|-|Keychain deterministic port|dotrill.go
PARTIAL|product-ideas.md|PI-45|S-standout|P0|-|Unified SecureJoin|PR#9723; PR#9726
PARTIAL|product-ideas.md|PI-46|S-standout|P2|-|Service authz harness|connector_service.go; PR#9700
PARTIAL|product-ideas.md|PI-47|S-M|P2|-|Field-access consistency|resolvers/annotations.go; PR#9705
PENDING|product-ideas.md|PI-48|M|P1|-|Superuser audit safe redirects|web-admin upgrade-callback
PENDING|product-ideas.md|PI-49|L-M|P1|-|Data-access audit and policy tester|project-runtime.ts:38-39
PENDING|product-ideas.md|PI-50|S-standout|P3|-|Post-deploy share loop|web-local deploy routes
PARTIAL|product-ideas.md|PI-51|M-standout|P2|-|AI cost budget metering|runtime/ai
PARTIAL|product-ideas.md|PI-52|M-standout|P2|-|Usage consumption dashboard|resolvers/project_storage.go
PARTIAL|product-ideas.md|PI-53|M-L|P2|-|Billing issue console|admin/billing
PENDING|product-ideas.md|PI-54|M|P3|-|Embed telemetry|EmbedStore
PARTIAL|product-ideas.md|PI-55|M-L|P1|-|Capacity metrics safe autoscaler|admin/provisioner
```
