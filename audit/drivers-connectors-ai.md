# Connector Framework & AI/Notifier Drivers — Code Review

**Scope:** `runtime/drivers/{drivers,connectors,registry,catalog,repo,models,notifier,ai,information_schema}.go`, `runtime/drivers/{admin,openai,gemini,claude,slack,mock}`, plus the connection lifecycle in `runtime/connection_cache.go`, `runtime/connections.go`, `runtime/pkg/conncache/conncache.go`, and the file-API entrypoints in `runtime/repos.go` / `runtime/server/repos.go` / `runtime/drivers/file/repo.go`.  |  **Health:** needs-attention

## Overview

This area defines the `Driver`/`Handle` abstraction and the concrete drivers for the control-plane admin service, the three first-party LLM providers (OpenAI, Gemini, Claude), the Slack notifier, and test mocks. The abstraction is clean and the connection cache (`conncache`) is carefully written and well-tested for the hard concurrency cases (open-during-close, close-during-open, hanging detection). The AI drivers are mostly mechanical proto↔SDK translation.

The biggest risks are on the repo/file surface and the open/close lifecycle rather than in the LLM translation code. The `RepoStore` file APIs perform no path-traversal validation at any layer, so a caller with edit access can read/write/delete files outside the project root — a tenant-isolation concern on shared runtimes. The connection cache leaks a successfully-opened handle if the context is cancelled in the same instant, and the admin repo nil-derefs when a deployment is archive-backed (no Git). The notifier fails the entire recipient batch on the first Slack API error. None of the AI message-conversion functions have unit tests, and Claude has no test file at all.

## Findings

### drivers-connectors-ai-01 [HIGH] RepoStore file APIs allow path traversal outside the project root
- **Category:** security  **Confidence:** high  **Location:** runtime/drivers/admin/repo.go:257 (and :150, :188, :221, :295, :325, :361); runtime/drivers/file/repo.go:74/93/128; runtime/repos.go:42; runtime/server/repos.go:139
- **Problem:** Every `RepoStore` file operation resolves the caller-supplied path with a bare `filepath.Join(root, path)` and no check that the result stays within `root`. The gRPC handlers (`PutFile`, `GetFile`, `DeleteFile`, `RenameFile`, `CreateDirectory`, `UploadMultipartFile`) and the `runtime.*File` wrappers pass `req.Path` straight through. `IsIgnored` only matches a fixed prefix list and does not defend against `..`. Because `filepath.Join` cleans the result, a path like `../../../../etc/…` or `/../../<other-instance-id>/…` escapes the instance's repo root.
- **Failure scenario:** A Rill Cloud user with `EditRepo` on a project calls `PutFile(path="../../../<victim-instance>/rill.yaml", blob=…)`. On a runtime host that co-locates multiple instances under a common data dir, this overwrites another tenant's project files (or writes into arbitrary host locations writable by the runtime process). `GetFile`/`DeleteFile` give the same read/delete primitive. In Rill Developer the same escape reaches anywhere on the developer's filesystem.
- **Recommendation:** Reject or safely resolve traversal at the driver boundary. Clean the path and require it to remain under `root` (e.g. `rel, err := filepath.Rel(root, filepath.Join(root, path)); if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) { reject }`, or use `os.Root`/`filepath.IsLocal` on Go 1.20+). Apply it in one shared helper used by both `admin/repo.go` and `file/repo.go`, and add regression tests with `..` inputs.

**Verification:** CONFIRMED — every RepoStore file op resolves the caller path with a bare `filepath.Join(root, path)` and no containment check. `file/repo.go` `Get` (:74), `Put` (:128), `MkdirAll` (:151), `Rename` (:163/:165), `Delete` (:178), `Hash` (:93), `Stat` (:113) all join `c.root` with the caller path directly. `admin/repo.go` does the same in `Get` (:150), `Stat` (:221), `Put` (:257), `MkdirAll` (:295), `Rename` (:325/:326), `Delete` (:361). The server handlers in `runtime/server/repos.go` pass `req.Path` straight into `s.runtime.GetFile/PutFile/DeleteFile/RenameFile` (:116, :139, :181, :203), and `runtime/repos.go` wrappers pass `path` unmodified to `repo.Get/Put` (e.g. PutFile at :42-51). `drivers.IsIgnored` only matches a fixed prefix list, no `..` defense. No `filepath.Rel`/`filepath.IsLocal`/containment check anywhere in the chain.

