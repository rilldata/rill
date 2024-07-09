<script lang="ts">
  import { RillTheme } from "@rilldata/web-common/layout";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import type { Query } from "@tanstack/query-core";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";
  import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { ApplicationBuildMetadata } from "@rilldata/web-common/layout/build-metadata";
  import { initMetrics } from "@rilldata/web-common/metrics/initMetrics";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";
  import ResourceWatcher from "@rilldata/web-common/features/entity-management/ResourceWatcher.svelte";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  /** This function will initialize the existing node stores and will connect them
   * to the Node server.
   */
  initializeNodeStoreContexts();

  const appBuildMetaStore: Writable<ApplicationBuildMetadata> =
    getContext("rill:app:metadata");

  queryClient.getQueryCache().config.onError = (
    error: AxiosError,
    query: Query,
  ) => errorEventHandler?.requestErrorEventHandler(error, query);

  let removeJavascriptListeners: () => void;

  onMount(async () => {
    const config = await runtimeServiceGetConfig();
    await initMetrics(config);
    removeJavascriptListeners = errorEventHandler.addJavascriptErrorListeners();

    featureFlags.set(false, "adminServer");
    featureFlags.set(config.readonly, "readOnly");

    appBuildMetaStore.set({
      version: config.version,
      commitHash: config.build_commit,
    });
  });

  /**
   * Async mount doesnt support an unsubscribe method.
   * So we need this to make sure javascript listeners for error handler is removed.
   */
  onMount(() => {
    return () => removeJavascriptListeners?.();
  });

  $: ({ host, instanceId } = $runtime);
</script>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <ResourceWatcher {host} {instanceId}>
      <div class="body h-screen w-screen overflow-hidden absolute">
        <slot />
      </div>
    </ResourceWatcher>
  </QueryClientProvider>
</RillTheme>

<NotificationCenter />

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
