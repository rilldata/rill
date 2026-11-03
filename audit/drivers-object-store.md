# Object Store & Warehouse Drivers — Code Review

**Scope:** `runtime/drivers/{s3,gcs,azure,athena,bigquery,redshift,snowflake,databricks,salesforce,https,file}`, `runtime/drivers/{object_store,warehouse,file_store}.go`, `runtime/pkg/blob`, `runtime/pkg/globutil`, `runtime/pkg/gcputil`, `runtime/pkg/fileutil` (path helpers), plus the callers in `runtime/ai`, `runtime/server/connector_service.go`, `runtime/resolvers/glob.go`.  |  **Health:** needs-attention

## Overview

These drivers glob-list, download, and unload data from object stores and warehouses into local temp files for ingestion, and expose bucket browsing to the AI assistant and the connector UI. The shared blob download pipeline (`pkg/blob/download.go`) is carefully written and its channel/goroutine choreography holds up. The `path_prefixes` access-scoping feature, however, is only half-enforced: the non-glob `ListObjects` path checks it, but the `ListObjectsForGlob` and `DownloadFiles` paths bypass it entirely, and even the enforced path uses non-boundary `strings.HasPrefix` checks that leak sibling prefixes. The biggest concrete defects are a Redshift `UNLOAD` that fails on any query containing a quote, a leaked BigQuery client on every successful ingest, and unsanitized object-store keys written straight into local temp paths. The warehouse drivers share good structure but repeat a temp-file cleanup bug that orphans S3 unload output when `Close()` errors. Overall the code is readable, but the security-boundary logic and several error/cleanup paths need attention.

## Findings

### drivers-object-store-01 [HIGH] Redshift `UNLOAD` breaks on any query containing a single quote
- **Category:** correctness  **Confidence:** high  **Location:** `runtime/drivers/redshift/warehouse.go:120`
- **Problem:** `finalSQL := fmt.Sprintf("UNLOAD ('%s') TO '%s/' IAM_ROLE '%s' FORMAT AS PARQUET", sourceProperties.SQL, ...)` embeds the user's query inside a single-quoted Redshift string literal without doubling internal single quotes. Redshift requires single quotes inside the `UNLOAD ('...')` query string to be escaped by doubling. The same package already ships `escapeStringValue` (`information_schema.go:206`, which does `strings.ReplaceAll(s, "'", "''")`) but it is not used here. Athena's equivalent (`athena/warehouse.go:123`) is unaffected because Athena's `UNLOAD (query)` is not quoted.
- **Failure scenario:** A perfectly ordinary model query such as `SELECT * FROM orders WHERE country = 'US'` produces `UNLOAD ('SELECT * FROM orders WHERE country = 'US'') TO ...`. The embedded `'US'` terminates the string literal, so Redshift raises a syntax error and the model fails to ingest. Nearly every real query with a string literal, date literal, or `LIKE` pattern hits this.
- **Recommendation:** Escape the query before interpolating: `strings.ReplaceAll(sourceProperties.SQL, "'", "''")` (reuse `escapeStringValue`-style logic). Add a test with a query containing a literal string.

**Verification:** CONFIRMED — `runtime/drivers/redshift/warehouse.go:120` interpolates `sourceProperties.SQL` directly into a single-quoted `UNLOAD ('%s')` string with no quote doubling; `escapeStringValue` exists at `information_schema.go:206` and is used elsewhere in that file but not here. Athena's variant at `warehouse.go:123` is unquoted (`UNLOAD (%s)`), confirming the contrast. Any SQL containing a single quote breaks the literal.

**Manual Reproduction Steps (for QA):**
1. Configure a Redshift connector in a Rill project (`connectors/redshift.yaml`) with valid `role_arn`, `workgroup`/`cluster_identifier`, and `output_location`.
2. Create a model that ingests from Redshift with a query containing an ordinary string literal, e.g. `sql: SELECT * FROM orders WHERE country = 'US'`.
3. Trigger a reconcile/refresh of the model (save the file, or `rill project refresh`).
4. Observe the ingest fails with a Redshift syntax error; the generated `UNLOAD ('SELECT ... 'US'')` is malformed. A query with no quotes succeeds, confirming the quote is the trigger.

