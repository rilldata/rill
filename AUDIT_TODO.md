# Rill Code Audit — Pending Findings

**Last verified:** 2026-09-20, item by item cross-referenced with changed files, against `origin/main` merged into `review_code` (HEAD `008315ff3`).
**Consolidates:** the `audit/` area reports, `audit/MUST-FIX.md`, and `REVIEW_FINDINGS_TODO.md` (both superseded by this file). Full failure scenarios and test guidance for MF items remain in the linked `audit/*.md` reports.
**Now/next/later work queue:** derived from this file in [`audit/todo/`](audit/todo/README.md).

**107 findings pending** — **39 must-fix** (stable MF ids) · **62 medium** (1 partial) · **6 low**.

### 2026-09-20 re-verification result

- 53 PRs merged 2026-08-30 → 2026-09-20. Changed files were intersected with every finding's cited paths (31 cited files touched); each overlapping finding was re-checked at its cited location. **3 findings fixed, 1 partially mitigated:**
  - **MF-33 FIXED** — canvas `timeAndFilterStore` inner-subscription leak: the unified expression filter refactor ([#9746](https://github.com/rilldata/rill/pull/9746)) rewrote it as a single flat `derived([...])` with no inner `subscribe(set)` calls (`BaseCanvasComponent.ts:190+`).
  - **G-4 FIXED** — expression mapping no longer drops falsy literals: `if (expr.val !== undefined)` at `map-metrics-resolver-query-to-dashboard.ts:139-140`.
  - **G-5 FIXED (obsolete)** — the `dimensionsWithInListFilter` marker mechanism no longer exists anywhere in `web-common` (removed by the unified expression filter refactor #9746/#9886); the cited flaw cannot occur.
  - **MF-38 PARTIAL** — the new `ExpressionFilterManager.setUrlParams` wraps `expandCompressedParams` in try/catch, but the main `convertURLToExplorePreset.ts:92` call is still unguarded and the `String.fromCharCode(...spread)` overflow in `compression.ts:23` remains. Still pending.
- Everything else re-checked at touched locations survives verbatim: MF-07/08/09/11/13/14/20/21/23/32/34/37/41, A-7/8/9/10, B-1/8/10/11, C-2, F-3, G-1/2, L-1/2/3/4. Notably `MetricsViewAnnotations` still assigns `limit = &req.Offset` (B-10) despite the file being touched.
- Cited line drift: MF-34 is now `layout-util.ts:186-325` (`generateArrayRearrangeFunction`; `insertRow` splices `newArray` at `:220,258,285` without splicing `touchedRows`); MF-32 is now `rduckdb/db.go:953`; A-9 is now `organizations.go:1110-1140`; L-1 is now `controller.go:1220`; MF-21 send is now `controller.go:1386`.
- Related but not findings: `feat: map` ([#8931](https://github.com/rilldata/rill/pull/8931)) landed native map charts, and `skills/*.md` AI agent skills shipped ([#9851](https://github.com/rilldata/rill/pull/9851)) — both items from the competitive roadmaps in `audit/*/roadmap.md`.

### 2026-08-30 re-verification result

- Every previously pending finding was re-checked against current code (42 PRs merged 2026-07-31 → 2026-08-30). **Zero pending findings were fixed** in that window; several cited files were refactored (SSE refactor #9261/#9260, pivot orchestration #9538, URI links #9588, view-as invalidation #9804, catalog-flush race #9815) but the specific failure modes survived in every case. Line numbers below are current.
- Removed as fixed: “runtime repo file APIs allow paths that escape the project root” (fixed by [#9723](https://github.com/rilldata/rill/pull/9723) via `drivers.ResolveRepoPath`, verified across all file/admin repo driver methods) — it was still listed as pending in `REVIEW_FINDINGS_TODO.md`.
- Reclassified: credit-trial start (A-7) is now **partial** — a `balance <= 0` guard makes retries idempotent, but concurrent starts can still double-grant.
- The GitHub-issue triage reports (`OPEN_ISSUES_REPORT.md`, `codex_issues_report.md`) were also re-checked on 2026-08-30: every tracked backlog issue is still open on GitHub, so both reports remain current as written.

---

## Must-fix (39)

MF ids are stable and never renumbered. Grouped by theme; each item links its source report.

### Security boundaries

- [ ] **MF-01 — Unsafe web-admin redirects after billing/device flows.** Untrusted `redirect` param opened verbatim, even after a swallowed mutation error; device flow also interpolates un-encoded `user_code` — `web-admin/src/routes/[organization]/-/upgrade-callback/+page.svelte:97-103`, `web-admin/src/features/billing/plans/upgrade-to-plan.ts:58`, `web-admin/src/routes/-/auth/device/+page.svelte:24,32`. ([web-admin](audit/web-admin.md))
- [ ] **MF-13 — Magic tokens replay the creator's `admin` flag and privileged attributes.** Shared/embed links inherit `admin:true`, defeating `{{ .user.admin }}` policies — `admin/server/magic_tokens.go:90-99`, `admin/server/server.go:385`, `admin/server/runtime_jwt.go:113-119`. ([admin-auth-rbac](audit/admin-auth-rbac.md))
- [ ] **MF-15 — Object-store `path_prefixes` bypassed by glob listing and downloads.** Only `ListObjects` enforces prefixes; `ListObjectsForGlob`/`DownloadFiles` go straight to the bucket — `runtime/drivers/s3/object_store.go:74-106`, `gcs/object_store.go:81-94`, `azure/object_store.go:87-98`. ([drivers-object-store](audit/drivers-object-store.md))
- [ ] **MF-16 — Object-store prefix checks are not boundary-aware.** Raw `strings.HasPrefix`, so allowing `data` also admits `database/` — `runtime/drivers/object_store.go:177-180,276`. ([drivers-object-store](audit/drivers-object-store.md))
- [ ] **MF-17 — AI `query_sql` executes mutations despite `ReadOnlyHint: true`.** Arbitrary SQL passed to the generic resolver with only row limit/timeout — `runtime/ai/query_sql.go:41,91-97`. ([ai](audit/ai.md))
- [ ] **MF-18 — Local credentials written world-readable.** `os.WriteFile(..., 0o644)` and `MkdirAll(dotrill, os.ModePerm)` — `cli/pkg/dotrill/dotrill.go:113,362`. ([cli](audit/cli.md))
- [ ] **MF-19 — Local PKCE auth/logout handlers accept external redirects.** Raw `?redirect=` used as post-auth origin and logout target — `cli/pkg/local/server.go:815-827,878,897-901`. ([cli](audit/cli.md))
- [ ] **MF-30 — ClickHouse object-store credentials leak into SQL traces/logs.** S3/GCS keys embedded verbatim in generated `s3(...)`/`gcs(...)` SQL executed through the otelsql-instrumented connection — `runtime/drivers/clickhouse/model_executor_objectstore_self.go:127-161`. ([deep-connectors](audit/deep-connectors.md))
- [ ] **MF-39 — TDD row headers render dimension values through `innerHTML` (stored XSS).** `escapeHtml` from #9588 covers only link `href`/`title`, not the value itself — `web-common/src/features/dashboards/time-dimension-details/TDDTable.svelte:210-214,262`. ([deep-audit-time-dimension-detail](audit/deep-audit-time-dimension-detail.md))
- [ ] **MF-40 — Data-controlled `javascript:`/`data:` URIs reach dashboard links.** `makeHref` has no safe-scheme allowlist; #9588 widened exposure to leaderboard and pivot links — `web-common/src/features/dashboards/dashboard-utils.ts:155-180`. ([deep-audit-time-dimension-detail](audit/deep-audit-time-dimension-detail.md))

### Crashes and panics on ordinary input

- [ ] **MF-04 — Pinot schema enumeration panics on a transient table-schema failure.** Per-table lookup errors swallowed, leaving nil slots that `sort.Slice` dereferences — `runtime/drivers/pinot/information_schema.go:274-287`. ([drivers-olap](audit/drivers-olap.md))
- [ ] **MF-05 — `logbuffer.GetLogs(limit=0)` panics.** Only `limit < 0` is remapped; `0` indexes `logs[-1]` — `runtime/pkg/logbuffer/buffer.go:77-93`. ([runtime-pkg](audit/runtime-pkg.md))
- [ ] **MF-38 — Malformed or oversized `gzipped_state` crashes dashboard load (PARTIAL).** `ExpressionFilterManager.setUrlParams` now guards decompression, but `convertURLToExplorePreset.ts:92` calls `expandCompressedParams` unguarded and the `String.fromCharCode(...bytes)` spread overflow remains — `web-common/src/features/dashboards/url-state/compression.ts:20-23,36`. ([deep-audit-state-management](audit/deep-audit-state-management.md))
- [ ] **MF-28 — rduckdb `deleteRemote` panics on listing errors and deletes the wrong `meta.json`.** Deletes root-level `meta.json` instead of `<table>/meta.json`; non-EOF iterator error falls through to nil deref — `runtime/pkg/rduckdb/remote.go:276,287-294`. ([deep-connectors](audit/deep-connectors.md))

### Deadlocks, hangs and races

- [ ] **MF-06 — A slow log subscriber deadlocks instance logging.** `AddEntry` does unbuffered sends under `b.mu`; `WatchLogs` fan-out is a synchronous blocking `srv.Send` — `runtime/pkg/logbuffer/buffer.go:38-49`, `runtime/server/instances.go:274-276`. ([runtime-pkg](audit/runtime-pkg.md))
- [ ] **MF-07 — `DeleteInstance` blocks forever after registry-store delete fails.** `registryCache.delete` returns a nil channel on error and the caller unconditionally receives — `runtime/registry.go:114-120,398-401`. ([runtime-core](audit/runtime-core.md))
- [ ] **MF-08 — AI session notification strands the producer under the session lock.** `Subscribe` is unbuffered; `AddMessage` sends while holding `s.mu`; consumer exits on `ctx.Done()` — `runtime/ai/ai.go:653,889-895`, `runtime/server/chat.go:385-388`. ([ai](audit/ai.md))
- [ ] **MF-09 — Static runtime slot allocation is a TOCTOU race (overcommit).** Capacity read and unconditional upsert are separate plain statements — `admin/provisioner/static/static.go:94-121`, `admin/database/postgres/postgres.go:794-810`. ([admin-billing-provision](audit/admin-billing-provision.md))
- [ ] **MF-21 — Late reconcile completions block forever after controller shutdown.** Plain `c.completed <- inv` on an unbuffered channel; shutdown drain gives up after 30s (#9815 did not touch this) — `runtime/controller.go:148,325-350,1335`. ([REVIEW_FINDINGS](audit/00-summary.md))
- [ ] **MF-22 — DuckDB connection-count metrics data race.** Callback reads `dbConnCount` without `dbCond.L` — `runtime/drivers/duckdb/duckdb.go:213-214,718-748`. ([REVIEW_FINDINGS](audit/00-summary.md))

### Data loss and silent failure

- [ ] **MF-10 — File rename/delete helpers swallow errors; callers navigate as if success.** — `web-common/src/features/entity-management/actions/actions.ts:227-231,271-275`, callers `RenameAssetModal.svelte:106-118`, `FileExplorer.svelte:155-167,192-196`. ([frontend-shared](audit/frontend-shared.md))
- [ ] **MF-11 — Failed chat send loses the draft and keeps a false optimistic message.** `startStreaming` resolves on "close" and never rejects; rollback in `handleTransportError` unreachable — `web-common/src/features/chat/core/conversation.ts:377-388,436-451,705-723`. ([frontend-runtime-client](audit/frontend-runtime-client.md))
- [ ] **MF-14 — Editable deployment deletion continues after checkpoint push failure.** `GitPush` failure logged at Warn, then deprovision/branch delete/DB delete proceed — `admin/deployments.go:414-422,465-476,480`. ([admin-projects-deploy](audit/admin-projects-deploy.md))
- [ ] **MF-20 — Failed delete reconciliation still hard-deletes the catalog resource.** Error only logged before `c.catalog.delete(...)` — `runtime/controller.go:1457-1467`. ([REVIEW_FINDINGS](audit/00-summary.md))
- [ ] **MF-25 — Exhausted deployment reconcile jobs never persist a terminal error state.** River error handler only logs at max attempts; deployments stay “Provisioning…” forever — `admin/jobs/river/river.go:873-880` (only `DeploymentStatusErrored` write is the enqueue-failure path at `admin/deployments.go:817`). ([REVIEW_FINDINGS](audit/00-summary.md))
- [ ] **MF-29 — rduckdb deletes the old remote version even when new metadata upload fails.** Failed `meta.json` upload only logged; old version deleted before `return err` — `runtime/pkg/rduckdb/remote.go:244-255`. ([deep-connectors](audit/deep-connectors.md))
- [ ] **MF-31 — Salesforce batch-result fetch errors become empty files.** Errors logged and replaced with `bytes.NewReader(nil)`; ingestion proceeds silently partial — `runtime/drivers/salesforce/bulk.go:211-231`. ([deep-connectors](audit/deep-connectors.md))

### Correctness, leaks and wrong results

- [ ] **MF-12 — Redshift `UNLOAD` breaks on any query containing a single quote.** Query interpolated unescaped — `runtime/drivers/redshift/warehouse.go:120`. ([drivers-object-store](audit/drivers-object-store.md))
- [ ] **MF-23 — ClickHouse partial-open failures leak pools/embedded process.** After `embed.start()` and read-handle open, later failures return without cleanup — `runtime/drivers/clickhouse/clickhouse.go:224-395` (esp. `:303,308,322,336-339`). ([REVIEW_FINDINGS](audit/00-summary.md))
- [ ] **MF-24 — Runtime-client JWT can become `Bearer undefined` after waiting for refresh.** Token checked before the await, header written after — `web-common/src/runtime-client/v2/runtime-client.ts:64-67`. ([REVIEW_FINDINGS](audit/00-summary.md))
- [ ] **MF-26 — Migration reconciliation re-executes the current version.** `for v := from; v <= to; v++` with `from = state.Version` re-runs the persisted version every reconcile — `runtime/reconcilers/migration.go:76-83`. ([REVIEW_FINDINGS](audit/00-summary.md))
- [ ] **MF-27 — Embed singleton retains the previous token/host/instance/external user.** `EmbedStore.init(url)` only when no singleton exists; store is immutable with no reset (#9761 didn't change init) — `web-admin/src/routes/-/embed/+layout.ts:7-9`, `web-common/src/features/embeds/embed-store.ts:37-49`. ([REVIEW_FINDINGS](audit/00-summary.md))
- [ ] **MF-32 — rduckdb registers a non-removable DB-stats callback per write-handle open.** `otelsql.RegisterDBStatsMetrics` on every model build with no unregister — `runtime/pkg/rduckdb/db.go:956` (via `acquireWriteConn:780`). ([deep-connectors](audit/deep-connectors.md))
- [ ] **MF-34 — Canvas row insertion marks the wrong rows touched.** `touchedRows` never spliced when `insertRow` splices `newArray`; flags desync from indices after upward moves — `web-common/src/features/canvas/layout-util.ts:196-318`. ([deep-audit-canvas](audit/deep-audit-canvas.md))
- [ ] **MF-35 — Pivot row-total merging can assign cells to the wrong dimension rows.** Skeleton built from filtered `mergedRowTotals` while cells indexed against unfiltered `rowDimensionValues` — `web-common/src/features/dashboards/pivot/pivot-table-transformations.ts:66-86,222-252`, `pivot-data-store.ts:403-416,445-455`. ([deep-audit-pivot](audit/deep-audit-pivot.md))
- [ ] **MF-36 — Pivot expanded-table cache discards new pages and fresh data.** Cached `expandedTableMap[configKey]` returned unconditionally (moved by #9538, flaw intact) — `web-common/src/features/dashboards/pivot/pivot-data-assembly.ts:90-96`. ([deep-audit-pivot](audit/deep-audit-pivot.md))
- [ ] **MF-37 — Module-level pivot `lastKey` breaks pagination across instances.** Shared `let lastKey` across all pivot instances — `web-common/src/features/dashboards/pivot/pivot-data-config.ts:25,164-166`, `web-common/src/features/canvas/components/pivot/util.ts:98,239-241,329-331`. ([deep-audit-pivot](audit/deep-audit-pivot.md))
- [ ] **MF-41 — TDD global scroll position is never reset or scoped.** Module-level `lastKnownPosition` restored with no per-dashboard scoping — `web-common/src/features/dashboards/time-dimension-details/time-dimension-data-store.ts:469`, `TDDTable.svelte:403-408`. ([deep-audit-time-dimension-detail](audit/deep-audit-time-dimension-detail.md))
- [ ] **MF-42 — Metrics SQL accepts unsupported operators into invalid IR.** `operator` default case passes unknown opcodes through as `metricsview.Operator(op.String())`; fails late with confusing errors (#9716/#9721 didn't touch this) — `runtime/metricsview/metricssql/parser.go:509-512`. ([deep-metrics-sql](audit/deep-metrics-sql.md))

---

## Medium (62)

### A. Admin, provisioning, billing and GitHub (11)

- [ ] **A-1 `CreateProject` failure paths leave partial state.** Project tx commits before deploy-input validation; deployment creation/primary update not compensated — `admin/projects.go:24-115`. Fix: validate before commit; add compensating cleanup or an idempotent outbox flow.
- [ ] **A-2 `used_flusher` drops IDs after DB update failures.** Maps swapped before `flushToDB`, whose errors are only logged — `admin/used_flusher.go` (`flush`/`flushToDB`). Fix: requeue failed IDs or clear batches only after successful update.
- [ ] **A-3 `used_flusher` shutdown races the background goroutine.** `Close` does `cancel(); flush(); flushWg.Wait()` while the ticker flush can still `flushWg.Add(1)`; goroutine never joined — `admin/used_flusher.go` (`Close`/`runBackground`). Fix: join the goroutine and stop it before the final flush.
- [ ] **A-4 Kubernetes readiness ping ignores context and HTTP status.** Final `retryClient.Get(pingURL)` without ctx; `resp.StatusCode` never checked — `admin/provisioner/kubernetes/kubernetes.go:424`. Fix: `NewRequestWithContext`, bounded timeout, require 2xx.
- [ ] **A-5 Kubernetes provisioner setup panics on bad config.** Nil `TemplatePaths` deref, `template.Must`, `strings.Split(host, "//")[1]` — `admin/provisioner/kubernetes/kubernetes.go:125-133,206`. Fix: validate spec fields; parse host with `url.Parse`; return wrapped errors.
- [ ] **A-6 Static ClickHouse provisioning does not roll back DDL after later failure.** User/grant errors leave orphan DB/user; `Deprovision` exits early when `cfg.DSN == ""` — `admin/provisioner/clickhousestatic/provisioner.go:110+,264-266`. Fix: best-effort rollback or persist cleanup identifiers before DDL.
- [ ] **A-7 Credit-trial start is not serialized (PARTIAL).** A `balance <= 0` guard makes retries idempotent, but there is no per-org lock/idempotency key, so concurrent starts across the three call sites can double-grant — `admin/billing.go:203,283-291`, `admin/server/billing.go:149`, `admin/jobs/river/org_jobs.go:150`. Fix: per-org advisory lock or idempotency key around trial start.
- [ ] **A-8 Project-wide deployment updates have a data race on `prodErr`.** Goroutines assign the shared variable without synchronization — `admin/deployments.go:148-175`. Fix: mutex or per-deployment results through the errgroup.
- [ ] **A-9 Deployment annotations not refreshed when org custom domain changes.** `req.CustomDomain != org.CustomDomain` compared after `org` is reassigned with the new value, so the refresh branch never runs — `admin/server/organizations.go:1046-1079`. Fix: capture the old domain before update.
- [ ] **A-10 GitHub webhook processing is synchronous and unbounded.** `s.admin.ProcessGithubEvent(context.Background(), event)` inline in the handler — `admin/server/github.go:926`. Fix: enqueue a durable job after signature validation; bounded worker contexts; idempotent event keys.
- [ ] **A-11 Managed GitHub repos orphaned when DB metadata insert fails.** External repo created, then `InsertManagedGitRepo` failure returns with no compensation; cleanup job only sees DB-tracked repos — `admin/github.go:263-274`. Fix: best-effort delete on insert failure, or persist a pending row first.

### B. Runtime controller, reconcilers and server (11)

- [ ] **B-1 Catalog store release skipped on `newCatalogCache` init failures.** `NextControllerVersion`/`FindResources` errors return without `release()` — `runtime/catalog_cache.go:52-55,73-76`. Fix: release on all post-acquire error paths.
- [ ] **B-2 Controller accepts mutations after shutdown starts.** `c.closed` set only after the 30s drain + reconciler close; `checkRunning` is the sole guard — `runtime/controller.go:364,855`. Fix: mark stopping before leaving the event loop or reject mutations at shutdown start.
- [ ] **B-3 Staged model cleanup uses the canceled context.** `stageResultManager.Delete(ctx, stageResult)` deferred on the same cancellable ctx; leak accepted by comment — `runtime/reconcilers/model.go:1505-1513`. Fix: fresh bounded background context for cleanup.
- [ ] **B-4 Migration templates panic on non-model lookups.** Unconditional `res.Resource.(*runtimev1.Resource_Model)` type assert in template `Lookup` — `runtime/reconcilers/migration.go:149-150`. Fix: check `res.GetModel()` and return a template error.
- [ ] **B-5 Metrics-view reconciliation treats any model lookup error as “external table”.** Branches on `err == nil` and falls back to `Table = Model` for transient/context errors too — `runtime/reconcilers/metrics_view.go:81-88,94-98`. Fix: fall back only on `drivers.ErrResourceNotFound`.
- [ ] **B-6 Refresh-trigger processing is not durable before direct reconciles.** `r.C.Reconcile(ctx, rn)` fired for generic resources with progress kept only in the in-memory cloned spec; retry re-fires them — `runtime/reconcilers/refresh_trigger.go:87-94`. Fix: persist progress or make retries idempotent.
- [ ] **B-7 Alert ref-hash lookup failures silently change the hash.** All `c.Get` errors (incl. cancellation) collapse to `stateVersion = -1` — `runtime/reconcilers/alert.go:445-451`. Fix: return non-not-found errors from `refsStateHash`.
- [ ] **B-8 Report notification dirty state overwritten across notifiers.** Single `sent` flag per iteration; earlier successful email forgotten when a later notifier fails → duplicate send on retry — `runtime/reconcilers/report.go:619-635`. Fix: accumulate `anySent`.
- [ ] **B-9 Runtime SSE adapter sends are not cancellation-aware.** Plain `events <- event` while `serveSSEUntilClose` can return early on write error, stranding producers — `runtime/server/sse.go:137,171,218,265`, `runtime/server/chat.go:525,545`. Fix: select on `ctx.Done()` for every send.
- [ ] **B-10 `MetricsViewAnnotations` passes request `offset` as `limit`.** `if req.Offset != 0 { limit = &req.Offset }`, `offset` stays nil — `runtime/server/queries_metrics.go:630-633`. Fix: assign `offset = &req.Offset`.
- [ ] **B-11 Unknown time-grain enum values panic instead of validating.** `TimeGrainFromProto` default case panics; called on client-supplied enums — `runtime/metricsview/query.go:397-398` (call sites `queries_metrics.go:641,652`). Fix: return `(TimeGrain, error)` or validate at API boundaries.

### C. Drivers and connectors (10)

- [ ] **C-1 DuckDB metrics registration/context not cleaned up on open failure.** `ctx/cancel` and `registration` created before `reopenDB`; error path only closes remote — `runtime/drivers/duckdb/duckdb.go:191,213,219-229`. Fix: deferred cleanup until successful return.
- [ ] **C-2 ClickHouse `openHandle` leaks handles on ping/register failure and HTTP retry.** Ping failures return without `db.Close()`; retry reassigns `db` without closing the first — `runtime/drivers/clickhouse/clickhouse.go:940-960`. Fix: close failed handles; guard with cleanup defer.
- [ ] **C-3 Athena result pagination ignores the caller context.** `GetQueryResults(context.Background(), ...)` for next pages — `runtime/drivers/athena/olap.go:208`. Fix: store the query ctx on the rows object.
- [ ] **C-4 Redshift result pagination ignores the caller context.** `GetStatementResult(context.Background(), ...)` — `runtime/drivers/redshift/olap.go:242`. Fix: same as C-3.
- [ ] **C-5 Embedded ClickHouse binary download is unbounded.** Raw `http.Get(url)`, no ctx or timeout — `runtime/drivers/clickhouse/embed.go:437-448`. Fix: `NewRequestWithContext` + bounded client timeout.
- [ ] **C-6 HTTPS connector ignores configured `headers` and `path_prefixes`.** Connector-level values decoded but never merged into requests or enforced; only per-model headers apply — `runtime/drivers/https/https.go:53-58,259-266`. Fix: merge connector headers with documented precedence; enforce URL-prefix boundaries.
- [ ] **C-7 Azure connector uses default host credentials despite `allow_host_access=false`.** `NewDefaultAzureCredential` constructed unconditionally — `runtime/drivers/azure/object_store.go:174`. Fix: gate on `AllowHostAccess`.
- [ ] **C-8 Cached connector handles evicted on delete but not spec update.** `EvictWhere` only when `connector == nil`; credential rotation leaves the old handle live — `runtime/runtime.go:248-254`. Fix: evict on successful update too.
- [ ] **C-9 Driver handles opened after caller cancellation are leaked.** `openAndMigrate` returns an error when `ctx.Err() != nil` without closing the successfully opened handle — `runtime/connection_cache.go:169-182`. Fix: close the handle on every post-open error path.
- [ ] **C-10 Connector open-time validation is not cancelable.** StarRocks 30s background ping (`starrocks.go:350`), Pinot `dbx.Ping()` (`pinot.go:212`), ClickHouse `db.Ping()` (`clickhouse.go:942,959`), Slack `AuthTest()` (`slack.go:86`), Salesforce context-free login (`salesforce.go:217`). Fix: thread the caller context or defer validation to `Ping(ctx)`.

### D. Shared runtime utilities (6)

- [ ] **D-1 Blob download temp dir not removed when listing fails.** Temp dir created before `pagination.CollectAll`; failure returns without `os.RemoveAll` — `runtime/pkg/blob/download.go:71-83`. Fix: cleanup on all pre-iterator error returns.
- [ ] **D-2 Blob iterator `Next(ctx)` ignores cancellation.** Blocks on `<-it.batchCh` with no select on `ctx.Done()` — `runtime/pkg/blob/download.go:183`.
- [ ] **D-3 Git merge helpers don't accept a caller context.** `MergeWithStrategy`/`MergeWithBailOnConflict` (+`mergeInProgress`) use `Run(context.Background(), ...)` — `runtime/pkg/gitutil/sync.go:481-529`.
- [ ] **D-4 Connection-cache hanging checks nil-deref without `HangingFunc`.** Goroutine started whenever `CheckHangingInterval != 0`; optional callback called unconditionally — `runtime/pkg/conncache/conncache.go:137,446,450`. Fix: nil-check or require the callback in `New`.
- [ ] **D-5 `LazyWatcher.Subscribe` nil-derefs during concurrent close.** Unlocks after `count++`, then dereferences `w.watcher` which `Close` can nil out — `runtime/pkg/filewatcher/lazy.go:31,38-61`. Fix: capture the watcher under the mutex.
- [ ] **D-6 Admin repo git helpers don't receive the caller context.** `resetToRemoteTrackingBranch`, `setGitConfig`, `setRemoteURL`, `setFetchBranch`, `currentBranch` all use `context.Background()` — `runtime/drivers/admin/repo_git.go:397-433`.

### E. Frontend runtime client (web-common) (6)

- [ ] **E-1 Aborted requests keep waiting for a fresh JWT.** `waitForFreshJwt` polls on `disposed` + 60s deadline only; `req.signal` never consulted — `web-common/src/runtime-client/v2/runtime-client.ts:119-140`. Fix: reject immediately on abort.
- [ ] **E-2 Intentional SSE stop surfaces as an error.** `abort("SSE stream stopped by client")` string reason fails the `error?.name !== "AbortError"` suppression — `web-common/src/runtime-client/sse/sse-fetch-client.ts:110,126`. Fix: abort without a custom reason or treat `signal.aborted` as intentional.
- [ ] **E-3 SSE reconnect hook errors emitted after close/pause.** `onBeforeReconnect` catch emits `error` before checking `signal.aborted` — `web-common/src/runtime-client/sse/sse-connection.ts:190-201`.
- [ ] **E-4 Request queue cross-resource scheduling ignores request priority.** Every `NameEntry.weight` initialized to `ACTIVE_PRIORITY`; drain schedules by name-heap weight only — `web-common/src/runtime-client/v2/request-queue.ts:176-194,210-222`. Fix: track max pending priority per resource.
- [ ] **E-5 `removeByName` doesn't cancel or track active entries.** Only queued heap entries rejected; fired entries untracked, no AbortController — `web-common/src/runtime-client/v2/request-queue.ts:147-166,196-208`.
- [ ] **E-6 OLAP table lists not invalidated when only the connector changes.** Invalidation gated on table-name changes — `web-common/src/runtime-client/invalidation/resource-invalidators.ts:264-283`. Fix: include `connectorName != previousConnectorName`.

### F. Cloud frontend (web-admin) (7)

- [ ] **F-1 Embed iframe RPC handlers not disposed on route teardown.** `createIframeRPCHandler()` disposer discarded in `onMount`; `window` "message" listeners stack — `web-admin/src/routes/-/embed/+layout.svelte:111-123`, `web-common/src/lib/rpc.ts:79-91`. Fix: compose both cleanups in the `onMount` return.
- [ ] **F-2 "View As" falls back to real project permissions while mocked permissions load.** `mockUser.permissions ?? projectData?.projectPermissions` with `permissions` undefined during the fetch (#9804 fixed invalidation, not this) — `web-admin/src/features/projects/project-runtime.ts:38-39`, `web-admin/src/routes/[organization]/[project]/+layout.svelte:205-211`. Fix: treat mock mode as authoritative with a loading/empty permission set.
- [ ] **F-3 Alert/report metadata actions lack permission and confirmation gates.** Edit/one-click delete gated only on `!createdByCode`; `ProjectAccessControls` wraps only the owner block — `web-admin/src/features/alerts/metadata/AlertMetadata.svelte:149-163`, `web-admin/src/features/scheduled-reports/metadata/ReportMetadata.svelte:201-209`. Fix: wrap mutations in the access guard; confirm destructive actions.
- [ ] **F-4 Create-project form has duplicate submit paths.** Form `onsubmit={submit}` plus button `type="submit"` with `onClick={submit}` — `web-admin/src/features/projects/CreateProjectForm.svelte:148-175`. Fix: one submission path; disable while pending.
- [ ] **F-5 Report-run and project-wake polling loops are unbounded.** 1s loop until history changes; `while (true)` refetch every 2s until deployment appears — `web-admin/src/features/scheduled-reports/metadata/RunNowButton.svelte:39-53`, `web-admin/src/features/projects/RedeployProjectCTA.svelte:38-52`. Fix: abortable polling with timeout/max attempts and teardown cleanup.
- [ ] **F-6 Destructive delete dialogs close before the async delete finishes.** `onDelete(...)` not awaited; `open = false` immediately — `web-admin/src/features/public-urls/DeletePublicURLConfirmDialog.svelte:18-25`, `web-admin/src/features/projects/environment-variables/DeleteDialog.svelte:65-72`. Fix: await, show pending/error, close only on success.
- [ ] **F-7 GitHub connection component reads project data before the loading guard.** Top-level reactive destructure of `$proj.data` before the `{#if $proj.data}` check — `web-admin/src/features/projects/github/ProjectGithubConnection.svelte:19-21,37`.

### G. Shared frontend features (web-common) (11 — G-4 and G-5 fixed 2026-09-20; ids are stable and not renumbered)

- [ ] **G-1 Canvas `lastVisitedState` keyed by canvas name, not instance ID.** Module-level `Map` shared across projects/embeds — `web-common/src/features/canvas/stores/canvas-entity.ts:54,612,673`. Fix: key by the same `${instanceId}::${canvasName}` used by the store registry.
- [ ] **G-2 Dashboard time-range derivation applies stale async results.** `onTimeDimensionSelect`/`onSelectRange` apply whichever `deriveInterval` resolves last — `web-common/src/features/dashboards/filters/Filters.svelte:245-262,308-337`. Fix: generation counter or abort signal.
- [ ] **G-3 Canvas time-interval derivation applies stale async results.** `Promise.all(...).then(set)` with no guard — `web-common/src/features/canvas/stores/time-state.ts:163-220`. Fix: derived cleanup or generation counter.
- [ ] **G-6 Runtime file read errors swallowed as missing content.** `RuntimeFileIO.read` catches everything and returns `undefined` — `web-common/src/features/entity-management/file-io.ts:38-48`. Fix: typed not-found only for 404; rethrow the rest.
- [ ] **G-7 Uploaded filenames and generated DuckDB paths are not sanitized/quoted.** `data/${file.name}` unsanitized; path interpolated into single-quoted SQL and double-quoted YAML — `web-common/src/features/sources/modal/file-upload.ts:18`, `web-common/src/features/sources/sourceUtils.ts:126,129-148`. Fix: safe basename, URL-encode params, SQL-escape literals.
- [ ] **G-8 Generated local-file source detection never matches.** Generator emits `type: model` + `connector: duckdb` while `useIsLocalFileConnector` checks `yaml?.type === "duckdb"` — `web-common/src/features/sources/sourceUtils.ts:16-22,119-127`, `selectors.ts:60-62`. Fix: shared marker/helper for generation and detection.
- [ ] **G-9 Add-data import leaves committed secret edits after later failure.** Env session committed before model write; rollback only via user-initiated `cleanupAndBack` — `web-common/src/features/add-data/manager/steps/import.ts:218`, `GenerateDashboardStatus.svelte:105-121`. Fix: commit env after file creation succeeds, or auto-rollback on failure.
- [ ] **G-10 Source/model column derived stores leak inner subscriptions.** Inner `derived(...).subscribe(set)` return discarded — `web-common/src/features/sources/selectors.ts:81-102`, `web-common/src/features/models/selectors.ts:33-53`. Fix: return the unsubscribe from the outer callback.
- [ ] **G-11 File invalidation races overlapping write/delete events.** Fire-and-forget `void handleFileEvent(...)` with no per-path serialization while the write handler awaits `fetchContent(true)` — `web-common/src/features/entity-management/file-and-resource-watcher.ts:121-128`, `web-common/src/runtime-client/invalidation/file-invalidators.ts:64`. Fix: serialize per path or attach event versions.
- [ ] **G-12 Directory delete events don't clear child artifact state.** Delete cleanup inside `if (!event.isDir)`; dir deletes only refetch the list — `web-common/src/runtime-client/invalidation/file-invalidators.ts:55-101`. Fix: remove child artifacts/queries by prefix.
- [ ] **G-13 File invalidation ignores any path containing `.db`.** `event.path.includes(".db")` early return also matches e.g. `/models/my.dbt.sql` — `web-common/src/runtime-client/invalidation/file-invalidators.ts:49`. Fix: basename/extension rules for real DB artifacts only.

---

## Low (6)

- [ ] **L-1 `markPending` clears errors on the wrong node.** IDLE branch calls `updateErrorAndWarning(n, nil, nil)` inside the visit over descendant `dn` — `runtime/controller.go:1169`.
- [ ] **L-2 Partition JSON serialization panics instead of returning an error.** `json.Unmarshal` panic + `must(structpb.NewStruct(...))` in `modelPartitionToPB` — `runtime/server/controller.go:516-532`.
- [ ] **L-3 Canvas validation helpers ignore metrics-view lookup errors.** `continue` on any `r.C.Get` error in `validateMetricsViewTimeConsistency`/`validateRequiredFilters` — `runtime/reconcilers/canvas.go:266-268,337-339`.
- [ ] **L-4 Download-token parsing does not bound decompressed size.** `gzipDecompress` ends in unbounded `io.ReadAll` — `runtime/server/downloads.go:469-475`.
- [ ] **L-5 Auth-token cache entries not invalidated on revocation.** `PurgeAuthTokenCache` exists but has zero callers; revoked tokens valid for up to the cache TTL — `admin/auth_token.go:250-270`.
- [ ] **L-6 `Runtime.Repo` returns a non-nil release after rejecting a non-repo connector.** Returns `nil, release, err` after already releasing; defensive callers double-release — `runtime/connections.go:73-76`. Fix: return `nil, nil, err`.

---

## Suggested fix order

1. **Immediate security boundaries:** MF-39/MF-40 (stored XSS, unsafe schemes), MF-13 (magic-token privilege replay), MF-17 (destructive AI SQL), MF-30 (credentials in traces), MF-01/MF-19 (open redirects).
2. **Process-wide crash/deadlock risks:** MF-05/MF-06 (logbuffer), MF-07 (instance deletion), MF-08 (AI session send), MF-04 (Pinot panic), MF-21 (controller shutdown).
3. **Irrecoverable or silent data loss:** MF-14, MF-20, MF-29, MF-31, then MF-10/MF-11.
4. **Concurrency and capacity invariants:** MF-09, MF-22, MF-23, MF-25, MF-32, A-7, A-8.
5. **Wrong-result and stale-state bugs:** MF-26, MF-34 through MF-38, MF-41, MF-42, then the remaining medium items by area.

## Appendix: resolved-finding register

25 PRs merged 2026-07-11 → 2026-07-31 exactly resolved audit findings (see git history of `audit/MUST-FIX.md` predecessor for the original one-line register): #9683 (ClickHouse static DB-name collision), #9699 (`.env` pushed to wrong project), #9700 (ConnectorService authorization), #9701 (DashboardStateSync lock leak), #9702 (component-variable parse panic), #9703 (nil time ranges in AI resolver), #9704 (security-blind query cache key), #9705 (annotations-resolver access check), #9706 (BigQuery client leak), #9707 (alert notifier nil deref), #9709 (full rebuild after model rename), #9711 (partition catalog rows leaked), #9712 (`IS FALSE` inversion), #9714 (QueryBatch RPC removed), #9715 (aggregation export panic), #9716 (metrics SQL clause drops), #9717 (Summary security-filter leak), #9719 (`NOT BETWEEN` inversion), #9720 (decimal literal loss), #9721 (joins silently ignored), #9722 (canvas tab-group PDF export), #9723 (repo-root escape in AI file tools and repo drivers), #9724 (annotation validation parent fields), #9725 (parquet timeseries export panic), #9726 (`ResolveLocalPath` sandbox bypass).

No pending finding was resolved between 2026-07-31 and 2026-08-30.

Resolved 2026-08-30 → 2026-09-20 (verified at HEAD `008315ff3`): **MF-33** (canvas `timeAndFilterStore` subscription leak — eliminated by the unified expression filter refactor, #9746), **G-4** (falsy filter literals dropped — `expr.val !== undefined`), **G-5** (canvas in-list filter marker — mechanism removed by #9746/#9886). **MF-38** partially mitigated (guarded decompression in `ExpressionFilterManager` only) and remains pending.