**Manual Reproduction Steps (for QA):**
1. Start Rill Developer on a project (`rill start`) so the local runtime and file API are running.
2. Note the project root directory and create a marker file just outside it, e.g. `~/outside-project/secret.txt` with some text.
3. Using the runtime file API (the same endpoint the editor uses), issue a GetFile request with a traversal path, e.g. `GET /v1/instances/{instance_id}/files/file?path=../outside-project/secret.txt` (or a PutFile with `path=../outside-project/pwned.yaml` and a blob).
4. Observe that the contents of the out-of-root file are returned (or the out-of-root file is created/overwritten), confirming the path escapes the project root.
5. (Cloud analog) A user with EditRepo can call PutFile with `path=../../<other-instance-id>/rill.yaml` on a host co-locating instances to reach another tenant's files.

### drivers-connectors-ai-02 [MEDIUM] Successfully-opened handle is leaked when the open context is already cancelled
- **Category:** bug  **Confidence:** high  **Location:** runtime/connection_cache.go:169-182
- **Problem:** `openAndMigrate` opens the driver, then does `if err == nil && ctx.Err() != nil { err = fmt.Errorf("timed out …") }` and later `if err != nil { return nil, err }`. When `drivers.Open` returns a valid `handle` (err == nil) but the context expired in the meantime, `err` is overwritten and the function returns without ever calling `handle.Close()`. The Migrate branch below correctly closes on failure, but this earlier branch does not. Since `conncache` only stores the returned error (no handle), the opened handle is orphaned forever.
- **Failure scenario:** The cache's `OpenTimeout` (10 min) elapses just as a DuckDB/ClickHouse connection finishes opening, or the whole cache is closed (`c.ctx` cancelled) during an in-flight open. The driver's connection, file locks, and background goroutines stay alive with no owner and are never released, accumulating across restarts/timeouts.
- **Recommendation:** Close the handle before returning in the cancellation branch, e.g. `if err == nil && ctx.Err() != nil { handle.Close(); return nil, fmt.Errorf("timed out while opening driver %q", cfg.driver) }`, or restructure so any non-nil `handle` on an error path is closed.

**Verification:** CONFIRMED — at `runtime/connection_cache.go:169-172`, `handle, err := drivers.Open(...)` then `if err == nil && ctx.Err() != nil { err = fmt.Errorf("timed out ...") }`. When `Open` succeeds but the context is already cancelled, `err` is overwritten and the `if err != nil { return nil, err }` at :180-182 returns without ever calling `handle.Close()`. The successfully-opened `handle` local is never returned nor closed — orphaned. The Migrate branch below (:184-191) correctly closes on failure, confirming the asymmetry.

**Manual Reproduction Steps (for QA):**
Not user-reproducible via UI (requires a precise context-cancellation race at the instant a driver finishes opening). Confirm in code: `runtime/connection_cache.go:169-182` — the branch where `err == nil && ctx.Err() != nil` overwrites `err` and returns via `if err != nil` without calling `handle.Close()`. Compare against the Migrate branch at :184-191 which does close on error.

### drivers-connectors-ai-03 [MEDIUM] Admin repo nil-derefs `r.git` on archive-backed (non-Git) deployments
- **Category:** bug  **Confidence:** medium  **Location:** runtime/drivers/admin/repo.go:392 (Watch); also :251 (Put), :289 (MkdirAll), :316 (Rename), :355 (Delete)
- **Problem:** These methods guard with `if r.git != nil && !r.git.editable() { return … }` and then unconditionally call `root := r.git.root()`. When the deployment is served from a tarball archive (`meta.GitUrl == ""` in `checkHandshake`, so `r.git == nil` but `r.archive != nil`), the guard is skipped and `r.git.root()` dereferences a nil pointer. `Watch` additionally leaves the read lock held on this panic path (it unlocks manually rather than via `defer`). The contrast is stark: `Commit`/`CommitAndPush`/`MergeToBranch`/`Status`/`Diff` all correctly `if r.git == nil { return … }` first, and `roots()` only appends `r.git.root()` when non-nil. Note that `Root`, `Hash`, `CommitHash`, and `CommitTimestamp` are actually *safe* on archive-backed deployments because they check `if r.archive != nil` before falling back to `r.git.root()`; they would only nil-deref in the practically-unreachable "virtual-only" state where both git and archive are nil. So the reachable defect is specifically `Put`/`MkdirAll`/`Rename`/`Delete`/`Watch`.
- **Failure scenario:** An archive-backed instance with `rill.watch_repo` enabled (or any client that calls `Put`/`Delete`/`Rename` on such an instance) triggers `r.git.root()` on a nil `r.git`, panicking the request goroutine instead of returning a clean "repo is not editable/watchable" error.
- **Recommendation:** Restructure the guards to handle the nil-git case explicitly, e.g. `if r.git == nil || !r.git.editable() { return fmt.Errorf("repo is not editable") }` (and, in `Watch`, release the read lock before returning). Prefer `r.roots()`/archive fallbacks where a read root is needed.

