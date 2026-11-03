# Verified audit: `runtime-server.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **11** numbered findings. Every source ID appears in one disposition table and one manifest line.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| runtime-server-03 | MEDIUM | P1 | JWKS bootstrap creates 20 refreshers on success | `runtime/server/auth/jwt.go:202-224` lacks a success `break`; only the final refresher is stopped at `:236-239` | Break on success and test first-attempt success plus retry. |
| runtime-server-04 | MEDIUM | P1 | SSE producers can strand on disconnected clients | Bare sends at `runtime/server/sse.go:137,171,218,265` and `runtime/server/chat.go:525,545` target unbuffered channels | Make sends select on context cancellation or use a bounded funnel that is drained on exit. |
| runtime-server-05 | MEDIUM | P1 | SSE validates event types after starting streams | `runtime/server/sse.go:84,111-229` sorts and launches per type before `default` rejects an invalid type | Validate the complete event-type set before launching any goroutine. |
| runtime-server-08 | LOW | P2 | Public health endpoint returns internal errors | `runtime/server/health.go:88-97`; route is intentionally public | Return a generic public failure and log detailed component errors. |
| runtime-server-09 | LOW | P2 | Global mutex serializes SQL JSON rewriting | `runtime/server/queries.go:19-26,110-160` | Confirm whether the parser is unsafe; remove or narrowly scope the mutex and benchmark. |
| runtime-server-10 | LOW | P1 | Export appends HTTP errors after streaming begins | `runtime/server/downloads.go:328-356` | Track response commitment; after streaming begins, abort/log instead of appending an error body. |
| runtime-server-11 | MEDIUM | P1 | Raw-query limit rewrite is DuckDB-only | `runtime/server/queries.go:39-47,110-160` invokes `json_serialize_sql`/`json_deserialize_sql` on the selected connector | Gate by dialect and use a portable subquery/post-read cap elsewhere. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| runtime-server-01 | FIXED | All seven connector handlers now require `ReadInstance` at `runtime/server/connector_service.go:15,42,85,121,146,179,212`, with authorization tests. | PR #9700, `d4553f382` |
| runtime-server-02 | REMOVE (fixed by deletion) | The unused `QueryBatch` RPC, generated API surface, implementation, and tests no longer exist. | PR #9714, `887f35301` |

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| runtime-server-06 | LOW | P2 | `Claims` still aliases/mutates `Attrs`, but normal interception pre-populates `id`, reducing live race reachability. | `runtime/server/auth/claims.go:29-36,66-70`; clone the map and run concurrent claims tests. |
| runtime-server-07 | LOW | P3 | Decompression is unbounded, but the authenticated server-generated token is capped at 7.5 KB. | `runtime/server/downloads.go:30-33,413-414,468-475`; add a decompressed-size ceiling and bomb test. |

Counts: **7 pending, 2 fixed/remove, 2 partial/needs-test, 0 invalid = 11 total**.

## Machine-readable manifest

```text
FIXED|runtime-server.md|runtime-server-01|HIGH|-|-|ConnectorService handlers lacked authorization|runtime/server/connector_service.go:15,42,85,121,146,179,212; PR#9700
REMOVE|runtime-server.md|runtime-server-02|HIGH|-|-|QueryBatch sent concurrently on one stream|RPC and runtime/server/batch_query.go removed by PR#9714
PENDING|runtime-server.md|runtime-server-03|MEDIUM|P1|-|JWKS bootstrap leaks refreshers|runtime/server/auth/jwt.go:202-224
PENDING|runtime-server.md|runtime-server-04|MEDIUM|P1|-|SSE producer sends ignore cancellation|runtime/server/sse.go:137,171,218,265; runtime/server/chat.go:525,545
PENDING|runtime-server.md|runtime-server-05|MEDIUM|P1|-|SSE validates types after goroutine launch|runtime/server/sse.go:84,111-229
PARTIAL|runtime-server.md|runtime-server-06|LOW|P2|-|JWT claims mutate and alias attributes|runtime/server/auth/claims.go:29-36,66-70
PARTIAL|runtime-server.md|runtime-server-07|LOW|P3|-|Download-token decompression has no output bound|runtime/server/downloads.go:30-33,468-475
PENDING|runtime-server.md|runtime-server-08|LOW|P2|-|Public health leaks component errors|runtime/server/health.go:88-97
PENDING|runtime-server.md|runtime-server-09|LOW|P2|-|Global mutex serializes SQL JSON rewrite|runtime/server/queries.go:19-26
PENDING|runtime-server.md|runtime-server-10|LOW|P1|-|Export appends error after streaming|runtime/server/downloads.go:328-356
PENDING|runtime-server.md|runtime-server-11|MEDIUM|P1|-|Raw query limit rewrite is DuckDB-only|runtime/server/queries.go:39-47,110-160
```
