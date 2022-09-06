<script lang="ts">
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    ApplicationStore,
    duplicateSourceName,
  } from "$lib/application-state-stores/application-store";
  import {
    assetsVisible,
    assetVisibilityTween,
    importOverlayVisible,
    inspectorVisibilityTween,
    inspectorVisible,
    layout,
    quickStartDashboardOverlay,
    SIDE_PAD,
  } from "$lib/application-state-stores/layout-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import HideLeftSidebar from "$lib/components/icons/HideLeftSidebar.svelte";
  import HideRightSidebar from "$lib/components/icons/HideRightSidebar.svelte";
  import MoreHorizontal from "$lib/components/icons/MoreHorizontal.svelte";
  import SurfaceViewIcon from "$lib/components/icons/SurfaceView.svelte";
  import DuplicateSource from "$lib/components/modal/DuplicateSource.svelte";
  import ExportingDataset from "$lib/components/overlay/ExportingDataset.svelte";
  import FileDrop from "$lib/components/overlay/FileDrop.svelte";
  import ImportingTable from "$lib/components/overlay/ImportingTable.svelte";
  import PreparingImport from "$lib/components/overlay/PreparingImport.svelte";
  import QuickStartDashboard from "$lib/components/overlay/QuickStartDashboard.svelte";
  import SurfaceControlButton from "$lib/components/surface/SurfaceControlButton.svelte";
  import { getContext } from "svelte";
  import AssetsSidebar from "./_surfaces/assets/index.svelte";
  import InspectorSidebar from "./_surfaces/inspector/index.svelte";
  import Workspace from "./_surfaces/workspace/index.svelte";

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

  /** Workaround for hiding inspector for now. Post July 19 2022 we will remove this
   * in favor of ironing out more modular routing and suface management.
   */
  const views = {
    Table: {
      hasInspector: true,
      bg: "bg-gray-100",
    },
    Model: {
      hasInspector: true,
      bg: "bg-gray-100",
    },
    MetricsDefinition: {
      hasInspector: false,
      bg: "bg-gray-100",
    },
    MetricsExplorer: {
      hasInspector: false,
      bg: "bg-white",
    },
  };

  $: activeEntityType = $app?.activeEntity?.type;
  $: hasInspector = activeEntityType
    ? views[activeEntityType]?.hasInspector
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
  {#if hasInspector}
    <SurfaceControlButton
      show={true}
      right="{($layout.inspectorWidth - 12 - 24) *
        (1 - $inspectorVisibilityTween) +
        12 * $inspectorVisibilityTween}px"
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
    class="box-border fixed {views[activeEntityType]?.bg || 'bg-gray-100'}"
    style:padding-left="{$assetVisibilityTween * SIDE_PAD}px"
    style:padding-right="{$inspectorVisibilityTween *
      SIDE_PAD *
      (hasInspector ? 1 : 0)}px"
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
        on MetricsExplorer for now.
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