**Verification:** CONFIRMED — in `runtime/drivers/admin/repo.go`, `Put` (:248-251), `MkdirAll` (:286-289), `Rename` (:313-316), `Delete` (:352-355), and `Watch` (:388-392) all guard with `if r.git != nil && !r.git.editable() { return ... }` then unconditionally call `root := r.git.root()`. When `r.git == nil` (archive-backed: `meta.GitUrl == ""` sets `r.git = nil` at :937), the guard is skipped and `r.git.root()` nil-derefs. `Watch` unlocks manually (`r.mu.RUnlock()` inside the guard at :389) so the panic path after the guard leaves the read lock held. By contrast `Commit`/`CommitAndPush`/`MergeToBranch`/`Status`/`Diff` all check `if r.git == nil` first (:574, :598, :619, :451, :498), and `Root`/`Hash`/`CommitHash`/`CommitTimestamp` check `if r.archive != nil` before falling back (:95, :177, :633, :647) — confirming the audit's corrected scope.

**Manual Reproduction Steps (for QA):**
1. Deploy a project to Rill Cloud in an archive-backed (tarball) configuration rather than a Git-connected one, so the runtime serves files from an archive with no Git remote.
2. Ensure the instance has `rill.watch_repo` enabled (or trigger any edit operation on the project).
3. Trigger a file write/rename/delete on that instance (e.g. edit a file through the API), or let the project parser start its repo watch.
4. Observe that the request goroutine panics (nil pointer dereference on `r.git`) instead of returning a clean "repo is not editable/watchable" error.

If an archive-backed editable/watched instance cannot be produced in the test environment, confirm in code: `runtime/drivers/admin/repo.go:251/289/316/355/392` — `root := r.git.root()` reached when `r.git == nil`.

### drivers-connectors-ai-04 [MEDIUM] Updating a connector's config does not evict the stale cached handle
- **Category:** bug  **Confidence:** medium  **Location:** runtime/runtime.go:247-253; registry.go edit path (EditInstance called with `restartController=false`)
- **Problem:** `UpdateInstanceConnector` evicts cached connections only when the connector is being *removed* (`if connector == nil`, runtime.go:248-253), and calls `EditInstance(ctx, inst, false)`, which does not restart the controller and therefore does not reach the `evictInstanceConnections` call in `registryCache` (registry.go:496-497). That eviction site is explicitly commented "When an instance is edited, connector config may have changed. So we want to evict all open connections for that instance … this is the only place where we can do it safely" — confirming eviction-on-edit is the intended behavior, yet the in-place-upsert path bypasses it. Reachable on any connector spec change via the connector reconciler (reconcilers/connector.go:90, which calls `UpdateInstanceConnector(ctx, …, t.Spec)` when the spec hash changes). Because the cache key (`generateKey`) includes the config, the next acquire opens a *new* handle under a new key while the old handle remains cached until idle LRU eviction. Harmless for most drivers (a lingering idle connection); the concrete hazard is drivers holding an exclusive OS resource keyed by something unchanged by the edit — e.g. an embedded DuckDB file when a non-DSN setting (threads, memory_limit) changes.
- **Failure scenario:** A connector that holds an exclusive resource (e.g. an embedded DuckDB file, which permits a single read-write handle) has a non-DSN setting changed (threads, memory_limit, etc.). The new config produces a new cache key and a second `Open` against the same file, which conflicts with the still-cached old handle ("database is locked") until the idle LRU eventually closes it.
- **Recommendation:** On connector upsert, proactively evict cached connections for that `instanceID`+`name` (as is already done on removal) so the stale handle is closed before the new config is used.

**Verification:** CONFIRMED — `UpdateInstanceConnector` (`runtime/runtime.go:211`) calls `EditInstance(ctx, inst, false)` and only evicts via `r.connCache.EvictWhere(...)` inside `if connector == nil` (the removal branch). `EditInstance` (:97) delegates to `registryCache.edit`, and `edit` (registry.go:375) only calls `r.restartController(iwc)` when `restartController` is true (:391-392); `restartController` is the only path reaching `r.rt.evictInstanceConnections` (registry.go:497). Since the upsert passes `false`, the stale handle is not evicted. The connector reconciler reaches this on any spec change: `reconcilers/connector.go:90` calls `UpdateInstanceConnector` when `specHash != t.State.SpecHash`. Because `generateKey` (connection_cache.go:195-209) includes config values, a changed config yields a new key and a second `Open`.

