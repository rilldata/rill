# Verified audit: `queries.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **12** numbered findings. Every source ID is classified once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| queries-02 | MEDIUM | P1 | Millisecond grain panics null-fill | `runtime/queries/timeutil.go:12-35` returns `InfDuration`; unchecked assertions remain at `runtime/queries/metricsview_timeseries.go:348-350,379-381` | Reject unsupported grains or return `(duration,error)` and remove concrete assertions. |
| queries-04 | MEDIUM | P2 | Profiling APIs execute caller SQL expressions | `runtime/queries/column_topk.go:70-78`; `runtime/queries/column_timeseries.go:638-642` | Keep strictly behind `ReadProfiling`, document trusted SQL, and allowlist/parse aggregation expressions. |
| queries-07 | LOW | P2 | Column timeseries leaks cursor on `rows.Err` | `runtime/queries/column_timeseries.go:181-222`; error at `:217-219` returns before close | Defer `rows.Close()` immediately after acquisition; test an injected cursor error. |
| queries-09 | LOW | P2 | Constant-column histogram derives invalid buckets | `runtime/queries/column_numeric_histogram.go:285-302` computes step/count when range is zero | Return one deterministic bucket for `min == max`; add constant/all-null tests. |
| queries-10 | LOW | P3 | Comparison measure indexes are dead bookkeeping | Fields at `runtime/queries/metricsview_comparison_toplist.go:47-50`, written at `:451-474`, never read | Remove both fields/counters and simplify metadata construction. |
| queries-11 | LOW | P2 | Aggregation export always probes time range | `runtime/queries/metricsview_aggregation.go:165-179` binds timestamps for every time-dimensioned view | Bind only when rollup routing needs it, while preserving header time-range rewriting. |
| queries-12 | LOW | P2 | Null-filled series silently truncates at 5000 buckets | `runtime/queries/metricsview_timeseries.go:328-376` | Return a typed cap error/truncation flag or make the cap configurable and explicit. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| queries-01 | FIXED | Export now nil-checks the request time range at `runtime/queries/metricsview_aggregation.go:165-168` and rewrites before header access, with regression tests. | PR #9715, `101dce8ea` |
| queries-05 | FIXED | Timeseries export supplies `CODE_TIMESTAMP` at `runtime/queries/metricsview_timeseries.go:177-186`; Parquet rejects unmapped types instead of panicking at `runtime/queries/metricsview.go:343-351`. | PR #9725, `5479bf4f3` |

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| queries-03 | MEDIUM | P1 | Nil dereference remains, but requires StarRocks plus exactly one non-null row (`stddev_samp` NULL). | `runtime/queries/column_desc_stats.go:88-97`; guard every pointer and integration-test the StarRocks case. |
| queries-06 | LOW | P2 | Panic branches remain, but supported drivers normally scan bucket timestamps into accepted types. | `runtime/queries/metricsview_timeseries.go:247-260`; `runtime/queries/column_timeseries.go:190-197`; return errors and test a string timestamp driver stub. |
| queries-08 | LOW | P2 | `max` assertion remains guarded only by non-nil `min`; inconsistent null aggregates are dialect/data dependent. | `runtime/queries/column_time_range.go:100-114`; independently type-check min and max and test partial-null rows. |

Counts: **7 pending, 2 fixed/remove, 3 partial/needs-test, 0 invalid = 12 total**.

## Machine-readable manifest

```text
FIXED|queries.md|queries-01|HIGH|-|-|Aggregation export panicked without time range|runtime/queries/metricsview_aggregation.go:165-168; PR#9715
PENDING|queries.md|queries-02|MEDIUM|P1|-|Millisecond grain panics null fill|runtime/queries/timeutil.go:12-35; runtime/queries/metricsview_timeseries.go:348-350,379-381
PARTIAL|queries.md|queries-03|MEDIUM|P1|-|StarRocks single-row stats can dereference nil sd|runtime/queries/column_desc_stats.go:88-97
PENDING|queries.md|queries-04|MEDIUM|P2|-|Profiling APIs inline SQL expressions|runtime/queries/column_topk.go:70-78; runtime/queries/column_timeseries.go:638-642
FIXED|queries.md|queries-05|MEDIUM|-|-|Timeseries Parquet schema omitted time type|runtime/queries/metricsview_timeseries.go:177-186; runtime/queries/metricsview.go:343-351; PR#9725
PARTIAL|queries.md|queries-06|LOW|P2|-|Unexpected timestamp scan type panics|runtime/queries/metricsview_timeseries.go:247-260; runtime/queries/column_timeseries.go:190-197
PENDING|queries.md|queries-07|LOW|P2|-|Column timeseries leaks cursor on rows error|runtime/queries/column_timeseries.go:181-222
PARTIAL|queries.md|queries-08|LOW|P2|-|Column time range asserts max type under min guard|runtime/queries/column_time_range.go:100-114
PENDING|queries.md|queries-09|LOW|P2|-|Constant histogram computes invalid bucket count|runtime/queries/column_numeric_histogram.go:285-302
PENDING|queries.md|queries-10|LOW|P3|-|Comparison measure indexes are dead|runtime/queries/metricsview_comparison_toplist.go:47-50,451-474
PENDING|queries.md|queries-11|LOW|P2|-|Aggregation export performs redundant timestamp probe|runtime/queries/metricsview_aggregation.go:165-179
PENDING|queries.md|queries-12|LOW|P2|-|Null-filled timeseries silently truncates|runtime/queries/metricsview_timeseries.go:328-376
```
