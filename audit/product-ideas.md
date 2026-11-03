# Rill — Cross-Cutting Product Ideas

Opinionated roadmap fuel, synthesized from the 24 area code reviews in this folder. These go beyond the per-area feature lists: they combine related gaps into single bets, pull recurring failure modes up into product primitives, and add a few genuinely new ideas grounded in what the code already does. Each idea says what it is, why it matters, rough effort (S/M/L/XL), and which existing code it builds on. **★ marks standout bets** — the highest-leverage, most differentiated moves.

The through-line across every area: *the happy path is polished, but failure paths dead-end silently, and several powerful primitives already exist in the code but are unreachable, half-wired, or unenforced.* The best ideas below turn those into first-class features.

---

## Onboarding & PLG

### ★ One-click "aha": seeded datasets + offline sample-data generator
Ship a handful of small, real Parquet datasets (e-commerce, web analytics, ad spend) on the welcome screen, backed by a bundled DuckDB `range()`/`random()` mock generator that runs with no login and no network. Today the splash "Generate sample data" CTA drives entirely through the developer chat agent and silently dead-ends (~30s spinner, empty project) whenever AI is unavailable or the flag is off — the single worst first-run gap. A deterministic local fallback makes the fastest-possible-aha path work for the exact no-data, no-login segment most first-runs are in.
- **Value:** Kills the top onboarding dead-end; reaches a real dashboard in one click. **Effort:** M. **Builds on:** `web-common/src/features/sample-data`, `welcome/ProjectCards.svelte`, `runtime/server/examples.go`, DuckDB driver.

### Reconciliation-driven first-dashboard progress (replace the 5s hack)
Replace the fixed `setTimeout(5000)` + full-page reload after example unpack, and the opaque generate/import spinners, with a live progress panel bound to resource reconcile status, plus per-step timeouts and "continue in background." Every local example selection currently pays a guaranteed 5s of spinner regardless of whether the project is actually ready — directly inflating time-to-first-dashboard, the core PLG metric.
- **Value:** Cuts time-to-first-dashboard; turns silent waits into legible, escapable progress. **Effort:** M. **Builds on:** `ProjectCards.svelte:62`, `OnboardingWorkspace`, `waitForResourceReconciliation`.

### Guided onboarding checklist + fixed activation funnel instrumentation
A persistent "Import → Model → Metrics → Dashboard → Deploy/Share" tracker with live completion state, backed by explicit funnel events (`welcome_viewed → source_added → metrics_created → dashboard_viewed → deployed`). Also fix `emitStartEvent`, which returns a nil error instead of the connector error and silently drops the `AppStart` telemetry for any project with a misconfigured connector — skewing activation data toward already-working projects.
- **Value:** Orientation for new users + a measurable, unskewed funnel to optimize. **Effort:** M. **Builds on:** `cli/pkg/local/app.go:544`, welcome/home routes, behavioral telemetry.

### Backend-driven example gallery with reachability preflight
Drive the welcome cards from `ListExamples` (extended to carry the OLAP connector and entry-point file) instead of a hardcoded frontend list that can drift, and pre-flight-check reachability so an example needing live network/credentials (e.g. the ClickHouse GitHub-analytics example) is labeled or gated rather than landing the user on an errored dashboard. Also surface unpack failures (today swallowed by a bare `catch`).
- **Value:** No broken-example dead-ends; the gallery stays honest as examples change. **Effort:** M. **Builds on:** `welcome/constants.ts`, `runtime/server/examples.go`, `runtime/pkg/examples`.

### NEW — "Explain this dashboard" first-run narrative
After the auto-generator produces a metrics view + dashboard, offer a one-tap AI narrative ("here's what this dashboard shows, and three things worth looking at") using the analyst agent. It also gracefully covers the dimension-less-dashboard case (all-numeric tables produce a KPI-only dashboard today with no explanation), turning a confusing "empty" result into a teaching moment.
- **Value:** Converts a generated artifact into understanding; showcases the AI on first contact. **Effort:** M. **Builds on:** `runtime/ai` analyst agent, `generate_metrics_view.go`, `generate_canvas_dashboard.go`.

---

## Data Connectors & Ingestion

### ★ Guaranteed-clean connector test-drive
Validate credentials and preview available tables *against a temporary/in-memory spec before* writing any YAML or committing `.env` secrets. Today `createConnector`/`saveAiConnector` commit secrets and write the connector file *before* validation, and rollback is wired only to the explicit "Back" button — so a user who enters a real API key, fails the test, and closes the modal leaves the plaintext key in `.env` and a non-reconciling connector in the project (pushed to Cloud on deploy). Add a real `Ping` for OpenAI/Gemini/Claude (all three currently return nil, so "test connection" reports success with an invalid key).
- **Value:** Silent post-import failures become inline, correctable errors; no leftover secrets by construction. **Effort:** M. **Builds on:** `add-data/manager/steps/connector.ts`, `connectors/ai/saveAiConnector.ts`, `ConnectorForm.svelte`, driver `Ping()`.

