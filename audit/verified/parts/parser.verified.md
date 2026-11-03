# Verified audit: `parser.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **10** numbered findings. Every source ID is classified once below.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| parser-03 | MEDIUM | P1 | Multi-file `.env` reparse can erase a real change | `runtime/parser/parser.go:433-480,572-575`; line 473 assigns instead of accumulating | Use `modifiedDotEnv = modifiedDotEnv || changed`; test changed then unchanged env paths in one batch. |
| parser-04 | LOW | P2 | Empty display name panics `InitEmpty` | `runtime/parser/init.go:18-21` indexes `displayName[0]` | Validate non-empty input or supply a default before indexing; test the REST/RPC entry point. |
| parser-05 | LOW | P2 | Environment overrides bypass unknown-field validation | Base decoder enables known fields at `runtime/parser/parse_node.go:346-356`; override uses `yaml.Node.Decode` at `:370-375` | Decode overrides through a known-fields-aware overlay or validate override keys against the same schema. |
| parser-06 | LOW | P2 | Comparison dimension accepts a measure | `runtime/parser/parse_metrics_view.go:665-668` uses `!ok && v != nameIsDimension` | Change to `!ok || v != nameIsDimension`; add measure/dimension cases. |
| parser-08 | LOW | P3 | `default_measures` error names wrong property | `runtime/parser/parse_metrics_view.go:624-628` says `default_dimensions` | Correct the message and pin it in a parser test. |
| parser-10 | LOW | P2 | Bad SQL position becomes line 4294967295 | `runtime/parser/parse_model.go:340-348,410-425` casts `-1` to `uint32` | Omit/clamp location on invalid positions; test EOF and out-of-range offsets. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| parser-01 | FIXED | `ComponentVariableYAML.Proto` now returns an error from `pbutil.ToValue` at `runtime/parser/parse_component.go:29-41`; parser regression tests cover unsupported values. | PR #9702, `22b82781c` |
| parser-02 | FIXED | Annotation validation now checks `annotation.Table`/`annotation.Model` at `runtime/parser/parse_metrics_view.go:735-756`, with parent/annotation matrix tests. | PR #9724, `53a15b6f7` |

## Partial / Needs-test

| ID | Original severity | Priority | Why partial | Current evidence / next test |
|---|---:|---:|---|---|
| parser-07 | LOW | P3 | New inline explores now inherit `rill.yaml` theme through shared parsing, but the legacy auto-emitted explore still assigns only `tmp.DefaultTheme`. | Fixed inline path: `runtime/parser/parse_explore.go:227-238,330-338` (PR #9680); residual legacy path: `runtime/parser/parse_metrics_view.go:946-970`. Apply the fallback there or retire legacy emission. |
| parser-09 | LOW | P3 | Panic-after-insert pattern remains but preceding dry runs and single-threaded insertion make it currently unreachable. | `runtime/parser/parse_metrics_view.go:850-861`; `runtime/parser/parse_canvas.go:244-251`; refactor insertion transactionally and fault-inject the second insert. |

Counts: **6 pending, 2 fixed/remove, 2 partial/needs-test, 0 invalid = 10 total**.

## Machine-readable manifest

```text
FIXED|parser.md|parser-01|HIGH|-|-|Component variable value could panic project parse|runtime/parser/parse_component.go:29-41; PR#9702
FIXED|parser.md|parser-02|MEDIUM|-|-|Annotation validation checked parent fields|runtime/parser/parse_metrics_view.go:735-756; PR#9724
PENDING|parser.md|parser-03|MEDIUM|P1|-|Multi-env reparse can erase change flag|runtime/parser/parser.go:433-480,572-575
PENDING|parser.md|parser-04|LOW|P2|-|InitEmpty panics on empty display name|runtime/parser/init.go:18-21
PENDING|parser.md|parser-05|LOW|P2|-|Environment override bypasses known fields|runtime/parser/parse_node.go:346-356,370-375
PENDING|parser.md|parser-06|LOW|P2|-|Comparison dimension accepts measure|runtime/parser/parse_metrics_view.go:665-668
PARTIAL|parser.md|parser-07|LOW|P3|-|Metrics-view explore misses project theme fallback|runtime/parser/parse_explore.go:227-238; runtime/parser/parse_metrics_view.go:946-970; PR#9680
PENDING|parser.md|parser-08|LOW|P3|-|default_measures error names default_dimensions|runtime/parser/parse_metrics_view.go:624-628
PARTIAL|parser.md|parser-09|LOW|P3|-|Post-insert failures panic|runtime/parser/parse_metrics_view.go:850-861; runtime/parser/parse_canvas.go:244-251
PENDING|parser.md|parser-10|LOW|P2|-|Invalid SQL position wraps to huge line|runtime/parser/parse_model.go:340-348,410-425
```
