# Rill Deep Code Audit — Executive Summary & Roadmap

**Date:** 2026-07-19
**Method:** 24 area-scoped review agents, each followed by an independent adversarial verifier that re-checked every finding against the actual code (removing false positives, fixing severities and `file:line` citations), plus a cross-cutting product-idea synthesis. Static review only — no builds or tests were run.
**Corpus:** ~150k LOC Go (`runtime/`, `admin/`, `cli/`) + ~230k LOC TypeScript/Svelte (`web-common/`, `web-admin/`, `web-local/`).

Per-area detail lives in the sibling files (see [README.md](./README.md)). This document rolls everything up and categorizes it three ways, as requested: **by severity**, **by priority**, and **new features to build**.

## Remediation update (2026-07-25)

- Of the original 27 unique Tier-1 findings (28 witnesses because the cache leak was found twice), **15 are fixed and merged; 12 remain**.
- The all-source pass also includes `REVIEW_FINDINGS_TODO.md` and the later deep-dive audit files. After deduplication, **22 must-fix findings are fixed and 42 remain pending**.
- **24 merged PRs** from the 2026-07-11 through 2026-07-25 window exactly match audit findings. The authoritative pending list and PR mapping are in [../AUDIT_TODO.md](../AUDIT_TODO.md) (re-verified 2026-08-30).
- Counts below are the original audit snapshot and remain useful as corpus/severity totals; they are not decremented when remediation lands.

---

## 1. Headline numbers

**278 verified findings** across 24 areas.

| Severity | Count | | Priority | Count | Meaning |
|---|---|---|---|---|---|
| High | 29 | | **P0** | 28 | High severity **and** high confidence — confirmed serious defects, fix now |
| Medium | 90 | | **P1** | 44 | High severity (lower confidence) or high-confidence medium security/data/concurrency — next |
| Low | 159 | | **P2** | 47 | Remaining medium — planned |
| | | | **P3** | 159 | Low — backlog / opportunistic |

**By category:** correctness 89 · bug 40 · security 39 · product-improvement 34 · concurrency 30 · performance 16 · tech-debt 13 · readability 10 · data-loss 5 · simplification 2.

**Overall read:** the codebase is fundamentally sound and well-tested in its core (parser, metricsview, reconcilers all landed "fair" with strong test suites). The risk is concentrated in **cross-cutting classes** that recur across many areas — path-traversal sandboxing, security-context cache keys, resource cleanup on error paths, goroutine/lock lifecycle, and stale-async handling in the frontend. Fixing these as *classes* (shared helpers + test harnesses) will retire far more than the individual line items.

**Areas needing the most attention:** `runtime-core`, `resolvers`, `runtime-server`, `runtime-pkg`, `ai`, `drivers-object-store`, `drivers-connectors-ai`, `admin-billing-provision`, `frontend-dashboard-state` (all rated *needs-attention*).

---

## 2. Cross-cutting themes (the highest-leverage fixes)

These patterns appear in multiple independent areas. Each is worth a single coordinated fix rather than N scattered ones.

1. **Path traversal / sandbox escape (security).** The same weakness — building a filesystem path from user/agent input without confining it to the project root — appears in the AI file tools (`runtime/ai/file_read.go`, `file_write.go`, `develop_file.go`), the admin & file repo stores (`runtime/drivers/admin/repo.go`, `runtime/drivers/file/repo.go`), `runtime/pkg/fileutil/fileutil.go:ResolveLocalPath`, and the object-store `path_prefixes` checks (`runtime/drivers/*` — non-boundary prefix match). **One `SecureJoin`-style helper + one root-confinement helper, adopted everywhere, closes the entire class.**

2. **Security-context-blind cache keys (data leakage).** `runtime/resolver.go:210` builds the query-result cache key without `AdditionalRules`/`SkipChecks`, so results computed under one row-access policy can be served to a different security context. Found independently by the `resolvers` and `runtime-core` agents — it is one bug with two witnesses, and it is the single most dangerous item in the report.

