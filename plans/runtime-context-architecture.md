# Runtime Context Architecture Plan

## Summary

This document proposes a new architecture for managing runtime context (host, instanceId, JWT) in the frontend. The new architecture enables multi-runtime support, cleaner component code, and a migration path from Orval/REST to Connect/gRPC.

## Problem Statement

### The Immediate Bug (PR #8559)

Canvas navigation between projects causes errors because:

1. SvelteKit load functions call `setRuntime()` during navigation
2. The global `runtime` store updates immediately
3. Old components (still mounted) react to the new `instanceId`
4. They attempt to access canvas entities that don't exist for the new instanceId
5. Error occurs before old components unmount

### The Underlying Architecture Issue

The current architecture uses a **global mutable store** (`runtime`) that:

- Is updated from load functions (wrong timing)
- Is read reactively by components (causes race conditions)
- Cannot support multiple runtimes simultaneously
- Mixes concerns (auth, routing, data fetching)

## Options Considered

### Option A: Quick Fix (PR #8559)

- Remove `setRuntime` from load functions
- Set runtime via `RuntimeProvider` component (after old tree unmounts)
- Add `{#key instanceId}` to force component remount
- Add `enabled: !!instanceId` guards on queries

**Verdict:** Fixes the immediate bug but doesn't address architectural issues.

### Option B: HTTP Client State (PR #8572)

- Remove the `runtime` store entirely
- Store host, instanceId, JWT on `httpClient` singleton
- Components call `httpClient.getInstanceId()` (non-reactive)

**Verdict:** Cleaner than A, but commits to a singleton pattern that doesn't support multi-runtime. Would be a detour if we want multi-runtime later.

### Option C: Context-Based Architecture with Connect Web

- Use Svelte context to provide runtime configuration
- Migrate from Orval/REST to Connect/gRPC
- Generate TanStack Query hooks that use context
- Support multiple runtimes by nesting providers

**Verdict:** Recommended. Solves the immediate problem, enables multi-runtime, and aligns with the desired migration to Connect/gRPC.

## Recommended Architecture

### Core Concepts

```
┌─────────────────────────────────────────────────────────┐
│  RuntimeProvider                                        │
│    - Creates Connect transport with host + auth         │
│    - Sets transport in Svelte context                   │
│    - Children only render when transport is ready       │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  use{Service}() Factory Hook                            │
│    - Calls getContext() to get transport                │
│    - Creates Connect client                             │
│    - Returns query/mutation creators                    │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Component                                              │
│    const { createGetExploreQuery } = useRuntimeService()│
│    const query = createGetExploreQuery({ name });       │
└─────────────────────────────────────────────────────────┘
```

### RuntimeProvider

```svelte
<!-- RuntimeProvider.svelte -->
<script lang="ts">
  import { setContext } from 'svelte';
  import { createConnectTransport } from '@connectrpc/connect-web';
  import type { Transport } from '@connectrpc/connect';

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;

  const transport = createConnectTransport({
    baseUrl: host,
    interceptors: jwt ? [authInterceptor(jwt)] : [],
  });

  // Provide both transport and instanceId via context
  setContext('runtime', { transport, instanceId });
</script>

{#if host && instanceId}
  <slot />
{/if}
```

### Generated Service Hook

```typescript
// Generated: useRuntimeService.ts
import { getContext } from 'svelte';
import { createClient } from '@connectrpc/connect';
import { createQuery, createMutation } from '@tanstack/svelte-query';
import { RuntimeService } from '../proto/gen/rill/runtime/v1/api_connect';
import type { Transport } from '@connectrpc/connect';

interface RuntimeContext {
  transport: Transport;
  instanceId: string;
}

export function useRuntimeService() {
  const { transport, instanceId } = getContext<RuntimeContext>('runtime');
  const client = createClient(RuntimeService, transport);

  return {
    instanceId,

    createGetExploreQuery: (
      params: { name: string },
      options?: { query?: CreateQueryOptions }
    ) => createQuery({
      queryKey: ['RuntimeService', 'getExplore', instanceId, params],
      queryFn: () => client.getExplore({ instanceId, ...params }),
      enabled: !!instanceId,
      ...options?.query,
    }),

    createListResourcesQuery: (
      params: { kind?: string },
      options?: { query?: CreateQueryOptions }
    ) => createQuery({
      queryKey: ['RuntimeService', 'listResources', instanceId, params],
      queryFn: () => client.listResources({ instanceId, ...params }),
      enabled: !!instanceId,
      ...options?.query,
    }),

    // ... other RPCs
  };
}

// For use in load functions (explicit transport)
export function createRuntimeServiceClient(transport: Transport) {
  return createClient(RuntimeService, transport);
}
```

### Component Usage

```svelte
<script lang="ts">
  const { instanceId, createGetExploreQuery } = useRuntimeService();

  export let exploreName: string;

  const exploreQuery = createGetExploreQuery({ name: exploreName });
</script>

{#if $exploreQuery.isLoading}
  <LoadingSpinner />
{:else if $exploreQuery.data}
  <ExploreDashboard explore={$exploreQuery.data} />
{/if}
```

### Load Function Usage

```typescript
// +layout.ts
export async function load({ params }) {
  // Fetch runtime config (host, jwt) from admin API or parent
  const runtimeConfig = await fetchProjectRuntime(params.org, params.project);

  return {
    runtime: runtimeConfig, // Passed to RuntimeProvider via data prop
  };
}
```

