# Runtime Server + Auth — Code Review
**Scope:** `runtime/server/*.go` (all handlers) and `runtime/server/auth/*.go` (interceptors, JWT, claims)  |  **Health:** needs-attention

## Overview
This area implements the runtime's gRPC/REST surface (query, connector, chat/AI, MCP, downloads, repo/git, instances, controller) plus the JWT auth interceptors that gate every request. The code is generally clean, consistently structured, and the vast majority of handlers correctly gate on `auth.GetClaims(ctx, instanceID).Can(...)`. The security model is sound in principle: an issuer (admin) mints instance-scoped JWTs, and an audience (runtime) validates them and resolves per-instance permissions. Crucially, the auth interceptors only *populate* claims on the context — they never enforce per-method permissions — so every handler is responsible for its own gate.

The biggest risks are (1) one entire service — `ConnectorService` — that performs no authorization at all (verified: it is the *only* data-access handler file in the package with zero `auth.GetClaims` calls), exposing object-storage listings and OLAP schema enumeration to any caller, with cross-tenant potential; (2) `QueryBatch` calling `srv.Send` from multiple goroutines, which violates the gRPC single-writer contract; and (3) a JWKS bootstrap loop that leaks ~19 background refresher goroutines on every normal startup. Two SSE goroutine-leak paths (blocking channel sends with no `ctx` select, and event-type validation that runs after streams are launched) round out the concurrency issues. The AI/generation, download-token bounding, and repo/git handlers are otherwise solid.

## Findings

### runtime-server-01 [HIGH] `ConnectorService` handlers perform no authorization

