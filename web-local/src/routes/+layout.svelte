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
  import WelcomePageRedirect from "@rilldata/web-common/features/welcome/WelcomePageRedirect.svelte";
  import type { ApplicationBuildMetadata } from "@rilldata/web-common/layout/build-metadata";
  import { initMetrics } from "@rilldata/web-common/metrics/initMetrics";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";
  import ResourceWatcher from "@rilldata/web-common/features/entity-management/ResourceWatcher.svelte";
  import type { LayoutData } from "./$types";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";

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

  export let data: LayoutData;

  onMount(async () => {
    const config = await runtimeServiceGetConfig();
    await initMetrics(config);

    featureFlags.set(false, "adminServer");
    featureFlags.set(config.readonly, "readOnly");

    appBuildMetaStore.set({
      version: config.version,
      commitHash: config.build_commit,
    });
  });
</script>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <WelcomePageRedirect>
      <ResourceWatcher host={data.host} instanceId={data.instanceId}>
        <div class="body h-screen w-screen overflow-hidden absolute">
          <slot />
        </div>
      </ResourceWatcher>
    </WelcomePageRedirect>
  </QueryClientProvider>
</RillTheme>

<NotificationCenter />

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
