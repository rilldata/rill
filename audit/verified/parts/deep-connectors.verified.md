# Verified: deep-connectors.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **20** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| deep-connectors-01 | HIGH | P1 | rduckdb remote delete uses wrong metadata key and can panic | `runtime/pkg/rduckdb/remote.go:276,285-298`; root `meta.json` is deleted and non-EOF errors fall through to `obj.Key` | Delete `<table>/meta.json`; return on iterator error before dereference; bound retries/errors. |
| deep-connectors-02 | HIGH | P0 | Failed metadata upload deletes last good remote version | `runtime/pkg/rduckdb/remote.go:244-255`; old version deletion still runs when metadata upload returns an error | Delete old version only after metadata commit succeeds; validate referenced files during pull. |
| deep-connectors-03 | HIGH | P0 | ClickHouse ingestion exports storage credentials in SQL telemetry | `runtime/drivers/clickhouse/model_executor_objectstore_self.go:126-163`; credentials are SQL literals; `clickhouse.go:933+` tracing and `dialect.go:237-240` sanitizer do not mask them | Suppress/sanitize sensitive statements and move credentials out of SQL; escape all literals. |
| deep-connectors-04 | HIGH | P0 | Salesforce fetch errors become successful empty files | `runtime/drivers/salesforce/bulk.go:201-231`; `fetchBatchResult` still returns only `io.Reader` and converts failures to empty readers | Return and propagate errors, fail failed batches, and retry transient retrieval failures. |
| deep-connectors-05 | HIGH | P1 | rduckdb leaks an OTEL callback per write handle | `runtime/pkg/rduckdb/db.go:956` and `generic.go:189`; unregistrable DB metrics are registered on ephemeral handles | Do not register on write handles, or retain/unregister registrations in release. |
| deep-connectors-06 | MEDIUM | P2 | Object-store pagination rebuilds clients/credentials | `runtime/drivers/s3/object_store.go:61-82`, GCS/Azure equivalents; every page opens/closes a bucket and re-resolves auth/region | Cache bucket/client or at least resolved credentials/config for the connection lifetime. |
| deep-connectors-07 | MEDIUM | P1 | Postgres OLAP defaults to one connection | `runtime/drivers/postgres/postgres.go:211-216,425-431`; default remains 1 and is applied to dashboard pool | Raise default to a bounded 5–10 and expose/document tuning. |
| deep-connectors-08 | MEDIUM | P1 | MySQL/Snowflake OLAP pools are unbounded | `runtime/drivers/mysql/mysql.go:384-390`; `runtime/drivers/snowflake/snowflake.go:381-386`; neither sets maximum open connections | Add bounded configurable max/idle pool sizes. |
| deep-connectors-09 | MEDIUM | P2 | Druid cancel hook fires after successful queries | `runtime/drivers/druid/druidsqldriver/druid_api_sql_driver.go:116-129`; returned `AfterFunc` stop function remains discarded | Retain and stop the callback when rows finish/close and on all post-registration errors. |
| deep-connectors-10 | MEDIUM | P1 | Druid clean EOF can silently truncate arrayLines | `runtime/drivers/druid/druidsqldriver/druid_api_sql_driver.go:319-335,483`; decoder has no terminal-success marker check | Require/track the trailing terminator and error when it is absent. |
| deep-connectors-11 | MEDIUM | P1 | BigQuery logs partial writes but returns success | `runtime/drivers/bigquery/warehouse.go:308-310,354-360`; row-count shortfall is only logged | Return an error for verified count shortfall and test both formats. |
| deep-connectors-12 | MEDIUM | P1 | rduckdb generic pool connection leaks on USE failure | `runtime/pkg/rduckdb/generic.go:538-556`; acquired connection returns on `USE` error without close | Close on every post-acquire error path. |
| deep-connectors-13 | MEDIUM | P1 | rduckdb dirty flags can permanently serve stale data | `runtime/pkg/rduckdb/db.go:470-477,562-569,715-724,741-751`; failed pulls are marked clean and metadata-write failures can return success | Clear dirty only after successful pull; propagate metadata errors and keep state dirty until catalog convergence. |
| deep-connectors-14 | MEDIUM | P2 | Incremental DuckDB refresh copies/uploads whole table | `runtime/pkg/rduckdb/db.go:512-518`; `io.go:44-65`; `remote.go:204-234` | Use reflinks/copy-on-write locally and an incremental remote representation; document cost meanwhile. |
| deep-connectors-15 | MEDIUM | P1 | ClickHouse file ingest double-buffers whole files | `runtime/drivers/clickhouse/model_executor_filestore_self.go:151-162` | Stream/chunk into ClickHouse instead of constructing a full SQL string. |
| deep-connectors-16 | MEDIUM | P2 | HTTPS refresh leaks a full temp file | `runtime/drivers/https/https.go:277`; consumer `runtime/drivers/clickhouse/model_executor_filestore_self.go:104+` lacks ownership cleanup | Return a cleanup callback/ownership bit and keep temp data in instance-scoped storage. |
| deep-connectors-17 | LOW | P2 | Generic RenameTable does not quote identifiers | `runtime/pkg/rduckdb/generic.go:416` | Apply `safeSQLName` to old and new identifiers. |
| deep-connectors-18 | LOW | P2 | Retry helpers ignore most transient failures | `runtime/pkg/blob/download.go:323-340`; `runtime/pkg/rduckdb/remote.go:302-320` | Classify 429/5xx/network resets/timeouts and retry with jittered exponential backoff. |
| deep-connectors-19 | LOW | P2 | rduckdb read acquisition serializes schema work | `runtime/pkg/rduckdb/db.go:329-349,1235-1274`; every read does serialized `USE`, first read replays attachments | Prewarm snapshots before publication, avoid redundant `USE`, and align idle pool size. |
| deep-connectors-20 | MEDIUM | P1 | Remote discovery returns handled `errNotFound` | `runtime/pkg/rduckdb/remote.go:165-172`; successful catalog add falls through to `return err` | Continue after adding a verified new table; return only unhandled errors. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| — | None | Recent merged connector PRs on this snapshot do not replace these paths; every cited defect remains present. | — |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| — | — | — | None | Trigger-focused fixtures are still needed for regression coverage, but all twenty underlying code defects are directly present. |

