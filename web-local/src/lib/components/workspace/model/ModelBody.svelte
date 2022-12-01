<script lang="ts">
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    useRuntimeServiceRenameFileAndReconcile,
    V1PutFileAndReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { SIDE_PAD } from "@rilldata/web-local/lib/application-config";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import Editor from "@rilldata/web-local/lib/components/Editor.svelte";
  import Portal from "@rilldata/web-local/lib/components/Portal.svelte";
  import ConnectedPreviewTable from "@rilldata/web-local/lib/components/preview-table/ConnectedPreviewTable.svelte";
  import { drag } from "@rilldata/web-local/lib/drag";
  import { localStorageStore } from "@rilldata/web-local/lib/store-utils";
  import { renameFileArtifact } from "@rilldata/web-local/lib/svelte-query/actions";
  import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { getFileFromName } from "@rilldata/web-local/lib/util/entity-mappers";
  import { getContext } from "svelte";
  import { tweened } from "svelte/motion";
  import type { Writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { notifications } from "../../notifications";
  import WorkspaceHeader from "../core/WorkspaceHeader.svelte";

  export let modelName: string;

  const queryHighlight = getContext("rill:app:query-highlight");

  $: runtimeInstanceId = $runtimeStore.instanceId;
  $: getModel = useRuntimeServiceGetCatalogEntry(runtimeInstanceId, modelName);
  const updateModel = useRuntimeServicePutFileAndReconcile();
  const renameModel = useRuntimeServiceRenameFileAndReconcile();

  // track innerHeight to calculate the size of the editor element.
  let innerHeight;

  let showPreview = true;
  let modelPath: string;
  $: modelPath = getFileFromName(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;

  // TODO: does this need any sanitization?
  $: titleInput = modelName;

  function invalidateForModel(queryHash, modelName) {
    const r = new RegExp(
      `\\/v1/instances\\/[a-zA-Z0-9-]+\\/queries\\/[a-zA-Z0-9-]+\\/tables\\/${modelName}`
    );
    return r.test(queryHash);
  }

  function formatModelName(str) {
    return str?.trim().replaceAll(" ", "_").replace(/\.sql/, "");
  }

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Source name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.target.value = modelName; // resets the input
      return;
    }

    try {
      await renameFileArtifact(
        runtimeInstanceId,
        modelName,
        e.target.value,
        EntityType.Model,
        $renameModel
      );
    } catch (err) {
      console.error(err.response.data.message);
    }
  };

  /** model body layout elements */
  // TODO: should there be a session lived ID here instead of name?
  const outputLayout = localStorageStore(`${modelName}-output`, {
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
    "rilapp:inspector-visibility-tween"
  ) as Writable<number>;

  const navigationWidth = getContext(
    "rill:app:navigation-width-tween"
  ) as Writable<number>;

  const navVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Writable<number>;

  async function updateModelContent(content: string) {
    // cancel all existing analytical queries currently running.
    await queryClient.cancelQueries({
      fetching: true,
      predicate: (query) => {
        return invalidateForModel(query.queryHash, modelName);
      },
    });
    // TODO: why is the response type not present?
    const resp = (await $updateModel.mutateAsync({
      data: {
        instanceId: runtimeInstanceId,
        path: getFileFromName(modelName, EntityType.Model),
        blob: content,
      },
    })) as V1PutFileAndReconcileResponse;
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

    if (!resp.errors.length) {
      // re-fetch existing finished queries
      await queryClient.resetQueries({
        predicate: (query) => {
          return invalidateForModel(query.queryHash, modelName);
        },
      });
    }
  }
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
    {#if $getModel?.data?.entry?.model}
      <div class="h-full grid p-5 pt-0 overflow-auto">
        {#key modelName}
          <Editor
            content={$getModel?.data?.entry?.model?.sql}
            selections={$queryHighlight}
            on:write={(evt) => updateModelContent(evt.detail.content)}
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
      style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
      style:padding-right="{(1 - $inspectorVisibilityTween) * SIDE_PAD}px"
      style:right="{$inspectorVisibilityTween * $inspectorWidth + 16}px"
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

  {#if $getModel?.data?.entry}
    <div style:height="{$outputPosition}px" class="p-6 flex flex-col gap-6">
      <div
        class="rounded border border-gray-200 border-2 overflow-auto h-full grow-1 {!showPreview &&
          'hidden'}"
        class:border={!!modelError}
        class:border-gray-300={!!modelError}
      >
        <div
          style="{modelError ? 'filter: brightness(.9);' : ''}
            transition: filter 200ms;
          "
          class="relative h-full"
        >
          <ConnectedPreviewTable objectName={modelName} />
        </div>
        <!--TODO {:else}-->
        <!--  <div-->
        <!--    class="grid items-center justify-center italic pt-3 text-gray-600"-->
        <!--  >-->
        <!--    no columns selected-->
        <!--  </div>-->
        <!--{/if}-->
      </div>
      {#if modelError}
        <div
          transition:slide={{ duration: 200 }}
          class="error break-words overflow-auto p-6 border-2 border-gray-300 font-bold text-gray-700 w-full shrink-0 max-h-[60%] z-10 bg-gray-100"
        >
          {modelError}
        </div>
      {/if}
    </div>
  {/if}
</div>
