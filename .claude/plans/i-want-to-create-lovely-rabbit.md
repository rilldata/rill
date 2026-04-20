# SSE Client Cleanup — Layered Abstractions + Tests

## Context

SSE powers three client flows today: file/resource watching (`FileAndResourceWatcher`), chat completion streaming (`Conversation`), and project log viewing (`ProjectLogsPage`). The client code has grown organically and the layers now blur:

- JSON parsing + event-type routing is re-implemented in every consumer (`file-and-resource-watcher.ts:119-147`, `conversation.ts:376-397`, `ProjectLogsPage.svelte:127-147`).
- `SSEConnectionManager` mixes reconnect logic with auto-close lifecycle. The auto-close exists to manage the browser's 6-connection HTTP quota, and has been retrofitted with `disableAutoClose()` for cloud-editor keep-alive (`sse-connection-manager.ts:151-175`). Two concerns, one class.
- Lifecycle policy (visibility, blur, focus heartbeats) lives inside `FileAndResourceWatcher.svelte:51-68`, unreachable from unit tests.
- `FileAndResourceWatcher` is a ~565-line god class combining transport wiring, JSON parsing, per-resource-kind query invalidation, seen-files bookkeeping, and store integration.
- The singleton `fileAndResourceWatcher` export made sense when Rill Developer hosted one runtime at a time. Rill Cloud editing introduces project/branch switching across runtimes, so the assumption of a single long-lived instance is no longer semantically accurate.
- Tests cover essentially none of this. Only `conversation.spec.ts` exists, and it targets fork behavior, not SSE.

Goal: introduce clean, testable layers; migrate all three consumers onto them (three different consumer shapes — multi-event-type watcher, single-stream chat, single-type logs — are the right stress test for the abstraction); ship a meaningful unit-test suite; fix one real bug (JWT refresh on long-lived streams).

In scope:
- Extract SSE into a layered stack under `web-common/src/runtime-client/sse/`.
- Drop the `fileAndResourceWatcher` singleton. Each mount owns its own watcher; status is exposed via Svelte context.
- Move auto-close/lifecycle policy out of `SSEConnection` into an optional `SSEConnectionLifecycle` layer.
- Migrate `FileAndResourceWatcher`, `Conversation`, and `ProjectLogsPage` onto the new layers.
- Extract per-resource-kind invalidation logic out of the watcher into pure helpers.
- Rename `SSEConnectionManager` → `SSEConnection`.
- Unify the two duplicate `EventEmitter` wirings.
- Add an opt-in `onBeforeReconnect` hook so the cloud editor can refresh its JWT before long-lived stream reconnects.

Explicitly out of scope: any user-visible behavior change beyond the JWT refresh path.

Branch: `ericgreen/sse-client-cleanup` off `ericgreen/cloud-editing-mvp`.

---

## Target Layer Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│ 7. Domain Consumers                                              │
│    FileAndResourceWatcher │ Conversation │ ProjectLogsPage       │
│    (wire subscriber + lifecycle + invalidators)                  │
└──────────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────────┐
│ 6. Pure Invalidators (new)                                       │
│    runtime-client/invalidation/file-invalidators.ts              │
│    runtime-client/invalidation/resource-invalidators.ts          │
│    (per-resource-kind functions: (event, queryClient, …) => void)│
└──────────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────────┐
│ 5. SSESubscriber<TEventMap> (new)                                │
│    Typed decoder registry: event.type → JSON.parse → T           │
│    Emits typed domain events; one place for parse + narrowing    │
│    Normalizes untagged frames (type=undefined) to "message" per  │
│    SSE spec, so success frames from the chat endpoint route      │
│    correctly through the "message" decoder.                      │
└──────────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────────┐
│ 4. SSEConnectionLifecycle (new, optional)                        │
│    Idle auto-close + visibility/focus heartbeat policy           │
│    Injectable; consumers that want a persistent connection       │
│    simply don't attach this layer.                               │
└──────────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────────┐
│ 3. SSEConnection (renamed from SSEConnectionManager)             │
│    Reconnect + exponential backoff + stable-threshold + status   │
│    Auto-close logic removed (moved to layer 4)                   │
│    onBeforeReconnect?: () => Promise<void> hook for JWT refresh  │
└──────────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────────┐
│ 2. SSEFetchClient                                                │
│    fetch + AbortController + JWT header + stream plumbing        │
│    Thinner once parser is extracted                              │
└──────────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────────┐
│ 1. sse-protocol.ts (new)                                         │
│    Pure parser: ReadableStream<Uint8Array> → AsyncIterable<SSEMessage> │
│    Also exports parseSSELine / isEventComplete as pure helpers   │
└──────────────────────────────────────────────────────────────────┘
```

Each layer has one job, one test file, and no upward dependencies.

---

## Subscriber API — handling the default event

The admin AI endpoint emits success frames with no `event:` line (so `SSEMessage.type === undefined`) and emits failures as `event: error`. To avoid forcing consumers to think about this, `SSESubscriber` normalizes `type === undefined` to `"message"` (the default event name in the SSE spec) before decoder lookup.

```ts
type Decoder<T> = (data: string) => T;

