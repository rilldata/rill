<script lang="ts">
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import type { ApplicationStore } from "../../application-state-stores/application-store";
  import {
    assetVisibilityTween,
    inspectorVisibilityTween,
    layout,
    modelPreviewVisibilityTween,
    modelPreviewVisible,
    SIDE_PAD,
  } from "../../application-state-stores/layout-store";
  import { drag } from "../../drag";
  import Editor from "../Editor.svelte";

  import type { DerivedModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import type { PersistentModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "../../application-state-stores/model-stores";
  import { updateModelQueryApi } from "../../redux-store/model/model-apis";
  import Portal from "../Portal.svelte";
  import { PreviewTable } from "../preview-table";

  const store = getContext("rill:app:store") as ApplicationStore;
  const queryHighlight = getContext("rill:app:query-highlight");
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let showPreview = true;

  let currentModel: PersistentModelEntity;
  $: activeEntityID = $store?.activeEntity?.id;
  $: currentModel =
    activeEntityID && $persistentModelStore?.entities
      ? $persistentModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  let currentDerivedModel: DerivedModelEntity;
  $: currentDerivedModel =
    activeEntityID && $derivedModelStore?.entities
      ? $derivedModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;

  // track innerHeight to calculate the size of the editor element.
  let innerHeight;
</script>

<svelte:window bind:innerHeight />

<div class="editor-pane bg-gray-100">
  <div
    style:height="calc({innerHeight}px - {(1 - $modelPreviewVisibilityTween) *
      $layout.modelPreviewHeight}px - var(--header-height))"
  >
    {#if $store && $persistentModelStore?.entities && $derivedModelStore?.entities && currentModel}
      <div class="h-full grid p-5 pt-0 overflow-auto">
        {#key currentModel?.id}
          <Editor
            content={currentModel.query}
            selections={$queryHighlight}
            on:write={(evt) =>
              updateModelQueryApi(currentModel.id, evt.detail.content)}
          />
        {/key}
      </div>
    {/if}
  </div>

  {#if $modelPreviewVisible}
    <Portal target=".body">
      <div
        class="fixed z-50 drawer-handler h-4 hover:cursor-col-resize translate-y-2 grid items-center ml-2 mr-2"
        style:bottom="{(1 - $modelPreviewVisibilityTween) *
          $layout.modelPreviewHeight}px"
        style:left="{(1 - $assetVisibilityTween) * $layout.assetsWidth + 16}px"
        style:right="{(1 - $inspectorVisibilityTween) * $layout.inspectorWidth +
          16}px"
        style:padding-left="{$assetVisibilityTween * SIDE_PAD}px"
        style:padding-right="{$inspectorVisibilityTween * SIDE_PAD}px"
        use:drag={{
          minSize: 200,
          maxSize: innerHeight - 200,
          side: "modelPreviewHeight",
          orientation: "vertical",
          reverse: true,
        }}
      >
        <div class="border-t border-gray-300" />
        <div class="absolute right-1/2 left-1/2 top-1/2 bottom-1/2">
          <div
            class="border-gray-400 border bg-white rounded h-1 w-8 absolute -translate-y-1/2"
          />
        </div>
      </div>
    </Portal>
  {/if}

  {#if currentModel}
    <div
      style:height="{(1 - $modelPreviewVisibilityTween) *
        $layout.modelPreviewHeight}px"
      class="p-6 grid"
      style:grid-template-rows="auto {currentDerivedModel?.error ? "auto" : ""}"
    >
      <div
        class="rounded border border-gray-200 border-2 overflow-auto h-full  {!showPreview &&
          'hidden'}"
        class:border={!!currentDerivedModel?.error}
        class:border-gray-300={!!currentDerivedModel?.error}
      >
        {#if currentDerivedModel?.preview && currentDerivedModel?.profile}
          <div
            style="{currentDerivedModel?.error ? 'filter: brightness(.9);' : ''}
            transition: filter 200ms;
          "
            class="relative h-full"
          >
            <PreviewTable
              rows={currentDerivedModel.preview}
              columnNames={currentDerivedModel.profile}
              rowOverscanAmount={20}
            />
          </div>
        {:else}
          <div
            class="grid items-center justify-center italic pt-3 text-gray-600"
          >
            no columns selected
          </div>
        {/if}
        {#if currentDerivedModel?.error}
          <div
            transition:slide={{ duration: 200 }}
            class="error break-words overflow-auto p-6 mb-3 border-t-2 border-gray-300 font-bold text-gray-700 sticky bottom-0 z-10 bg-gray-100"
          >
            {currentDerivedModel.error}
          </div>
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
