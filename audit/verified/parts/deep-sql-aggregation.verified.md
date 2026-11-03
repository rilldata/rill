# Verified audit: `deep-sql-aggregation.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **16** numbered findings. Every source ID is classified once below; none was removed by the recent merged fixes.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| deep-sql-aggregation-01 | HIGH | P1 | Two-phase pct-of-total comparison emits invalid SQL | `runtime/metricsview/executor/executor_rewrite_two_phase_comparisons.go:28-45,100-107` excludes percent totals from inline base measures | Bail out for unsupported derived measures or carry their dependencies/results into phase two; add compute matrix tests. |
| deep-sql-aggregation-02 | HIGH | P1 | Comparison top-N runs before HAVING | Two-phase clears HAVING at `executor_rewrite_two_phase_comparisons.go:47-66`; approximate pushdown at `executor_rewrite_approx_comparisons.go:108-170` has no HAVING guard | Skip pushdown/two-phase with HAVING, or safely push HAVING into phase one. |
| deep-sql-aggregation-03 | MEDIUM | P1 | Half-open ranges bypass query time caps | `runtime/metricsview/executor/executor_enforce_query_limits.go:41-58` uses `End.Sub(Start)` without resolving missing bounds | Resolve/reject half-open bounds before cap comparison and test both directions. |
| deep-sql-aggregation-04 | MEDIUM | P1 | Half-open ranges can select incomplete rollups | `runtime/metricsview/executor/executor_rewrite_rollup.go:132-133,220-230,503-537` | Normalize each absent side to base min/max before coverage/alignment checks. |
| deep-sql-aggregation-05 | MEDIUM | P1 | Missing comparison range panics while formatting error | `runtime/metricsview/ast.go:603-604` detects nil then dereferences it in the error arguments | Capture a safe comparison dimension before the guard; add no-range comparison-time test. |
| deep-sql-aggregation-06 | MEDIUM | P1 | Spine without time range panics | `runtime/metricsview/executor/executor_rewrite_time.go:45-54` copies `qry.TimeRange` without a nil guard | Return a typed validation error when a spine lacks a time range. |
| deep-sql-aggregation-07 | MEDIUM | P1 | DST comparisons mis-join local buckets | Comparison adjustment in `runtime/metricsview/ast.go:1525-1576`; DuckDB/Druid interval helpers shift outside local-time truncation | Shift calendar offsets in local time and add per-dialect DST boundary tests. |
| deep-sql-aggregation-08 | MEDIUM | P1 | Druid native search uses logical names | `runtime/metricsview/executor/executor.go:830-851` appends `f.Name`; native request consumes those strings | Use physical column/expression mapping and map values back to logical names; test renamed dimensions. |
| deep-sql-aggregation-09 | MEDIUM | P1 | ClickHouse wrap corrupts CTE leaf metadata | `executor_wrap_clickhouse_computed_time_dim.go:75-87`; `runtime/metricsview/ast.go:709-728` copies CTE/raw flags into the inner node | Reset/transfer CTE metadata deliberately or skip CTE leaves; add CH SQL-generation coverage. |
| deep-sql-aggregation-10 | MEDIUM | P1 | Pivot staging removes the source row cap | `runtime/metricsview/executor/executor_pivot.go:98-103` clears limit, then stages all rows at `:143-165` before pivot checks | Impose a hard underlying bound derived from pivot caps and fail fast when exceeded. |
| deep-sql-aggregation-11 | LOW | P2 | Druid exactify keeps HAVING without its measures | `runtime/metricsview/executor/executor_rewrite_druid_exactify.go:33-62` copies HAVING but only adds the sort measure | Include all HAVING dependencies or skip exactify for unsupported HAVING. |
| deep-sql-aggregation-12 | LOW | P2 | Two-phase comparison can apply OFFSET twice | `executor_rewrite_two_phase_comparisons.go:47-66` copies offset; only the approximate walk clears the root | Clear root limit/offset inside the two-phase rewrite; test with approximate comparisons disabled. |
| deep-sql-aggregation-13 | LOW | P2 | Sub-day rollup ignores fractional-offset timezone | `executor_rewrite_rollup.go:312-356`; UTC/local assumptions in `runtime/metricsview/executor/rollup_grain.go:86-105` | Reject incompatible sub-day timezone offsets or align consistently in UTC; add half-hour-zone tests. |
| deep-sql-aggregation-14 | LOW | P2 | Multi-grain comparison ignores physical time expression | `runtime/metricsview/ast.go:1550-1566` rebuilds from escaped logical `timeDim` | Resolve the actual dimension expression before applying comparison offset. |
| deep-sql-aggregation-15 | LOW | P2 | Array NIN drops NULL-array rows | `runtime/metricsview/astexpr.go:662-702`; scalar NIN null handling differs | Emit `NOT contains(...) OR array_column IS NULL`; test NULL arrays/elements. |
| deep-sql-aggregation-16 | MEDIUM | P1 | Pre-bound timestamp results omit `Now` | Builders populate min/max/watermark only at `runtime/resolvers/metricsview_time_range.go:191-253` and `runtime/queries/metricsview_aggregation.go:231-287`; evaluation consumes zero `Now` at `executor_rewrite_time.go:83-102` | Populate `Now` or default it during evaluation; test MCP/metrics resolver with a now-anchored expression. |