class SSESubscriber<TMap extends Record<string, unknown>> {
  constructor(
    connection: SSEConnection,
    decoders: { [K in keyof TMap]: Decoder<TMap[K]> },
    opts?: {
      onUnknown?: (message: SSEMessage) => void;
      onParseError?: (err: unknown, message: SSEMessage) => void;
    }
  );

  on<K extends keyof TMap>(
    type: K,
    handler: (payload: TMap[K]) => void,
  ): () => void;
}
```

Chat event map:

```ts
{
  message: V1CompleteStreamingResponse;     // untagged frames route here
  error: { code: string; error: string };   // event:error frames
}
```

Watcher event map:

```ts
{
  file: V1WatchFilesResponse;
  resource: V1WatchResourcesResponse;
  error: V1WatchErrorPayload;   // the unified /sse endpoint emits event:error on gRPC errors
}
```

Logs event map:

```ts
{
  log: V1WatchLogsResponse;
  error: V1WatchErrorPayload;
}
```

---

## Lifecycle Strategies

`SSEConnectionLifecycle` is **optional**. Omit it entirely and the connection stays open until the consumer calls `close()`. Attach it and the connection participates in the browser's connection budget via visibility/idle-based pausing.

Two lifecycle strategies are used by current consumers:

| Preset       | short | normal | When to use                                                                                                          |
| ------------ | ----- | ------ | -------------------------------------------------------------------------------------------------------------------- |
| `aggressive` | 20 s  | 2 min  | Rill Developer (local). Browser 6-connection limit bites: SSE, queries, and dev assets all share `localhost:<port>`. |
| `none`       | —     | —      | Rill Cloud editor + `Conversation` + `ProjectLogsPage`. Don't attach `SSEConnectionLifecycle` at all.                |

Only these two actual call sites exist today — `FileAndResourceWatcher` is mounted in `web-local/src/routes/+layout.svelte` (aggressive) and `web-admin/src/routes/[organization]/[project]/-/edit/+layout.svelte` (none). If a future surface (e.g. cloud explorer watcher) needs an intermediate strategy, add a preset then — don't pre-build it.

The current `keepAlive` prop on `FileAndResourceWatcher.svelte` maps 1:1 to "no lifecycle attached", so it's superseded by a single `lifecycle: "aggressive" | "none"` prop. The old `keepAlive` prop is removed.

---

## JWT refresh — ownership and wiring

`runtimeClient.getJwt()` reads mutable state from the `RuntimeProvider`. The provider is driven by the admin edit layout's `projectQuery` (`web-admin/src/routes/[organization]/[project]/-/edit/+layout.svelte:57-77`), which backs `createAdminServiceGetProject(organization, project, branch ? { branch } : undefined)`.

Critically, that query only polls (at 2 s) while the deployment is in a transitional state (`PENDING`, `UPDATING`, `STOPPED`, `STOPPING`). Once it reaches `RUNNING`, `refetchInterval` returns `false` and the cached JWT sits untouched until the next page navigation or explicit invalidation. In a steady-state cloud editor session — which is exactly the long-lived case this refactor cares about — there is no ambient refresh. If a JWT expires mid-stream and the server 401s, the reconnect would blindly reuse the stale token.

Fix: `SSEConnection` gains `onBeforeReconnect?: () => Promise<void>`. The hook runs after the backoff delay, before the next `client.start(...)`. For the cloud editor, this is **the** JWT refresh path — not a redundant backup.

Ownership: the hook is provided by the route that owns the JWT — `web-admin/src/routes/[organization]/[project]/-/edit/+layout.svelte`. That route already knows the branch, organization, and project; the hook does:

```ts
import { getAdminServiceGetProjectQueryKey } from "@rilldata/web-admin/client";

