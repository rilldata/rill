<script lang="ts">
  import type { SelectionRange } from "@codemirror/state";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { getAllErrorsForFile } from "@rilldata/web-common/features/entity-management/resources-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import type { QueryHighlightState } from "@rilldata/web-common/features/models/query-highlight-store";
  import {
    createQueryServiceTableRows,
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import HorizontalSplitter from "../../../layout/workspace/HorizontalSplitter.svelte";
  import type { LayoutElement } from "../../../layout/workspace/types";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useModel, useModelFileIsEmpty } from "../selectors";
  import { sanitizeQuery } from "../utils/sanitize-query";
  import Editor from "./Editor.svelte";

  export let modelName: string;
  export let focusEditorOnMount = false;

  const queryClient = useQueryClient();

  const queryHighlight: Writable<QueryHighlightState> = getContext(
    "rill:app:query-highlight",
  );

  $: runtimeInstanceId = $runtime.instanceId;
  const updateModel = createRuntimeServicePutFile();

  const limit = 150;

  // track innerHeight to calculate the size of the editor element.
  let innerHeight: number;

  let showPreview = true;
  let modelPath: string;
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelSqlQuery = createRuntimeServiceGetFile(runtimeInstanceId, modelPath);

  $: modelEmpty = useModelFileIsEmpty(runtimeInstanceId, modelName);

  $: modelSql = $modelSqlQuery?.data?.blob ?? "";
  $: hasModelSql = typeof modelSql === "string";

  $: modelQuery = useModel(runtimeInstanceId, modelName);

  let sanitizedQuery: string;
  $: sanitizedQuery = sanitizeQuery(modelSql ?? "");

  $: allErrors = getAllErrorsForFile(
    queryClient,
    $runtime.instanceId,
    modelPath,
  );
  $: modelError = $allErrors?.[0]?.message;

  $: tableQuery = createQueryServiceTableRows(
    runtimeInstanceId,
    $modelQuery.data?.model?.state?.table ?? "",
    {
      limit,
    },
  );

  $: runtimeError = $tableQuery.error?.response.data;

  const outputLayout = getContext(
    "rill:app:output-layout",
  ) as Writable<LayoutElement>;
  const outputPosition = getContext(
    "rill:app:output-height-tween",
  ) as Writable<number>;
  const outputVisibilityTween = getContext(
    "rill:app:output-visibility-tween",
  ) as Writable<number>;

  async function updateModelContent(content: string) {
    const hasChanged = sanitizeQuery(content) !== sanitizedQuery;

    try {
      if (hasChanged) {
        httpRequestQueue.removeByName(modelName);
        // cancel all existing analytical queries currently running.
        await queryClient.cancelQueries({
          fetchStatus: "fetching",
          predicate: (query) => isProfilingQuery(query, modelName),
        });
      }

      await $updateModel.mutateAsync({
        instanceId: runtimeInstanceId,
        path: getFileAPIPathFromNameAndType(modelName, EntityType.Model),
        data: {
          blob: content,
        },
      });

      sanitizedQuery = sanitizeQuery(content);
    } catch (err) {
      console.error(err);
    }
  }
  $: selections = $queryHighlight?.map((selection) => ({
    from: selection?.referenceIndex,
    to: selection?.referenceIndex + selection?.reference?.length,
  })) as SelectionRange[];

  let errors: string[] = [];
  $: {
    errors = [];
    // only add error if sql is present
    if (modelSql !== "") {
      if (modelError) errors.push(modelError);
      if (runtimeError) errors.push(runtimeError.message);
    }
  }
</script>

<svelte:window bind:innerHeight />

<div class="editor-pane h-full pb-3">
  <div
    class="p-5"
    style:height="calc({innerHeight}px - {$outputPosition *
      $outputVisibilityTween}px - var(--header-height))"
  >
    {#if hasModelSql}
      <div class="h-full grid overflow-auto">
        {#key modelName}
          <Editor
            content={modelSql}
            {selections}
            focusOnMount={focusEditorOnMount}
            on:write={(evt) => updateModelContent(evt.detail.content)}
          />
        {/key}
      </div>
    {/if}
  </div>
  {#if $outputLayout.visible}
    <HorizontalSplitter className="px-5" />
  {/if}
  {#if hasModelSql}
    <div class="p-5 flex flex-col gap-6" style:height="{$outputPosition}px">
      <div
        class="rounded border border-gray-200 border-2 overflow-auto h-full grow-1 {!showPreview &&
          'hidden'}"
      >
        <div
          style="{modelError || runtimeError ? 'filter: brightness(.9);' : ''}
            transition: filter 200ms;
          "
          class="relative h-full"
        >
          {#if !$modelEmpty?.data}
            <ConnectedPreviewTable
              objectName={$modelQuery?.data?.model?.state?.table}
              loading={resourceIsLoading($modelQuery?.data)}
              {limit}
            />
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
      {#if errors.length > 0}
        <div
          transition:slide={{ duration: 200 }}
          class="error break-words overflow-auto p-6 border-2 border-gray-300 font-bold text-gray-700 w-full shrink-0 max-h-[60%] z-10 bg-gray-100 flex flex-col gap-2"
        >
          {#each errors as error}
            <div>{error}</div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
