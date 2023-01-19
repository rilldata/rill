<script lang="ts">
  import { page } from "$app/stores";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import DuplicateSource from "@rilldata/web-common/features/sources/add-source/DuplicateSource.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/add-source/FileDrop.svelte";
  import { duplicateSourceName } from "@rilldata/web-common/features/sources/sources-store";
  import BlockingOverlayContainer from "@rilldata/web-common/features/temp/BlockingOverlayContainer.svelte";
  import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
  import { QueryClientProvider } from "@sveltestack/svelte-query";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";
  import { runtimeStore } from "../../application-state-stores/application-store";
  import type { ApplicationBuildMetadata } from "../../application-state-stores/build-metadata";
  import { fileArtifactsStore } from "../../application-state-stores/file-artifacts-store";
  import {
    importOverlayVisible,
    overlay,
    quickStartDashboardOverlay,
  } from "../../application-state-stores/overlay-store";
  import { initMetrics } from "../../metrics/initMetrics";
  import { getArtifactErrors } from "../../svelte-query/getArtifactErrors";
  import { createQueryClient } from "../../svelte-query/globalQueryClient";
  import PreparingImport from "../overlay/PreparingImport.svelte";
  import QuickStartDashboard from "../overlay/QuickStartDashboard.svelte";
  import BasicLayout from "./BasicLayout.svelte";
  import {
    addReconcilingOverlay,
    syncFileSystemPeriodically,
  } from "./sync-file-system";

  const queryClient = createQueryClient();

  const appBuildMetaStore: Writable<ApplicationBuildMetadata> =
    getContext("rill:app:metadata");
  onMount(async () => {
    const config = await runtimeServiceGetConfig();

    runtimeStore.set({
      instanceId: config.instance_id,
      readOnly: config.readonly,
    });

    appBuildMetaStore.set({
      version: config.version,
      commitHash: config.build_commit,
    });

    const res = await getArtifactErrors(config.instance_id);
    fileArtifactsStore.setErrors(res.affectedPaths, res.errors);

    return initMetrics(config);
  });

  syncFileSystemPeriodically(
    queryClient,
    runtimeStore,
    page,
    fileArtifactsStore
  );
  $: addReconcilingOverlay($page.url.pathname);

  let dbRunState = "disconnected";
  let runstateTimer;

  function debounceRunstate(state) {
    if (runstateTimer) clearTimeout(runstateTimer);
    setTimeout(() => {
      dbRunState = state;
    }, 500);
  }

  let showDropOverlay = false;

  // TODO: add new global run state
  $: debounceRunstate("disconnected");

  function isEventWithFiles(event: DragEvent) {
    let types = event.dataTransfer.types;
    return types && types.indexOf("Files") != -1;
  }
</script>

<QueryClientProvider client={queryClient}>
  <div class="body">
    {#if $importOverlayVisible}
      <PreparingImport />
    {:else if $quickStartDashboardOverlay?.show}
      <QuickStartDashboard
        sourceName={$quickStartDashboardOverlay.sourceName}
        timeDimension={$quickStartDashboardOverlay.timeDimension}
      />
    {:else if showDropOverlay}
      <FileDrop bind:showDropOverlay />
    {:else if $overlay !== null}
      <BlockingOverlayContainer
        bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
      >
        <div slot="title">
          <span class="font-bold">{$overlay?.title}</span>
        </div>
      </BlockingOverlayContainer>
    {/if}

    {#if $duplicateSourceName !== null}
      <DuplicateSource />
    {/if}

    <div
      class="index-body absolute w-screen h-screen"
      on:dragenter|preventDefault|stopPropagation
      on:dragleave|preventDefault|stopPropagation
      on:dragover|preventDefault|stopPropagation={(e) => {
        if (isEventWithFiles(e)) showDropOverlay = true;
      }}
      on:drag|preventDefault|stopPropagation
      on:drop|preventDefault|stopPropagation
    >
      <BasicLayout>
        <slot />
      </BasicLayout>
    </div>
  </div>
</QueryClientProvider>

<NotificationCenter />
