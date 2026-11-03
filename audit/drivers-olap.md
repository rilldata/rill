# OLAP Drivers — Code Review
**Scope:** `runtime/drivers/{olap.go, dialect.go, information_schema.go}` and the OLAP drivers `duckdb/`, `clickhouse/`, `druid/` (+ `druidsqldriver/`), `pinot/` (+ `sqldriver/`), `starrocks/`, `mysql/`, `postgres/`, `sqlite/`.  |  **Health:** fair

## Overview
This area implements the data-plane OLAP abstraction: a shared `OLAPStore`/`Dialect`/`InformationSchema` contract plus one driver per engine. DuckDB and ClickHouse are the heavyweight drivers (embedded process management, connection pools gated by meta/olap semaphores, DB reopen logic, model execution); Druid and Pinot are HTTP-backed custom `database/sql` drivers; MySQL/Postgres/StarRocks are thin `sqlx` wrappers. The abstraction is generally well structured and the semaphore-based pool gating is thoughtfully done. The biggest risks cluster around (1) connection/process/goroutine lifecycle on failure paths, (2) a metrics-callback data race in DuckDB, and (3) a handful of correctness bugs in the newer/HTTP drivers (a reachable nil-panic in Pinot, in-place mutation of the caller's `Statement`, and type-scan mismatches). None of the query hot paths are grossly wrong (DuckDB/ClickHouse `Query` release their connections correctly on every branch), but several partial-open and error paths leak resources that accumulate under repeated reconcile failures.

## Findings

### drivers-olap-01 [HIGH] Pinot `All()` panics on nil table when any schema fetch fails
- **Category:** bug  **Confidence:** high  **Location:** `runtime/drivers/pinot/information_schema.go:266-288`
- **Problem:** `All` allocates `tables := make([]*drivers.OlapTable, len(paginatedTables))` and fills each slot from a parallel `errgroup`. When `c.Lookup` fails, the goroutine logs and `return nil` (error swallowed), leaving that index `nil`. It then calls `sort.Slice(tables, func(i, j int){ return tables[i].Name < tables[j].Name })`, which dereferences `tables[i].Name` on the nil element. (The per-goroutine writes to distinct `tables[i]` indices are themselves race-free; the defect is the retained nil slot.)
- **Failure scenario:** A user opens the connector/table explorer against Pinot; one table's `/schema` endpoint returns a transient non-200 (restarting controller) or a decode error. `Lookup` returns an error, its slot stays nil, and `sort.Slice` (running in the caller goroutine after `g.Wait()`) dereferences nil → panic. Even if the sort is removed, the returned slice contains nil entries that downstream code will deref.
- **Recommendation:** Collect results into a local slice guarded by a mutex (append only successes), or compact out nil entries before sorting. Surface the per-table error via the logger rather than swallowing it (and remove the `fmt.Printf`, see drivers-olap-14).

**Verification:** CONFIRMED — `runtime/drivers/pinot/information_schema.go:266` allocates `tables := make([]*drivers.OlapTable, len(paginatedTables))`; on `c.Lookup` error the goroutine does `fmt.Printf(...); return nil` (`:275-277`) leaving the slot nil, and `sort.Slice` at `:286-288` dereferences `tables[i].Name` on the nil element. Exactly as described.

**Manual Reproduction Steps (for QA):**
1. Create a Rill project with a Pinot connector (`connector: pinot` with a valid `dsn` pointing at a Pinot cluster that has at least 2 tables).
2. Start Rill Developer (`rill start`) and confirm the Pinot tables are listed (e.g. in the connector/table explorer in the sidebar, or via the `/v1/instances/{id}/connectors/{name}/olap/tables` API).
3. Restart the Pinot controller (or use a firewall rule/proxy to make the controller's `/schemas/<oneTable>` endpoint return HTTP 500 for one specific table) while leaving the rest of the cluster reachable.
4. Refresh the table listing in Rill while the controller is unavailable for that table.
5. Observed bug: the Rill runtime panics with a nil-pointer dereference (visible as a crash/stack trace in the terminal running `rill start`) instead of showing the table list.

### drivers-olap-02 [MEDIUM] DuckDB data race on `dbConnCount` in the metrics callback
- **Category:** concurrency  **Confidence:** high  **Location:** `runtime/drivers/duckdb/duckdb.go:214` (read) vs `:730,:744` (writes)
- **Problem:** The observable-gauge callback registered in `Open` reads `int64(c.dbConnCount)` with no synchronization, while `acquireDB` (`:730`) and its release closure (`:744`) mutate `c.dbConnCount` under `c.dbCond.L`. `dbConnCount` is a plain `int` (no atomics). The OTEL collection goroutine and query goroutines therefore race.
- **Failure scenario:** Under concurrent queries the metric collector reads `dbConnCount` while a query increments/decrements it. This is a data race by the Go memory model (flagged by `-race`, undefined behavior; torn reads possible on 32-bit). The reported gauge can also be stale/garbage.
- **Recommendation:** Make `dbConnCount` an `atomic.Int64` (or read it under `c.dbCond.L` inside the callback). Atomic is simplest since the callback must not block.

**Verification:** CONFIRMED — `runtime/drivers/duckdb/duckdb.go:214` reads `int64(c.dbConnCount)` in the OTEL callback with no lock; `dbConnCount` is a plain `int` (`:320`) mutated under `c.dbCond.L` at `:730` (increment) and `:744` (decrement in the release closure). The callback runs on the metrics-collection goroutine, so this is a data race per the Go memory model.

**Manual Reproduction Steps (for QA):**
Not user-reproducible through the UI (the race has no reliably visible symptom; at worst a garbage gauge value). Confirm in code/test: build the runtime with the race detector (`go test -race ./runtime/drivers/duckdb/...` with a test that runs concurrent queries while an OTEL metrics reader collects, or run `rill start` from a `-race` build with metrics export enabled and issue concurrent dashboard queries). The race detector flags the unsynchronized read at `runtime/drivers/duckdb/duckdb.go:214` versus the writes at `:730`/`:744`.

### drivers-olap-03 [MEDIUM] ClickHouse `Open` leaks read/write DB handles and the embedded server on post-connect failures
- **Category:** bug  **Confidence:** high  **Location:** `runtime/drivers/clickhouse/clickhouse.go:308-389`
- **Problem:** After `openHandle` succeeds (and possibly a second `writeDB`), `Open` runs a version-check query (`:326-340`) and a settings probe (`:347`). If the write-DSN open fails (`:321-323`), or the version `row.Scan` errors (`:335-336`), or `!isSupported` (`:338-340`), the function returns before the `c := &Connection{...}` at `:368` and therefore without closing `db`/`writeDB` and without calling `embed.stop()`. `Close` (`:471-492`) is the only place that performs this cleanup, and it is unreachable because `c` was never returned. For provisioned/embedded mode, `embed.start()` already incremented the singleton `refs` (embed.go:70), so the embedded ClickHouse **process is never shut down** and its ref count leaks upward permanently.
- **Failure scenario:** A managed/local ClickHouse whose first version query transiently fails (or an old server) causes `Open` to error; the runtime retries reconcile, each attempt leaks a `*sql.DB` pool, and the embedded server keeps running with a ref count that can never reach 0.
- **Recommendation:** On every error path after `db`/`writeDB`/`embed` are created, close them (`errors.Join` the cleanup). A `defer` that closes them unless a `success` flag is set is the cleanest pattern.

**Verification:** CONFIRMED — `runtime/drivers/clickhouse/clickhouse.go:308` opens `db`; the error returns at `:317-318` (write-DSN parse), `:321-323` (write open), `:335-336` (version scan), and `:338-340` (`!isSupported`) all return without closing `db`/`writeDB` or calling `embed.stop()`. The `Connection` struct that would own cleanup is only built at `:368`, and `Close` (`:471-492`) is the only cleanup site. `embed.start()` increments the singleton `refs` at `embed.go:70`, so the embedded process leaks on these paths.

**Manual Reproduction Steps (for QA):**
1. Run a ClickHouse server older than version 22.7 (e.g. `docker run clickhouse/clickhouse-server:22.6`).
2. Create a Rill project with a ClickHouse connector pointing at it (`host`/`port` in `connector.yaml`) and start Rill Developer.
3. The connector errors with "clickhouse version must be 22.7 or higher"; trigger repeated reconciles by saving the connector YAML file several times.
4. Observed bug: each failed open leaks a `*sql.DB` pool. Confirm via the runtime's pprof endpoint (`/debug/pprof/goroutine`) showing an accumulating count of `database/sql.(*DB).connectionOpener` goroutines, or via netstat showing lingering connections to the ClickHouse port. The embedded-server variant (managed ClickHouse process never stopped) requires the version/settings probe to fail after `embed.start()`, which is harder to stage manually — confirm in code at `clickhouse.go:299-340`.

### drivers-olap-04 [MEDIUM] DuckDB `Open` leaks the metrics-callback registration when `reopenDB` fails
- **Category:** bug  **Confidence:** high  **Location:** `runtime/drivers/duckdb/duckdb.go:213-229`
- **Problem:** `Open` registers the `connections_in_use` callback at `:213` (storing `c.registration`), then calls `c.reopenDB`. On failure it closes `remote` but never calls `c.registration.Unregister()`. The registration retains the closure `c` on the global meter and continues firing forever, reading `c.dbConnCount` (racily, see -02). Note the `periodicallyEmitStats`/`periodicallyCheckConnDurations` goroutines are started only *after* `reopenDB` succeeds (`:231-233`), so this path leaks no goroutine and `c.cancel()` is not strictly required here — but the registration leak alone keeps the whole `connection` object alive. On a failed open `c.db` is nil (`reopenDB` returns before/with a nil handle), so there is no leaked DB pool here — the leak is the registration + retained object.
- **Failure scenario:** A project whose DuckDB can't be opened (locked file, bad path) fails `Open`; the caller discards the handle but the OTEL registration retains `c` and periodically reads its `dbConnCount`. Repeated reconcile failures leak one registration + one retained `connection` each.
- **Recommendation:** On the `reopenDB` error path, `_ = c.registration.Unregister()` (and `c.cancel()` for tidiness) before returning, mirroring `Close` (`:362-372`).

**Verification:** CONFIRMED — `runtime/drivers/duckdb/duckdb.go:213-216` registers the callback into `c.registration`; the `reopenDB` failure path at `:220-229` closes only `remote` and returns without `c.registration.Unregister()` or `c.cancel()`. The stats goroutines start after success at `:231-233`, matching the refined description (registration + retained object leak only).

**Manual Reproduction Steps (for QA):**
1. Start Rill Developer on a project (`rill start` in project A).
2. Without stopping it, start a second Rill instance on the same project directory; the second instance fails with "failed to open database (is Rill already running?)".
3. The failure itself is expected; the leak (a live OTEL callback retaining the discarded `connection` object per failed open) is not visible in the UI.
4. Confirm in code: `runtime/drivers/duckdb/duckdb.go:220-229` returns on `reopenDB` error without calling `c.registration.Unregister()`; a heap profile after repeated failed opens shows retained `duckdb.connection` objects held by the global meter's callback list.

### drivers-olap-05 [MEDIUM] ClickHouse `contextWithoutDeadline` leaks a goroutine (and context) per query on read-only clusters
- **Category:** concurrency  **Confidence:** medium  **Location:** `runtime/drivers/clickhouse/olap.go:533-540` (used at `:513,:521,:529`)
- **Problem:** When `supportSettings == false` (read-only ClickHouse user — a documented, real configuration), `SQLConn.ExecContext`/`QueryxContext` wrap the context with `contextWithoutDeadline`, which spawns `go func(){ <-parent.Done(); cancel() }()` and returns only the derived `ctx` (the `cancel` is unreachable to the caller). The goroutine and the derived context live until `parent` is Done. For queries run under the `ensuredCtx` that `WithConnection` builds from `context.Background()` (`olap.go:63`), `parent` is never Done → the goroutine and context leak permanently. Even for request-scoped parents, the goroutine outlives the query until the parent is cancelled.
- **Failure scenario:** A read-only ClickHouse connector doing `WithConnection` work that runs queries on `ensuredCtx`: each such query leaks a goroutine + context permanently; other queries leak until their parent context is cancelled. Under sustained load this accumulates.
- **Recommendation:** Return the `cancel` and call it once the query finishes (e.g. attach it to the `Result` cleanup), or use `context.WithoutCancel` alone if propagating cancellation isn't required. At minimum, stop deriving from `context.Background()` in a way that never terminates.

**Verification:** CONFIRMED — `runtime/drivers/clickhouse/olap.go:533-540` spawns `go func(){ <-parent.Done(); cancel() }()` and discards `cancel` from the caller's reach; it is invoked from `ExecContext`/`QueryxContext`/`QueryRowContext` at `:513`, `:521`, `:529` whenever `supportSettings == false`. `WithConnection` builds `ensuredCtx` from `context.Background()` at `olap.go:63`, whose `Done()` channel never closes, so goroutines spawned under it leak permanently.

**Manual Reproduction Steps (for QA):**
1. Create a ClickHouse user with `readonly=1` and configure a Rill ClickHouse connector with that user (Rill logs will show settings probes failing; `supportSettings` becomes false).
2. Use the project normally: create models/metrics views on ClickHouse tables and browse dashboards for a while (this exercises `WithConnection`/query paths).
3. Capture goroutine profiles periodically: `curl http://localhost:9091/debug/pprof/goroutine?debug=1` (local runtime debug port).
4. Observed bug: a monotonically growing count of goroutines parked in `contextWithoutDeadline.func1` (`runtime/drivers/clickhouse/olap.go:536`) that never drops, even when the system is idle.

### drivers-olap-06 [LOW] ClickHouse and Pinot mutate the caller's `Statement.Query` in place
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/drivers/clickhouse/olap.go:168`, `runtime/drivers/pinot/olap.go:89`
- **Problem:** `Query` rewrites the input pointer's field: ClickHouse does `stmt.Query += "\n SETTINGS " + sqlSettings`; Pinot does `stmt.Query = fmt.Sprintf("SET timeoutMS=%d; %s", c.timeoutMS, stmt.Query)`. `Statement` is passed by pointer and is not documented as owned by the callee. No current caller in the runtime was found to reuse/retry the same `*Statement`, so the impact today is latent — hence LOW / correctness rather than a live concurrency bug.
- **Failure scenario:** If a caller ever reuses or retries the same `*Statement`, the suffix/prefix is appended again (`... SETTINGS x SETTINGS x`, or `SET timeoutMS=..; SET timeoutMS=..; ..`), producing invalid SQL or double-applied settings. A caller that logs `stmt.Query` after the call also sees the mutated string. Concurrent reads of the same `*Statement` from another goroutine would additionally be a data race.
- **Recommendation:** Build the final SQL in a local variable and pass it to `QueryxContext`; never write back to `stmt.Query`.

**Verification:** CONFIRMED — the in-place mutation exists at both cited sites: `runtime/drivers/clickhouse/olap.go:168` (`stmt.Query += "\n SETTINGS " + sqlSettings`) and `runtime/drivers/pinot/olap.go:89` (`stmt.Query = fmt.Sprintf("SET timeoutMS=%d; %s", ...)`). The harm is latent since no current caller reuses a `*Statement`, matching the LOW rating.

**Manual Reproduction Steps (for QA):**
Not user-reproducible today — no runtime caller retries or reuses the same `*Statement`, so there is no visible misbehavior to trigger. Confirm in code: `runtime/drivers/clickhouse/olap.go:168` and `runtime/drivers/pinot/olap.go:89` write the rewritten SQL back into the caller-owned `stmt.Query` instead of a local variable.

### drivers-olap-07 [MEDIUM] StarRocks scans `INT128`/`LARGEINT` into `sql.NullInt64`
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/drivers/starrocks/olap.go:328` (schema map `:204-205`)
- **Problem:** `databaseTypeToRuntimeType` maps `LARGEINT` → `Type_CODE_INT128` (`:204-205`), but `prepareScanDest` maps `INT128` → `&sql.NullInt64{}` (`:328-329`). StarRocks returns `LARGEINT` over the MySQL wire protocol as a string; scanning a 128-bit value that exceeds int64 range into `NullInt64` invokes `strconv.ParseInt(_, 10, 64)`, which errors.
- **Failure scenario:** Querying a StarRocks table with a `LARGEINT` column containing a value > 2^63-1 returns a scan error (`converting driver.Value ... out of range`) instead of the data, silently breaking that query/dashboard. (`INT8` is also mapped to `NullInt16` at `:322-323`, a lesser width mismatch that carries the value under an `INT8`-typed schema.)
- **Recommendation:** Scan `INT128`/`LARGEINT` into `sql.NullString` (like DECIMAL) or a big-int-aware type and carry it as a string, matching how MySQL/DuckDB handle 128-bit values.

**Verification:** CONFIRMED — `runtime/drivers/starrocks/olap.go:204-205` maps `LARGEINT` → `Type_CODE_INT128`, and `prepareScanDest` at `:328-329` scans `INT64, INT128` into `&sql.NullInt64{}`. A LARGEINT value outside int64 range arrives as a string over the MySQL wire protocol and fails `strconv.ParseInt(..., 10, 64)` inside `NullInt64.Scan`. The secondary `INT8` → `NullInt16` mismatch is also present at `:322-323`.

**Manual Reproduction Steps (for QA):**
1. On a StarRocks cluster, create a table with a LARGEINT column and insert a value greater than 2^63-1, e.g.: `CREATE TABLE t (id INT, big LARGEINT) ...; INSERT INTO t VALUES (1, 170141183460469231731687303715884105727);`
2. Configure a StarRocks connector in a Rill project pointing at that database.
3. Query the table through Rill (e.g. build a model `SELECT * FROM t` or preview the table in the connector explorer).
4. Observed bug: the query fails with a scan error like `converting driver.Value type ... ("170141183460469231731687303715884105727") to a int64: value out of range` instead of returning the row.

### drivers-olap-08 [MEDIUM] Pinot `goValue` returns a parse error as a cell value and can panic on `BOOLEAN`
- **Category:** bug  **Confidence:** medium  **Location:** `runtime/drivers/pinot/sqldriver/driver.go:229-235`
- **Problem:** In the `TIMESTAMP` branch, on `time.Parse` error the code does `return err` — the `error` object becomes the column value returned into `dest[i]` (an invalid `driver.Value`) instead of failing the row. In the `BOOLEAN` branch (`:235`), `r.results.Get(...).(bool)` is an unchecked type assertion (only a nil-check at `:195` precedes it); Pinot's own numeric branches guard against values arriving as strings (`:200-201`, `:206`), so a boolean returned as a non-bool (string/number) would panic inside `Rows.Next`.
- **Failure scenario:** A malformed/unexpected timestamp string yields a result row whose cell is an `error`, corrupting downstream scans; a boolean column returned in an unexpected representation panics the query iterator.
- **Recommendation:** Surface parse failures by returning them from `Next` (not as the value), and use the comma-ok form for the bool assertion, returning nil / an error on mismatch.

**Verification:** CONFIRMED — `runtime/drivers/pinot/sqldriver/driver.go:229-232`: the `TIMESTAMP` branch does `t, err := time.Parse("2006-01-02 15:04:05.0", ...); if err != nil { return err }`, returning the `error` object as the cell value. The `BOOLEAN` branch at `:235` does an unchecked `r.results.Get(rowIdx, coldIdx).(bool)` guarded only by the nil check at `:195`, so a non-bool representation panics. Both defects exist as described; triggering them depends on Pinot returning an unexpected representation.

**Manual Reproduction Steps (for QA):**
1. Configure a Pinot connector in a Rill project against a cluster with a table containing a TIMESTAMP column.
2. Query the table through Rill (model preview or dashboard) with a filter/expression that makes the broker return a timestamp string not matching the exact layout `yyyy-MM-dd HH:mm:ss.S` (e.g. a value with no fractional-second digit, or with more than one).
3. Observed bug: the affected cells scan as garbage/errors downstream instead of the query failing cleanly. For the BOOLEAN panic, a boolean column returned by Pinot as a string/number crashes the query iterator (runtime panic in logs). If your Pinot version always emits the expected formats, this is not externally triggerable — confirm in code at `runtime/drivers/pinot/sqldriver/driver.go:229-235`.

### drivers-olap-09 [LOW] ClickHouse `openHandle` leaks the first `*sql.DB` on HTTP fallback and on metrics-registration failure
- **Category:** bug  **Confidence:** medium  **Location:** `runtime/drivers/clickhouse/clickhouse.go:933-976`
- **Problem:** When the native-protocol `db.Ping` fails with "unexpected packet [72]", the code reassigns `db = sqlx.NewDb(otelsql.OpenDB(...))` (`:951`) without closing the first `db`. Likewise, if `RegisterDBStatsMetrics` (`:971`) returns an error, the freshly opened `db` is returned-around (leaked, `:973`). Each `*sql.DB` owns a background connection-opener goroutine and pooled conns until `Close`.
- **Failure scenario:** Connecting to an HTTP ClickHouse port via the native default protocol leaks the native `*sql.DB` on every open; combined with reconcile retries this accumulates.
- **Recommendation:** Close the previous `db` before reassigning on the HTTP retry, and close `db` before returning the `RegisterDBStatsMetrics` error.

**Verification:** CONFIRMED — `runtime/drivers/clickhouse/clickhouse.go:951` reassigns `db = sqlx.NewDb(...)` on the "unexpected packet [72]" fallback without closing the first native-protocol `db`, and the `RegisterDBStatsMetrics` error path at `:971-973` returns without closing the freshly opened `db`. (All the earlier `Ping`-failure returns at `:940`, `:947`, `:957`, `:959` also leak `db`, making the finding slightly broader than stated.)

**Manual Reproduction Steps (for QA):**
1. Configure a Rill ClickHouse connector with `host` and `port` pointing at the ClickHouse HTTP port (8123) without setting a DSN or protocol (so the driver first tries native protocol).
2. Start Rill; the log shows the warning "ClickHouse connection was established with the HTTP protocol, consider using the native port for better performance", proving the fallback path ran.
3. Force repeated re-opens (edit and re-save the connector YAML several times).
4. Observed bug: each open leaks the abandoned native-protocol `*sql.DB`. Confirm via the runtime's `/debug/pprof/goroutine` profile showing an extra `database/sql.(*DB).connectionOpener` goroutine accumulating per open.

### drivers-olap-10 [LOW] ClickHouse read-only deadline stripping is bypassed for single-row queries; the `QueryRowContext` override is dead code
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/drivers/clickhouse/olap.go:525-531` vs `information_schema.go:403,507`
- **Problem:** `SQLConn` overrides `ExecContext`/`QueryxContext` to strip context deadlines when `supportSettings == false` (so the clickhouse-go driver doesn't auto-set `max_execution_time`, which errors in read-only mode). But the single-row meta queries in `LoadDDL` (`information_schema.go:403`) and `entityType` (`:507`) call the *embedded* `sqlx.Conn.QueryRowxContext` (a method `SQLConn` does not override), so they bypass the stripping. The `SQLConn.QueryRowContext` override (`:525`, returning `*sqlx.Row`) shadows the promoted `sql.Conn.QueryRowContext` by name but is never called by any code in the driver — dead code.
- **Failure scenario:** On a read-only cluster, `LoadDDL`/`entityType` (which use `QueryRowxContext`) can fail with "Cannot modify max_execution_time setting in readonly mode" whenever the caller's context carries a deadline.
- **Recommendation:** Add/route a `QueryRowxContext` override that strips the deadline (or centralize the stripping in `acquireConn` by wrapping the context once), and remove the unused `QueryRowContext` method.

**Verification:** CONFIRMED — `SQLConn` (`runtime/drivers/clickhouse/olap.go:504-531`) overrides `ExecContext`, `QueryxContext`, and `QueryRowContext` but not `QueryRowxContext`. `LoadDDL` calls `conn.QueryRowxContext` at `information_schema.go:403` and `entityType` at `:507`; both resolve to the embedded `sqlx.Conn` method, bypassing deadline stripping. A grep confirms `SQLConn.QueryRowContext` has no caller anywhere in the driver (the only other `QueryRowContext` use, `crud.go:428`, is on `c.writeDB`, not on `SQLConn`) — dead code as described.

**Manual Reproduction Steps (for QA):**
1. Configure a Rill ClickHouse connector with a `readonly=1` user (a cluster where settings cannot be modified).
2. In Rill Developer, open the connector's table explorer and view a table's details/DDL (this triggers `LoadDDL`, which runs `SHOW CREATE TABLE` via `QueryRowxContext` under a request context that carries a deadline).
3. Observed bug: the request fails with an error like "Cannot modify max_execution_time setting in readonly mode", while plain data queries against the same connector succeed (because their code path strips the deadline).

### drivers-olap-11 [LOW] Pinot query context is not propagated to the in-flight broker request
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/drivers/pinot/sqldriver/driver.go:63-77`
- **Problem:** `QueryContext` calls `c.pinotConn.ExecuteSQLWithParams("", query, params)` / `ExecuteSQL`, which are blocking and ignore `ctx` (`// TODO: cancel the query if ctx is done`). The retrier respects `ctx` only between attempts (`:65`).
- **Failure scenario:** A cancelled/timed-out request (user navigates away, `ExecutionTimeout` elapses) does not stop the running Pinot broker query; the goroutine blocks until Pinot responds, wasting a pooled connection and broker capacity.
- **Recommendation:** Run the Pinot client call in a goroutine and select on `ctx.Done()`, or use a Pinot client API that accepts a context/timeout, returning `ctx.Err()` promptly. (Contrast Druid, which issues a DELETE via `context.AfterFunc` in `druidsqldriver/druid_api_sql_driver.go:116`.)

**Verification:** CONFIRMED — `runtime/drivers/pinot/sqldriver/driver.go:63-77`: `QueryContext` calls the blocking `c.pinotConn.ExecuteSQLWithParams("", query, params)` / `ExecuteSQL` with no context plumbing, and the code carries the literal comment `// TODO: cancel the query if ctx is done` at `:73`. Only the retrier wrapper (`re.RunCtx`) consults `ctx`, and only between attempts.

**Manual Reproduction Steps (for QA):**
1. Configure a Rill Pinot connector and build an explore dashboard over a large Pinot table.
2. Trigger a heavy query (e.g. a high-cardinality leaderboard over the full time range), then immediately navigate away or refresh the page to cancel the request.
3. Watch the Pinot broker's query console or logs.
4. Observed bug: the query keeps running on the broker until it finishes on its own; cancellation in Rill does not abort it (contrast Druid, where cancellation issues a DELETE to the server).

### drivers-olap-12 [LOW] Config values are interpolated into SQL without escaping (StarRocks catalog, MotherDuck token)
- **Category:** security  **Confidence:** medium  **Location:** `runtime/drivers/starrocks/starrocks.go:362,379`; `runtime/drivers/duckdb/duckdb.go:513`
- **Problem:** StarRocks `initDB` builds `SELECT 1 FROM <catalog>.information_schema.schemata ...` (`:362`) and `... FROM <catalog>.information_schema.schemata WHERE ...` (`:379`) by `fmt.Sprintf` of the raw `catalog` config value (the database value at `:380` is correctly parameterized). DuckDB `reopenDB` builds `SET motherduck_token = '<token>'` by raw interpolation (`:513`), even though a `safeSQLString` helper is used just below for `extension_directory` (`:559`). Both are operator-supplied config (self-inflicted, not cross-tenant), but a value containing a quote breaks the statement, and there is no defense-in-depth.
- **Failure scenario:** A catalog name or token containing `'`/`.`/whitespace produces malformed SQL or, in principle, injected SQL run with the connector's privileges.
- **Recommendation:** Escape identifiers via the dialect's identifier-escaping and string values via `EscapeStringValue`/`safeSQLString` before interpolation.

**Verification:** CONFIRMED — StarRocks `initDB` interpolates the raw `catalog` config into SQL via `fmt.Sprintf` at `runtime/drivers/starrocks/starrocks.go:362` and `:379` (while the database value at `:380` is properly parameterized with `?`). DuckDB `reopenDB` interpolates the raw MotherDuck token at `runtime/drivers/duckdb/duckdb.go:513` (`fmt.Sprintf("SET motherduck_token = '%s'", c.config.Token)`), while `safeSQLString` is used just below at `:559` for `extension_directory`.

**Manual Reproduction Steps (for QA):**
1. In a Rill project, configure a StarRocks connector with a catalog value containing a quote, e.g. `catalog: "my'catalog"` in the connector YAML.
2. Start Rill; the connector fails with a StarRocks SQL syntax error (malformed `SELECT 1 FROM my'catalog.information_schema.schemata ...`) rather than a clean "catalog not found" message.
3. Similarly, configure a MotherDuck connector with a token containing a single quote (e.g. append `'` to a valid token); DuckDB init fails with a SQL parser error from the `SET motherduck_token` statement instead of an authentication error.

### drivers-olap-13 [LOW] Druid `Lookup` passes an unused argument to a placeholder-less readiness query
- **Category:** correctness  **Confidence:** low  **Location:** `runtime/drivers/druid/information_schema.go:92-93`
- **Problem:** The coordinator-readiness probe `c.db.QueryxContext(ctx, "SELECT * FROM sys.segments LIMIT 1", name)` passes `name` although the query has no placeholder. The Druid `database/sql` conn implements `QueryerContext` (`druidsqldriver/druid_api_sql_driver.go:55,100`), so `database/sql` skips the `NumInput` arg-count check and forwards `name` into `newDruidRequest`'s `parameters` array (`:477-497`) with no binding site.
- **Failure scenario:** Depending on Druid's tolerance of extra parameters, the probe may error or behave unexpectedly; at best it sends a meaningless parameter on every `Lookup`. Almost certainly a copy-paste leftover from the real INFORMATION_SCHEMA query below it.
- **Recommendation:** Drop the `name` argument from the readiness query.

**Verification:** CONFIRMED — `runtime/drivers/druid/information_schema.go:92-93`: `q := "SELECT * FROM sys.segments LIMIT 1"` followed by `c.db.QueryxContext(ctx, q, name)`. The Druid conn implements `driver.QueryerContext` (`druidsqldriver/druid_api_sql_driver.go:55`), so `database/sql` forwards the extra arg into `newDruidRequest`, which marshals it into the request's `parameters` array (`:478-482`, `:497`) with no placeholder to bind to.

**Manual Reproduction Steps (for QA):**
1. Configure a Rill Druid connector and open the table explorer or query a Druid-backed metrics view (any path that calls `Lookup`).
2. Capture the HTTP request Rill sends to the Druid router/broker `/druid/v2/sql` endpoint (via a proxy or Druid request logging).
3. Observed: the `SELECT * FROM sys.segments LIMIT 1` readiness probe carries a non-empty `parameters` array containing the table name even though the SQL has no `?` placeholder. Whether Druid errors or ignores it depends on the Druid version; the stray parameter itself is the defect.

### drivers-olap-14 [LOW] Pinot `All` logs per-table errors with `fmt.Printf`
- **Category:** readability  **Confidence:** high  **Location:** `runtime/drivers/pinot/information_schema.go:276`
- **Problem:** A schema-fetch failure is reported via `fmt.Printf(...)` to stdout instead of the injected `c.logger`, bypassing structured logging and log levels. (Also related to the swallowed error in drivers-olap-01. Note the sibling `LoadPhysicalSize` at `:324-351` already uses `c.logger.Warn(...)` correctly, so this is an inconsistency.)
- **Recommendation:** Use `c.logger.Warn(...)` with structured fields and `observability.ZapCtx(ctx)`.

**Verification:** CONFIRMED — `runtime/drivers/pinot/information_schema.go:276`: `fmt.Printf("Error fetching schema for table %s: %v\n", tableName, err)` inside the `All` errgroup, while the sibling `LoadPhysicalSize` at `:324-351` correctly uses `c.logger.Warn(..., observability.ZapCtx(ctx))`.

**Manual Reproduction Steps (for QA):**
1. Same setup as drivers-olap-01: Pinot connector with one table whose `/schemas/<table>` controller endpoint fails (controller restarting or blocked).
2. Refresh the table listing in Rill Developer.
3. Observed: the message "Error fetching schema for table ...: ..." appears on raw stdout of the `rill start` process, not in Rill's structured/leveled log output (note the process may then panic per drivers-olap-01 before you can compare).

### drivers-olap-17 [LOW] ClickHouse `otelsql.RegisterDBStatsMetrics` registers an un-unregisterable stats callback per `Open`
- **Category:** bug  **Confidence:** medium  **Location:** `runtime/drivers/clickhouse/clickhouse.go:971`
- **Problem:** `openHandle` calls `otelsql.RegisterDBStatsMetrics(db.DB, ...)`, which internally does `meter.RegisterCallback(...)` and **discards the returned `Registration`** (otelsql v0.27.0 `sql.go:120` — `_, err = ...`). The function's public signature returns only `error`, so the caller has no handle to unregister. Every `Open` (including every successful reconnect / cache eviction / config change) therefore leaves a permanent OTEL callback whose closure captures that specific `*sql.DB`. `Close` (`:471-492`) closes `readDB`/`writeDB` but cannot remove these callbacks, so the closed `*sql.DB` objects are retained by the meter and can never be garbage-collected.
- **Failure scenario:** On a long-running Rill Cloud runtime, connection churn (instance edits, LRU evictions, scale-to-zero reconnects) accumulates one leaked stats callback + retained closed `*sql.DB` per open, growing unboundedly over time. This is distinct from drivers-olap-09 (which is only the failure-path `*sql.DB` leak); this one occurs on the success path too and is not cleanable by the current code.
- **Recommendation:** Register DB-stats metrics once per unique (instance, connector) rather than per `Open`, or vendor a variant of `RegisterDBStatsMetrics` that returns the `metric.Registration` so `Close` can `Unregister()` it.

**Verification:** CONFIRMED — `runtime/drivers/clickhouse/clickhouse.go:971` calls `otelsql.RegisterDBStatsMetrics(db.DB, ...)` on every `openHandle`. Inspected the vendored dependency (`go.mod` pins `github.com/XSAM/otelsql v0.27.0`): `RegisterDBStatsMetrics` does `_, err = meter.RegisterCallback(...)`, discarding the `metric.Registration`, and returns only `error` — so `Connection.Close` (`:471-492`) has no way to unregister, and the callback closure retains the closed `*sql.DB` forever.

**Manual Reproduction Steps (for QA):**
Not directly user-visible. To observe on a running system: repeatedly edit and re-save a ClickHouse connector's YAML (forcing close/reopen cycles), then inspect the runtime's exported OTEL metrics or a heap profile — duplicate `db.sql.connection.*` callback registrations and retained closed `*sql.DB` objects accumulate one per reopen. Confirm in code: `runtime/drivers/clickhouse/clickhouse.go:971` plus `RegisterDBStatsMetrics` in otelsql v0.27.0 `sql.go` (registration discarded).

## Product Improvements

### drivers-olap-15 [LOW] `AllFromInformationSchema` hard-caps at 1000 tables per schema
- **Category:** product-improvement  **Confidence:** high  **Location:** `runtime/drivers/information_schema.go:68-100`
- **Problem:** The shared helper requests `ListTables(..., 1000, "")` per schema (`:78`) and returns an error (`schema has more than 1000 tables can not list all`) if a single schema exceeds 1000 tables (`:82-85`). It *does* paginate across database schemas by forwarding the outer `ListDatabaseSchemas` token (`:72`, returned at `:99`), but it cannot paginate table listing *within* a schema. Callers are `postgres`, `mysql`, `snowflake`, `bigquery`, `redshift`, `databricks`, `athena` (ClickHouse/Druid/Pinot have their own `All` implementations and are not affected). A large Postgres/MySQL/warehouse schema (>1000 tables) fails the "list all" path outright.
- **Recommendation:** Thread real pagination through (continue table pagination within a schema, resuming across schemas via a compound token) or stream tables rather than materializing all of them.

**Verification:** CONFIRMED — `runtime/drivers/information_schema.go:78` calls `i.ListTables(ctx, ..., 1000, "")` and `:82-85` returns the error `"schema has more than 1000 tables can not list all"` when the inner token is non-empty; the outer `ListDatabaseSchemas` token is forwarded/returned (`:72`, `:99`). Grep confirms the callers are exactly databricks, bigquery, snowflake, redshift, postgres, mysql, athena — not ClickHouse/Druid/Pinot.

**Manual Reproduction Steps (for QA):**
1. Create a Postgres database with a schema containing more than 1000 tables (script: `for i in $(seq 1 1001); do psql -c "CREATE TABLE bulk.t$i (id int);"; done`).
2. Configure a Postgres connector in a Rill project and open the connector's table explorer ("list all tables" view).
3. Observed: the listing fails with the error "schema has more than 1000 tables can not list all" instead of paginating.

### drivers-olap-16 [LOW] Pinot silently drops tables whose schema can't be fetched
- **Category:** product-improvement  **Confidence:** medium  **Location:** `runtime/drivers/pinot/information_schema.go:273-281`
- **Problem:** Beyond the panic in drivers-olap-01, the design intentionally swallows per-table errors, so a user sees an incomplete table list with no indication that some tables were skipped.
- **Recommendation:** Return partial results but attach a warning/diagnostic (or fail loudly for non-transient errors) so users understand why tables are missing.

**Verification:** CONFIRMED — `runtime/drivers/pinot/information_schema.go:273-281`: the `Lookup` error is swallowed (`return nil` after `fmt.Printf`), so the design intentionally drops failed tables with no user-facing signal. (In current code the nil slot causes the panic in drivers-olap-01 first; once that is fixed by compacting nils, this silent-drop behavior is what remains.)

**Manual Reproduction Steps (for QA):**
1. Same setup as drivers-olap-01: Pinot connector, one table whose controller `/schemas` endpoint transiently fails.
2. Refresh the table listing (after the -01 panic is fixed; today the panic masks this).
3. Observed: the affected table is silently missing from the list with no warning shown in the UI or API response.

## New Feature Ideas
- **Uniform driver "size + row-count" surface (S/M):** DuckDB/ClickHouse implement `EstimateSize`; Druid/Pinot/MySQL/Postgres/StarRocks return `-1`. A small `LoadPhysicalSize`-style estimate (Druid `sys.segments` and Pinot `/debug/tables` already exist — Pinot's `LoadPhysicalSize` is even implemented but untested) surfaced consistently would give users storage insight across all OLAP connectors.
- **Read-only connection validation on Open (S):** Several drivers only discover permission/mode problems at first query. A cheap capability probe at `Open` (settings-writable, table-list readable) surfaced as a clear connector error would improve the setup experience, especially for read-only ClickHouse/StarRocks users. (ClickHouse already computes `supportSettings` at Open — surfacing it would be cheap.)
- **Query-cancellation conformance (M):** Standardize context-aware cancellation across the HTTP drivers (Druid already issues a DELETE via `context.AfterFunc`; Pinot has a TODO — see drivers-olap-11). A shared "cancel in-flight on ctx.Done" helper would make timeouts/`ExecutionTimeout` behave consistently.
- **Structured dialect conformance tests (M):** A shared golden-file test matrix that runs the same logical query (unnest, date_trunc, safe-divide, ILIKE, regex) through every dialect would catch per-dialect SQL divergence that today is only exercised ad hoc.
- **Driver metrics parity + registration hygiene (S):** DuckDB/ClickHouse emit queue/query/total latency + connections-in-use; Druid/Pinot/MySQL/Postgres/StarRocks emit little. Extending the latency histograms to all drivers (and fixing the DuckDB gauge race in -02 and the ClickHouse stats-callback leak in -17) would make pool saturation and slow-query diagnosis uniform and leak-free.

## Notes & Missing Test Coverage
- **DuckDB reopen/close lifecycle** (`triggerReopen`, `acquireDB` counting, registration cleanup on `Open` failure) has no test exercising the failure paths in -02/-04; a `-race` test around concurrent `acquireDB` + metric collection would catch the `dbConnCount` race.
- **ClickHouse `Open` failure paths** (-03), `openHandle` HTTP fallback (-09), and the stats-callback registration leak (-17) are untested for resource cleanup; hard to unit-test without a server, but the embedded-server ref-count leak is testable with a fake `start`/`stop`.
- **Pinot `All` with a failing `Lookup`** (-01) is untested and is a straightforward table-driven test (mock the controller to 500 for one table) that would have caught the nil panic.
- **StarRocks `LARGEINT`/`INT128` scanning** (-07) lacks coverage; a fixture row with a >int64 value would demonstrate the scan failure.
- **Pinot `goValue`** timestamp-parse-error and boolean-type-mismatch branches (-08) are untested.
- Static review only (no runtime execution). The `Statement.Query` in-place mutation (-06) impact depends on whether any caller reuses/retries a `*Statement`; none was found in the runtime, so it is rated low/medium. The Druid extra-arg behavior (-13) depends on Druid's server-side tolerance, which could not be run.

## Verification Notes
An independent re-review opened and traced every cited location. All 16 original findings were confirmed in code; none were removed as false positives. The following were corrected or refined:

- **drivers-olap-04:** The original claimed a leaked goroutine and a "live connection." Verified that `periodicallyEmitStats`/`periodicallyCheckConnDurations` start only *after* `reopenDB` succeeds (`duckdb.go:231-233`), so the failure path leaks no goroutine, and `c.db` is nil after a failed `reopenDB`, so no DB pool leaks. Rewritten to focus on the genuine leak: the metrics-callback registration that retains `c`. Severity/confidence unchanged (MEDIUM/high).
- **drivers-olap-05:** Confidence lowered high→medium. The permanent-leak vector is confirmed for queries run on `ensuredCtx` (built from `context.Background()` at `olap.go:63`); for request-scoped parents the goroutine self-cleans on cancellation, so the "unbounded under sustained load" magnitude is conditional on the `ensuredCtx` path being exercised on a read-only cluster.
- **drivers-olap-06:** Downgraded MEDIUM→LOW and re-categorized concurrency→correctness. Searched the runtime for any caller that reuses/retries the same `*Statement`; none found, so the concurrent-race scenario is speculative and the realistic risk is double-application on reuse / mutated `stmt.Query` observed by the caller. The in-place mutation itself is confirmed at both locations.
- **drivers-olap-10:** Fully traced. Confirmed `LoadDDL` (`information_schema.go:403`) and `entityType` (`:507`) call the embedded `QueryRowxContext` and bypass deadline stripping, and that `SQLConn.QueryRowContext` has no caller (dead code). Location and rating retained (LOW/medium).
- **drivers-olap-15:** Corrected the affected-driver list — the original cited "big ClickHouse/Druid deployments," but those drivers do not use `AllFromInformationSchema`; the real callers are postgres/mysql/snowflake/bigquery/redshift/databricks/athena. Also clarified that cross-*schema* pagination does work (outer token forwarded); the hard limit is per-schema table pagination.
- **drivers-olap-17 (added):** New, verified. `otelsql.RegisterDBStatsMetrics` (v0.27.0 `sql.go:111-136`) registers a meter callback and discards the `Registration`, and `Connection.Close` has no way to unregister it — a per-`Open` accumulating leak on the success path, distinct from -09.

**Removed:** 0 findings. **Refined:** 6 (04, 05, 06, 10, 14, 15). **Added:** 1 (17). Locations spot-checked against the current tree; no invented line numbers were found in the surviving set, and the two location refs in -02 (`:730,:744`) and -03 (`:308-389`) were confirmed accurate.
