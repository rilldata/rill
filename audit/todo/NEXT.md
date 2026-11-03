# NEXT — after NOW clears (correctness, races, user-visible bugs)

Ids and full detail in [`../../AUDIT_TODO.md`](../../AUDIT_TODO.md).

## 1. Remaining must-fix: wrong results and stale state

- [ ] **MF-35** — Pivot row-total merging can assign cells to the wrong dimension rows.
- [ ] **MF-36** — Pivot expanded-table cache discards new pages and fresh data.
- [ ] **MF-37** — Module-level pivot `lastKey` breaks pagination across instances (explore `pivot-data-config.ts:25` + canvas `pivot/util.ts:106`). Verified still present 2026-09-20.
- [ ] **MF-34** — Canvas row insertion desyncs `touchedRows` from indices (now `layout-util.ts:186-325`, `generateArrayRearrangeFunction`). Verified still present 2026-09-20.
- [ ] **MF-41** — TDD global scroll position never reset or scoped per dashboard.
- [ ] **MF-26** — Migration reconciliation re-executes the current version every reconcile.
- [ ] **MF-27** — Embed singleton retains previous token/host/instance/external user across embeds.
- [ ] **MF-24** — Runtime-client JWT can become `Bearer undefined` after waiting for refresh.
- [ ] **MF-38** — Finish the `gzipped_state` hardening: guard `convertURLToExplorePreset.ts:92` and replace the `String.fromCharCode(...spread)` (partially mitigated by #9746's guarded path).
- [ ] **MF-42** — Metrics SQL passes unknown opcodes into invalid IR; fail early with positioned errors.
- [ ] **MF-12** — Redshift `UNLOAD` breaks on queries containing a single quote.
- [ ] **MF-10** — File rename/delete helpers swallow errors; callers navigate as if success.
- [ ] **MF-11** — Failed chat send loses the draft and keeps a false optimistic message (`conversation.ts` still resolves on "close" only). Verified still present 2026-09-20.

## 2. Remaining must-fix: concurrency and capacity

- [ ] **MF-09** — Static runtime slot allocation TOCTOU race (overcommit). Verified still present 2026-09-20.
- [ ] **MF-23** — ClickHouse partial-open failures leak pools/embedded process. Bundle with **C-1** (DuckDB open cleanup), **C-2** (`openHandle` ping-failure leaks — verified still present), **C-9** (post-cancel handle leak) as one "driver open/close hygiene" PR.
- [ ] **MF-32** — rduckdb registers a non-removable DB-stats callback per write-handle open (now `db.go:953`).
- [ ] **A-7** — Credit-trial start still not serialized (idempotent on retry, double-grants on concurrency). Verified still present 2026-09-20.
- [ ] **A-8** — `prodErr` data race in project-wide deployment updates (assigned inside `errgroup` goroutines, `deployments.go:148-175`). Verified still present 2026-09-20.
- [ ] **A-3** — `used_flusher` shutdown races its background goroutine.
- [ ] **D-5** — `LazyWatcher.Subscribe` nil-derefs during concurrent close.

## 3. Security-adjacent mediums

- [ ] **C-6** — HTTPS connector ignores configured `headers` and `path_prefixes` (documented security knob that does nothing).
- [ ] **C-7** — Azure connector uses default host credentials despite `allow_host_access=false`.
- [ ] **G-7** — Uploaded filenames / generated DuckDB paths unsanitized into SQL/YAML.
- [ ] **L-5** — Revoked auth tokens stay valid until cache TTL (`PurgeAuthTokenCache` has zero callers).
- [ ] **L-4** — Download-token gzip decompression unbounded (`io.ReadAll`).

## 4. High-value user-visible mediums

- [ ] **F-2** — "View As" falls back to real project permissions while mocked permissions load (#9840 added cloud-edit view-as; the fallback at `project-runtime.ts:38-39` is unchanged).
- [ ] **B-8** — Report notifier retry duplicates earlier successful sends (single `sent` flag; verified still present 2026-09-20).
- [ ] **B-11** — Client-supplied time-grain enums panic instead of validating (`query.go:356,383,410`).
- [ ] **B-5** — Metrics-view reconciliation treats transient model-lookup errors as "external table".
- [ ] **B-7** — Alert ref-hash lookup failures silently change the hash (spurious alert re-sends).
- [ ] **G-2 + G-3** — Stale async time-interval derivations race in explore filters and canvas time-state (same generation-counter fix pattern).
- [ ] **E-1 + E-2** — Aborted requests keep waiting for JWT; intentional SSE stop surfaces as an error.
- [ ] **F-5 + F-6** — Unbounded polling loops; destructive dialogs close before the delete finishes.
- [ ] **L-2 + L-3** — Partition JSON serialization panics; canvas validators swallow metrics-view lookup errors (both verified still present 2026-09-20).

## 5. GitHub backlog hygiene (near-zero engineering)

- [ ] Close the **12 verified-implemented issues** from [`OPEN_ISSUES_REPORT.md`](../../OPEN_ISSUES_REPORT.md) Phase 0 (`#3764 #5202 #5061 #5775 #5575 #4007 #4446 #5469 #5645 #4675 #4821 #5905`) — re-confirm each is still open first.
- [ ] Re-scope the 11 partially-implemented issues to their remaining gaps (table in the report).
- [ ] Apply the proposed `type:` / `area:` / `priority:` label normalization.