3. **Missing authorization on data/control endpoints.** `ConnectorService` handlers do no authz at all; the `metrics_annotations` resolver never calls `CanAccess()`; the AI `query_sql` tool runs non-read-only SQL despite advertising `ReadOnlyHint`; magic tokens replay the creator's `admin` flag. A per-service authz test harness would prevent regressions.

4. **Resource cleanup missing on error paths.** ClickHouse and DuckDB leak DB handles / embedded processes / metric registrations when open partially fails; BigQuery leaks its client on *success*; connector handles opened after context cancellation are never closed. Consistent `defer`-with-rollback on partial-open is the pattern.

5. **Goroutine / lock lifecycle.** Unbuffered sends under locks and unbounded loops strand goroutines: `logbuffer.AddEntry` deadlocks on a slow subscriber, the AI session send blocks the producer, `DeleteInstance` blocks forever on store-delete failure, `QueryBatch` calls `srv.Send` concurrently, the JWKS bootstrap never breaks on success. The frontend analog is the `updating` lock in `DashboardStateSync` that is never released on two early returns, permanently freezing URL↔state sync.

6. **Crash-on-valid-input panics.** Unchecked type assertions and nil derefs turn ordinary requests into panics: a bare `value: 2024-01-01` in a component panics the *entire* project parse; `MetricsViewAggregation.Export` panics with no time range; the AI resolver panics when `time_range` is omitted; Pinot `All()` panics on a nil table; `logbuffer.GetLogs` panics when `limit == 0`. None of `runtime/parser`/several drivers have a `recover()`.

7. **Silent failure & partial-state (data-loss / trust).** Frontend `renameFileArtifact`/`deleteFileArtifact` swallow errors so callers navigate as if they succeeded; `duplicateFileArtifact` overwrites without a collision check; the CLI `DeployProject` pushes `.env` to the *wrong* project after a name collision; editable-deployment deletion discards uncommitted work when the checkpoint push fails; `CreateProject` can orphan a project (and a managed GitHub repo) on partial failure.

8. **Stale-async in the frontend.** Canvas time-interval derivation and dashboard time-range derivation apply out-of-order async results with no request sequencing; canvas `lastVisitedState` is keyed by canvas *name* not instance, leaking state across projects.

---

## 3. Priority backlog

### P0 — Original confirmed serious defects (28 witnesses / 27 unique; 15 fixed, 12 pending)

Grouped by theme. Format: **area** — issue — `location`.

