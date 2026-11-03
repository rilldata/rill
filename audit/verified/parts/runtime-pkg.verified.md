# Verified audit: `runtime-pkg.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **12** numbered findings, including the two numbered Product Improvements. Every source ID is classified once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| runtime-pkg-13 | HIGH | P0 | `GetLogs(limit=0)` panics | `runtime/pkg/logbuffer/buffer.go:73-95` indexes `logs[limit-1]` after allocating length zero | Return an empty slice immediately for zero and add API/replay regression tests. |
| runtime-pkg-02 | HIGH | P1 | Slow log subscriber blocks under buffer mutex | `runtime/pkg/logbuffer/buffer.go:38-48,52-55` | Snapshot clients outside the mutex and use bounded/non-blocking per-client delivery with drop accounting. |
| runtime-pkg-04 | MEDIUM | P2 | DuckDB write handles leak OTel callbacks | `runtime/pkg/rduckdb/db.go:955-957`; `runtime/pkg/rduckdb/generic.go:188-190` | Register DB stats only for long-lived handles or use an API that returns an unregisterable registration. |
| runtime-pkg-03 | MEDIUM | P1 | Telemetry HTTP can hang shutdown forever | `runtime/pkg/activity/sink_intake.go:57-68,121-131` | Use a timeout-bound client/request and context-bounded close. |
| runtime-pkg-05 | LOW | P2 | DuckDB AST JSON is built by string interpolation | `runtime/pkg/duckdbsql/ast_rewrite.go:97-130,251-262` | Construct Go values and `json.Marshal` them. |
| runtime-pkg-07 | LOW | P2 | File-watcher error field has a data race | `runtime/pkg/filewatcher/filewatcher.go:84-104` | Synchronize all `w.err` access with one mutex or atomic pointer; add `-race` coverage. |
| runtime-pkg-09 | LOW | P2 | Blob download materializes the entire listing | `runtime/pkg/blob/download.go:76-83` | Stream pages into the bounded download pipeline or cap object count. |
| runtime-pkg-10 | LOW | P2 | Download/copy paths use overly broad modes | `runtime/pkg/blob/download.go:230-242`; `runtime/pkg/fileutil/fileutil.go:67,102` | Use 0700 directories and 0600 files unless broader access is explicitly required. |
| runtime-pkg-11 | product-improvement | P3 | Log loss is invisible to clients | `runtime/pkg/logbuffer/buffer.go:41-48` | Emit per-buffer/per-subscriber dropped-line counters or marker events. |
| runtime-pkg-12 | product-improvement | P3 | Blob ingestion tuning is fixed and opaque | `runtime/pkg/blob/download.go:26-29,68,208` | Expose bounded connector options and progress/throughput metrics. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| runtime-pkg-01 | FIXED | `ResolveLocalPath` now cleans both absolute and relative paths and checks `finalPath == root` or `root + separator` containment at `runtime/pkg/fileutil/fileutil.go:317-340`; regression tests cover absolute `..` and sibling-prefix bypasses. | PR #9726, `ec882259d3cb6dbf38a168534eaf1b3b2a42074` |

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| runtime-pkg-06 | MEDIUM | P1 | The write escapes `tempDir`, but exploitability depends on attacker-controlled object keys in a configured bucket. | `runtime/pkg/blob/download.go:230-242`; reject `..`, verify canonical containment, and test hostile object names. |

Counts: **10 pending, 1 fixed/remove, 1 partial/needs-test, 0 invalid = 12 total**.

## Machine-readable manifest

```text
FIXED|runtime-pkg.md|runtime-pkg-01|HIGH|-|-|ResolveLocalPath sandbox traversal|runtime/pkg/fileutil/fileutil.go:317-340; PR#9726
PENDING|runtime-pkg.md|runtime-pkg-13|HIGH|P0|-|GetLogs limit zero panics|runtime/pkg/logbuffer/buffer.go:73-95
PENDING|runtime-pkg.md|runtime-pkg-02|HIGH|P1|-|Slow log subscriber blocks under mutex|runtime/pkg/logbuffer/buffer.go:38-48
PENDING|runtime-pkg.md|runtime-pkg-04|MEDIUM|P2|-|DuckDB write handles leak OTel callbacks|runtime/pkg/rduckdb/db.go:955-957
PENDING|runtime-pkg.md|runtime-pkg-03|MEDIUM|P1|-|Telemetry HTTP can hang shutdown|runtime/pkg/activity/sink_intake.go:57-68,121-131
PARTIAL|runtime-pkg.md|runtime-pkg-06|MEDIUM|P1|-|Blob object key can escape download temp directory|runtime/pkg/blob/download.go:230-242
PENDING|runtime-pkg.md|runtime-pkg-05|LOW|P2|-|DuckDB AST JSON uses unsafe interpolation|runtime/pkg/duckdbsql/ast_rewrite.go:97-130
PENDING|runtime-pkg.md|runtime-pkg-07|LOW|P2|-|File watcher error field races|runtime/pkg/filewatcher/filewatcher.go:84-104
PENDING|runtime-pkg.md|runtime-pkg-09|LOW|P2|-|Blob download eagerly collects all objects|runtime/pkg/blob/download.go:76-83
PENDING|runtime-pkg.md|runtime-pkg-10|LOW|P2|-|Downloaded and copied files use broad modes|runtime/pkg/blob/download.go:230-242; runtime/pkg/fileutil/fileutil.go:67,102
PENDING|runtime-pkg.md|runtime-pkg-11|product-improvement|P3|-|Log loss has no client signal|runtime/pkg/logbuffer/buffer.go:41-48
PENDING|runtime-pkg.md|runtime-pkg-12|product-improvement|P3|-|Blob ingestion tuning is hardcoded|runtime/pkg/blob/download.go:26-29
```
