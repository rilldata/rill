# Runtime Shared Utilities — Code Review
**Scope:** `runtime/pkg/{conncache, singleflight, priorityqueue, schedule, ratelimit, retrier, dag, graceful, logbuffer, bufferutil, filewatcher, blob, rduckdb, catalog, duckdbsql, rilltime, activity, securetoken, pagination, globutil, pathutil, fileutil}` (plus call-site checks in `runtime/server`, `runtime/pkg/logutil`, `runtime/drivers/file`, `runtime/drivers/duckdb`)  |  **Health:** needs-attention

## Overview
This area holds the low-level concurrency primitives, path/glob helpers, SQL-AST manipulation, and storage plumbing shared across the runtime. The core concurrency utilities are, on the whole, carefully written and well-tested: `conncache`, `singleflight`, `priorityqueue`/`semaphore`, `schedule`, and `dag` all hold up to close tracing, and `rduckdb`'s catalog uses disciplined reference counting. The biggest risks are concentrated in a few spots: (1) `logbuffer.GetLogs` panics with an index-out-of-range on the very common `limit == 0` request because the proto validation permits it and there is no panic-recovery interceptor in the runtime gRPC chain; (2) the same file's log fan-out performs a blocking channel send while holding its mutex, which can wedge or deadlock an instance's entire logging pipeline behind one slow log-streaming client; and (3) `fileutil.ResolveLocalPath`, the sandbox that is supposed to keep the local `file`/DuckDB connectors inside the project root in hosted deployments, is bypassable via absolute paths and sibling-directory prefixes. Secondary resource-lifecycle gaps include an unbounded OpenTelemetry callback leak in `rduckdb` write handles and a timeout-less HTTP client in the telemetry sink. The SQL-AST rewrite layer builds JSON by unescaped string interpolation, which breaks table rewriting for identifiers containing special characters. Most utilities are simple and readable; the exceptions (`blob.ListObjectsForGlob`, `rduckdb/db.go`, `rilltime`) are large and would benefit from more targeted tests.

## Findings

### runtime-pkg-01 [HIGH] `ResolveLocalPath` sandbox is bypassable (path traversal)

