<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, onMount, setContext } from "svelte";
  import { featureFlags } from "../../features/feature-flags";
  import { invalidateRuntimeQueries } from "../invalidation";
  import {
    getRuntimeClient,
    evictRuntimeClient,
    RUNTIME_CONTEXT_KEY,
  } from "./context";
  import type { AuthContext, JwtRefresher } from "./runtime-client";
  import { createEnvFileStore } from "@rilldata/web-common/features/env-management/env-file-store.ts";

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;
  export let authContext: AuthContext = "user";
  // Fetches a fresh JWT when a request finds the current one expired, e.g. after a laptop sleep
  // skipped the scheduled refetch. Without it, requests wait for the parent to pass a new `jwt`.
  export let refreshJwt: JwtRefresher | undefined = undefined;

  // Returns a cached instance if a load function already created one for this host+instanceId.
  // If host/instanceId change, the parent's {#key} re-mounts us.
  const client = getRuntimeClient({ host, instanceId, jwt, authContext });
  setContext(RUNTIME_CONTEXT_KEY, client);
  featureFlags.setRuntimeClient(client);

  const envStoreUnsub = createEnvFileStore(client);

  // Handle JWT-only changes (15-min refresh, View As with same host)
  $: {
    const authContextChanged = client.updateJwt(jwt, authContext);
    client.setJwtRefresher(refreshJwt);
    if (authContextChanged)
      void invalidateRuntimeQueries(queryClient, instanceId);
  }

  // A tab that was asleep, hidden or offline has likely missed the scheduled JWT refetch.
  // Refresh as soon as it is back so the first requests don't have to wait for it.
  onMount(() => {
    const refreshIfStale = () => void client.refreshJwtIfStale();
    const refreshWhenVisible = () => {
      if (document.visibilityState !== "hidden") refreshIfStale();
    };
    window.addEventListener("online", refreshIfStale);
    window.addEventListener("focus", refreshWhenVisible);
    document.addEventListener("visibilitychange", refreshWhenVisible);
    return () => {
      window.removeEventListener("online", refreshIfStale);
      window.removeEventListener("focus", refreshWhenVisible);
      document.removeEventListener("visibilitychange", refreshWhenVisible);
    };
  });

  onDestroy(() => {
    featureFlags.clearRuntimeClient();
    evictRuntimeClient(client);
    client.dispose();
    envStoreUnsub();
  });
</script>

<slot />
