<script lang="ts">
  import type { SelectionRange } from "@codemirror/state";
  import Portal from "@rilldata/web-common/components/Portal.svelte";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import {
    useRuntimeServiceGetFile,
    useRuntimeServicePutFileAndReconcile,
    V1PutFileAndReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { SIDE_PAD } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import ConnectedPreviewTable from "@rilldata/web-local/lib/components/preview-table/ConnectedPreviewTable.svelte";
  import { drag } from "@rilldata/web-local/lib/drag";
  import {
    invalidateAfterReconcile,
    invalidationForProfileQueries,
  } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { modelIsEmpty } from "../utils/model-is-empty";
  import { sanitizeQuery } from "../utils/sanitize-query";
  import Editor from "./Editor.svelte";

  export let modelName: string;

  const queryClient = useQueryClient();

  const queryHighlight = getContext("rill:app:query-highlight");

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const updateModel = useRuntimeServicePutFileAndReconcile();

  // track innerHeight to calculate the size of the editor element.
  let innerHeight;

  let showPreview = true;
  let modelPath: string;
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;
  $: modelSqlQuery = useRuntimeServiceGetFile(runtimeInstanceId, modelPath);

  $: modelEmpty = modelIsEmpty(runtimeInstanceId, modelName);

  $: modelSql = $modelSqlQuery?.data?.blob;
  $: hasModelSql = typeof modelSql === "string";

  let sanitizedQuery: string;
  $: sanitizedQuery = sanitizeQuery(modelSql ?? "");

  const outputLayout = getContext("rill:app:output-layout");
  const outputPosition = getContext("rill:app:output-height-tween");
  const outputVisibilityTween = getContext(
    "rill:app:output-visibility-tween"
  ) as Writable<number>;

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

  async function updateModelContent(content: string) {
    const hasChanged = sanitizeQuery(content) !== sanitizedQuery;

    if (hasChanged) {
      httpRequestQueue.removeByName(modelName);
      // cancel all existing analytical queries currently running.
      await queryClient.cancelQueries({
        fetching: true,
        predicate: (query) => {
          return invalidationForProfileQueries(query.queryHash, modelName);
        },
      });
    }

    // TODO: why is the response type not present?
    const resp = (await $updateModel.mutateAsync({
      data: {
        instanceId: runtimeInstanceId,
        path: modelPath,
        blob: content,
      },
    })) as V1PutFileAndReconcileResponse;

    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
    if (!resp.errors.length && hasChanged) {
      sanitizedQuery = sanitizeQuery(content);
    }
    return invalidateAfterReconcile(
      queryClient,
      $runtimeStore.instanceId,
      resp
    );
  }

  $: selections = $queryHighlight?.map((selection) => ({
    from: selection.referenceIndex,
    to: selection.referenceIndex + selection.reference.length,
  })) as SelectionRange[];
</script>

<svelte:window bind:innerHeight />

<div class="editor-pane bg-gray-100">
  <div
    style:height="calc({innerHeight}px - {$outputPosition *
      $outputVisibilityTween}px - var(--header-height))"
  >
    {#if hasModelSql}
      <div class="h-full  p-5  grid overflow-auto">
        {#key modelName}
          <Editor
            {modelName}
            content={modelSql}
            {selections}
            on:write={(evt) => updateModelContent(evt.detail.content)}
          />
        {/key}
      </div>
    {/if}
  </div>
  <Portal target=".body">
    {#if $outputLayout.visible}
      <div
        class="fixed drawer-handler h-4 hover:cursor-col-resize translate-y-2 grid items-center"
        style:bottom="{$outputPosition * $outputVisibilityTween}px"
        style:left="{(1 - $navVisibilityTween) * $navigationWidth + 20}px"
        style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
        style:padding-right="{(1 - $inspectorVisibilityTween) * SIDE_PAD}px"
        style:right="{$inspectorVisibilityTween * $inspectorWidth + 20}px"
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
    {/if}
  </Portal>

  {#if hasModelSql}
    <div style:height="{$outputPosition}px" class="p-5 flex flex-col gap-6">
      <div
        class="rounded border border-gray-200 border-2 overflow-auto h-full grow-1 {!showPreview &&
          'hidden'}"
      >
        <div
          style="{modelError ? 'filter: brightness(.9);' : ''}
            transition: filter 200ms;
          "
          class="relative h-full"
        >
          {#if !$modelEmpty?.data}
            <ConnectedPreviewTable objectName={modelName} />
          {/if}
        </div>
        <!--TODO {:else}-->
        <!--  <div-->
        <!--    class="grid items-center justify-center pt-3 text-gray-600"-->
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
