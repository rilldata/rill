# Dashboard State Management — Code Review

**Scope:** `web-common/src/features/dashboards/{stores, state-managers, url-state, proto-state, time-controls, time-dimension-details}`  |  **Health:** needs-attention

## Overview
This area owns the explore dashboard's in-memory state (`metricsExplorerStore`), the cascading
initialization from multiple sources (session storage, URL, bookmark/token, YAML defaults, Rill
defaults), the two-way sync between store and URL (`DashboardStateSync`), the URL <-> state
serializers (`url-state`, legacy `proto-state`), and time-range/comparison derivation
(`time-controls`). The code is well-factored and the cascading-merge and URL-conversion layers are
carefully written with a lot of edge-case handling. The concentrated risk is the `DashboardStateSync`
boolean `updating` lock: on two early-return paths it is never released, which permanently wedges the
URL/state sync, and its coarse boolean semantics also make concurrent navigations fragile. Time-range
derivation is mostly pure and defensive. Test coverage is strong for URL conversion but absent for the
`DashboardStateSync` lock lifecycle.

## Findings

### frontend-dashboard-state-01 [high] `updating` lock is never released on two early-return paths, permanently freezing URL/state sync

**Status (2026-07-25):** ✅ Fixed and merged in [#9701](https://github.com/rilldata/rill/pull/9701).

- **Category:** concurrency  **Confidence:** high  **Location:** `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:215,225,235,288`
- **Problem:** `handleURLChange` sets `this.updating = true` (line 215) and only resets it to `false` on the success path (line 288). The two guard clauses after it — `if (!rillDefaultExploreURLParams) return;` (line 225) and `if (!partialExplore) return;` (line 235) — return without resetting `updating`. There is no `try/finally`. Note these two returns occur **before** the `resolveTimeRanges` await (line 239), so the leak is a plain synchronous early return, not an await race. Secondarily, `gotoNewState` (line 312) and any throw between `updating = true` and the reset (e.g. from `getUrlForExploreState`) leak the lock the same way.
- **Failure scenario:** In Rill Developer, editing the metrics-view/explore YAML triggers a refetch of the valid-spec / default-state compound queries. If an `afterNavigate` fires while `rillDefaultExploreURLParams` is transiently `undefined` (line 225) or `getExploreStateFromURLParams` returns `undefined` because a dependent query has not reloaded (line 235; the code comment "This shouldn't ideally happen" acknowledges the input can be `undefined`), `handleURLChange` returns with `updating` stuck at `true`. From then on every `handleURLChange` returns at line 214 and every `gotoNewState` returns at line 311, so filter/time/pivot changes stop updating the URL, session storage, and the "most recent" store until a full page reload.
- **Recommendation:** Wrap the body of `handleURLChange` and `gotoNewState` in `try { ... } finally { this.updating = false; }` so the lock is always released regardless of early return or throw.

**Verification:** CONFIRMED — `DashboardStateSync.ts:215` sets `this.updating = true`; the guard returns at line 225 (`if (!rillDefaultExploreURLParams) return;`) and line 235 (`if (!partialExplore) return;`) exit without resetting it, and there is no `try/finally`. The only resets are at lines 288, 347, and 352. Once leaked, line 214 (`if (this.updating || !this.initialized) return;`) and line 311 permanently short-circuit both sync paths. `gotoNewState`'s `await goto(newUrl)` at line 351 before the reset at 352 can likewise leak on throw.

**Manual Reproduction Steps (for QA):**
1. In Rill Developer, open an explore dashboard and keep the YAML editor for its metrics view or explore open in a second window/split.
2. Edit and save the metrics view YAML (e.g. rename a measure label) so the spec queries refetch, and within the same moment navigate inside the dashboard (click a dimension leaderboard row to drill in, or press the browser Back button) so `afterNavigate` fires while the refetched queries are transiently unloaded.
3. After the dashboard settles, change a filter or the time range picker.
4. Expected: the browser URL updates to reflect the new filter/time range. Actual (when the race is hit): the URL never changes again for any interaction, and reopening the dashboard from another page restores stale state, until a full page reload.
5. This is timing-dependent; if it does not reproduce in a few tries, confirm in code: `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:215,225,235` — `updating = true` followed by two bare `return`s with no `finally`.

### frontend-dashboard-state-02 [low] `decompressUrlParams` decodes bytes with `String.fromCharCode` rather than `TextDecoder` (currently correct, but fragile)
- **Category:** readability  **Confidence:** low  **Location:** `web-common/src/features/dashboards/url-state/compression.ts:11,22`
- **Problem:** `compressUrlParams` encodes with `new TextEncoder().encode(params)` (UTF-8) while `decompressUrlParams` reconstructs the string with `String.fromCharCode(...decompressedParams)`, which maps each byte to a code unit rather than decoding UTF-8. The encode/decode are therefore not true inverses in the general case. **This is not a live bug today:** the only caller compresses `cleanedUrlParams.toString()` (`convert-partial-explore-state-to-url-params.ts:76`), and `URLSearchParams.toString()` percent-encodes every non-ASCII character to pure ASCII (`Zürich` -> `Z%C3%BCrich`). Every byte fed to `TextEncoder` is therefore < 128, so `String.fromCharCode` round-trips losslessly and `new URLSearchParams(...)` re-decodes the percent-escapes back to the original value. The `RangeError`-on-spread concern is also not reachable: compression only triggers above 2000 chars and the URL is hard-capped near 16 KB (`url-length-limits.ts`), well under the argument-count limit that would make `String.fromCharCode(...)` throw.
- **Failure scenario:** No user-visible failure with the current single call site. The risk is latent: if a future caller ever passes a non-percent-encoded string (raw non-ASCII) to `compressUrlParams`, the value would come back as mojibake with no error.
- **Recommendation:** Decode with `new TextDecoder().decode(decompressedParams)` to make the round trip symmetric and robust regardless of caller, and add a round-trip unit test covering non-ASCII and large payloads (there are currently no compression tests).

**Verification:** PLAUSIBLE — the asymmetry exists exactly as described: `compression.ts:11` uses `new TextEncoder().encode(params)` (UTF-8) while `compression.ts:22` uses `String.fromCharCode(...decompressedParams)` (byte-to-code-unit). The sole caller (`convert-partial-explore-state-to-url-params.ts:76`) compresses `cleanedUrlParams.toString()`, which is pure percent-encoded ASCII, so no live bug; the issue is latent fragility only.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — with the current single call site the round trip is lossless (verified: apply enough filters with non-ASCII values, e.g. `Zürich`, to push the URL over 2000 chars so it compresses into the gzipped param, then reload the URL; filters restore correctly). Confirm in code: `web-common/src/features/dashboards/url-state/compression.ts:11,22` — `TextEncoder` on encode vs `String.fromCharCode` on decode; a future caller passing raw non-ASCII would get mojibake.

### frontend-dashboard-state-03 [medium] Boolean `updating` lock drops a competing navigation/state-change made during the in-flight `resolveTimeRanges` await
- **Category:** concurrency  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:208-300`
- **Problem:** `handleURLChange` awaits `resolveTimeRanges` (a network call) while holding `updating = true`. `afterNavigate` fires `handleURLChange` fire-and-forget (`DashboardStateManager.svelte:95`), and store mutations fire `gotoNewState` synchronously. A second navigation or state change that arrives during the await returns immediately at line 214 (or line 311 for `gotoNewState`) and is never merged/synced. When the first call resumes it recomputes the redirect from its own (now stale) state and may `goto` (line 296), so the second change can be silently dropped, leaving URL and store desynced.
- **Failure scenario:** A user quickly changes two time-range presets (or a filter then a time range) while the first range is still resolving on the server. The second change's URL/state sync is dropped; the dashboard can snap back to the first selection until the next interaction.
- **Recommendation:** Replace the single boolean with a monotonically increasing "generation" token captured before the await and re-checked after; abort (don't apply, don't `goto`) if a newer generation started. Alternatively serialize navigations through a queue.

**Verification:** PLAUSIBLE — `handleURLChange` holds `updating = true` (`DashboardStateSync.ts:215`) across the `await resolveTimeRanges` at lines 239-255 (only when the metrics view has a `timeDimension`), and any call arriving during the await returns at line 214 (`gotoNewState` at line 311) without being queued or retried. `DashboardStateManager.svelte:95` calls `handleURLChange` fire-and-forget from `afterNavigate`. The drop is real but requires a second change to land inside the in-flight network window, so it is timing-conditional.

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard whose metrics view has a time dimension (any standard project) in Rill Developer or Cloud.
2. In browser devtools, set network throttling to "Slow 3G" to widen the `resolveTimeRanges` window.
3. Quickly change the time range preset twice in succession (e.g. "Last 7 days" then immediately "Last 30 days"), or apply a dimension filter and immediately change the time range.
4. Expected: the URL and dashboard settle on the second selection. Actual (when hit): the second change's URL/session-storage sync is dropped; the URL can snap back to reflect the first selection until the next interaction.
5. If flaky, confirm in code: `DashboardStateSync.ts:214,239-255,311` — boolean lock held across an await with competing calls returning silently.

### frontend-dashboard-state-04 [low] `handleExploreInit` can double-initialize because the `initialized` guard precedes an `await`
- **Category:** concurrency  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:129-201`
- **Problem:** The re-entrancy guard `if (this.initialized) return;` (line 131) runs before `await resolveTimeRanges` (line 150), and `this.initialized = true` is only set after the await (lines 190/201). The `initExploreState` subscription (line 70) can emit more than once. If a second emission arrives while the first `handleExploreInit` is suspended on the await, both invocations pass the guard and both run `metricsExplorerStore.init` + `goto`.
- **Failure scenario:** With a time dimension present (so `resolveTimeRanges` does a network round trip) a spec refetch produces a fresh `initExploreState` value during the await; the store is `init`-ed twice and two `replaceState` navigations occur. Impact is bounded because init runs before the dashboard is interactive (the `initialized` guard gates `handleURLChange`/`gotoNewState`), but the redundant re-init and double navigation can clobber any state applied in the interim.
- **Recommendation:** Set a synchronous in-flight flag before the await (e.g. `this.initializing = true`), or set `this.initialized = true` up front and revert on failure, so only one initialization can proceed.

**Verification:** PLAUSIBLE — the guard `if (this.initialized) return;` is at `DashboardStateSync.ts:131`, the `await resolveTimeRanges` at lines 145-161 (gated on `metricsViewSpec.timeDimension`), and `this.initialized = true` is only set at lines 190 and 201, after the await. The subscription at lines 70-82 fires on every `initExploreState` emission, so a second emission during the await passes the guard and runs `metricsExplorerStore.init` + `goto` a second time. Real race, but confined to the init window before the dashboard is interactive.

**Manual Reproduction Steps (for QA):**
Not reliably user-reproducible — the window is the initial-load network call and the double-init is mostly invisible (redundant `replaceState`). A possible attempt: with "Slow 3G" throttling, load an explore dashboard with a time dimension in Rill Developer and save a YAML edit to its metrics view during the initial spinner, then check for state flicker/reset once loaded. Otherwise confirm in code: `DashboardStateSync.ts:131,145-161,190,201` — guard read before the await, flag set after it.

### frontend-dashboard-state-05 [low] `memoizeMetricsStore` cache key concatenates names without a separator (collisions) and never evicts
- **Category:** correctness  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/state-managers/memoize-metrics-store.ts:15`
- **Problem:** The cache key is `metricsName + exploreName` with no delimiter, so `("sales","report")` and `("sale","sreport")` both map to `"salesreport"` and return the same memoized store. This backs `useTimeControlStore`, `useTimeSeriesDataStore`, `usePivotForExplore`, and `useDimensionTableData`. The module-level `Map` is also never cleared, so every distinct dashboard visited in a session retains its derived stores.
- **Failure scenario:** Two metrics-view/explore pairs whose concatenations collide would render the wrong time-control/TDD/pivot store for one of them (low probability, requires adversarial-ish naming). More commonly, long-lived sessions accumulate never-evicted store entries.
- **Recommendation:** Use a delimiter that cannot appear in resource names (e.g. `` `${metricsName}\n${exploreName}` ``) and consider bounding/evicting the cache when the context tears down.

**Verification:** PLAUSIBLE — `memoize-metrics-store.ts:15` is `const key = metricsName + exploreName;` with no delimiter, and the closure-level `Map` (line 10) has no eviction. Both facts confirmed; an actual collision requires two name pairs whose concatenations coincide, which is unlikely with normal naming.

**Manual Reproduction Steps (for QA):**
1. In a Rill Developer project, create two metrics views named `sales` and `sale`, and matching explores named `report` (on `sales`) and `sreport` (on `sale`), with clearly different time settings (e.g. different `smallest_time_grain` or data ranges).
2. Open the `report` explore dashboard, then in the same browser tab/session open the `sreport` explore dashboard.
3. Expected: each dashboard shows its own time controls / time-dimension-detail / pivot behavior. Actual: the second dashboard reuses the first dashboard's memoized time-control/TDD/pivot store (both keys concatenate to `salesreport`), so its time controls reflect the wrong metrics view.
4. The no-eviction aspect is not directly visible; confirm in code: `web-common/src/features/dashboards/state-managers/memoize-metrics-store.ts:10,15`.

### frontend-dashboard-state-06 [low] Proto time serialization stores milliseconds in the `seconds` field
- **Category:** tech-debt  **Confidence:** high  **Location:** `web-common/src/features/dashboards/proto-state/toProto.ts:218-221`, `web-common/src/features/dashboards/proto-state/fromProto.ts:484-486`
- **Problem:** `toTimeProto` writes `seconds: BigInt(date.getTime())` (milliseconds) and `fromTimeProto` reads `new Date(Number(timestamp.seconds))` (treating the value as milliseconds). The two are internally consistent so the frontend round-trips correctly, but the proto `Timestamp.seconds` field is populated with a millisecond value.
- **Failure scenario:** No current impact: a search of `runtime/`, `admin/`, and `cli/` found no Go consumer that reads these bookmark/legacy-URL proto timestamps as epoch seconds. The risk is latent — a future non-frontend consumer that reads `seconds` as real epoch seconds would compute a timestamp ~1000x in the future.
- **Recommendation:** Either store real seconds (`BigInt(Math.floor(date.getTime() / 1000))`) plus `nanos` and migrate readers, or add a prominent comment documenting the intentional misuse.

**Verification:** PLAUSIBLE — `toProto.ts:218-222` (`toTimeProto`) writes `seconds: BigInt(date.getTime())` (milliseconds) and `fromProto.ts:484-486` (`fromTimeProto`) reads `new Date(Number(timestamp.seconds))` as milliseconds. Internally consistent misuse of the `Timestamp.seconds` field; latent only, no current cross-language consumer affected.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — the frontend writes and reads the field symmetrically, so bookmarks and legacy proto URLs round-trip correctly in the UI. Confirm in code: `web-common/src/features/dashboards/proto-state/toProto.ts:218-222` vs `web-common/src/features/dashboards/proto-state/fromProto.ts:484-486`; any new consumer treating `seconds` as epoch seconds would be ~1000x off.

### frontend-dashboard-state-07 [low] Session-storage TDD restore casts a possibly-`null` measure param to `string`
- **Category:** correctness  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/state-managers/loaders/explore-web-view-store.ts:150-152`
- **Problem:** `storedExploreState.tdd.expandedMeasureName = searchParams.get(ExploreStateURLParams.ExpandedMeasure) as string;` assigns `null` (cast to `string`) when the `measure` param is absent, rather than falling back to `""` or the stored value. This path is reachable: `shouldSkipSessionStorage` allows through a URL of `?view=ttd` (single `view` param, no `measure`), and `searchParams.get(ExpandedMeasure)` then returns `null`.
- **Failure scenario:** Landing on the TDD view URL without a `measure` param sets `expandedMeasureName` to `null`; downstream code that assumes a string (measure lookups / set membership) can misbehave or blank the TDD table.
- **Recommendation:** Default to `""`: `searchParams.get(...) ?? ""`.

**Verification:** CONFIRMED — `explore-web-view-store.ts:150-152` assigns `searchParams.get(ExploreStateURLParams.ExpandedMeasure) as string` to `expandedMeasureName`; with no `measure` param this is `null`. Reachability verified: `shouldSkipSessionStorage` (`explore-web-view-store.ts:187-206`) does not skip a URL whose only param is `view` (size 1 with `view` set fails all three skip conditions), so `?view=ttd` flows into the TDD branch at line 142.

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard, click "Expand" on a measure to enter the time-dimension-detail (TDD) view, so TDD state is saved in session storage.
2. Go back to the main explore view, then manually edit the browser URL to exactly `?view=ttd` (delete every other query param, including `measure`) and press Enter.
3. Expected: TDD view opens with a sensible default (e.g. previously expanded measure or graceful fallback). Actual: the restored state has `expandedMeasureName = null`; the TDD table can render blank or misbehave since downstream code assumes a string.
4. If the UI happens to recover, confirm in code: `web-common/src/features/dashboards/state-managers/loaders/explore-web-view-store.ts:150-152` — `null` cast to `string`.

### frontend-dashboard-state-08 [low] `cleanUrlParams` time-grain special case falls through into the generic cleaning branch
- **Category:** readability  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/url-state/clean-url-params.ts:27-49`
- **Problem:** The `if (key === TimeGrain)` block (lines 27-34) sets the grain when the time range differs from default but does not `return`/`continue`, so execution falls through to the generic `hasDefaultAndIsUnmodified` logic below. All four cases were traced and the two blocks happen to combine to the correct output, but the control flow is hard to reason about and fragile to future edits.
- **Failure scenario:** No current incorrect output. A maintainer changing the generic branch could unknowingly re-clean the grain the special case intended to preserve.
- **Recommendation:** Make the branches mutually exclusive: handle the `TimeGrain` key fully and `continue` inside its block, leaving the generic path for all other keys.

**Verification:** CONFIRMED — `clean-url-params.ts:27-34`: the `if (key === ExploreStateURLParams.TimeGrain)` block calls `cleanedParams.set(key, value)` when the time range differs from default and then falls through (no `return` inside the `forEach` callback) into the generic `hasNoDefaultAndValueIsEmpty` / `hasDefaultAndIsUnmodified` logic at lines 36-49. Traced the four range/grain combinations; the combined result is correct today (the special-case `set` survives because the generic branch either re-sets the same value or returns early after the set), so this is a readability/fragility issue only.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — current output is correct for all cases (spot-check: change the time range away from default with the default grain; the `grain` param is correctly kept in the URL). Confirm in code: `web-common/src/features/dashboards/url-state/clean-url-params.ts:27-34` — the `TimeGrain` branch lacks a `return`, so execution continues into the generic cleaning logic below.

### frontend-dashboard-state-09 [low] TDD "Total" row body sums current + comparison while its header shows the current-period total
- **Category:** correctness  **Confidence:** low  **Location:** `web-common/src/features/dashboards/time-dimension-details/time-dimension-data-store.ts:227-277`
- **Problem:** In `prepareTimeData` with time comparison enabled, the "Total" row header (lines 228-236) uses `total` (the current-period grand total, from `timeSeries.total[measureName]`) and a sparkline over `v[measureName]` (current only), but the "Total" row body cells (lines 267-277) are `current + comparison` per bucket. Header and body for the same row therefore represent different quantities, and the body cells do not sum to the header value.
- **Failure scenario:** In TDD with time comparison on, the first table row's per-bucket cells show current+previous summed values that don't reconcile with the row's displayed total. Whether this is a defect depends on the intended meaning of the "Total" row in comparison mode, which is not documented.
- **Recommendation:** Confirm the intended semantics; if the row is meant to show the current period (matching the non-comparison branch on line 323 and the header), change the body to `sanitizeMeasure(v[measureName])`.

**Verification:** CONFIRMED — in `prepareTimeData` (`time-dimension-data-store.ts:207`), the "Total" row header at lines 228-236 shows `total` (passed from `timeSeries?.total[measureName]` at line 400-401, the current-period grand total) with a current-only sparkline, while the "Total" row body cells at lines 267-277 compute `sanitizeMeasure(v[measureName]) + sanitizeMeasure(v['comparison.'+measureName])` per bucket. Header and body of the same row represent different quantities; whether this is intended is undocumented, hence the low confidence on severity stands.

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard with a time dimension and expand a measure into the time-dimension-detail (TDD) view.
2. Enable time comparison (comparison toggle in the time controls, e.g. "Previous period"), and make sure no comparison dimension is selected.
3. Look at the first table row labeled "Total": note the total value in the second fixed column, then manually sum a few of its per-bucket cells and compare against the "Current"/"Previous" rows below.
4. Expected: each "Total" cell equals the "Current" row cell (matching the header total), or the header equals current+previous. Actual: each "Total" body cell equals Current + Previous for that bucket, while the row's header total is the current period only, so the cells do not reconcile with the header.

## Product Improvements

### frontend-dashboard-state-10 [low] Comparison time range is not server-resolved during init / URL sync
- **Category:** product-improvement  **Confidence:** low  **Location:** `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:148-160,241-254`
- **Problem:** The `selectedComparisonTimeRange` argument to `resolveTimeRanges` is commented out in both `handleExploreInit` and `handleURLChange`, so only the primary range gets its `start`/`end` resolved via the network. This appears deliberate (comparison ranges are typically derived client-side relative to the primary range), but for comparison ranges that depend on server resolution the values are left unresolved at these entry points.
- **Recommendation:** Either resolve the comparison range alongside the primary range or document why client-side derivation is sufficient; add a test asserting the comparison `start`/`end` are populated for a rill-time comparison on a shared link.

**Verification:** CONFIRMED — the commented-out `selectedComparisonTimeRange` arguments are present in both `handleExploreInit` (`DashboardStateSync.ts:149,155`) and `handleURLChange` (`DashboardStateSync.ts:243,249`); only the primary range is passed to `resolveTimeRanges`. Whether this causes a user-visible defect depends on comparison ranges that require server resolution, so this remains a confirm/document product note rather than a proven bug.

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard, select a custom/rill-time time range and enable a time comparison, then copy the URL.
2. Open the copied URL in a fresh tab (new session, so init runs from the URL).
3. Expected: the comparison period renders immediately with correct start/end. Actual (if affected): comparison data may be missing or use unresolved bounds on first paint for comparison ranges that need server resolution.
4. If no visible difference, confirm in code: `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:149,155,243,249` — comparison range resolution commented out.

### frontend-dashboard-state-11 [low] Invalid URL-state errors are silently dropped in the load path
- **Category:** product-improvement  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/state-managers/loaders/DashboardStateDataLoader.ts:329-335`
- **Problem:** `getMergedExploreState` calls `convertURLSearchParamsToExploreState` and destructures only `partialExploreState`, discarding the rich `errors` array (missing measures/dimensions, bad time grain, etc.) that the converter already produces. A user opening a stale/shared link that references removed fields gets a silently degraded dashboard with no feedback.
- **Recommendation:** Surface a non-blocking toast/banner summarizing dropped URL parameters (the error strings are already produced by `error-message-helpers`), gated by the existing `ignore_errors` param.

**Verification:** CONFIRMED — `DashboardStateDataLoader.ts:329-335`: `const { partialExploreState: partialExploreStateFromUrl } = convertURLSearchParamsToExploreState(...)` destructures only the state; the converter's `errors` array is discarded and nothing in `getMergedExploreState` surfaces it.

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard, apply a dimension filter and select a couple of measures, and copy the URL.
2. Edit the URL by hand to reference a non-existent field (e.g. change the `measures` param value to `bogus_measure`, or the filter expression to use a dimension name that does not exist) and load it in a fresh tab.
3. Expected: some feedback (toast/banner) that parts of the link were invalid and dropped. Actual: the dashboard loads silently with the invalid parameters ignored; the user gets no indication the shared view is degraded.

### frontend-dashboard-state-12 [low] `handleURLChange` releases the lock before its own `replaceState` `goto`, causing a redundant re-entrant sync pass
- **Category:** concurrency  **Confidence:** medium  **Location:** `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:288-299`
- **Problem:** Unlike `handleExploreInit` and `gotoNewState`, which hold `updating = true` across their `goto`, `handleURLChange` sets `this.updating = false` at line 288 and only then calls `goto(redirectUrl, { replaceState: true })` at line 296. That navigation triggers `afterNavigate` -> `handleURLChange` again with `updating` already released, so the whole body (including another `resolveTimeRanges` network call and `mergePartialExplorerEntity`) runs a second, redundant time before terminating at the search-equality check. This is both an inconsistency with the sibling methods and a small hot-path inefficiency.
- **Recommendation:** Hold the lock across the `goto` (release in a `finally` after it, consistent with the fix for finding 01), so the redirect's own `afterNavigate` is skipped like it is for the other two methods.

**Verification:** CONFIRMED — `DashboardStateSync.ts:288` sets `this.updating = false` before the `goto(redirectUrl, { replaceState: true })` at lines 296-299. The redirect's `afterNavigate` (`DashboardStateManager.svelte:92-95`) re-enters `handleURLChange` with the lock already released, so the full body (including another `resolveTimeRanges` call when a time dimension exists) runs a second time before terminating at the `redirectUrl.search === pageState.url.search` check. By contrast `handleExploreInit` (lines 197-201) and `gotoNewState` (lines 351-352) hold their flag across `goto`.

**Manual Reproduction Steps (for QA):**
1. Open an explore dashboard with a time dimension, interact with it (set filters/time range) so session storage holds view state, then navigate to another page in the app.
2. Open browser devtools Network tab and filter for the time-range resolution request (`metrics-view` time range query).
3. Navigate back to the dashboard using a link with a partial URL (e.g. paste the dashboard URL stripped to just `?view=explore`), so `handleURLChange` augments state from session storage and issues a redirect.
4. Expected: one sync pass / one time-range resolution per navigation. Actual: the redirect triggers a second, redundant `handleURLChange` pass (a duplicate time-range resolution request appears).
5. Alternatively confirm in code: `web-common/src/features/dashboards/state-managers/loaders/DashboardStateSync.ts:288,296` — lock released before the redirect `goto`.

## New Feature Ideas
- **URL <-> state property/fuzz tests (S/M):** Generate random `ExploreState`, serialize to URL params, parse back, and assert equivalence (including compression, non-ASCII filters, all web views). Would guard the many round-trip edge cases (and lock in the compression behavior verified in finding 02). High value, low risk.
- **Deterministic sync engine with generation tokens (M):** Replace the `updating` boolean in `DashboardStateSync` with a small state machine / generation counter that is race-safe and always releases its lock. Directly addresses findings 01, 03, 04, 12.
- **"Copy clean link" / "Reset to default view" affordance (S):** A button that emits the minimal cleaned (and, if needed, compressed) URL and one that resets to the YAML/Rill default preset. Users routinely share bloated URLs today; the URL-too-long banner is the only current signal.
- **Shorter shareable state via server-side state tokens (L):** Persist large explore states server-side and reference them by a short id in the URL, bypassing the 16KB header limit and the client-side compression path entirely.
- **Human-readable time-range chips from rill-time (M):** Surface the parsed `RillTime` (already available in `time-controls`) as an editable, validated expression with inline error messages, reducing reliance on opaque preset names in the URL.
- **State-diff debugger panel (S):** A dev-only overlay showing the cascading-merge inputs (session/URL/bookmark/YAML/Rill-default) and the winning source per key. Would make the merge order in `DashboardStateDataLoader` far easier to debug.

## Notes & Missing Test Coverage
- **No tests for the `DashboardStateSync` lock lifecycle** — findings 01/03/04/12 involve `updating`/`initialized` transitions. `DashboardStateManager.spec.ts` exists but the guard-clause leak and the re-entrancy races are not exercised. Note that `resolveTimeRanges` is skipped under `import.meta.env.VITEST` (DashboardStateSync.ts:145,239), so the await-related races (03/04) can't reproduce under the current harness without changes — **but finding 01's leak is pre-await and would reproduce today.**
- **No tests for `compression.ts`** — add round-trip tests with ASCII, non-ASCII, and large inputs to lock in behavior (see finding 02).
- **`resolveTimeRanges` mutates its input `DashboardTimeControls` objects in place** (`rill-time-ranges.ts:26,58-59`); currently safe because callers pass freshly-derived objects, but worth a comment or a defensive clone if these objects are ever shared.
- **Not fully verified:** the exact intended semantics of the TDD "Total" row (finding 09) — flagged low-confidence pending design confirmation. The `getComparisonInterval` path in `time-control-store.ts` was checked and is safe (`Interval<true>` guarantees non-null `start`/`end`), so no finding was raised there.

## Verification Notes
Every finding was re-checked against the code; the following changes were made:

- **Finding 02 (was high / correctness / high) — core claim refuted, downgraded to low / readability / low.** The reviewer missed that the only caller compresses `cleanedUrlParams.toString()` (`convert-partial-explore-state-to-url-params.ts:76`). `URLSearchParams.toString()` percent-encodes all non-ASCII into pure ASCII, so every byte handed to `TextEncoder`/`gzip` is < 128 and `String.fromCharCode` round-trips losslessly; `new URLSearchParams(...)` then re-decodes the escapes. There is **no** non-ASCII corruption. The `RangeError` is also unreachable: compression triggers above 2000 chars and the URL is capped near 16 KB (`url-length-limits.ts`), far below the spread argument-count limit. Rewrote it as a minor robustness/readability nit (prefer `TextDecoder`).
- **Finding 01 — confirmed, kept at high.** Corrected the location/emphasis: the two leaking returns (lines 225, 235) are **synchronous and precede** the `resolveTimeRanges` await, so the leak is a plain early-return bug (not an await race) and is reproducible even under the VITEST harness. Trimmed the recommendation (removing the `handleExploreInit`/`initialized` `finally` suggestion, which conflated it with finding 04).
- **Finding 04 — kept, severity medium -> low.** The double-init is real, but init runs before the dashboard is interactive, bounding the impact.
- **Finding 05 — kept, severity medium -> low.** The missing separator is a genuine latent collision but requires adversarial naming; the no-eviction leak is minor.
- **Finding 06 — kept, category correctness -> tech-debt.** Confirmed via a Go search that no `runtime/`/`admin/`/`cli/` consumer reads these proto timestamps, so impact is latent only.
- **Finding 07 — kept at low.** Confirmed the `null` assignment is reachable through a `?view=ttd` URL (verified against `shouldSkipSessionStorage`).
- **Finding 08 — kept at low.** Traced all four grain/time-range cases; the fall-through produces correct output today, so this is readability/fragility only (not correctness). Tightened the line range.
- **Finding 09 — kept at low / low.** The header-vs-body inconsistency is real and verifiable (`total` is the current-period grand total; the body sums current + comparison), but the intended semantics are undocumented, so confidence stays low.
- **Finding 10 — kept, severity medium -> low, confidence -> low.** The commented-out comparison resolution is deliberate; reframed as a "confirm/document" product note rather than a defect.
- **Findings 03 and 11 — kept** with tightened descriptions.
- **Added finding 12 (new, low):** `handleURLChange` releases the lock before its own `replaceState` `goto`, triggering a redundant re-entrant sync pass and diverging from the two sibling methods that hold the lock across `goto`.

No findings were removed outright (0 removed); one high-severity finding (02) had its central claim refuted and was downgraded to low, several severities/categories were recalibrated, and one finding was added.
