<script lang="ts">
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import type { ApplicationStore } from "$lib/app-store";
  import { dataModelerService } from "$lib/app-store";
  import { cubicOut as easing } from "svelte/easing";
  import Editor from "$lib/components/Editor.svelte";
  import Header from "./Header.svelte"
  import { drag } from "$lib/drag";
  import Table from "$lib/components/icons/Parquet.svelte"
  import { modelPreviewVisibilityTween, modelPreviewVisible, layout, assetVisibilityTween, inspectorVisibilityTween, SIDE_PAD } from "$lib/layout-store";
  
  import PreviewTable from "$lib/components/table/PreviewTable.svelte";
  import type {
      PersistentTableEntity
  } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
  import type {
      DerivedTableEntity
  } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
  import { DerivedTableStore, PersistentTableStore } from "$lib/tableSTores";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import Portal from "$lib/components/Portal.svelte";
  
  const store = getContext("rill:app:store") as ApplicationStore;
  const queryHighlight = getContext("rill:app:query-highlight");
  const persistentTableStore = getContext('rill:app:persistent-table-store') as PersistentTableStore;
  const derivedTableStore = getContext('rill:app:derived-table-store') as DerivedTableStore;
  
  let errorLineNumber;
  
  let showPreview = true;
  
  let currentTable: PersistentTableEntity;
  $: currentTable = ($store?.activeEntity && $persistentTableStore?.entities) ?
      $persistentTableStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
  let currentDerivedTable: DerivedTableEntity;
  $: currentDerivedTable = ($store?.activeEntity && $derivedTableStore?.entities) ?
      $derivedTableStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
  // track innerHeight to calculate the size of the editor element.
  let innerHeight;
  
  </script>
  
  <svelte:window bind:innerHeight />
  <Header currentEntity={currentTable} icon={Table} />

  <div class="editor-pane bg-gray-100">
    <div  
      style:height="calc({innerHeight}px - {(1 - $modelPreviewVisibilityTween) * $layout.modelPreviewHeight}px - var(--header-height))"
    >
    {#if $store && $persistentTableStore?.entities && $derivedTableStore?.entities && currentTable}
      <div class="h-full grid p-5 pt-0 overflow-auto">
        {#key currentTable?.id}
          <Editor 
            editable={false}
            content={`SELECT * from ${currentTable.name}`}
            name={currentTable.name}
            selections={$queryHighlight}
            errorLineNumber={ currentTable.id === $store.activeEntity.id ? errorLineNumber : undefined }
            on:release-focus={() => {
              //dataModelerService.dispatch('releaseActiveQueryFocus', [{ id: q.id }]);
            }}
            on:model-profile={() => {
              //dataModelerService.dispatch('computeModelProfile', [{ id: currentQuery.id }]);
            }}
            on:rename={(evt) => {
              dataModelerService.dispatch('updateModelName', [currentTable.id, evt.detail]);
            }}
            on:write={(evt) => {
                // dataModelerService.dispatch('setActiveAsset', [EntityType.Model, currentTable.id]);
                // dataModelerService.dispatch('updateModelQuery', [currentTable.id, evt.detail.content])
            }}
        />
      {/key}
      </div>
    {/if}
  </div>
  
  {#if $modelPreviewVisible}
  <Portal target=".body">
    <div
    class='fixed z-50 drawer-handler h-4 hover:cursor-col-resize translate-y-2 grid items-center ml-2 mr-2'
    style:bottom="{(1 - $modelPreviewVisibilityTween) * $layout.modelPreviewHeight}px"
    style:left="{(1 - $assetVisibilityTween) * $layout.assetsWidth + 16}px"
    style:right="{(1 - $inspectorVisibilityTween) * $layout.inspectorWidth + 16}px"
    style:padding-left="{($assetVisibilityTween * SIDE_PAD)}px"
    style:padding-right="{($inspectorVisibilityTween * SIDE_PAD)}px"
    use:drag={{ minSize: 200, maxSize: innerHeight - 200,  side: 'modelPreviewHeight', orientation: "vertical", reverse: true  }}>
      <div class="border-t border-gray-300" />
      <div class="absolute right-1/2 left-1/2 top-1/2 bottom-1/2">
        <div class="border-gray-400 border bg-white rounded h-1 w-8 absolute -translate-y-1/2" />
      </div>
  </div>
  </Portal>
  {/if}
  
  {#if currentTable}
      <div
        style:height="{(1 - $modelPreviewVisibilityTween) * $layout.modelPreviewHeight}px"
        class="p-6 "
      >
      <div class="rounded border border-gray-200 border-2  overflow-auto  h-full  {!showPreview && 'hidden'}"
        class:border={!!currentDerivedTable?.error}
      class:border-gray-300={!!currentDerivedTable?.error}
        >
        {#if currentDerivedTable?.error}
        <!-- <div 
          transition:slide={{ duration: 200, easing }} 
          class="error font-bold rounded-lg p-5 text-gray-700"
        >
          {currentDerivedTable.error} 
        </div> -->
        {:else if currentDerivedTable?.preview && currentDerivedTable?.profile}
          <PreviewTable rows={currentDerivedTable.preview} columnNames={currentDerivedTable.profile} />
        {:else}
          <div class="grid items-center justify-center italic">no columns selected</div>
        {/if}
      </div>
    </div>
    {/if}
  </div>
  <style>
  
  .editor-pane {
    height: calc(100vh - var(--header-height));
  }
  </style>
  