## Fixed / Remove

None.

## Partial / Needs-test

None. Dialect-specific entries above have direct current code evidence and remain actionable; their recommended regression tests are part of the fixes, not a reason to defer classification.

Counts: **16 pending, 0 fixed/remove, 0 partial/needs-test, 0 invalid = 16 total**.

## Machine-readable manifest

```text
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-01|HIGH|P1|-|Two-phase percent-total comparison emits invalid SQL|runtime/metricsview/executor/executor_rewrite_two_phase_comparisons.go:28-45,100-107
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-02|HIGH|P1|-|Comparison top-N runs before HAVING|runtime/metricsview/executor/executor_rewrite_two_phase_comparisons.go:47-66; executor_rewrite_approx_comparisons.go:108-170
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-03|MEDIUM|P1|-|Half-open ranges bypass time caps|runtime/metricsview/executor/executor_enforce_query_limits.go:41-58
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-04|MEDIUM|P1|-|Half-open ranges can select incomplete rollup|runtime/metricsview/executor/executor_rewrite_rollup.go:132-133,220-230,503-537
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-05|MEDIUM|P1|-|Missing comparison range panics error path|runtime/metricsview/ast.go:603-604
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-06|MEDIUM|P1|-|Spine without time range panics|runtime/metricsview/executor/executor_rewrite_time.go:45-54
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-07|MEDIUM|P1|-|DST comparisons mis-join buckets|runtime/metricsview/ast.go:1525-1576
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-08|MEDIUM|P1|-|Druid search uses logical dimension names|runtime/metricsview/executor/executor.go:830-851
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-09|MEDIUM|P1|-|ClickHouse wrap corrupts CTE metadata|runtime/metricsview/executor/executor_wrap_clickhouse_computed_time_dim.go:75-87; runtime/metricsview/ast.go:709-728
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-10|MEDIUM|P1|-|Pivot staging is unbounded|runtime/metricsview/executor/executor_pivot.go:98-103,143-165
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-11|LOW|P2|-|Druid exactify omits HAVING measures|runtime/metricsview/executor/executor_rewrite_druid_exactify.go:33-62
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-12|LOW|P2|-|Two-phase comparison double-applies OFFSET|runtime/metricsview/executor/executor_rewrite_two_phase_comparisons.go:47-66
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-13|LOW|P2|-|Sub-day rollup ignores fractional timezone|runtime/metricsview/executor/executor_rewrite_rollup.go:312-356; runtime/metricsview/executor/rollup_grain.go:86-105
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-14|LOW|P2|-|Multi-grain comparison ignores physical time expression|runtime/metricsview/ast.go:1550-1566
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-15|LOW|P2|-|Array NIN excludes NULL arrays|runtime/metricsview/astexpr.go:662-702
PENDING|deep-sql-aggregation.md|deep-sql-aggregation-16|MEDIUM|P1|-|Pre-bound timestamps omit Now|runtime/resolvers/metricsview_time_range.go:191-253; runtime/queries/metricsview_aggregation.go:231-287
```
