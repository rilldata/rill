<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, setContext } from "svelte";
  import { featureFlags } from "../../features/feature-flags";
  import { invalidateRuntimeQueries } from "../invalidation";
  import {
    getRuntimeClient,
    evictRuntimeClient,
    RUNTIME_CONTEXT_KEY,
  } from "./context";
  import type { AuthContext } from "./runtime-client";

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;
  export let authContext: AuthContext = "user";
  export let externalUserId: string | null = null;

  // Returns a cached instance if a load function already created one for this host+instanceId.
  // If host/instanceId change, the parent's {#key} re-mounts us.
  const client = getRuntimeClient({
    host,
    instanceId,
    jwt,
    authContext,
    externalUserId,
  });
  setContext(RUNTIME_CONTEXT_KEY, client);
  featureFlags.setRuntimeClient(client);

  // Handle JWT-only changes (15-min refresh, View As with same host)
  $: {
    const authContextChanged = client.updateJwt(jwt, authContext);
    if (authContextChanged)
      void invalidateRuntimeQueries(queryClient, instanceId);
  }

  onDestroy(() => {
    featureFlags.clearRuntimeClient();
    evictRuntimeClient(client);
    client.dispose();
  });
</script>

<slot />
