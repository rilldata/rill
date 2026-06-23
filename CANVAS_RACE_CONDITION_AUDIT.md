# Canvas Dashboard — Race Condition Audit

**Scope:** `web-common/src/features/canvas/` (Svelte 4, TanStack Query, SvelteKit).
**Method:** Four independent sub-agents audited distinct concern areas — async/promise ordering, Svelte store/subscription lifecycle, component lifecycle & resource management, and URL/filter/time state sync. The Go backend (`runtime/canvas`) was checked and contains **no concurrency primitives**, confirming the nondeterminism is frontend-only.

Multiple agents converged independently on the same root causes (noted below) — a strong signal these are real.

---

## Root causes (the two structural defects driving most symptoms)

These two patterns explain the bulk of the user-reported nondeterministic behavior:

### A. The URL is a bidirectional source of truth with no reentrancy guard
Every state mutation calls `goto()`; the resulting `$page.url` change re-runs the reactive `onUrlChange` block, which writes the URL *back* into state. Because `onUrlChange` is `async` and unguarded, concurrent navigations interleave and a **stale invocation can write last, clobbering newer state**. Compounded by handlers that read `window.location.search` (which only updates *after* a `goto` commits) instead of the in-memory source of truth.

### B. Subscriptions opened in constructors are never torn down
`CanvasEntity` and every `BaseCanvasComponent` open spec subscriptions in their constructors and discard the unsubscriber. The "reset to defaults" path (`removeCanvasStore`) and `CanvasEntity.unsubscribe()` are **no-ops**. Stale entity/component instances keep reacting to spec emissions and **race the live instance** over URL and YAML file writes.

---

## Findings by severity

### 🔴 HIGH

