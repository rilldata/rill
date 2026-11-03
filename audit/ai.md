# AI Assistant / Agents — Code Review

**Scope:** `runtime/ai/*.go` (router/analyst/developer/feedback agents, orchestration in `ai.go`, tool implementations: file_read/write/search/list, query_sql, metrics_view_*, list_buckets/objects, list_tables, show_table, create_chart, navigate, canvas_get, project_status, develop_file), `runtime/ai/instructions/*`, `mcp.go`, `util.go`. Cross-referenced: `runtime/server/chat.go`, `runtime/server/mcp.go`, `runtime/repos.go`, `runtime/drivers/file/repo.go`, `runtime/drivers/admin/repo.go`, `runtime/resolvers/sql.go`, `runtime/metricsview/executor/*`, `runtime/drivers/claude/claude.go`, `runtime/drivers/sqlite/catalog.go`.  |  **Health:** needs-attention

## Overview
This package implements a well-structured, session-based agent framework: a `Runner` registers typed tools, a `Session` records a call-tree of messages, and `Complete` drives an LLM tool-call loop shared by chat, scheduled reports, and the MCP server. Access control is generally applied consistently (per-tool `CheckAccess` + `ApplySecurityPolicy` on metrics resources — verified in `metrics_view_list.go`). Tool outputs are deliberately bounded (row caps, snippet limits, truncation). The biggest risks are on the developer-tool surface, where LLM-generated inputs reach the filesystem and OLAP connectors: file-path arguments are not constrained to the project root (path traversal), and `query_sql` runs arbitrary statements while advertising itself as read-only. There is also a real concurrency defect in the streaming path where the message-subscriber send can strand the producer under the session lock on context cancellation, leaking the request goroutine. Secondary issues involve limit validation, a latent fail-open branch in shared-session access, error/stack-trace handling that leaks into the LLM context, and a latent message-persistence hazard in `Flush`.

## Findings

