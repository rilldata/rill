# Reconcilers — Code Review

**Scope:** `runtime/reconcilers/*.go` (model, source, metrics_view, explore, canvas, component, alert, report, api, connector, theme, migration, project_parser, refresh_trigger, util). Cross-checked against `runtime/drivers/sqlite/catalog.go`, `runtime/drivers/duckdb/model_manager.go`, `runtime/connections.go`, and the local/noop admin services.  |  **Health:** fair

## Overview

The reconcilers implement per-resource reconcile loops driven by the controller, and the code is generally thoughtful: idempotency and context-cancellation are handled deliberately, state is kept serve-able mid-reconcile, and spec/ref hashing is used to avoid redundant work. The most complex reconciler (`model.go`, ~2100 lines) carries most of the risk — partitioning, incremental execution, rename, and delete paths. The biggest risks I found are two correctness bugs that fire on common operations: a nil-pointer dereference when an alert with a non-email notifier runs without admin metadata (the default in Rill Developer), and a stale `prevResult` on the model rename path that forces a full rebuild (dropping incremental state). Secondary issues include an orphaned-partition-row leak on model delete, a duplicate report-delivery window on retry, and a partition-pagination query that silently returns nothing past the first batch. Test coverage is thin for several sizeable reconcilers (`report.go`, `project_parser.go`, `migration.go` have no dedicated tests).

## Findings

### reconcilers-01 [HIGH] Alert notification via non-email notifier dereferences nil `adminMeta`