#### H1. `onUrlChange` has no reentrancy guard — concurrent navigations clobber state
**Location:** [stores/canvas-entity.ts:500-521](web-common/src/features/canvas/stores/canvas-entity.ts#L500-L521), triggered from [CanvasInitialization.svelte:80-84](web-common/src/features/canvas/CanvasInitialization.svelte#L80-L84)
`onUrlChange` awaits `handleCanvasRedirect` (which awaits a bookmark `fetchQuery` + `goto`), then writes `filterManager`/`searchParams`/`timeManager` using the **captured (now stale) `searchParams`**. A second URL change starts a second concurrent run with no guard; the older run can write last, reverting filters/time to a previous URL and firing queries against the wrong state.
**Fix:** Capture a per-call sequence token; bail before each post-await write if a newer call started. Or serialize through a queue. At minimum re-read the URL after each `await`.

#### H2. Stale write ordering inside `onUrlChange` — snapshot saved before time state applied
**Location:** [stores/canvas-entity.ts:517-521](web-common/src/features/canvas/stores/canvas-entity.ts#L517-L521)
The four sequential writes aren't atomic w.r.t. synchronously-recomputing derived stores. `saveSnapshot(searchParams.toString())` runs **before** `timeManager.state.onUrlChange(...)`, persisting a pre-time-state snapshot into the module-level map; `viewingDefaultsStore` recomputes mid-sequence on a partially-applied state.
**Fix:** Apply all state first, `saveSnapshot` last. Better, batch the writes so derived stores recompute once.

#### H3. `interval` derived store: stale `Promise.all` resolves last and wins
**Location:** [stores/time-state.ts:158-208](web-common/src/features/canvas/stores/time-state.ts#L158-L208)
The `interval` derived fires `Promise.all(deriveInterval...)` on every input change and writes the result in `.then()`. Svelte `derived` does **not** cancel the prior run. Two rapid time-range changes (A then B) where B resolves first → A's `.then()` overwrites the store with the **old** range's interval. Dashboard renders data for a range that no longer matches the URL.
**Fix:** Capture a generation token before the await; in `.then()` `return` if a newer run started.

#### H4. `applyFiltersToUrl` reads stale `window.location.search` — concurrent filter actions clobber each other
**Location:** [stores/filter-manager.ts:865-897](web-common/src/features/canvas/stores/filter-manager.ts#L865-L897), `clearAllFilters` [:846-863](web-common/src/features/canvas/stores/filter-manager.ts#L846-L863)
Filter actions are `async` and fire-and-forget from click handlers. Each snapshots `new URLSearchParams(window.location.search)` then `await goto(...)`. Two changes before the first `goto` commits → the second reads the URL **without** the first's change and overwrites it. Last-write-wins on shared URL state; the first filter edit is silently lost. (Same root cause: [canvas-entity.ts:132](web-common/src/features/canvas/stores/canvas-entity.ts#L132).)
**Fix:** Build params from the in-memory source of truth (`get(page).url.searchParams` or the entity's `searchParams` store), not `window.location.search`. Or serialize the actions.

#### H5. `CanvasEntity` spec subscription leaks; recreated instances keep firing on stale state
**Location:** [stores/canvas-entity.ts:181-189](web-common/src/features/canvas/stores/canvas-entity.ts#L181-L189), [:524-526](web-common/src/features/canvas/stores/canvas-entity.ts#L524-L526); [state-managers/state-managers.ts:47-53](web-common/src/features/canvas/state-managers/state-managers.ts#L47-L53)
The constructor's `specStore.subscribe` unsubscriber is never called — `unsubscribe()` is commented out, and `removeCanvasStore` only deletes the registry entry. On YAML-save "reset to defaults," a fresh entity is built while the old one's callback still runs `processSpec(data)` on the next emission — two entities racing to drive the URL/`fileArtifact`. Leak compounds across every save+revisit cycle.
**Fix:** Make `removeCanvasStore` call `store.canvasEntity.unsubscriber()` before deleting; restore the body of `unsubscribe()`.

#### H6. `BaseCanvasComponent` spec subscriptions never cleaned up
**Location:** [components/BaseCanvasComponent.ts:140-148](web-common/src/features/canvas/components/BaseCanvasComponent.ts#L140-L148)
Each component constructor subscribes to `specStore` (pushing into `localFilters`/`localTimeControls`) and discards the unsubscriber. Components are replaced in `processRows` without disposal, so orphaned subscriptions keep mutating filter/time state of deleted widgets — which can be shared via the parent's `filterManager`/`timeManager`.
**Fix:** Store the unsubscriber, add `destroy()`, and call it in `processRows`/`removeComponent` before replacing or deleting a component.

#### H7. Module-level `lastVisitedState` shared across instances/tabs, keyed only by name
**Location:** [stores/canvas-entity.ts:52](web-common/src/features/canvas/stores/canvas-entity.ts#L52), used [:542](web-common/src/features/canvas/stores/canvas-entity.ts#L542) [:595](web-common/src/features/canvas/stores/canvas-entity.ts#L595) [:602-604](web-common/src/features/canvas/stores/canvas-entity.ts#L602-L604)
The registry keys by `instanceId::canvasName`, but `lastVisitedState` keys only by canvas **name**. Two instances of the same canvas (embed + main app, two runtime instances) share one snapshot: instance A saves its filter/time URL; instance B loading with empty params gets redirected to A's state — cross-instance contamination. Never cleared on unmount.
**Fix:** Key by `instanceId::canvasName` (mirror `makeCanvasId`), or store the snapshot on the instance and clear it in a real teardown.

#### H8. Keyed-each by positional index reuses wrong component instance on reorder
**Location:** [EditableCanvasRow.svelte:187](web-common/src/features/canvas/EditableCanvasRow.svelte#L187), [StaticCanvasRow.svelte:36](web-common/src/features/canvas/StaticCanvasRow.svelte#L36)
`{#each itemIds as id, columnIndex (columnIndex)}` keys by slot position, not by component id. On drag-reorder/insert/delete, Svelte reuses the instance at that position and swaps only the `component` prop — so a moved widget renders into another widget's DOM node, and `bind:`/selection state attaches to the wrong slot. Worse with optimistic reorders landing before server reconciliation.
**Fix:** Key by stable `id`: `{#each itemIds as id, columnIndex (id)}` (guard empty-string ids for unresolved items).

---

### 🟡 MEDIUM

#### M1. Optimistic reorder via `processRows` races server reconcile; positional IDs drop component state
**Location:** [stores/canvas-entity.ts:643-707](web-common/src/features/canvas/stores/canvas-entity.ts#L643-L707), `generateId` [:709](web-common/src/features/canvas/stores/canvas-entity.ts#L709)
Component identity is purely positional (`generateId(row, column)`). A reorder renumbers components, so a moved widget arrives under a new id, `existingClass` is undefined, and a **brand-new** `BaseCanvasComponent` is constructed — losing `visible`, local filter/time state, and in-flight inspector edits. The code comments already flag this ("Once we have stable IDs, this can be simplified").
**Fix:** Stable component IDs (root fix). Short-term: migrate local state when replacing an instance at the same path.

#### M2. `IntersectionObserver` created in `onMount` is never disconnected
**Location:** [CanvasComponent.svelte:16-32](web-common/src/features/canvas/CanvasComponent.svelte#L16-L32)
No `onDestroy` → `observer.disconnect()`. With frequent remounts (H8 + drag ghosts), observers leak; a callback firing after teardown calls `component.visible.set(true)` on a dead component. (Contrast the correct cleanup in [ComponentHeader.svelte:34-35](web-common/src/features/canvas/ComponentHeader.svelte#L34-L35).)
**Fix:** `onDestroy(() => observer?.disconnect())`.

#### M3. Non-atomic YAML read-modify-write across `saveDefaultFilters` and inspector edits
**Location:** [stores/canvas-entity.ts:363-467](web-common/src/features/canvas/stores/canvas-entity.ts#L363-L467) vs [components/BaseCanvasComponent.ts:348-358](web-common/src/features/canvas/components/BaseCanvasComponent.ts#L348-L358)
Both paths do read-modify-write against `editorContent` via `updateEditorContent`, with awaits interleaving and no locking/version check. A late `updateYAML` write can clobber the filter write or vice-versa. (`saveDefaultFilters` does correctly read `parsedContent` *after* its await per the in-code comment — the hazard is the **uncoordinated second writer**.)
**Fix:** Funnel all YAML mutations through a single serialized writer that re-reads current content, mutates, and writes synchronously — never holding a parsed document across an await.

#### M4. `saveDefaultFilters` hardcoded `setTimeout(100)` + subscribe-then-unsub navigation
**Location:** [stores/canvas-entity.ts:363-369](web-common/src/features/canvas/stores/canvas-entity.ts#L363-L369), [:457-466](web-common/src/features/canvas/stores/canvas-entity.ts#L457-L466)
The 100ms wait is a guess at filter-propagation time; slower propagation (or continued typing) → stale filters persisted to YAML. The trailing navigation `goto`s on the *next* `defaultUrlParamsStore` emission — but that store only updates if the value actually changed (re-saving identical defaults → navigation never fires, `unsub` leaks), and an unrelated emission triggers a wrong-params `goto`. No guard against two overlapping saves.
**Fix:** Drive off an explicit "filters committed" signal (await the pending URL sync); compute the target params directly and `goto` once; disable/guard re-entry.

#### M5. `defaultUrlParamsStore` equality is `URLSearchParams.toString()` — order-sensitive, causes flapping
**Location:** [stores/canvas-entity.ts:260-266](web-common/src/features/canvas/stores/canvas-entity.ts#L260-L266)
`toString()` preserves insertion order, so the same logical defaults with keys in a different order compare unequal and re-`set` the store, re-triggering `viewingDefaultsStore` and the M4 subscription. (`viewingDefaultsStore`'s own body at [:222-239](web-common/src/features/canvas/stores/canvas-entity.ts#L222-L239) is correctly order-independent — only the upstream populate is buggy.)
**Fix:** Canonicalize (sort keys) before comparing, or compare entry-by-entry.

#### M6. Time state applied field-by-field — transient mixed (range, zone) fires queries
**Location:** [stores/time-state.ts:261-280](web-common/src/features/canvas/stores/time-state.ts#L261-L280)
`onUrlChange` sets `urlRangeStore`/`urlGrainStore`/`urlTimeZoneStore`/`urlComparisonRangeStore` in sequence; each recomputes the async `interval` derived. An interval can be computed from the **new range but old timezone** and resolve out of order. Comparison mode (`showTimeComparisonStore` vs `comparisonRangeStore`) has the same split-write hazard.
**Fix:** Apply parsed params in one batched update; guard the async `interval` against out-of-order resolution (generation counter).

#### M7. `firstTimeLoad` skip-first-emission flag is fragile to emission ordering
**Location:** [stores/canvas-entity.ts:104-107](web-common/src/features/canvas/stores/canvas-entity.ts#L104-L107), [:181-189](web-common/src/features/canvas/stores/canvas-entity.ts#L181-L189)
The constructor's `this.spec` comes from a *different* query (`createQueryServiceResolveCanvas`) than `this.specStore` (`useCanvas`). If they're not in sync at construction (cache miss, `enabled` gating), the first *real* emission is silently swallowed by the flag and a genuinely newer spec is dropped.
**Fix:** Dedupe by comparing incoming spec against last-processed spec instead of a boolean; or don't pre-call `processSpec` in the constructor.

#### M8. `replaceState` redirect can clobber a concurrent `pushState` filter write
**Location:** [stores/canvas-entity.ts:541-599](web-common/src/features/canvas/stores/canvas-entity.ts#L541-L599)
`handleCanvasRedirect` awaits a bookmark `fetchQuery` then `goto(..., {replaceState:true})`. A user filter write (push) landing during that window can be replaced by the redirect resolving last. Same root cause as H1, specific to replace-vs-push ordering.
**Fix:** Guard the redirect with H1's in-flight token; abort if the URL gained params during the fetch.

---

### 🟢 LOW

#### L1. Non-reactive `componentsStore` refresh is conditional — rows can stay stale
**Location:** [stores/canvas-entity.ts:643-707](web-common/src/features/canvas/stores/canvas-entity.ts#L643-L707) (esp. [:701-704](web-common/src/features/canvas/stores/canvas-entity.ts#L701-L704))
`_rows.refresh()` runs only when `(!didUpdateRowCount && createdNewComponent) || isFirstLoad`. The common in-place-update case and pure deletions flag neither condition → `refresh()` skipped. Leaf components subscribed to their own stores still update, so impact is limited today, but it's exactly the "did we remember to refresh?" trap.
**Fix:** Make `componentsStore` reactive, or refresh unconditionally at the end of `processRows`, or track a single `dirty` flag.

#### L2. `viewingDefaultsStore` transient flicker from non-atomic multi-dependency updates
**Location:** [stores/canvas-entity.ts:191-242](web-common/src/features/canvas/stores/canvas-entity.ts#L191-L242)
Six dependencies update from independent code paths (`onUrlChange` vs `processSpec`); during settle it emits transient values. Only consumer today is a button label (cosmetic flicker). Becomes a real bug if ever used to gate a write/navigation.
**Fix:** Acceptable as a label; if reused for side effects, settle/debounce.

---

## Recommended fix order

1. **H5 + H6** (subscription teardown) — single structural fix that removes a whole class of stale-instance races; lowest risk, highest leverage.
2. **H1 + H2 + H8** (reentrancy guard, write ordering, stable keys) — the core URL/render correctness defects.
3. **H3 + H4 + H7** (async generation guards, in-memory source of truth, instance-scoped snapshot).
4. **M-tier** hardening, then **L-tier**.

H8/M1 share one root cause — **no stable component IDs** — which the code comments repeatedly call out. Addressing that resolves several findings at once.

---
*Developed in collaboration with Claude Code*