### ★ Unified object-store access-policy boundary
Extract one `authorizePath(op, bucket, path)` used by `ListObjects`, `ListObjectsForGlob`, `DownloadFiles`, and model `Delete`, with delimiter-boundary prefix matching. Today `path_prefixes` is only half-enforced: the non-glob list path checks it, but glob listing and downloads bypass it entirely (a `*` in the path escapes the restriction), and even the enforced path uses non-boundary `HasPrefix` so `data` matches `database/`. Model `Delete` has the same non-boundary flaw and can over-delete sibling prefixes.
- **Value:** Makes `path_prefixes` a trustworthy boundary — the prerequisite for safe self-serve/multi-tenant bucket browsing by the UI and AI assistant. **Effort:** M. **Builds on:** `runtime/drivers/{s3,gcs,azure}/object_store.go`, `object_store.go:177`, `s3/model_manager.go`.

### Per-bucket credential preflight + SSRF/misconfig guardrails
Extend `Ping` to validate credentials against a specific bucket/prefix and return actionable errors (region mismatch, missing `s3:ListBucket`, anonymous fallback engaged, `path_prefixes` mismatch), and add allow/deny-list guards to the HTTPS `FilePaths` fetch, which currently issues `http.DefaultClient.Do` against any user URL (internal metadata endpoints, RFC1918 hosts) with no restriction.
- **Value:** Users self-diagnose connector setup from clear errors instead of cryptic SDK failures; closes an SSRF vector. **Effort:** S. **Builds on:** `drivers/*/Ping`, `runtime/drivers/https/https.go:259`.

### Glob progress, object caps, and resumable downloads
Surface listing/download progress and add configurable max-objects/max-bytes caps for globs, feeding the download pipeline incrementally instead of materializing the full listing first (`blob.Download` calls `CollectAll` before the first byte). Persist the download manifest so a failed ingest resumes only the missing files.
- **Value:** Predictable behavior and no unbounded memory on `**` over huge buckets; cheap retries on large sources. **Effort:** M (caps/progress) / L (resumable). **Builds on:** `runtime/pkg/blob/download.go`.

### Cross-schema table pagination + uniform size/row-count surface
Make `AllFromInformationSchema` paginate *within* a schema instead of erroring past 1000 tables (large ClickHouse/Druid/Postgres schemas fail today with an opaque message), and surface a consistent `LoadPhysicalSize`-style estimate across all OLAP connectors (DuckDB/ClickHouse implement it; Druid/Pinot/MySQL/Postgres/StarRocks return -1, though Druid `sys.segments` and Pinot `/debug/tables` already exist).
- **Value:** Big schemas become usable; storage/scale insight everywhere instead of a blank -1. **Effort:** M. **Builds on:** `runtime/drivers/information_schema.go`, per-driver `EstimateSize`.

### NEW — `.env` secret manager view
A project-settings panel listing which `.env` keys each connector references, flagging orphaned/committed-but-unused secrets (like those left by a failed connector test) with one-click removal.
- **Value:** Secret hygiene becomes discoverable and self-serve. **Effort:** M. **Builds on:** `env-management`, connector YAML parsing.

---

## Metrics Layer & Modeling

### ★ Ratio / "per"-dimension derived measures
Wire up `addDerivedMeasureWithPer`, currently a stub that returns `support for "per" not implemented`, to support measures like "revenue per active user" natively. This is a commonly-requested BI primitive that today silently errors, forcing users to pre-compute ratios in the model.
- **Value:** Unlocks ratio-style measures — a staple of real dashboards — without model gymnastics. **Effort:** L. **Builds on:** `runtime/metricsview/ast.go:910`.

### ★ Structured query validation surface
Build the `ValidateQuery` stub (currently `panic("not implemented")`) into a real API that returns typed errors: unknown field, invalid grain, missing required dimension, cap exceeded. This lets the frontend and the AI assistant correct queries *before* executing, and would pre-empt the handful of panics reachable via the raw-`Query` resolver (malformed `cast` conditions, invalid time grains on duplicate time-floor dimensions).
- **Value:** Turns query errors into structured, self-correctable feedback — the missing contract the AI layer needs. **Effort:** M. **Builds on:** `runtime/metricsview/executor/executor.go:166`, `astexpr.go`, `ast.go`.

