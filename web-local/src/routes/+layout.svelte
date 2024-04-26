<script lang="ts">
  import { RillTheme } from "@rilldata/web-common/layout";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { beforeNavigate } from "$app/navigation";
  import { retainFeaturesFlags } from "@rilldata/web-common/features/feature-flags";
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
  import AddSourceModal from "@rilldata/web-common/features/sources/modal/AddSourceModal.svelte";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/modal/FileDrop.svelte";
  import SourceImportedModal from "@rilldata/web-common/features/sources/modal/SourceImportedModal.svelte";
  import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import {
    importOverlayVisible,
    overlay,
  } from "@rilldata/web-common/layout/overlay-store";
  import PreparingImport from "@rilldata/web-common/features/sources/modal/PreparingImport.svelte";

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

  beforeNavigate(retainFeaturesFlags);

  export let data;

  let showDropOverlay = false;

  $: overlayData = $overlay;

  onMount(async () => {
    const config = await runtimeServiceGetConfig();
    await initMetrics(config);

    featureFlags.set(false, "adminServer");
    featureFlags.set(config.readonly, "readOnly");
    // Disable AI when running e2e tests
    featureFlags.set(!import.meta.env.VITE_PLAYWRIGHT_TEST, "ai");

    appBuildMetaStore.set({
      version: config.version,
      commitHash: config.build_commit,
    });
  });

  function isEventWithFiles(event: DragEvent) {
    let types = event?.dataTransfer?.types;
    return types && types.indexOf("Files") != -1;
  }
</script>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <WelcomePageRedirect>
      <ResourceWatcher host={data.host} instanceId={data.instanceId}>
        <main
          role="application"
          class="index-body body absolute w-screen h-screen flex overflow-hidden"
          on:drag|preventDefault|stopPropagation
          on:drop|preventDefault|stopPropagation
          on:dragenter|preventDefault|stopPropagation
          on:dragleave|preventDefault|stopPropagation
          on:dragover|preventDefault|stopPropagation={(e) => {
            if (isEventWithFiles(e)) showDropOverlay = true;
          }}
        >
          <slot />

          {#if $importOverlayVisible}
            <PreparingImport />
          {:else if showDropOverlay}
            <FileDrop bind:showDropOverlay />
          {:else if overlayData !== null}
            <BlockingOverlayContainer
              bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
            >
              <div slot="title" class="font-bold">
                {overlayData?.title}
              </div>
              <svelte:fragment slot="detail">
                {#if overlayData?.detail}
                  <svelte:component
                    this={overlayData.detail.component}
                    {...overlayData.detail.props}
                  />
                {/if}
              </svelte:fragment>
            </BlockingOverlayContainer>
          {/if}

          <AddSourceModal />
          <SourceImportedModal sourcePath={$sourceImportedPath} />
          <NotificationCenter />
        </main>
      </ResourceWatcher>
    </WelcomePageRedirect>
  </QueryClientProvider>
</RillTheme>

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
