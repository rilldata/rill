<script lang="ts">
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { initPylonWidget } from "@rilldata/web-common/features/help/initPylonWidget";
  import { RillTheme } from "@rilldata/web-common/layout";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createLocalServiceGetMetadata } from "@rilldata/web-common/runtime-client/local-service";
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import type { Query } from "@tanstack/query-core";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { ApplicationBuildMetadata } from "@rilldata/web-common/layout/build-metadata";
  import { initMetrics } from "@rilldata/web-common/metrics/initMetrics";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";
  import ResourceWatcher from "@rilldata/web-common/features/entity-management/ResourceWatcher.svelte";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import RepresentingUserBanner from "@rilldata/web-common/features/authentication/RepresentingUserBanner.svelte";

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
  initPylonWidget();

  let removeJavascriptListeners: () => void;

  $: metadata = createLocalServiceGetMetadata({
    query: {
      queryClient,
    },
  });
  $: if ($metadata.data) {
    initMetrics($metadata.data).then(() => {
      removeJavascriptListeners =
        errorEventHandler.addJavascriptErrorListeners();
    });

    featureFlags.set(false, "adminServer");
    featureFlags.set($metadata.data.readonly, "readOnly");

    appBuildMetaStore.set({
      version: $metadata.data.version,
      commitHash: $metadata.data.buildCommit,
    });
  }

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
    {#if !$metadata.data?.deployOnly || $page.route.id === "/(misc)/deploy"}
      <ResourceWatcher {host} {instanceId}>
        <div class="body h-screen w-screen overflow-hidden absolute">
          <RepresentingUserBanner />
          <slot />
        </div>
      </ResourceWatcher>
    {:else}
      <ErrorPage
        fatal
        statusCode={500}
        header="Deploy only"
        body="This is a deploy only server. Please stop the server and reload the page."
      />
    {/if}
  </QueryClientProvider>
</RillTheme>

<NotificationCenter />

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