### Percent-of-total via window function (one query, no edge-case bugs)
Replace the separate totals round-trip query with `SUM(m) OVER ()` where the dialect supports it. This removes an extra OLAP query per request *and* sidesteps two live correctness bugs in one change: the bare-division/precision-truncated total (`%#f` gives divide-by-zero on a zero total) and the hard scan error when the total is NULL over an empty filtered set.
- **Value:** Fewer queries, and percent-of-total stops returning inf/NaN or erroring the whole dashboard. **Effort:** M. **Builds on:** `executor_rewrite_percent_of_totals.go`, `ast.go:537`.

### Native regex / "matches" filter operator across dialects
Expose a first-class regex/glob filter operator, properly escaped per dialect, so search and advanced filters behave identically across DuckDB/ClickHouse/Druid/Pinot. Today the Druid LIKE→regex conversion passes user metacharacters through unescaped (`"3.5"` matches `"3x5"`; `"(EU)"` errors the query), and the frontend `DimensionFilterMode` has Select/Contains/InList but no regex mode.
- **Value:** Power-user filtering that comma-lists can't express, consistent everywhere. **Effort:** M. **Builds on:** `astexpr.go:787`, `DimensionFilterMode`, dialect layer.

### Partition failure management + "why did this rebuild?" diagnostics
Surface errored/skipped partitions per model and let users re-run just the failed ones (the catalog already tracks per-partition `Error`/`ExecutedOn`, and `RefreshTrigger` already supports `AllErroredPartitions`), and record the trigger reason (spec change vs ref-hash vs manual vs missing-output) into model state. The latter would have made the rename→full-rebuild regression obvious in the field.
- **Value:** First-class recovery instead of full refresh; expensive rebuilds become explainable. **Effort:** M. **Builds on:** `runtime/reconcilers/model.go`, catalog partition tables, `refresh_trigger.go`.

### NEW — Semantic diff & downstream-impact preview for metrics views
When editing a metrics view (or model), show what changes downstream before save: measures/dimensions added or removed, which explores/canvases and alerts reference them, and which shared URLs would break. Pairs naturally with surfacing dropped fields on stale shared links (the URL-state loader already computes an `errors` array it currently discards).
- **Value:** Confidence when refactoring the semantic layer; fewer broken dashboards and links. **Effort:** M. **Builds on:** parser DAG/refs, `DashboardStateDataLoader`, resource graph.

---

## Dashboarding & Exploration

### ★ Deterministic dashboard state-sync engine
Replace the coarse boolean `updating` lock in `DashboardStateSync` with a small generation-token state machine that is race-safe and always releases (finally-based). One boolean today causes a cluster of bugs: it is never released on two early-return paths (permanently freezing URL/state sync until reload), drops competing navigations made during an in-flight time-range resolve, and double-initializes on a re-entrant init.
- **Value:** Fixes four related concurrency defects at the root; makes shared links and rapid interactions reliable. **Effort:** M. **Builds on:** `state-managers/loaders/DashboardStateSync.ts`.

