# Runtime Parser — Code Review
**Scope:** `runtime/parser/*.go` (parser.go, template.go, parse_node.go, connectors.go, parse_rillyaml.go, parse_dotenv.go, and all per-resource-type parse funcs: model, source, metrics_view, explore, canvas, component, theme, alert, report, api, connector, migration; plus partials: data, schedule, field_selector, security_policy; schema.go, init.go, util.go)  |  **Health:** fair

## Overview
The parser turns a project's YAML/SQL files into an in-memory DAG of `Resource`s and supports incremental reparsing driven by a file watcher. The code is generally well-structured, heavily commented, and backed by a large test suite (`parser_test.go` alone is ~3k lines). Error collection is careful: most parse funcs validate fully *before* calling `insertResource` (there are explicit "must not return an error after this point" comments), and errors are attributed to the right file path via `pathError`.

The biggest risks are (1) a user-reachable `panic` when a component/canvas variable's `value:` cannot be converted to protobuf (e.g. a bare date), which — because the package contains **no** `recover()` anywhere — fails the entire project parse rather than one file; (2) a copy/paste validation bug in metrics-view annotations that validates the parent's data source instead of the annotation's; and (3) an incremental-reparse bug where a real `.env` change can be dropped when several `.env` files are reparsed together (traced through to the reconciler, which then skips `updateConfig`). There are also several lower-severity correctness/consistency issues (an `&&`/`||` inversion, KnownFields not enforced on environment-override blocks, a wrong error string, an empty-string index panic in `InitEmpty`, and a garbage line number on out-of-range SQL error positions). None are cross-tenant security holes — templating/SQL is authored and executed within a single project's trust boundary — but the panics are effective self-DoS on malformed input.

## Findings

### parser-01 [HIGH] Component/canvas variable `value:` can panic the whole project parse

