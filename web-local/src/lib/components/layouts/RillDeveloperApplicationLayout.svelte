<script lang="ts">
  import { page } from "$app/stores";
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    ApplicationStore,
    duplicateSourceName,
    runtimeStore,
  } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { config } from "@rilldata/web-local/lib/application-state-stores/application-store.js";
  import {
    assetsVisible,
    assetVisibilityTween,
    importOverlayVisible,
    inspectorVisibilityTween,
    inspectorVisible,
    layout,
    overlay,
    quickStartDashboardOverlay,
    SIDE_PAD,
  } from "@rilldata/web-local/lib/application-state-stores/layout-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import AssetsSidebar from "@rilldata/web-local/lib/components/assets/index.svelte";
  import DuplicateSource from "@rilldata/web-local/lib/components/assets/sources/DuplicateSource.svelte";
  import HideLeftSidebar from "@rilldata/web-local/lib/components/icons/HideLeftSidebar.svelte";
  import HideRightSidebar from "@rilldata/web-local/lib/components/icons/HideRightSidebar.svelte";
  import MoreHorizontal from "@rilldata/web-local/lib/components/icons/MoreHorizontal.svelte";
  import SurfaceViewIcon from "@rilldata/web-local/lib/components/icons/SurfaceView.svelte";
  import InspectorSidebar from "@rilldata/web-local/lib/components/inspector/index.svelte";
  import NotificationCenter from "@rilldata/web-local/lib/components/notifications/NotificationCenter.svelte";
  import ExportingDataset from "@rilldata/web-local/lib/components/overlay/ExportingDataset.svelte";
  import FileDrop from "@rilldata/web-local/lib/components/overlay/FileDrop.svelte";
  import ImportingTable from "@rilldata/web-local/lib/components/overlay/ImportingTable.svelte";
  import PreparingImport from "@rilldata/web-local/lib/components/overlay/PreparingImport.svelte";
  import QuickStartDashboard from "@rilldata/web-local/lib/components/overlay/QuickStartDashboard.svelte";
  import SurfaceControlButton from "@rilldata/web-local/lib/components/surface/SurfaceControlButton.svelte";
  import ConfigProvider from "@rilldata/web-local/lib/config/ConfigProvider.svelte";
  import { initMetrics } from "@rilldata/web-local/lib/metrics/initMetrics";
  import {
    createQueryClient,
    queryClient,
  } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { fetchWrapper } from "@rilldata/web-local/lib/util/fetchWrapper";
  import { QueryClientProvider } from "@sveltestack/svelte-query";
  import { getContext, onMount } from "svelte";
  import BlockingOverlayContainer from "../overlay/BlockingOverlayContainer.svelte";
  createQueryClient();

  onMount(async () => {
    const instanceResp = await fetchWrapper("v1/runtime/instance-id", "GET");

    runtimeStore.set({
      instanceId: instanceResp.instanceId,
      repoId: instanceResp.repoId,
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

  const app = getContext("rill:app:store") as ApplicationStore;
  $: debounceRunstate($app?.status || "disconnected");

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  // get any importing tables
  $: derivedImportedTable = $derivedTableStore?.entities?.find(
    (table) => table.status === EntityStatus.Importing
  );
  $: persistentImportedTable = $persistentTableStore?.entities?.find(
    (table) => table.id === derivedImportedTable?.id
  );
  // get any exporting datasets.
  $: derivedExportedModel = $derivedModelStore?.entities?.find(
    (model) => model.status === EntityStatus.Exporting
  );
  $: persistentExportedModel = $persistentModelStore?.entities?.find(
    (model) => model.id === derivedExportedModel?.id
  );

  // consider moving this logic to the individual pages
  const views = {
    Table: {
      bg: "bg-gray-100",
    },
    Model: {
      bg: "bg-gray-100",
    },
    MetricsDefinition: {
      bg: "bg-gray-100",
    },
    MetricsExplorer: {
      bg: "surface",
    },
  };

  $: activeEntityType = $app?.activeEntity?.type;

  const routesWithAnInspector = ["/source/", "/model/"];
  $: hasInspector = routesWithAnInspector.some((route: string) =>
    $page.url.pathname.includes(route)
  );
  function isEventWithFiles(event: DragEvent) {
    let types = event.dataTransfer.types;
    return types && types.indexOf("Files") != -1;
  }

  /** workaround for hiding inspector when there's a page error.
   * We should refactor this inspector to work with a named slot instead of this current approach.
   */
  $: hasNoError = $page.status < 400 ? 1 : 0;
</script>

<QueryClientProvider client={queryClient}>
  <ConfigProvider {config}>
    <div class="body">
      {#if derivedExportedModel && persistentExportedModel}
        <ExportingDataset tableName={persistentExportedModel.name} />
      {:else if derivedImportedTable && persistentImportedTable}
        <ImportingTable
          importName={persistentImportedTable.path}
          tableName={persistentImportedTable.name}
        />
      {:else if $importOverlayVisible}
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
        class="index-body absolute w-screen h-screen bg-gray-100"
        on:dragenter|preventDefault|stopPropagation
        on:dragleave|preventDefault|stopPropagation
        on:dragover|preventDefault|stopPropagation={(e) => {
          if (isEventWithFiles(e)) showDropOverlay = true;
        }}
        on:drag|preventDefault|stopPropagation
        on:drop|preventDefault|stopPropagation
      >
        <!-- left assets pane expansion button -->
        <!-- make this the first element to select with tab by placing it first.-->
        <SurfaceControlButton
          left="{($layout.assetsWidth - 12 - 24) * (1 - $assetVisibilityTween) +
            12 * $assetVisibilityTween}px"
          on:click={() => {
            assetsVisible.set(!$assetsVisible);
          }}
          show={true}
        >
          {#if $assetsVisible}
            <HideLeftSidebar size="20px" />
          {:else}
            <SurfaceViewIcon size="16px" mode={"hamburger"} />
          {/if}
          <svelte:fragment slot="tooltip-content">
            {#if $assetVisibilityTween === 0} close {:else} show {/if} sidebar
          </svelte:fragment>
        </SurfaceControlButton>

        <!-- inspector pane hide -->
        {#if hasInspector && hasNoError}
          <SurfaceControlButton
            show={true}
            right="{($layout.inspectorWidth - 12 - 24) *
              (1 - $inspectorVisibilityTween * hasNoError) +
              12 * $inspectorVisibilityTween * hasNoError}px"
            on:click={() => {
              inspectorVisible.set(!$inspectorVisible);
            }}
          >
            {#if $inspectorVisible}
              <HideRightSidebar size="20px" />
            {:else}
              <MoreHorizontal size="16px" />
            {/if}
            <svelte:fragment slot="tooltip-content">
              {#if $assetVisibilityTween === 0} close {:else} show {/if} sidebar
            </svelte:fragment>
          </SurfaceControlButton>
        {/if}
        <!-- assets sidebar component -->
        <!-- this is where we handle navigation -->
        <div
          aria-hidden={!$assetsVisible}
          class="box-border	 assets fixed"
          style:left="{-$assetVisibilityTween * $layout.assetsWidth}px"
        >
          <AssetsSidebar />
        </div>

        <!-- workspace component -->
        <div
          class="box-border fixed {views[activeEntityType]?.bg ||
            'bg-gray-100'}"
          style:left="{$layout.assetsWidth * (1 - $assetVisibilityTween)}px"
          style:padding-left="{$assetVisibilityTween * SIDE_PAD}px"
          style:padding-right="{$inspectorVisibilityTween *
            SIDE_PAD *
            hasNoError *
            (hasInspector ? 1 : 0)}px"
          style:right="{hasInspector && hasNoError
            ? $layout.inspectorWidth * (1 - $inspectorVisibilityTween)
            : 0}px"
          style:top="0px"
        >
          <slot />
        </div>

        <!-- inspector sidebar -->
        <!-- Workaround: hide the inspector on MetricsDefinition or 
            on MetricsExplorer for now.
          Once we refactor how layout routing works, we will have a better solution to this.
      -->
        {#if hasInspector && hasNoError}
          <div
            class="fixed"
            aria-hidden={!$inspectorVisible}
            style:right="{$layout.inspectorWidth *
              (1 - $inspectorVisibilityTween)}px"
          >
            <InspectorSidebar />
          </div>
        {/if}
      </div>
    </div>
  </ConfigProvider>
</QueryClientProvider>

<NotificationCenter />
