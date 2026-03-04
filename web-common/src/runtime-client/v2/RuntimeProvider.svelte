<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, setContext } from "svelte";
  import { featureFlags } from "../../features/feature-flags";
  import { invalidateRuntimeQueries } from "../invalidation";
  import {
    getRuntimeClient,
    evictRuntimeClient,
    RUNTIME_CONTEXT_KEY,
    runtimeClientStore,
  } from "./context";
  import type { AuthContext } from "./runtime-client";

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;
  export let authContext: AuthContext = "user";

  // Returns a cached instance if a load function already created one for this host+instanceId.
  // If host/instanceId change, the parent's {#key} re-mounts us.
  const client = getRuntimeClient({ host, instanceId, jwt, authContext });
  setContext(RUNTIME_CONTEXT_KEY, client);
  runtimeClientStore.set(client);
  featureFlags.setRuntimeClient(client);

  // Handle JWT-only changes (15-min refresh, View As with same host)
  $: {
    const authContextChanged = client.updateJwt(jwt, authContext);
    if (authContextChanged)
      void invalidateRuntimeQueries(queryClient, instanceId);
  }

  onDestroy(() => {
    featureFlags.clearRuntimeClient();
    runtimeClientStore.update((c) => (c === client ? null : c));
    evictRuntimeClient(client);
    client.dispose();
  });
</script>

<slot />
