<script lang="ts">
  import { setContext, onDestroy } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { RuntimeClient, type AuthContext } from "./runtime-client";
  import { RUNTIME_CONTEXT_KEY } from "./context";
  import { invalidateRuntimeQueries } from "../invalidation";
  import { featureFlags } from "../../features/feature-flags";

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;
  export let authContext: AuthContext = "user";

  // Created once per mount. If host/instanceId change, the parent's {#key} re-mounts us.
  const client = new RuntimeClient({ host, instanceId, jwt, authContext });
  setContext(RUNTIME_CONTEXT_KEY, client);
  featureFlags.setRuntimeClient(client);

  // Handle JWT-only changes (15-min refresh, View As with same host)
  $: {
    const authContextChanged = client.updateJwt(jwt, authContext);
    if (authContextChanged)
      void invalidateRuntimeQueries(queryClient, instanceId);
  }

  onDestroy(() => client.dispose());
</script>

{#if host && instanceId}
  <slot />
{/if}