const onBeforeReconnect = async () => {
  await queryClient.invalidateQueries({
    queryKey: getAdminServiceGetProjectQueryKey(
      organization,
      project,
      branch ? { branch } : undefined,
    ),
  });
  // next runtimeClient.getJwt() call returns the fresh token
};
```

**Branch is part of the query key.** Without it, the invalidation misses the branch-scoped cache entry and the stream keeps reconnecting with a stale token. The route already computes `branch = extractBranchFromPath($page.url.pathname)` (edit layout line 36); pass the same value into the hook.

The route passes this through `FileAndResourceWatcher.svelte` as a new `onBeforeReconnect` prop. Short-lived consumers (web-local, chat, logs) don't pass it and behavior is unchanged.

**Failure contract.** If the hook rejects (network is down, admin API 401s, project was deleted out from under the session):

- The pending `client.start(...)` is **skipped** — we do not open a transport with a possibly-stale token.
- The rejection is surfaced through `SSEConnection`'s `error` event so consumers can render it.
- **The attempt counts.** `retryAttempts` is incremented the same way a transport failure would be. A repeatedly-failing refresh walks the status from `CONNECTING` to `CLOSED` after `maxRetryAttempts`, just like a repeatedly-failing fetch. Treating hook failures as "free" would leave the connection stuck in `CONNECTING` forever, which is worse UX than a terminal `CLOSED` with a retry affordance.
- A new `start()` call (or manual UI retry) resets `retryAttempts` to zero, same as any session restart.

---

## Changes by File

### New files (`web-common/src/runtime-client/sse/`)

The existing SSE modules move into a dedicated `sse/` directory alongside the new ones — keeps this subsystem discoverable and lets the barrel export be the single import point.

- `sse/sse-protocol.ts` — extract `parseSSELine`, `isEventComplete`, `isValidEvent` from `sse-fetch-client.ts:41-78`. Add `readSSEStream(stream): AsyncIterable<SSEMessage>` as the pure parser. Zero dependencies.
- `sse/sse-fetch-client.ts` — moved + simplified. Consumes `readSSEStream`. Retain `SSEHttpError`.
- `sse/sse-connection.ts` — renamed from `sse-connection-manager.ts`. Remove auto-close API (`scheduleAutoClose`, `disableAutoClose`, `enableAutoClose`, `autoCloseThrottler`, `autoCloseDisabled` — moved to `SSEConnectionLifecycle`). Keep `ConnectionStatus`, backoff, stable-threshold, `resumeIfPaused()`, `pause()`, `close()`, `cleanup()`. Keep `heartbeat()` as a deprecated compatibility alias. Add `onBeforeReconnect` constructor option.
- `sse/sse-connection-lifecycle.ts` — new. Class that takes an `SSEConnectionLifecycleConnection` + `idleTimeouts` and listens to `document.visibilitychange` / `window.blur` / activity signals (or accepts an injected signal source for testing). Arms/cancels an idle throttler that calls `connection.pause()`. Exposes `start()` and `stop()`.
- `sse/sse-subscriber.ts` — new. Generic `SSESubscriber<TMap>` per the API sketch above. Normalizes undefined event type to `"message"`.
- `sse/index.ts` — barrel re-export.

### New invalidator modules (`web-common/src/runtime-client/invalidation/`)

- `resource-invalidators.ts` — one pure function per resource kind, pulled from `file-and-resource-watcher.ts:310-559`: `invalidateForConnectorWrite`, `invalidateForSourceOrModelWrite`, `invalidateForMetricsViewWrite`, `invalidateForExploreWrite`, `invalidateForCanvasWrite`, `invalidateForComponentWrite`, plus the delete variants. Signatures take `(event, previous, queryClient, instanceId, deps)` where `deps` carries `fileArtifacts`, `connectorExplorerStore`, `sourceIngestionTracker`, `runtimeClient`.
- `file-invalidators.ts` — pure functions pulled from `file-and-resource-watcher.ts:205-260`: `handleFileWrite`, `handleFileDelete`, plus the throttled `listFiles` refetch driver. Takes `(event, queryClient, instanceId, deps, throttler)`.

### Modified files

- `web-common/src/lib/event-emitter.ts` — unify. Both `SSEFetchClient` and `SSEConnection` currently instantiate `EventEmitter` and manually bind `on`/`once`. Add a small `withEventEmitter` mixin helper (or an exported binding convention) so both classes use one pattern.
- `web-common/src/features/entity-management/file-and-resource-watcher.ts`:
  - Invalidator branches extracted (see above).
  - `FileAndResourceWatcher` becomes a thin class: constructor takes `{host, instanceId, runtimeClient, queryClient, lifecycle, onBeforeReconnect, deps}`. Builds an `SSEConnection` + `SSESubscriber` + optional `SSEConnectionLifecycle`; routes typed events to the invalidator helpers.
  - **Remove** `export const fileAndResourceWatcher = new FileAndResourceWatcher()`. Instantiate per-mount in the Svelte component.
- `web-common/src/features/entity-management/FileAndResourceWatcher.svelte`:
  - Construct the `FileAndResourceWatcher` **synchronously at the top of `<script>`**, not in `onMount`. `setContext` must run during component initialization — descendants calling `getContext` during their own init would see `undefined` otherwise. The props (`host`, `instanceId`) and the runtime client (via `useRuntimeClient()` getContext) are all available synchronously.
  - Call `setContext(WATCHER_CONTEXT_KEY, { status, watcher })` at init time so `RuntimeTrafficLights.svelte` can subscribe.
  - In `onMount`, do only the side-effectful parts: `watcher.start(url)` and any lifecycle attachment.
  - In `onDestroy`, call `watcher.close(true)`.
  - Remove all window-level event handlers — they're now inside `SSEConnectionLifecycle`.
  - Keep the `{#if $status === CLOSED}` error gate.
  - Replace the `keepAlive` boolean prop with `lifecycle: "aggressive" | "none"`.
  - Accept an optional `onBeforeReconnect: () => Promise<void>` prop.
- `web-common/src/features/entity-management/RuntimeTrafficLights.svelte`:
  - Read the watcher context via `getContext(WATCHER_CONTEXT_KEY)`.
  - Fallback: if the context is absent (component rendered outside a watcher provider), default to a static `writable(ConnectionStatus.CLOSED)` store and render the component in its closed state. Don't throw. A unit test covers this.
- `web-common/src/features/chat/core/conversation.ts` — replace raw `SSEFetchClient` usage (`conversation.ts:370-414`) with `SSESubscriber`. Define the event map above. Drop the ad-hoc `JSON.parse` + try/catch in the message handler; drop the `handleServerError(message.data)` branch in favor of the typed `error` decoder. Keep the existing transport-error handler for `SSEHttpError`. No `SSEConnectionLifecycle`.
- `web-admin/src/features/projects/status/logs/ProjectLogsPage.svelte` — replace the local `SSEConnectionManager` construction + inline `handleMessage` (`ProjectLogsPage.svelte:60-147`) with `SSEConnection` + `SSESubscriber<{ log: V1WatchLogsResponse; error: V1WatchErrorPayload }>`. The subscriber owns the `JSON.parse` and the `message.type !== "log"` filter. Keep the page's UI concerns (status badge, scroll-to-bottom, retry button). `retryConnection()` still calls `connection.start(url, opts)`. No `SSEConnectionLifecycle`.
- `web-admin/src/features/projects/status/logs/log-store.ts` — **new, pure.** Extract the log state machine out of the component: a class (or plain module) exposing `addLog(log: V1Log)`, `getAll()`, `getFiltered({levels, search})`, and a ring-buffer cap of `MAX_LOGS`. The Svelte component becomes a thin view over this store. Keeps the spec below at pure-helper level so web-admin doesn't need jsdom configured.

### Call-site updates

- `web-local/src/routes/+layout.svelte:111` — pass `lifecycle="aggressive"` to `<FileAndResourceWatcher>`.
- `web-admin/src/routes/[organization]/[project]/-/edit/+layout.svelte:172` — replace `keepAlive={true}` with `lifecycle="none"`; pass the `onBeforeReconnect` prop described above.

### Test runner scripts

- `web-admin/package.json` — add `"test:unit": "vitest run"`. The package already has `vitest` as a devDependency but no script. The root `CLAUDE.md` notes that vitest already works in web-admin via `cd web-admin && npx vitest run`; this just makes the workflow discoverable and scriptable.

---

## Tests

All tests use vitest. Mock `fetch` with a helper that returns a custom `ReadableStream` to exercise the parser.

**`web-common/src/runtime-client/sse/sse-protocol.spec.ts`** — pure, fast.

- Single event with single data line.
- Multi-line `data:` accumulates with newlines.
- `event:` field sets type.
- Comments (`:keepalive`) are ignored.
- Empty line is the event boundary.
- Event split across two chunks reassembles.
- CRLF line endings.
- Trailing partial line in buffer is held until next chunk.
- Unknown fields (e.g. `id:`, `retry:`) are ignored — assert, so we notice if support is added.

**`web-common/src/runtime-client/sse/sse-fetch-client.spec.ts`**

- Non-2xx response → emits `SSEHttpError` with `status` and `statusText`.
- Missing response body → emits generic `Error("No response body")`.
- `AbortError` from controller does **not** emit `error`.
- `getJwt` return value is sent as `Authorization: Bearer …`.
- `stop()` during streaming aborts the controller and fires `close`.
- `cleanup()` clears listeners (subsequent emit is a no-op).

**`web-common/src/runtime-client/sse/sse-connection.spec.ts`** — use fake timers.

- Initial `start()`: status goes CLOSED → CONNECTING → OPEN on fetch open.
- Network error with `retryOnError: true` → backoff delays of `1000 * 2 ** n`, stops at `maxRetryAttempts`.
- `maxRetryAttempts` hit → status lands on CLOSED.
- Connection stable > 5 s → `retryAttempts` resets to 0 (both via the open-then-stable path and the `wasStable` close path).
- `pause()` then `heartbeat()` → reconnects.
- `pause()` resets `retryAttempts`.
- `close(cleanup: true)` clears listeners.
- `onBeforeReconnect` is awaited before each retry.
- `onBeforeReconnect` rejection: skips the transport call entirely, emits `error` with the rejection reason, **increments `retryAttempts`** like any other failed attempt, and lands on `CLOSED` after `maxRetryAttempts` consecutive hook failures.

**`web-common/src/runtime-client/sse/sse-connection-lifecycle.spec.ts`**

- Simulated `document.visibilitychange` to hidden → arms idle throttler → `connection.pause()` fires after configured timeout.
- Visibilitychange to visible → calls `connection.resumeIfPaused()`.
- Lifecycle `stop()` removes window/document listeners (no leak on unmount).
- Not attaching (omitting the layer) is a no-op.

**`web-common/src/runtime-client/sse/sse-subscriber.spec.ts`**

- Registered decoder invoked with correct `event.type`; typed payload emitted.
- Untagged message (`event.type === undefined`) routes to the `"message"` decoder. **If no `message` decoder is registered, falls through to `onUnknown`.**
- Unknown event type → `onUnknown` called with the raw message.
- Decoder throws → `onParseError` called; no typed emit.

**`web-common/src/runtime-client/invalidation/resource-invalidators.spec.ts`**

- One `describe` per resource kind. Drive each with a synthetic `V1WatchResourcesResponse` + a mocked `queryClient` (mocks for `invalidateQueries`/`refetchQueries`/`setQueryData`/`getQueryData`) and assert the expected query keys are hit.
- Guards: `resourceVersionChanged` vs. `resourceFinishedReconciling`, new-connector detection, source → model connector swap, missing-table-name short-circuits.
- Delete path: connector removal calls `connectorExplorerStore.deleteItem`; source/model delete invalidates OLAP list tables for the previous connector.

**`web-common/src/runtime-client/invalidation/file-invalidators.spec.ts`**

- `.db` path early return.
- Write event on `/rill.yaml` invalidates dev JWT key + calls `invalidate("app:init")` + emits `rill-yaml-updated`.
- Delete event on `/rill.yaml` invalidates `app:init` but not dev JWT.
- `seenFiles` promotion logic (new files trigger listFiles refetch; delete also triggers).

**`web-common/src/features/entity-management/file-and-resource-watcher.spec.ts`**

- Integration-ish: feed raw `SSEMessage` values through an injected fake `SSEConnection`; assert the correct invalidator is called with the right args.
- Reconnect → `invalidateAll()` + `fileArtifacts.init()` called.

**`web-common/src/features/entity-management/RuntimeTrafficLights.spec.ts`**

- Renders a default/closed state when no watcher context is provided — doesn't throw, doesn't subscribe to an undefined store.

**`web-common/src/features/chat/core/conversation.spec.ts`** — extend existing file.

- Streaming: given an injected fake subscriber, a `message` event (via untagged frame → normalized to `"message"`) with a `V1CompleteStreamingResponse` updates the message cache and emits the `message` event.
- `error` type-tagged message surfaces through `streamError` with the server's error text.
- Transport error path (`SSEHttpError`) still surfaces via `streamError` with the formatted transport message.

**`web-admin/src/features/projects/status/logs/log-store.spec.ts`** — runs via `web-admin`'s vitest. Pure, no jsdom needed (tests the extracted `log-store.ts`, not the Svelte component).

- `addLog` assigns a monotonic `_id` counter.
- `addLog` past `MAX_LOGS` drops the oldest entry (ring buffer).
- `getFiltered({levels: ["error"]})` returns only error logs.
- `getFiltered({search: "foo"})` matches against `message` and `jsonPayload`, case-insensitively.
- `getFiltered` with both filters combines them (intersection).
- Empty filters return everything.

If web-admin's vitest config ever needs a DOM environment for other specs, it's a one-line addition (`test: { environment: "jsdom" }` in `web-admin/vite.config.ts`). For this refactor we explicitly avoid that dependency by keeping all logic in a pure module.

---

## Critical files to modify or create

**Create**

- `web-common/src/runtime-client/sse/sse-protocol.ts`
- `web-common/src/runtime-client/sse/sse-connection-lifecycle.ts`
- `web-common/src/runtime-client/sse/sse-subscriber.ts`
- `web-common/src/runtime-client/sse/index.ts`
- `web-common/src/runtime-client/invalidation/file-invalidators.ts`
- `web-common/src/runtime-client/invalidation/resource-invalidators.ts`
- `web-admin/src/features/projects/status/logs/log-store.ts`
- All `*.spec.ts` files listed in the Tests section.

**Modify (move + rename + shrink)**

- `web-common/src/runtime-client/sse-fetch-client.ts` → `web-common/src/runtime-client/sse/sse-fetch-client.ts`
- `web-common/src/runtime-client/sse-connection-manager.ts` → `web-common/src/runtime-client/sse/sse-connection.ts`
- `web-common/src/features/entity-management/file-and-resource-watcher.ts`
- `web-common/src/features/entity-management/FileAndResourceWatcher.svelte`
- `web-common/src/features/entity-management/RuntimeTrafficLights.svelte`
- `web-common/src/features/chat/core/conversation.ts`
- `web-common/src/features/chat/core/conversation.spec.ts`
- `web-admin/src/features/projects/status/logs/ProjectLogsPage.svelte`
- `web-admin/package.json` (add `test:unit` script)
- `web-common/src/lib/event-emitter.ts`

**Call sites to update**

- `web-local/src/routes/+layout.svelte:111`
- `web-admin/src/routes/[organization]/[project]/-/edit/+layout.svelte:172`

**Reuse (don't rewrite)**

- `web-common/src/lib/throttler.ts` — `SSEConnectionLifecycle` uses it.
- `web-common/src/lib/event-bus/event-bus.ts` — `file-invalidators` emits `rill-yaml-updated` through it.
- `web-common/src/runtime-client/invalidation.ts` — existing `invalidateConnectorQueries` / `invalidateMetricsViewData` / `invalidateComponentData` / `invalidateProfilingQueries` helpers stay; `resource-invalidators` call them.

---

## PR staging

The total blast radius (move + rename + new layers + singleton removal + invalidator extraction + three consumer migrations) is too large for one reviewable PR. Split into two on the same branch, landing sequentially:

**PR 1 — Core SSE layers + tests.**

- New files: `sse/sse-protocol.ts`, `sse/sse-connection-lifecycle.ts`, `sse/sse-subscriber.ts`, `sse/index.ts`; all their `*.spec.ts`.
- Move + rename: `sse-fetch-client.ts`, `sse-connection-manager.ts` → `sse/sse-connection.ts`. Class rename `SSEConnectionManager` → `SSEConnection`. Re-export the old name from the old path as a deprecated alias so consumers compile unchanged.
- Keep auto-close methods on `SSEConnection` as thin forwarders to a new lazily-attached `SSEConnectionLifecycle` — deprecation shim, removed in PR 2.
- Add `onBeforeReconnect` constructor option (wired, no caller uses it yet).
- Unify `EventEmitter` usage.
- No consumer-side changes. Safe to land in isolation.

**PR 2 — Consumer migrations + singleton removal.**

- Extract `file-invalidators.ts` and `resource-invalidators.ts`; their specs.
- Refactor `FileAndResourceWatcher` onto the new layers, drop the singleton, add the context provider.
- Update `FileAndResourceWatcher.svelte` (`lifecycle` prop, `onBeforeReconnect` prop, context wiring).
- Update `RuntimeTrafficLights.svelte` to read from context with a safe fallback + spec.
- Migrate `Conversation` onto `SSESubscriber`.
- Migrate `ProjectLogsPage` onto `SSEConnection` + `SSESubscriber`; add `web-admin` vitest script.
- Update call sites in `web-local` and the cloud-editor route; wire `onBeforeReconnect` in the editor.
- Delete the deprecation shims from PR 1.

Either PR can be landed independently if reviewers prefer (PR 1 alone is a pure introduce-new-layers change; PR 2 depends on PR 1).

---

## Verification

End-to-end check after PR 2 lands on the branch:

1. **Unit tests**
   - `npm run test -w web-common` — all new web-common specs pass; existing `conversation.spec.ts` still passes.
   - `npm run test:unit -w web-admin` (or `cd web-admin && npx vitest run`) — `ProjectLogsPage.spec.ts` passes.
2. **Type + lint**: `npm run quality`.
3. **Rill Developer (`lifecycle="aggressive"`)**: `rill devtool start local`. Open a project, edit a SQL file, confirm:
   - Dashboard refreshes on save.
   - `RuntimeTrafficLights` reflects connection state correctly.
   - Connection pauses after ~2 min of inactivity (tab hidden); reconnects on focus.
4. **Rill Cloud editor (`lifecycle="none"`)**: `rill devtool start cloud`, open `-/edit`. Confirm:
   - File writes stream through without pausing.
   - Leaving the tab hidden for 20+ min does **not** close the connection.
   - Triggering a reconcile routes resource events to invalidations (dashboards update).
5. **Chat streaming**: open an AI conversation, send a prompt, confirm the response streams token-by-token (validates untagged-frame → "message" normalization). Force a server-side error and confirm the typed `error` channel surfaces the message in `streamError`.
6. **Project logs**: open the cloud project's `Status → Logs` view. Confirm logs stream live, filter/search still work, scroll-to-bottom behavior preserved, and `Retry` reconnects after a forced disconnect.
7. **Branch switching (cloud editing)**: switch between two project branches. Confirm the watcher tears down cleanly and the new runtime's watcher comes up. Exercises the dropped-singleton path and is the main behavioral motivation for per-mount instances.
8. **Long-lived JWT**: with the cloud editor open, force a JWT-expiry simulation (shorten TTL in dev) and confirm the `onBeforeReconnect` hook refreshes JWT cleanly on reconnect.
9. **Graceful fallback**: unit test `RuntimeTrafficLights.spec.ts` confirms the component renders a default-closed state when no watcher context is provided.