```svelte
<!-- +layout.svelte -->
<script>
  export let data;
</script>

<RuntimeProvider
  host={data.runtime.host}
  instanceId={data.runtime.instanceId}
  jwt={data.runtime.jwt}
>
  <slot />
</RuntimeProvider>
```

### Load Function Data Fetching

When load functions need to fetch data:

```typescript
// +page.ts
export async function load({ params, parent }) {
  const { runtime } = await parent();

  const transport = createConnectTransport({
    baseUrl: runtime.host,
    interceptors: runtime.jwt ? [authInterceptor(runtime.jwt)] : [],
  });

  const client = createRuntimeServiceClient(transport);
  const explore = await client.getExplore({
    instanceId: runtime.instanceId,
    name: params.exploreName
  });

  return { explore };
}
```

## Multi-Runtime Support

The context-based architecture naturally supports multiple runtimes:

```svelte
<!-- Compare two projects side-by-side -->
<div class="comparison">
  <RuntimeProvider {...projectA}>
    <ProjectDashboard />
  </RuntimeProvider>

  <RuntimeProvider {...projectB}>
    <ProjectDashboard />
  </RuntimeProvider>
</div>
```

Each subtree uses its own transport and instanceId. Components don't need to know which runtime they're in.

## Code Generator

### Input

The generator reads from `*_connect.ts` files produced by `protoc-gen-es`:

```typescript
// proto/gen/rill/runtime/v1/api_connect.ts
export const RuntimeService = {
  typeName: "rill.runtime.v1.RuntimeService",
  methods: {
    getExplore: {
      name: "GetExplore",
      I: GetExploreRequest,
      O: GetExploreResponse,
      kind: MethodKind.Unary,
    },
    // ...
  }
};
```

### Output

For each service, generates:

1. `use{Service}.ts` - Factory hook for components (context-based)
2. Query key generators for cache management
3. Type exports for request/response

### Generator Scope

Estimated ~300-500 lines of TypeScript. Responsibilities:

- Parse service definitions from `*_connect.ts`
- Determine query vs mutation (unary GETs → query, others → mutation)
- Generate TanStack Query wrappers with proper typing
- Generate query key factories

## Migration Strategy

### Phase 0: Immediate Fix (Now)

Merge PR #8559's quick fix to unblock the Canvas navigation bug. This is compatible with the long-term architecture.

### Phase 1: Infrastructure (1-2 weeks)

1. Create `RuntimeProvider` component
2. Write the code generator
3. Generate hooks for `LocalService` (already using Connect)
4. Validate pattern works end-to-end

### Phase 2: Incremental Migration

Migrate RPCs incrementally, not all at once:

1. **Per-RPC migration:** Generate Connect hook for one RPC, update components, verify
2. **Coexistence:** Old Orval hooks and new Connect hooks can coexist
3. **Priority order:**
   - Start with low-traffic RPCs to validate
   - Then high-pain RPCs (ones causing issues)
   - Leave rarely-used RPCs for last

### Phase 3: Cleanup

Once a service is fully migrated:

1. Remove Orval-generated code for that service
2. Update Orval config to exclude migrated services
3. Eventually remove Orval dependency entirely

## JWT Handling

### Refresh Flow

```typescript
// RuntimeProvider handles JWT refresh
<script>
  export let jwt: string | undefined;
  export let onJwtExpiring: () => Promise<string>;

  const transport = createConnectTransport({
    baseUrl: host,
    interceptors: [
      createAuthInterceptor(jwt, onJwtExpiring),
    ],
  });
</script>
```

The interceptor can detect expiring JWTs and trigger refresh before requests fail.

### User Impersonation

```typescript
// Switch to viewing as another user
async function impersonateUser(userId: string) {
  const newJwt = await adminService.getImpersonationToken(userId);
  // Update RuntimeProvider props → new transport created → queries refetch
}
```

Since transport is recreated when props change, queries automatically use the new auth context.

## Open Questions

1. **Query key namespacing:** Should query keys include a version or hash to handle proto schema changes?

2. **Streaming RPCs:** How should server-streaming RPCs (like `WatchResources`) integrate with TanStack Query?

3. **Error handling:** Should the generator produce error type mappings from Connect errors to application errors?

4. **Caching strategy:** Should we generate cache update helpers for common patterns (optimistic updates, cache invalidation)?

## Appendix: Comparison with Current Architecture

| Aspect | Current (Orval + Global Store) | Proposed (Connect + Context) |
|--------|-------------------------------|------------------------------|
| Runtime config | Global mutable store | Svelte context per subtree |
| Client generation | Orval from OpenAPI | Custom generator from protobuf |
| Protocol | REST/HTTP | Connect (gRPC-compatible) |
| Multi-runtime | Not supported | Supported via nested providers |
| Type safety | Generated from OpenAPI | Generated from protobuf |
| Load function support | Problematic (caused bug) | Supported with explicit client |
| Migration | N/A | Incremental, per-RPC |

## References

- [PR #8559: Canvas navigation fix](https://github.com/rilldata/rill/pull/8559)
- [PR #8572: Refactor instanceId handling](https://github.com/rilldata/rill/pull/8572)
- [Connect Web documentation](https://connectrpc.com/docs/web/getting-started/)
- [TanStack Query Svelte](https://tanstack.com/query/latest/docs/framework/svelte/overview)
- [Connect-Query (React reference)](https://github.com/connectrpc/connect-query-es)