**Security · path traversal / sandbox escape**
- **ai** — File tools (`file_read`/`file_write`/`develop_file`) allow paths outside the project root — `runtime/ai/file_read.go:52`, `file_write.go:61`, `develop_file.go:59`
- **drivers-connectors-ai** — `RepoStore` file APIs allow path traversal outside the project root — `runtime/drivers/admin/repo.go:257` (+ file repo, `runtime/repos.go:42`)
- ✅ [#9726](https://github.com/rilldata/rill/pull/9726) — **runtime-pkg** — `ResolveLocalPath` sandbox bypass

**Security · authorization & tenant isolation**
- ✅ [#9704](https://github.com/rilldata/rill/pull/9704) — **resolvers / runtime-core** — security-context-blind query cache key *(one bug, two witnesses)*
- ✅ [#9705](https://github.com/rilldata/rill/pull/9705) — **resolvers** — missing `metrics_annotations` access check
- ✅ [#9700](https://github.com/rilldata/rill/pull/9700) — **runtime-server** — missing `ConnectorService` authorization
- **web-admin** — Open redirect via unvalidated `redirect` query param — `web-admin/src/routes/[organization]/-/upgrade-callback/+page.svelte:34`
- ✅ [#9683](https://github.com/rilldata/rill/pull/9683) — **admin-billing-provision** — clickhouse-static database-name collision

**Concurrency · deadlocks, goroutine leaks, races**
- **runtime-pkg** — `logbuffer.AddEntry` blocks/deadlocks on a slow subscriber — `runtime/pkg/logbuffer/buffer.go:38-49`
- **runtime-core** — `DeleteInstance` blocks forever when the registry store delete fails — `runtime/registry.go:114-120`
- **ai** — Subscriber send under the session lock strands the producer on cancellation — `runtime/ai/ai.go:898-900`
- ✅ [#9714](https://github.com/rilldata/rill/pull/9714) — **runtime-server** — concurrent `QueryBatch` stream sends (RPC removed)
- ✅ [#9701](https://github.com/rilldata/rill/pull/9701) — **frontend-dashboard-state** — `updating` lock leak
- **admin-billing-provision** — Static runtime slot allocation is a TOCTOU race (slot overcommit) — `admin/provisioner/static/static.go:93-122`

**Panics on valid input**
- ✅ [#9702](https://github.com/rilldata/rill/pull/9702) — **parser** — component/canvas `value:` panic
- ✅ [#9715](https://github.com/rilldata/rill/pull/9715) — **queries** — aggregation export nil-time-range panic
- ✅ [#9703](https://github.com/rilldata/rill/pull/9703) — **resolvers** — AI resolver nil-time-range panic
- ✅ [#9707](https://github.com/rilldata/rill/pull/9707) — **reconcilers** — non-email alert notifier nil dereference
- **drivers-olap** — Pinot `All()` panics on nil table when any schema fetch fails — `runtime/drivers/pinot/information_schema.go:266-288`
- **runtime-pkg** — `logbuffer.GetLogs` panics (index `-1`) when `limit == 0` — `runtime/pkg/logbuffer/buffer.go:93`

**Data-loss, leaks & silent failure**
- ✅ [#9699](https://github.com/rilldata/rill/pull/9699) — **cli** — `.env` variables pushed to the wrong project
- ✅ [#9722](https://github.com/rilldata/rill/pull/9722) — **frontend-canvas** — tab groups omitted from PDF export
- **frontend-shared** — `renameFileArtifact`/`deleteFileArtifact` swallow errors; callers treat failure as success — `.../entity-management/actions/actions.ts:206-232`
- **frontend-runtime-client** — Failed chat send loses the user's draft and never rolls back the optimistic message — `.../chat/core/conversation.ts:378`
- ✅ [#9709](https://github.com/rilldata/rill/pull/9709) — **reconcilers** — full rebuild after model rename
- **drivers-object-store** — Redshift `UNLOAD` breaks on any query containing a single quote — `runtime/drivers/redshift/warehouse.go:120`
- ✅ [#9706](https://github.com/rilldata/rill/pull/9706) — **drivers-object-store** — BigQuery client leak

### P1 — Next (44)

Full list with `area` and one-line. (Details and `file:line` in each area file.)

- **admin-auth-rbac** — Magic tokens capture the creator's `admin` flag + custom attributes and replay them
- **admin-database** — Four billing read methods bypass the transaction context
- **admin-projects-deploy** — Unguarded `*proj.PrimaryDeploymentID` deref panics on hibernated/undeployed projects
- **admin-projects-deploy** — Editable deployment deletion discards uncommitted work when checkpoint push fails
- **admin-projects-deploy** — `LookupCanvas` has an inverted state check (NotFound when spec exists; panics)
- **admin-projects-deploy** — `SudoUpdateOrganizationCustomDomain` never propagates to deployments (dead condition)
- **admin-projects-deploy** — `CreateProject` leaves an orphaned project (and possibly managed GitHub repo) on partial failure
- **ai** — `query_sql` executes arbitrary (non-read-only) SQL despite `ReadOnlyHint`
- **cli** — Unsynchronized `pkceAuthenticators` map races between `/auth` and `/auth/callback`
- **cli** — `Helper.Client()/Telemetry()` mutate shared state from concurrent local-server RPCs
- **cli** — Open redirect in the local PKCE auth/logout handlers
- **cli** — Credentials and config in `~/.rill` written world-readable
- **cli** — `ListMatchingProjects` swallows non-inference errors and returns an empty list
- **drivers-connectors-ai** — Successfully-opened handle leaked when the open context is already cancelled
- **drivers-object-store** — `path_prefixes` access scoping bypassed by glob listing and downloads
- **drivers-object-store** — `path_prefixes` enforcement uses non-boundary prefix matches
- **drivers-object-store** — `ResolveLocalPath` root confinement uses a non-boundary prefix check
- **drivers-object-store** — Redshift result pagination fetches with `context.Background()`
- **drivers-object-store** — Athena/Redshift `autoDeleteFileIterator.Close` skips S3 cleanup when inner Close errors
- **drivers-olap** — DuckDB data race on `dbConnCount` in the metrics callback
- **drivers-olap** — ClickHouse `Open` leaks read/write DB handles + embedded server on post-connect failure
- **drivers-olap** — DuckDB `Open` leaks the metrics-callback registration when `reopenDB` fails
- **frontend-canvas** — In-list→Select conversion never clears the in-list marker (discarded `.filter()`)
- **frontend-canvas** — Async time-interval derivation has no request sequencing; stale results overwrite current
- **frontend-canvas** — `lastVisitedState` keyed by canvas name only; leaks state across instances/projects
- **frontend-dashboard-viz** — Big number hides all comparison deltas when the comparison value is exactly 0
- **frontend-dashboard-viz** — Dimension-table export always sorts by a measure, ignoring "sort by dimension"
- **frontend-runtime-client** — Status-page Tables list not invalidated on reconcile (wrong query keys)
- **frontend-runtime-client** — `removeByName/clear` leave a live abort listener that throws when it later fires
- **frontend-shared** — `waitForResourceReconciliation` compares `stateVersion` strings with `>`, hanging at digit boundaries
- **frontend-shared** — `duplicateFileArtifact` overwrites an existing `(copy)` file with no collision check
- **metricsview** — LIKE→regex conversion does not escape regex metacharacters (Druid ILIKE/search wrong results)
- ✅ [#9724](https://github.com/rilldata/rill/pull/9724) — **parser** — metrics-view annotation validation checked the parent's model/table
- **parser** — Incremental `.env` reparse can drop a real change when multiple `.env` files reparse together
- **plg-first-run** — Splash "Generate sample data" is ungated and can silently dead-end
- **queries** — Millisecond time grain panics timeseries null-fill via unchecked type assertion
- **queries** — Profiling APIs inline caller-supplied SQL expressions (`ColumnTopK.Agg`, column-timeseries)
- ✅ [#9711](https://github.com/rilldata/rill/pull/9711) — **reconcilers** — partition rows leaked on materialized-model deletion
- **reconcilers** — Report retry can re-deliver to notifiers that already succeeded
- **reconcilers** — `markPartitionsExecutedExceptFirst` stops after the first 1000 partitions
- **resolvers** — Built-in `/metrics-sql` API silently ignores its documented `additional_where`
- **runtime-pkg** — Telemetry intake sink uses `http.DefaultClient` with no timeout; `Close` can hang
- **runtime-server** — `OpenAudience` JWKS bootstrap loop never breaks on success, leaking refreshers
- **web-admin** — "View As" falls back to the real admin's permissions while mocked permissions load

### P2 (47) and P3 (159)

Planned / backlog. Enumerated in full within each area file. The P2 set is mostly medium-confidence correctness and product-robustness gaps; P3 is low-severity polish, readability, dead-code removal, and defensive hardening.

---

## 4. New features to build

The full, themed catalog is in **[product-ideas.md](./product-ideas.md)** — **55 synthesized ideas across 10 themes** (plus 144 area-specific ideas embedded in the individual reports). Highlights below.

### Standout bets (marked by the synthesis)

| Theme | Idea | Effort |
|---|---|---|
| Onboarding & PLG | One-click "aha": seeded datasets + offline sample-data generator | M |
| Data & Ingestion | Guaranteed-clean connector test-drive (test + preview before import) | M |
| Data & Ingestion | Unified object-store access-policy boundary | M |
| Metrics Layer | Ratio / "per"-dimension derived measures | L |
| Metrics Layer | Structured query-validation surface | M |
| Dashboarding | Deterministic dashboard state-sync engine | M |
| Dashboarding | Complete + tab-aware canvas export | M |
| AI / NL Analytics | Sandboxed AI: read-only introspection role + project-root jail | M |
| AI / NL Analytics | "Fix with AI" recovery, standardized everywhere | M |
| AI / NL Analytics | Scheduled AI narrative reports | M |
| Collaboration | Short share links via server-side state tokens | XL |
| Collaboration | Scoped magic tokens + origin-pinned embeds | M |
| Performance | Streaming / chunked exports | L |
| Performance | Security-aware cache keys + a correctness harness | M |
| Dev Experience | Reversible file operations: trash/undo + multi-select + conflict dialog | M |
| Dev Experience | `rill deploy --dry-run` + named profiles + CI login | M |
| Governance | One `SecureJoin` helper to kill the path-traversal class | S |
| Governance | Per-service authz test harness | S |
| Monetization | Post-deploy → share growth loop | S |
| Monetization | Per-session AI cost budget + usage-based metering | M |
| Monetization | Customer-facing usage & consumption dashboard | M |

Several standout bets (security-aware cache keys, sandboxed AI, tab-aware canvas export, deterministic state-sync, `SecureJoin`, authz harness) **double as the fix for a P0/P1 class** — building the feature retires the bug. Prioritize those.

---

## 5. PLG & first-user-experience

Full report: **[plg-first-user-experience.md](./plg-first-user-experience.md)** (and the agent's extra brainstorming in `plg_improvement_ideas.md`).

**Friction found in the first-run funnel (`rill start` → sample/import → auto-generated dashboard → deploy/share):**
- Splash "Generate sample data" is **ungated** and silently dead-ends if the developer agent can't start (P1).
- "Generate dashboard with AI" can **spin forever** on a hard backend error; example/blank unpack failures are **swallowed with no feedback**.
- Local example selection blocks on a **fixed 5-second `setTimeout`** then hard-reloads.
- `/deploy?org=<org>` redirects to a **non-existent route (404)** — broken share/deploy funnel.
- Every "add data" run generates **four resources** (model + metrics view + explore + canvas) per source — noisy for a first project.
- Auto-generated metrics view can yield a **dimension-less dashboard** with no warning; welcome examples are **hardcoded in the frontend** instead of using `ListExamples`.

**Top PLG features (activation-focused):** offline/deterministic sample-data generator · one-click seeded datasets on the splash · reconciliation-driven first-dashboard progress bar with skip · guided onboarding checklist · test-connection + table-preview before import · **post-deploy share funnel** + CLI "deploy to share" nudge · first-run activation-funnel instrumentation.

---

## 6. Suggested sequencing

**Now (this cycle) — P0 + the cheap class-killers.**
Ship the `SecureJoin` + root-confinement helpers (retires 6+ path-traversal findings), fix the security-context cache key (`runtime/resolver.go:210`) and add the security-aware cache-key harness, add `ConnectorService`/annotation authz, and clear the panic and deadlock P0s (parser, logbuffer, `DeleteInstance`, `QueryBatch`, `DashboardStateSync` lock). Fix the CLI wrong-project `.env` push and the frontend error-swallowing.

**Next — P1 + PLG activation.**
Driver resource-leak cleanups, `path_prefixes` boundary enforcement, magic-token scope fix, editable-deployment checkpoint safety, `CreateProject` rollback. In parallel, ship the PLG activation set (seeded datasets, first-dashboard progress, fix `/deploy` 404, connector test-drive) and the post-deploy share loop.

**Later — P2/P3 + bigger bets.**
Structured query-validation surface, streaming exports, ratio measures, short share links, customer-facing usage dashboard, and the dead-code / readability cleanup.

---
*Generated by a multi-agent code audit (find → adversarially verify → synthesize). Every P0/P1 was independently re-checked against the source before inclusion.*