**Status (2026-07-25):** ✅ Fixed and merged in [#9726](https://github.com/rilldata/rill/pull/9726).

- **Category:** security  **Confidence:** high  **Location:** runtime/pkg/fileutil/fileutil.go:317-333
- **Problem:** The guard is `if !allowHostAccess && !strings.HasPrefix(finalPath, root)`. Two defects: (a) for absolute inputs the code sets `finalPath = path` **without `filepath.Clean`** (the `filepath.Join`, which would clean, only runs for relative paths, lines 324-326), so `..` segments in an absolute path survive; (b) the prefix check has no trailing path separator, so a sibling directory whose name starts with `root` passes. Note the *relative* `..` case is actually caught because `filepath.Join` cleans it, so the two live bypasses are absolute-path `..` and the sibling-prefix collision.
- **Failure scenario:** In hosted deployments `AllowHostAccess=false` is used as the file/DuckDB sandbox (confirmed at `drivers/duckdb/duckdb.go:550`, "Hack: Using AllowHostAccess as a proxy indicator for a hosted environment"). A source/model with `path: /repo/../../etc/passwd` (absolute) is not cleaned → `HasPrefix("/repo/../../etc/passwd", "/repo")` is true → returned verbatim → the OS resolves it to `/etc/passwd`. Alternatively, with root `/data/proj123`, `path: ../proj123-x/secret` cleans to `/data/proj123-x/secret`, which still passes the separator-less prefix check — a cross-tenant read if project directories are siblings. Callers (`drivers/file/file_store.go:36`, `drivers/duckdb/model_executor_self.go:298,540`) then glob/attach the resolved path, giving a tenant arbitrary-file read on the runtime host.
- **Recommendation:** Compute `finalPath = filepath.Clean(finalPath)` for both branches, then check `finalPath == root || strings.HasPrefix(finalPath, root+string(filepath.Separator))`. Consider a shared `SecureJoin`-style helper reused by all path-resolving code.

**Verification:** CONFIRMED — `runtime/pkg/fileutil/fileutil.go:317-333`. The absolute branch (line 324) sets `finalPath = path` verbatim with no `filepath.Clean`, and the guard at line 328 is `!strings.HasPrefix(finalPath, root)` with no trailing separator. Both defects (absolute-path `..` surviving, and sibling-directory prefix collision such as `/data/proj123` matching `/data/proj123-x`) are present exactly as described. Confirmed the hosted sandbox uses `AllowHostAccess=false` at `runtime/drivers/duckdb/duckdb.go` (the "Hack: Using AllowHostAccess as a proxy indicator for a hosted environment" comment) and that `runtime/drivers/file/file_store.go` resolves via this function.

**Manual Reproduction Steps (for QA):**
1. Configure a `duckdb` or local `file` connector with `allow_host_access: false` (this is the state hosted Rill Cloud runs the file/DuckDB sandbox in).
2. Create a source/model whose `path` is an absolute path containing `..`, e.g. `path: /<project_root>/../../etc/hosts`.
3. Reconcile/ingest the source; observe the file resolves outside the project root and is read, instead of being rejected with "path is outside repo root".
4. Alternatively, with project root `/data/proj123`, point `path` at `../proj123-x/<file>` and observe it resolves to a sibling tenant directory that still passes the prefix check.
5. To confirm in code / add a regression test: extend `runtime/pkg/fileutil/fileutil_test.go` with cases `ResolveLocalPath("/root/../etc/x", "/root", false)` and `ResolveLocalPath("../root-x/secret", "/root", false)`; both currently return a path instead of an error.

### runtime-pkg-13 [HIGH] `logbuffer.GetLogs` panics with index-out-of-range when `limit == 0`
- **Category:** bug  **Confidence:** high  **Location:** runtime/pkg/logbuffer/buffer.go:93 (guard at :77)
- **Problem:** `GetLogs` only special-cases `limit < 0` ("all"). For `limit == 0` it computes `limit = min(0, Count()) = 0`, allocates `logs := make([]*Log, 0)`, then in the `ReverseIterateUntil` callback executes `logs[limit-1] = item.Value`, i.e. `logs[-1]`, which is a runtime panic ("index out of range [-1]") as soon as the callback reaches any entry at or above `minLvl`. The empty-buffer case does not iterate, so the panic requires a non-empty buffer — the normal state of any active instance.
- **Failure scenario:** `limit == 0` is reachable, not theoretical: the proto rule is `int32 limit = 3 [(validate.rules).int32 = {gte: -1}]` (`proto/rill/runtime/v1/api.proto:810`), so `0` (also the default for an unset field) passes `grpc_validator`. `Server.GetLogs` (`runtime/server/instances.go:238`) forwards `int(req.Limit)` straight through; `WatchLogs` with `Replay=true, ReplayLimit=0` hits the same path (`instances.go:266`, `api.proto:821`). The runtime gRPC interceptor chain (`runtime/server/server.go:158-165`) has no `grpc_recovery`, so any authorized caller (a project viewer in Cloud) sending `limit=0` panics the handler goroutine and crashes the runtime process — a denial of service.
- **Recommendation:** Treat `limit == 0` as "return zero logs" (early-return an empty slice), or clamp/validate at the boundary. Separately, add a panic-recovery interceptor to the runtime gRPC chain so a single handler bug cannot take down the process, and add a `logbuffer` test covering `limit` of `0`, `-1`, and values larger than the buffer.

**Verification:** CONFIRMED — `runtime/pkg/logbuffer/buffer.go:73-96`. With `limit == 0` and a non-empty buffer, line 80 sets `limit = min(0, Count()) = 0`, line 82 allocates a zero-length slice, and the `ReverseIterateUntil` callback executes `logs[limit-1]` = `logs[-1]` at line 93 for the first entry whose level meets `minLvl` — a hard `index out of range [-1]` panic. Proto permits `0`: `int32 limit = 3 [(validate.rules).int32 = {gte: -1}]` (`proto/rill/runtime/v1/api.proto:810`) and `replay_limit` likewise (:821). `runtime/server/instances.go:238` forwards `int(req.Limit)` unchanged, and the `Replay` path (:266) forwards `int(req.ReplayLimit)`. The gRPC interceptor chain (`runtime/server/server.go:158-166`) has no `grpc_recovery`, so the panic crashes the handler goroutine/process.

**Manual Reproduction Steps (for QA):**
1. Open a Rill project with an active instance that has emitted at least one INFO-or-above log line (the normal state).
2. Call the runtime API `GetLogs` (or `WatchLogs` with `replay=true`) with `limit` (resp. `replay_limit`) explicitly set to `0` — e.g. `grpcurl -d '{"instance_id":"<id>","limit":0}' <runtime-host> rill.runtime.v1.RuntimeService/GetLogs`, or send the equivalent REST request with `limit=0`. Note: an unset `limit` also defaults to `0`.
3. Observe the runtime handler panics with "index out of range [-1]" and the process crashes (no recovery interceptor), rather than returning an empty log list.
4. To confirm in code / add a regression test: `logbuffer` has no test file; add a test that pushes one log then calls `GetLogs(true, 0, LOG_LEVEL_INFO)` and expect it not to panic.

### runtime-pkg-02 [HIGH] `logbuffer.AddEntry` blocks (and can deadlock) on a slow log subscriber
- **Category:** concurrency  **Confidence:** high  **Location:** runtime/pkg/logbuffer/buffer.go:38-49
- **Problem:** `AddEntry` holds `b.mu` and does a blocking send `client <- message` on an **unbuffered** channel (created at buffer.go:53) for every subscriber (line 47). The subscriber goroutine (`WatchLogs`, buffer.go:52) drains that channel and, in the runtime's real callback, calls `srv.Send(...)` — a gRPC network write (`runtime/server/instances.go:274-279`). `AddEntry` is the sink for the buffered zap core (`runtime/pkg/logutil/handler.go:45`), i.e. every instance log line.
- **Failure scenario:** A `WatchLogs` client reads slowly or its TCP send stalls → `srv.Send` blocks → the watcher never returns to receive → the next `AddEntry` blocks on the send **while holding `b.mu`** → all further logging for that instance (and `GetLogs`, which takes `RLock`) is wedged. A hard deadlock is also reachable: if the watcher's `ctx` is cancelled while `AddEntry` is mid-send, `WatchLogs` may take the `ctx.Done()` branch of its `select` and return; its deferred `removeClient` (buffer.go:118) then blocks on `b.mu`, which `AddEntry` still holds while blocked on a send no one will receive.
- **Recommendation:** Do not hold the lock during delivery. Give each subscriber a buffered channel and use a non-blocking send (`select { case ch <- msg: default: /* drop + count */ }`), or hand off to a per-subscriber goroutine that owns the network send. This also removes the `panic("client closed!")` landmine at buffer.go:61.

**Verification:** CONFIRMED — `runtime/pkg/logbuffer/buffer.go:38-49`. `AddEntry` acquires `b.mu` (line 38) and, still holding it, does a blocking `client <- message` (line 47) over channels created unbuffered at line 53 (`make(chan *runtimev1.Log)`). The subscriber's `WatchLogs` callback ultimately calls `srv.Send` (`runtime/server/instances.go:274-279`), a network write. A stalled subscriber therefore blocks `AddEntry` while it holds the mutex, wedging all further logging and any `GetLogs` (RLock). The deadlock path via `removeClient` (line 118-123, taking `b.mu`) racing a mid-send `AddEntry` is also real.

**Manual Reproduction Steps (for QA):**
1. Start a Rill instance and open a `WatchLogs` stream (e.g. `rill project logs -f` in Cloud, or a direct `WatchLogs` gRPC call) from a client that reads slowly or whose TCP receive window is stalled (e.g. pause the reader / suspend the process).
2. Generate continuous log activity on that instance (trigger reconciles, run queries) so `AddEntry` is called repeatedly.
3. Observe that once the stalled subscriber's channel fills (immediately, since unbuffered), all logging for that instance stops progressing and other log operations (further `GetLogs`) hang — the mutex is held during the blocked send.
4. For the deadlock variant, cancel the slow watcher's context while `AddEntry` is mid-send; `removeClient`'s deferred `b.mu.Lock()` blocks against the still-held lock.
5. To confirm in code / add a concurrency test: a `logbuffer` test with a subscriber that never drains its channel plus a concurrent `AddEntry` will hang (demonstrates the block under `-timeout`).

### runtime-pkg-04 [MEDIUM] `rduckdb` leaks an OpenTelemetry callback per write handle
- **Category:** performance  **Confidence:** high  **Location:** runtime/pkg/rduckdb/db.go:956 (also generic.go:189)
- **Problem:** `openDBAndAttach` calls `otelsql.RegisterDBStatsMetrics(db.DB, …)`. In `otelsql@v0.27.0` that function registers a meter callback via `meter.RegisterCallback` and **discards the returned `metric.Registration` internally** (`sql.go:120`, `_, err = meter.RegisterCallback(...)`); it returns only `error`. The callback closure captures the `*sql.DB`. Write handles are opened and closed on **every** write (`acquireWriteConn` → `openDBAndAttach(read=false)`, from `CreateTableAsSelect`, `MutateTable`, `RenameTable`, …), and the callback is never removed. (The read handle at db.go:284 is opened once and is not the concern.)
- **Failure scenario:** A project doing periodic model refreshes accumulates one live callback (with its instruments, each holding a now-closed DB) per write for the process lifetime. Memory grows without bound, and every metrics scrape iterates all stale callbacks calling `Stats()` on closed DBs, so collection cost grows linearly with total writes.
- **Recommendation:** Because `RegisterDBStatsMetrics` does not hand back the `Registration`, the caller cannot `Unregister()` it — so register DB-stats metrics **only for the long-lived read `dbHandle`** and skip it for ephemeral write handles. Alternatively, upgrade to an `otelsql` version that returns the registration (or scope a fresh `MeterProvider` per handle that can be shut down) and cancel it when the handle closes.

**Verification:** CONFIRMED — `runtime/pkg/rduckdb/db.go:955` (inside `openDBAndAttach`, line 880) calls `otelsql.RegisterDBStatsMetrics`, and `openDBAndAttach(read=false)` is invoked from `acquireWriteConn` (db.go:780), which runs on every write (`CreateTableAsSelect` db.go:410, `MutateTable`/rename db.go:523). The read handle is opened once (db.go:284). Verified in the vendored dependency `github.com/XSAM/otelsql@v0.27.0/sql.go:120`: `_, err = meter.RegisterCallback(...)` discards the returned `metric.Registration`, so it can never be unregistered; the callback closure captures the (soon-closed) `*sql.DB`. The identical pattern is at `runtime/pkg/rduckdb/generic.go:189`. `go.mod` pins `otelsql v0.27.0`.

**Manual Reproduction Steps (for QA):**
1. Run a Rill project with a model that refreshes on a schedule (or trigger many manual model re-materializations) so write handles are opened/closed repeatedly.
2. Let it run over many refresh cycles while an OTel meter provider is configured (metrics enabled).
3. Observe process RSS growing monotonically and metric-scrape latency increasing over time, with one retained meter callback (each calling `Stats()` on a closed DB) accumulating per write.
4. To confirm in code: inspect `runtime/pkg/rduckdb/db.go:955` vs the single read-handle open at db.go:284, and the discarded registration at `otelsql@v0.27.0/sql.go:120`.

### runtime-pkg-03 [MEDIUM] Telemetry intake sink uses `http.DefaultClient` with no timeout; `Close` can hang
- **Category:** concurrency  **Confidence:** high  **Location:** runtime/pkg/activity/sink_intake.go:127 (and Close at :57-68)
- **Problem:** Each `flush` spawns a goroutine that calls `s.send`, which uses `http.DefaultClient.Do` with a request built by `http.NewRequest` (no context, no timeout). `Close` closes `stop`, does a final flush, then `s.wg.Wait()`.
- **Failure scenario:** If the intake endpoint accepts the connection but never responds, `send` blocks indefinitely; `Close` then blocks forever on `wg.Wait()`, stalling process shutdown. Under sustained backpressure, each buffer-full `Emit` (sink_intake.go:50-52) launches another stuck goroutine, so goroutines and buffers accumulate. Requires a misbehaving/black-holed endpoint, but a well-behaved shutdown should not be at its mercy.
- **Recommendation:** Use a dedicated `*http.Client` with a request timeout (and ideally a bounded number of in-flight flushes). Give `Close` a ctx-bounded wait so shutdown cannot hang on a wedged HTTP call.

**Verification:** CONFIRMED — `runtime/pkg/activity/sink_intake.go:121-131`. `send` builds the request with `http.NewRequest` (no context) and executes it via `http.DefaultClient.Do` (line 127), which has no timeout. `flush` (lines 90-107) spawns a goroutine per flush tracked by `s.wg`, and `Close` (lines 57-68) ends with an unbounded `s.wg.Wait()`. A black-holed intake endpoint makes `send` block forever, so `Close` hangs shutdown; sustained backpressure launches more stuck goroutines from `Emit` (lines 50-52).

**Manual Reproduction Steps (for QA):**
1. Configure the runtime/CLI telemetry intake URL to point at an endpoint that accepts TCP connections but never sends an HTTP response (e.g. a netcat listener or a firewalled host).
2. Generate telemetry events so the buffer flushes.
3. Stop/shut down the process (which calls `intakeSink.Close`).
4. Observe shutdown hangs indefinitely on `wg.Wait()` waiting for the wedged HTTP request, instead of exiting.
5. To confirm in code: `sink_intake.go:127` uses `http.DefaultClient.Do` with a context-less request and no client timeout.

### runtime-pkg-06 [MEDIUM] Blob download writes to `filepath.Join(tempDir, obj.Path)` without containment (path traversal)
- **Category:** security  **Confidence:** medium  **Location:** runtime/pkg/blob/download.go:230-242
- **Problem:** The download target is `filepath.Join(it.tempDir, obj.Path)` where `obj.Path` is the raw object-store key (`obj.Key`, set verbatim at `blob/bucket.go:219,287`). Object keys can legitimately contain `..` segments; `filepath.Join` cleans them, so a key such as `a/../../../tmp/x` resolves outside `tempDir`. `os.MkdirAll(filepath.Dir(...))` + `os.OpenFile(..., O_CREATE|O_TRUNC)` then write there.
- **Failure scenario:** Ingesting from a bucket whose keys an attacker can influence (a shared/public bucket, or the tenant's own bucket on the multi-tenant cloud runtime) lets the download write/overwrite files outside the temp dir on the runtime host. Reachability is gated on controlling source-bucket keys, so this is a hardening gap rather than a turnkey exploit, but it matters most on a shared runtime.
- **Recommendation:** After joining, verify the result stays within `tempDir` (clean + `HasPrefix(tempDir+sep)`), or reject keys containing `..` path segments before writing.

**Verification:** PLAUSIBLE — `runtime/pkg/blob/download.go:230` computes `filename := filepath.Join(it.tempDir, obj.Path)` and lines 231/242 `os.MkdirAll`/`os.OpenFile(..., O_CREATE|O_TRUNC)` write there, with no containment check on `obj.Path` (the raw object key). `filepath.Join` cleans `..` segments, so a key like `a/../../../tmp/x` escapes `tempDir`. Real but conditional: reachability requires an object store whose keys contain `..` and are attacker-influenced (shared/public bucket, or a tenant's own bucket on a shared runtime), so it is a hardening gap rather than a turnkey exploit.

**Manual Reproduction Steps (for QA):**
1. On a runtime with a source ingesting from an object store (S3/GCS/local blob), create/upload an object whose key contains `..` traversal segments (e.g. `data/../../../tmp/pwned.txt`), if the store permits such keys.
2. Configure a source with a glob that matches that key and reconcile/ingest it.
3. Observe the runtime writes the downloaded content outside the `blob_ingestion` temp dir (e.g. `/tmp/pwned.txt`), overwriting host files.
4. To confirm in code / add a test: `runtime/pkg/blob/download.go:230`; a unit test with a fake bucket returning an entry `Path: "../escape.txt"` should show the written file lands outside `tempDir`.

### runtime-pkg-05 [LOW] `duckdbsql` builds AST JSON via unescaped `fmt.Sprintf`
- **Category:** correctness  **Confidence:** medium  **Location:** runtime/pkg/duckdbsql/ast_rewrite.go:100-108 (also :251-262, :314-327, :436-452)
- **Problem:** Table names, aliases, column names, and string values are interpolated straight into JSON string literals with `%s`/`%v` and then `json.Unmarshal`ed. `createBaseTable`, `createTableFunction`, `createKeyedFunctionArg`, and `createFunctionCall` do no JSON escaping; `createGenericValue` escapes strings with Go's `%q`, which is close to but not identical to JSON escaping (and `%v` on a float `NaN`/`Inf` produces invalid JSON).
- **Failure scenario:** A model whose FROM clause aliases a table as `"a""b"` (alias `a"b`) or uses an identifier containing a quote/backslash yields malformed JSON → `json.Unmarshal` fails → `rewriteToBaseTable` returns an error and table-reference rewriting fails for an otherwise-valid query. Note the values here originate from DuckDB's own parsed-AST serialization (not free-form attacker text), so the practical impact is a confusing rewrite failure on legitimate special-character identifiers rather than key injection; impact is contained since the user is rewriting their own query.
- **Recommendation:** Construct these nodes as Go maps/structs and `json.Marshal` them, rather than interpolating into JSON text. At minimum, `json.Marshal` each interpolated value.

**Verification:** CONFIRMED — `runtime/pkg/duckdbsql/ast_rewrite.go:97-116` (`createBaseTable`) and :118-130 (`createTableFunction`) interpolate `alias`/`table_name` straight into a JSON string literal via `fmt.Sprintf("...%s...", toString(ast, astKeyAlias), name)` and then `json.Unmarshal`, with no JSON escaping. An alias/identifier containing a `"` or `\` produces malformed JSON and the unmarshal fails, so `rewriteToBaseTable` errors. As the review notes, values originate from DuckDB's own AST serialization (not free-form attacker text), so impact is a contained rewrite failure on legitimate special-character identifiers — hence LOW.

**Manual Reproduction Steps (for QA):**
1. Create a model/source whose SQL FROM clause aliases a table with an identifier containing a double quote, e.g. `SELECT * FROM t AS "a""b"` (alias `a"b`), where the model path triggers table-reference rewriting (source rewrite to a base table).
2. Reconcile the model.
3. Observe the rewrite fails (JSON unmarshal error) for an otherwise valid query, instead of rewriting successfully.
4. To confirm in code / add a test: `runtime/pkg/duckdbsql` — feed a query with a quote-containing alias through the rewrite path and observe `createBaseTable` returns a JSON error.

### runtime-pkg-07 [LOW] `filewatcher` writes `w.err` without synchronization (data race)
- **Category:** concurrency  **Confidence:** high  **Location:** runtime/pkg/filewatcher/filewatcher.go:92-94 vs :102
- **Problem:** `closeWithErr` deliberately avoids `w.mu` (comment at :86) and writes `w.err` (lines 92-94), while `Subscribe` reads `w.err` under `w.mu` (line 102). The `w.closed` atomic guards only the once-only close, not `w.err`, and the two accesses share no common synchronization. (The post-`<-w.done` read at :121 is safe via the `close(w.done)` happens-before; the :102 read is not.)
- **Failure scenario:** `Close()`/watcher failure runs concurrently with a `Subscribe` call; `-race` flags the write/read of `w.err`, and in principle `Subscribe` can observe a torn/stale interface value.
- **Recommendation:** Store the error in an `atomic.Pointer[error]` (or guard it consistently with `w.mu`) so both writer and readers use the same synchronization.

**Verification:** CONFIRMED — `runtime/pkg/filewatcher/filewatcher.go:84-98` (`closeWithErr`) deliberately avoids `w.mu` (comment at :85-86) and writes `w.err` at lines 92-94, while `Subscribe` reads `w.err` under `w.mu` at line 102. The `w.closed` atomic (line 87) guards only once-only close, not the `w.err` field; the two accesses share no common synchronization. The later read at :121 is safe via the `close(w.done)` happens-before, but the :102 read is not, so `-race` will flag the write/read pair.

**Manual Reproduction Steps (for QA):**
1. Not user-reproducible through the product (internal race, timing-dependent). Confirm in code: `runtime/pkg/filewatcher/filewatcher.go:92-94` (unsynchronized `w.err` write) versus `:102` (`w.err` read under `w.mu`).
2. To demonstrate: run the filewatcher tests with the race detector, `go test -race ./runtime/pkg/filewatcher/`, exercising a `Subscribe` call concurrent with `Close()`/a watcher failure; the detector reports the `w.err` data race.

### runtime-pkg-09 [LOW] `blob.Download` materializes the entire object listing in memory before downloading
- **Category:** performance  **Confidence:** medium  **Location:** runtime/pkg/blob/download.go:76-83
- **Problem:** `pagination.CollectAll` (`pagination/page_token.go:49`) is used to eagerly fetch **all** matching `ObjectStoreEntry`s into a slice before any download starts; `CollectAll` appends every page with no cap.
- **Failure scenario:** A glob matching millions of objects (large partitioned lakes) forces the full listing into memory up front, spiking RSS before the first batch is produced, independent of `BatchSizeBytes`.
- **Recommendation:** Stream listing pages into the download pipeline instead of collecting all entries first, or enforce a configurable maximum object count with a clear error.

**Verification:** CONFIRMED — `runtime/pkg/blob/download.go:76-83`. `Download` calls `pagination.CollectAll(...)` to eagerly gather every matching `ObjectStoreEntry` into `entries` before the download/batch goroutines start (lines 86-102). `pagination.CollectAll` (`runtime/pkg/pagination/page_token.go`) appends every page with no cap, so a glob matching a very large number of objects materializes the full listing in memory up front, independent of `BatchSizeBytes`.

**Manual Reproduction Steps (for QA):**
1. Point a source at an object store with a glob matching a very large number of objects (e.g. a partitioned lake with hundreds of thousands/millions of files).
2. Reconcile/ingest the source and monitor the runtime's memory.
3. Observe RSS spikes at the start (during listing) before the first batch is downloaded, scaling with object count rather than batch size.
4. To confirm in code: `runtime/pkg/blob/download.go:76` (`pagination.CollectAll`) collects all entries before any download begins.

### runtime-pkg-10 [LOW] Overly permissive file modes (`os.ModePerm` / 0777) for downloaded and copied files
- **Category:** security  **Confidence:** medium  **Location:** runtime/pkg/blob/download.go:231,242; runtime/pkg/fileutil/fileutil.go:67,102
- **Problem:** Downloaded data files/dirs and embedded-dir copies are created with `os.ModePerm` (0777) rather than restrictive modes. In practice the process umask (usually 022) strips this to 0755, so the effective concern is **world-readable** ingested/materialized data, not world-writable — but the requested mode is still needlessly broad.
- **Failure scenario:** On a shared/multi-tenant host, other local users (or co-located processes) can read ingested data and materialized files.
- **Recommendation:** Use 0600 for files and 0700 for directories (matching `OpenTempFileInDir`, which already uses 0o750), unless a broader mode is specifically required.

**Verification:** CONFIRMED — `runtime/pkg/blob/download.go:231` (`os.MkdirAll(..., os.ModePerm)`) and :242 (`os.OpenFile(..., os.ModePerm)`) create downloaded dirs/files with 0777; `runtime/pkg/fileutil/fileutil.go:67` (`os.MkdirAll(dst, os.ModePerm)`) and :102 (`os.WriteFile(entryDst, data, os.ModePerm)`) do the same for embedded-dir copies. As noted, a typical 022 umask reduces the effective mode to 0755, so the accurate concern is world-*readable* ingested/materialized data on a shared host, not world-writable — hence the medium confidence.

**Manual Reproduction Steps (for QA):**
1. On a shared/multi-tenant host (default umask 022), ingest data from any source so files are downloaded/materialized via the blob downloader.
2. Inspect the resulting files/directories: `ls -l` shows mode `0755`/`-rwxr-xr-x` (world-readable), rather than owner-only `0600`/`0700`.
3. To confirm in code: `runtime/pkg/blob/download.go:231,242` and `runtime/pkg/fileutil/fileutil.go:67,102` pass `os.ModePerm` (0777).

## Product Improvements

### runtime-pkg-11 [product-improvement] Log streaming silently loses events with no signal to the client
- **Category:** product-improvement  **Confidence:** medium  **Location:** runtime/pkg/logbuffer/buffer.go:46-49
- **Problem:** Once the fan-out in finding 02 is made non-blocking, a slow `WatchLogs`/`rill project logs` client will drop entries with no indication. Today the buffer is also silently bounded (`bufferutil.BoundedCircularBuffer` evicts oldest on overflow), so replayed logs can be incomplete without any marker.
- **Recommendation:** Emit an explicit "N log lines dropped" marker (per subscriber and for buffer eviction) so users viewing logs know the stream is not complete.

**Verification:** CONFIRMED — `runtime/pkg/logbuffer/buffer.go:46-49` fans out to subscribers with no drop signal, and the backing `bufferutil.BoundedCircularBuffer` (`runtime/pkg/logbuffer/buffer.go:17,41-44`) evicts oldest entries on overflow with no marker, so replayed logs can be silently incomplete. This is a product-improvement observation about existing behavior, not a defect.

**Manual Reproduction Steps (for QA):**
1. Generate more log lines on an instance than the buffer's `maxMessageCount`/`maxBufferSize`, then run `WatchLogs`/`rill project logs` with replay.
2. Observe the replayed logs simply start partway through, with no indication that earlier lines were evicted/dropped.
3. To confirm in code: no drop counter exists at `runtime/pkg/logbuffer/buffer.go:46-49` nor around the circular-buffer eviction.

### runtime-pkg-12 [product-improvement] Blob ingestion tuning is hardcoded and unobservable
- **Category:** product-improvement  **Confidence:** medium  **Location:** runtime/pkg/blob/download.go:26,29
- **Problem:** `_concurrentBlobDownloadLimit = 8` and `_defaultBatchSizeBytes = 1 GB` are compile-time constants. Users ingesting from high-latency or high-throughput stores cannot tune parallelism/batch size, and there is no surfaced progress/throughput for long ingests.
- **Recommendation:** Expose download concurrency and batch size as connector/source options, and surface per-ingest progress (files done / total, bytes, throughput) that the existing metrics already compute.

**Verification:** CONFIRMED — `runtime/pkg/blob/download.go:26` (`const _concurrentBlobDownloadLimit = 8`) and :29 (`const _defaultBatchSizeBytes = 1 GB`) are compile-time constants with no connector/source override, and there is no user-facing per-ingest progress surface. This is a product-improvement observation, not a defect.

**Manual Reproduction Steps (for QA):**
1. Configure a source ingesting from a high-latency object store and attempt to tune download parallelism or batch size via connector/source YAML.
2. Observe there is no such option; parallelism is fixed at 8 and batch size at 1 GB.
3. To confirm in code: `runtime/pkg/blob/download.go:26,29` are `const` with no config binding.

## New Feature Ideas
- **Shared `SecureJoin` path helper (S).** A single containment-checked join used by `fileutil.ResolveLocalPath`, the blob downloader, and any other path resolution, eliminating the class of bugs in findings 01 and 06. Backs the sandbox with tests for `..`, absolute inputs, symlinks, and sibling prefixes.
- **Bounded pub/sub primitive (M).** Extract the fan-out pattern shared by `logbuffer` and `filewatcher` into one package with per-subscriber buffered channels, non-blocking delivery, and a drop counter — so no producer can be blocked by a slow subscriber. Removes the concurrency footguns in both.
- **`otelsql` handle-stats lifecycle policy (S).** Since `RegisterDBStatsMetrics` (v0.27.0) does not return a cancellable `Registration`, adopt a policy of registering DB-stats only on long-lived handles (or upgrade the dependency / scope a per-handle `MeterProvider` that can be shut down), applied everywhere short-lived DB handles are opened (finding 04).
- **rilltime fuzz + property tests (M).** A `go test -fuzz` harness over the rilltime grammar plus round-trip/eval invariants would guard the large, hand-written evaluator against regressions and pin down invariants like "an ordinal interval always carries ≥1 ordinal" (see verification notes on the former finding 08).
- **Structured AST builders for `duckdbsql` (M).** Replace the `fmt.Sprintf`-into-JSON node constructors with typed Go structs marshalled via `encoding/json`, removing the escaping hazard in finding 05 and making the rewrite layer far easier to read and extend.
- **Fail-open rate-limit observability (S).** `ratelimit.isServerConnError` intentionally allows requests through during Redis outages (`ratelimiter.go:47-57,125-155`); add a metric/log so operators can see when rate limiting is effectively disabled.

## Notes & Missing Test Coverage
- **`logbuffer` has no test file at all.** Both the `limit == 0` panic (finding 13) and the blocking-fan-out/deadlock (finding 02) are exactly what unit/concurrency tests would expose (edge-case limits; slow subscriber + concurrent `AddEntry` + subscriber ctx cancellation).
- **`fileutil.ResolveLocalPath`** has tests (`fileutil_test.go`) but they do not cover absolute-path `..` escapes or sibling-directory prefixes; add cases asserting rejection for both.
- **`conncache` is solid and well-tested.** One dead-defensive spot: `periodicallyCheckHangingConnections` calls `e.handle.Driver()` on a closing entry (conncache.go:449) assuming `e.handle != nil` without stating why; traced to be unreachable with a nil handle, so not filed.
- **`rduckdb/db.go` and `catalog.go`** carry the most intricate logic (write/meta semaphores, snapshot/version reference counting). The reference counting looks internally consistent and `localDirty` is consistently guarded by `writeSem`; no concrete defect found, but the surface is large and would benefit from targeted concurrency tests around snapshot acquisition racing with `DropTable`/`RenameTable`.
- **`singleflight`** correctly reads `c.val`/`c.err` outside the lock only after `c.ctx.Done()` (happens-before established via `cancel`); note the shared call context takes its `Value(...)` lookups from the *first* caller only — fine for identical work, but a subtle gotcha if per-caller context values ever matter.
- **`blob.ListObjectsForGlob`** is complex pagination/glob-filtering logic; it appears correct but is under-tested for the `startAfter`/`end`/directory-accumulation edge cases and is the most likely place for a latent off-by-one.
- **`priorityqueue.Semaphore`, `schedule`, `graceful`, `ratelimit`, `pathutil`, `globutil`, `securetoken`** were re-read during verification and look correct.

## Verification Notes
An adversarial re-verification traced every original finding to the cited code and call sites. Summary of changes:

- **Removed 1 finding (former `runtime-pkg-08`, "rilltime previous-period panic on empty ordinals").** Confirmed unreachable: `OrdinalInterval.Ordinals` is populated only by the participle grammar `@@ (Of @@)*` (rilltime.go:163), whose leading mandatory `@@` guarantees ≥1 ordinal whenever `i.Ordinal != nil`, and there is no programmatic construction of an `OrdinalInterval` with an empty slice anywhere in non-test code. So `i.Ordinal.Ordinals[0]` (rilltime.go:470) cannot index into an empty slice today; the sibling guard in `OrdinalInterval.eval` (rilltime.go:535) is dead-defensive. Adding the same guard to `previousPeriod` for symmetry is a harmless nicety but not a live bug, so it is not filed. The general robustness angle is covered by the rilltime fuzz-testing feature idea.
- **Added 1 finding (`runtime-pkg-13`, HIGH):** `logbuffer.GetLogs` index-out-of-range panic on `limit == 0`. Verified end to end: proto allows `gte: -1` (api.proto:810/821), `Server.GetLogs`/`WatchLogs` forward the value unchanged (instances.go:238,266), the buffer iterates a non-empty buffer and hits `logs[-1]` (buffer.go:93), and the runtime gRPC chain has no recovery interceptor (server.go:158-165) — a reachable DoS the original review missed.
- **Corrected `runtime-pkg-04`'s recommendation:** the original advice to "capture the `metric.Registration` and `Unregister()` it" is infeasible with `otelsql@v0.27.0`, whose `RegisterDBStatsMetrics` discards the registration internally and returns only `error` (verified in the vendored `sql.go:111-120`). The viable fix is to register DB-stats only for long-lived handles. Also noted the identical pattern at `generic.go:189`.
- **Downgraded `runtime-pkg-05`** from MEDIUM to LOW: the "injects/overrides sibling JSON keys" scenario overstated the risk (the interpolated values come from DuckDB's own AST serialization, not arbitrary attacker text); the real, contained impact is a rewrite failure on legitimate special-character identifiers.
- **Refined `runtime-pkg-10`** (confidence high→medium): the code does pass `0777`, but the process umask normally reduces the effective mode to `0755`, so the accurate concern is world-*readable* data, not world-writable.
- **Refined `runtime-pkg-01`** to note that the *relative* `..` case is already caught by `filepath.Join`'s cleaning; the two genuine bypasses are absolute-path `..` and the sibling-prefix collision. Confirmed `AllowHostAccess=false` is the hosted-environment sandbox (duckdb.go:550).
- **Confirmed unchanged:** `runtime-pkg-02`, `-03`, `-06`, `-07`, `-09`, `-11`, `-12` (minor location/wording tightening only). Findings are now ordered most-severe first.