### ai-01 [HIGH] File tools allow path traversal outside the project root
- **Category:** security  **Confidence:** high  **Location:** `runtime/ai/file_read.go:52`, `runtime/ai/file_write.go:61`, `runtime/ai/develop_file.go:59`
- **Problem:** `ReadFile`/`WriteFile`/`DevelopFile` take an LLM-controlled `args.Path`, prepend `/` if missing, and pass it straight to `Runtime.GetFile/PutFile/DeleteFile`. Those call `repo.Get/Put/Delete`, which in both the local driver (`drivers/file/repo.go:74,128,178`) and the cloud driver (`drivers/admin/repo.go:150,241,…`) do `filepath.Join(root, path)` with **no containment check**. `filepath.Join` cleans `..` segments, so the join escapes `root`. The leading-`/` normalization does not help — verified empirically: `filepath.Join("/srv/rill/projects/abc", "/../../../../etc/passwd")` = `/etc/passwd`, and `"/models/../../../../../etc/passwd"` also resolves to `/etc/passwd`. `runtime/repos.go` performs no validation before delegating to the driver.
- **Failure scenario:** A prompt-injection payload embedded in project data, a bucket object name, a table/column comment, or a shared/forked conversation steers the developer agent to `read_file` `../../.env`, cloud-credential files, or another project's files on the runtime host — then exfiltrate the contents in its reply. `write_file`/`develop_file` with the same traversal can create or overwrite files outside the project where the repo is editable (always in Rill Developer; editable cloud repos). In Rill Cloud this crosses the project sandbox on a shared runtime host. Preconditions: `UseAI` + `EditRepo` claims and the `developer_agent` feature flag; these tools are `internal` (not exposed to external MCP clients), but the path argument is still model-chosen inside the chat agent.
- **Recommendation:** Constrain paths in the AI tools before hitting the repo: reject inputs containing `..` segments and verify the cleaned relative path is local (e.g. `filepath.IsLocal` on the leading-slash-stripped form, or `strings.HasPrefix(filepath.Join(root, clean), root+string(os.PathSeparator))`). Add the same guard in the repo drivers as defense-in-depth, since these APIs now receive untrusted, model-generated paths.
- **Note:** `file_search.go` is **not** a traversal vector (removed from the prior draft's location list): it lists files via `ListGlob`/`os.DirFS`, which rejects `..`, and only reads paths the walk returns (all within root).

**Verification:** CONFIRMED — `file_read.go:52-56`, `file_write.go:61-85`, and `develop_file.go:59-61` prepend a leading `/` and pass the model-chosen path straight to `Runtime.GetFile/PutFile/DeleteFile`. Those (`runtime/repos.go:21-87`) delegate to `repo.Get/Put/Delete` with no validation, and the local driver does `filepath.Join(c.root, filePath)` with no containment check (`drivers/file/repo.go:74,113,128,178`); the admin driver is identical (`drivers/admin/repo.go:150,221,257`). `filepath.Join` cleans `..`, so a leading-`/` path like `/../../etc/passwd` escapes root.

**Manual Reproduction Steps (for QA):**
1. In Rill Developer, open a project and start an AI chat with the developer agent enabled (requires the `developer_agent` feature flag, `UseAI` and `EditRepo`).
2. Ask the agent to "read the file at `../../../../etc/passwd`" (or on the write side, "create a file at `../../escaped.txt`").
3. Observe the agent returns contents of a file outside the project root, or creates a file outside the project directory.
4. If UI phrasing resists, confirm in code: `runtime/ai/file_read.go:52` (no `..` rejection before `GetFile`) and `runtime/drivers/file/repo.go:74` (`filepath.Join(c.root, filePath)` with no prefix check).

### ai-02 [HIGH] Subscriber send under the session lock can strand the producer on cancellation (goroutine leak / hang)
- **Category:** concurrency  **Confidence:** high  **Location:** `runtime/ai/ai.go:898-900` (`AddMessage`), unbuffered channel at `runtime/ai/ai.go:656-661` (`Subscribe`), consumed at `runtime/server/chat.go:376-407`
- **Problem:** `AddMessage` holds `s.mu.Lock()` and does a blocking send `sub <- msg` on an **unbuffered** channel (`Subscribe` → `make(chan *Message)`) with no `select` on any cancellation signal. In `CompleteStreaming`, the only reader is a goroutine whose `select` also has a `case <-ctx.Done(): return`. When the context is canceled (client stop, disconnect, or the completion timeout) while a message is in flight, Go may pick the `ctx.Done()` case over the `sub` receive; the reader returns and the producer's `sub <- msg` never rendezvous, blocking permanently while holding `s.mu`.
- **Failure scenario:** During a streaming completion the client disconnects. The agent loop has just received an LLM response and calls `AddMessage` (e.g. `ai.go:1356` for a text block, or `ai.go:964` for a tool-call message) → blocks forever on the send. `session.CallTool`/`Complete` never returns, so `CompleteStreaming` never returns, the deferred `Unsubscribe`/`Flush` never run, the session is never persisted, and the request goroutine (holding `s.mu`) is leaked for the process lifetime. `Unsubscribe` itself would also block trying to acquire `s.mu`. Repeated cancellations accumulate leaked goroutines and permanently wedge each affected session object.
- **Recommendation:** Never block on a channel send while holding a lock, and make the send cancellation-aware. Have the reader close a `done` channel on exit and use `select { case sub <- msg: case <-done: }`, or use buffered channels with a non-blocking send, or notify subscribers outside the critical section. Ensure `Unsubscribe` can always unblock the producer.

**Verification:** CONFIRMED — `Subscribe` creates an unbuffered `make(chan *Message)` (`ai.go:657`); `AddMessage` holds `s.mu.Lock()` (`ai.go:893-894`) and does a blocking bare `sub <- msg` with no `select` (`ai.go:898-900`). The only reader is the `CompleteStreaming` goroutine whose `select` also has `case <-ctx.Done(): return` (`chat.go:386-388`), so on cancellation the reader can exit while the producer is stranded on the send, holding `s.mu` forever. `Unsubscribe` (`ai.go:664-669`) then blocks on `s.mu` too.

**Manual Reproduction Steps (for QA):**
1. Start an AI chat completion that streams (e.g. a longer analyst/developer prompt that produces several tool calls and text blocks).
2. While the response is still streaming, abruptly disconnect the client (close the browser tab, kill the network, or cancel the request).
3. Repeat several times. On a server with goroutine profiling (`/debug/pprof/goroutine`), observe goroutines accumulating and never exiting, blocked in `AddMessage` on the channel send.
4. Confirm in code: `runtime/ai/ai.go:898-900` (blocking send under lock) vs `runtime/server/chat.go:387` (reader's `ctx.Done()` case that can win the select).

### ai-03 [MEDIUM] `query_sql` executes arbitrary (non-read-only) SQL despite `ReadOnlyHint`
- **Category:** security  **Confidence:** high  **Location:** `runtime/ai/query_sql.go:37-47` (annotation), `query_sql.go:92-100` (resolve); resolver at `runtime/resolvers/sql.go:162-166`
- **Problem:** The tool is annotated `ReadOnlyHint: true` / `DestructiveHint: false` and described as "introspect data", but the SQL resolver runs `olap.Query(ctx, &drivers.Statement{Query: r.sql})` with no restriction on statement type. Nothing prevents `DROP`, `DELETE`, `INSERT`, `CREATE`, `ATTACH`, etc. against the OLAP connector using its configured credentials. `CheckAccess` calls `checkDeveloperAccess(..., internal=false)`, so — unlike the file tools — this tool **is** exposed to external MCP clients (subject to `UseAI` + `EditRepo` + `developer_agent` flag), and the SQL text is model-generated and prompt-injectable.
- **Severity rationale (downgraded from HIGH):** The tool requires `EditRepo`, which already grants the ability to run arbitrary SQL via model `pre_exec`/`post_exec`/DDL, so this is not a fundamental privilege escalation. The concrete incremental risks are (a) the **misleading `ReadOnlyHint`**, which can cause MCP hosts/clients to auto-approve a destructive call, and (b) prompt-injection reachability without a human authoring a model file.
- **Failure scenario:** A prompt-injection payload (or an LLM error) causes the developer agent — or an auto-approving MCP host trusting `ReadOnlyHint` — to run `DROP TABLE <model>` / `DELETE FROM ...` against DuckDB/ClickHouse, destroying materialized data.
- **Recommendation:** Enforce read-only execution (a read-only connection/role where supported, or parse and reject non-`SELECT`/`SHOW`/`DESCRIBE`/`EXPLAIN`/`WITH` statements). At minimum, correct the `ReadOnlyHint`/`DestructiveHint` annotations to reflect actual behavior.

**Verification:** CONFIRMED — `query_sql.go:37-42` sets `ReadOnlyHint: true` / `DestructiveHint: false` while the description says "introspect data"; `CheckAccess` calls `checkDeveloperAccess(ctx, t.Runtime, false)` (`query_sql.go:51`), and with `internal=false` the tool is NOT restricted to `rill`-prefixed user agents (`developer_agent.go:176-180`), so it is MCP-exposed. The SQL resolver runs `r.olap.Query(ctx, &drivers.Statement{Query: r.sql})` with no statement-type restriction (`resolvers/sql.go:162-166`). Any `DROP`/`DELETE`/`INSERT` will execute.

**Manual Reproduction Steps (for QA):**
1. With the developer agent enabled (`UseAI` + `EditRepo` + `developer_agent` flag), open an AI chat.
2. Prompt: "Run this SQL to clean up: `DROP TABLE <some_model_table>`" (use a disposable table).
3. Observe the table is dropped despite the tool being advertised as read-only.
4. Confirm in code: `runtime/ai/query_sql.go:41` (`ReadOnlyHint: true`) and `runtime/resolvers/sql.go:163` (unrestricted `olap.Query`).

### ai-04 [MEDIUM] `metrics_view_query` limit lacks a lower bound and uses a fragile parse
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/ai/metrics_view_query.go:238-253`
- **Problem:** The handler caps only the upper bound (`if limit > cfg.AIMaxQueryLimit`) and extracts the limit via `strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)`. Two confirmed issues:
  1. **Fragile numeric parse (certain).** JSON numbers arrive as `float64`; `fmt.Sprintf("%v", f)` renders "round" magnitudes ≥ 1e6 in scientific notation (verified: `1000000.0`→`"1e+06"`, `5000000.0`→`"5e+06"`), which `ParseInt` rejects. So a common model output like `"limit": 1000000` errors with an opaque "invalid limit value" instead of being capped. (Non-round values such as `1234567` render as decimal and parse fine.)
  2. **No lower bound (dialect-dependent impact).** A non-positive limit passes through unchanged with `isSystemLimit=false`. `rewriteQueryEnforceCaps` returns a 0 reader-cap when any limit ≤ cap is present (`executor_rewrite_enforce_caps.go:28-33`), and the builder renders the value verbatim (`astsql.go:236-238` → `LIMIT -1`). On DuckDB (the default) `LIMIT -1` is a query error; on lenient dialects it can mean "unlimited", streaming the full result set into `data [][]any`. `LIMIT 0` silently returns zero rows with no warning.
- **Failure scenario:** The model emits `"limit": 1000000` and gets an opaque parse error; or emits `"limit": -1`/`0` and gets a dialect error or (on a lenient dialect) an unbounded materialization that bypasses `AIMaxQueryLimit`.
- **Recommendation:** Decode the limit numerically (`json.Number` / type-switch on `float64`/`int64`) rather than round-tripping through `%v`, and clamp to `[1, AIMaxQueryLimit]`, treating `<= 0` as the default.

**Verification:** CONFIRMED — `metrics_view_query.go:241` does `strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)`, which fails on `float64` round magnitudes ≥ 1e6 rendered in scientific notation (`fmt.Sprintf("%v", 1000000.0)` = `"1e+06"`), erroring with the opaque "invalid limit value" (`:243`). Only the upper bound is clamped (`if limit > cfg.AIMaxQueryLimit`, `:245`); no lower-bound check, so a non-positive limit flows through with `isSystemLimit=false`.

**Manual Reproduction Steps (for QA):**
1. In an AI chat, ask a question that makes the analyst request exactly one million rows, or craft an MCP `query_metrics_view` call with `"limit": 1000000`.
2. Observe the query fails with "invalid limit value" instead of being capped to `AIMaxQueryLimit`.
3. For the lower-bound path, issue a `query_metrics_view` with `"limit": -1` or `"limit": 0` and observe a dialect error (DuckDB `LIMIT -1`) or zero rows.
4. Confirm in code: `runtime/ai/metrics_view_query.go:241` (`%v`+ParseInt) and `:245` (only upper bound clamped).

### ai-05 [LOW] Shared-session access control fails open if the boundary message is ever missing
- **Category:** security  **Confidence:** medium  **Location:** `runtime/ai/ai.go:123-152`
- **Problem:** For a non-owner viewing a shared session, the loader appends every message and only stops when it reaches `SharedUntilMessageID` (`if m.ID == retrieveUntilMessageID { break }`). If that sentinel ID is not present in the returned message set, the loop runs to completion and **all** messages — including content past the shared boundary — are returned to the non-owner. This fails open rather than closed.
- **Currently not reachable (severity downgraded from MEDIUM):** The boundary is validated at share time — `ShareConversation` resolves it via `session.LatestMessage(...)` against an actual message (`chat.go:134-143`) — and messages are append-only (no `DeleteAIMessage` path exists), so `SharedUntilMessageID` always matches a loaded message today. This is defensive hardening, not a live disclosure.
- **Failure scenario:** Any future invariant break (message pruning/rewrites, a stale ID after a schema change) would silently disclose the entire private conversation to any non-owner holding a share link.
- **Recommendation:** Resolve the boundary to an `Index` up front and filter `Index <= boundaryIndex`; if the sentinel is not found among loaded messages, deny access (fail closed) instead of returning everything.

**Verification:** PLAUSIBLE — the fail-open code path is real: `ai.go:135-152` appends every loaded message and only `break`s when `m.ID == retrieveUntilMessageID` (`:149`); if that sentinel is absent, the loop runs to completion and returns all messages to the non-owner. As the finding states, it is not currently reachable — the boundary is validated at share time and messages are append-only (no `DeleteAIMessage` in the sqlite catalog), so the sentinel always matches a loaded message today.

**Manual Reproduction Steps (for QA):** Not user-reproducible with current code (no way to make the sentinel go missing). Confirm in code: `runtime/ai/ai.go:149` — the loop `break`s only on ID match with no fallback; if `retrieveUntilMessageID` is not found, every message (past the shared boundary) is returned to a non-owner viewer of a shared session.

### ai-06 [LOW] Stack traces are embedded in errors surfaced to the LLM and persisted
- **Category:** tech-debt  **Confidence:** high  **Location:** `runtime/ai/ai.go:1334`, `runtime/ai/ai.go:990-996`
- **Problem:** `Complete` wraps completion failures with `... (stack: %s)` using `debug.Stack()`, and the panic recovery in `Call` builds an error containing a 64 KB goroutine stack. For tool-call errors these become `MessageContentTypeError` result messages (`Call` lines 1030-1041), which are serialized back into the LLM context on the next turn and written to the catalog/activity log (`Flush`).
- **Failure scenario:** A transient LLM error or a tool panic injects a multi-kilobyte Go stack trace into the conversation history that is re-sent to the model (token waste and prompt noise) and stored; internal paths/symbols leak into user-visible errors.
- **Recommendation:** Log stacks via the logger only; keep the error returned to the loop/user concise (no `debug.Stack()`), and truncate panic detail before it becomes a result message.

**Verification:** CONFIRMED — `ai.go:1334` wraps completion failures with `... (stack: %s)` using `string(debug.Stack())`, and the panic recovery in `Call` builds `NewInternalError(fmt.Errorf("panic caught: %v\n\n%s", err, string(stack)))` from a 64 KB `goruntime.Stack` buffer (`ai.go:991-995`). Tool-call errors become `MessageContentTypeError` result messages (`ai.go:1031-1041`) that are re-serialized into the LLM context and persisted (`Flush` records error content verbatim, `ai.go:584-585`).

**Manual Reproduction Steps (for QA):** Hard to trigger deliberately via UI (requires a tool panic or transient LLM error). Confirm in code: `runtime/ai/ai.go:1334` (`debug.Stack()` in the returned error string) and `runtime/ai/ai.go:991-995` (64 KB panic stack becomes an error message). Note the error content is stored and re-sent to the model, not redacted (`ai.go:584-585`).

### ai-07 [LOW] Final assistant response keeps only the last text block
- **Category:** correctness  **Confidence:** low  **Location:** `runtime/ai/ai.go:1402-1414`
- **Problem:** When assembling the final result, the loop assigns `outVal = block.Text` per block, so a response with multiple text blocks keeps only the last one. This is reachable: the Claude driver maps `thinking` blocks to `ContentBlock_Text` (`drivers/claude/claude.go:452-455`), so a final message with extended thinking arrives as `[Text(thinking), Text(answer)]`. In the common case the dropped block is the leading thinking block (benign), but a genuine answer split across two text blocks would lose its first fragment.
- **Refuted sub-claim:** The prior draft also claimed a non-text block would hit `default` and hard-fail the completion. This is **not reachable** — `aiv1.ContentBlock` has only `Text`/`ToolCall`/`ToolResult`, the driver converts thinking→text and drops unknown blocks, and a final (no-tool-call) message therefore contains only `Text` blocks. The `default` branch is effectively dead here.
- **Recommendation:** Concatenate text blocks for string/`outputText` results rather than overwriting.

**Verification:** CONFIRMED — `ai.go:1402-1414` loops over `result.Content` and assigns `outVal = block.Text` (`:1407`) per text block, so only the last block survives. The `default` branch at `:1411` is effectively dead for a final (no-tool-call) message, as the finding notes. Whether the dropped block is a benign leading thinking block or a genuine answer fragment depends on how the driver splits blocks.

**Manual Reproduction Steps (for QA):** Not reliably user-reproducible (depends on the model emitting multiple text blocks in its final message). Confirm in code: `runtime/ai/ai.go:1407` — `outVal = block.Text` overwrites on each iteration instead of concatenating, so a multi-text-block final answer keeps only the last fragment.

### ai-08 [LOW] `promptToTitle` truncates on a byte boundary
- **Category:** correctness  **Confidence:** high  **Location:** `runtime/ai/router_agent.go:238-239`
- **Problem:** `title[:47]` slices by bytes; a multi-byte UTF-8 rune straddling byte 47 is split, producing an invalid string stored as the conversation title and shown in the UI / sent over the API.
- **Recommendation:** Truncate by runes (`[]rune(title)[:47]`) or use `utf8`-aware truncation.

**Verification:** CONFIRMED — `router_agent.go:238-239`: `if len(title) > 50 { title = title[:47] + "..." }` slices by bytes. A multi-byte UTF-8 rune straddling byte 47 is split into an invalid string.

**Manual Reproduction Steps (for QA):**
1. Start a new AI conversation whose first message is a long prompt (>50 characters) containing multi-byte characters (e.g. emoji or CJK text) positioned so one falls across byte 47, e.g. many emoji: "😀😀😀😀😀😀😀😀😀😀😀😀😀please analyze my sales data trends over time".
2. Observe the auto-generated conversation title in the UI shows a broken/replacement character (�) where the rune was split.
3. Confirm in code: `runtime/ai/router_agent.go:239` (`title[:47]` byte slice).

### ai-09 [LOW] Inconsistent session locking: read accessors and `Flush` bypass the mutex
- **Category:** concurrency  **Confidence:** medium  **Location:** `runtime/ai/ai.go:686-742` (`NextIndex`, `Message`, `LatestMessage`, `Messages`), `runtime/ai/ai.go:535-599` (`Flush`), `runtime/ai/ai.go:884` (`Index` computed before the lock)
- **Problem:** `AddMessage` mutates `s.messages`/`s.messagesDirty` under `s.mu.Lock()`, but the read accessors iterate `s.messages` with no lock, `Flush` both reads and writes `s.dtoDirty`/`s.messagesDirty` and iterates `s.messages` unlocked, and `AddMessage` even computes `Index: s.NextIndex()` (a read of `len(s.messages)`) before acquiring the lock. This is not a live race today: a session is driven by a single goroutine, and the streaming reader only touches per-message content and the (read-only) tools map via `UnmarshalMessageContent` (verified — it does not read `s.messages`). It is a latent data race the moment any code reads/flushes messages from a second goroutine (the natural evolution of the subscriber/streaming path).
- **Recommendation:** Take `s.mu.RLock()`/`Lock()` uniformly in the accessors and `Flush` (or snapshot under lock), so the locking discipline is consistent and future concurrent readers are safe.

**Verification:** CONFIRMED (latent) — `AddMessage` mutates `s.messages`/`s.messagesDirty` under `s.mu.Lock()` (`ai.go:893-897`), but `NextIndex`/`Message`/`LatestMessage`/`Messages` iterate `s.messages` with no lock (`ai.go:686-742`), `Flush` reads/writes `s.dtoDirty`/`s.messagesDirty` and iterates `s.messages` unlocked (`ai.go:535-599`), and `AddMessage` computes `Index: s.NextIndex()` (a `len(s.messages)` read) at `ai.go:884` before acquiring the lock at `:893`. Not a live race today (single driving goroutine), but a data race the moment a second goroutine reads/flushes.

**Manual Reproduction Steps (for QA):** Not user-reproducible (no live race in current single-goroutine flow). Confirm in code: `runtime/ai/ai.go:686-742` (lock-free accessors), `ai.go:535-599` (lock-free `Flush`), and `ai.go:884` (`NextIndex()` read before `s.mu.Lock()` at `:893`). Run `go test -race ./runtime/ai/...` after adding any concurrent reader to surface it.

### ai-10 [LOW] Dead code: `Session.RootID()`
- **Category:** tech-debt  **Confidence:** high  **Location:** `runtime/ai/ai.go:905-921`
- **Problem:** `RootID()` has no callers anywhere in the repo and contains two `panic` paths (empty parent at line 908, missing message at line 913; the `for range 100` bound is a non-panicking fail-safe that returns the last-seen root). Dead code carrying panic risk.
- **Recommendation:** Remove it, or if it is intended for upcoming durable-execution work, add a test and replace the panics with returned errors.

**Verification:** CONFIRMED — `RootID()` at `ai.go:905-921` has two `panic` paths (`:908` empty parent, `:913` missing message); the `for range 100` is a non-panicking fail-safe. A repo-wide search finds no callers.

**Manual Reproduction Steps (for QA):** Not user-reproducible (dead code, never invoked). Confirm in code: `runtime/ai/ai.go:905` (`func (s *Session) RootID()`), and `grep -rn "RootID" runtime/` returns only the definition, no callers.

### ai-13 [LOW] `msg.dirty` is never reset, so re-flushing a session would duplicate-insert messages
- **Category:** correctness  **Confidence:** low  **Location:** `runtime/ai/ai.go:890` (set), `runtime/ai/ai.go:541,564,598` (`Flush`), sink at `runtime/drivers/sqlite/catalog.go:551-559`
- **Problem:** `AddMessage` sets `msg.dirty = true` and nothing ever resets it. `Flush` gates the message loop on `s.messagesDirty` (reset to `false` after a flush) but never clears the per-message `dirty` flag. `InsertAIMessage` is a plain `INSERT` on primary key `id`. Consequently, if the **same in-memory `Session` object** is ever flushed a second time after new messages are added, every previously-persisted message (still `dirty=true`) is re-inserted and the flush fails with a `UNIQUE constraint` error.
- **Currently latent:** every caller today creates a fresh session and flushes once — the streaming/non-streaming chat handlers use a deferred single `Flush`, and the MCP handler is `Stateless: true` (`server/mcp.go:59`), reloading messages clean (`dirty=false`) on each request. So the double-flush precondition is not met in current code.
- **Failure scenario:** A future change that reuses a session across flushes (stateful MCP, periodic flushing during long streams, or an added mid-turn `Flush`) would immediately break persistence with duplicate-key errors.
- **Recommendation:** Set `msg.dirty = false` after a successful `InsertAIMessage` (and rely on it rather than the coarse `s.messagesDirty`), or make `InsertAIMessage` an idempotent upsert.

**Verification:** CONFIRMED (latent) — `AddMessage` sets `dirty: true` (`ai.go:890`) and nothing resets the per-message flag; `Flush` gates on `s.messagesDirty` and resets only that (`ai.go:562,598`), never clearing `msg.dirty` (`ai.go:564`). `InsertAIMessage` is a plain `INSERT` on PK `id` (`sqlite/catalog.go:551-555`), so a second flush of a reused session would re-insert all still-`dirty` messages and fail with a UNIQUE constraint. Latent because all callers flush once per fresh session and MCP is `Stateless: true` (`server/mcp.go:59`).

**Manual Reproduction Steps (for QA):** Not user-reproducible with current code (no session is flushed twice). Confirm in code: `runtime/ai/ai.go:890` (set `dirty=true`, never reset), `ai.go:562-598` (`Flush` clears only `s.messagesDirty`), `runtime/drivers/sqlite/catalog.go:552` (plain `INSERT`). A stateful-MCP or mid-stream re-flush change would trigger duplicate-key errors.

## Product Improvements

### ai-11 [MEDIUM] No aggregate cost/iteration budget across nested agents
- **Category:** product-improvement  **Confidence:** high  **Location:** `runtime/ai/ai.go:1139-1150` (`Complete`), `developer_agent.go:102` (20), `analyst_agent.go:198` (20), `develop_file.go:131` (10)
- **Problem:** `MaxIterations` is per-loop only. Router → Developer (20 iterations) → each iteration may call `develop_file` (its own 10-iteration loop) → each of those may call `write_file`/`query_sql`. A single user turn can fan out to hundreds of LLM calls with no session- or turn-level token/call budget; the only backstop is the wall-clock `AICompletionTimeoutSeconds`. On the Rill-managed (billable) connector this can produce surprising token spend from one prompt.
- **Recommendation:** Track cumulative LLM calls / tokens on the session and enforce a per-turn budget that halts the loop with a graceful "budget exhausted" message, independent of wall-clock time.

**Verification:** CONFIRMED — `MaxIterations` is enforced per `Complete` loop only. Developer agent uses 20 (`developer_agent.go:102`), analyst 20 (`analyst_agent.go:198`), `develop_file` 10 (`develop_file.go:131`), and the developer loop's tools include `DevelopFileName` (`developer_agent.go:99`), which spawns its own nested loop. There is no session- or turn-level token/call budget; the only cross-loop backstop is the wall-clock `AICompletionTimeoutSeconds`.

**Manual Reproduction Steps (for QA):** Not a defect to "reproduce" — it is a missing guardrail. Confirm in code: `runtime/ai/develop_file.go:131` (nested 10-iteration loop) invoked from `runtime/ai/developer_agent.go:99,102` (outer 20-iteration loop), with no aggregate counter anywhere in `Complete` (`ai.go` completion loop).

### ai-12 [MEDIUM] Analyst context seeding keyed on call count goes stale after dashboard navigation
- **Category:** product-improvement  **Confidence:** medium  **Location:** `runtime/ai/analyst_agent.go:105-161`
- **Problem:** The agent pre-invokes dashboard context (`get_canvas`, `query_metrics_view_summary`, `get_metrics_view`) only when `first := len(analyst calls) == 1`. The inline TODO (line 109) already flags this: when the user navigates to a different explore/canvas between turns, the second turn's new metrics view and summary are never seeded, so the analysis relies on the model re-discovering context (the tools are available, but the priming is skipped).
- **Recommendation:** Seed based on whether the current `Explore`/`Canvas`/metrics-view set has already been introduced in the session (dedupe by resource name), not on the raw call count.

**Verification:** CONFIRMED — `analyst_agent.go:106` computes `first := len(s.Messages(FilterByType(MessageTypeCall), FilterByTool(AnalystAgentName))) == 1`, and all dashboard-context pre-invocation (`get_canvas`, `query_metrics_view_summary`, `get_metrics_view`, `list_metrics_views`) is gated on `if first` (`:111-161`). The inline TODO at `:109` explicitly flags that this is unsafe when the user navigates to a different dashboard between turns.

**Manual Reproduction Steps (for QA):**
1. Open an explore/canvas dashboard A and ask the analyst agent a question (first turn seeds context for A).
2. Navigate to a different dashboard B, then ask a second question referencing B in the same conversation.
3. Observe the agent does not pre-load B's metrics view / summary (it must re-discover context via tools, if at all).
4. Confirm in code: `runtime/ai/analyst_agent.go:106` (`first` keyed on call count) and the TODO at `:109`.

## New Feature Ideas
- **Sandboxed, read-only data-introspection role (S/M):** A dedicated read-only execution path for `query_sql`/`list_tables`/`show_table` (separate connection or role) so introspection cannot mutate data even under prompt injection. Directly addresses ai-03 and hardens the whole developer surface.
- **Project-root-jailed filesystem tool layer (S):** A single `projectPath(instanceID, userPath) (string, error)` helper enforcing containment, reused by every AI file tool (and ideally the repo drivers). Closes ai-01 and prevents regressions as new file tools are added.
- **Per-conversation token/cost meter + budget (M):** Surface cumulative input/output/cached tokens (already recorded as metrics) back to the user in the chat UI, with a configurable per-turn ceiling. Pairs with ai-11 and gives admins cost predictability.
- **Tool-call audit trail with approval gates for destructive tools (M):** For `write_file`/`develop_file`/`query_sql`, optionally require explicit user confirmation (the MCP annotations already carry `DestructiveHint`) before executing, with a visible diff/statement preview. Mitigates prompt-injection blast radius.
- **AI-assisted conversation summarization for truncation (M/L):** Replace the fixed keep-first-4/last-16 window in `maybeTruncateMessages` (a self-described placeholder) with rolling summarization so long developer sessions retain earlier decisions instead of dropping them.
- **Structured citations as first-class objects (S):** `query_metrics_view` returns `open_url` strings the model must inline; promoting these to structured citation objects attached to messages would make citation rendering reliable and decouple it from prompt compliance.

## Notes & Missing Test Coverage
- **No tests exercise the security-critical branches.** The test harness (`ai_test.go:newSession`) always sets `SkipChecks: true` and a `rill-evals` user agent, so `CheckAccess` gating, the `internal` MCP restriction, shared-session boundaries (`SharedUntilMessageID`), and `ForkSession` cloning are effectively untested. Add unit tests that run with `SkipChecks:false` and non-`rill` user agents.
- **No path-traversal tests** for `read_file`/`write_file`/`develop_file` (ai-01). Add cases asserting that `..`-containing paths are rejected and cannot escape the project root.
- **No tests for `maybeTruncateMessages`** (ai.go:1605) — the unbalanced tool-call/result pruning is subtle (relies on call-before-result ordering and only inspects the first tool block per message). It appears correct for the current one-block-per-message shape, but table-driven tests covering orphaned calls/results at the truncation boundary would lock it in.
- **No concurrency/streaming test** for the subscriber path (ai-02); a test that cancels the context mid-stream and asserts the producer unblocks and the goroutine exits would lock in the fix.
- **`list_buckets.go:70-82`** collects all pages with no overall cap — benign for normal accounts but unbounded in principle.

## Verification Notes
Every finding was re-checked against the cited code; the empirical Go behaviors (path traversal via `filepath.Join`, `fmt.Sprintf("%v", float64)` scientific notation) were reproduced with standalone snippets.

Changes made during verification:
- **ai-01:** Confirmed (HIGH/high). Removed `file_search.go` from the location — it goes through `ListGlob`/`os.DirFS`, which rejects `..`, so it is not a traversal vector. Added the precondition detail (these file tools are `internal`, i.e. not on external MCP, but the path is still model-chosen inside the chat agent).
- **ai-02:** Confirmed (HIGH/high). Tightened the mechanism: the hang arises from the `select` race between the unbuffered `sub <- msg` and the reader's `ctx.Done()` case (both can be ready; the reader may win). Corrected line anchors.
- **ai-03:** Confirmed behavior; **severity downgraded HIGH → MEDIUM** (confidence raised to high on the facts). `EditRepo` already permits arbitrary SQL via models, so the real incremental risks are the misleading `ReadOnlyHint` (auto-approval) and prompt-injection reachability. Verified `internal=false` (MCP-exposed) via `checkDeveloperAccess`.
- **ai-04:** Confirmed. Separated the two sub-issues by confidence: the `%v` scientific-notation parse failure is certain (reproduced); the unbounded-memory path is dialect-dependent (DuckDB errors on `LIMIT -1`), so it stays medium.
- **ai-05:** Confirmed as a fail-open code path but **not currently reachable** (boundary validated at share time via `LatestMessage`; messages are append-only). **Severity downgraded MEDIUM → LOW**; reframed as defensive hardening.
- **ai-06:** Confirmed (LOW). Category changed `readability → tech-debt` (error-handling hygiene + minor info leak + token waste). Traced the panic-stack path into `MessageContentTypeError` result messages.
- **ai-07:** **Partially refuted.** The "non-text/thinking block → hard error" scenario is unreachable (`aiv1.ContentBlock` has no thinking type; the Claude driver maps thinking→`Text` and drops unknown blocks; final messages are Text-only). Rewrote to keep only the accurate residual ("keeps only the last text block"), which is usually benign. Stays LOW/low.
- **ai-08:** Confirmed (LOW/high). Corrected line anchor to `router_agent.go:238-239`.
- **ai-09:** Confirmed as latent (not live) — verified the streaming reader does not touch `s.messages`. **Expanded** to include `Flush` (which reads/writes dirty flags and iterates messages unlocked) and the pre-lock `Index` computation.
- **ai-10:** Confirmed dead code (no callers). Corrected "three panic paths" → **two** (the loop bound is a non-panicking fail-safe).
- **ai-11 / ai-12:** Confirmed; verified line anchors (analyst `MaxIterations` at `analyst_agent.go:198`).
- **ai-13 (added):** New, verified finding — `msg.dirty` is set once and never reset, so re-flushing a reused session object would duplicate-insert via a plain `INSERT`. Currently latent because all callers flush once per fresh session (MCP is `Stateless`). LOW/low.

No findings were removed outright (0). Net: 12 original findings retained (2 downgraded in severity, 1 partially refuted/rewritten, 1 recategorized) plus 1 added.
