<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, setContext } from "svelte";
  import { featureFlags } from "../../features/feature-flags";
  import { invalidateRuntimeQueries } from "../invalidation";
  import { RUNTIME_CONTEXT_KEY, runtimeClientStore } from "./context";
  import { RuntimeClient, type AuthContext } from "./runtime-client";

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;
  export let authContext: AuthContext = "user";

  // Created once per mount. If host/instanceId change, the parent's {#key} re-mounts us.
  const client = new RuntimeClient({ host, instanceId, jwt, authContext });
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
    runtimeClientStore.update((c) => (c === client ? null : c));
    client.dispose();
  });
</script>

<slot />