**Status (2026-07-25):** ✅ Fixed and merged in [#9702](https://github.com/rilldata/rill/pull/9702).

- **Category:** bug  **Confidence:** high  **Location:** runtime/parser/parse_component.go:34 (call sites: parse_component.go:145 & :155, parse_canvas.go:151, and inline canvas components via `parseComponentYAML` at parse_canvas.go:447)
- **Problem:** `ComponentVariableYAML.Proto()` does `val, err := structpb.NewValue(y.Value)` and then `panic(...)` on error instead of returning it. `Value` is `any` decoded straight from YAML (yaml.v3), so it can hold types `structpb.NewValue` rejects — most notably a `time.Time` (yaml.v3 resolves bare dates like `2024-01-01` to `time.Time` per the YAML 1.1 timestamp rule). Every other parse path in this package converts the `structpb` error into a normal parse error; this one panics. Verified: there is **no** `recover()` anywhere in `runtime/parser`, and `parseNode` is not wrapped in one, so the panic aborts the whole `Parse`/`Reparse`.
- **Failure scenario:** A component (or canvas) with `input:\n  - name: start\n    value: 2024-01-01` — a very plausible default for a date input — decodes `value` to `time.Time`; `structpb.NewValue` returns "invalid type: time.Time"; the code panics. The panic is caught only by the controller's reconcile recover (`runtime/controller.go:~1319`) and turned into a hard error on the *entire* ProjectParser, so no resources in the project load. A single bad `value:` takes down the whole project instead of producing a localized parse error on one file.
- **Recommendation:** Return the error instead of panicking: `if err != nil { return nil, fmt.Errorf("invalid default value: %w", err) }` (the two call sites already handle a returned error). Consider pre-normalizing `time.Time` (and other YAML scalar types) to strings before `structpb.NewValue`.

**Verification:** CONFIRMED — `runtime/parser/parse_component.go:32-35` calls `structpb.NewValue(y.Value)` and executes `panic(fmt.Errorf("invalid default value: %w", err))` on error; both call sites (parse_component.go:145,155) and the canvas path (parse_canvas.go:151, parse_canvas.go:447 via `parseComponentYAML`) already handle a returned error, so the panic is gratuitous. `grep -rn "recover()" runtime/parser/` returns nothing, so the panic escapes the whole parse.

**Manual Reproduction Steps (for QA):**
1. Start Rill Developer (`rill start`) on any working project.
2. Create `components/panic-test.yaml` with:
   ```yaml
   type: component
   markdown:
     content: "hello"
   input:
     - name: start
       value: 2024-01-01
   ```
   (yaml.v3 decodes the bare date `2024-01-01` into a `time.Time`, which `structpb.NewValue` rejects.)
3. Expected: a parse error localized to `components/panic-test.yaml` (e.g. "invalid input variable at index 0: invalid default value ...") while all other resources keep working.
4. Actual: the whole project parse fails — every dashboard/model in the project becomes unavailable and the runtime logs show a recovered panic with "invalid type: time.Time" originating from the ProjectParser reconcile. Deleting the file restores the project.

### parser-02 [MEDIUM] Metrics-view annotation validation checks the parent's `model`/`table`, not the annotation's

**Status (2026-07-25):** ✅ Fixed and merged in [#9724](https://github.com/rilldata/rill/pull/9724).

- **Category:** correctness  **Confidence:** high  **Location:** runtime/parser/parse_metrics_view.go:753-758
- **Problem:** Inside `for _, annotation := range tmp.Annotations`, the two "must set exactly one of model/table" checks reference `tmp.Table`/`tmp.Model` (the metrics view's top-level fields) even though the error messages and intent are about the annotation. The annotation's own `annotation.Model`/`annotation.Table` (distinct struct fields, parse_metrics_view.go:79-87) are never validated. Because the metrics view already rejects setting both `model` and `table` earlier (line 294-295), the first check here is dead code, and the second check (`tmp.Table == "" && tmp.Model == ""`) only ever fires for `parent`-based derived views.
- **Failure scenario:** Two-directional: (a) A metrics view with `model: sales` and an annotation that sets *neither* `model` nor `table` passes validation (because `tmp.Model != ""`), then falls through to `annotation.Name = annotation.Table` = `""`, emitting a spec annotation with empty Name/Model/Table (parse_metrics_view.go:933-942) — a silent misconfiguration. (b) A `parent`-based derived view with any annotation is always rejected with "must set a value for either the model or table field for annotation", even when the annotation itself sets a valid model/table.
- **Recommendation:** Change the two checks to use `annotation.Table`/`annotation.Model`. Also decide explicitly whether annotations are allowed on `parent`-based derived views.

**Verification:** CONFIRMED — `runtime/parser/parse_metrics_view.go:753-758` checks `tmp.Table`/`tmp.Model` (the metrics view's own fields) inside the `for _, annotation := range tmp.Annotations` loop; the annotation struct's own `Model`/`Table` fields (parse_metrics_view.go:79-87) are never validated. The parent-level both-set case is already rejected at :294-295, so the first check is dead; the second only fires when both parent fields are empty, i.e. `parent:`-based views (:297-298 allows that combination).

**Manual Reproduction Steps (for QA):**
1. In a running Rill Developer project with a model `sales`, create a derived metrics view `metrics/derived.yaml`:
   ```yaml
   version: 1
   type: metrics_view
   parent: some_existing_metrics_view
   annotations:
     - name: releases
       model: sales
   ```
2. Expected: no error about the annotation's model/table (it sets `model: sales`).
3. Actual: parse error `must set a value for either the "model" field or the "table" field for annotation`, because the check reads the parent-level `tmp.Model`/`tmp.Table` (both empty for `parent`-based views).
4. Converse case: on a normal metrics view (`model: sales`), add `annotations: [{name: foo}]` with neither `model` nor `table`. Expected: a validation error; Actual: it parses cleanly and the spec (visible at `http://localhost:9009/v1/instances/default/resource?name.kind=rill.runtime.v1.MetricsView&name.name=<mv>`) contains an annotation with empty `model` and `table`.

### parser-03 [MEDIUM] Incremental `.env` reparse can drop a real change when multiple `.env` files are reparsed together
- **Category:** correctness  **Confidence:** high  **Location:** runtime/parser/parser.go:433 (decl) & :468-480 (assignment on :472)
- **Problem:** `modifiedDotEnv` is a single bool shared across the whole loop (declared at :433), but it is *assigned* (`modifiedDotEnv = !maps.Equal(...)`) rather than OR-accumulated for each `.env` path. If a batch reparses two `.env` files where a later-processed one is unchanged, its `false` overwrites the earlier `true`. `Diff.ModifiedDotEnv` then reports `false`, and the reconciler (`runtime/reconcilers/project_parser.go:344`, verified: `updateConfig := diff == nil || diff.ModifiedDotEnv || diff.Reloaded`) skips `reconcileProjectConfig`, so the changed environment variables are not applied to the instance config and the controller is not restarted until the next full reload.
- **Failure scenario:** User edits `.env` (real change) and `.prod.env` (saved but content-identical), and the watcher delivers both paths in one `Reparse`. `.prod.env` processed last → `modifiedDotEnv` ends `false` → the `.env` change is silently not applied.
- **Recommendation:** Accumulate: `modifiedDotEnv = modifiedDotEnv || !maps.Equal(p.DotEnv[path], oldDotEnv)` (and on parse error, `modifiedDotEnv = true`). Keep the per-path `continue` based on a local `changed` variable rather than the shared flag.

**Verification:** CONFIRMED — `runtime/parser/parser.go:433` declares the single shared `modifiedDotEnv := false`, and `:472` plainly assigns `modifiedDotEnv = !maps.Equal(p.DotEnv[path], oldDotEnv)` inside the per-path loop, so a later unchanged `.env` path overwrites an earlier `true`. `:573` propagates it as `Diff.ModifiedDotEnv`, and `runtime/reconcilers/project_parser.go:344` (`updateConfig := diff == nil || diff.ModifiedDotEnv || diff.Reloaded`) confirms a `false` skips applying the config.

**Manual Reproduction Steps (for QA):**
Timing-dependent: both `.env` paths must arrive in the same watcher batch, with the unchanged one iterated after the changed one. Best-effort repro:
1. Start Rill Developer on a project that has both a root `.env` (e.g. `MY_VAR=a`) and a nested `sub/.env` whose content you will not change, plus a resource that consumes the variable (e.g. a model using `{{ .env.MY_VAR }}`).
2. In one shell command, write both files at once so the watcher delivers them in one `Reparse` batch: `printf 'MY_VAR=b\n' > .env && touch sub/.env` (or re-save `sub/.env` with identical content).
3. Expected: the runtime restarts the controller and the model re-resolves with `MY_VAR=b`.
4. Actual (when the batch lands together and `sub/.env` is processed last): no controller restart; the model still sees `MY_VAR=a` until another edit or a full reload.
5. If the race does not land, confirm in code: `runtime/parser/parser.go:433,472` (plain assignment, not OR-accumulation) plus `runtime/reconcilers/project_parser.go:344`.

### parser-04 [LOW] `InitEmpty` panics on an empty display name
- **Category:** bug  **Confidence:** high  **Location:** runtime/parser/init.go:20
- **Problem:** `if !isAlphabetic(displayName[0])` indexes `displayName[0]` without checking for the empty string, so an empty `displayName` causes an index-out-of-range panic. The runtime `UnpackEmpty` handler (`runtime/server/examples.go:140`) passes `req.DisplayName` straight through, and the generated proto validation (`UnpackEmptyRequest.validate`, `proto/gen/rill/runtime/v1/api.pb.validate.go:5202`) has "no validation rules for DisplayName", so an empty `display_name` reaches this line.
- **Reachability correction:** The CLI caller (`cli/cmd/initialize/init.go:186`) is **not** a viable trigger — `projectName := filepath.Base(projectPath)` (init.go:97) never returns an empty string. The only reachable path is the `UnpackEmpty` gRPC/REST handler (gated by `EditRepo` auth) with an explicit empty `display_name`. The runtime gRPC chain has **no** `grpc_recovery` interceptor (`runtime/server/server.go:158-166`); the panic is instead recovered per-request by `net/http` (requests are served through the vanguard transcoder over the HTTP mux), so it degrades one request rather than crashing the process.
- **Failure scenario:** A client calls `UnpackEmpty` with `display_name: ""` → panic → the single request fails (connection reset / error), leaving no project created and a confusing stack in the logs.
- **Recommendation:** Guard for empty input before indexing, e.g. `if displayName == "" { displayName = "Untitled" }` (or return a clear validation error).

**Verification:** CONFIRMED — `runtime/parser/init.go:20` (`if !isAlphabetic(displayName[0])`) indexes without an empty-string guard. `runtime/server/examples.go:140` passes `req.DisplayName` straight through, and the REST route is `POST /v1/instances/{instance_id}/files/unpack-empty` (proto/rill/runtime/v1/api.proto:139) with no validation rule on `display_name`. The CLI path (`cli/cmd/initialize/init.go:186`) is not a trigger since `filepath.Base` never returns "".

**Manual Reproduction Steps (for QA):**
1. Start Rill Developer (`rill start`) in an empty directory (no `rill.yaml`), so `UnpackEmpty` is permitted.
2. Call the runtime REST endpoint with an empty display name: `curl -X POST http://localhost:9009/v1/instances/default/files/unpack-empty -H 'Content-Type: application/json' -d '{"display_name":"","olap":"duckdb"}'`.
3. Expected: either a project created with a default name or a clean validation error.
4. Actual: the request fails abnormally (empty reply/connection reset or 500) and the runtime log contains an `index out of range [0] with length 0` panic stack from `parser.InitEmpty`; no project files are created. The process itself stays up (per-request recovery by `net/http`).

### parser-05 [LOW] Environment-override blocks bypass unknown-field (KnownFields) validation
- **Category:** correctness  **Confidence:** high  **Location:** runtime/parser/parse_node.go:367-372 (vs. the KnownFields base decode at :345-349)
- **Problem:** For resource types decoded with `knownFields=true` (metrics view, explore, canvas, alert, report, component), the base YAML is decoded from `node.YAMLRaw` with `dec.KnownFields(true)` so a typo'd top-level key is a hard error. But the environment-specific override is applied via `node.YAMLOverride.Decode(dst)`, and `YAMLOverride` is a `*yaml.Node` (parse_node.go:23) whose `Decode` does not support KnownFields. Unknown/typo'd keys inside a `dev:`, `prod:`, or `environment_overrides:` block are therefore silently ignored.
- **Failure scenario:** A user writes `prod:\n  smalles_time_grain: hour` (typo) expecting it to take effect in production; the base-level equivalent would error, but in the override block it is silently dropped, so the override never applies and no error is shown.
- **Recommendation:** Re-encode the override node to raw YAML and decode it with `KnownFields(true)` (as done for the base), or otherwise validate override keys against the destination struct.

**Verification:** CONFIRMED — `runtime/parser/parse_node.go:345-349` decodes the base YAML with `dec.KnownFields(true)`, but the environment override at :367-372 uses `node.YAMLOverride.Decode(dst)` on a `*yaml.Node` (declared parse_node.go:23), and `yaml.Node.Decode` has no KnownFields support, so unknown keys in override blocks are silently dropped.

**Manual Reproduction Steps (for QA):**
1. In a running Rill Developer project (local environment is `dev`), open a metrics view YAML and first add a typo'd key at the top level, e.g. `smalles_time_grain: hour`. Observe the expected hard parse error ("field smalles_time_grain not found...").
2. Remove it and instead put the same typo inside a dev override block:
   ```yaml
   dev:
     smalles_time_grain: hour
   ```
3. Expected: the same unknown-field parse error.
4. Actual: no error is shown, the file parses cleanly, and the override silently does nothing (the metrics view's smallest time grain is unchanged).

### parser-06 [LOW] `default_comparison.dimension` accepts a measure name (`&&` should be `||`)
- **Category:** correctness  **Confidence:** high  **Location:** runtime/parser/parse_metrics_view.go:684
- **Problem:** `if v, ok := names[strings.ToLower(tmp.DefaultComparison.Dimension)]; !ok && v != nameIsDimension` uses `&&` where every analogous check in the file uses `||` (e.g. :467, :644, :813). With `nameIsMeasure=1`, `nameIsDimension=2`, map-miss=0: a non-existent name is still rejected (`!ok`=true and `v(0)!=2`=true), but a name that exists as a *measure* (`ok`=true → `!ok`=false) short-circuits the `&&` to false, so the measure is accepted as the default comparison dimension.
- **Failure scenario:** `default_comparison:\n  mode: dimension\n  dimension: total_revenue` where `total_revenue` is a measure parses without error and emits an invalid preset.
- **Recommendation:** Change `&&` to `||` to match the other name-kind validations. This is a deprecated field, but the fix is trivial and prevents an invalid spec.

**Verification:** CONFIRMED — `runtime/parser/parse_metrics_view.go:684` reads `if v, ok := names[strings.ToLower(tmp.DefaultComparison.Dimension)]; !ok && v != nameIsDimension`, whereas the analogous checks use `||` (:467, :644, :637). Constants at :276-277 (`nameIsMeasure=1`, `nameIsDimension=2`) confirm a measure name yields `ok=true`, short-circuiting the `&&` to false and skipping the error.

**Manual Reproduction Steps (for QA):**
1. In a Rill Developer project, create a metrics view (without a `version:` field, so it accepts the deprecated key) that has a measure named `total_revenue`, and add:
   ```yaml
   default_comparison:
     mode: dimension
     dimension: total_revenue
   ```
2. Expected: a parse error like `default comparison dimension "total_revenue" doesn't exist` (it is a measure, not a dimension).
3. Actual: the file parses without error; the emitted explore preset carries the measure name as its comparison dimension (visible in the resource API or as a broken comparison selection in the explore dashboard). Sanity check: using a truly non-existent name (e.g. `dimension: nope`) still errors, confirming only the measure case slips through.

### parser-07 [LOW] Metrics-view-emitted explore ignores the `rill.yaml` project theme fallback
- **Category:** product-improvement  **Confidence:** high  **Location:** runtime/parser/parse_metrics_view.go:988 (auto-emitted explore) and :1134/:1167 (inline explore) vs. runtime/parser/parse_explore.go:172-174
- **Problem:** A standalone explore falls back to `p.RillYAML.Theme` when no local/default theme is set (parse_explore.go:172-174). The default explore auto-emitted from a metrics view sets `Theme = tmp.DefaultTheme` directly (:988), and the inline explore uses `parseThemeRef` (:1134) without the `p.RillYAML.Theme` fallback (:1167) — so identical dashboards are themed inconsistently depending on whether the explore is auto-generated or hand-written.
- **Failure scenario:** A project sets `theme:` in `rill.yaml`; hand-written explores pick it up, but the auto-emitted/inline explores for metrics views render with the default theme.
- **Recommendation:** Apply the same `p.RillYAML.Theme` fallback when emitting the metrics-view explore (and the inline explore), or document the difference intentionally.

**Verification:** CONFIRMED — `runtime/parser/parse_explore.go:171-174` applies the `p.RillYAML.Theme` fallback for standalone explores; the auto-emitted explore sets `e.ExploreSpec.Theme = tmp.DefaultTheme` directly (parse_metrics_view.go:988) and the inline explore sets `r.ExploreSpec.Theme = themeName` from `parseThemeRef` (:1134, :1167), neither consulting `p.RillYAML.Theme`.

**Manual Reproduction Steps (for QA):**
1. In a Rill Developer project, create a theme (e.g. `themes/mytheme.yaml` with distinctive primary colors) and set `theme: mytheme` in `rill.yaml`.
2. Create a hand-written explore (`explores/manual.yaml` referencing a metrics view, with no `theme:` key) and open it: it renders with `mytheme` colors (the rill.yaml fallback).
3. Create a legacy metrics view without a `version:` field and without `default_theme:` (this auto-emits an explore), and/or a metrics view with an inline `explore:` block without a theme.
4. Expected: the auto-emitted/inline explore also renders with `mytheme`.
5. Actual: it renders with the default (unthemed) look; the explore spec's `theme` field is empty in the resource API.

### parser-08 [LOW] Wrong property name in the `default_measures` validation error
- **Category:** readability  **Confidence:** high  **Location:** runtime/parser/parse_metrics_view.go:645
- **Problem:** The loop validating `tmp.DefaultMeasures` (line 643) returns `measure %q referenced in "default_dimensions" not found` — it names `default_dimensions` for a `default_measures` failure (contrast the correct `default_dimensions` loop at :466-468), which will mislead users debugging the error.
- **Recommendation:** Change the message to reference `"default_measures"`.

**Verification:** CONFIRMED — `runtime/parser/parse_metrics_view.go:643-646`: the `tmp.DefaultMeasures` loop returns `measure %q referenced in "default_dimensions" not found`; the correct dimension-loop message is at :466-468.

**Manual Reproduction Steps (for QA):**
1. In a Rill Developer metrics view, add `default_measures: [does_not_exist]` (with no `default_dimensions` key at all).
2. Expected: an error naming `default_measures`.
3. Actual: the parse error reads `measure "does_not_exist" referenced in "default_dimensions" not found`, pointing the user at a key they never set.

### parser-09 [LOW] Fragile "already inserted, so panic on failure" pattern in canvas/metrics-view insertion
- **Category:** tech-debt  **Confidence:** medium  **Location:** runtime/parser/parse_metrics_view.go:876-879, runtime/parser/parse_canvas.go:247-251
- **Problem:** Because inline explores/components are inserted *before* their parent resource, these paths `panic` if the subsequent parent/child `insertResource` fails, relying on a preceding `insertDryRun` (parse_metrics_view.go:1143, parse_canvas.go:456/459) to prove the panic is unreachable. It is currently safe (kinds differ, single-threaded, dry-run precedes), but it is a maintenance landmine: any future validation added between the dry-run and the real insert would turn a normal error into a process-affecting panic (the package has no `recover()`).
- **Failure scenario:** A later change adds a fallible step between `parseAndInsertInlineExplore` and the metrics-view `insertResource`; a malformed file now panics instead of producing a parse error.
- **Recommendation:** Restructure so all fallible work completes before any `insertResource`, or buffer the child defs and insert parent-then-children only after all validation, removing the need for the panics.

**Verification:** PLAUSIBLE — the panics exist as described (`runtime/parser/parse_metrics_view.go:874-881` panics if the metrics-view insert fails after the inline explore was created; `runtime/parser/parse_canvas.go:244-251` panics if an inline component insert fails after the canvas insert), guarded only by preceding `insertDryRun` calls (parse_metrics_view.go:1143, parse_canvas.go:456-462). Currently unreachable (dry-run precedes, kinds differ, single-threaded), so this is latent tech-debt rather than an active bug.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — no project file content currently makes the dry-run pass and the real insert fail. Confirm in code: `runtime/parser/parse_metrics_view.go:874-881` and `runtime/parser/parse_canvas.go:244-251` (the `panic(...)` calls), noting the package contains no `recover()` and that any future fallible step added between `insertDryRun` and `insertResource` would convert a parse error into a whole-parse panic.

### parser-10 [LOW] Out-of-range SQL error position yields a garbage line number (`uint32(-1)`)
- **Category:** correctness  **Confidence:** medium  **Location:** runtime/parser/parse_model.go:347 (helper at :412-415)
- **Problem:** `findLineNumber(text, pos)` returns `-1` when `pos < 0 || pos >= len(text)`. The single caller wraps it as `Line: uint32(findLineNumber(node.SQL, posError.Position))` — so an out-of-range position becomes `uint32(-1)` = 4294967295 in the error's `CharLocation.Line`. DuckDB commonly reports a position at end-of-input for "unexpected EOF"-style syntax errors, where `posError.Position == len(node.SQL)`, hitting the `pos >= len(text)` branch.
- **Failure scenario:** A model with a trailing/incomplete SQL statement produces a parse error whose location points at line 4294967295, which the editor cannot map to a real line.
- **Recommendation:** Clamp/normalize the out-of-range case to a valid line (e.g. the last line, or `1`) rather than `-1`, or have the caller treat a negative return as "no location".

**Verification:** CONFIRMED — `runtime/parser/parse_model.go:412-415` returns `-1` for `pos < 0 || pos >= len(text)`, and the sole caller at :347 wraps it as `Line: uint32(findLineNumber(node.SQL, posError.Position))`, so any out-of-range position becomes 4294967295 in the error's `CharLocation.Line`. Reaching it requires DuckDB to report a position at/after end-of-input (typical for unexpected-EOF syntax errors), which makes the garbage line number user-visible.

**Manual Reproduction Steps (for QA):**
1. In a Rill Developer project using DuckDB, create `models/broken.sql` containing an incomplete statement that errors at end-of-input, e.g. `SELECT * FROM` (no trailing newline).
2. Open the file in the Rill Developer editor and observe the parse error for the model.
3. Expected: the syntax error points at the last real line of the file (or carries no location).
4. Actual (when DuckDB reports the error position at end-of-input): the error's location line is 4294967295 (`uint32(-1)`), so the editor cannot highlight the offending line. If a given DuckDB version reports an in-range position for this input, try variants that end mid-token (e.g. `SELECT 1 WHERE`), or confirm in code at `runtime/parser/parse_model.go:347` and :412-415.

## Product Improvements
- **parser-07** (above): project-level theme fallback is not applied to auto-emitted/inline explores — a visible inconsistency for users who set a global theme.
- **Unknown-field feedback in override blocks (parser-05):** users cannot currently tell that a typo'd key inside `dev:`/`prod:` silently does nothing; enforcing KnownFields there would turn a confusing no-op into an actionable error.
- **Localized failure isolation (parser-01/parser-09):** a single malformed file can currently fail the whole project parse via panic (the package has no `recover()`); converting these to per-file parse errors would substantially improve the "one bad file shouldn't break everything" experience that the rest of the parser already aims for.

## New Feature Ideas
- **Deprecation-warning pass (S/M):** The parser silently accepts a large set of deprecated aliases (`title`→`display_name`, `subtitle`→`description`, `label`→`display_name`, `dashboards`→`metrics_views`, legacy `email:` notify blocks, etc.). Emitting `addParseWarning` for each would guide users off deprecated syntax without breaking them.
- **Schema-backed validation wired into parse errors (M):** `schema.go` already embeds and parses JSON Schemas per resource type but they are only exposed for editor tooling. Running each resource's YAML through its schema during parse (as warnings first) would produce far better messages than ad-hoc `fmt.Errorf` checks and would catch structural mistakes (and typos in override blocks) uniformly.
- **Per-node panic recovery (S):** Wrap each `parseNode` call in a `recover` that converts an unexpected panic into a localized parse error on that node's paths. The package currently has no `recover()` at all, so any panic (parser-01, parser-04, parser-09) is a whole-project failure — defense-in-depth here directly addresses the top finding's blast radius.
- **Env-override linting (S):** Warn when an `environment_overrides:` block targets an environment the project never runs as, or when the same key is set both at top level and in every override (dead config).
- **Cycle/ref pre-validation warnings (M/L):** Ref-cycle and dangling-ref detection is entirely deferred to the reconciler. A parser-side pass that flags obvious cycles/self-refs would surface problems at edit time.
- **Consistent connector-analysis without live I/O (M):** `AnalyzeConnectors` performs network calls (`HasAnonymousSourceAccess` at connectors.go:112, `TertiarySourceConnectors` at :127) during analysis; a cached/offline mode would make connector metadata extraction deterministic and fast for tooling.

## Notes & Missing Test Coverage
- **parser-01:** No test exercises a component/canvas variable `value:` with a date, a non-string-keyed map, or other `structpb`-incompatible scalar — the exact inputs that panic. Add negative tests for these.
- **parser-02:** No test for a metrics-view annotation that omits both `model` and `table` (silently accepted on model/table-based views), nor for an annotation on a `parent`-based view (always rejected).
- **parser-03:** `parse_dotenv_test.go` only reparses a single `.env` at a time. There is no test for a multi-`.env` batch where a later file is unchanged, which is exactly the case that loses the flag.
- **parser-04:** No test calls `InitEmpty`/`UnpackEmpty` with an empty display name.
- **parser-06:** No test for `default_comparison.dimension` pointing at a measure.
- **parser-10:** No test for a DuckDB parse error whose position is at end-of-input.
- **Minor / not filed as findings:** (a) `GetDotEnvPerEnvironment` relies on lexicographic path sort to make "deeper `.env` wins" (parser.go:342); this holds for normal directory names but inverts if a nested directory name begins with a byte `< '.'` (e.g. `-foo/`). (b) Alert/report `renotify_after` is parsed and validated but only stored when `renotify` is also set (parse_alert.go:324-326, :352-354), so a lone `renotify_after` is silently ignored. Both are low impact.

## Verification Notes
Independently re-verified all nine original findings against the code; all nine are real. Corrections made:
- **parser-04:** downgraded HIGH-adjacent MEDIUM → **LOW**. The reviewer listed the CLI (`cli/cmd/initialize/init.go`) as a trigger, but `projectName := filepath.Base(projectPath)` never returns `""`, so the CLI path cannot panic. The only reachable trigger is the `UnpackEmpty` gRPC/REST handler with an explicit empty `display_name`. Also verified the runtime gRPC chain has **no** recovery interceptor (server.go:158-166); the panic is instead recovered per-request by `net/http` via the vanguard transcoder, so it degrades one request rather than crashing the process — hence low severity. Location/claim text corrected.
- **parser-01:** confirmed and strengthened — verified there is **no** `recover()` anywhere in `runtime/parser` and `parseNode` is not wrapped in one, so the panic aborts the entire parse (only the controller-level recover catches it). Added the inline-canvas-component call site (parse_canvas.go:447 via `parseComponentYAML`).
- **parser-03:** traced end-to-end into `runtime/reconcilers/project_parser.go:344`, confirming a dropped `ModifiedDotEnv` skips `reconcileProjectConfig` (env not applied, controller not restarted). Added the declaration line (:433).
- **parser-06:** verified the concrete constant values (`nameIsMeasure=1`, `nameIsDimension=2`, map-miss=0) that make the `&&` accept a measure while still rejecting non-existent names.
- **parser-02 / parser-05 / parser-07 / parser-08 / parser-09:** confirmed as written; added precise cross-reference line numbers (annotation struct at :79-87 and spec build at :933-942; `YAMLOverride` type at parse_node.go:23; correct `default_dimensions` loop at :466-468; dry-run guards at :1143 / :456).
- **Added parser-10** (new, LOW/medium): `findLineNumber` returns `-1` on out-of-range positions, which the single caller casts to `uint32`, producing a 4-billion line number in model SQL parse-error locations. Reachable for DuckDB end-of-input errors.
- **Refuted / dropped candidate:** a suspected nil-pointer deref at connectors.go:142 (`resolverProps.Fields` for a "sql"/"glob" API resolver) is **not** a bug — `parseDataYAML` always returns a non-nil `*structpb.Struct` from `structpb.NewStruct(make(map[string]any))`, and API resources always set `ResolverProperties` from that result. Not filed.

Findings removed as false positives: **0**. One finding downgraded (parser-04), one added (parser-10).
