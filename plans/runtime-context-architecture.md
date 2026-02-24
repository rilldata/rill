# Runtime Context + ConnectRPC Migration

## Context

Rill's frontend uses a global mutable Svelte store (`runtime`) to hold connection info (host, instanceId, JWT) for talking to the runtime data plane. This singleton prevents supporting multiple simultaneous runtimes, which is needed for cloud editing (PR #8912, where a dev deployment coexists with production). The cloud editing MVP hacks around the global store with conditional rendering, manual overwriting, and cleanup-on-destroy. This plan replaces the global store with Svelte context + ConnectRPC, building on Brian's POC (PR #8603) which introduced the proto gen pipeline and proved the ConnectRPC approach.

**Key design decisions:**
- **`RuntimeClient` class** — encapsulates transport, JWT lifecycle, and service clients. Each `RuntimeProvider` creates its own instance and sets it in Svelte context. Not a singleton; scoped to the provider subtree.
- JWT reactivity via **mutable ref inside the class** — transport interceptor reads `this.currentJwt`, updated by the provider when props change. No re-mounting needed.
- No singleton manager — **Svelte context IS the registry**; nesting providers gives multi-runtime
- Code generator built early — 62+ Orval-generated functions are too many to hand-write
- Global store removed ASAP — bridge exists only during migration, not as a permanent pattern

---

## Phase 1: Foundation

Branch off Brian's `bgh/connectrpc-poc`, keeping:
- `proto/buf.gen.runtime.yaml` (buf config for ConnectRPC TS stubs)
- Generated `*_connect.ts` files in `web-common/src/proto/gen/rill/runtime/v1/`
- `@connectrpc/connect`, `@connectrpc/connect-web` dependencies in `web-common/package.json`

Discard:
- `ProjectProvider.svelte`, `project-manager.ts` (replaced by Svelte context)
- `connectrpc.ts` (replaced by code generator output)
- All consumer-side changes (Leaderboard, timeseries-data-store, etc.)

Create `web-common/src/runtime-client/v2/`:
```
v2/
  runtime-client.ts     # RuntimeClient class
  context.ts            # useRuntimeClient() — reads from getContext()
  RuntimeProvider.svelte # Creates RuntimeClient, sets context
  gen/                  # Code generator output (Phase 2)
```

### `RuntimeClient` class

```typescript
// runtime-client.ts
import { createClient, type Client, type Transport } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryService } from "../../proto/gen/rill/runtime/v1/queries_connect";
import { RuntimeService } from "../../proto/gen/rill/runtime/v1/api_connect";
import { ConnectorService } from "../../proto/gen/rill/runtime/v1/connectors_connect";

export type AuthContext = "user" | "mock" | "magic" | "embed";

export class RuntimeClient {
  readonly instanceId: string;
  readonly transport: Transport;

  // JWT state (mutable; read by the transport interceptor)
  private currentJwt: string | undefined;
  private jwtReceivedAt: number;
  private authContext: AuthContext;

  // Cached service clients (created once per RuntimeClient)
  private _queryService: Client<typeof QueryService> | null = null;
  private _runtimeService: Client<typeof RuntimeService> | null = null;
  private _connectorService: Client<typeof ConnectorService> | null = null;

  constructor(opts: {
    host: string;
    instanceId: string;
    jwt?: string;
    authContext?: AuthContext;
  }) {
    this.instanceId = opts.instanceId;
    this.currentJwt = opts.jwt;
    this.jwtReceivedAt = opts.jwt ? Date.now() : 0;
    this.authContext = opts.authContext ?? "user";

    this.transport = createConnectTransport({
      baseUrl: opts.host,
      interceptors: [(next) => async (req) => {
        if (this.currentJwt) {
          await this.waitForFreshJwt();
          req.header.set("Authorization", `Bearer ${this.currentJwt}`);
        }
        return next(req);
      }],
    });
  }

  // Called by RuntimeProvider when the parent passes a new JWT prop
  updateJwt(jwt: string | undefined, authContext?: AuthContext): boolean {
    const authContextChanged = !!this.authContext && !!authContext
      && authContext !== this.authContext;
    if (jwt !== this.currentJwt) {
      this.currentJwt = jwt;
      this.jwtReceivedAt = Date.now();
    }
    if (authContext) this.authContext = authContext;
    return authContextChanged; // caller invalidates queries if true
  }

  // Getter for JWT (used by SSE clients and other non-query consumers)
  getJwt(): string | undefined { return this.currentJwt; }

  // Lazy service client getters
  get queryService() {
    return this._queryService ??= createClient(QueryService, this.transport);
  }
  get runtimeService() {
    return this._runtimeService ??= createClient(RuntimeService, this.transport);
  }
  get connectorService() {
    return this._connectorService ??= createClient(ConnectorService, this.transport);
  }

  // Port of maybeWaitForFreshJWT from http-client.ts:50-70
  private async waitForFreshJwt(): Promise<void> {
    if (this.authContext === "embed") return; // 24h TTL, skip
    const expiresAt = this.jwtReceivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;
    while (Date.now() + JWT_EXPIRY_WARNING_WINDOW > expiresAt) {
      await new Promise((r) => setTimeout(r, 50));
      // Loop exits when provider calls updateJwt() with a fresh token
    }
  }

  dispose(): void {
    // Future: clean up SSE connections, cancel pending requests, etc.
  }
}
```

### `RuntimeProvider.svelte`

`host` and `instanceId` are stable for the provider's lifetime. If they change (e.g. navigating between projects), the **parent layout** re-mounts the provider via `{#key}`. JWT-only changes — including View As / impersonation and periodic refresh — are handled in-place via `client.updateJwt()`.

```svelte
<script lang="ts">
  import { setContext, onDestroy } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { RuntimeClient } from "./runtime-client";
  import { RUNTIME_CONTEXT_KEY } from "./context";
  import { invalidateRuntimeQueries } from "../invalidation";
  import { runtime } from "../runtime-store"; // BRIDGE (temporary)

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;
  export let authContext: AuthContext = "user";

  // Created once. If host/instanceId change, parent's {#key} re-mounts us.
  const client = new RuntimeClient({ host, instanceId, jwt, authContext });
  setContext(RUNTIME_CONTEXT_KEY, client);

  // JWT-only changes (common: 15-min refresh, View As with same host)
  $: {
    const authContextChanged = client.updateJwt(jwt, authContext);
    if (authContextChanged) void invalidateRuntimeQueries(queryClient, instanceId);
  }

  // BRIDGE (temporary): also set global store for unmigrated Orval consumers
  $: runtime.setRuntime(queryClient, host, instanceId, jwt, authContext);

  onDestroy(() => client.dispose());
</script>

{#if host && instanceId}<slot />{/if}
```

### Wire into layouts

- **`web-admin/.../[project]/+layout.svelte`**:
  - Swap import to `v2/RuntimeProvider`
  - Move `RuntimeProvider` higher to wrap BOTH `ProjectTabs` and `<slot />` (currently `ProjectTabs` is outside the provider, lines 169-176)
  - Wrap `RuntimeProvider` in `{#key}` keyed on `host::instanceId` so that project navigation correctly re-mounts the provider (View As does NOT change host/instanceId — it only changes the JWT and auth context):
    ```svelte
    {#key `${effectiveHost}::${effectiveInstanceId}`}
      <RuntimeProvider host={effectiveHost} instanceId={effectiveInstanceId} jwt={effectiveJwt} {authContext}>
        {#if onProjectPage && deploymentStatus === RUNNING}
          <ProjectTabs ... />
        {/if}
        <slot />
      </RuntimeProvider>
    {/key}
    ```
- **`web-local/src/routes/+layout.svelte`**: add `v2/RuntimeProvider` wrapping the app (host from env, instanceId="default", no JWT). No `{#key}` needed (host/instanceId never change).
- **`web-local/src/hooks.client.ts`**: keep `runtime.set()` during bridge period

### `featureFlags` → per-client

`web-common/src/features/feature-flags.ts` is a module-level singleton that subscribes to the global `runtime` store. In the new architecture, feature flags are **per-RuntimeClient** (they're fetched from the instance's API, so different runtimes may have different flags).

During the bridge period, `featureFlags` continues to use the global store. During consumer migration (Phase 4), it becomes a `RuntimeClient` property:
```typescript
class RuntimeClient {
  readonly featureFlags: FeatureFlagStore;
  constructor(opts) {
    this.featureFlags = new FeatureFlagStore(this);
  }
}
```
Components access via `const client = useRuntimeClient(); $: flags = client.featureFlags;`

**Validation:** Both web-admin and web-local load correctly (bridge keeps old consumers alive).

---

## Phase 2: Code Generator

Create `web-common/scripts/generate-query-hooks.ts` — a build-time script that reads the ConnectRPC `*_connect.ts` service descriptors and produces TanStack Query hooks.

### Input

The three generated files:
- `web-common/src/proto/gen/rill/runtime/v1/api_connect.ts` (RuntimeService)
- `web-common/src/proto/gen/rill/runtime/v1/queries_connect.ts` (QueryService)
- `web-common/src/proto/gen/rill/runtime/v1/connectors_connect.ts` (ConnectorService)

### Classification config

`web-common/scripts/generate-query-hooks-config.ts`:
- **ConnectorService**: all methods → query (all GET)
- **QueryService**: all unary methods → query (semantically read-only, POST for complex bodies), except `export`/`exportReport`/`query` → mutation; `queryBatch` → skip (streaming)
- **RuntimeService**: methods starting with `Get`/`List`/`Ping`/`Health`/`Analyze` → query; `Watch*`/`CompleteStreaming` → skip (streaming); `IssueDevJWT`/`GetModelPartitions` → query (per Orval config); rest → mutation

### JSON bridge: Orval types in the public API

ConnectRPC uses proto message classes internally, but proto `oneof` fields use discriminated unions (`{ case: "ident", value: "foo" }`) while the existing Orval types use flat optional fields (`{ ident?: string }`). This type mismatch affects `Expression` (~60 files) and `Resource` (~19 files). If consumers had to change their type usage during migration, the scope would expand significantly.

Instead, the generated hooks use **Orval-compatible types in their public API** and convert to/from proto internally. The protobuf canonical JSON format uses the same flat keys as Orval, and proto message classes provide `fromJson()` / `toJson()` methods for lossless conversion:

```
Consumer passes:    { where: { ident: "foo" } }    ← V1Expression (flat oneof)
fromJson converts:  proto Expression               ← { case: "ident", value: "foo" }
  (RPC call uses proto internally)
toJson converts:    response back to flat JSON
Consumer receives:  { resource: { meta: ... } }    ← V1Resource (flat oneof)
```

The generator reads `index.schemas.ts` at generation time to discover which Orval types exist:

- **Response types:** 100% have `V1{ProtoType}` counterparts. Always use the Orval type.
- **Request types:** ~25% have `V1{ProtoType}` counterparts (POST endpoints with JSON bodies). GET endpoints don't have V1 request types because Orval represents their parameters separately. The generator falls back to `PartialMessage<ProtoType>` for those, which is safe since GET requests have no oneofs.

This means consumers can migrate to v2 hooks while keeping their existing Orval types — no type-level changes needed.

### Output per unary query method (4 tiers)

```typescript
// --- Tier 1: Raw function (JSON bridge: fromJson on input, toJson on output) ---
export function queryServiceMetricsViewAggregation(
  client: RuntimeClient,
  request: Omit<V1MetricsViewAggregationRequest, "instanceId">,
  options?: { signal?: AbortSignal },
): Promise<V1MetricsViewAggregationResponse> {
  return client.queryService.metricsViewAggregation(
    MetricsViewAggregationRequest.fromJson(
      { instanceId: client.instanceId, ...request } as unknown as JsonValue,
    ),
    { signal: options?.signal },
  ).then(r => r.toJson() as unknown as V1MetricsViewAggregationResponse);
}

// --- Tier 2: Query key (no client needed) ---
export function getQueryServiceMetricsViewAggregationQueryKey(
  instanceId: string,
  request?: Omit<V1MetricsViewAggregationRequest, "instanceId">,
): QueryKey
// Format: ["QueryService", "metricsViewAggregation", instanceId, request]

// --- Tier 3: Query options ---
export function getQueryServiceMetricsViewAggregationQueryOptions(
  client: RuntimeClient,
  request: Omit<V1MetricsViewAggregationRequest, "instanceId">,
  options?: { query?: Partial<CreateQueryOptions<V1MetricsViewAggregationResponse>> },
): CreateQueryOptions<V1MetricsViewAggregationResponse> & { queryKey: QueryKey }

// --- Tier 4: Convenience hook ---
export function createQueryServiceMetricsViewAggregation(
  client: RuntimeClient,
  request: Omit<V1MetricsViewAggregationRequest, "instanceId">,
  options?: { query?: Partial<CreateQueryOptions<V1MetricsViewAggregationResponse>> },
): CreateQueryResult<V1MetricsViewAggregationResponse>
```

Key details:
- Generated hooks take `client: RuntimeClient` as the first argument
- `instanceId` is **omitted from the request type** and injected from `client.instanceId`
- Raw functions call `client.queryService.*` (cached service client, not recreated per call)
- Query keys use `["QueryService", "metricsViewAggregation", instanceId, request]` format
- `enabled` defaults to `!!client.instanceId` (generator can add more based on config)
- Mutations follow the same pattern but produce `createMutation` / `getMutationOptions`
- For SSE/streaming consumers, `client.getJwt()` provides current JWT without global store access

### Output location

```
web-common/src/runtime-client/v2/gen/
  query-service.ts       # All QueryService hooks
  runtime-service.ts     # All RuntimeService hooks
  connector-service.ts   # All ConnectorService hooks
  index.ts               # Barrel re-export
```

### npm integration

Add to `web-common/package.json`: `"generate:query-hooks": "tsx scripts/generate-query-hooks.ts"`
Wire into Makefile after `buf generate`.

**Validation:** Generated files compile. Function count matches expected (~59 query + ~28 mutation methods).

---

## Phase 3: Update Invalidation Logic

Before migrating consumers, update cache invalidation to support both old (URL-path) and new (service/method) query key formats.

### Files to modify

**`web-common/src/runtime-client/invalidation.ts`:**
- `invalidateRuntimeQueries`: match `key[0].startsWith("rill.runtime.v1.") && key[2] === instanceId` (new) OR `key[0].startsWith("/v1/instances/${instanceId}")` (old bridge)
- `invalidationForMetricsViewData`: match `key[0] === "QueryService" && key[1] includes "metricsView" && request contains metricsViewName` (new) OR URL regex (old)
- `invalidateAllMetricsViews`: same dual-match pattern
- `invalidateComponentData`: match `key[1] === "resolveComponent" && request.component === name`

**`web-common/src/runtime-client/query-matcher.ts`:**
- `isRuntimeQuery`: match new key format too
- `isProfilingQuery` / `isTableProfilingQuery` / `isColumnProfilingQuery`: match new keys (these use `queryHash` regex which won't work with new format — rewrite to inspect `queryKey` structure)

**Validation:** Existing tests pass (old keys still match). New key format also matches.

---

## Phase 4: Consumer Migration

Migrate all ~100 consumer files from Orval imports + `$runtime` to v2 imports + `useRuntimeClient()`.

### Migration pattern

**.svelte files:**
```diff
- import { createQueryServiceFoo } from "@rilldata/web-common/runtime-client";
- import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
+ import { createQueryServiceFoo, useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
+ const client = useRuntimeClient();

- $: query = createQueryServiceFoo($runtime.instanceId, param, body, opts);
+ $: query = createQueryServiceFoo(client, { param, ...body }, opts);
  // Types stay the same — V1Expression, V1Resource, etc.
```

**.ts query factories** (accept `RuntimeClient` instead of `instanceId: string`):
```diff
- export function useExplore(instanceId: string, name: string) {
-   return createRuntimeServiceGetExplore(instanceId, { name });
+ export function useExplore(client: RuntimeClient, name: string) {
+   return createRuntimeServiceGetExplore(client, { name });
```

**`derived` store patterns** (become simpler since client is stable):
```diff
- const queryOpts = derived(runtime, ($r) =>
-   getQueryServiceFooQueryOptions($r.instanceId, params));
- const query = createQuery(queryOpts);
+ const query = createQueryServiceFoo(client, params);
```

### Critical migration: `state-managers.ts`

`web-common/src/features/dashboards/state-managers/state-managers.ts` threads `runtime: Writable<Runtime>` through the dashboard state system. Replace with `client: RuntimeClient` in the `StateManagers` interface and all downstream consumers.

### Migration order (by dependency, leaves first)

1. **Chart providers** (6 files) — `web-common/src/features/components/charts/*/`
2. **Canvas components** (~15 files) — `web-common/src/features/canvas/`
3. **Connectors / column profile** (~10 files) — `web-common/src/features/connectors/`, `column-profile/`
4. **Dashboard state managers** (~15 files) — `state-managers.ts`, selectors, time controls, pivot
5. **Explore + dashboard data** (~15 files) — leaderboard, time-series, totals
6. **Chat, alerts, reports** (~15 files)
7. **File management / workspaces** (~15 files)
8. **web-admin specific** (~15 files) — project status, dashboards listing
9. **web-local specific** (~5 files)
10. **Special cases** — `query-options.ts`, SSE clients, `StreamingQueryBatch`

### SSE / streaming migration

These consumers currently read JWT from the global `runtime` store. Migrate them to accept `RuntimeClient` (or its `getJwt()` method):
- `sse-fetch-client.ts`: change `start()` to accept `getJwt: () => string | undefined` instead of reading `get(runtime).jwt`
- `StreamingQueryBatch.ts`: accept `RuntimeClient` in constructor; use `client.instanceId`, `client.getJwt()`
- `FileAndResourceWatcher.svelte`: pass `client.getJwt` as the JWT getter
- `file-and-resource-watcher.ts`: accept `instanceId` as constructor parameter

**Validation per batch:** `npm run check`, `npm run test -w web-common` after each batch. Full e2e after batches 4 and 8.

---

## Phase 5: Remove Bridge + Cleanup

Once all consumers are migrated:

### Delete
- `web-common/src/runtime-client/runtime-store.ts`
- `web-common/src/runtime-client/http-client.ts`
- `web-common/src/runtime-client/fetchWrapper.ts`
- `web-common/src/runtime-client/http-request-queue/` (entire directory)
- `web-common/src/runtime-client/gen/` (Orval output)
- `web-common/orval.config.ts`
- `orval` dependency from `web-common/package.json`

### Update
- `v2/RuntimeProvider.svelte`: remove the `$: runtime.setRuntime(...)` bridge line
- `web-local/src/hooks.client.ts`: remove `runtime.set()` call
- `invalidation.ts` / `query-matcher.ts`: remove old URL-path matching branches
- Move `v2/` contents to `web-common/src/runtime-client/` (or update barrel export)
- Update all `runtime-client/v2` imports to `runtime-client`

### Verify
- `grep -r "runtime-store" web-common/ web-admin/ web-local/` → zero results
- `grep -r "from.*runtime-client/gen/" web-common/ web-admin/ web-local/` → zero results
- Full test suite passes
- Manual smoke: explore dashboard, canvas dashboard, View As, file editing, chat

---

## Phase 6: Native Proto Types (optional)

Once the bridge is in place and all consumers use v2 hooks, there's an optional future phase to remove the JSON bridge and use proto types natively. This would replace `V1Expression` with proto `Expression` (discriminated union `{ case: "ident", value: "foo" }`), `V1Resource` with proto `Resource`, and so on across ~80 files. The JSON bridge adds minimal overhead (serialization of typical API payloads is sub-millisecond), so this phase is entirely optional and can be deferred indefinitely.

---

## Addressing Previous Feedback

Responses to comments on the original design doc (PR #8590):

**Brian's comment (ProjectProvider vs RuntimeProvider):** We use `RuntimeProvider` because the abstraction is the data plane connection, not the project. Multiple edit sessions create multiple runtimes for the same project. However, we adopt Brian's key ideas: the `RuntimeClient` class (encapsulating transport, JWT, and service clients), the dual return pattern (`create()` + `options`), and the ConnectRPC proto gen pipeline from his branch.

**Aditya: "Is there a library for [the code generator] already?"** — No. `connect-query-es` only supports React. Brian also noted this gap ([connectrpc/connect-query-es#324](https://github.com/connectrpc/connect-query-es/issues/324)). We build a custom generator (~300-500 lines).

**Aditya: "Load function fetching doesn't populate the TanStack cache"** — Correct. Our plan avoids load function data fetching entirely. All queries go through TanStack Query hooks (which manage the cache). If load functions need to pre-fetch, they can use `queryClient.prefetchQuery()` with the generated query options.

**Aditya: "What does removing Orval give us if the backend REST endpoints still exist?"** — The backend doesn't change (gRPC-Gateway continues serving REST). The benefit is removing the global `httpClient` + `runtime` store on the frontend — the core architectural issue. ConnectRPC clients use the transport from Svelte context, enabling multi-runtime support.

**Aditya: "Would the impersonation function sit above the RuntimeProvider?"** — Yes. The project layout (parent of RuntimeProvider) handles View As: it fetches mock credentials via `GetDeploymentCredentials` and passes them as props to RuntimeProvider. `GetDeploymentCredentials` returns the same production host/instanceId with a different JWT (confirmed in `admin/server/deployment.go:642-647`), so View As is handled entirely by `client.updateJwt()` — no re-mount needed.

---

## PR Strategy

- **PR 1 (infrastructure):** Phases 1-3. RuntimeProvider, code generator (with JSON bridge), invalidation updates. Proves the pattern end-to-end without touching consumers.
- **PR 2 (migration: charts + canvas):** Batches 1-3 — chart providers, canvas components, connectors, column profile. Leaf components, low coupling.
- **PR 3 (migration: dashboard core):** Batches 4-5 — `state-managers.ts`, dashboard selectors, time controls, pivot, leaderboard, time-series, totals. The hardest batch (threading `RuntimeClient` through the state manager system).
- **PR 4 (migration: features):** Batches 6-7 — chat, alerts, scheduled reports, file management, workspaces.
- **PR 5 (migration: app shells + SSE):** Batches 8-10 — web-admin routes/features, web-local routes, SSE clients, `StreamingQueryBatch`.
- **PR 6 (cleanup):** Phase 5. Remove bridge, delete Orval/global store, move `v2/` to primary path.
