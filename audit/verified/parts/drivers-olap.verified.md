# Verified: drivers-olap.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **17** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| drivers-olap-01 | HIGH | P1 | Pinot schema error leaves nil then panics in sort | `runtime/drivers/pinot/information_schema.go:266-288`; failed lookups leave nil slots dereferenced by `sort.Slice` | Append only successful lookups to a synchronized slice and return/log structured errors. |
| drivers-olap-02 | MEDIUM | P1 | DuckDB connection metric has data race | `runtime/drivers/duckdb/duckdb.go:214,320,730,744`; callback reads plain `int` outside the condition lock | Use `atomic.Int64` or read under the same lock. |
| drivers-olap-03 | MEDIUM | P1 | ClickHouse Open leaks resources after connect | `runtime/drivers/clickhouse/clickhouse.go:308-368`; several post-open errors precede construction of the owning `Connection` | Defer close/embedded-stop until a success flag transfers ownership. |
| drivers-olap-04 | MEDIUM | P2 | DuckDB failed reopen leaks metric registration | `runtime/drivers/duckdb/duckdb.go:213-229`; reopen failure omits unregister/cancel | Unregister the callback and cancel before returning. |
| drivers-olap-05 | MEDIUM | P1 | ClickHouse deadline wrapper leaks a goroutine per query | `runtime/drivers/clickhouse/olap.go:513-540`; cancellation goroutine can wait forever on a background-derived parent | Return and invoke cancel with result cleanup, or use a no-goroutine context strategy. |
| drivers-olap-06 | LOW | P3 | ClickHouse/Pinot mutate caller statement | `runtime/drivers/clickhouse/olap.go:168`; `runtime/drivers/pinot/olap.go:89` | Construct final SQL in local variables. |
| drivers-olap-07 | MEDIUM | P1 | StarRocks LARGEINT scans into int64 | `runtime/drivers/starrocks/olap.go:204-205,322-329`; INT128 is assigned `sql.NullInt64` | Scan 128-bit values as strings/big integers and preserve precision. |
| drivers-olap-08 | MEDIUM | P1 | Pinot conversion returns errors as values and can panic | `runtime/drivers/pinot/sqldriver/driver.go:229-235`; timestamp error becomes a cell and bool uses unchecked assertion | Return conversion errors through `Rows.Next`; use checked conversions. |
| drivers-olap-09 | LOW | P2 | ClickHouse openHandle leaks DBs on fallback/error | `runtime/drivers/clickhouse/clickhouse.go:940-978`; native DB is replaced without close and registration failures return without close | Close every allocated DB before reassignment/error. |
| drivers-olap-10 | LOW | P2 | Single-row ClickHouse queries bypass deadline stripping | `runtime/drivers/clickhouse/olap.go:504-531`; override is `QueryRowContext`, while callers use embedded `QueryRowxContext` | Override/route `QueryRowxContext`, centralizing context adaptation. |
| drivers-olap-11 | LOW | P2 | Pinot broker request ignores query context | `runtime/drivers/pinot/sqldriver/driver.go:63-77`; TODO remains and blocking client call has no context | Use a context-aware Pinot API or cancellation/timeout wrapper. |
| drivers-olap-12 | LOW | P2 | Config values enter SQL without escaping | `runtime/drivers/starrocks/starrocks.go:362-380`; `runtime/drivers/duckdb/duckdb.go:513` | Escape identifiers/strings with dialect-safe helpers or parameterize. |
| drivers-olap-13 | LOW | P3 | Druid readiness query receives unused argument | `runtime/drivers/druid/information_schema.go:92-93` | Remove the `name` argument. |
| drivers-olap-14 | LOW | P3 | Pinot library prints lookup errors to stdout | `runtime/drivers/pinot/information_schema.go:273-277` | Use structured logger plus request context. |
| drivers-olap-17 | LOW | P2 | ClickHouse registers unremovable metric callback per open | `runtime/drivers/clickhouse/clickhouse.go:978`; dependency API discards registration | Register once per long-lived connector or use an API returning a registration that `Close` unregisters. |
| drivers-olap-15 | LOW | P3 | Shared information-schema helper stops at 1000 tables | `runtime/drivers/information_schema.go:68-99` | Implement compound schema/table pagination or stream results. |
| drivers-olap-16 | LOW | P2 | Pinot silently omits schema lookup failures | `runtime/drivers/pinot/information_schema.go:273-281` | Return an actionable partial-result diagnostic or fail on non-transient errors. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| — | None | No merged ancestor of the snapshot replaces or repairs these code paths. | — |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| — | — | — | None | Each finding's cited defect is still directly visible in current code; runtime/provider fixtures are useful regression coverage but are not needed to establish the code defect. |

