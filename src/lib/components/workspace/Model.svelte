<script lang="ts">
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import {
    assetVisibilityTween,
    inspectorVisibilityTween,
    layout,
    modelPreviewVisibilityTween,
    modelPreviewVisible,
    SIDE_PAD,
  } from "$lib/application-state-stores/layout-store";
  import Editor from "$lib/components/Editor.svelte";
  import { drag } from "$lib/drag";
  import { getContext } from "svelte";
  import { cubicOut as easing } from "svelte/easing";
  import { slide, fade } from "svelte/transition";

  import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import Portal from "$lib/components/Portal.svelte";
  import { PreviewTable } from "$lib/components/preview-table";
  import { updateModelQueryApi } from "$lib/redux-store/model/model-apis";

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
      class="p-6 "
    >
      <div
        class="rounded border border-gray-200 border-2  overflow-auto  h-full  {!showPreview &&
          'hidden'}"
        class:border={!!currentDerivedModel?.error}
        class:border-gray-300={!!currentDerivedModel?.error}
      >
        <div class="flex h-full flex-col">
          {#if currentDerivedModel?.preview && currentDerivedModel?.profile}
            <div class="relative flex-auto min-h-0">
              {#if currentDerivedModel?.error}
                <div
                  transition:fade={{ duration: 200 }}
                  style:background="rgba(0,0,0,0.2)"
                  class="absolute z-10 h-full w-full pointer-events-none z-100"
                />
              {/if}
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
              transition:slide={{ duration: 200, easing }}
              class="error border-2 border-gray-300 m-3 font-bold p-2 text-gray-700"
            >
              {currentDerivedModel.error}
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .editor-pane {
    height: calc(100vh - var(--header-height));
  }
</style>
