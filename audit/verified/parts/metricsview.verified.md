# Verified audit: `metricsview.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **8** numbered findings. Every source ID is classified once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| metricsview-05 | medium | P1 | LIKE-to-regex corrupts Druid search semantics | `runtime/metricsview/astexpr.go:787-796` replaces `%` but does not quote regex metacharacters or translate `_` | Quote literal chunks, translate SQL wildcards deliberately, and add Druid search cases. |
| metricsview-06 | medium | P1 | Percent-of-total uses unsafe, rounded division | `runtime/metricsview/ast.go:558-569` emits bare `/` with `%#f` total | Use dialect safe-divide and a bound/full-precision total. |
| metricsview-09 | medium | P1 | NULL percent total aborts the query | `runtime/metricsview/executor/executor_rewrite_percent_of_totals.go:82-98` scans aggregates into `*float64` | Scan nullable values and combine with safe-divide NULL/zero semantics. |
| metricsview-02 | medium | P1 | Malformed cast expression panics | `runtime/metricsview/astexpr.go:146-181` indexes two operands without validating length | Validate arity and return a typed query error before indexing. |
| metricsview-04 | medium | P1 | Invalid grain can panic the pre-validation pass | `runtime/metricsview/ast.go:194-214` calls `ToTimeutil`; validation occurs later; panic is in `runtime/metricsview/query.go:321-344` | Validate every grain before min-grain comparison; add duplicate-time-dimension malformed input test. |
| metricsview-07 | low | P2 | Annotation OLAP handle map is never populated | `runtime/metricsview/executor/executor_validate.go:649-665` reads `olaps[connector]` but never assigns it | Store the acquired handle by connector and release each unique handle once. |
| metricsview-08 | low | P1 | Primary time dimension bypasses field access | `runtime/metricsview/ast.go:647-669`, explicitly before `CanAccessField` | Enforce field access or document/gate the exception explicitly; add deny-policy tests. |

## Fixed / Remove

None.

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| metricsview-01 | medium | P2 | Offset-loss asymmetry remains, but a production query shape with two pushdown spine levels and live OFFSET has not been demonstrated. | `runtime/metricsview/executor/executor_rewrite_limit_pushdown.go:10-74`; add a constructed two-level AST test before changing behavior. |

Counts: **7 pending, 0 fixed/remove, 1 partial/needs-test, 0 invalid = 8 total**.

## Machine-readable manifest

```text
PENDING|metricsview.md|metricsview-05|medium|P1|-|LIKE regex conversion does not escape metacharacters|runtime/metricsview/astexpr.go:787-796
PENDING|metricsview.md|metricsview-06|medium|P1|-|Percent of total divides by rounded raw literal|runtime/metricsview/ast.go:558-569
PENDING|metricsview.md|metricsview-09|medium|P1|-|NULL percent total hard-errors|runtime/metricsview/executor/executor_rewrite_percent_of_totals.go:82-98
PENDING|metricsview.md|metricsview-02|medium|P1|-|Malformed cast condition panics|runtime/metricsview/astexpr.go:146-181
PENDING|metricsview.md|metricsview-04|medium|P1|-|Invalid time grain can panic pre-validation|runtime/metricsview/ast.go:194-214; runtime/metricsview/query.go:321-344
PARTIAL|metricsview.md|metricsview-01|medium|P2|-|Nested limit pushdown can lose OFFSET|runtime/metricsview/executor/executor_rewrite_limit_pushdown.go:10-74
PENDING|metricsview.md|metricsview-07|low|P2|-|Annotation validation does not cache OLAP handles|runtime/metricsview/executor/executor_validate.go:649-665
PENDING|metricsview.md|metricsview-08|low|P1|-|Primary time dimension bypasses field access|runtime/metricsview/ast.go:647-669
```
