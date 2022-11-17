<script lang="ts">
  import {
    useRuntimeServiceGetCatalogObject,
    useRuntimeServicePutFileAndMigrate,
    useRuntimeServiceRenameFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import { SIDE_PAD } from "@rilldata/web-local/lib/application-config";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import Editor from "@rilldata/web-local/lib/components/Editor.svelte";
  import Portal from "@rilldata/web-local/lib/components/Portal.svelte";
  import { PreviewTable } from "@rilldata/web-local/lib/components/preview-table";
  import { drag } from "@rilldata/web-local/lib/drag";
  import { localStorageStore } from "@rilldata/web-local/lib/store-utils";
  import { getContext } from "svelte";
  import { tweened } from "svelte/motion";
  import type { Writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import WorkspaceHeader from "../core/WorkspaceHeader.svelte";

  export let modelName: string;

  const queryHighlight = getContext("rill:app:query-highlight");
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  $: getModel = useRuntimeServiceGetCatalogObject(
    $runtimeStore.instanceId,
    modelName
  );
  const updateModel = useRuntimeServicePutFileAndMigrate();
  const renameModel = useRuntimeServiceRenameFileAndMigrate();

  $: currentModel = $persistentModelStore?.entities
    ? $persistentModelStore.entities.find((q) => q.tableName === modelName)
    : undefined;

  $: currentDerivedModel = $derivedModelStore?.entities
    ? $derivedModelStore.entities.find((q) => q.id === currentModel.id)
    : undefined;

  // track innerHeight to calculate the size of the editor element.
  let innerHeight;

  let showPreview = true;

  let titleInput = currentModel?.name;
  $: titleInput = currentModel?.name;

  function formatModelName(str) {
    return str?.trim().replaceAll(" ", "_").replace(/\.sql/, "");
  }

  const onChangeCallback = async (e) => {
    // CHECK: do I have to rename the entity in the Node backend too?
    $renameModel.mutate(
      {
        data: {
          repoId: $runtimeStore.repoId,
          instanceId: $runtimeStore.instanceId,
          fromPath: `models/${modelName}.sql`,
          toPath: `models/${e.target.value}.sql`,
        },
      },
      {
        onError: (err) => {
          console.error(err.response.data.message);
        },
      }
    );
  };

  /** model body layout elements */
  const outputLayout = localStorageStore(`${currentModel.id}-output`, {
    value: 500,
    visible: true,
  });
  const outputPosition = tweened($outputLayout.value, { duration: 50 });
  outputLayout.subscribe((state) => {
    outputPosition.set(state.value);
  });

  const inspectorWidth = getContext(
    "rill:app:inspector-width-tween"
  ) as Writable<number>;

  const inspectorVisibilityTween = getContext(
    "rill:app:inspector-visibility-tween"
  ) as Writable<number>;

  const navigationWidth = getContext(
    "rill:app:navigation-width-tween"
  ) as Writable<number>;

  const navVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Writable<number>;
</script>

<svelte:window bind:innerHeight />

<WorkspaceHeader
  {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
/>

<div class="editor-pane bg-gray-100">
  <div
    style:height="calc({innerHeight}px - {$outputPosition}px -
    var(--header-height))"
  >
    {#if $persistentModelStore?.entities && $derivedModelStore?.entities && currentModel}
      <div class="h-full grid p-5 pt-0 overflow-auto">
        {#key currentModel?.id}
          <Editor
            content={currentModel.query}
            selections={$queryHighlight}
            on:write={(evt) =>
              $updateModel.mutate(
                {
                  data: {
                    repoId: $runtimeStore.repoId,
                    instanceId: $runtimeStore.instanceId,
                    path: `models/${currentModel.tableName}`,
                    blob: evt.detail.content,
                  },
                },
                {
                  onError: (err) => {
                    console.error(err.response.data.message);
                  },
                }
              )}
          />
        {/key}
      </div>
    {/if}
  </div>

  <Portal target=".body">
    <div
      class="fixed drawer-handler h-4 hover:cursor-col-resize translate-y-2 grid items-center ml-2 mr-2"
      style:bottom="{$outputPosition}px"
      style:left="{(1 - $navVisibilityTween) * $navigationWidth + 16}px"
      style:right="{$inspectorVisibilityTween * $inspectorWidth + 16}px"
      style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
      style:padding-right="{(1 - $inspectorVisibilityTween) * SIDE_PAD}px"
      use:drag={{
        minSize: 200,
        maxSize: innerHeight - 200,
        side: "modelPreviewHeight",
        store: outputLayout,
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

  {#if currentModel}
    <div style:height="{$outputPosition}px" class="p-6 flex flex-col gap-6">
      <div
        class="rounded border border-gray-200 border-2 overflow-auto h-full grow-1 {!showPreview &&
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
      </div>
      {#if currentDerivedModel?.error}
        <div
          transition:slide={{ duration: 200 }}
          class="error break-words overflow-auto p-6 border-2 border-gray-300 font-bold text-gray-700 w-full shrink-0 max-h-[60%] z-10 bg-gray-100"
        >
          {currentDerivedModel.error}
        </div>
      {/if}
    </div>
  {/if}
</div>