**Status (2026-07-25):** ✅ Fixed and merged in [#9700](https://github.com/rilldata/rill/pull/9700).

- **Category:** security  **Confidence:** high  **Location:** `runtime/server/connector_service.go:12,35,74,106,127,156,185`
- **Problem:** Every handler in `connector_service.go` — `ListBuckets`, `ListObjects`, `OLAPListTables`, `OLAPGetTable`, `ListDatabaseSchemas`, `ListTables`, `GetTable` — jumps straight to `s.runtime.AcquireHandle(...)`/`s.runtime.OLAP(...)` with no `auth.GetClaims(...).Can(...)` check. `AcquireHandle`/`OLAP` take no claims argument and perform no per-user authorization, so nothing gates these RPCs. Every other data-access handler in the package gates access (verified by cross-checking: `connector_service.go` is the only file that calls `AcquireHandle`/`OLAP`/`AsObjectStore`/`AsInformationSchema` yet contains zero `auth.GetClaims` calls; compare `queries.go:30` requiring `ReadOLAP`, `connectors.go:26` requiring `ReadInstance`, `assets.go:26` requiring `ReadObjects`, `generate_metrics_view.go:49` requiring `EditRepo`). These RPCs are registered on the auth-intercepted gRPC server (`server.go:171`) and transcoded to REST (verified in `proto/rill/runtime/v1/connectors.proto`): `GET /v1/instances/{instance_id}/connectors/{connector}/buckets`, `GET /v1/olap/tables`, `GET /v1/connectors/tables`, `GET /v1/connectors/table_metadata`, etc. Notably `OLAPListTables`→`/v1/olap/tables` and `OLAPGetTable`→`/v1/connectors/olap/table` carry no `{instance_id}` in the path at all — the instance is a plain request parameter with no path-level scoping.
- **Failure scenario:** On Rill Cloud the browser calls the runtime directly with an admin-issued JWT. Because the handlers never call `.Can(...)`, a project **viewer** — or even an anonymous caller (no `Authorization` header yields empty claims, `interceptors.go:143`, which the handler never inspects) — can call `OLAPListTables`/`GetTable`/`ListObjects` to enumerate every physical table, its schema, and every object-storage bucket/object reachable with the project's credentials — data that is never exposed through a metrics view and that raw `Query` deliberately restricts to `ReadOLAP`. Worse, the handlers never verify the JWT is scoped to the requested `instance_id`; on a multi-tenant runtime, a valid JWT for instance A (same audience) passes `ParseAndValidate` and can pass `instance_id=B` to enumerate a co-located tenant's tables/storage.
- **Recommendation:** Add the same gate used by the analogous handlers: require `ReadOLAP` for the OLAP/table endpoints and `ReadObjects` for the bucket/object endpoints (or `ReadInstance` for admin-only introspection). Local Rill Developer is unaffected because auth-disabled claims set `SkipChecks`.

**Verification:** CONFIRMED — every handler in `connector_service.go` (`ListBuckets:12`, `ListObjects:35`, `OLAPListTables:74`, `OLAPGetTable:106`, plus `ListDatabaseSchemas`/`ListTables`/`GetTable` further down) goes straight to `s.runtime.AcquireHandle`/`s.runtime.OLAP` with no `auth.GetClaims(...).Can(...)` call; grepping the file confirms zero `auth.GetClaims` references. Contrast `queries.go:30` (`Query` requires `ReadOLAP`) and `downloads.go:306` (`TableRows` requires `ReadOLAP`). The service is registered under the auth-intercepted gRPC server (`server.go:171`) and transcoded to REST (`server.go:182`); `connectors.proto` confirms `OLAPListTables`→`GET /v1/olap/tables` and `OLAPGetTable`→`GET /v1/connectors/olap/table` carry no `{instance_id}` path segment. Anonymous callers get empty-permission claims (`interceptors.go:143-149`) that the handler never inspects.

**Manual Reproduction Steps (for QA):**
1. On a Rill Cloud deployment (or any runtime started with auth enabled), obtain the runtime host and a valid instance ID.
2. Call the OLAP tables endpoint with NO Authorization header: `curl -i 'https://<runtime-host>/v1/olap/tables?instance_id=<instance_id>&connector=<olap-connector>'`.
3. Expected (secure): HTTP 401/403 "action not allowed". Actual: HTTP 200 with a full list of physical tables and their schemas, despite no credentials.
4. Repeat with a viewer-scoped JWT (a token that lacks `ReadOLAP`) against `GET /v1/instances/<id>/connectors/<connector>/buckets` and `/v1/connectors/table_metadata`; observe bucket listings and table metadata returned that the same token cannot reach via `POST /v1/instances/<id>/query`.
5. Multi-tenant check: using a valid JWT minted for instance A (same runtime audience), pass `instance_id=<B>` to `/v1/olap/tables`; observe instance B's tables are enumerable.

### runtime-server-02 [HIGH] `QueryBatch` calls `srv.Send` concurrently from multiple goroutines

**Status (2026-07-25):** ✅ Fixed by removing the unused `QueryBatch` RPC in [#9714](https://github.com/rilldata/rill/pull/9714).

- **Category:** concurrency  **Confidence:** high  **Location:** `runtime/server/batch_query.go:21-24`
- **Problem:** `QueryBatch` spawns one `errgroup` goroutine per query, and each goroutine calls `srv.Send(resp)` on the shared `QueryService_QueryBatchServer` stream (line 23). gRPC server streams are explicitly *not* safe for concurrent `SendMsg` from different goroutines — the generated `Send` calls `ServerStream.SendMsg`, which keeps unsynchronized per-message state (compression buffers, header-sent flag), so concurrent sends are a data race that can corrupt the wire framing or panic. There is no mutex or channel funneling the sends.
- **Failure scenario:** A dashboard issues a `QueryBatch` with several sub-queries. Two queries finish at nearly the same time; their goroutines call `srv.Send` simultaneously → interleaved/garbled frames delivered to the client, a `-race` failure, or a transport panic. This is a batched-load path, so overlap is common.
- **Recommendation:** Serialize the sends: either guard `srv.Send` with a `sync.Mutex`, or have each worker push its `*QueryBatchResponse` onto a channel that one dedicated goroutine drains and sends. (Note: the existing TODO at `batch_query.go:11` covers *performance* reworks — upfront access checks, caching, a bounded pool — not this concurrency bug.)

**Verification:** CONFIRMED — `batch_query.go:16-27`: `g, ctx := errgroup.WithContext(...)` then `for idx, qry := range req.Queries { g.Go(func() error { resp := s.forwardQuery(...); return srv.Send(resp) }) }`. Each per-query goroutine calls `srv.Send` on the shared stream with no mutex or funneling channel. gRPC server streams are documented as unsafe for concurrent `SendMsg`.

**Manual Reproduction Steps (for QA):**
1. Build/run the runtime with the Go race detector enabled: `go build -race` (or run the test suite under `go test -race`).
2. Issue a `QueryService/QueryBatch` gRPC call containing several sub-queries that return quickly and finish near-simultaneously (e.g. multiple `MetricsViewTotals` against a small metrics view). The dashboard batched-load path exercises this naturally.
3. Expected (safe): all sub-responses delivered intact. Actual: intermittent `-race` "DATA RACE" report on `ServerStream.SendMsg`, or garbled/interleaved frames / transport panic under concurrency.
4. If not reproducible on demand: confirm in code at `batch_query.go:21-24` — multiple `g.Go` closures each call `srv.Send` with no serialization.

### runtime-server-03 [MEDIUM] `OpenAudience` JWKS bootstrap loop never breaks on success, leaking refreshers
- **Category:** concurrency  **Confidence:** high  **Location:** `runtime/server/auth/jwt.go:205-224`
- **Problem:** The retry loop `for i := 0; i < 20; i++ { jwks, err = keyfunc.Get(...) ; if err != nil { …sleep… } }` has no `break` on success. `keyfunc.Get` (with `RefreshInterval: time.Hour`) starts a background goroutine that refreshes the JWKS over HTTP on an interval. When the fetch succeeds, the loop keeps iterating and calls `keyfunc.Get` again on every remaining iteration, creating up to 20 independent background-refreshing `*keyfunc.JWKS`. Only the last is assigned to `a.jwks`; the other 19 are never `EndBackground()`-ed (`Close()` at jwt.go:238 stops only the retained one).
- **Failure scenario:** Normal startup (JWKS reachable on first try) leaks 19 goroutines, each performing an HTTP JWKS refresh every hour for the life of the process, plus 20× the JWKS fetches at boot.
- **Recommendation:** `break` out of the loop as soon as `err == nil` (and only sleep/retry on error). This also matches the loop's stated intent ("retry in case the admin server is not ready").

**Verification:** CONFIRMED — `jwt.go:205-224`: `for i := 0; i < 20; i++ { jwks, err = keyfunc.Get(jwksURL, keyfunc.Options{... RefreshInterval: time.Hour ...}); if err != nil { ...sleep... } }`. There is no `break` on `err == nil`, so on a first-try success the loop still calls `keyfunc.Get` on every remaining iteration. Each `keyfunc.Get` starts a background refresher goroutine; only the last is assigned to `a.jwks`, and `Close()` (`jwt.go:238`) calls `EndBackground()` on just that one, leaking the other 19.

**Manual Reproduction Steps (for QA):**
1. Start the runtime (or admin+runtime devtool) so `OpenAudience` runs against a reachable JWKS endpoint (normal startup — first fetch succeeds).
2. Not directly user-visible; confirm via one of: (a) add a goroutine-count assertion / `pprof` goroutine dump right after startup and observe ~20 JWKS-refresh goroutines instead of 1; or (b) point the runtime at a JWKS server and count the burst of `.well-known/jwks.json` HTTP requests at boot — expected 1, actual up to 20, then hourly refreshes multiplied by ~20.
3. How to confirm in code: `jwt.go:205-224`, note the missing `break` after `if err != nil`.

### runtime-server-04 [MEDIUM] SSE plumbing blocks on an unbuffered channel with no `ctx` select, stranding goroutines on client disconnect
- **Category:** concurrency  **Confidence:** medium  **Location:** `runtime/server/sse.go:137,171,218,265` and `runtime/server/chat.go:525`
- **Problem:** The per-stream shim callback does a bare blocking send: `events <- event; return nil` (sse.go:137/171/218, chat.go:525), and the background goroutine's terminal error send (`events <- &sseEvent{Event:"error",…}`, sse.go:265) is likewise unguarded. `serveSSEUntilClose` (sse.go:285) returns on the *first* write error (sse.go:322/326/330) and then abandons the channel without draining it. A plain channel send is not unblocked by `ctx` cancellation, so once the consumer has returned, any producer still committed to a send blocks forever. The clean disconnect path (ctx cancellation → streams return → `grp.Wait()` returns → `close(events)`) is handled, but it races against a producer that has already committed to `events <- ev` when `serveSSEUntilClose` returns on a write error.
- **Failure scenario:** A browser tab watching `/sse?events=file,resource` disconnects while events are flowing. `serveSSEUntilClose` receives one event, its write fails, and it returns; a second stream (or the terminal error-event send) is mid-`events <- …` with no consumer and blocks permanently. `grp.Wait()` never returns, so the coordinating goroutine plus the blocked worker leak. Because dashboards keep SSE connections open and clients disconnect on every navigation/refresh, these leaks accumulate over the life of a long-running runtime.
- **Recommendation:** Make every send select on the stream context: `select { case events <- ev: case <-ctx.Done(): return ctx.Err() }`. Alternatively, have `serveSSEUntilClose` keep draining (discarding) the channel until it is closed so producers always unblock.

**Verification:** CONFIRMED — bare unguarded sends `events <- event; return nil` at `sse.go:137` (file shim), `:171` (resource shim), `:218` (log shim), and `chat.go:525` (streaming completion shim); the terminal error send `events <- &sseEvent{Event: "error", ...}` at `sse.go:265` is likewise unguarded. `serveSSEUntilClose` (`sse.go:311-333`) `return`s on the first write error (`sse.go:322/326/330`) and does not drain the channel afterward. The channel is unbuffered (`sse.go:111 events := make(chan *sseEvent)`), and a bare channel send is not unblocked by ctx cancellation, so a producer already committed to a send after the consumer returns blocks forever, stranding `grp.Wait()`.

**Manual Reproduction Steps (for QA):**
1. Open a dashboard/UI view that holds an SSE connection to `/v1/instances/<id>/sse?events=file,resource&resources_replay=true` (multiple event types increases overlap).
2. Force a write error on the consumer: disconnect the client abruptly mid-stream (kill the tab / drop the TCP connection) while events are actively flowing.
3. Repeat navigation/refresh many times to accumulate.
4. Expected (no leak): server goroutines for the closed connection are reclaimed. Actual: leaked goroutine pairs (blocked producer + `grp.Wait()` coordinator) accumulate; observe growth via `/debug/pprof/goroutine` or a steadily rising goroutine count over the process lifetime.
5. How to confirm in code: `sse.go:137/171/218/265` bare sends + `sse.go:311-333` early `return` without drain.

### runtime-server-05 [MEDIUM] SSE event-type validation runs *after* goroutines are started, leaking the ones already launched
- **Category:** bug  **Confidence:** medium  **Location:** `runtime/server/sse.go:112-229`
- **Problem:** The `switch eventType` that validates and launches streams handles unknown values in its `default` case (line 226) by writing `http.Error(...)` and `return`ing. But `eventTypes` is sorted in place (line 84), so a valid type that sorts before an invalid one (e.g. `events=file,zzz` → `[file, zzz]`) starts `grp.Go(WatchFiles)` first, then hits `default` and returns from the handler. `serveSSEUntilClose` and the `grp.Wait()` coordinator are never reached, so the already-started stream goroutine has no consumer and — combined with the bare send in finding 04 — blocks the moment it produces an event. With `..._replay=true` (which the UI commonly requests), the stream emits immediately, making the leak near-certain rather than racy.
- **Failure scenario:** A client sends `events=resource,bogus&resources_replay=true`. `WatchResources` is launched and immediately replays existing resources into `events <- …`; the handler then 400s without ever consuming the channel, so the resource-stream goroutine leaks. Repeated malformed requests accumulate leaked goroutines.
- **Recommendation:** Validate all requested event types up front (before the `grp.Go` loop) and reject unknown values there, so no stream is started unless every requested type is valid.

**Verification:** CONFIRMED — `eventTypes` is sorted in place at `sse.go:84` (`slices.Sort(eventTypes)`), and the `switch eventType` loop (`sse.go:112-230`) launches `grp.Go(...)` for each valid type as it iterates, only hitting the `default: http.Error(...); return` (`sse.go:226-228`) when it reaches an invalid one. Because of the sort, `events=file,zzz` sorts to `[file, zzz]`, so `WatchFiles` is started before the handler 400s and returns — `serveSSEUntilClose`/`grp.Wait()` are never reached, leaving the started stream with no consumer. Combined with the bare send (finding 04) and `resources_replay=true` (immediate emission), the leak is near-certain.

**Manual Reproduction Steps (for QA):**
1. Send a request mixing a valid event type (sorting first) with an invalid one, plus replay: `curl -N 'https://<runtime-host>/v1/instances/<id>/sse?events=file,zzz&files_replay=true' -H 'Authorization: Bearer <token>'`.
2. Expected (clean): HTTP 400 "unknown event type" and no server-side stream started. Actual: HTTP 400 returned, but the `file` stream goroutine was already launched and (with replay) immediately produces an event into a channel with no consumer, leaking the goroutine.
3. Repeat the malformed request many times; observe goroutine growth via `/debug/pprof/goroutine`.
4. How to confirm in code: `sse.go:84` sort + `sse.go:112-228` launch-before-validate ordering.

### runtime-server-06 [LOW] `jwtClaims.Claims` mutates and aliases the shared `Attrs` map
- **Category:** concurrency  **Confidence:** medium  **Location:** `runtime/server/auth/claims.go:29-36`
- **Problem:** `Claims()` takes `attrs := c.Attrs` and, when the key is absent, writes `attrs["id"] = c.Subject` directly into the JWT's shared map, then returns that same map as `SecurityClaims.UserAttributes`. `Claims` is called repeatedly and concurrently (e.g. `QueryBatch` fan-out, each sub-handler calling `auth.GetClaims`). This is not *currently* a live `fatal error: concurrent map writes` only because `parseClaims` (interceptors.go:173) calls `claims.Claims("")` once, synchronously, before the handler runs — pre-populating `"id"` so later concurrent calls take the read-only branch. Every returned `SecurityClaims` also aliases the same underlying map, so any downstream mutation of `UserAttributes` would leak across requests. The safety here rests on a load-bearing, non-obvious invariant (the observability pre-warm).
- **Failure scenario:** If the pre-warm at interceptors.go:173 is removed/reordered, or `Claims` is invoked concurrently via a path that doesn't pre-warm, two goroutines racing the `attrs["id"]=…` write crash the process. Separately, any code that writes to `UserAttributes` corrupts the shared JWT map.
- **Recommendation:** Make `Claims` pure: build a fresh map (`maps.Clone(c.Attrs)`) before injecting `"id"`, so it never mutates or aliases shared state.

**Verification:** PLAUSIBLE — `claims.go:29-36`: `attrs := c.Attrs` (aliases the JWT's map, not a copy), and when `"id"` is absent writes `attrs["id"] = c.RegisteredClaims.Subject` directly into it, then returns that same map as `UserAttributes` (`:47`, `:69`). This is a real aliasing/latent-race hazard. It is not a live crash today because `interceptors.go:173` calls `claims.Claims("")` once synchronously before any handler runs, pre-populating `"id"` so later concurrent calls (e.g. `QueryBatch` fan-out) take the read-only branch at `:34`. The safety rests entirely on that non-obvious pre-warm, hence PLAUSIBLE rather than CONFIRMED-live.

**Manual Reproduction Steps (for QA):**
1. Not user-reproducible in current code (the interceptors.go:173 pre-warm masks the write race).
2. To demonstrate the latent bug: write a Go unit test that constructs a `jwtClaims` with `Attrs` lacking `"id"` and calls `Claims("")` concurrently from many goroutines under `-race` (bypassing the pre-warm) — expect a "concurrent map writes" / race report. Separately, mutate the returned `UserAttributes` map and observe the change leaking into a second `Claims(...)` result (shared-map aliasing).
3. How to confirm in code: `claims.go:30` (`attrs := c.Attrs`, no clone) and `claims.go:35` (write into shared map); safety dependency at `interceptors.go:173`.

### runtime-server-07 [LOW] `gzipDecompress` reads the download-token payload with no size bound
- **Category:** security  **Confidence:** low  **Location:** `runtime/server/downloads.go:475`
- **Problem:** `gzipDecompress` does `io.ReadAll(gzip.NewReader(...))` with no limit. This is defense-in-depth only, and it is **not currently exploitable**: the compressed payload is a field of a token that `s.codec` (securetoken) authenticates and encrypts, so it cannot be forged; the payload is itself produced by the server from a `proto.Marshal` of an `ExportRequest`; and generation caps the *encoded* token at `maxDownloadTokenSize` = 7.5 KB (`downloads.go:33,413`), bounding the compressed input well under that.
- **Failure scenario:** Low. A future change that widens where tokens originate, or a codec regression, would leave an unbounded gzip expansion on the parse path.
- **Recommendation:** Wrap the reader with `io.LimitReader` at a generous ceiling (e.g. a few MB) and error if exceeded.

**Verification:** CONFIRMED (as latent/defense-in-depth, not currently exploitable) — `downloads.go:469-476`: `gzipDecompress` does `io.ReadAll(gzip.NewReader(bytes.NewReader(v)))` with no size bound. The compressed bytes come from `downloadToken.Request`, decoded via `s.codec.Decode` (`downloads.go:423`) which authenticates/decrypts the token, so it cannot be forged; the payload is server-produced from `proto.Marshal(ExportRequest)` (`downloads.go:387-392`); and `generateDownloadToken` caps the encoded token at `maxDownloadTokenSize = 7.5 KB` (`downloads.go:33,413`). So the unbounded read exists but is not reachable with attacker-controlled large input today.

**Manual Reproduction Steps (for QA):**
1. Not user-reproducible — download tokens are encrypted/authenticated by `s.codec` and server-generated, and capped at 7.5 KB encoded, so no caller can supply an oversized/malicious gzip payload.
2. How to confirm in code: `downloads.go:475` (`io.ReadAll` with no `io.LimitReader`); note the mitigating controls at `downloads.go:408-415` (encode + size cap) and `downloads.go:423` (`codec.Decode`).

### runtime-server-08 [LOW] Unauthenticated `/v1/health` leaks internal component error strings
- **Category:** security  **Confidence:** medium  **Location:** `runtime/server/health.go:97`
- **Problem:** `healthCheckHandler` is mounted **without** auth middleware (`server.go:191` wraps only `observability.Middleware`, unlike sibling routes that add `auth.HTTPMiddleware`). On failure it returns raw internal error text in the 500 body — verified via `httputil.WriteError`, which serializes `{"error": err.Error()}` — as `"controller: <err>, repo: <err>, olap: <err>"`. The authenticated `Health` RPC (health.go:19) requires `ManageInstances`; this public probe requires nothing.
- **Failure scenario:** An unauthenticated probe of `/v1/health` during an incident can read internal error messages (connector/driver details, paths) that aid reconnaissance.
- **Recommendation:** Return a generic status for the public health probe and log the detailed error server-side; keep detailed errors on the authenticated `Health`/`InstanceHealth` RPCs.

**Verification:** CONFIRMED — `/v1/health` is mounted with only `observability.Middleware`, no `auth.HTTPMiddleware` (`server.go:191`), unlike `/v1/download` and the other routes at `server.go:194-212`. On failure `healthCheckHandler` returns `httputil.Error(http.StatusInternalServerError, errors.New("controller: "+h.Controller+", repo: "+h.Repo+", olap: "+h.OLAP))` (`health.go:97`), and `httputil.WriteError` serializes `{"error": err.Error()}` into the body (`httputil.go:68`). So an unauthenticated 500 body leaks raw internal component error strings.

**Manual Reproduction Steps (for QA):**
1. Induce a component failure on a runtime instance (e.g. a broken OLAP connector or repo so the controller/repo/olap health string is non-empty).
2. Probe the endpoint with NO auth: `curl -i 'https://<runtime-host>/v1/health'`.
3. Expected (secure): generic 500 status with no internal detail. Actual: 500 with body like `{"error":"controller: <detail>, repo: <detail>, olap: <detail>"}` exposing connector/driver/path detail to an unauthenticated caller.
4. Contrast: the authenticated `Health` RPC requires `ManageInstances` (`health.go:19`).

### runtime-server-09 [LOW] Global mutex serializes all SQL JSON (de)serialization on the query path
- **Category:** performance  **Confidence:** low  **Location:** `runtime/server/queries.go:19`
- **Problem:** A package-level `var mutex sync.Mutex` guards every `jsonvalue.Unmarshal` in `unmarshalJSON`, which is called from `ensureLimits`/`createConstantLimit`/`createLimitModifier` on the raw-SQL `Query` path. `jsonvalue.Unmarshal` builds a fresh value and has no obvious need for a process-wide lock, so this reads as an unexplained global bottleneck. Practical impact is small because the path is only hit for raw `Query` with a non-zero limit, and each call is dominated by two OLAP round-trips — but the lock is undocumented.
- **Failure scenario:** Concurrent raw-SQL queries with a `limit` contend on one global lock. Likely minor.
- **Recommendation:** Document why the lock exists; if it guards non-thread-safe library global state, scope it more tightly or replace the library. If it is unnecessary, remove it.

**Verification:** CONFIRMED — `queries.go:19` declares package-level `var mutex sync.Mutex`; `unmarshalJSON` (`queries.go:21-26`) locks it around every `jsonvalue.Unmarshal`. It is called from `ensureLimits` (`queries.go:130`), `createConstantLimit` (`queries.go:256`), and `createLimitModifier` (`queries.go:274`), all reachable only from `Query` (`queries.go:29`) when `req.Limit != 0` (`queries.go:46`). The global lock has no documented rationale. Practical impact is minor since the path is dominated by two OLAP round-trips per call.

**Manual Reproduction Steps (for QA):**
1. Not meaningfully user-observable (impact is a minor, OLAP-round-trip-bound serialization point, not a correctness bug).
2. How to confirm in code: `queries.go:19-26` — the undocumented package-level `sync.Mutex` guarding `jsonvalue.Unmarshal`, hit only on raw `Query` with a non-zero `Limit`.

### runtime-server-10 [LOW] Export handler writes an HTTP 500 + error text after streaming has begun, corrupting the download
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/server/downloads.go:355-358`
- **Problem:** `downloadHandler` calls `q.Export(...)` which, via `PreWriteHook`, sets the `Content-Type`/`Content-Disposition` headers and then streams rows directly to `w` (downloads.go:328-354). If `Export` fails *after* the first bytes are written (e.g., OLAP connection drop, context timeout mid-stream), the handler falls through to `http.Error(w, err.Error(), http.StatusInternalServerError)` (line 356). Because the response is already committed (status 200, headers sent, partial body written), `WriteHeader(500)` is a no-op ("superfluous WriteHeader" warning) and the error string is appended to the partially-streamed data — the client receives a 200 with a corrupt file rather than a clean failure.
- **Failure scenario:** A large CSV export fails halfway through; the user downloads a truncated file with `internal error: ...` text appended and no error status to signal the failure.
- **Recommendation:** Track whether any bytes have been written (or whether `PreWriteHook` ran); if streaming has started, stop appending an HTTP error body and instead log + abort the connection (or reset the stream) so the client sees a failed transfer.

**Verification:** CONFIRMED — `downloadHandler` calls `q.Export(req.Context(), s.runtime, ..., w, &runtime.ExportOptions{...})` (`downloads.go:328-354`), whose `PreWriteHook` sets `Content-Type`/`Content-Disposition` on `w` and Export then streams rows to `w`. On any error the handler unconditionally falls through to `http.Error(w, err.Error(), http.StatusInternalServerError)` (`downloads.go:355-357`) with no check for whether bytes were already written. Once headers/body are committed, `WriteHeader(500)` is a no-op ("superfluous WriteHeader") and the error text is appended to the partial file.

**Manual Reproduction Steps (for QA):**
1. Start a large CSV/XLSX export via the dashboard "Export" flow (or `GET /v1/download?token=...`) against a big result set.
2. Cause the export to fail mid-stream (e.g. drop/kill the OLAP connection, or trigger a context timeout after some rows have streamed).
3. Expected (clean): the client receives a non-200 status signaling failure. Actual: HTTP 200 already sent; the downloaded file is truncated with `internal error: ...` text appended at the end, and no error status. Server logs show a "superfluous response.WriteHeader" warning.
4. How to confirm in code: `downloads.go:355-358` — `http.Error` after `q.Export` may have already streamed headers + partial body.

### runtime-server-11 [MEDIUM] Raw-query limit rewrite uses DuckDB-only `json_serialize_sql`/`json_deserialize_sql`, breaking `Query` with a limit on non-DuckDB OLAPs
- **Category:** product-improvement  **Confidence:** medium  **Location:** `runtime/server/queries.go:110-160`
- **Problem:** `ensureLimits` rewrites the SQL by round-tripping it through the *target* OLAP connector's `json_serialize_sql(...)` and `json_deserialize_sql(...)` (queries.go:112,142). These functions are DuckDB-specific. `Query` (queries.go:29) accepts any connector via `s.runtime.OLAP(ctx, req.InstanceId, req.Connector)`, so a raw query against ClickHouse/Druid/Pinot with a non-zero `Limit` executes `select json_serialize_sql(...)` on that engine and fails with an opaque "unknown function" error before the user's query ever runs.
- **Failure scenario:** A user of a ClickHouse-backed project runs a raw SQL query with a row limit set and gets a confusing DuckDB-function error instead of their results.
- **Recommendation:** Gate the JSON-based rewrite to DuckDB dialects; for other engines, either skip the rewrite and apply a post-hoc row cap on the result stream, or wrap the query as `SELECT * FROM (<sql>) LIMIT n` where the dialect allows it.

**Verification:** CONFIRMED — `Query` (`queries.go:29`) resolves any connector via `s.runtime.OLAP(ctx, req.InstanceId, req.Connector)` (`queries.go:39`) and, when `req.Limit != 0` (`queries.go:46`), calls `ensureLimits`, which executes `select json_serialize_sql(?::VARCHAR)::BLOB` (`queries.go:112`) and `select json_deserialize_sql(json(?))` (`queries.go:142`) against that same connector. These functions are DuckDB-specific, so a raw query with a limit against a ClickHouse/Druid/Pinot connector fails on the serialize call with an "unknown function" error before the user's SQL runs. No dialect guard is present.

**Manual Reproduction Steps (for QA):**
1. Create a project whose OLAP connector is ClickHouse (or another non-DuckDB engine).
2. Call `POST /v1/instances/<id>/query` (QueryService/Query) with a valid raw `sql`, the ClickHouse `connector`, and a non-zero `limit` (e.g. `{"sql":"SELECT 1","connector":"clickhouse","limit":100}`).
3. Expected: 100-row-capped result (or the query's natural result). Actual: an opaque engine error such as "unknown function json_serialize_sql", because `ensureLimits` runs DuckDB-only SQL on ClickHouse.
4. Control: the same request with `limit` omitted/0 skips `ensureLimits` and succeeds, confirming the limit path is the trigger.
5. How to confirm in code: `queries.go:110-160`.

## Product Improvements
- **runtime-server-08** (above) — public health endpoint leaking internal error detail.
- **runtime-server-11** (above) — raw-query limits unsupported on non-DuckDB OLAP connectors.
- **`resolveExportLimit` base-0 semantics** — `resolveExportLimit` (downloads.go:361) treats a base limit of `0` as "unlimited" and then ignores any per-query override (`override < res` is never true when `res == 0`). This appears intentional for full exports but can export far more rows than the originating dashboard view; worth confirming against product expectations.

## New Feature Ideas
- **Per-service authz test harness (S).** A table-driven test that iterates every RPC and asserts a minimally-privileged/anonymous caller is rejected would have caught finding 01 and prevents future regressions. High value for a security-critical surface.
- **Backpressure-aware SSE with heartbeats (M).** Add periodic keep-alive comments and a bounded, `ctx`-aware event buffer to the SSE multiplexer so slow/disconnected clients are detected promptly and never strand goroutines (addresses findings 04/05 structurally and improves proxy compatibility).
- **Download-token introspection endpoint (S).** A read-only "what will this export contain" endpoint (query summary, row cap, claims scope, expiry) would let the UI warn before a large export and help debug the "token too large" error surfaced at downloads.go:413.
- **Streaming/chunked exports (L).** Large exports currently run a single `q.Export` to the response writer with a 1-hour token; a resumable, chunked export (object-store staged + signed URL) would remove the header-size ceiling and browser-timeout limits, and would also resolve finding 10 (no in-band error possible once streaming starts).
- **Structured error taxonomy for AI generation (S).** `generateMetricsViewYAMLWithAI`/`generateCanvasDashboardYAMLWithAI` collapse many failure modes (timeout, invalid YAML, no valid measures) into a warn-and-fallback path; surfacing a typed reason to the client would improve the "AI failed, used simple generator" UX.

## Notes & Missing Test Coverage
- **Concurrency is untested where it matters most.** There is no test exercising `QueryBatch` with multiple queries under `-race` (finding 02), nor an SSE test that disconnects mid-stream to assert no goroutine leak (findings 04/05), nor a test asserting `OpenAudience` starts exactly one refresher (finding 03). These are precisely the paths most likely to break silently.
- **`auth` package** has good unit coverage (`jwt_test.go`, `interceptors_test.go`), but none covers the bootstrap-retry loop or the `jwtClaims.Claims` concurrent-read/aliasing behavior (finding 06).
- **`connector_service.go` has no test file at all**, which is consistent with the missing authz (finding 01) never being asserted.
- The exact multi-tenancy blast radius of finding 01 (whether co-located instances always share a single runtime audience) depends on deployment topology and was not traced end-to-end; the code-level absence of any authz check is certain regardless, and matches an unintended gap relative to every sibling handler.

## Verification Notes
All nine original findings were independently re-verified against the code and **all survived** (removed_count = 0); several were tightened for accuracy, and two new findings were added.

- **runtime-server-01 (kept, HIGH):** Confirmed. `RegisterConnectorServiceServer` at `server.go:171` mounts these RPCs under the auth interceptors, which only populate claims. Strengthened with a verified cross-check: `connector_service.go` is the *only* data-access file in the package with zero `auth.GetClaims` calls. Corrected/expanded the REST paths against `proto/rill/runtime/v1/connectors.proto` (e.g. `OLAPListTables` → `/v1/olap/tables`, with no `{instance_id}` in the path). Confirmed anonymous callers get empty claims that the handler never inspects.
- **runtime-server-02 (kept, HIGH):** Confirmed the per-query `errgroup` goroutines each call `srv.Send`. Corrected the report's claim that "the existing TODO anticipates a rework here" — the TODO at `batch_query.go:11` is about performance (upfront access checks, cache, pool), not the Send race.
- **runtime-server-03 (kept, MEDIUM, high confidence):** Confirmed no `break` on success; the loop calls `keyfunc.Get` on all 20 iterations, leaking 19 background refreshers. Location refined to jwt.go:205-224.
- **runtime-server-04 (kept, MEDIUM):** Confirmed bare sends at all cited lines and that `serveSSEUntilClose` returns on the first write error without draining. Clarified that the clean ctx-cancellation path *is* handled and the leak is the race where a producer is already committed to a send when the consumer returns.
- **runtime-server-05 (kept, MEDIUM):** Confirmed `slices.Sort` (sse.go:84) orders event types so a valid one launches before an invalid one triggers the `default` 400. Added the concrete detail that `..._replay=true` (used by the UI) makes the leak near-certain rather than merely racy.
- **runtime-server-06 (kept, LOW):** Confirmed the shared-map mutation/aliasing and the pre-warm at interceptors.go:173. Reframed honestly: this is **not** a live data race today (the pre-warm makes concurrent calls read-only); it is a latent fragility + cross-request aliasing hazard. Confidence left at medium.
- **runtime-server-07 (kept, LOW):** Confirmed the unbounded `io.ReadAll` (corrected location to downloads.go:475, the read itself). Emphasized it is defense-in-depth only and not currently exploitable (authenticated/encrypted token, 7.5 KB encoded cap, server-generated payload).
- **runtime-server-08 (kept, LOW):** Confirmed `/v1/health` has no auth middleware (`server.go:191`) and that `httputil.WriteError` writes `err.Error()` into the body. Recategorized from product-improvement to **security** (unauthenticated info disclosure), which is the more accurate classification; it remains listed under Product Improvements as well.
- **runtime-server-09 (kept, LOW):** Confirmed the package-level mutex around `jsonvalue.Unmarshal`. Reframed to emphasize the real issue is an undocumented global lock; practical performance impact is minor because the path is OLAP-round-trip-bound.
- **runtime-server-10 (NEW, LOW):** Verified that `downloadHandler` calls `http.Error` after `q.Export` may already have streamed headers + partial body (downloads.go:328-358), yielding a corrupt 200 download on mid-stream failures.
- **runtime-server-11 (NEW, MEDIUM):** Verified `ensureLimits` (queries.go:110-160) round-trips SQL through DuckDB-only `json_serialize_sql`/`json_deserialize_sql` against the request's chosen connector, so `Query` with a non-zero `Limit` fails on non-DuckDB OLAPs. This promotes the reviewer's prior "Non-DuckDB raw-query limits" product note to a tracked finding.