**Manual Reproduction Steps (for QA):**
1. In a Rill project, define a connector that holds an exclusive OS resource keyed by something other than the config value you will change — e.g. an embedded DuckDB connector backed by a file (single read-write handle).
2. Start the project and let the connector open (issue a query so the connection is cached).
3. Edit the connector YAML to change a non-DSN setting that does not change the underlying file path (e.g. `threads` or `memory_limit`), and save so the connector reconciler runs.
4. Immediately issue a query that uses the connector. Observe a "database is locked"/conflict error (a second Open against the same file) that persists until the old idle handle is LRU-evicted.

If the exclusive-resource setup is impractical, confirm in code: `runtime/runtime.go:211-253` (eviction only when `connector == nil`) plus `registry.go:391-392,497` (eviction gated on `restartController`).

### drivers-connectors-ai-05 [MEDIUM] Slack notifier aborts the whole recipient batch on the first send error
- **Category:** product-improvement  **Confidence:** high  **Location:** runtime/drivers/slack/notifier.go:196-202 (sendTextToChannels), :214-224 (sendTextToUsers), :90-108 (sendReportToUsers); SendScheduledReport:68-76
- **Problem:** Each fan-out loop returns immediately on the first Slack API error, so remaining channels/users never receive the message, and there is no retry/backoff for transient failures. `SendScheduledReport` also sequences channels → webhooks → users; a failure in the channels step skips webhooks and users entirely.
- **Failure scenario:** A report is configured for five users; the first `GetUserByEmail`/`PostMessage` fails (rate limit, a single deactivated user, transient 5xx). The other four users silently get nothing, and the alert/report is reported as failed even though most deliveries could have succeeded.
- **Recommendation:** Isolate per-recipient failures (continue the loop, accumulate errors with `errors.Join`), add bounded retry/backoff for retryable Slack errors, and return a partial-success signal so the scheduler can distinguish "all failed" from "some failed".

**Verification:** CONFIRMED — every Slack fan-out loop returns on the first error with no per-recipient isolation or retry: `sendTextToChannels` (notifier.go:196-201), `sendTextToUsers` (:214-223), `sendReportToUsers` (:100-107), and `sendTextViaWebhooks` (:228-236). `SendScheduledReport` (:68-76) sequences channels → webhooks → users, each with `if err := ...; err != nil { return err }`, so a failure in the channels step skips webhooks and users entirely.

**Manual Reproduction Steps (for QA):**
1. Configure a scheduled report or alert with a Slack notifier targeting several users (e.g. 5 email addresses).
2. Make the first recipient fail deterministically: use an email that does not resolve to a Slack user (so `GetUserByEmail` errors), while the remaining recipients are valid.
3. Trigger the report/alert send.
4. Observe that none of the later recipients receive the message and the whole send is reported as failed, even though the valid users could have been delivered to.

### drivers-connectors-ai-06 [LOW] Gemini `Complete` nil-derefs `res.UsageMetadata`
- **Category:** bug  **Confidence:** medium  **Location:** runtime/drivers/gemini/gemini.go:333-335
- **Problem:** After converting the response, the code reads `res.UsageMetadata.PromptTokenCount` / `.CachedContentTokenCount` / `.CandidatesTokenCount`. Confirmed against the vendored SDK (`google.golang.org/genai@v1.44.0/types.go:2680`): `UsageMetadata` is `*GenerateContentResponseUsageMetadata` with `json:"...,omitempty"` and can be nil. `convertResponseToRillMessage` checks `Candidates`/`Content` but nothing checks `UsageMetadata`. The OpenAI and Claude drivers use value-typed `res.Usage`, so they are not exposed to this. Reachability caveat: the real Gemini `GenerateContent` endpoint reliably populates usage metadata, so this is primarily a hazard with proxies/custom base URLs or unusual responses — hence a defensive-nil-check finding rather than a routinely-hit crash.
- **Failure scenario:** A Gemini response (or a proxy/custom base URL) returns without usage metadata; the driver panics on the nil deref instead of returning the completion. The panic surfaces as an opaque internal error for the AI request.
- **Recommendation:** Guard with `if res.UsageMetadata != nil { … }` before reading token counts (leave them at 0 otherwise).

**Verification:** PLAUSIBLE — at `runtime/drivers/gemini/gemini.go:332-335` the driver reads `res.UsageMetadata.PromptTokenCount`, `.CachedContentTokenCount`, `.CandidatesTokenCount` with no nil check; `convertResponseToRillMessage` validates `Candidates`/`Content` but not `UsageMetadata`. The genai SDK types `UsageMetadata` as a nilable pointer with `omitempty`. OpenAI/Claude use value-typed `res.Usage` so they are unaffected. Reachability is conditional: the real Gemini endpoint reliably populates usage metadata, so a crash requires a proxy/custom base URL or an unusual response returning nil — hence PLAUSIBLE (latent defensive gap) rather than CONFIRMED.

