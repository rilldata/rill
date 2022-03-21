<script lang="ts">
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import type { ApplicationStore } from "$lib/app-store";
import { dataModelerService } from "$lib/app-store";
import { cubicOut as easing } from "svelte/easing";
import Editor from "$lib/components/Editor.svelte";
import { drag } from "$lib/drag";
import { modelPreviewVisibilityTween, modelPreviewVisible, layout, assetVisibilityTween, inspectorVisibilityTween, SIDE_PAD } from "$lib/layout-store";

import PreviewTable from "$lib/components/table/PreviewTable.svelte";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
    DerivedModelEntity
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { DerivedModelStore, PersistentModelStore } from "$lib/modelStores";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import Portal from "$lib/components/Portal.svelte";

const store = getContext("rill:app:store") as ApplicationStore;
const queryHighlight = getContext("rill:app:query-highlight");
const persistentModelStore = getContext('rill:app:persistent-model-store') as PersistentModelStore;
const derivedModelStore = getContext('rill:app:derived-model-store') as DerivedModelStore;

let errorLineNumber;

let showPreview = true;

let currentModel: PersistentModelEntity;
$: currentModel = ($store?.activeEntity && $persistentModelStore?.entities) ?
    $persistentModelStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
let currentDerivedModel: DerivedModelEntity;
$: currentDerivedModel = ($store?.activeEntity && $derivedModelStore?.entities) ?
    $derivedModelStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;

// track innerHeight to calculate the size of the editor element.
let innerHeight;

</script>

<svelte:window bind:innerHeight />

<div class="editor-pane">
  <div  
    style:height="calc({innerHeight}px - {(1 - $modelPreviewVisibilityTween) * $layout.modelPreviewHeight}px - var(--header-height))"
  >
  {#if $store && $persistentModelStore?.entities && $derivedModelStore?.entities && currentModel}
    <div class="h-full grid p-5 pt-0 overflow-auto">
      {#key currentModel?.id}
        <Editor 
          content={currentModel.query}
          name={currentModel.name}
          selections={$queryHighlight}
          errorLineNumber={ currentModel.id === $store.activeEntity.id ? errorLineNumber : undefined }
          on:down={() => { dataModelerService.dispatch('moveModelDown', [currentModel.id]); }}
          on:up={() => { dataModelerService.dispatch('moveModelUp', [currentModel.id]); }}
          on:delete={() => { dataModelerService.dispatch('deleteModel', [currentModel.id]); }}
          on:receive-focus={() => {
              dataModelerService.dispatch('setActiveAsset', [EntityType.Model, currentModel.id]);
          }}
          on:release-focus={() => {
            //dataModelerService.dispatch('releaseActiveQueryFocus', [{ id: q.id }]);
          }}
          on:model-profile={() => {
            //dataModelerService.dispatch('computeModelProfile', [{ id: currentQuery.id }]);
          }}
          on:rename={(evt) => {
            dataModelerService.dispatch('updateModelName', [currentModel.id, evt.detail]);
          }}
          on:write={(evt) => {
              dataModelerService.dispatch('setActiveAsset', [EntityType.Model, currentModel.id]);
              dataModelerService.dispatch('updateModelQuery', [currentModel.id, evt.detail.content])
          }}
      />
    {/key}
    </div>
  {/if}
</div>

{#if $modelPreviewVisible}
<Portal>
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

{#if currentModel}
    <div
      style:height="{(1 - $modelPreviewVisibilityTween) * $layout.modelPreviewHeight}px"
      class="p-6 "
    >
    <div class="rounded border border-gray-200 border-2  overflow-auto  h-full  {!showPreview && 'hidden'}"
     class:border={!!currentDerivedModel?.error}
    class:border-gray-300={!!currentDerivedModel?.error}
     >
      {#if currentDerivedModel?.error}
      <div 
        transition:slide={{ duration: 200, easing }} 
        class="error font-bold rounded-lg p-5 text-gray-700"
      >
        {currentDerivedModel.error}
      </div>
      {:else if currentDerivedModel?.preview && currentDerivedModel?.profile}
        <PreviewTable rows={currentDerivedModel.preview} columnNames={currentDerivedModel.profile} />
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
