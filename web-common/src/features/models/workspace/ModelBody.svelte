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
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useModel, useModelFileIsEmpty } from "../selectors";
  import { sanitizeQuery } from "../utils/sanitize-query";
  import Editor from "./Editor.svelte";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { quintOut } from "svelte/easing";

  const QUERY_DEBOUNCE_TIME = 400;

  export let modelName: string;
  export let focusEditorOnMount = false;

  const queryClient = useQueryClient();

  const queryHighlight: Writable<QueryHighlightState> = getContext(
    "rill:app:query-highlight",
  );

  const updateModel = createRuntimeServicePutFile();
  const limit = 150;

  let errors: string[] = [];

  $: runtimeInstanceId = $runtime.instanceId;
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelSqlQuery = createRuntimeServiceGetFile(runtimeInstanceId, modelPath);

  $: modelEmpty = useModelFileIsEmpty(runtimeInstanceId, modelName);

  $: modelSql = $modelSqlQuery?.data?.blob ?? "";
  $: hasModelSql = typeof modelSql === "string";

  $: modelQuery = useModel(runtimeInstanceId, modelName);

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

  $: workspaceLayout = $workspaces;

  $: tableHeight = workspaceLayout.table.height;
  $: tableVisible = workspaceLayout.table.visible;

  $: selections = $queryHighlight?.map((selection) => ({
    from: selection?.referenceIndex,
    to: selection?.referenceIndex + selection?.reference?.length,
  })) as SelectionRange[];

  $: {
    errors = [];
    // only add error if sql is present
    if (modelSql !== "") {
      if (modelError) errors.push(modelError);
      if (runtimeError) errors.push(runtimeError.message);
    }
  }

  async function updateModelContent(e: CustomEvent<{ content: string }>) {
    const { content } = e.detail;
    const hasChanged = sanitizeQuery(content) !== sanitizedQuery;

    try {
      if (hasChanged) {
        httpRequestQueue.removeByName(modelName);
        // cancel all existing analytical queries currently running.
        await queryClient.cancelQueries({
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

  const debounceUpdateModelContent = debounce(
    updateModelContent,
    QUERY_DEBOUNCE_TIME,
  );
</script>

<div class="editor-pane h-full overflow-hidden w-full flex flex-col">
  {#if hasModelSql}
    <div
      class="p-5 size-full flex-shrink-1 overflow-hidden"
      style:min-height="150px"
    >
      {#key modelName}
        <Editor
          content={modelSql}
          {selections}
          focusOnMount={focusEditorOnMount}
          on:write={debounceUpdateModelContent}
        />
      {/key}
    </div>
  {/if}

  {#if $tableVisible}
    <div
      class="p-5 w-full relative flex flex-none flex-col gap-2"
      style:height="{$tableHeight}px"
      style:max-height="75%"
      transition:slide={{ duration: 300, easing: quintOut }}
    >
      <Resizer
        max={600}
        direction="NS"
        side="top"
        bind:dimension={$tableHeight}
      >
        <HorizontalSplitter />
      </Resizer>
      <div
        class="table-wrapper"
        class:brightness-90={modelError || runtimeError}
      >
        {#if !$modelEmpty?.data}
          <ConnectedPreviewTable
            objectName={$modelQuery?.data?.model?.state?.table}
            loading={resourceIsLoading($modelQuery?.data)}
            {limit}
          />
        {/if}
      </div>
      {#if errors.length > 0}
        <div
          transition:slide={{ duration: 200 }}
          class="error bottom-4 break-words overflow-auto p-6 border-2 border-gray-300 font-bold text-gray-700 w-full shrink-0 max-h-[60%] z-10 bg-gray-100 flex flex-col gap-2"
        >
          {#each errors as error}
            <div>{error}</div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .table-wrapper {
    transition: filter 200ms;
    @apply relative rounded w-full overflow-hidden border-gray-200 border-2 h-full;
  }
</style>