**Status (2026-07-25):** ✅ Fixed and merged in [#9707](https://github.com/rilldata/rill/pull/9707).

- **Category:** bug  **Confidence:** high  **Location:** `runtime/reconcilers/alert.go:927`
- **Problem:** In `popCurrentExecution`, the email branch guards `if adminMeta != nil && adminMeta.RecipientURLs != nil` (line 892), but the default (non-email) notifier branch unconditionally does `urls, ok := adminMeta.RecipientURLs[""]`. `executeAll` explicitly tolerates a nil `adminMeta`: `admin.GetAlertMetadata` returning `drivers.ErrNotImplemented` leaves `adminMeta == nil` and swallows the error (alert.go:542-545). Both the local admin (`cli/pkg/local/admin.go:41`) and the noop admin return `(nil, ErrNotImplemented)`.
- **Failure scenario:** In Rill Developer, an alert configured with a Slack (or any non-email) notifier fails a check with `notify_on_fail: true`. `msg` is built, the notifier loop hits the default case, and `adminMeta.RecipientURLs` dereferences nil → panic. The controller recovers the panic (controller.go:1318) and records the alert error as `panic: runtime error: invalid memory address...`, so the notification never sends and the user sees an internal error instead.
- **Recommendation:** Guard the non-email branch the same way the email branch is guarded (check `adminMeta != nil` before touching `RecipientURLs`), and degrade gracefully to blank links when metadata is unavailable — matching the "alerts still work, links just won't be present" contract stated in the code comment at alert.go:522.

**Verification:** CONFIRMED — alert.go:927 does `urls, ok := adminMeta.RecipientURLs[""]` with no nil guard, while the email branch at alert.go:892 guards `adminMeta != nil && adminMeta.RecipientURLs != nil`. `executeAll` swallows `drivers.ErrNotImplemented` from `admin.GetAlertMetadata` leaving `adminMeta == nil` (alert.go:542-544), and the local admin service returns exactly `(nil, drivers.ErrNotImplemented)` (cli/pkg/local/admin.go:40-43). Dereferencing a field on a nil struct pointer panics in Go.

**Manual Reproduction Steps (for QA):**
1. In Rill Developer (`rill start`), open a project with a working metrics view.
2. Add a Slack connector (`connectors/slack.yaml` with a `bot_token`).
3. Add an alert YAML (`type: alert`) against the metrics view whose data query returns rows (so the alert fails/fires), with `notify: slack: channels: ["#some-channel"]` (no email notifier).
4. Save and let the alert reconcile (or trigger it via the UI / `rill project refresh` locally).
5. Expected: a Slack notification is sent (or, at worst, a clean error about missing links). Actual: the alert resource enters an error state whose message contains `panic: runtime error: invalid memory address or nil pointer dereference`, and no notification is sent.

### reconcilers-02 [HIGH] Model rename forces a full rebuild because `prevResult` is not updated to the renamed table

**Status (2026-07-25):** ✅ Fixed and merged in [#9709](https://github.com/rilldata/rill/pull/9709).

- **Category:** correctness  **Confidence:** high  **Location:** `runtime/reconcilers/model.go:191` (rename) and `model.go:271-272` (exists check)
- **Problem:** On the rename path, `prevManager.Rename` renames the physical output table to the new name and `updateStateWithResult` records the new `State.ResultTable`, but the local `prevResult` variable (built at model.go:124-129 from the *old* table name) is never updated. Later, `exists, err = prevManager.Exists(ctx, prevResult)` looks up the old table name, which no longer exists after the rename (`duckdb/model_manager.go:104-141`, `UsedModelName == true`, the default). `exists` comes back false, so `resolveTrigger` computes `firstRun = true` (model.go:1912) → `specChanged = true` → a full reset run.
- **Failure scenario:** A user renames `orders.sql` to `orders_v2.sql` (unchanged SQL, so the parser emits a rename, not delete+create). The reconciler renames the DuckDB table cheaply, then immediately concludes the output is missing (`tableExists=false`) and fully rebuilds the model. In RESET mode (the default) `firstRun=true` → `specChanged=true` → `reset=true`. For an incremental model this forces `incrementalRun=false` (model.go:1112), so it discards `IncrementalState` and re-processes the full model logic rather than the cheap rename — expensive for large models. Note: this is a rebuild/state-loss regression, not data duplication (a reset re-runs the full-refresh branch, which replaces rather than appends).
- **Confirmed:** `Rename` only physically renames the table when `resProps.UsedModelName == true` (duckdb/model_manager.go:110-112), which is the default. When the table name is explicitly pinned (`UsedModelName == false`), `Rename` is a no-op and `prevResult` stays valid, so this bug does not fire — the regression is limited to the (common) model-name-derived-table case.
- **Recommendation:** After a successful `Rename`, assign the returned `renameRes` back into `prevResult` (and keep `model` state in sync) so the subsequent `Exists`/trigger logic sees the renamed table. Alternatively, skip the `Exists` recheck when a rename just succeeded.

**Verification:** CONFIRMED — the rename branch (model.go:181-205) captures `renameRes` at model.go:191 and passes it only to `updateStateWithResult`; the local `prevResult` (built at model.go:124-129 from the pre-rename `State.ResultTable`) is never reassigned. The later existence check `prevManager.Exists(ctx, prevResult)` (model.go:272) therefore looks up the old table, which DuckDB's `Rename` physically renamed away when `UsedModelName` is true (runtime/drivers/duckdb/model_manager.go:104-131). `exists=false` makes `firstRun=true` in `resolveTrigger` (model.go:1912) → `specChanged=true` → a full reset run in the default RESET mode.

**Manual Reproduction Steps (for QA):**
1. In Rill Developer, create a model that is expensive enough to notice a rebuild (e.g. a model over a large source, or an `incremental: true` model with `state:`), and let it build fully. Note the build duration in the logs.
2. Rename the model file (e.g. `orders.sql` → `orders_v2.sql`) without changing its contents, and save.
3. Expected: the underlying table is cheaply renamed and the model does not re-execute; for an incremental model, its incremental state is preserved. Actual: the model immediately re-executes a full build (same duration as step 1 visible in logs), and an incremental model loses its `IncrementalState` (restarts from scratch, observable via the model's state/refreshed data).

### reconcilers-03 [MEDIUM] Deleting a materialized partitioned model leaks its partition rows in the catalog

**Status (2026-07-25):** ✅ Fixed and merged in [#9711](https://github.com/rilldata/rill/pull/9711).

- **Category:** correctness  **Confidence:** high  **Location:** `runtime/reconcilers/model.go:161-170`
- **Problem:** In the deletion branch, when `prevManager != nil` (the model produced output) the code drops the output table and `return`s immediately at line 169, so `r.clearPartitions` is only reached in the fall-through branch (`prevManager == nil`, line 172). `clearPartitions` → `DeleteModelPartitions` is the only cleanup for `model_partitions` rows (verified: the sole caller of `DeleteModelPartitions` is `clearPartitions`, and the `model_partitions`/`model_splits` table (migrations 0024/0027) has no foreign key or `ON DELETE CASCADE`). `PartitionsModelId` is a fresh `uuid.NewString()` per model (model.go:915), so a delete+recreate cannot reclaim the old rows.
- **Failure scenario:** A partitioned incremental model runs successfully (so `ResultConnector != ""` → `prevManager != nil`), then its file is deleted. The output table is dropped, but every row it created in the `model_partitions` SQLite table is orphaned forever. Repeated create/delete churn of partitioned models grows the catalog DB unbounded.
- **Recommendation:** Call `r.clearPartitions(ctx, model)` in the `prevManager != nil` delete branch as well (before or after `prevManager.Delete`), so partition metadata is cleaned up on every delete path.

**Verification:** CONFIRMED — in the deletion branch (model.go:160-177), when `prevManager != nil` the code calls `prevManager.Delete` and returns at model.go:168-169 without ever calling `r.clearPartitions`; the `clearPartitions` call at model.go:172 is only reached when `prevManager == nil` (model never produced output). `clearPartitions` (model.go:1092) is the sole caller of `DeleteModelPartitions`, and `model_partitions` has no cascade, so the rows are orphaned. `PartitionsModelId` is a fresh UUID per model, so re-creating the model cannot reclaim them.

**Manual Reproduction Steps (for QA):**
1. In Rill Developer, create a partitioned model, e.g. a YAML model with `partitions: sql: SELECT range AS num FROM range(0,50)` and `incremental: true`, and let it build.
2. Verify partition rows exist in the catalog: `sqlite3 <project>/tmp/meta.db "SELECT COUNT(*) FROM model_partitions;"` (should be 50).
3. Delete the model file and wait for reconcile (the output table is dropped).
4. Re-run the query from step 2. Expected: 0 rows remain for that model. Actual: all 50 rows are still present (orphaned). Repeating create/delete grows the count each cycle.

### reconcilers-04 [MEDIUM] Report retry can re-deliver to notifiers that already succeeded
- **Category:** correctness  **Confidence:** high  **Location:** `runtime/reconcilers/report.go:594-610`
- **Problem:** `sendReport` tracks whether notifications were sent with a single `sent` variable that is *overwritten* on each notifier iteration (`sent, err = ...`) rather than OR-accumulated. The returned bool is the "dirty" flag that `executeSingle` uses to decide whether an interrupted run may auto-retry (report.go:467-475). If an earlier notifier delivered successfully and a later notifier is then cancelled/fails before sending, `sent` reflects only the last (unsent) notifier — `sendNonEmailNotification`/`sendEmailNotification` both return `false` on a pre-delivery failure (report.go:649/655/660 and 616) — so the run is treated as "clean" and retried.
- **Failure scenario:** A report has an email notifier and a Slack notifier. Email delivery succeeds (`sent=true`); the Slack send is then cancelled (context deadline) and returns `(false, err)`. `sent` is now `false` at the `return sent, nil, err` on line 605-607 → `dirtyErr=false`. In `executeSingle`, a `context.Canceled/DeadlineExceeded` error with `dirtyErr=false` sets `retry=true` (report.go:469-475). `executeAll` then clears `CurrentExecution` ("pretends it never happened", report.go:358-367) and does not advance `previousWatermark`, so the next reconcile recomputes the same execution time and re-sends to the already-delivered email recipients.
- **Traced end-to-end:** confirmed the re-delivery path through `executeSingle` → `executeAll` → `calculateReportExecutionTimes` (unchanged watermark). The generalized trigger is "2+ notifiers where an earlier one delivered and a later one is cancelled" (email+Slack is one instance).
- **Recommendation:** Accumulate delivery state across notifiers (`sent = sent || ...`), and return `dirty=true` if any notifier delivered before a later failure. Consider tracking per-notifier delivery so retries can skip already-delivered notifiers.

**Verification:** CONFIRMED — `sendReport` initializes `sent := false` at report.go:594 and overwrites it on each notifier iteration (`sent, err = ...` at report.go:599 and 605), returning it on error (report.go:601, 607). `sendNonEmailNotification` returns `false` on any pre-delivery failure (report.go:645-661), so a cancelled later notifier erases an earlier successful delivery. In `executeSingle`, `dirtyErr=false` plus `context.Canceled`/`DeadlineExceeded` sets `retry=true` (report.go:466-476); `executeAll` then clears `CurrentExecution` (report.go:358-367) without advancing the watermark, so the same execution time is recomputed and the already-delivered email recipients are re-sent.

**Manual Reproduction Steps (for QA):**
1. Not reproducible in Rill Developer: `sendReport` exits early there because the local admin returns `ErrNotImplemented` for report metadata (report.go:530-535 logs "Skipped sending report").
2. In Rill Cloud: create a report with two notifiers — an email recipient and a Slack channel — and a small `timeout_seconds`.
3. Arrange for the Slack delivery to be slow or unreachable (e.g. blocked egress to slack.com) so the email sends successfully but the Slack send hits the deadline.
4. Expected: the interrupted run is marked dirty and not auto-retried (or a retry skips email). Actual: the run is treated as clean, auto-retries, and the email recipient receives the same report twice.
5. If the environment cannot be arranged, confirm in code: runtime/reconcilers/report.go:594-611 (`sent` overwritten per iteration) and report.go:466-476 (retry decision).

### reconcilers-05 [MEDIUM] `markPartitionsExecutedExceptFirst` stops after the first 1000 partitions
- **Category:** correctness  **Confidence:** high  **Location:** `runtime/reconcilers/model.go:1596-1635`
- **Problem:** The loop paginates pending partitions using `AfterKey`, but `FindModelPartitions` implements `AfterKey` with `AND (executed_on < ? OR (executed_on = ? AND key > ?))` (sqlite/catalog.go:170-173) while `WherePending` requires `executed_on IS NULL` (catalog.go:158-160). For pending rows `executed_on` is NULL, so the `AfterKey` predicate is always NULL/false — any `WherePending=true` query with a non-empty `AfterKey` returns zero rows. Thus the second loop iteration always returns empty and breaks.
- **Failure scenario:** A model with >1000 partitions does a forced incremental run with no prior state. Only the first batch of 1000 is handled (partitions 1-999 marked executed, partition 0 kept pending); partitions 1000..N are never marked and are executed in this run instead of being deferred — violating the "run only the first partition" invariant and, for append-mode partitions, processing far more data than intended.
- **Recommendation:** Paginate pending partitions by `idx`/`key` alone (the query already orders by `idx DESC`) rather than reusing the `executed_on`-based `AfterKey` predicate, or drive the loop off a monotonically advancing offset. Add a test with >1000 partitions.

**Verification:** CONFIRMED — `markPartitionsExecutedExceptFirst` (model.go:1596-1637) loops with `WherePending: true` and a growing `afterKey`, batch size `_modelPendingPartitionsBatchSize = 1000` (model.go:38). In `FindModelPartitions`, `WherePending` adds `executed_on IS NULL` (runtime/drivers/sqlite/catalog.go:159-161), while a non-empty `AfterKey` adds `AND (executed_on < ? OR (executed_on = ? AND key > ?))` (catalog.go:170-173). For pending rows `executed_on` is NULL, so the AfterKey predicate evaluates to NULL (false) and the second iteration always returns zero rows, breaking the loop after the first 1000.

**Manual Reproduction Steps (for QA):**
1. In Rill Developer, create a partitioned incremental model that will hit the `prevResult == nil && incrementalRun` path (model.go:1216-1221): e.g. YAML model with `incremental: true`, `change_mode: manual`, and `partitions: sql: SELECT range AS num FROM range(0,2500)`.
2. Since change mode is manual, the model waits for a manual trigger. Trigger an *incremental* refresh (not a full refresh) from the UI or API.
3. Expected: exactly 1 partition executes; the other 2499 are marked executed without running ("at least one partition runs" invariant). Actual: partitions in the first batch (minus the first) are marked executed, but partitions 1000-2499 remain pending and are all executed in this run — observable as ~1501 partition executions in the logs, or via the model's partitions list (GetModelPartitions API / model partitions UI) showing per-partition `elapsed_ms` for rows that should have been skipped.

### reconcilers-06 [MEDIUM] Migration reconciler re-executes the current top version on every reconcile
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/reconcilers/migration.go:83`
- **Problem:** The loop is `for v := from; v <= to; v++` where `from = mig.State.Version` (the last-applied version). Because the lower bound is inclusive, version `from` re-executes every time `Reconcile` runs, and there is no spec-hash gate. `Reconcile` runs whenever the migration is (re)reconciled — a ref-state change (if the migration has refs), a manual re-trigger, or a controller state reset — so the highest already-applied migration SQL re-executes on each such invocation. Even a migration with no `max_version` has `Spec.Version == 0`, so `from == to == 0` and its SQL re-runs on every reconcile.
- **Failure scenario:** A migration whose SQL is not perfectly idempotent (e.g. `INSERT INTO audit VALUES (...)` or `ALTER TABLE ... ADD COLUMN c`) re-runs on every deploy/restart, duplicating rows or erroring on the second run. The whole point of tracking `State.Version` — to avoid re-running applied migrations — is defeated.
- **Recommendation:** Advance from `from + 1` for already-applied state, and handle the initial run explicitly (e.g. treat an unset state version as a sentinel, or 1-index versions) so each version executes exactly once. If re-running is genuinely intended, drop the `State.Version` machinery and document the idempotency requirement.

**Verification:** CONFIRMED — migration.go:76-97: `from := mig.State.Version`, `to := mig.Spec.Version`, then `for v := from; v <= to; v++` executes version `from` (the last-applied version) again on every `Reconcile` invocation. There is no spec-hash or "already applied" gate anywhere in the reconciler, and a migration without `max_version` has `from == to == 0`, so its SQL re-runs on each reconcile.

**Manual Reproduction Steps (for QA):**
1. In Rill Developer, create a migration file, e.g. `migrations/m1.yaml` with `type: migration`, `connector: duckdb`, and non-idempotent SQL such as `ALTER TABLE some_table ADD COLUMN extra INT` (or `CREATE TABLE t (x INT)`).
2. Let it reconcile once successfully.
3. Force a re-reconcile of the migration resource without changing its SQL. Two reliable ways: (a) give the migration a ref to a model (e.g. reference the model with `{{ ref "my_model" }}` in the SQL) and then refresh that model — the ref-state change re-reconciles the migration; or (b) issue a refresh trigger targeting the migration resource via the CreateTrigger API.
4. Expected: the already-applied version is skipped (no-op). Actual: the SQL executes again and the migration errors with "column already exists" / "table already exists" (or silently duplicates rows for an `INSERT`).

### reconcilers-07 [LOW] Migration "version too large" check underflows on downgrade
- **Category:** correctness  **Confidence:** high  **Location:** `runtime/reconcilers/migration.go:79`
- **Problem:** `from` and `to` are `uint32`. `if to-from > 100` underflows when `to < from` (a user lowering `max_version`), producing a huge value that trips the "difference ... is too large" error with a misleading message. The subsequent loop would be a no-op anyway.
- **Failure scenario:** A user reduces `max_version` from 5 to 3; reconcile fails with "difference between migration versions 5 and 3 is too large" instead of a clear "downgrades are not supported" (or a no-op).
- **Recommendation:** Explicitly handle `to <= from` (no-op or a clear "downgrade not supported" error) before the magnitude check, and compute the delta with a signed comparison.

**Verification:** CONFIRMED — migration.go:76-80: `from`/`to` derive from `mig.State.Version`/`mig.Spec.Version`, which are proto `uint32` fields (and `executeMigration` takes `version uint32`, migration.go:109). `if to-from > 100` at migration.go:79 underflows when `to < from`, producing a value near 2^32 that trips the "too large" error with a misleading message.

**Manual Reproduction Steps (for QA):**
1. In Rill Developer, create a migration with `max_version: 5` (any idempotent SQL) and let it apply, so `State.Version == 5`.
2. Edit the file to lower it to `max_version: 3` and save.
3. Expected: a no-op or a clear "downgrade not supported" message. Actual: the migration resource errors with "difference between migration versions 5 and 3 is too large".

### reconcilers-08 [LOW] `attemptRename` mutates `def.ModelSpec` trigger flags while merely comparing
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/reconcilers/project_parser.go:634` (calls `newResourceIfModified`) and `project_parser.go:684-688`
- **Problem:** `newResourceIfModified` copies the *existing* resource's ephemeral trigger flags into `def.ModelSpec` (to compare specs), mutating the shared parser `def` in place. `attemptRename` calls it against each delete-candidate. If no rename matches, the mutated `def` (now carrying an unrelated deleted resource's `Trigger`/`TriggerFull`/`TriggerPartitions`) is passed to `putParserResourceDef` for a fresh insert.
- **Failure scenario:** A new model is inserted in the same reparse where an unrelated model with `Trigger=true` is being deleted; the rename probe copies `Trigger=true` into the new model's spec, so the freshly created model spuriously triggers an extra refresh.
- **Recommendation:** Compare against a clone of `def` (or copy trigger flags into a temporary) inside the rename/comparison path so the source `def` is not mutated when no rename occurs.

**Verification:** CONFIRMED — `newResourceIfModified` mutates `def.ModelSpec.Trigger/TriggerFull/TriggerPartitions` in place from the existing resource before comparing (project_parser.go:683-691). `attemptRename` calls it at project_parser.go:634 for each delete candidate (loop at project_parser.go:424-441), and when no rename matches, the now-mutated `def` is inserted fresh via `putParserResourceDef(ctx, inst, self, def, nil)` (project_parser.go:444), carrying over the deleted model's ephemeral trigger flags. Preconditions are narrow: the delete candidate must be a model with the same kind and identical refs, and must have a trigger flag set at reparse time.

**Manual Reproduction Steps (for QA):**
Practically not user-observable: a freshly inserted model runs anyway, so a spuriously copied `Trigger` flag produces no visible behavior difference in most cases (a copied `TriggerFull`/`TriggerPartitions` could subtly alter the first run mode). Not user-reproducible in a reliable way — confirm in code: runtime/reconcilers/project_parser.go:683-691 (in-place mutation of `def.ModelSpec`), project_parser.go:634 (rename probe), and project_parser.go:444 (mutated `def` used for the insert).

### reconcilers-09 [LOW] PATCH change-mode persists the new spec hash before the execution runs
- **Category:** correctness  **Confidence:** medium  **Location:** `runtime/reconcilers/model.go:1965-1971`
- **Problem:** In `resolveTrigger`, PATCH mode writes `model.State.SpecHash = specHash` and persists it *before* returning the trigger and running the execution. If that execution fails while `StageChanges` is enabled (so failure does not clear state), the next reconcile sees `specChanged == false` and does not re-detect the definition change; a model without a refresh schedule then won't retry until manually triggered.
- **Failure scenario:** A patch-mode incremental model's new logic errors on first run (StageChanges on); the spec hash is already recorded as applied, so the model keeps serving old data with no automatic retry and no "spec changed" signal.
- **Recommendation:** Persist the new spec hash only after a successful execution (in `updateStateAfterExecution`), or record a separate "pending patch" marker so a failed patch is retried on the next reconcile.

**Verification:** PLAUSIBLE — the code fact is real: in PATCH mode `resolveTrigger` persists `model.State.SpecHash = specHash` via `UpdateState` before any execution (model.go:1964-1970), and a failed execution with `StageChanges` enabled does not clear state (model.go:423-436). However, the claimed consequence is weaker than stated: in PATCH mode `specChanged` never feeds into the returned trigger (`normal: Trigger || scheduledTrigger || firstRun`, `reset: TriggerFull` — model.go:1972-1976), so a spec edit alone triggers no execution at all, and even if the hash were persisted only after success, the re-detected `specChanged` on the next reconcile would still not trigger a run. The behavioral delta of the early persist is therefore limited (e.g. downstream models whose spec hash incorporates this model's `State.SpecHash` at model.go:527-529 see the change before the new logic has produced data), rather than the "lost retry" described.

**Manual Reproduction Steps (for QA):**
Not meaningfully user-reproducible as a behavioral regression: in patch mode the spec change by itself does not start a run, so there is no "failed patch execution" to observe being skipped. Confirm in code: runtime/reconcilers/model.go:1964-1976 (hash persisted inside `resolveTrigger` before returning a trigger that does not include `specChanged`), and model.go:423-436 (`StageChanges` keeps state on failure).

### reconcilers-10 [LOW] `truncateSummary` slices by bytes and can split UTF-8 runes
- **Category:** bug  **Confidence:** high  **Location:** `runtime/reconcilers/report.go:788-793`
- **Problem:** `summary[:maxLen-3]` slices raw bytes, which can split a multi-byte UTF-8 rune and yield invalid UTF-8 in report summaries.
- **Failure scenario:** An AI report summary whose 497th byte lands mid-rune (common with emoji/non-ASCII) produces a garbled trailing character in the delivered Slack/email summary. Cosmetic only.
- **Scope note (corrected):** The prior draft claimed a possible negative-index panic for `maxLen < 3`. Verified this is **not currently reachable**: the sole call site passes `maxLen == 500` (report.go:575), and the function is unexported, so the underflow is theoretical. The only real defect today is the rune-splitting.
- **Recommendation:** Truncate on rune boundaries (e.g. convert to `[]rune`, or use `utf8`-aware trimming); if the function ever gains new callers, also guard `maxLen < len(ellipsis)`.

**Verification:** CONFIRMED — `truncateSummary` (report.go:787-793) does `summary[:maxLen-3]`, a raw byte slice that can split a multi-byte UTF-8 rune. The sole caller passes `maxLen == 500` (report.go:575, AI report summaries), so the negative-index underflow is indeed unreachable today; the rune-splitting is the only live defect and is cosmetic.

**Manual Reproduction Steps (for QA):**
Not practically user-reproducible: it requires an AI-generated report summary longer than 500 bytes whose 497th byte falls mid-rune (emoji/non-ASCII), which cannot be controlled deterministically. Confirm in code: runtime/reconcilers/report.go:788-792 — or add a quick unit test calling `truncateSummary(strings.Repeat("é", 300), 500)` and checking `utf8.ValidString` on the result (it returns invalid UTF-8).

## Product Improvements

### reconcilers-11 [MEDIUM] Project-parser watch failures retry immediately with no backoff
- **Category:** product-improvement  **Confidence:** medium  **Location:** `runtime/reconcilers/project_parser.go:258`
- **Problem:** When `repo.Watch` returns a non-cancellation error, the reconciler logs and returns `ReconcileResult{Err: err, Retrigger: time.Now()}`, asking the controller to retry immediately. A persistently failing watch (e.g. a repo/filesystem error that keeps recurring) produces a tight error/log loop with no backoff.
- **Failure scenario:** A transient-but-recurring filesystem or repo error causes the parser to re-enter watch, fail, log "Stopped watching for file changes, retrying...", and immediately retry — spamming logs and burning CPU.
- **Recommendation:** Apply a bounded exponential backoff (with jitter) to the `Retrigger` time for repeated watch failures, resetting on a successful watch establishment.

**Verification:** CONFIRMED — project_parser.go:250-259: on a non-context-cancellation watch/re-parse error, the reconciler logs "Stopped watching for file changes, retrying..." and returns `runtime.ReconcileResult{Err: err, Retrigger: time.Now()}` (project_parser.go:258), i.e. an immediate retry with no backoff or jitter anywhere in the path.

**Manual Reproduction Steps (for QA):**
1. Run `rill start` on a project.
2. Break the file watcher persistently, e.g. remove read permission from the project directory (`chmod 000 <project>`) or delete/replace the directory the watcher is bound to.
3. Watch the terminal logs. Expected: retries with increasing delay (backoff). Actual: a tight loop of "Stopped watching for file changes, retrying..." error lines repeating with no delay, with corresponding CPU/log churn, until the underlying error is fixed.

## New Feature Ideas

- **Partition failure management UI/API (S–M):** Surface errored/skipped partitions per model and let users re-run just the failed ones. The catalog already tracks per-partition `Error`/`ExecutedOn` and `RefreshTrigger` supports `AllErroredPartitions`; a thin API + UI would turn this into a first-class recovery flow instead of a full refresh.
- **Catalog GC for orphaned partition rows (S):** A periodic sweep that deletes `model_partitions` rows whose `model_id` no longer corresponds to a live model would bound catalog growth and mitigate reconcilers-03 defensively even if a delete path is missed.
- **"Why did this rebuild?" reconcile diagnostics (S):** Record the trigger reason (spec change vs ref-hash change vs manual vs missing-output) into model state so users and support can see why an expensive rebuild happened — this would have made reconcilers-02 obvious in the field.
- **Migration dry-run / plan (M):** A command that reports which migration versions will execute (from → to) and warns on downgrades/re-runs before applying, giving users confidence in the migration mechanism.
- **Configurable execution-history retention (S):** Alert (`alertExecutionHistoryLimit = 25`) and report (`reportExecutionHistoryLimit = 10`) history limits are hard-coded; exposing them per-resource would help debugging noisy alerts and audit-heavy reports.
- **Idempotent-delivery keys for notifiers (M):** Attach an execution-scoped idempotency key to each notification so retries (see reconcilers-04) can be deduplicated by the notifier/email layer, eliminating duplicate report/alert delivery entirely.

## Notes & Missing Test Coverage

- **No dedicated tests** for `report.go` (932 lines, complex delivery/interval logic), `project_parser.go` (743 lines, rename/diff reconciliation), `migration.go`, `connector.go`, `refresh_trigger.go`, `component.go`, `theme.go`, `api.go`, or `source.go`. The two highest-value gaps are `report.go` (the retry/dirty logic in reconcilers-04) and `project_parser.go` (rename vs delete+create decisions).
- **reconcilers-05** would be caught by a test that syncs >1000 partitions and asserts the "except-first" invariant; there appears to be no partition-count boundary test.
- **reconcilers-02** (rename → rebuild) is not covered; a test asserting that renaming a materialized model does *not* clear `IncrementalState` / does not re-execute would lock in the fix.
- **Performance note (not a formal finding):** `queryOutputTableStats` runs `SELECT COUNT(*)` plus a physical-size lookup after every model execution (model.go:771-813, and again after the first partition). It's cheap on DuckDB/ClickHouse metadata but could be a full scan on some OLAPs; worth confirming for large external output tables.
- I could not fully confirm the *intended* semantics of migration versioning (reconcilers-06) without product docs; the finding is based on the `State.Version` tracking machinery implying once-only execution. Confidence is medium accordingly.

## Verification Notes

Independently re-verified every finding against the source (traced call sites, error paths, and existing guards). **All 11 findings survived** — none were false positives — but several were corrected for accuracy:

- **Nothing removed.** `removed_count = 0`.
- **reconcilers-01** — Confirmed. `localAdminService.GetAlertMetadata` returns `(nil, drivers.ErrNotImplemented)` (cli/pkg/local/admin.go:41-43); `executeAll` swallows that error leaving `adminMeta == nil` (alert.go:542-545); the non-email branch dereferences `adminMeta.RecipientURLs[""]` unguarded (alert.go:927) while the email branch guards it (alert.go:892). Controller recovers the panic (controller.go:1319). Notably, the *report* equivalent is not affected because `sendReport` returns early on `ErrNotImplemented` (report.go:532-535) — this asymmetry confirms the alert path is a genuine oversight.
- **reconcilers-02** — Confirmed, with two corrections: (a) removed the speculative "duplicated data for append-mode" claim — a reset run sets `incrementalRun=false` and re-runs the full-refresh logic (replace, not append); the real harm is an expensive full rebuild plus loss of `IncrementalState`. (b) Documented the `UsedModelName` guard: the bug only fires when the output table name is derived from the model name (the default); a pinned table name makes `Rename` a no-op so `prevResult` stays valid.
- **reconcilers-03** — Confirmed and tightened the location to `model.go:161-170`. Verified `DeleteModelPartitions` has exactly one caller (`clearPartitions`) and the table has no cascade (migrations 0024/0027), and that `PartitionsModelId` is a per-model UUID, so orphaned rows are never reclaimed.
- **reconcilers-04** — Confidence raised **medium → high** after tracing the full re-delivery path (`sendReport` overwrite → `dirtyErr=false` → `retry=true` → `CurrentExecution` cleared → unchanged watermark → same execution time recomputed). Generalized the trigger to "2+ notifiers where an earlier delivered and a later is cancelled".
- **reconcilers-05** — Confirmed. `WherePending` requires `executed_on IS NULL` (sqlite/catalog.go:158-160); the `AfterKey` predicate `executed_on < ? OR (executed_on = ? AND key > ?)` (catalog.go:170-173) is NULL/false for those rows, so the second loop iteration always returns empty. Verified the sibling execution loop (model.go:1274) is correct because executed partitions drop out of `WherePending` without needing `AfterKey`.
- **reconcilers-06** — Confirmed the code behavior (inclusive lower bound re-runs version `from`). Corrected the framing from "every runtime restart" to "every reconcile invocation after apply" (ref-state change / manual re-trigger / state reset), since an idle migration is not necessarily re-reconciled on a plain restart. Kept at medium confidence: the behavior is certain, only the *intended* semantics are undocumented.
- **reconcilers-07** — Confirmed. `from`/`to` are `uint32` (passed to `executeMigration(version uint32)`), so `to-from` underflows when `to < from`.
- **reconcilers-08** — Confirmed the in-place mutation of `def.ModelSpec` (project_parser.go:684-688) via `attemptRename` → `newResourceIfModified`. The mutation is *intended* for the spec comparison (see the explicit NOTE at refresh_trigger.go:188/222 about keeping trigger flags out of the modified-check), but it leaks into the subsequent insert when no rename matches. Impact is narrow (requires an unrelated delete candidate with identical refs and a set trigger flag), so kept at low severity / medium confidence.
- **reconcilers-09** — Confirmed. In PATCH mode, `model.State.SpecHash` is persisted before the trigger returns (model.go:1965-1971), and a failed execution with `StageChanges` on does not clear state (model.go:422-435), so `specChanged` reads false on the next reconcile. Narrow (PATCH + a run that actually executes and fails + no schedule to re-trigger); kept low/medium.
- **reconcilers-10** — Corrected: the claimed negative-index **panic is not reachable** (sole caller passes `maxLen == 500`, report.go:575; function is unexported). Retitled and rescoped to the only real defect — byte-slicing can split a UTF-8 rune — which is cosmetic. Kept at low.
- **reconcilers-11** — Confirmed as a valid product-improvement: `repo.Watch` failure returns `Retrigger: time.Now()` (project_parser.go:258) with no backoff.

Additional sweep of the in-scope files not central to the original findings (`connector.go`, `source.go`, `component.go`, `api.go`, `theme.go`, `metrics_view.go`, `refresh_trigger.go`, `util.go`, and the partition execution/termination paths in `model.go`) surfaced no new high-confidence bugs worth adding. The partition execution loop (model.go:1274-1364) terminates correctly because errored partitions still get `ExecutedOn` set (model.go:1411-1418) and cancellations propagate out of `errgroup.Wait`. One minor observation not raised to a finding: `metrics_view.go:173` assigns `mv.State.ValidSpec = mv.Spec` after mutating `mv.Spec.Table`/`Connector` in place (they share the pointer), which is consistent with the codebase's resolved-spec persistence pattern and not a defect.
