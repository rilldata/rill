<script lang="ts">
  import type { SelectionRange } from "@codemirror/state";
  import {
    useRuntimeServiceGetFile,
    useRuntimeServicePutFileAndReconcile,
    useRuntimeServiceRenameFileAndReconcile,
    V1PutFileAndReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { SIDE_PAD } from "@rilldata/web-local/lib/application-config";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import Editor from "@rilldata/web-local/lib/components/Editor.svelte";
  import Portal from "@rilldata/web-local/lib/components/Portal.svelte";
  import ConnectedPreviewTable from "@rilldata/web-local/lib/components/preview-table/ConnectedPreviewTable.svelte";
  import { drag } from "@rilldata/web-local/lib/drag";
  import {
    invalidateAfterReconcile,
    invalidationForProfileQueries,
  } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { sanitizeQuery } from "@rilldata/web-local/lib/util/sanitize-query";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { fade, slide } from "svelte/transition";
  import { runtimeStore } from "../../../application-state-stores/application-store";

  export let modelName: string;

  const queryClient = useQueryClient();

  const queryHighlight = getContext("rill:app:query-highlight");

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const updateModel = useRuntimeServicePutFileAndReconcile();
  const renameModel = useRuntimeServiceRenameFileAndReconcile();

  // track innerHeight to calculate the size of the editor element.
  let innerHeight;

  let showPreview = true;
  let modelPath: string;
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;
  $: modelSqlQuery = useRuntimeServiceGetFile(runtimeInstanceId, modelPath);

  $: modelSql = $modelSqlQuery?.data?.blob;
  $: hasModelSql = typeof modelSql === "string";

  let sanitizedQuery: string;
  $: sanitizedQuery = sanitizeQuery(modelSql ?? "");

  /** model body layout elements */
  // TODO: should there be a session lived ID here instead of name?
  // const outputLayout = localStorageStore(`${modelName}-output`, {
  //   value: 500,
  //   visible: true,
  // });
  // const outputPosition = tweened($outputLayout.value, { duration: 50 });
  // outputLayout.subscribe((state) => {
  //   outputPosition.set(state.value);
  // });

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
    } else {
      resp.affectedPaths = resp.affectedPaths.filter(
        (affectedPath) => affectedPath !== modelPath
      );
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
      <div class="h-full pb-5 grid overflow-auto">
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
        transition:fade
        class="fixed drawer-handler h-4 hover:cursor-col-resize translate-y-2 grid items-center"
        style:bottom="{$outputPosition * $outputVisibilityTween}px"
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
    {/if}
  </Portal>

  {#if hasModelSql}
    <div style:height="{$outputPosition}px" class="flex flex-col gap-6">
      <div
        class="rounded overflow-auto h-full grow-1 {!showPreview && 'hidden'}"
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