### ★ Complete + tab-aware canvas export
Render tab-group content in the PDF/PNG export (today it renders only free rows, so every widget inside a tab group is silently dropped — a canvas that's entirely tabbed exports as a header-only PDF or fails outright), then let the export dialog choose "active tab only" vs "all tabs as sections," and add a post-export summary ("2 of 14 components could not be rendered") since `captureCanvasBlocks` already degrades per-block failures to warnings.
- **Value:** Exports stop silently losing data; teams that organize canvases into tabs get trustworthy PDFs. **Effort:** M. **Builds on:** `exports/pdf/CanvasPdfExportView.svelte`, `canvas-entity.ts` layout, `capture.ts`.

### Persisted, resizable column widths + shareable clean links
Store pivot/dimension-table column widths in explore state keyed by metrics view (they're component-local writables today, recomputed and lost on every navigation/reload), and add "Copy clean link" / "Reset to default view" buttons that emit the minimal cleaned (and if needed compressed) URL. Also fix export sort parity — dimension-table and pivot exports currently ignore the on-screen sort, so the downloaded file is ordered differently than what the user sees.
- **Value:** Wide tables feel stable; shared links are short and canonical; exports match the screen. **Effort:** M. **Builds on:** `NestedTable.svelte`, `url-state`, `dimension-table-export.ts`, `pivot-export.ts`.

### Promote pivot selection to a global filter + export just the selection
The pivot already tracks click-selected cells/columns/rows and has a self-filtering engine; add "apply this selection as a global filter" and "export only the selected region." Drill-down workflows currently require manually re-creating the same filter globally.
- **Value:** Streamlines the analyze-then-filter loop analysts do constantly. **Effort:** M. **Builds on:** `pivot-click-selection.ts`, `buildPivotFilter`.

### NEW — Human-readable rill-time expression editor
Surface the parsed `RillTime` (already available in time-controls) as an editable, validated time-range expression with inline error messages and named chips, reducing reliance on opaque preset names in the URL and giving power users direct control.
- **Value:** Transparent, precise time control; better shared-link legibility. **Effort:** M. **Builds on:** `time-controls`, `rilltime` grammar.

### State-diff debugger panel + cross-measure hover sync (dev/UX polish)
A dev-only overlay showing the cascading-merge inputs (session/URL/bookmark/YAML/Rill-default) and the winning source per key, plus cross-measure annotation hover sync so hovering an x-position on one time-series chart highlights it across all stacked measure charts.
- **Value:** Makes intricate state-origin bugs debuggable; easier visual correlation across measures. **Effort:** S. **Builds on:** `DashboardStateDataLoader`, `MeasureChart`.

---

## AI Assistant & Natural-Language Analytics

### ★ Sandboxed AI: read-only introspection role + project-root jail
A dedicated read-only execution path for `query_sql`/`list_tables`/`show_table` (separate connection or role) so introspection can't mutate data even under prompt injection, and a single project-root-jailed `projectPath(instanceID, userPath)` helper reused by every AI file tool. Today `read_file`/`write_file`/`develop_file` pass model-chosen paths straight to `filepath.Join(root, path)` with no containment check (so `../../.env` escapes the project), and `query_sql` runs arbitrary DDL/DELETE while advertising `ReadOnlyHint: true` (which MCP hosts may auto-approve).
- **Value:** Hardens the entire developer-AI surface against prompt injection and cross-tenant escape. **Effort:** M. **Builds on:** `runtime/ai/{file_read,file_write,develop_file,query_sql}.go`, `resolvers/sql.go`, repo drivers.

### ★ "Fix with AI" recovery, standardized everywhere
Promote a single recovery action set — Retry / Retry-without-AI / Edit YAML / Fix with AI — onto *every* failure surface: parse errors, failed metrics/dashboard generation, invalid shared-URL state, connector validation errors. Right now each surface handles failure differently (or dead-ends), and the powerful "fix with AI" affordance exists only in isolated spots.
- **Value:** No failure state is a dead-end; the AI becomes the universal escape hatch, reinforcing its value on every error. **Effort:** M. **Builds on:** `runtime/ai` develop/generate tools, `ImportDataStatus`, parse-error surfaces.

### ★ Scheduled AI narrative reports
Let a scheduled report deliver a natural-language "what changed this week" narrative (anomalies, top movers, notable segments) via the existing report/alert + notifier infra, instead of only a static dashboard snapshot. The AI resolver already runs the analyst agent for reports — but it panics today when `time_range`/`comparison_time_range` are omitted (the common case), so nearly every AI report fails; fixing that unlocks this.
- **Value:** A differentiated, high-retention collaboration surface — insight pushed to inboxes/Slack, not just links. **Effort:** M. **Builds on:** `runtime/resolvers/ai.go`, `reconcilers/report.go`, notifiers.

### Per-conversation cost meter + turn/session budget
Surface cumulative input/output/cached tokens (already recorded as metrics) back to the user in the chat UI, and enforce a per-turn/session token-and-call budget that halts gracefully. Today `MaxIterations` is per-loop only: router → developer (20) → `develop_file` (10) → `write_file`/`query_sql` can fan out to hundreds of LLM calls with only a wall-clock backstop, producing surprising spend on the billable managed connector.
- **Value:** Cost predictability for admins; the accounting hook for usage-based AI billing. **Effort:** M. **Builds on:** `runtime/ai/ai.go` completion loop, `InstanceConfig` query-limit knobs, provider token accounting.

### Approval gates + audit trail for destructive AI tools
For `write_file`/`develop_file`/`query_sql`, optionally require explicit user confirmation with a visible diff/statement preview before executing (the MCP annotations already carry `DestructiveHint`), and log every tool call.
- **Value:** Bounds prompt-injection blast radius on the highest-impact tools; auditability for teams. **Effort:** M. **Builds on:** tool annotations, chat UI, session persistence.

### Configurable AI agents + rolling-summary truncation + structured citations
Three compounding upgrades to the agent framework: allow registering additional agents (forecasting, anomaly) in the `ai` resolver instead of hard-rejecting anything but `analyst_agent`; replace the fixed keep-first-4/last-16 truncation window (a self-described placeholder that drops earlier decisions) with rolling summarization; and promote `query_metrics_view` `open_url` strings to first-class citation objects so rendering doesn't depend on prompt compliance.
- **Value:** Extensible AI use-cases, long sessions that retain context, reliable citations. **Effort:** L. **Builds on:** `ai.go:70`, `maybeTruncateMessages`, `query_metrics_view`.

---

## Collaboration & Sharing

### ★ Short share links via server-side state tokens
Persist large explore/canvas states server-side and reference them by a short id in the URL, bypassing the 16KB header limit and the client-side gzip/base64 compression path entirely. Combine with a universal "Copy link to this view" affordance across explore and canvas (canvas `lastVisitedState`/bookmarks already capture URL state, but keying is inconsistent and can bleed across projects).
- **Value:** Sharing "just works" — short, robust links regardless of dashboard complexity. **Effort:** XL (tokens) + S (share button). **Builds on:** `url-state`, `compression.ts`, canvas `lastVisitedState`, admin bookmark storage.

### ★ Scoped magic tokens + origin-pinned embeds
Let issuers pick an explicit least-privilege attribute profile when creating a magic/public link (never `admin`, choose which custom attributes carry), and add an `allowed_origin` embed param with a `connect` handshake so the iframe validates the parent origin and scopes all `postMessage` to it. Today a shareable magic link inherits the creating admin's `admin: true` attribute (subverting `{{ .user.admin }}` row/field security), and the embed RPC accepts messages from any origin and replies with `targetOrigin: "*"`.
- **Value:** Safe embeds and shared dashboards for enterprise — predictable data access, no cross-frame leakage. **Effort:** M. **Builds on:** `admin/server/magic_tokens.go`, `runtime_jwt.go`, `web-common/src/lib/rpc.ts`, embed layout.

### Public URL management surface + immediate revocation
Add expiry countdown, one-click revoke, and created-by/last-used columns to the public/magic-token table, guard issuance against past dates (a negative TTL is accepted today), and make revoke take effect instantly by evicting the token from the auth cache (revocation currently lags up to 10s because no revoke path clears the validation cache).
- **Value:** Sharing tokens become auditable and revocation becomes trustworthy. **Effort:** M. **Builds on:** `PublicURLsTable`, `CreatePublicURLForm`, `admin/auth_token.go`.

### Reliable delivery: idempotency keys + per-recipient isolation
Attach an execution-scoped idempotency key to every notification and give all notifiers per-recipient error isolation with bounded retry/backoff and a structured partial-success result. Today a report retry can re-deliver to notifiers that already succeeded (the `sent` flag is overwritten, not accumulated), and the Slack notifier aborts the whole recipient batch on the first error, so one bad address silently drops the rest and marks the alert failed.
- **Value:** No duplicate or silently-dropped alerts/reports — table stakes for a notification product. **Effort:** M. **Builds on:** `reconcilers/report.go`, `drivers/slack/notifier.go`, notifier interface.

### NEW — Dashboard comments & annotations
A lightweight comment/annotation primitive on explore and canvas dashboards (pin a note to a time range, a leaderboard value, or a widget), delivered via the same admin storage that backs bookmarks. Collaboration today is entirely out-of-band (paste a link into Slack); in-context discussion is a natural retention driver.
- **Value:** Turns dashboards into a collaboration surface, not just a viewer. **Effort:** L. **Builds on:** bookmark/admin storage, dashboard state, notifier fan-out.

---

## Performance & Scale

### ★ Streaming / chunked exports
Replace the single `q.Export`-to-response-writer (with a 1-hour token and a 16KB-ish header-encoded query) with a resumable, chunked export staged in object storage and delivered via signed URL. This removes the header-size ceiling and browser-timeout limits for large exports and eliminates a class of corrupt-download bugs where a mid-stream failure appends an error body to a committed 200 response.
- **Value:** Large exports become reliable and resumable; no more truncated/corrupt files. **Effort:** L. **Builds on:** `runtime/server/downloads.go`, resolver export path, object storage.

### ★ Security-aware cache keys + a correctness harness
Fold the full security context (`AdditionalRules`, `SkipChecks`, resolved row filter) into the query-result cache key, and add a test harness that, for every resolver and legacy query, asserts that varying each security dimension changes the key. Today the query cache hashes only `UserAttributes`, so two magic-token links to the same embedded dashboard with different locked filters (carried as `AdditionalRules`) can serve each other's rows — a default-on cross-context leak for managed metrics views.
- **Value:** Closes a real data-leak vector and prevents regressions as resolvers are added. **Effort:** S (harness) + M (key fix). **Builds on:** `runtime/resolver.go:210`, `runtime/security.go`, `metricsview/executor.CacheKey`.

### Bound the unbounded: capped in-memory results + streaming decode
Non-export resolve paths buffer the entire result set into `[]*structpb.Struct` before returning; add a capped/streamed decode with an explicit "result too large" error to harden toplist/rows/aggregation against pathological requests. Pair with clamping the AI `metrics_view_query` limit to `[1, AIMaxQueryLimit]` (a non-positive limit currently flows through to `LIMIT -1`/unbounded materialization on lenient dialects).
- **Value:** Removes OOM footguns on the hot query path; clients get a clear signal instead of a crash. **Effort:** M. **Builds on:** `rowsToData`, `runtime/ai/metrics_view_query.go`.

### Extend pre-aggregation routing (rollups + approx comparison)
Broaden rollup routing to non-primary time dimensions and validate measure-referenced dimensions, and add an approximation strategy for delta/ratio sorts (`rewriteApproxComparisonNode` bails on `sortDelta` today). Lets more queries hit pre-aggregated tables and lets comparison dashboards sort by change on high-cardinality dimensions without full scans.
- **Value:** Faster dashboards at scale without full-resolution scans. **Effort:** M–L. **Builds on:** `executor_rewrite_rollup.go`, `executor_rewrite_approx_comparisons.go`.

### Connection-cache & provisioner observability
Expose a debug endpoint for open handles per instance/driver (refs, age, hanging state — the conncache metrics already exist internally), emit slot-used/slot-total gauges per host with alerting at the existing thresholds, and add an orphaned-resource reconciler that reclaims provisioner resources with no live deployment. Also fixes several leaks (BigQuery client per ingest, ClickHouse embedded-server ref, un-unregisterable otelsql callbacks).
- **Value:** Operators can see and pre-empt pool churn, leaks, and slot exhaustion instead of discovering them via "database is locked" or provisioning failures. **Effort:** S–M. **Builds on:** `conncache`, `admin/provisioner/*`, `provisioner_resources`.

---

## Developer Experience

### ★ Reversible file operations: trash/undo + multi-select + conflict dialog
Soft-delete files to a recoverable `.trash` area (or an in-session undo toast), add shift/ctrl multi-select with bulk move/delete behind one confirmation, and a shared "replace / keep both / cancel" dialog reused by duplicate, upload, and rename. The file surface is scary today: rename/delete swallow errors so callers treat failures as success and navigate to broken paths, duplicate silently overwrites an existing `(copy)`, and every operation is one-at-a-time.
- **Value:** The core code-first editing surface stops being a place where one misclick loses work. **Effort:** M. **Builds on:** `entity-management/actions/actions.ts`, `file-explorer`, runtime file API.

### ★ `rill deploy --dry-run` + named profiles + CI login
Print the resolved deploy plan (target org/project, deploy mode, git remote/subpath, and the exact `.env` keys to be pushed) without mutating anything; let `~/.rill` hold multiple named credential+org profiles (`rill --profile work deploy`); and add a first-class non-interactive `rill login --token`/stdin for CI. Deploy today silently pushes every `.env` variable with no preview, and a name-collision retry can push local secrets to the *wrong* pre-existing project.
- **Value:** Confidence before destructive/first deploys, multi-tenant workflows, and clean CI onboarding. **Effort:** M. **Builds on:** `cli/cmd/{deploy,project}`, `cli/pkg/local/server.go`, `dotrill`.

### Parser as a linter: deprecation warnings + schema validation + panic isolation
Turn the parser into a proper authoring linter: emit warnings for the many silently-accepted deprecated aliases (`title→display_name`, `dashboards→metrics_views`, legacy email blocks…), run each resource's YAML through its already-embedded JSON Schema during parse for uniform structural errors and typo detection, and wrap each `parseNode` in a `recover()` so one bad file localizes to that file instead of failing the whole project.
- **Value:** Smoother upgrades, consistent error messages, and a single malformed file can't take down the project. **Effort:** S (warnings + recover) / M (schema). **Builds on:** `runtime/parser` (`schema.go`, `parseNode`, alias handling).

### "Verify my project" linting + migration dry-run
One command that reports parse-time DAG problems (obvious ref cycles, self-refs, dangling refs — currently only caught later in the reconciler), unused resources, and deprecated syntax; plus a migration dry-run/plan that reports which versions will execute (from→to) and warns on downgrades/re-runs before applying (the migration reconciler re-executes the current top version on every reconcile today).
- **Value:** Catch DAG and migration problems at edit time, not after reconcile. **Effort:** M. **Builds on:** parser refs, `reconcilers/migration.go`.

### OS keychain credentials + deterministic local port
Store the access token (and per-env/assume-user tokens) in the OS keyring with a plaintext-YAML fallback, and on `rill start` fail fast or auto-select the next free port with a printed URL instead of racing a 250ms health-check that can open the browser onto a *different* Rill already bound to the port. Credentials are world-readable (`0644`) in `~/.rill` today.
- **Value:** Matches dev-CLI expectations for credential safety; no "opened the wrong app" confusion. **Effort:** M (keychain) / S (port). **Builds on:** `dotrill.go`, `PollServer` in `cli/pkg/local/app.go`.

---

## Governance, Security & Multi-tenancy

### ★ One SecureJoin helper to kill the path-traversal class
A single containment-checked `secureJoin(root, userPath)` — cleans the result and requires it to stay within `root`, backed by tests for `..`, absolute inputs, symlinks, and sibling-prefix cases — reused by every path resolver: the AI file tools, both `RepoStore` drivers, the blob downloader, and `fileutil.ResolveLocalPath`. Several independent HIGH findings are the same missing check: `filepath.Join(root, path)` with no boundary, letting `EditRepo` users or a steered AI read/write/delete outside the project (cross-tenant on shared runtimes).
- **Value:** Eliminates a whole class of tenant-isolation bugs in one place instead of patching each call site. **Effort:** S. **Builds on:** `runtime/ai/file_*.go`, `drivers/{file,admin}/repo.go`, `pkg/blob/download.go`, `pkg/fileutil`.

### ★ Per-service authz test harness
A table-driven test that iterates every RPC and asserts a minimally-privileged or anonymous caller is rejected. This would have caught the entire `ConnectorService` (`ListBuckets`/`OLAPListTables`/`GetTable`/`ListObjects`) shipping with *no* authorization — a viewer or even an anonymous caller can enumerate every physical table, schema, and bucket, and some endpoints don't even scope to the requested `instance_id`.
- **Value:** Closes a live cross-tenant enumeration gap and prevents future regressions on a security-critical surface where interceptors only populate claims and never enforce. **Effort:** S. **Builds on:** `runtime/server/connector_service.go`, auth interceptors.

### Field-level access consistency pass
Make field-level security uniform: add the missing `CanAccess()` gate to the `metrics_annotations` resolver (the one metrics-view-backed resolver that skips it, leaking annotation rows to users denied the view), and decide/document whether the primary time dimension should honor `CanAccessField` (it bypasses field checks today). Fold this into the cache-key fix so cached results can't cross security contexts.
- **Value:** The security policy a user configures is actually enforced on every path. **Effort:** S–M. **Builds on:** `resolvers/annotations.go`, `metricsview/ast.go`, resolver cache key.

### Superuser audit log + immediate revocation + redirect allow-list
Persist every Superuser / force-access / assume-user action (who, whom, when, why) with an admin UI; wire token/session revocation to evict the auth cache so "log out everywhere" is instant; and route every redirect (auth start, logout, device code, PKCE, billing callback) through one validated allow-list helper (open redirects exist in the billing callback, local PKCE handlers, and device-auth today).
- **Value:** Compliance and incident-response for a support-heavy control plane; systematic open-redirect prevention. **Effort:** M. **Builds on:** `admin/server/{sudo,auth}`, `auth_token.go`, redirect validation.

### NEW — Data-access audit log + policy tester
A queryable audit log of who queried what through which token/JWT, and a policy tester (extending "View As") that simulates a metrics-view security policy against arbitrary user attributes and shows the resolved row/field filters — so policy authors can verify a restricted role *before* sharing. The current "View As" can silently show admin permissions during load and stale data across user switches, undermining exactly this use case.
- **Value:** Enterprise-grade governance and confidence that policies do what authors intend. **Effort:** L (audit) / M (tester). **Builds on:** runtime security engine, `view-as-user`, JWT minting.

---

## Monetization & Growth Loops

### ★ Post-deploy → share growth loop
After a successful deploy (local `rill start` detecting a working dashboard, or the Cloud deploy flow), present an explicit share step: copy public URL, invite teammates, open in Cloud, plus a clickable `rill deploy` nudge in the terminal for local users. Today the funnel ends at deploy; the share/invite/viral step is missing entirely.
- **Value:** Turns the terminal endpoint of the funnel into a viral loop — the single highest-leverage growth move here. **Effort:** S. **Builds on:** deploy flow, public-URL issuance, CLI output.

### ★ Per-session AI cost budget + usage-based metering
Meter cumulative input/output/cached tokens per conversation (already recorded internally), enforce a configurable per-session budget, and surface it — the foundation for usage-based AI pricing on the billable managed connector. Pairs directly with the AI cost-meter idea above.
- **Value:** Cost guardrails for admins today; a metered AI revenue line tomorrow. **Effort:** M. **Builds on:** `runtime/ai` token accounting, `InstanceConfig`, billing reporter.

### ★ Customer-facing usage & consumption dashboard
A self-serve view of an org's own usage — storage per connector, query volume, AI spend, seats, credits — built on the existing `project_storage` resolver and billing-reporter metrics (parallelize the currently-sequential per-connector size estimation so one slow connector doesn't stall it). Ties naturally into in-product upgrade prompts at quota/usage thresholds.
- **Value:** Transparency drives trust and self-serve expansion; the natural home for upgrade CTAs. **Effort:** M. **Builds on:** `resolvers/project_storage.go`, `billing_reporter.go`, quotas.

### Billing-issue admin console + idempotent reconcile
An internal view to list/inspect/resolve per-org billing issues (trial, payment-failed invoices, credit state) and manually re-run repair, backed by treating Orb as the source of truth and reconciling subscription/plan/quotas/issues in one convergent operation. Replaces several overlapping Init/Repair/Start paths that today produce duplicate emails, abandoned hibernation on retry, and >100-project pagination gaps.
- **Value:** Operators can see and fix stuck billing state instead of it living only in DB rows and daily jobs. **Effort:** M (console) / L (idempotent reconcile). **Builds on:** `admin/billing`, `admin/jobs/river`.

### NEW — Embed analytics & per-embed telemetry
Emit per-embed usage telemetry (which embedded dashboards are viewed, by which parent origin, how often), enabled by the embed origin-handshake work. Gives sales/CS an expansion signal and customers a usage view of their embedded analytics.
- **Value:** Turns embeds from a feature into a measurable growth/expansion surface. **Effort:** M. **Builds on:** embed RPC handshake, `EmbedStore`, telemetry.

### Provisioner capacity metrics + safeguarded autoscaler
Emit slot-used/slot-total and cluster-health gauges with alerting, and finish/re-enable the disabled autoscaler behind the real `AutoscalerCron` flag with per-project min/max policies, dry-run logging, and nil-guards. Also fix the `clickhouse-static` name truncation that drops the unique resource ID (a real data-loss/collision risk) and the static-slot TOCTOU that overcommits hosts.
- **Value:** Right-sized compute (cost) and headroom visibility before overcommit/exhaustion. **Effort:** M (metrics) / L (autoscaler). **Builds on:** `admin/provisioner/{static,clickhousestatic}`, `run_autoscaler.go`.

---

## Quick Wins
High value, low effort (S) — pick these up first:

- **SecureJoin path helper** — one small helper closes multiple HIGH path-traversal findings across AI tools, repo drivers, blob, and fileutil. *(Governance)*
- **Per-service authz test harness** — a table-driven test that would have caught the unauthenticated `ConnectorService`. *(Governance)*
- **Post-deploy → share step + CLI `rill deploy` nudge** — the cheapest viral-loop move. *(Growth)*
- **Cache-key correctness harness** — asserts security dimensions change the key; guards the cross-context leak. *(Perf/Governance)*
- **Parser deprecation warnings + per-node `recover()`** — smoother upgrades and blast-radius isolation. *(DevEx)*
- **Fix `emitStartEvent` + add funnel events** — unskews activation analytics. *(PLG)*
- **`metrics_annotations` `CanAccess()` gate** — one missing check closes a policy-bypass. *(Governance)*
- **Field-consistent secret redaction + real AI `Ping`** — no leaked keys via `Config()`; "test connection" tells the truth. *(Connectors)*
- **Deterministic local port + `rill deploy --dry-run`** — remove two common footguns. *(DevEx)*
- **Immediate token revocation (cache eviction) + redirect allow-list helper** — make revoke instant and close open redirects. *(Governance)*

## Bigger Bets
High-leverage, larger investments (L/XL) — the differentiated platform moves:

- **One-click aha (seeded datasets + offline sample generator)** — fixes the worst first-run dead-end; the fastest path to a real dashboard. *(PLG)*
- **Structured query validation surface** — the typed contract that lets the frontend *and* the AI self-correct queries before executing. *(Metrics/AI)*
- **Deterministic dashboard state-sync engine** — retires a whole cluster of URL/state concurrency bugs at the root. *(Dashboarding)*
- **Scheduled AI narrative reports + "Fix with AI" everywhere** — push insight to inboxes and make the AI the universal recovery path. *(AI/Collaboration)*
- **Sandboxed AI (read-only role + project-root jail)** — safe to give the developer AI real power. *(AI/Governance)*
- **Short share links via server-side state tokens** — sharing that "just works" regardless of dashboard complexity. *(Collaboration)*
- **Scoped magic tokens + origin-pinned embeds** — the trust foundation for enterprise embedding. *(Collaboration/Governance)*
- **Streaming/chunked exports** — reliable, resumable large exports; removes header/timeout ceilings. *(Perf)*
- **Ratio / "per" derived measures** — a first-class BI primitive that today silently errors. *(Metrics)*
- **Usage/consumption dashboard + per-session AI metering** — transparency plus the metered-AI revenue line. *(Monetization)*
- **Reversible file operations (trash/undo/multi-select)** — makes the code-first editing surface safe to use. *(DevEx)*
