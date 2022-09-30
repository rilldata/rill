<script lang="ts">
  import { browser } from "$app/environment";
  import { page } from "$app/stores";
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    ApplicationStore,
    createStore,
    duplicateSourceName,
  } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { config } from "@rilldata/web-local/lib/application-state-stores/application-store.js";
  import {
    assetsVisible,
    assetVisibilityTween,
    importOverlayVisible,
    inspectorVisibilityTween,
    inspectorVisible,
    layout,
    quickStartDashboardOverlay,
    SIDE_PAD,
  } from "@rilldata/web-local/lib/application-state-stores/layout-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import {
    createDerivedModelStore,
    createPersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import { createQueryHighlightStore } from "@rilldata/web-local/lib/application-state-stores/query-highlight-store";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import {
    createDerivedTableStore,
    createPersistentTableStore,
  } from "@rilldata/web-local/lib/application-state-stores/table-stores";

  import AssetsSidebar from "@rilldata/web-local/lib/components/assets/index.svelte";
  import HideLeftSidebar from "@rilldata/web-local/lib/components/icons/HideLeftSidebar.svelte";
  import HideRightSidebar from "@rilldata/web-local/lib/components/icons/HideRightSidebar.svelte";
  import MoreHorizontal from "@rilldata/web-local/lib/components/icons/MoreHorizontal.svelte";
  import SurfaceViewIcon from "@rilldata/web-local/lib/components/icons/SurfaceView.svelte";
  import InspectorSidebar from "@rilldata/web-local/lib/components/inspector/index.svelte";
  import DuplicateSource from "@rilldata/web-local/lib/components/modal/DuplicateSource.svelte";
  import notificationStore from "@rilldata/web-local/lib/components/notifications/";
  import NotificationCenter from "@rilldata/web-local/lib/components/notifications/NotificationCenter.svelte";
  import ExportingDataset from "@rilldata/web-local/lib/components/overlay/ExportingDataset.svelte";
  import FileDrop from "@rilldata/web-local/lib/components/overlay/FileDrop.svelte";
  import ImportingTable from "@rilldata/web-local/lib/components/overlay/ImportingTable.svelte";
  import PreparingImport from "@rilldata/web-local/lib/components/overlay/PreparingImport.svelte";
  import QuickStartDashboard from "@rilldata/web-local/lib/components/overlay/QuickStartDashboard.svelte";
  import SurfaceControlButton from "@rilldata/web-local/lib/components/surface/SurfaceControlButton.svelte";
  import ConfigProvider from "@rilldata/web-local/lib/config/ConfigProvider.svelte";
  import { initMetrics } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { syncApplicationState } from "@rilldata/web-local/lib/redux-store/application/application-apis";
  import {
    createQueryClient,
    queryClient,
  } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import type { ApplicationMetadata } from "@rilldata/web-local/lib/types";
  import { QueryClientProvider } from "@sveltestack/svelte-query";
  import { getContext, onMount, setContext } from "svelte";
  import "../app.css";
  import "../fonts.css";

  let store;
  let queryHighlight = createQueryHighlightStore();

  const applicationMetadata: ApplicationMetadata = {
    version: RILL_VERSION, // constant defined in svelte.config.js
    commitHash: RILL_COMMIT, // constant defined in svelte.config.js
  };

  setContext("rill:app:metadata", applicationMetadata);

  if (browser) {
    store = createStore();
    setContext("rill:app:store", store);
    setContext("rill:app:query-highlight", queryHighlight);
    setContext(`rill:app:persistent-table-store`, createPersistentTableStore());
    setContext(`rill:app:derived-table-store`, createDerivedTableStore());
    setContext(`rill:app:persistent-model-store`, createPersistentModelStore());
    setContext(`rill:app:derived-model-store`, createDerivedModelStore());
    notificationStore.listenToSocket(store.socket);
    syncApplicationState(store);
  }

  createQueryClient();

  onMount(() => {
    initMetrics();
  });

  let dbRunState = "disconnected";
  let runstateTimer;

  function debounceRunstate(state) {
    if (runstateTimer) clearTimeout(runstateTimer);
    setTimeout(() => {
      dbRunState = state;
    }, 500);
  }

  $: debounceRunstate($store?.status || "disconnected");

  // FROM OLD INDEX.SVELTE

  let showDropOverlay = false;

  const app = getContext("rill:app:store") as ApplicationStore;

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
      bg: "bg-white",
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
      {/if}

      {#if $duplicateSourceName !== null}
        <DuplicateSource />
      {/if}

      <div
        class="index-body absolute w-screen h-screen bg-gray-100"
        on:drop|preventDefault|stopPropagation
        on:drag|preventDefault|stopPropagation
        on:dragenter|preventDefault|stopPropagation
        on:dragover|preventDefault|stopPropagation={(e) => {
          if (isEventWithFiles(e)) showDropOverlay = true;
        }}
        on:dragleave|preventDefault|stopPropagation
      >
        <!-- left assets pane expansion button -->
        <!-- make this the first element to select with tab by placing it first.-->
        <SurfaceControlButton
          show={true}
          left="{($layout.assetsWidth - 12 - 24) * (1 - $assetVisibilityTween) +
            12 * $assetVisibilityTween}px"
          on:click={() => {
            assetsVisible.set(!$assetsVisible);
          }}
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
          class="box-border	 assets fixed"
          aria-hidden={!$assetsVisible}
          style:left="{-$assetVisibilityTween * $layout.assetsWidth}px"
        >
          <AssetsSidebar />
        </div>

        <!-- workspace component -->
        <div
          class="box-border fixed {views[activeEntityType]?.bg ||
            'bg-gray-100'}"
          style:padding-left="{$assetVisibilityTween * SIDE_PAD}px"
          style:padding-right="{$inspectorVisibilityTween *
            SIDE_PAD *
            hasNoError *
            (hasInspector ? 1 : 0)}px"
          style:left="{$layout.assetsWidth * (1 - $assetVisibilityTween)}px"
          style:top="0px"
          style:right="{hasInspector && hasNoError
            ? $layout.inspectorWidth * (1 - $inspectorVisibilityTween)
            : 0}px"
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