**Manual Reproduction Steps (for QA):**
Not routinely user-reproducible against the real Gemini API (it always returns usage metadata). To force it: point the Gemini connector at a custom `base_url`/proxy that returns a valid completion response with the `usageMetadata` field omitted, then make any AI request through it; the request panics with a nil pointer dereference. Otherwise confirm in code: `runtime/drivers/gemini/gemini.go:333-335` reads `res.UsageMetadata.*` with no preceding nil check.

### drivers-connectors-ai-07 [LOW] Slack alert ERROR status sets the "Open" link to the edit URL
- **Category:** bug  **Confidence:** high  **Location:** runtime/drivers/slack/notifier.go:137
- **Problem:** In `SendAlertStatus`, the `ASSERTION_STATUS_ERROR` branch sets `OpenLink: htemplate.URL(s.EditLink)` (the PASS and FAIL branches correctly use `s.OpenLink`). This is a verified copy/paste slip: the `alert_status.slack` template renders `<{{ .OpenLink }}|Open in browser>` and, separately, `<{{ .EditLink }}|click here>` for editing — so on an errored alert both links resolve to the edit URL and the "Open in browser" action never reaches the dashboard.
- **Failure scenario:** A user receives an errored-alert Slack message and clicks "Open"; they land on the alert editor instead of the alert/dashboard view.
- **Recommendation:** Use `htemplate.URL(s.OpenLink)` for `OpenLink` in the error branch.

**Verification:** CONFIRMED — in `SendAlertStatus` (`runtime/drivers/slack/notifier.go`), the ERROR branch sets `OpenLink: htemplate.URL(s.EditLink)` at :137 (with `EditLink: htemplate.URL(s.EditLink)` at :138), whereas the PASS branch (:120-121) and FAIL branch (:128-129) correctly use `s.OpenLink`. The `alert_status.slack` template renders `<{{ .OpenLink }}|Open in browser>` and separately `<{{ .EditLink }}|click here>`, so on an errored alert both resolve to the edit URL.

**Manual Reproduction Steps (for QA):**
1. Create an alert wired to a Slack notifier, using a valid dashboard/metrics view.
2. Cause the alert to error during evaluation (e.g. reference a dimension/measure that no longer exists, or otherwise make the alert query fail so its status is ERROR rather than PASS/FAIL).
3. Wait for the errored-alert Slack message to arrive.
4. Click the "Open in browser" link. Observe it lands on the alert editor (edit URL) instead of the dashboard/alert view, i.e. the same destination as the "click here" edit link.

### drivers-connectors-ai-08 [LOW] AI drivers expose the API key via `Handle.Config()` while Slack deliberately does not
- **Category:** security  **Confidence:** low  **Location:** runtime/drivers/openai/openai.go:268-272; runtime/drivers/claude/claude.go:229-233; runtime/drivers/gemini/gemini.go:254-258 (vs runtime/drivers/slack/slack.go:98-100)
- **Problem:** The OpenAI/Claude/Gemini `Config()` implementations decode the full `configProperties` (including `api_key`) back into a map. The Slack handle instead returns an empty map to avoid leaking `bot_token`, showing the intended pattern. Verified that `Handle.Config()` is currently consumed only internally — model executors `mapstructure.Decode` it into typed props, and it feeds the connection-cache key — with no path that serializes it into an API response, so there is no active exposure. The inconsistency remains a latent secret-leak footgun if any connector-introspection/debug surface ever returns `Handle.Config()`.
- **Failure scenario:** A future connector-analyze/debug endpoint serializes `Handle.Config()` and inadvertently returns the raw `api_key` to clients.
- **Recommendation:** Redact or omit `Secret` properties from `Config()` consistently across drivers (drive it off the `PropertySpec.Secret` flag), or return an empty map as Slack does.

**Verification:** CONFIRMED (as a latent inconsistency, low severity) — OpenAI `Config()` (openai.go:268-272), Claude `Config()` (claude.go:229-233), and Gemini `Config()` (gemini.go:254-258) all `mapstructure.Decode(h.config, &configMap)`, returning the full config including `api_key`. Slack `Config()` (slack.go:98-100) returns `map[string]any{}` to avoid leaking `bot_token`. The audit's own call-site sweep (no API serialization path for `Handle.Config()`) means there is no active leak today, so this is a latent hardening item, matching the LOW rating.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — there is currently no API/UI surface that serializes `Handle.Config()` to a client. Confirm in code: compare `runtime/drivers/openai/openai.go:268-272`, `runtime/drivers/claude/claude.go:229-233`, `runtime/drivers/gemini/gemini.go:254-258` (return full config incl. `api_key`) against `runtime/drivers/slack/slack.go:98-100` (returns empty map).

