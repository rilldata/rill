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
    layout,
    assetVisibilityTween,
    assetsVisible,
    inspectorVisibilityTween,
    inspectorVisible,
    SIDE_PAD,
    importOverlayVisible,
  } from "$lib/application-state-stores/layout-store";
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import PreparingImport from "$lib/components/overlay/PreparingImport.svelte";

  let showDropOverlay = false;
  let assetsHovered = false;

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

<div
  class="absolute w-screen h-screen bg-gray-100"
  on:drop|preventDefault|stopPropagation
  on:drag|preventDefault|stopPropagation
  on:dragenter|preventDefault|stopPropagation
  on:dragover|preventDefault|stopPropagation={() => {
    showDropOverlay = true;
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
    style:right="{$layout.inspectorWidth * (1 - $inspectorVisibilityTween)}px"
  >
    <Workspace />
  </div>

  <!-- inspector sidebar -->
  <div
    class="fixed"
    aria-hidden={!$inspectorVisible}
    style:right="{$layout.inspectorWidth * (1 - $inspectorVisibilityTween)}px"
  >
    <InspectorSidebar />
  </div>
</div>