Counts: **20 pending, 0 fixed/remove, 0 partial/needs-test, 0 invalid = 20 total**.

## Machine-readable ledger

```text
PENDING|deep-connectors.md|deep-connectors-01|HIGH|P1|-|rduckdb remote-delete panic and wrong metadata key|runtime/pkg/rduckdb/remote.go:276,285-298
PENDING|deep-connectors.md|deep-connectors-02|HIGH|P0|-|rduckdb deletes good version after failed metadata upload|runtime/pkg/rduckdb/remote.go:244-255
PENDING|deep-connectors.md|deep-connectors-03|HIGH|P0|-|ClickHouse emits storage credentials in SQL telemetry|runtime/drivers/clickhouse/model_executor_objectstore_self.go:126-163
PENDING|deep-connectors.md|deep-connectors-04|HIGH|P0|-|Salesforce fetch failure becomes empty success|runtime/drivers/salesforce/bulk.go:201-231
PENDING|deep-connectors.md|deep-connectors-05|HIGH|P1|-|rduckdb write-handle metric callback leak|runtime/pkg/rduckdb/db.go:956; runtime/pkg/rduckdb/generic.go:189
PENDING|deep-connectors.md|deep-connectors-06|MEDIUM|P2|-|Object-store page rebuilds client and credentials|runtime/drivers/s3/object_store.go:61-82
PENDING|deep-connectors.md|deep-connectors-07|MEDIUM|P1|-|Postgres OLAP single-connection default|runtime/drivers/postgres/postgres.go:211-216,425-431
PENDING|deep-connectors.md|deep-connectors-08|MEDIUM|P1|-|MySQL and Snowflake pools unbounded|runtime/drivers/mysql/mysql.go:384-390; runtime/drivers/snowflake/snowflake.go:381-386
PENDING|deep-connectors.md|deep-connectors-09|MEDIUM|P2|-|Druid cancel hook not stopped|runtime/drivers/druid/druidsqldriver/druid_api_sql_driver.go:116-129
PENDING|deep-connectors.md|deep-connectors-10|MEDIUM|P1|-|Druid arrayLines truncation undetected|runtime/drivers/druid/druidsqldriver/druid_api_sql_driver.go:319-335,483
PENDING|deep-connectors.md|deep-connectors-11|MEDIUM|P1|-|BigQuery partial write returns success|runtime/drivers/bigquery/warehouse.go:308-310,354-360
PENDING|deep-connectors.md|deep-connectors-12|MEDIUM|P1|-|rduckdb generic connection leak|runtime/pkg/rduckdb/generic.go:538-556
PENDING|deep-connectors.md|deep-connectors-13|MEDIUM|P1|-|rduckdb dirty protocol serves stale data|runtime/pkg/rduckdb/db.go:470-477,562-569,715-724,741-751
PENDING|deep-connectors.md|deep-connectors-14|MEDIUM|P2|-|Incremental DuckDB full copy and upload|runtime/pkg/rduckdb/db.go:512-518; runtime/pkg/rduckdb/remote.go:204-234
PENDING|deep-connectors.md|deep-connectors-15|MEDIUM|P1|-|ClickHouse file ingest double buffers|runtime/drivers/clickhouse/model_executor_filestore_self.go:151-162
PENDING|deep-connectors.md|deep-connectors-16|MEDIUM|P2|-|HTTPS temp file leak|runtime/drivers/https/https.go:277
PENDING|deep-connectors.md|deep-connectors-17|LOW|P2|-|rduckdb RenameTable unescaped identifiers|runtime/pkg/rduckdb/generic.go:416
PENDING|deep-connectors.md|deep-connectors-18|LOW|P2|-|Transient retries too narrow|runtime/pkg/blob/download.go:323-340; runtime/pkg/rduckdb/remote.go:302-320
PENDING|deep-connectors.md|deep-connectors-19|LOW|P2|-|rduckdb read path serialization|runtime/pkg/rduckdb/db.go:329-349,1235-1274
PENDING|deep-connectors.md|deep-connectors-20|MEDIUM|P1|-|Remote discovery returns handled not-found|runtime/pkg/rduckdb/remote.go:165-172
```
