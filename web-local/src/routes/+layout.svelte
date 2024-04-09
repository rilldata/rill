<script lang="ts">
  import { RillTheme } from "@rilldata/web-common/layout";
  import { addViewportListener } from "@rilldata/web-common/lib/viewport-utils";
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import FileDrop from "@rilldata/web-common/features/sources/modal/FileDrop.svelte";
  import SourceImportedModal from "@rilldata/web-common/features/sources/modal/SourceImportedModal.svelte";
  import {
    duplicateSourceName,
    sourceImportedName,
  } from "@rilldata/web-common/features/sources/sources-store";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import type { ApplicationBuildMetadata } from "@rilldata/web-common/layout/build-metadata";
  import { initMetrics } from "@rilldata/web-common/metrics/initMetrics";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";
  import AddSourceModal from "@rilldata/web-common/features/sources/modal/AddSourceModal.svelte";
  import PreparingImport from "@rilldata/web-common/features/sources/modal/PreparingImport.svelte";
  import { addSourceModal } from "@rilldata/web-common/features/sources/modal/add-source-visibility";
  import WelcomePageRedirect from "@rilldata/web-common/features/welcome/WelcomePageRedirect.svelte";
  import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
  import {
    importOverlayVisible,
    overlay,
  } from "@rilldata/web-common/layout/overlay-store";
  import { retainFeaturesFlags } from "@rilldata/web-common/features/feature-flags";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import type { Query } from "@tanstack/query-core";
  import type { AxiosError } from "axios";
  import { beforeNavigate } from "$app/navigation";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { startWatchFilesClient } from "@rilldata/web-common/features/entity-management/watch-files-client";
  import { startWatchResourcesClient } from "@rilldata/web-common/features/entity-management/watch-resources-client";

  let showDropOverlay = false;

  addViewportListener();
  initializeNodeStoreContexts();

  queryClient.getQueryCache().config.onError = (
    error: AxiosError,
    query: Query,
  ) => errorEventHandler?.requestErrorEventHandler(error, query);

  beforeNavigate(retainFeaturesFlags);

  onMount(() => {
    const stopWatchFilesClient = startWatchFilesClient(queryClient);
    const stopWatchResourcesClient = startWatchResourcesClient(queryClient);
    const stopJavascriptErrorListeners =
      errorEventHandler?.addJavascriptErrorListeners();
    void fileArtifacts.init(queryClient, "default");

    return () => {
      stopWatchFilesClient();
      stopWatchResourcesClient();
      stopJavascriptErrorListeners?.();
    };
  });

  const appBuildMetaStore: Writable<ApplicationBuildMetadata> =
    getContext("rill:app:metadata");

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

<QueryClientProvider client={queryClient}>
  <RillTheme>
    <div class="body">
      {#if $importOverlayVisible}
        <PreparingImport />
      {:else if showDropOverlay}
        <FileDrop bind:showDropOverlay />
      {:else if $overlay !== null}
        <BlockingOverlayContainer
          bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
        >
          <div slot="title" class="font-bold">
            {$overlay?.title}
          </div>
          <svelte:fragment slot="detail">
            {#if $overlay?.detail}
              <svelte:component
                this={$overlay.detail.component}
                {...$overlay.detail.props}
              />
            {/if}
          </svelte:fragment>
        </BlockingOverlayContainer>
      {/if}

      {#if $addSourceModal || $duplicateSourceName}
        <AddSourceModal />
      {/if}
      <SourceImportedModal source={$sourceImportedName} />

      <div
        class="index-body absolute w-screen h-screen"
        on:dragenter|preventDefault|stopPropagation
        on:dragleave|preventDefault|stopPropagation
        on:dragover|preventDefault|stopPropagation={(e) => {
          if (isEventWithFiles(e)) showDropOverlay = true;
        }}
        on:drag|preventDefault|stopPropagation
        on:drop|preventDefault|stopPropagation
        role="application"
      >
        <WelcomePageRedirect>
          <slot />
        </WelcomePageRedirect>
      </div>
    </div>

    <NotificationCenter />
  </RillTheme>
</QueryClientProvider>

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
