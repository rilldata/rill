# NOW — fix immediately (security, crashes, data loss)

Ids and full detail in [`../../AUDIT_TODO.md`](../../AUDIT_TODO.md). Ordered for execution; the first two groups are exploitable or outage-grade.

## 1. Exploitable security boundaries

- [ ] **MF-39** — Stored XSS: TDD row headers render dimension values through `innerHTML` (`TDDTable.svelte`). Same class: **PIVO-12** (`RegularTable.svelte`) — fix together with one shared escaping path.
- [ ] **MF-40** — `javascript:`/`data:` URIs reach dashboard links (`makeHref`, no scheme allowlist); #9588 widened exposure to leaderboard + pivot links.
- [ ] **MF-13** — Magic tokens replay the creator's `admin` flag and privileged attributes into `{{ .user.admin }}` policies (`magic_tokens.go`, `runtime_jwt.go`).
- [ ] **MF-17** — AI `query_sql` executes mutations despite `ReadOnlyHint` (`runtime/ai/query_sql.go`). Gate to read-only statements or route through a read-only session.
- [ ] **MF-30** — ClickHouse object-store credentials embedded verbatim in traced SQL (`model_executor_objectstore_self.go`).
- [ ] **MF-01** — Unsafe web-admin redirects after billing/device flows (open redirect + unencoded `user_code`).
- [ ] **MF-19** — Local PKCE auth/logout handlers accept external `?redirect=` targets (`cli/pkg/local/server.go`).
- [ ] **MF-15 + MF-16** — Object-store `path_prefixes` bypassed by glob/download paths, and prefix checks not boundary-aware (`data` admits `database/`). One PR: boundary-aware check applied to all three entry points across s3/gcs/azure.
- [ ] **MF-18** — Local credentials written world-readable (`dotrill.go`: `0o644`, `os.ModePerm`).

## 2. Process-wide crash / deadlock risks

- [ ] **MF-05** — `logbuffer.GetLogs(limit=0)` panics (indexes `logs[-1]`). Trivial guard + test.
- [ ] **MF-06** — Slow log subscriber deadlocks instance logging (unbuffered send under `b.mu`; blocking `srv.Send` fan-out).
- [ ] **MF-07** — `DeleteInstance` blocks forever on nil channel after registry delete fails (`registry.go`). Verified still present 2026-09-20.
- [ ] **MF-08** — AI session `AddMessage` sends to unbuffered subscriber channels while holding `s.mu`; exited consumer strands the producer (`ai.go:672-916`). Verified still present 2026-09-20.
- [ ] **MF-04** — Pinot schema enumeration panics on transient per-table failures (nil slots into `sort.Slice`).
- [ ] **MF-21** — Late reconcile completions block forever after controller shutdown (plain `c.completed <- inv`, now `controller.go:1386`).

## 3. Irrecoverable or silent data loss

- [ ] **MF-14** — Editable deployment deletion proceeds after checkpoint `GitPush` failure → user work destroyed (`deployments.go:410-480`). Pair with **admin-projects-deploy-01** and **deep-cloud-deploy-13** (same lifecycle).
- [ ] **MF-20** — Failed delete reconciliation still hard-deletes the catalog resource (`controller.go:1508-1521`). Verified still present 2026-09-20.
- [ ] **MF-29** — rduckdb deletes the old remote version even when new `meta.json` upload fails (`remote.go:244-255`). Pair with **MF-28** (deleteRemote panics + wrong `meta.json` path) in one rduckdb PR.
- [ ] **MF-31** — Salesforce batch-result fetch errors become empty files; ingestion silently partial (`bulk.go`).
- [ ] **MF-25** — Exhausted deployment reconcile jobs never persist a terminal error state; deployments show "Provisioning…" forever (`river.go`).

## 4. One-line / trivial correctness fixes (bundle into a single cleanup PR)

- [ ] **B-10** — `MetricsViewAnnotations` assigns `limit = &req.Offset` and never sets `offset` (`queries_metrics.go:630-640`). Verified still present 2026-09-20 — genuinely a one-character fix.
- [ ] **L-1** — `markPending` clears errors on `n` instead of descendant `dn` (`controller.go:1220`).
- [ ] **L-6** — `Runtime.Repo` returns a non-nil release after already releasing (`connections.go`).
- [ ] **MF-22** — DuckDB `dbConnCount` metrics callback reads without the lock (data race; run with `-race` to confirm).
