<script lang="ts">
  import {
    duplicateSourceName,
    runtimeStore,
  } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import {
    importOverlayVisible,
    overlay,
    quickStartDashboardOverlay,
  } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import DuplicateSource from "@rilldata/web-local/lib/components/navigation/sources/DuplicateSource.svelte";
  import NotificationCenter from "@rilldata/web-local/lib/components/notifications/NotificationCenter.svelte";
  import FileDrop from "@rilldata/web-local/lib/components/overlay/FileDrop.svelte";
  import PreparingImport from "@rilldata/web-local/lib/components/overlay/PreparingImport.svelte";
  import QuickStartDashboard from "@rilldata/web-local/lib/components/overlay/QuickStartDashboard.svelte";
  import { initMetrics } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { createQueryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { fetchWrapperDirect } from "@rilldata/web-local/lib/util/fetchWrapper";
  import { QueryClientProvider } from "@sveltestack/svelte-query";
  import { onMount } from "svelte";
  import BlockingOverlayContainer from "../overlay/BlockingOverlayContainer.svelte";
  import BasicLayout from "./BasicLayout.svelte";

  const queryClient = createQueryClient();

  onMount(async () => {
    const localConfig = await fetchWrapperDirect(
      `${RuntimeUrl}/local/config`,
      "GET"
    );

    runtimeStore.set({
      instanceId: localConfig.instance_id,
    });

    return initMetrics();
  });

  let dbRunState = "disconnected";
  let runstateTimer;

  function debounceRunstate(state) {
    if (runstateTimer) clearTimeout(runstateTimer);
    setTimeout(() => {
      dbRunState = state;
    }, 500);
  }

  // FROM OLD INDEX.SVELTE

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
