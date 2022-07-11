<script lang="ts">
  import { getContext } from "svelte";
  import Workspace from "./_surfaces/workspace/index.svelte";
  import InspectorSidebar from "./_surfaces/inspector/index.svelte";
  import AssetsSidebar from "./_surfaces/assets/index.svelte";

  import SurfaceViewIcon from "$lib/components/icons/SurfaceView.svelte";
  import SurfaceControlButton from "$lib/components/surface/SurfaceControlButton.svelte";

  import ImportingTable from "$lib/components/overlay/ImportingTable.svelte";
  import ExportingDataset from "$lib/components/overlay/ExportingDataset.svelte";
  import FileDrop from "$lib/components/overlay/FileDrop.svelte";

  import type {
    PersistentModelStore,
    DerivedModelStore,
  } from "$lib/application-state-stores/model-stores";
  import type {
    PersistentTableStore,
    DerivedTableStore,
  } from "$lib/application-state-stores/table-stores";

  import {
    ApplicationStore,
    config,
  } from "$lib/application-state-stores/application-store";
  import {
    layout,
    assetVisibilityTween,
    assetsVisible,
    inspectorVisibilityTween,
    inspectorVisible,
    SIDE_PAD,
    importOverlayVisible,
  } from "$lib/application-state-stores/layout-store";
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { HttpStreamClient } from "$lib/http-client/HttpStreamClient";
  import { store } from "$lib/redux-store/store-root";
  import PreparingImport from "$lib/components/overlay/PreparingImport.svelte";
  import DuplicateSource from "$lib/components/modal/DuplicateSource.svelte";

  let showDropOverlay = false;
  let assetsHovered = false;

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

  HttpStreamClient.create(`${config.server.serverUrl}/api`, store.dispatch);

  /** Workaround for hiding inspector for now. Post July 19 2022 we will remove this
   * in favor of ironing out more modular routing and suface management.
   */
  const views = {
    Source: {
      hasInspector: true,
    },
    Model: {
      hasInspector: true,
    },
    MetricsDefinition: {
      hasInspector: false,
    },
    MetricsLeaderboard: {
      hasInspector: false,
    },
  };

  $: activeEntityType = $app?.activeEntity?.type;
  $: hasInspector = activeEntityType
    ? views[activeEntityType].hasInspector
    : false;
  function isEventWithFiles(event: DragEvent) {
    let types = event.dataTransfer.types;
    return types && types.indexOf("Files") != -1;
  }
</script>

{#if derivedExportedModel && persistentExportedModel}
  <ExportingDataset tableName={persistentExportedModel.name} />
{:else if derivedImportedTable && persistentImportedTable}
  <ImportingTable
    importName={persistentImportedTable.path}
    tableName={persistentImportedTable.name}
  />
{:else if $importOverlayVisible}
  <PreparingImport />
{:else if showDropOverlay}
  <FileDrop bind:showDropOverlay />
{/if}

<DuplicateSource />

<div
  class="absolute w-screen h-screen bg-gray-100"
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
    show={assetsHovered || !$assetsVisible}
    left="{($layout.assetsWidth - 12 - 24) * (1 - $assetVisibilityTween) +
      12 * $assetVisibilityTween}px"
    on:click={() => {
      assetsVisible.set(!$assetsVisible);
    }}
  >
    <SurfaceViewIcon
      size="16px"
      mode={$assetsVisible ? "right" : "hamburger"}
    />
    <svelte:fragment slot="tooltip-content">
      {#if $assetVisibilityTween === 0} hide {:else} show {/if} models and sources
    </svelte:fragment>
  </SurfaceControlButton>

  <!-- assets sidebar component -->
  <!-- this is where we handle navigation -->
  <div
    class="box-border	 assets fixed"
    aria-hidden={!$assetsVisible}
    on:mouseover={() => {
      assetsHovered = true;
    }}
    on:mouseleave={() => {
      assetsHovered = false;
    }}
    on:focus={() => {
      assetsHovered = true;
    }}
    on:blur={() => {
      assetsHovered = false;
    }}
    style:left="{-$assetVisibilityTween * $layout.assetsWidth}px"
  >
    <AssetsSidebar />
  </div>

  <!-- workspace component -->
  <div
    class="box-border bg-gray-100 fixed"
    style:padding-left="{$assetVisibilityTween * SIDE_PAD}px"
    style:padding-right="{$inspectorVisibilityTween * SIDE_PAD}px"
    style:left="{$layout.assetsWidth * (1 - $assetVisibilityTween)}px"
    style:top="0px"
    style:right="{hasInspector
      ? $layout.inspectorWidth * (1 - $inspectorVisibilityTween)
      : 0}px"
  >
    <Workspace />
  </div>

  <!-- inspector sidebar -->
  <!-- Workaround: hide the inspector on MetricsDefinition or 
        on MetricsLeaderboard for now.
      Once we refactor how layout routing works, we will have a better solution to this.
  -->
  {#if hasInspector}
    <div
      class="fixed"
      aria-hidden={!$inspectorVisible}
      style:right="{$layout.inspectorWidth * (1 - $inspectorVisibilityTween)}px"
    >
      <InspectorSidebar />
    </div>
  {/if}
</div>