### drivers-object-store-02 [MEDIUM] `path_prefixes` access scoping is bypassed by glob listing and downloads
- **Category:** security  **Confidence:** high  **Location:** `runtime/drivers/s3/object_store.go:81` (`ListObjectsForGlob`) and `:101` (`DownloadFiles`); identical in `gcs/object_store.go:88`/`:94` and `azure/object_store.go:94`/`:98`
- **Problem:** `path_prefixes` is documented as "a list of bucket path prefixes that this connector is allowed to access," and `ListObjects`/`ListBuckets` enforce it via `drivers.ListObjects(ctx, c.config.PathPrefixes, ...)`. But `ListObjectsForGlob` calls `blobBucket.ListObjectsForGlob(...)` directly with no prefix argument, and `DownloadFiles` calls `bucket.Download(...)` with no prefix check. Verified across all three drivers. The AI tool (`runtime/ai/list_bucket_objects.go:96-100`) and the gRPC `ListObjects` handler (`runtime/server/connector_service.go:47-53`) choose between the two purely based on whether the input string contains a glob metacharacter, so a caller escapes the scoping simply by including a `*`.
- **Failure scenario:** A connector is configured with `path_prefixes: [s3://bucket/allowed/]` to restrict what the assistant/UI can browse. A request for glob `secret/**` (or a direct `DownloadFiles("s3://bucket/secret/…")`) returns/reads objects under `secret/`, which the `path_prefixes` restriction was meant to forbid. The IAM credentials still gate raw access, but the intended per-connector logical boundary is defeated.
- **Severity note (revised down from HIGH):** The entry points are all gated to developer-level access (`checkDeveloperAccess` for the AI tool; connector access for the gRPC handler), and the config comments frame `path_prefixes` partly as a credential-routing convenience ("route access through the appropriate connector based on the blob path"), not solely a hard security boundary. It becomes a genuine privilege boundary only when the connector (and its credentials) is managed at a higher level than the actor editing project YAML — in that case a developer/steered-AI can read outside the admin-set prefix using the managed credentials. Real bypass, but not an unprivileged-actor / cross-tenant escalation, hence MEDIUM.
- **Recommendation:** Enforce `path_prefixes` in `ListObjectsForGlob` and `DownloadFiles` too (validate the glob's literal prefix / the download path against the allowed prefixes before listing/downloading), or centralize the check so no object-store entry point can skip it. Use delimiter-boundary matching (see finding 05).

**Verification:** CONFIRMED — `s3/object_store.go:81` (`ListObjectsForGlob`) calls `blobBucket.ListObjectsForGlob(...)` directly and `:101` (`DownloadFiles`) calls `bucket.Download(...)` with no `PathPrefixes` argument, while `ListObjects` at `:70` passes `c.config.PathPrefixes` into `drivers.ListObjects`. Same pattern in `gcs/object_store.go:88` and `azure/object_store.go:94`. The AI tool routes on `fileutil.IsGlob(args.Path)` (`runtime/ai/list_bucket_objects.go:96-99`) and the gRPC handler on presence of `req.Glob` (`runtime/server/connector_service.go:48-53`), so a `*` in the path selects the unchecked path.

**Manual Reproduction Steps (for QA):**
1. Configure an S3 (or GCS/Azure) connector with `path_prefixes: ["s3://mybucket/allowed/"]` and credentials that can read the whole bucket.
2. Place a test object under a sibling prefix, e.g. `s3://mybucket/secret/data.parquet`.
3. Ask the AI assistant to browse the bucket using a glob path that includes a metacharacter, e.g. `secret/**` on `mybucket` (the assistant's list-bucket tool). Alternatively define a model source that downloads `s3://mybucket/secret/**`.
4. Observe the `secret/` objects are listed/downloaded despite not being under the configured `allowed/` prefix. Repeat step 3 with a non-glob path `secret/` and observe it is correctly rejected by `path_prefixes` — confirming only the glob path bypasses the check.

### drivers-object-store-03 [HIGH] BigQuery client is leaked on every successful `QueryAsFiles`

**Status (2026-07-25):** ✅ Fixed and merged in [#9706](https://github.com/rilldata/rill/pull/9706).

- **Category:** bug  **Confidence:** high  **Location:** `runtime/drivers/bigquery/warehouse.go:213`
- **Problem:** `QueryAsFiles` opens a `*bigquery.Client` (with a Storage Read gRPC client) and stores it in `fileIterator.client`. `client.Close()` is only called on the error paths inside `QueryAsFiles`. The iterator's `Close()` does only `os.RemoveAll(f.tempDir)` and `Next()` never closes the client either, so on every successful ingest the client and its underlying gRPC connections/goroutines are never released. Compare Snowflake/Databricks, whose iterators close `db`/`conn` in both `Next` and `Close`.
- **Failure scenario:** A project with several BigQuery-backed models refreshing on a schedule leaks one client per refresh, accumulating gRPC connections and goroutines until the runtime exhausts file descriptors or memory over long uptimes.
- **Recommendation:** Close `f.client` in `fileIterator.Close()` (and in the early-close defer in `Next()`), guarding against double close.

**Verification:** CONFIRMED — `runtime/drivers/bigquery/warehouse.go:213` `fileIterator.Close()` does only `return os.RemoveAll(f.tempDir)`; `f.client` (stored at `:191-197`) is closed only on error paths inside `QueryAsFiles` and never in `Next()` or `Close()`. On a successful ingest the `*bigquery.Client` and its Storage Read gRPC client leak. Snowflake and Databricks iterators close their handles in both `Next` and `Close`, confirming the omission is a real leak.

**Manual Reproduction Steps (for QA):**
Not user-visible in a single action (symptom is slow FD/goroutine accumulation). Reproduce by load: configure a BigQuery-backed model and schedule frequent refreshes (e.g. every minute) over a long period, then watch the runtime process's open file descriptor / goroutine count grow without bound. Confirm in code: `runtime/drivers/bigquery/warehouse.go:213` — `Close()` never calls `f.client.Close()`; contrast `snowflake/warehouse.go` / `databricks/warehouse.go` iterators which close their connections in both `Next` and `Close`.

### drivers-object-store-04 [MEDIUM] Object-store keys are joined into local paths without sanitization (zip-slip)
- **Category:** security  **Confidence:** medium  **Location:** `runtime/pkg/blob/download.go:230`
- **Problem:** `filename := filepath.Join(it.tempDir, obj.Path)` uses the raw object key as a relative local path to preserve hive-partition hierarchy. `filepath.Join` cleans `..` segments, so a key containing `../` resolves *outside* `tempDir`. There is no check that the result stays within `tempDir`.
- **Failure scenario:** A bucket the connector reads contains an object whose key is `../../../../tmp/rill/<other-instance>/evil.parquet` (object stores permit `..` in keys). The download writes to `/tmp/rill/<other-instance>/evil.parquet`, escaping the per-ingest temp dir and potentially overwriting another instance's files or planting files the runtime later ingests.
- **Recommendation:** After `filepath.Join`, verify `strings.HasPrefix(filename, filepath.Clean(it.tempDir)+string(os.PathSeparator))` and reject keys that escape; or sanitize keys (reject `..` components) before joining.

**Verification:** PLAUSIBLE — `runtime/pkg/blob/download.go:230` does `filename := filepath.Join(it.tempDir, obj.Path)` with the raw object key and no confinement check before `os.MkdirAll`/`os.OpenFile`. `filepath.Join` cleans `..`, so a key with leading `../` segments resolves outside `tempDir`. Latent because it requires an object whose key literally contains `..` path segments; `obj.Path` comes from the glob listing of the configured bucket, so exploitation needs a bucket containing such a key (achievable in object stores that permit `..` in keys, but not a typical setup).

**Manual Reproduction Steps (for QA):**
1. In a test S3/GCS bucket, upload an object whose key contains parent-directory segments, e.g. `../../../../tmp/rill-zipslip-test/evil.parquet` (use the raw S3 API / `aws s3api put-object --key`, since normal tools may normalize the key).
2. Configure a connector for that bucket and a model that globs to include the object (e.g. `s3://bucket/**`).
3. Trigger ingestion and check whether a file appears at `/tmp/rill-zipslip-test/evil.parquet` (outside the per-ingest `blob_ingestion*` temp dir).
4. If it does, the key escaped confinement. Confirm in code at `runtime/pkg/blob/download.go:230` — no `HasPrefix(tempDir)` guard after the join.

### drivers-object-store-05 [MEDIUM] `path_prefixes` enforcement uses non-boundary prefix matches
- **Category:** security  **Confidence:** high  **Location:** `runtime/drivers/object_store.go:177`
- **Problem:** In `drivers.ListObjects`, allowed prefixes are compared with `strings.HasPrefix(path, ap)` where `ap` is the configured prefix with only its leading slash trimmed (`buildAllowedPrefixesForBucket`, line 260, does not normalize a trailing delimiter). So `s3://bucket/data` yields `ap = "data"` and matches any path that merely starts with the string `data`, not just the `data/` "directory." The nested-prefix reduction at line 276 has the same non-boundary flaw.
- **Failure scenario:** Config allows `s3://bucket/data`. A request to list `database/` satisfies `strings.HasPrefix("database/", "data")`, so the caller lists a sibling prefix that was never meant to be in scope.
- **Recommendation:** Normalize allowed prefixes to end in the delimiter and compare on delimiter boundaries (treat an exact `path == ap` and `path` under `ap + "/"` as the allowed cases), rather than raw string prefix.

**Verification:** CONFIRMED — `runtime/drivers/object_store.go:177` uses `strings.HasPrefix(path, ap)` and `:180` `strings.HasPrefix(ap, path)`; `buildAllowedPrefixesForBucket` (`:253-287`) only trims the leading `/` (`:260`) and never appends a trailing delimiter, and the nested-prefix reduction at `:276` uses the same raw `strings.HasPrefix`. So allowed prefix `data` matches requested path `database/`.

**Manual Reproduction Steps (for QA):**
1. Configure an S3 connector with `path_prefixes: ["s3://mybucket/data"]` (note: no trailing slash).
2. Ensure the bucket has a sibling prefix such as `s3://mybucket/database/` containing objects.
3. Ask the AI assistant / connector browse UI to list path `database/` on `mybucket` (non-glob).
4. Observe the objects under `database/` are listed even though only `data` was allowed, because `strings.HasPrefix("database/", "data")` is true.

### drivers-object-store-06 [MEDIUM] `ResolveLocalPath` root confinement uses a non-boundary prefix check
- **Category:** security  **Confidence:** high  **Location:** `runtime/pkg/fileutil/fileutil.go:328`
- **Problem:** The file connector confines paths with `if !allowHostAccess && !strings.HasPrefix(finalPath, root)`. When `path` is absolute, `finalPath = path` unchanged, and the check passes for any absolute path that shares a string prefix with `root` without a separator boundary.
- **Failure scenario:** With `root = /var/data/proj` and `allow_host_access=false`, a model `path: /var/data/proj-other/secrets.csv` passes `strings.HasPrefix("/var/data/proj-other/secrets.csv", "/var/data/proj")` and is read, escaping the intended project root and reading a sibling project's files in a multi-tenant layout.
- **Recommendation:** Compare against `root + string(os.PathSeparator)` (or use `filepath.Rel` and reject results starting with `..`).
- **Verified:** `root` is set from `filepath.Abs` (`file/file.go:89-96`), which cleans trailing separators, so the non-boundary match is real. Reachable via the `file` connector's `FilePaths` → `resolveLocalPath` (`file/file_store.go:18,36`). Practical exploitation is narrow: it requires `allow_host_access=false`, an absolute (or `..`-containing) model `path`, and a sibling directory whose name shares `root` as a string prefix (e.g. `/data/proj` vs `/data/proj-other`). Treat as a defense-in-depth boundary fix.

**Verification:** CONFIRMED — `runtime/pkg/fileutil/fileutil.go:328` is `if !allowHostAccess && !strings.HasPrefix(finalPath, root)`. When `path` is absolute, `finalPath = path` (`:323-326`), so any absolute path sharing `root` as a raw string prefix (without a separator boundary) passes. `root` originates from `filepath.Abs` so it has no trailing separator, making `/var/data/proj-other/...` pass against `root=/var/data/proj`.

**Manual Reproduction Steps (for QA):**
This is primarily a self-hosted / multi-tenant concern and hard to hit through the normal single-project UI. Confirm in code: `runtime/pkg/fileutil/fileutil.go:328` — `strings.HasPrefix(finalPath, root)` with no `os.PathSeparator` boundary; reachable from the `file` connector `FilePaths` → `ResolveLocalPath` (`file/file_store.go:18,36`). To demonstrate in a test: call `ResolveLocalPath("/var/data/proj-other/secrets.csv", "/var/data/proj", false)` and observe it returns the path instead of the "outside repo root" error.

### drivers-object-store-07 [MEDIUM] Redshift result pagination fetches with `context.Background()`
- **Category:** concurrency  **Confidence:** high  **Location:** `runtime/drivers/redshift/olap.go:242`
- **Problem:** When paging past the first result page, `rows.Next()` calls `r.client.GetStatementResult(context.Background(), ...)`. The query's context (available at `newRows`, line 147, but not retained) is dropped, so subsequent pages are neither cancellable nor timeout-bound and lose trace propagation. The initial page (line 155) correctly uses the caller's context, making the pagination inconsistent.
- **Failure scenario:** A large Redshift result is being streamed when the request is cancelled or times out. The first page respects cancellation, but every subsequent `GetStatementResult` runs to completion against Redshift regardless, wasting work and holding the row iterator open past cancellation.
- **Recommendation:** Store the context in `rows` (from `newRows`) and use it for follow-up `GetStatementResult` calls.

**Verification:** CONFIRMED — `runtime/drivers/redshift/olap.go:242` calls `r.client.GetStatementResult(context.Background(), ...)` for the next page, while the initial fetch in `newRows` (`:155`) uses the caller's `ctx`. The `rows` struct (`:133-143`) does not retain the context, so paged fetches are uncancellable and lose trace propagation.

**Manual Reproduction Steps (for QA):**
Not user-reproducible via UI in a simple observable way (the effect is that a cancelled/timed-out large query keeps paging server-side). Confirm in code: `runtime/drivers/redshift/olap.go:242` passes `context.Background()` to `GetStatementResult`; the `rows` struct at `:133` has no stored `ctx` to thread through. Contrast the correct first-page call at `:155`.

### drivers-object-store-08 [MEDIUM] Athena/Redshift `autoDeleteFileIterator.Close` skips S3 cleanup when the inner `Close` errors
- **Category:** bug  **Confidence:** high  **Location:** `runtime/drivers/redshift/warehouse.go:238` (and identical `runtime/drivers/athena/warehouse.go:253`)
- **Problem:** `Close()` returns early if `i.FileIterator.Close()` returns an error, so `i.cleanupFn()` (which deletes the `rill-tmp-*` UNLOAD output from S3) is never called on that path. The unloaded parquet files are then orphaned in the customer's bucket.
- **Failure scenario:** A transient error while closing the local download iterator (e.g. temp-dir removal fails) leaves the entire UNLOAD result set permanently in S3, silently accruing storage cost on every failed close.
- **Recommendation:** Always run `cleanupFn` (e.g. `defer`/`errors.Join`) so both the inner close error and the cleanup error are surfaced without skipping the S3 delete.
- **Verified:** Confirmed in both `redshift/warehouse.go:238-245` and `athena/warehouse.go:253-260`. The trigger (inner `blobIterator.Close` returning an error from `os.RemoveAll`/`bucket.Close`) is uncommon, so this leaks orphaned UNLOAD parquet into the customer bucket only on that error path; impact is storage cost/clutter, not data loss.

**Verification:** CONFIRMED — `redshift/warehouse.go:238-245`: `Close()` does `err := i.FileIterator.Close(); if err != nil { return err }` and only then `return i.cleanupFn()`, so an inner-close error skips `cleanupFn` (the S3 UNLOAD-output delete). Identical in `athena/warehouse.go:253-260`.

**Manual Reproduction Steps (for QA):**
Not reproducible through normal UI flow (requires the local temp-dir close to fail). Confirm in code: `runtime/drivers/redshift/warehouse.go:238-245` and `runtime/drivers/athena/warehouse.go:253-260` — the early `return err` on inner `Close()` failure bypasses `cleanupFn()`. To demonstrate: make `FileIterator.Close()` return an error (e.g. stub) and assert `cleanupFn` is never invoked, leaving `rill-tmp-*` parquet in the bucket.

### drivers-object-store-09 [MEDIUM] HTTPS connector discards its decoded config; connector-level `headers`/`path_prefixes` are ignored by `FilePaths`
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/drivers/https/https.go:105`
- **Problem:** `Open` decodes `config` into a `*ConfigProperties` (`cfg`) but then stores the raw `config` map (`config: config`) and never uses `cfg`. `FilePaths` (line 238) builds its request from `ModelInputProperties` (per-model props) only, so the connector-level `headers` and `path_prefixes` declared on `ConfigProperties` are never applied on the file-store download path. (They are honored on the DuckDB-secret ingestion path via `model_executor_self.go:735-742`, making the behavior inconsistent between the two ingestion routes.)
- **Failure scenario:** A user sets an `Authorization` header (or `path_prefixes`) on the `https` connector and defines a source that ingests via `FilePaths`; the header is silently dropped and the fetch fails with 401, or the prefix restriction is not enforced.
- **Recommendation:** Merge connector-level `headers` into the request in `FilePaths` (and decide whether `path_prefixes` should gate it), and drop the dead `cfg` decode or actually store/use it.

**Verification:** CONFIRMED — `runtime/drivers/https/https.go:105-115` decodes `config` into `cfg` (`*ConfigProperties`) but stores the raw map (`config: config`) and never references `cfg`. `FilePaths` (`:238-283`) builds the request solely from `ModelInputProperties` (`modelProp.Headers` at `:264`), so connector-level `headers`/`path_prefixes` on `ConfigProperties` are never applied on this path.

**Manual Reproduction Steps (for QA):**
1. Create an `https` connector with connector-level headers, e.g. `headers: { Authorization: "Bearer testtoken" }`.
2. Define a model/source that ingests via the file-store path from a URL that requires that header (or point at a request-inspection endpoint like `https://httpbin.org/headers`).
3. Trigger ingestion and inspect the outgoing request headers (or the fetched result). Observe the `Authorization` header from the connector config is absent — only per-model `headers` are sent. A per-model `headers` block on the source, by contrast, is applied.

### drivers-object-store-10 [MEDIUM] Redshift timestamp parsing rejects sub-second precision
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/drivers/redshift/olap.go:390`
- **Problem:** `convertStringValue` parses `timestamp` with `time.Parse(time.DateTime, val)` (`2006-01-02 15:04:05`) and `timestamptz` with `2006-01-02 15:04:05-07`. Redshift returns `TIMESTAMP`/`TIMESTAMPTZ` values with fractional seconds when present (microsecond precision). Go's `time.Parse` fails with "extra text" when the input has a `.` fraction the layout omits, so the whole row scan errors out.
- **Failure scenario:** Reading a table whose timestamp column holds `2023-06-01 12:30:45.123456` (via the OLAP path, e.g. `Head`/schema preview) fails with a parse error, blocking the query rather than returning the value.
- **Recommendation:** Use layouts that accept optional fractional seconds (append `.999999999`) or parse with a set of candidate layouts; add a test with fractional-second timestamps.
- **Verified:** The parse layouts at `olap.go:390-401` are exact and reachable (Redshift Data API returns temporal columns as `FieldMemberStringValue` → `convertStringValue`). Confidence stays medium because the exact string format the Data API emits (whether it always appends `.0`, only includes a fraction when non-zero, or never does) is not pinned down here; if it ever emits a fractional component the whole row scan errors. If the API always appends `.0`, severity would rise to high (every timestamp fails).

**Verification:** PLAUSIBLE — confirmed the exact layouts at `runtime/drivers/redshift/olap.go:390-401`: `timestamp` uses `time.DateTime` (`2006-01-02 15:04:05`, no fraction) and `timestamptz` uses `2006-01-02 15:04:05-07`. Go's `time.Parse` fails with "extra text" if the input carries a `.NNN` fraction the layout omits. Whether the Data API actually emits fractional seconds is not pinned down here, hence PLAUSIBLE rather than CONFIRMED.

**Manual Reproduction Steps (for QA):**
1. Configure a Redshift connector and point it at a table with a `TIMESTAMP` column holding a sub-second value, e.g. insert `2023-06-01 12:30:45.123456`.
2. Use the Redshift OLAP path in Rill (e.g. schema/data preview or a query that returns that column) rather than the UNLOAD ingest path.
3. Observe whether the row scan fails with a time-parse "extra text" error. If the value round-trips fine, the Data API is stripping/normalizing the fraction (finding does not fire in that config). Confirm layouts in code at `olap.go:390-401`.

### drivers-object-store-11 [MEDIUM] `blob.Download` materializes the entire glob listing in memory before downloading
- **Category:** performance  **Confidence:** medium  **Location:** `runtime/pkg/blob/download.go:76`
- **Problem:** `Download` calls `pagination.CollectAll(...)` to gather *all* matching `ObjectStoreEntry` values into a single slice before the first byte is downloaded. For globs matching very large numbers of objects this is unbounded memory and adds full-listing latency before any batch is produced. The batching pipeline downstream is streaming, but this upfront collection is not.
- **Failure scenario:** A source glob like `s3://bucket/events/**` over a bucket with millions of keys accumulates millions of entries in RAM up front, spiking memory and delaying ingestion start.
- **Recommendation:** Feed the download goroutine from the paginated listing incrementally (produce entries into a channel) instead of collecting them all first, or cap/opt-in the collection.
- **Verified:** Confirmed at `download.go:76-83`. Note the accumulation is of listing *metadata* (`ObjectStoreEntry`: a path string plus a few scalar fields), so memory scales with object count, not data volume — on the order of hundreds of MB only at the millions-of-objects scale. Real for `**` over very large buckets; not a concern for typical globs.

**Verification:** CONFIRMED — `runtime/pkg/blob/download.go:76-83` calls `pagination.CollectAll(...)` over `ListObjectsForGlob` into a single `entries` slice before any download starts (the iterator is constructed from `objects: entries` at `:89`). Accumulation is bounded by object count (metadata only), so it only matters at very large scale.

**Manual Reproduction Steps (for QA):**
1. Point a source glob at a bucket with a very large number of matching keys, e.g. `s3://bucket/events/**` over hundreds of thousands to millions of objects.
2. Trigger ingestion and watch runtime memory and time-to-first-download. Memory rises during the initial full listing and no bytes download until listing completes.
3. Compare with a narrow glob (few objects) where download starts promptly. Confirm in code at `download.go:76`.

### drivers-object-store-12 [LOW] Salesforce `getBatches` indexes `allBatches[0]` without a length check
- **Category:** bug  **Confidence:** medium  **Location:** `runtime/drivers/salesforce/bulk.go:162`
- **Problem:** The code guards `if allBatches != nil` and then dereferences `allBatches[0]`, but a non-nil empty slice passes the guard and panics. The `err` from `GetBatchesWithContext` (line 160) is also only checked afterward (line 178), so an error that returns an empty-but-non-nil slice would panic before the error is handled.
- **Failure scenario:** A PK-chunking job whose batch list comes back empty (or an SDK response yielding a non-nil empty slice) triggers an index-out-of-range panic in the ingestion goroutine.
- **Recommendation:** Check `len(allBatches) > 0` (and handle `err` first) before indexing.

**Verification:** CONFIRMED — `runtime/drivers/salesforce/bulk.go:162-164` guards `if allBatches != nil` then dereferences `allBatches[0]`; a non-nil empty slice passes the guard and panics with index-out-of-range. The `err` from `GetBatchesWithContext` (`:160`) is only checked later at `:178`, so an error returning an empty-but-non-nil slice panics first.

**Manual Reproduction Steps (for QA):**
Not reliably reproducible via UI (depends on the Salesforce SDK returning a non-nil empty batch slice for a PK-chunking job). Confirm in code: `runtime/drivers/salesforce/bulk.go:162` — `allBatches != nil` guard followed by `allBatches[0]` at `:163` with no `len()` check, and `err` handled only at `:178`. To demonstrate: unit-test `getBatches` with a session stub returning `([]force.BatchInfo{}, nil)` for a PK-chunking-enabled job and observe the panic.

### drivers-object-store-13 [LOW] Athena resolves the workgroup output location twice per query
- **Category:** performance  **Confidence:** high  **Location:** `runtime/drivers/athena/warehouse.go:52` and `:125`
- **Problem:** `QueryAsFiles` calls `sourceProperties.ResolveOutputLocation(ctx, client, c.config)` at line 52, and `unload` calls it again at line 125. When neither the source nor connector config sets `output_location`, each call issues a `GetWorkGroup` API request, so the same lookup runs twice per ingest.
- **Failure scenario:** Not a correctness bug, but every Athena ingest that relies on the workgroup's default output location makes a redundant `GetWorkGroup` call, adding latency and API usage.
- **Recommendation:** Resolve the output location once and pass it into `unload`.

**Verification:** CONFIRMED — `QueryAsFiles` calls `sourceProperties.ResolveOutputLocation(ctx, client, c.config)` at `runtime/drivers/athena/warehouse.go:52`, and `unload` calls it again at `:125`. When neither source nor config sets `output_location`, `ResolveOutputLocation` issues a `GetWorkGroup` request, so it runs twice per ingest.

**Manual Reproduction Steps (for QA):**
1. Configure an Athena connector with a `workgroup` but no explicit `output_location` (rely on the workgroup's default result location).
2. Ingest a model from Athena and capture the AWS API calls (CloudTrail, or an SDK request logger / proxy).
3. Observe two `GetWorkGroup` calls per single ingest. Setting `output_location` explicitly eliminates them, confirming the redundancy. Confirm in code at `athena/warehouse.go:52` and `:125`.

### drivers-object-store-14 [LOW] Azure `GetConnectionString` mutates shared config from a getter
- **Category:** concurrency  **Confidence:** medium  **Location:** `runtime/drivers/azure/azure.go:108`
- **Problem:** `GetConnectionString()` (pointer receiver) writes `c.ConnectionString = os.Getenv("AZURE_STORAGE_CONNECTION_STRING")` into the shared `*ConfigProperties` as a side effect of a getter. It is called from `newStorageClient` on the live `c.config`, so concurrent connection opens race on this write, and the getter's contract (read-only) is violated.
- **Failure scenario:** Two goroutines opening buckets on the same connector concurrently write `ConnectionString` without synchronization — a data race flagged by `-race`, and a surprising mutation of connector state.
- **Recommendation:** Compute the connection string into a local variable without mutating the receiver.

**Verification:** CONFIRMED — `runtime/drivers/azure/azure.go:108-113` `GetConnectionString()` (pointer receiver) writes `c.ConnectionString = os.Getenv(...)` when it is empty and `AllowHostAccess` is set. It is called from `newStorageClient` on the shared `c.config` (`azure/object_store.go:163`), invoked concurrently from `object_store.go:36`/`:154` and `azure.go:179`, so concurrent bucket opens race on the write.

**Manual Reproduction Steps (for QA):**
Not observable through the UI. Confirm in code: `runtime/drivers/azure/azure.go:108-113` mutates the receiver's `ConnectionString` from a getter, called via `newStorageClient` (`azure/object_store.go:163`) on the shared `c.config`. To demonstrate: with `AZURE_STORAGE_CONNECTION_STRING` set and `allow_host_access=true`, run the Azure driver under `go test -race` opening two buckets concurrently on the same connector and observe the data-race report on `ConnectionString`.

## Product Improvements

### drivers-object-store-15 [MEDIUM] HTTPS/`FilePaths` fetch has no SSRF guardrails
- **Category:** product-improvement  **Confidence:** medium  **Location:** `runtime/drivers/https/https.go:259`
- **Problem:** `FilePaths` issues `http.DefaultClient.Do` against any user-supplied URL with `HasAnonymousSourceAccess=true` and no allow/deny list. In Rill Cloud a model author can point a source at internal addresses (e.g. `http://169.254.169.254/…` metadata, RFC1918 hosts) and have the runtime fetch them.
- **Failure scenario:** A developer defines an `https` source `http://169.254.169.254/latest/meta-data/iam/security-credentials/…`; the runtime fetches cloud metadata and stores it as an ingestible file.
- **Recommendation:** Resolve and reject link-local/loopback/private ranges (and non-http(s) schemes) before fetching, ideally behind a config flag for self-hosted setups that legitimately need internal hosts.

**Verification:** CONFIRMED — `runtime/drivers/https/https.go:259-275` builds `http.NewRequestWithContext` from the user-supplied `path` and calls `http.DefaultClient.Do(req)` with no scheme/host allow-deny list; `HasAnonymousSourceAccess` returns `true` unconditionally (`:122-124`). A model author can target internal/link-local addresses.

**Manual Reproduction Steps (for QA):**
1. In a Rill project, define an `https` source pointing at an internal address, e.g. `path: http://169.254.169.254/latest/meta-data/` (or any RFC1918 host reachable from the runtime), on a deployment where such addresses resolve.
2. Trigger ingestion.
3. Observe the runtime fetches the internal resource and stores the response as an ingestible file, with no allow/deny-list rejection. Confirm in code at `https/https.go:259-275`.

### drivers-object-store-16 [MEDIUM] Object-store model `Delete` can over-delete via a non-boundary prefix
- **Category:** data-loss  **Confidence:** medium  **Location:** `runtime/drivers/s3/model_manager.go:42`
- **Problem:** `Delete` computes `base, _ := doublestar.SplitPattern(...)` and deletes everything under that prefix with no trailing-delimiter boundary. `MergePartitionResults` (line 61-68) explicitly warns about this class of data loss and guards against the bucket root, but `Delete` has no comparable guard, and the prefix match is non-boundary (deleting `output` also matches `output-archive/`).
- **Failure scenario:** A model output configured at `s3://bucket/output/*.parquet` deletes `output/` on refresh but also any sibling `output.../` keys sharing the literal prefix; an output glob directly at a shared prefix can delete unrelated data.
- **Recommendation:** Require the delete prefix to end in a delimiter (and be non-empty / non-root), mirroring the `MergePartitionResults` guard.

**Verification:** CONFIRMED — `runtime/drivers/s3/model_manager.go:42-43` computes `base, _ := doublestar.SplitPattern(strings.TrimPrefix(u.Path, "/"))` and passes it straight to `deleteObjectsInPrefix`, which deletes everything under that literal prefix (`ListObjectsV2` with `Prefix: &prefix`, no trailing-delimiter boundary). `MergePartitionResults` (`:61-68`) explicitly guards against the bucket-root data-loss case, but `Delete` has no comparable guard, and the S3 `Prefix` match is non-boundary, so prefix `output` also matches `output-archive/`.

**Manual Reproduction Steps (for QA):**
1. Configure an S3 model output at a path like `s3://bucket/output/*.parquet`.
2. In the same bucket, create a sibling prefix that shares the literal string, e.g. `s3://bucket/output-archive/keep.parquet`.
3. Trigger a model refresh/delete (which invokes `Delete` → `deleteObjectsInPrefix` with base prefix `output`).
4. Observe objects under `output-archive/` are also deleted because the S3 `Prefix` filter `output` matches them. Confirm in code at `s3/model_manager.go:42`.

## New Feature Ideas

- **Unified object-store access-policy layer (S/M):** Extract a single `authorizePath(op, bucket, path)` used by `ListObjects`, `ListObjectsForGlob`, `DownloadFiles`, and model `Delete`, with delimiter-boundary matching. Closes findings 02/05/16 in one place and makes `path_prefixes` a trustworthy boundary. User value: safe multi-tenant/self-serve bucket browsing for the AI assistant.
- **Server-side warehouse UNLOAD for Snowflake (M):** Snowflake currently streams Arrow batches through the runtime and rewrites Parquet locally (`snowflake/warehouse.go`), which is memory-heavy. Add an optional `COPY INTO <stage/s3>` + download path like Athena/Redshift. User value: large Snowflake extracts without runtime memory pressure.
- **Credential preflight / "Test connection" per bucket (S):** Extend `Ping` to validate credentials against a specific bucket/prefix and report actionable errors (region mismatch, missing `s3:ListBucket`, anonymous fallback engaged). User value: users debug connector setup without cryptic SDK errors.
- **Glob listing progress & object-count caps (M):** Surface listing/download progress and a configurable max-objects/max-bytes cap for globs (ties to finding 11). User value: predictable behavior and no OOM on `**` over huge buckets.
- **Incremental/resumable downloads (L):** Persist the download manifest so a failed ingest resumes only the missing files instead of re-downloading the whole glob. User value: faster, cheaper retries on large sources.
- **Structured warehouse type mapping with round-trip tests (M):** The Redshift/BigQuery/Snowflake type conversions are ad hoc (findings 10). A shared, table-tested type-mapping module (including fractional timestamps, decimals, arrays/structs) would reduce silent data-fidelity bugs. User value: correct values across warehouses.

## Notes & Missing Test Coverage

- **`pkg/blob/bucket.go:ListObjectsForGlob` pagination** (the `start`/`end`/`driverStartAfter` interplay across GCS `StartOffset` (inclusive), S3 `StartAfter` (exclusive), and Azure `StartFrom`, plus the directory-accumulation `currentDir` logic and the mid-page `driverPageToken = blob.FirstPageToken` reset at line 239) is intricate and I could not fully verify all provider paginate-and-filter combinations by reading alone. It deserves provider-specific table tests, especially the `obj.Key < driverStartAfter` hard-error branch (line 135-137) under Azure.
- **`pkg/blob/download.go` shutdown choreography** (`downloadFiles` → `downloadsCh` → `batchDownloads` → `batchCh`, with `Close` draining `batchCh`) appears deadlock-free because a goroutine blocked on the unguarded `downloadsCh <- …` (line 262) keeps `g.Wait()` from returning and thus keeps `batchDownloads` reading; still, the unguarded send is fragile and would benefit from a `select { case downloadsCh<-…; case <-ctx.Done(): }` plus an explicit test that `Close()` mid-download returns promptly.
- **Redshift `UNLOAD` (finding 01)** and **`convertStringValue` timestamps (finding 10)** have no tests exercising quotes/fractional seconds; both are high-likelihood real-world inputs.
- **`drivers.ListObjects` prefix logic** (`object_store.go`) has no test covering the sibling-prefix bypass (`data` vs `database`) or trailing-slash normalization.
- **BigQuery client lifecycle (finding 03):** no test asserts the client is closed after a successful `QueryAsFiles`; a leak test (or a `Close` returning the client-close error) would catch regressions.
- **Snowflake `fileIterator.Next`:** `f.batches[0].Fetch()` at line 205 is called for the empty/schema check and again inside the errgroup loop (line 255), and the first fetch's records are never `Release()`d. Worth confirming `ArrowBatch.Fetch` is cached/idempotent to avoid a double download and a minor memory retention. (Left as a note, not a finding: verifying the leak/double-download requires pinning gosnowflake's `ArrowBatch.Fetch` caching semantics, which could not be confirmed from source here.)

## Verification Notes

Every one of the 16 findings was independently re-checked against the cited code; all 16 were confirmed as real defects (nothing was a false positive, already-handled, or unreachable). `removed_count = 0`. The prior review was accurate on substance; the changes below are calibration, not retractions.

Changes made:

- **Finding 02 severity HIGH → MEDIUM.** The bypass is real and complete across s3/gcs/azure (verified: `ListObjectsForGlob` and `DownloadFiles` never call `drivers.ListObjects`, while `ListObjects` does), but all reachable entry points are gated to developer-level access and `path_prefixes` is framed partly as a credential-routing convenience. It is a genuine boundary only under admin-managed connectors; not an unprivileged-actor or cross-tenant escalation. Location corrected to the exact bypassing call sites (`:81`/`:101` in s3, and the gcs/azure equivalents).
- **Finding 06** confirmed reachable via the `file` connector and confirmed non-boundary (`root` comes from `filepath.Abs`, which strips trailing separators). Added an explicit note that practical exploitation is narrow (requires `allow_host_access=false` plus a sibling directory sharing the root name as a string prefix) — kept at MEDIUM as a defense-in-depth fix.
- **Finding 08** confirmed in both redshift and athena; clarified the trigger is an uncommon inner-`Close` error and the impact is orphaned S3 temp files (cost/clutter), not data loss.
- **Finding 10** confirmed the exact parse layouts and reachability; kept confidence medium and flagged that severity hinges on the Data API's actual timestamp string format (could be high if it always emits a fraction).
- **Finding 11** confirmed; clarified the up-front collection is listing *metadata* (bounded by object count, not data size), so the memory concern only bites at the millions-of-objects scale.

Cross-checks that strengthened findings by contrast: Snowflake (`snowflake/warehouse.go:152-159,182-188`) and Databricks (`databricks/warehouse.go:129-139,162-170`) both close `db`/`conn`/`rows` in *both* `Next` and `Close`, confirming BigQuery's omission (finding 03) is a genuine leak, not an accepted pattern. Redshift's `executeQuery` polling loop (`redshift.go:311-340`) correctly uses the caller's context for `DescribeStatement` and only falls back to `context.Background()` for the cancellation cleanup `CancelStatement` — this is the *correct* pattern and, by contrast, confirms the misuse in finding 07 (`olap.go:242`) where a normal pagination fetch drops the caller's context. The `blob/download.go` shutdown choreography was re-traced and is race-free and deadlock-free (channel-close happens-before edges cover every `downloadErr` read), so no finding was added there.

No new findings were added: the remaining candidates examined (world-writable file modes in `download.go` — mooted by the `0700` `MkdirTemp` parent; missing `OutputLocation` validation in redshift `QueryAsFiles`; the un-paginated expansion in `listObjectsFromPathPrefixes`) were each either not actually exploitable or too marginal to include without overstating.
