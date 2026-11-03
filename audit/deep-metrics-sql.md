# Metrics SQL — Deep-Dive Audit

**Date:** 2026-07-20
**Focus:** correctness and architecture of the metrics SQL dialect — the compiler that turns `SELECT … FROM <metrics_view> …` text into a `metricsview.Query`, the standalone WHERE-fragment parser, parse-time time-expression evaluation, and the capability gap versus the internal query representation.
**Scope:** `runtime/metricsview/metricssql/{parser,filter_parser,time_range_parser}.go`; consumers `runtime/resolvers/{metrics_sql,builtin_metrics_sql,sql,text}.go`, `runtime/queries/metricsview_aggregation.go`, `runtime/parser/parse_canvas.go`, `runtime/canvases.go`, `runtime/server/query_resolver.go`, `cli/cmd/query/query.go`; cross-referenced `runtime/metricsview/{query,ast,astexpr,query_expression_sql,query_expression_analyze}.go` and `runtime/metricsview/executor/executor_rewrite_time.go`; docs `docs/docs/developers/build/metrics-view/metrics-sql.md`.
**Method:** full read of the ~1,300-line core package and tests, empirical reproduction of every correctness finding by running the real compiler against crafted inputs (see per-finding "Reproduce"), and two parallel deep-dives mapping usage sites and the `metricsview.Query` capability surface.
**Health:** needs-attention — the top-level compile path silently drops unsupported clauses (JOIN / GROUP BY / DISTINCT / CTE / alias) and four filter constructs (`IS FALSE`, `NOT BETWEEN`, decimals, `||`) produce wrong results rather than errors.
**Remediation (2026-07-25):** findings 01–05 are fixed by [#9712](https://github.com/rilldata/rill/pull/9712), [#9721](https://github.com/rilldata/rill/pull/9721), [#9719](https://github.com/rilldata/rill/pull/9719), [#9720](https://github.com/rilldata/rill/pull/9720), and [#9716](https://github.com/rilldata/rill/pull/9716). Finding 08 remains pending P0; #9716 did not add an operator allowlist.

## Overview

Metrics SQL is a deliberately small dialect that compiles to `metricsview.Query` and inherits all downstream row-/field-level security. It is used far more widely than its size suggests: custom APIs, alerts, reports, canvas custom-chart components, canvas default filters, markdown templating (`metrics_sql` / `metrics_sql_rows`), the `whereSQL`/`havingSQL` parameters on every major metrics query (per request, on the hot path), the CLI (`rill query --resolver metrics_sql`), and the admin-only builtin `/metrics-sql` HTTP API. AI features do **not** generate metrics SQL — runtime AI tools and the MCP server all issue structured metricsview queries — with one exception: the canvas custom-chart AI agent authors `metrics_sql` into YAML, which makes parser error quality an AI-convergence concern.

The compiler is built on the TiDB (MySQL) parser in ANSI mode. The headline problem is not what it rejects but what it **silently accepts and discards**. The top-level `Compiler.Parse` path performs almost no clause validation: a JOIN keeps only the left table and drops the ON condition (F1); `DISTINCT`, `GROUP BY`, CTEs, and column aliases are dropped without a word (F7). Ironically the `IN (SELECT …)` subquery path (`filter_parser.go:218-250`) validates ~15 unsupported features exhaustively — the top-level path validates none of the same list. On top of that are four verified filter-level correctness bugs that return inverted or meaningless results: `IS FALSE` behaves identically to `IS TRUE` (F6, the most severe — a fully inverted boolean filter), `NOT BETWEEN` compiles to `BETWEEN` (F2), decimal literals are lost to an empty value (F3), and `||` compiles to logical OR (F4).

Architecturally the biggest limitation is **parse-time evaluation of time expressions**. `now()`, `time_range_start()`, `time_range_end()`, and `INTERVAL` arithmetic are all evaluated eagerly during compilation and baked into the query as literal timestamps — even though the executor already owns a lazy, execution-time, cache-key-aware time-resolution pipeline (`RewriteQueryTimeRanges`) that metrics SQL cannot reach because `metricsview.Expression` has no deferred/function node. This is acknowledged in a TODO at `parser.go:39-40`. It costs two extra warehouse round-trips per time range and pollutes/staleness-breaks the result cache (F10).

The metrics SQL surface also exposes only a fraction of `metricsview.Query`: pivots, spines, comparison time ranges, and every computed measure (`comparison_value`, `comparison_delta`, `comparison_ratio`, `percent_of_total`, `count_distinct`, `uri`) are unreachable from SQL (§4). Closing that gap — behind a strict deny-by-default compiler — is the main expansion opportunity.

## Findings

### deep-metrics-sql-01 [HIGH] `IS FALSE` / `IS NOT FALSE` compile identically to `IS TRUE` — fully inverted boolean filter

**Status (2026-07-25):** ✅ Fixed and merged in [#9712](https://github.com/rilldata/rill/pull/9712).

- **Category:** correctness  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/filter_parser.go:136-157` (`parseIsTruthOperation`)
- **Problem:** `parseIsTruthOperation` emits `eq(expr, "TRUE")` for the non-negated case and `neq(expr, "TRUE")` for the negated case, but it never inspects `node.True` (TiDB's `ast.IsTruthExpr.True` distinguishes `IS TRUE` from `IS FALSE`). As a result `flag IS FALSE` produces the exact same IR as `flag IS TRUE`, and `flag IS NOT FALSE` the same as `flag IS NOT TRUE`. A user filtering for false rows gets the true rows and vice versa. Two secondary defects in the same function: (a) it compares against the *string* `'TRUE'`, relying on implicit casts that work in DuckDB but are not reliable on ClickHouse/Druid; (b) the NULL semantics are wrong — `IS NOT TRUE` must include NULL rows, but `!= 'TRUE'` excludes them.
- **Failure scenario:** A report or alert with `WHERE is_active IS FALSE` silently reports on active rows; the operator sees inverted data with no error.
- **Reproduce:**
  ```go
  // in runtime/metricsview/metricssql, scratch test:
  expr, _ := ParseFilter("flag IS FALSE")
  b, _ := json.Marshal(expr); fmt.Println(string(b))
  // => {"cond":{"op":"eq","exprs":[{"name":"flag"},{"val":"TRUE"}]}}  -- identical to IS TRUE
  ```
- **Recommendation:** Branch on `node.True`; emit a boolean-typed value (`{Value: true/false}`), and model NULL-inclusive semantics explicitly for the negated forms (e.g. `expr IS NULL OR expr != TRUE`). Add tests for all four combinations.

### deep-metrics-sql-02 [HIGH] Top-level `JOIN` silently drops the second table and the `ON` condition

**Status (2026-07-25):** ✅ Fixed and merged in [#9721](https://github.com/rilldata/rill/pull/9721).

- **Category:** correctness  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/parser.go:134-149` (`parseFrom`)
- **Problem:** `parseFrom` only type-asserts `TableRefs.Left`. For `a JOIN b ON …`, TiDB builds `Join{Left: TableSource(a), Right: TableSource(b), On: …}` where `Left` *is* a `TableSource`, so the assertion passes and the right table plus the join condition are discarded with no error. Comma-joins (`FROM a, b`) *do* error (there `Left` is a nested `Join`), so behavior is also inconsistent. The subquery path already guards this correctly with `if sel.From.TableRefs.Right != nil` (`filter_parser.go:256`).
- **Failure scenario:** `SELECT pub FROM ad_bids_metrics JOIN other ON …` runs against only `ad_bids_metrics`, ignoring the intended join — wrong result set, no error surfaced.
- **Reproduce:** `Compiler.Parse("select pub from mv1 join mv2 on mv1.pub = mv2.pub")` → succeeds with `metrics_view:"mv1"`, `mv2` and ON gone.
- **Recommendation:** In `parseFrom`, error when `node.TableRefs.Right != nil` (and when `Left` is itself a `*ast.Join`).

### deep-metrics-sql-03 [HIGH] `NOT BETWEEN` compiles to `BETWEEN` — inverted range filter

**Status (2026-07-25):** ✅ Fixed and merged in [#9719](https://github.com/rilldata/rill/pull/9719).

- **Category:** correctness  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/filter_parser.go:375-409` (`parseBetween`)
- **Problem:** `parseBetween` never reads `n.Not`, so `x NOT BETWEEN 1 AND 10` produces the same `(x >= 1 AND x <= 10)` as `x BETWEEN 1 AND 10`. Docs list `NOT BETWEEN` as supported. Every row the user meant to exclude is included and vice versa.
- **Reproduce:** `ParseFilter("x NOT BETWEEN 1 AND 10")` → `{op:and,[{gte,x,1},{lte,x,10}]}` (identical to `BETWEEN`).
- **Recommendation:** When `n.Not`, emit `(x < low OR x > high)`; add tests for both forms.

### deep-metrics-sql-04 [HIGH] Decimal literals are lost — comparisons/IN against non-integers become meaningless

**Status (2026-07-25):** ✅ Fixed and merged in [#9720](https://github.com/rilldata/rill/pull/9720).

- **Category:** correctness  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/filter_parser.go:411-428` (`parseValueExpr`)
- **Problem:** `parseValueExpr` special-cases only `int64`/`uint64`. Decimal literals arrive as `*test_driver.MyDecimal` (production: `*types.MyDecimal`) and pass through unconverted, serializing to `{}`. Any filter with a non-integer numeric literal reaches the OLAP layer with an empty value.
- **Failure scenario:** `WHERE bid_price > 10.5` or `WHERE x IN (1.5, 2.5)` filters against `{}` — empty/garbage predicate; wrong rows or execution error depending on dialect.
- **Reproduce:** `ParseFilter("m > 10.5")` → `{op:gt,[{name:m},{val:{}}]}`.
- **Recommendation:** In `parseValueExpr`, handle the decimal type (via `String()`/`ToFloat64()`) and `float64`; add numeric-literal tests including IN lists.

### deep-metrics-sql-05 [HIGH] Silent clause drops in top-level SELECT: `DISTINCT`, `GROUP BY`, CTE, column alias

**Status (2026-07-25):** ✅ Fixed and merged in [#9716](https://github.com/rilldata/rill/pull/9716) (unsupported clauses now reject; `DISTINCT` and matching `GROUP BY` are supported explicitly).

- **Category:** correctness  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/parser.go:61-121` (`Parse`), `:182-214` (`parseSelect`, `TODO no alias handling` at `:186`)
- **Problem:** `Parse` never consults `stmt.SelectStmtOpts.Distinct`, `stmt.GroupBy`, `stmt.With`, or `stmt.WindowSpecs`, and `parseSelect` never reads `field.AsName` for plain columns. All of the following "succeed" with the offending clause silently discarded: `SELECT DISTINCT pub …`; `SELECT pub, measure_0 … GROUP BY dom` (even a *mismatched* group-by is dropped); `WITH x AS (…) SELECT …`; `SELECT pub AS publisher …` (result column stays `pub`). The alias drop is even asserted by a passing test (`parser_test.go:148-152`). Because custom APIs return JSON keyed by column name, a silently different column name is a correctness bug for programmatic consumers. Contrast `parseInSubquery` which rejects every one of these for subqueries.
- **Reproduce:** `Compiler.Parse("select pub as publisher from mv")` → `dimensions:[{name:"pub"}]`; `"select distinct pub from mv"` and `"with x as (select pub from mv) select pub from mv"` both succeed with the clause gone.
- **Recommendation:** Extract the subquery guard list (`filter_parser.go:218-250`) into a shared `validateSelectStmt` and call it from the top level too; implement plain-column aliases (see deep-metrics-sql-13 / §5.4).

### deep-metrics-sql-06 [MEDIUM] `||` (string concat) compiles to logical OR under the MySQL grammar
- **Category:** correctness  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/parser.go:391-392` (`operator`, `opcode.Or`), parser mode `parser.go:48`
- **Problem:** Despite `ModeANSI|ModeANSIQuotes`, `PIPES_AS_CONCAT` is not enabled, so TiDB parses `||` as logical OR. `a || b = 'ab'` becomes `OR(a, eq(b,'ab'))` — a silently mis-shaped boolean. A user writing standard-SQL concatenation gets wrong results with no error.
- **Reproduce:** `ParseFilter("a || b = 'ab'")` → `{op:or,[{name:a},{op:eq,[b,'ab']}]}`.
- **Recommendation:** Either add `mysql.ModePipesAsConcat` and support concat, or explicitly reject `||` with a clear error. Decide as part of the dialect definition (§5.5).

### deep-metrics-sql-07 [MEDIUM] Time expressions evaluated at parse time — extra round-trips, cache pollution, no execution-time resolution
- **Category:** architecture / performance  **Confidence:** high  **Location:** `runtime/metricsview/metricssql/time_range_parser.go:16-82`, `filter_parser.go:458-501`; TODO at `parser.go:39-40`; unused path `runtime/metricsview/executor/executor_rewrite_time.go:62-111`
- **Problem:** `now()`, `date_add`/`date_sub`, and `time_range_start`/`time_range_end` are evaluated during compilation and baked in as literal timestamps. `time_range_start`/`end` each call `GetTimestamps` (a min/max/watermark query) at parse time — two extra warehouse round-trips per range, and start/end are parsed independently so it is a pair per range. Meanwhile the executor already resolves rilltime lazily at execution time via `RewriteQueryTimeRanges` → `ResolveTimeRange`, keyed off an `executionTime` anchor that participates in the cache key (`executor.go:113`). The root cause is that `metricsview.Expression` has only `Name | Value | Condition | Subquery` node kinds (`query.go:236-241`) — no deferred/function node — so the compiler *cannot* emit "resolve later". Baking `now()` into a literal both defeats result caching (every compile differs) and, if a compiled query were cached, would freeze time. Bonus inconsistency: `time_range_start` yields a `time.Time` value while `now()` yields an RFC3339 *string*.
- **Recommendation:** (a) Lower time-dimension predicates into the structured `Query.TimeRange{Expression, TimeDimension}` so the executor resolves them (removes both round-trips and reuses existing cache-aware machinery); (b) longer term add a deferred function node to `metricsview.Expression` for `now()`/`date_add` in arbitrary positions. See §5.2.

### deep-metrics-sql-08 [MEDIUM] Unhandled operators pass parsing and fail late with confusing errors
- **Category:** correctness / UX  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/parser.go:373-399` (`operator` default arm)
- **Problem:** `operator()` falls through to `metricsview.Operator(op.String())` for anything unmapped, so arithmetic and exotic operators parse "successfully" into invalid IR: `measure_0 + 1 > 5` → op `"plus"`, `a XOR b` → `"xor"`, `a % 2 = 0` → `"mod"`. These fail downstream in the AST builder with `invalid expression operator "plus"` (`astexpr.go:140`) — no position, no hint — instead of a clear parse-time rejection. The `-` (unary minus) case is worse: `m > -5` errors with the internal `unsupported type *ast.UnaryOperationExpr, expression "-5"`, making filtering on negative numbers impossible.
- **Reproduce:** `ParseFilter("measure_0 + 1 > 5")` succeeds with op `"plus"`; `ParseFilter("m > -5")` errors on `UnaryOperationExpr`.
- **Recommendation:** Whitelist operators in `operator()` and return a positioned parse error for the rest; add a `*ast.UnaryOperationExpr` case handling `opcode.Minus` over a value (negate the literal).

### deep-metrics-sql-09 [MEDIUM] Canvas ref discovery swallows parse errors — silent security-context degradation
- **Category:** correctness / security  **Confidence:** high  **Location:** `runtime/canvases.go:140-141` (`if err == nil { … }`)
- **Problem:** During `ResolveCanvas`, referenced metrics views are discovered by compiling each `metrics_sql` component with `Compiler.Parse`; the result is used only when `err == nil`. A component query that fails to parse contributes no metrics-view ref, and since that ref list feeds security resolution for the canvas response, a malformed query degrades silently rather than erroring — the failure is invisible to author and viewer.
- **Recommendation:** Log the parse error at minimum; ideally propagate it so the canvas surfaces a resource error like other parse failures.
- **Note:** Canvas *default filters* (`parse_canvas.go:514`, `ParseFilter`) do surface parse failures as a resource parse-error state, and `whereSQL`/`havingSQL` (`metricsview_aggregation.go:576`) surface as per-request errors — so this swallow is the one silent path among the ParseFilter consumers.

### deep-metrics-sql-10 [MEDIUM] No parse-time name validation for WHERE / ORDER BY; late, position-less errors
- **Category:** UX / correctness  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/parser.go:243-255` (`parseOrderBy`), `filter_parser.go:37-91`
- **Problem:** `WHERE nosuchcol = 5` and `ORDER BY nonexistent` both parse successfully and fail later in AST construction (`invalid dimension reference %q`, `astexpr.go:734`), or not at all in bare-`ParseFilter` contexts (canvas default filters validate nothing). The compiler already holds `q.dims`/`q.measures` (built in `parseFrom`) but only consults them in `parseSelect`. Also, the top-level `parseHaving` does not extract HAVING measures the way `parseInSubquery` does (`filter_parser.go:319-330`), so `HAVING measure_0 > 10` requires also selecting `measure_0` — the same feature behaving two ways.
- **Recommendation:** Validate dimension/measure names in WHERE/HAVING/ORDER BY against the spec at parse time with positioned errors and "did you mean" suggestions; give `ParseFilter` an optional spec argument (the canvas path has the spec available). Unify HAVING-measure extraction across top-level and subquery paths.

### deep-metrics-sql-11 [LOW] `LIMIT` overflow wraps to −1; undocumented MySQL `LIMIT x, y` accepted
- **Category:** correctness  **Confidence:** high (verified)  **Location:** `runtime/metricsview/metricssql/parser.go:216-241` (`parseLimit`)
- **Problem:** `parseLimit` casts `uint64 → int` unchecked, so `LIMIT 18446744073709551615` becomes `-1` (either an error or "unlimited" downstream). Separately, MySQL's `LIMIT 5, 10` (offset 5, limit 10) is silently accepted but undocumented, and `LIMIT -1` fails with a raw TiDB error leaking internals.
- **Reproduce:** `Compiler.Parse("select pub from mv limit 18446744073709551615")` → `"limit":-1`; `"… limit 5, 10"` → `limit:10, offset:5`.
- **Recommendation:** Bound-check and reject negative/overflowing limits with a clean error; decide whether `LIMIT x,y` is part of the dialect and document or reject accordingly.

### deep-metrics-sql-12 [LOW] `LIKE` silently means `ILIKE`; no case-sensitive match; locale-ambiguous date parsing
- **Category:** correctness / dialect  **Confidence:** high (verified for LIKE; medium for dates)  **Location:** `runtime/metricsview/metricssql/parser.go:389-390` (LIKE→Ilike), `filter_parser.go:471` (`dateparse.ParseAny`)
- **Problem:** `opcode.Like` maps to `OperatorIlike`, so all LIKE is case-insensitive with no way to request case-sensitive matching (docs do note case-insensitivity, so this is a deliberate-but-irreversible dialect choice). Separately, `date_add`/`date_sub` parse their date operand with `dateparse.ParseAny`, which is locale-ambiguous: `'30-07-2024'` only works because 30 > 12; `'05-07-2024'` silently parses as May 7.
- **Recommendation:** Support an explicit `ILIKE` keyword (TiDB's `PatternLikeOrIlikeExpr` distinguishes them) and consider making `LIKE` case-sensitive in a future dialect version; restrict date operands to ISO-8601 and error otherwise.

### deep-metrics-sql-13 [LOW] Unusable auto-generated aliases; per-request parser construction; internal code duplication
- **Category:** maintainability / performance  **Confidence:** high  **Location:** `parser.go:196-208, 324-371`; `filter_parser.go:17`; `query_expression_sql.go` vs `astexpr.go`
- **Problem:** Several smaller items: (a) computed dimensions get unusable auto aliases — `date_trunc('SECOND', timestamp)` yields a result column literally named `DATE_TRUNC('SECOND', ""timestamp"")` and ordering on it requires reproducing that exact string (`parser_test.go:143`). (b) `ParseFilter` builds a fresh TiDB parser on every call and sits on hot per-request paths (`metricsview_aggregation.go:576`, called from toplist/totals/timeseries/rows/comparison); `newMetricsSQL` builds a fresh `Compiler` per request, so the "re-use the Compiler" doc comment never applies — a `sync.Pool` would remove per-request allocation. (c) `add`/`sub` (`parser.go:331-371`) are mirror-image duplicates; one function with a sign flag suffices. (d) There are two Expression→SQL emitters — `astexpr.go` (dialect-aware, parameterized) and `query_expression_sql.go` `ExpressionToSQL` (lossy display renderer used by `ConvertExpressionToMetricsSQL`) — with copy-pasted IN-flattening and nil→IS NULL logic; `ExpressionToSQL` doesn't handle `OperatorCast`, so Expression→SQL→ParseFilter round-tripping is not guaranteed lossless. (e) Identifier escaping in `resolvers/sql.go:281` is hardcoded to DuckDB (TODO) because the OLAP connector isn't known before template resolution — wrong quoting on ClickHouse/Druid edge cases.
- **Recommendation:** Give plain-column and computed dimensions real aliases (add an `Alias` field to `metricsview.Dimension`/`Measure`); pool parsers; collapse `add`/`sub`; make `ExpressionToSQL` a mode of the `astexpr` builder.

### deep-metrics-sql-14 [LOW] Template resolution runs before parsing — injection-shaped (contained by row security)
- **Category:** security  **Confidence:** medium  **Location:** `runtime/resolvers/metrics_sql.go:67` (`resolveTemplate` before `compiler.Parse` at `:116`)
- **Problem:** `newMetricsSQL` resolves Go-template interpolation on the raw SQL string before parsing, so a user attribute or API arg containing `' OR '1'='1` changes query *structure*, not just a value. This is mitigated — row-/field-level security policies are applied post-parse in the metricsview AST and the metrics-view boundary cannot be escaped — so the impact is filter-widening within already-permitted rows, not cross-tenant exfiltration. Still worth hardening.
- **Recommendation:** Prefer parsing first and injecting template values as bound parameters, or provide an escaping-aware template function and document that raw interpolation into string literals is unsafe.

## 4. Capability Gap: metrics SQL vs `metricsview.Query`

What the internal representation supports vs what SQL can express (`runtime/metricsview/query.go`):

| `metricsview.Query` capability | Expressible in metrics SQL? |
|---|---|
| Dimensions (plain), `Sort`, `Limit`/`Offset`, `Where`/`Having` | Yes |
| Dimension compute `time_floor` | Yes (`date_trunc`) — the only compute supported |
| Measures (plain) | Yes |
| Measure compute `count` / `count_distinct` | **No** — no aggregate function syntax |
| Measure compute `comparison_value` / `comparison_delta` / `comparison_ratio` | **No** |
| Measure compute `percent_of_total` | **No** |
| Measure compute `uri`, `comparison_time` | **No** |
| `PivotOn` | **No** |
| `Spine` (where-spine, time-spine / zero-fill) | **No** |
| `TimeRange` (structured: rilltime expression, iso duration/offset, round-to-grain, alt time dimension) | **No** — only baked-literal predicates via `time_range_start`/`end` |
| `ComparisonTimeRange` | **No** |
| `TimeZone` | **No** — only via resolver props, not SQL |
| `UseDisplayNames`, `Rows` | **No** |
| Subquery (IN-filter) | Yes (well-guarded) |

So the "SQL for your metrics layer" story currently covers roughly the toplist/aggregation case only. Everything distinctive about Rill dashboards — period comparisons, percent-of-total, pivots, zero-filled time series — is unreachable from metrics SQL, which limits its value for custom APIs, reports, and LLM authoring.

## 5. Architectural Recommendations

### 5.1 Make the compiler deny-by-default (foundational)
The recurring failure mode is TiDB's AST carrying more than the transform inspects. Invert the contract: after parsing, run a whitelist walker that errors on any clause/node not explicitly handled. Extract the subquery guard list (`filter_parser.go:218-250`) into a shared `validateSelectStmt(sel)` used by both the top-level and subquery paths, and check `SelectStmtOpts.Distinct`, `GroupBy`, `With`, `WindowSpecs`, `TableRefs.Right`, `field.AsName`, `field.WildCard`, set operators, etc. This one change converts deep-metrics-sql-02, -05, -08 and any future TiDB-AST surprise from silent-wrong-results into clear errors.

### 5.2 Defer time evaluation to the executor (key unlock)
Two designs, not mutually exclusive. **(a)** Lower time-dimension predicates (`timeDim > time_range_start('7D')`, or a `BETWEEN` on the time dim) into the structured `Query.TimeRange{Expression:"7D", TimeDimension:…}`; the executor's `RewriteQueryTimeRanges` then resolves at execution time with proper watermark/timezone/cache handling and both parse-time `GetTimestamps` round-trips disappear. **(b)** Add a fifth `Expression` node kind (`Function *FunctionCall{Name, Args}`) evaluated in the executor rewrite phase for `now`/`time_range_start`/`time_range_end`/`date_add`/`date_sub` — this is what the `parser.go:39` TODO asks for and also lets `date_add` apply to *columns* (dialect SQL) rather than only literals. Do (a) first (smaller, reuses existing machinery, covers the dominant case).

### 5.3 Unify the expression toolchain
Merge the two Expression→SQL emitters so IN-flattening / nil-handling / cast logic lives once and `ConvertExpressionToMetricsSQL` output round-trips through `ParseFilter`; give `ParseFilter` optional spec-based validation; pool parsers behind both entry points.

### 5.4 Expand the SQL surface toward Query parity (by value/effort)
1. **SELECT aliases** — honor `field.AsName`; add an `Alias` field to `metricsview.Dimension`/`Measure` (also fixes the unusable `date_trunc` aliases).
2. **Aggregate/computed-measure functions** — `count(*)`→`Count`, `count(distinct d)`→`CountDistinct`, `comparison_delta(m)` / `comparison_ratio(m)` / `percent_of_total(m)`→ the matching computes. Highest value for reports and LLM authoring.
3. **Comparison time range syntax** (e.g. a `COMPARE TO '-1P'` clause) → `ComparisonTimeRange`; without it the comparison computes are inert.
4. **Time zone** — `AT TIME ZONE` on `date_trunc` or a query-level clause → `Query.TimeZone`.
5. **Explicit GROUP BY / DISTINCT semantics**; **ORDER BY ordinals/expressions**.
6. **PIVOT → `PivotOn`; time spine / zero-fill → `Spine`** (lower priority; canvas/explore drive these structurally).
7. **Cross-view composition** — keep rejecting JOINs at the metricsview level, but longer term compile each referenced view to a compiled-SQL subquery and join in an outer DuckDB query via the `sql` resolver (a resolver-composition feature, not a parser feature; the per-view security model already fits).

### 5.5 Reconsider the parser dependency (strategic)
TiDB brings a MySQL dialect (the `||`-as-OR and locale-date quirks above), a very large dependency pinned to a **2023-11-24** pseudo-version (`go.mod:83`), and an AST far richer than the dialect needs — the direct cause of the silent-drop bugs. Do §5.1 now regardless. If §5.4 items 2–4 are pursued (new clauses/functions MySQL grammar lacks), that is the natural moment to move to a small owned grammar (as `rilltime` already does with participle): exact fit for the intended dialect, owned positioned error messages, and no oversized dependency.

### 5.6 Error quality & testing (cheap, compounding)
Prefix every error `metrics sql:`, attach `OriginTextPosition()` line/column, and add "did you mean" suggestions from the known dims/measures — this directly speeds the canvas custom-chart AI agent's convergence and improves reconcile-error readability. Add a golden corpus of (SQL → Query JSON | error) covering every clause and rejection (there are currently zero error-path tests in `parser_test.go`), fuzz `Parse`/`ParseFilter`, and add cross-dialect execution tests (ClickHouse, Druid) for `IS TRUE/FALSE`, casts, ILIKE, and time literals.

## 6. Documentation Drift
`docs/docs/developers/build/metrics-view/metrics-sql.md` claims `NOT BETWEEN` (broken, -03), `IS [NOT] TRUE/FALSE` (broken, -01), and `date_trunc(...) AS alias` (alias only works for `date_trunc`, plain aliases dropped, -05) are supported; implies numeric filters work (decimals broken -04, negatives error -08); and says "no joins" though JOINs silently drop rather than error (-02). Docs correctly warn the syntax is evolving without backward-compatibility guarantees — the window to fix dialect-level mistakes (LIKE→ILIKE, `||`) before the surface calcifies.

## 7. Priority of Fixes

Severity × confidence, ship-blocking correctness first.

| Priority | Findings | Item | Effort |
|---|---|---|---|
| ✅ **Fixed** | 01, 03, 04 | `IS FALSE`, `NOT BETWEEN`, and decimal literals — #9712, #9719, #9720 | — |
| ✅ **Fixed** | 02, 05 | JOIN rejection and top-level clause validation — #9721, #9716 | — |
| **P0 — pending** | 08 | Add an operator allowlist and unary-negative literal support; unsupported operators still pass into invalid IR | S |
| **P1 — dialect correctness** | 06, 11, 12 | Reject/support `\|\|`; bound-check LIMIT; ISO-8601 dates; ILIKE decision | S each |
| **P1 — architecture** | 07 | Lower time predicates into `Query.TimeRange`; remove parse-time `GetTimestamps` round-trips (§5.2a) | M |
| **P1 — safety/UX** | 09, 10, 14 | Fix canvas ref-discovery swallow; parse-time name validation + positioned errors; template-injection hardening | S–M |
| **P2 — surface expansion** | §4 gap | Aliases, computed-measure functions, comparison range, time zone (§5.4) | M–L |
| **P2 — cleanup / strategic** | 13, §5.3, §5.5 | Parser pooling, unify Expression→SQL emitters, deferred `Expression` node, own grammar | M–L |

## Appendix A: Manual Reproduction

A self-contained scratch test reproduces every verified finding. Drop this in `runtime/metricsview/metricssql/audit_scratch_test.go`, run `go test ./runtime/metricsview/metricssql/ -run TestAuditRepro -v`, then delete it.

```go
package metricssql

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

func fakeMV(name string) *runtimev1.Resource {
	return &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Kind: "rill.runtime.v1.MetricsView", Name: name}},
		Resource: &runtimev1.Resource_MetricsView{MetricsView: &runtimev1.MetricsView{
			State: &runtimev1.MetricsViewState{ValidSpec: &runtimev1.MetricsViewSpec{
				TimeDimension: "timestamp",
				Dimensions:    []*runtimev1.MetricsViewSpec_Dimension{{Name: "pub"}, {Name: "dom"}},
				Measures:      []*runtimev1.MetricsViewSpec_Measure{{Name: "measure_0"}},
			}},
		}},
	}
}

func TestAuditRepro(t *testing.T) {
	c := New(&CompilerOptions{GetMetricsView: func(ctx context.Context, n string) (*runtimev1.Resource, error) { return fakeMV(n), nil }})
	// Compile-path findings (02, 05, 08, 11)
	for _, sql := range []string{
		"select pub from mv1 join mv2 on mv1.pub = mv2.pub",  // 02: JOIN silently dropped
		"select distinct pub from mv",                        // 05: DISTINCT dropped
		"select pub as publisher from mv",                    // 05: alias dropped
		"with x as (select pub from mv) select pub from mv",  // 05: CTE dropped
		"select pub from mv where measure_0 + 1 > 5",         // 08: op "plus" passes parse
		"select pub from mv limit 18446744073709551615",      // 11: wraps to -1
	} {
		q, err := c.Parse(context.Background(), sql)
		if err != nil { fmt.Printf("PARSE %q\n  ERROR: %v\n", sql, err); continue }
		b, _ := json.Marshal(q); fmt.Printf("PARSE %q\n  OK: %s\n", sql, b)
	}
	// Filter-path findings (01, 03, 04, 06, 08)
	for _, f := range []string{
		"flag IS FALSE",         // 01: identical to IS TRUE
		"flag IS TRUE",          // 01: compare vs string 'TRUE'
		"x NOT BETWEEN 1 AND 10",// 03: identical to BETWEEN
		"m > 10.5",              // 04: decimal -> {}
		"a || b = 'ab'",         // 06: || -> OR
		"m > -5",                // 08: unary minus errors
	} {
		expr, err := ParseFilter(f)
		if err != nil { fmt.Printf("FILTER %q\n  ERROR: %v\n", f, err); continue }
		b, _ := json.Marshal(expr); fmt.Printf("FILTER %q\n  OK: %s\n", f, b)
	}
}
```

Observed output on this branch (abbreviated):

```
PARSE "…join mv2 on…"        OK: {"metrics_view":"mv1","dimensions":[{"name":"pub"}], …}   # 02 JOIN gone
PARSE "select distinct pub…" OK: dimensions=[pub]                                          # 05 DISTINCT gone
PARSE "…pub as publisher…"   OK: dimensions=[{"name":"pub"}]                               # 05 alias gone
PARSE "with x as…"           OK: (CTE gone)                                                # 05
PARSE "…measure_0 + 1 > 5"   OK: {"op":"plus", …}                                          # 08
PARSE "…limit 1844674407…"   OK: "limit":-1                                                # 11
FILTER "flag IS FALSE"       OK: {"op":"eq",[{name:flag},{val:"TRUE"}]}                     # 01 == IS TRUE
FILTER "flag IS TRUE"        OK: {"op":"eq",[{name:flag},{val:"TRUE"}]}                     # 01
FILTER "x NOT BETWEEN 1 AND 10" OK: {and,[{gte,x,1},{lte,x,10}]}                            # 03 == BETWEEN
FILTER "m > 10.5"            OK: {"op":"gt",[{name:m},{val:{}}]}                            # 04 decimal lost
FILTER "a || b = 'ab'"       OK: {"op":"or", …}                                            # 06 || as OR
FILTER "m > -5"              ERROR: unsupported type *ast.UnaryOperationExpr                # 08
```

An end-to-end alternative once a project is running: `rill query --local --resolver metrics_sql --properties sql="select pub from ad_bids_metrics where impressions is false"` against the `ad_bids` fixture (`runtime/testruntime/testdata/ad_bids`) reproduces the boolean-filter inversion at the API layer.

## Appendix B: Key File Index

| File | Role |
|---|---|
| `runtime/metricsview/metricssql/parser.go` | Compiler; FROM/SELECT/LIMIT/ORDER/HAVING transforms; operator map |
| `runtime/metricsview/metricssql/filter_parser.go` | WHERE-expression transforms; subquery guard; value/func handling; `ParseFilter` |
| `runtime/metricsview/metricssql/time_range_parser.go` | `time_range_start/end` → parse-time rilltime evaluation |
| `runtime/resolvers/metrics_sql.go` | `metrics_sql` resolver: templating, security callbacks, additional where/time-range merge |
| `runtime/resolvers/builtin_metrics_sql.go` | Admin-only builtin `/metrics-sql` API |
| `runtime/metricsview/query.go` | Target IR (`Query`, `Expression`, compute variants) |
| `runtime/metricsview/astexpr.go` | Expression → dialect SQL (parameterized, security-aware) |
| `runtime/metricsview/query_expression_sql.go` | Expression → display SQL (duplicate, lossy) |
| `runtime/metricsview/executor/executor_rewrite_time.go` | Execution-time rilltime resolution metrics SQL cannot reach today |
| `runtime/queries/metricsview_aggregation.go:576` | Per-request `ParseFilter` for `whereSQL`/`havingSQL` |
| `runtime/canvases.go:140` | Silent parse-error swallow in canvas ref discovery |
| `runtime/resolvers/sql.go:281` | DuckDB-hardcoded identifier escaping TODO |
| `docs/docs/developers/build/metrics-view/metrics-sql.md` | User-facing dialect reference (drift in §6) |
