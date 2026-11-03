# Verified audit: `deep-metrics-sql.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **14** numbered findings. Every source ID is classified once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| deep-metrics-sql-06 | MEDIUM | P1 | `||` is parsed as logical OR | Parser mode lacks pipes-as-concat at `runtime/metricsview/metricssql/parser.go:45-53`; operator maps `opcode.Or` at `:487-511` | Enable and implement concat, or reject `||` explicitly with a positioned error. |
| deep-metrics-sql-07 | MEDIUM | P2 | Dynamic time is evaluated while parsing | `runtime/metricsview/metricssql/time_range_parser.go:16-81`; compiler TODO at `parser.go:40-41` | Lower time predicates to structured `Query.TimeRange` or a deferred expression evaluated with execution time. |
| deep-metrics-sql-08 | MEDIUM | P2 | Unsupported operators fail late | `runtime/metricsview/metricssql/parser.go:487-511` converts unknown opcodes into arbitrary `metricsview.Operator` values | Make operator conversion return `(Operator,error)` from an allowlist and support unary negative literals. |
| deep-metrics-sql-09 | MEDIUM | P1 | Canvas ref discovery swallows SQL parse errors | `runtime/canvases.go:138-144` only records refs when `err == nil` | Propagate a resource error (or at minimum record/log it) so security/ref resolution never silently degrades. |
| deep-metrics-sql-10 | MEDIUM | P2 | WHERE/ORDER BY names are not validated while parsing | `runtime/metricsview/metricssql/filter_parser.go:39-92`; `parser.go:357-367` | Validate against resolved dimension/measure maps with positioned suggestions; align top-level HAVING extraction. |
| deep-metrics-sql-11 | LOW | P2 | LIMIT conversion can overflow | `runtime/metricsview/metricssql/filter_parser.go:425-448`; `parser.go:330-353` casts uint64 to int unchecked | Range-check into int64, reject negative/overflow, and decide/document comma-limit syntax. |
| deep-metrics-sql-12 | LOW | P2 | LIKE is always case-insensitive; dates are ambiguous | `runtime/metricsview/metricssql/parser.go:503-504`; date parsing path in `filter_parser.go:469-526` | Preserve LIKE/ILIKE distinction and require ISO-8601 date operands. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| deep-metrics-sql-01 | FIXED | Truth parsing uses `node.True` and correctly models NOT/NULL semantics at `runtime/metricsview/metricssql/filter_parser.go:138-186`, with a truth-table test matrix. | PR #9712, `ac7f6c232` |
| deep-metrics-sql-02 | FIXED | Top-level FROM now rejects both right-hand joins and non-table left nodes at `runtime/metricsview/metricssql/parser.go:190-203`. | PR #9721, `9c94a590b` |
| deep-metrics-sql-03 | FIXED | `parseBetween` now branches on `node.Not` and regression tests cover NOT BETWEEN/precedence. | PR #9719, `2c1972bf4` |
| deep-metrics-sql-04 | FIXED | Decimal AST values are converted from `MyDecimal` to float64 at `runtime/metricsview/metricssql/filter_parser.go:425-448`, with comparison/IN tests. | PR #9720, `21cabb2f9` |
| deep-metrics-sql-05 | FIXED | Shared SELECT validation rejects unsupported clauses, validates GROUP BY, and rejects dropped aliases at `runtime/metricsview/metricssql/parser.go:76-101,135-167,239-327`. | PR #9716, `da97afd56` |

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| deep-metrics-sql-13 | LOW | P3 | PR #9716 stopped silently dropping plain aliases, but computed auto-alias quality, per-call parser allocation, and duplicate emitters remain. | `runtime/metricsview/metricssql/parser.go:45-53,239-273`; `filter_parser.go:17-35`; split residual cleanup into scoped issues/benchmarks. |
| deep-metrics-sql-14 | LOW | P2 | Template injection shape remains, but post-parse metrics security contains it to already-authorized data. | `runtime/resolvers/metrics_sql.go:62-68,82-119`; parameterize template values and add hostile-attribute tests. |

Counts: **7 pending, 5 fixed/remove, 2 partial/needs-test, 0 invalid = 14 total**.

## Machine-readable manifest

```text
FIXED|deep-metrics-sql.md|deep-metrics-sql-01|HIGH|-|-|IS FALSE compiled as IS TRUE|runtime/metricsview/metricssql/filter_parser.go:138-186; PR#9712
FIXED|deep-metrics-sql.md|deep-metrics-sql-02|HIGH|-|-|Top-level JOIN was silently discarded|runtime/metricsview/metricssql/parser.go:190-203; PR#9721
FIXED|deep-metrics-sql.md|deep-metrics-sql-03|HIGH|-|-|NOT BETWEEN compiled as BETWEEN|runtime/metricsview/metricssql/filter_parser.go:115-135; PR#9719
FIXED|deep-metrics-sql.md|deep-metrics-sql-04|HIGH|-|-|Decimal literals were lost|runtime/metricsview/metricssql/filter_parser.go:425-448; PR#9720
FIXED|deep-metrics-sql.md|deep-metrics-sql-05|HIGH|-|-|Top-level SELECT silently dropped clauses|runtime/metricsview/metricssql/parser.go:76-101,135-167,239-327; PR#9716
PENDING|deep-metrics-sql.md|deep-metrics-sql-06|MEDIUM|P1|-|Pipes concat compiles as logical OR|runtime/metricsview/metricssql/parser.go:45-53,487-511
PENDING|deep-metrics-sql.md|deep-metrics-sql-07|MEDIUM|P2|-|Time expressions evaluate during parsing|runtime/metricsview/metricssql/time_range_parser.go:16-81
PENDING|deep-metrics-sql.md|deep-metrics-sql-08|MEDIUM|P2|-|Unsupported operators fail late|runtime/metricsview/metricssql/parser.go:487-511
PENDING|deep-metrics-sql.md|deep-metrics-sql-09|MEDIUM|P1|-|Canvas ref discovery swallows parse errors|runtime/canvases.go:138-144
PENDING|deep-metrics-sql.md|deep-metrics-sql-10|MEDIUM|P2|-|Filter and sort names lack parse validation|runtime/metricsview/metricssql/filter_parser.go:39-92; runtime/metricsview/metricssql/parser.go:357-367
PENDING|deep-metrics-sql.md|deep-metrics-sql-11|LOW|P2|-|LIMIT conversion can overflow|runtime/metricsview/metricssql/filter_parser.go:425-448; runtime/metricsview/metricssql/parser.go:330-353
PENDING|deep-metrics-sql.md|deep-metrics-sql-12|LOW|P2|-|LIKE is always ILIKE and dates are ambiguous|runtime/metricsview/metricssql/parser.go:503-504
PARTIAL|deep-metrics-sql.md|deep-metrics-sql-13|LOW|P3|-|Alias and parser-maintenance bundle partly improved|runtime/metricsview/metricssql/parser.go:45-53,239-273; PR#9716
PARTIAL|deep-metrics-sql.md|deep-metrics-sql-14|LOW|P2|-|Template resolution precedes parsing|runtime/resolvers/metrics_sql.go:62-68,82-119
```
