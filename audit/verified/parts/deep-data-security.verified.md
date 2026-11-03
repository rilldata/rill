# Verified: deep-data-security.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **3** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| deep-data-security-02 | MEDIUM | P1 | Raw-SQL security filters interpolate unescaped attributes | `runtime/security.go:759-767` renders `rule.Sql` with `parser.ResolveTemplate` and stores raw SQL; `runtime/metricsview/ast.go:366-395` carries the raw row filter without parameterizing its embedded template values | Replace raw interpolation with parameter binding/typed expressions. If compatibility requires raw SQL, expose an explicit SQL-string escaping helper and reject unsafe templates. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| deep-data-security-01 | FIXED — remove | `runtime/metricsview/executor/executor_summary.go:46-49,158-190` compiles row/query filters via `SecurityFilterSQL`, combines them with the sample predicate, and tests cover user attributes plus additional rules | PR #9717, commit `d40f49708adb7a73633f7b7ab2f43f8672201896` |
| deep-data-security-03 | INVALID AS BUG — remove | `runtime/metricsview/executor/executor.go:170-177` now explicitly documents global, cacheable timestamps as intended semantics and the trade-off. PR #9717 deliberately retained unfiltered time bounds while fixing value summaries. Track a product-semantics proposal separately if tenant-scoped bounds are desired. | Design clarification merged with PR #9717, commit `d40f49708adb7a73633f7b7ab2f43f8672201896` |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| — | — | — | None | No partial or test-gated findings remain in this source. |

Counts: **1 pending, 1 fixed/remove, 0 partial/needs-test, 1 invalid/remove = 3 total**.

## Machine-readable ledger

```text
FIXED|deep-data-security.md|deep-data-security-01|HIGH|-|-|Metrics summary leaked values across row security|runtime/metricsview/executor/executor_summary.go:46-49,158-190; PR #9717 d40f49708adb7a73633f7b7ab2f43f8672201896
PENDING|deep-data-security.md|deep-data-security-02|MEDIUM|P1|-|Raw-SQL row-filter attribute injection|runtime/security.go:759-767; runtime/metricsview/ast.go:366-395
INVALID|deep-data-security.md|deep-data-security-03|LOW|-|-|Global timestamp bounds are intentional semantics|runtime/metricsview/executor/executor.go:170-177; PR #9717 d40f49708adb7a73633f7b7ab2f43f8672201896
```