### drivers-connectors-ai-09 [LOW] `AllFromInformationSchema` caps at 1000 tables/schema and shadows the page token
- **Category:** product-improvement  **Confidence:** medium  **Location:** runtime/drivers/information_schema.go:78-99
- **Problem:** The helper lists tables with a hard-coded page size of 1000 and returns `fmt.Errorf("schema has more than 1000 tables can not list all")` if a single schema has more. The inner `ts, token, err :=` shadows the outer `token`, so the returned token is the schema-level token — correct but easy to misread. Large ClickHouse/Druid schemas silently fail to enumerate.
- **Failure scenario:** A user connects an OLAP source with >1000 tables in one schema; table listing (and anything built on `OLAP.All`) errors out with an opaque message.
- **Recommendation:** Paginate table listing within a schema (thread the inner token) instead of erroring, and rename the shadowed variable for clarity; longer term, merge `OLAPInformationSchema` and `InformationSchema` as the TODO notes.

**Verification:** CONFIRMED — `AllFromInformationSchema` (`runtime/drivers/information_schema.go:74-99`) calls `i.ListTables(ctx, schema.Database, schema.DatabaseSchema, 1000, "")` with a hard-coded page size of 1000, and `if token != "" { return nil, "", fmt.Errorf("schema has more than 1000 tables can not list all") }`. The inner `ts, token, err :=` at :78 shadows the outer `token` (from `ListDatabaseSchemas`), and the final `return tables, token, nil` returns the schema-level token — functionally correct but confusing, as the audit notes.

**Manual Reproduction Steps (for QA):**
1. Connect an OLAP source (e.g. ClickHouse or Druid) whose single database schema contains more than 1000 tables.
2. In the Rill UI, browse the connector's tables (or trigger anything that calls the OLAP `All` table listing).
3. Observe the listing fails with the opaque error "schema has more than 1000 tables can not list all" instead of paginating through the tables.

### drivers-connectors-ai-10 [LOW] Admin repo `close()` reads `configCtx` without holding the lock
- **Category:** concurrency  **Confidence:** low  **Location:** runtime/drivers/admin/repo.go:654-657
- **Problem:** `close()` reads and calls `r.configCtx`/`r.configCtxCancel` before acquiring any lock, whereas `configChanged()`/`checkHandshake` mutate those fields under the write lock. If a handshake-driven `configChanged()` races with `close()`, this is a data race on the pointer fields.
- **Failure scenario:** Handle close overlaps with an in-flight handshake that reassigns `configCtx`; under the race detector this is flagged, and in the worst case the wrong cancel func is invoked.
- **Recommendation:** Acquire the write lock (or reuse the existing read-lock section a few lines down) before touching `configCtx`/`configCtxCancel`.

**Verification:** CONFIRMED — `close()` (`runtime/drivers/admin/repo.go:654-657`) reads `r.configCtx` and calls `r.configCtxCancel()` before acquiring any lock (the `r.mu.RLock` is taken at :662, after this block). `configChanged()` (:985-990) and `checkHandshake` mutate `r.configCtx`/`r.configCtxCancel` under the write lock. A handshake-driven `configChanged()` racing with `close()` is an unsynchronized access to those pointer fields. Low severity as rated: a narrow race requiring close to overlap an in-flight handshake.

**Manual Reproduction Steps (for QA):**
Not user-reproducible via UI (requires a data race between handle close and an in-flight admin handshake). Confirm in code: `runtime/drivers/admin/repo.go:654-657` reads/invokes `r.configCtx`/`r.configCtxCancel` outside any lock, while `configChanged()` at :985-990 mutates them under the write lock. Detectable by running the admin repo tests under `go test -race` with concurrent handshake/close.

### drivers-connectors-ai-11 [LOW] Virtual repo sync writes admin-provided paths without traversal checks
- **Category:** security  **Confidence:** low  **Location:** runtime/drivers/admin/repo_virtual.go:66-84
- **Problem:** `syncInner` writes each virtual file to `filepath.Join(r.tmpDir, virtualDir, vf.Path)` and deletes via `os.Remove` on the same join, with no check that the result stays under `tmpDir`. `vf.Path` comes from the admin service (trusted), so this is defense-in-depth rather than a user-facing hole, but it mirrors finding 01.
- **Failure scenario:** A compromised or buggy admin response with `vf.Path` containing `..` writes/removes files outside the virtual tmp dir.
- **Recommendation:** Validate `vf.Path` against the virtual root (same shared helper as finding 01) before writing/removing.