Counts: **17 pending, 0 fixed/remove, 0 partial/needs-test, 0 invalid = 17 total**.

## Machine-readable ledger

```text
PENDING|drivers-olap.md|drivers-olap-01|HIGH|P1|-|Pinot All nil-table panic|runtime/drivers/pinot/information_schema.go:266-288
PENDING|drivers-olap.md|drivers-olap-02|MEDIUM|P1|-|DuckDB connection metric race|runtime/drivers/duckdb/duckdb.go:214,320,730,744
PENDING|drivers-olap.md|drivers-olap-03|MEDIUM|P1|-|ClickHouse Open resource leaks|runtime/drivers/clickhouse/clickhouse.go:308-368
PENDING|drivers-olap.md|drivers-olap-04|MEDIUM|P2|-|DuckDB failed-open metric leak|runtime/drivers/duckdb/duckdb.go:213-229
PENDING|drivers-olap.md|drivers-olap-05|MEDIUM|P1|-|ClickHouse context goroutine leak|runtime/drivers/clickhouse/olap.go:513-540
PENDING|drivers-olap.md|drivers-olap-06|LOW|P3|-|Drivers mutate Statement Query|runtime/drivers/clickhouse/olap.go:168; runtime/drivers/pinot/olap.go:89
PENDING|drivers-olap.md|drivers-olap-07|MEDIUM|P1|-|StarRocks INT128 scan overflow|runtime/drivers/starrocks/olap.go:204-205,322-329
PENDING|drivers-olap.md|drivers-olap-08|MEDIUM|P1|-|Pinot cell conversion error and panic|runtime/drivers/pinot/sqldriver/driver.go:229-235
PENDING|drivers-olap.md|drivers-olap-09|LOW|P2|-|ClickHouse openHandle DB leaks|runtime/drivers/clickhouse/clickhouse.go:940-978
PENDING|drivers-olap.md|drivers-olap-10|LOW|P2|-|ClickHouse single-row context bypass|runtime/drivers/clickhouse/olap.go:504-531
PENDING|drivers-olap.md|drivers-olap-11|LOW|P2|-|Pinot request ignores context|runtime/drivers/pinot/sqldriver/driver.go:63-77
PENDING|drivers-olap.md|drivers-olap-12|LOW|P2|-|Unescaped connector config in SQL|runtime/drivers/starrocks/starrocks.go:362-380; runtime/drivers/duckdb/duckdb.go:513
PENDING|drivers-olap.md|drivers-olap-13|LOW|P3|-|Druid readiness query extra arg|runtime/drivers/druid/information_schema.go:92-93
PENDING|drivers-olap.md|drivers-olap-14|LOW|P3|-|Pinot prints errors to stdout|runtime/drivers/pinot/information_schema.go:273-277
PENDING|drivers-olap.md|drivers-olap-17|LOW|P2|-|ClickHouse metric callback leak|runtime/drivers/clickhouse/clickhouse.go:978
PENDING|drivers-olap.md|drivers-olap-15|LOW|P3|-|Information schema 1000-table cap|runtime/drivers/information_schema.go:68-99
PENDING|drivers-olap.md|drivers-olap-16|LOW|P2|-|Pinot silently drops tables|runtime/drivers/pinot/information_schema.go:273-281
```