**Verification:** CONFIRMED (as defense-in-depth, low severity) — `syncInner` (`runtime/drivers/admin/repo_virtual.go:65-84`) computes `path := filepath.Join(r.tmpDir, virtualDir, vf.Path)` and then `os.Remove(path)` (delete branch) or `os.MkdirAll(filepath.Dir(path), ...)` + `os.WriteFile(path, ...)` with no check that `path` stays under `r.tmpDir`. `vf.Path` comes from the admin service's `PullVirtualRepo` response (trusted), so this is defense-in-depth mirroring finding 01, matching the LOW rating.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — `vf.Path` originates from the trusted admin service, not end-user input, so there is no user-facing path to inject `..`. Confirm in code: `runtime/drivers/admin/repo_virtual.go:66-84` — `filepath.Join(r.tmpDir, virtualDir, vf.Path)` used for write/remove with no containment check against `r.tmpDir`.

## Product Improvements

- **drivers-connectors-ai-05** — Slack notifier should tolerate partial delivery failures (per-recipient isolation, retry/backoff, partial-success reporting) instead of aborting the batch.
- **drivers-connectors-ai-07** — Fix the errored-alert "Open" link to point at the dashboard rather than the editor.
- **drivers-connectors-ai-09** — Enumerate schemas with >1000 tables via real pagination and return a clearer error in the meantime.
- **AI Ping is a no-op:** `openai`/`gemini`/`claude` `Ping()` always returns nil, so "test connection" in the UI reports success even with an invalid API key. A lightweight validation call (e.g. list models) would give real feedback (see finding-style note; Slack's `Ping` already does an `AuthTest`).

## New Feature Ideas

1. **Shared repo path-sandbox helper** — one `resolveInRepo(root, path)` used by every `RepoStore` implementation (admin, file) plus table-driven traversal tests. Closes finding 01/11 and prevents regressions. *Effort: S*
2. **Consistent secret redaction for `Handle.Config()`** — drive redaction off `PropertySpec.Secret` so no driver can accidentally echo an API key/token. *Effort: S*
3. **Notifier delivery resilience layer** — a small wrapper giving all notifiers per-recipient error isolation, bounded retry/backoff on retryable provider errors, and a structured partial-success result the scheduler can act on. *Effort: M*
4. **Real AI connector health checks** — implement `Ping` for OpenAI/Gemini/Claude with a cheap authenticated call so "test connection" and instance health reflect key validity. *Effort: S*
5. **Cross-schema table pagination** — make `AllFromInformationSchema` (and the OLAP `All` implementations built on it) paginate within a schema instead of failing past 1000 tables. *Effort: M*
6. **Per-session AI token/cost budget** — extend `InstanceConfig` (which already has query-limit knobs) with a per-session input/output token budget enforced in the completion loop, surfaced to users when hit. *Effort: M*
7. **Connection-cache introspection** — an admin/debug endpoint exposing open handles per instance/driver, refs, age, and hanging state (the metrics already exist internally) to diagnose leaks like finding 02/04. *Effort: M*

## Notes & Missing Test Coverage

- **No unit tests for AI message conversion.** `openai_test.go` and `gemini_test.go` are live integration tests (Gemini's is `t.Skip()`-ed; both require a real key), and `claude` has no test file at all. The proto↔SDK translation (`messageToOpenAI`/`messageFromOpenAI`, `convertMessages`/`normalizeContents`, Claude's per-block splitting) is exactly the kind of pure, edge-case-heavy logic that should be unit-tested (empty content, tool-call/tool-result pairing, role alternation, malformed tool-call JSON).
- **No traversal tests for `RepoStore`.** `drivers/repo_test.go`'s shared `testRepo` helper only exercises happy-path `Put`/`Get` with in-root paths; add `..`/absolute-path cases once finding 01 is fixed.
- **Archive-backed admin repo paths are untested.** `admin/repo_git_test.go` covers the Git flows well, but the `r.git == nil` (archive/virtual-only) branches that finding 03 concerns are not exercised.
- **Gemini `MaxOutputTokens` has no default** (gemini.go:290 passes `int32(0)` when unset, unlike the OpenAI/Claude drivers which default to 8192) — worth confirming Gemini treats 0 as "model default" and not an error; harmless if so, but inconsistent.
- **`generateKey`** joins config values with `:`/` ` via `fmt.Sprint`, so pathological values containing those delimiters could in principle collide across distinct configs; low risk, but a hash of a canonical encoding would be safer.
- The `conncache` package itself is solid and well-tested for the tricky lifecycle transitions; findings 02 and 04 are in the *runtime* code that drives it, not in the cache primitive.

## Verification Notes

An independent adversarial re-verification pass opened every cited file and traced call sites, error paths, and existing guards. All 11 original findings were confirmed reachable/valid in code and retained (removed_count = 0); the surviving set is accurate, but three had inaccuracies corrected and several were tightened with the verification evidence:

- **Finding 01 (path traversal):** Confirmed there is *no* traversal/containment check anywhere in the chain (`grep` for `filepath.IsLocal`/`filepath.Rel`/prefix checks across `server/repos.go`, `repos.go`, `drivers/repo.go`, and both repo drivers returned nothing), and `IsIgnored` (drivers/repo.go:124) only matches a fixed prefix list. The gRPC handlers gate on `ReadRepo`/`EditRepo` claims and pass `req.Path` straight through. Corrected the file-driver line reference `98 → 93` (line 98 is the `continue` inside `Hash`, not the `filepath.Join`). Severity kept HIGH: the missing check is certain; the cross-tenant blast radius is real on shared multi-instance cloud runtimes, while on the local `file` driver the "escape" only reaches the developer's own filesystem.
- **Finding 03 (nil-deref on archive-backed):** Corrected an overstatement. `Root`, `Hash`, `CommitHash`, and `CommitTimestamp` are in fact *safe* on archive-backed deployments because each checks `if r.archive != nil` before falling back to `r.git.root()`; they would only nil-deref in the practically-unreachable "virtual-only" state. The genuinely reachable defect is `Put`/`MkdirAll`/`Rename`/`Delete`/`Watch`, which call `root := r.git.root()` after a guard that permits `r.git == nil`. `Watch` reachability confirmed via reconcilers/project_parser.go:192 (gated on `rill.watch_repo`). Confidence kept medium: the code defect is definite, but it requires an archive-backed instance that is also editable/watched.
- **Finding 04 (stale handle not evicted):** Strengthened. Verified `edit` (registry.go:375-396) only restarts the controller — and thus only reaches `evictInstanceConnections` (registry.go:497) — when `restartController` is true, and `UpdateInstanceConnector` passes `false`. The eviction site's own comment ("this is the only place where we can do it safely") confirms eviction-on-edit is intended, so the upsert path is a real gap. Reachable via reconcilers/connector.go:90.
- **Finding 06 (Gemini nil-deref):** Confirmed `UsageMetadata` is a nilable pointer in the pinned SDK (`genai@v1.44.0/types.go:2680`). Added a reachability caveat: the real endpoint reliably populates it, so this is a defensive-nil-check hardening item, not a routinely-hit crash. Confidence held at medium.
- **Finding 07 (Slack ERROR link):** Confirmed by reading the `alert_status.slack` template — it renders `OpenLink` as "Open in browser" and `EditLink` as the edit link separately, so the ERROR branch (notifier.go:137) makes both point at the editor. Re-categorized from product-improvement to **bug** (it is a concrete wrong-output defect).
- **Finding 08 (AI Config() leaks api_key):** Confirmed all three AI handles decode the full config (incl. `api_key`) while Slack returns `map[string]any{}`. Verified via call-site sweep that `Handle.Config()` is consumed only internally (typed `mapstructure.Decode` in model executors and the cache key) with no API serialization path, so this stays a low/low latent-hardening item, not an active leak.
- **Findings 02, 05, 09, 10, 11:** Re-read and confirmed as originally described; no changes. (02: the handle from `drivers.Open` is a local var never returned/closed when `ctx.Err()` overrides `err` at connection_cache.go:170-172 — a definite orphan. 05: every Slack fan-out loop returns on the first error and `SendScheduledReport` sequences channels→webhooks→users. 09: hard-coded `1000` page size with an error past it; the inner `token` at information_schema.go:78 does shadow the outer, and the returned token is the schema-level one — functionally correct. 10: `close()` reads/invokes `configCtx`/`configCtxCancel` at repo.go:655-657 outside the lock while `configChanged()` mutates them under the write lock. 11: `syncInner` joins admin-provided `vf.Path` with no containment check — trusted input, so defense-in-depth/low.)

No high-signal issues beyond the original set survived scrutiny. Candidates considered and dropped: Claude's per-block message splitting producing consecutive same-role messages (cannot confirm the Anthropic Beta API rejects this without runtime evidence — speculative); `generateKey` delimiter collisions (already noted, negligible risk); and an empty-`roots()` `Get` returning `("", nil)` (unreachable because `r.virtual` is always initialized on the first handshake).
