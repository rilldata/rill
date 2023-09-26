<script lang="ts">
  import type { SelectionRange } from "@codemirror/state";
  import Portal from "@rilldata/web-common/components/Portal.svelte";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import type { QueryHighlightState } from "@rilldata/web-common/features/models/query-highlight-store";
  import {
    embeddedSourcesError,
    filterKnownEmbeddedSources,
  } from "@rilldata/web-common/features/models/utils/embedded";
  import {
    getEmbeddedReferences,
    Reference,
  } from "@rilldata/web-common/features/models/utils/get-table-references";
  import { useEmbeddedSources } from "@rilldata/web-common/features/sources/selectors";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import {
    createQueryServiceTableRows,
    createRuntimeServiceGetFile,
    createRuntimeServicePutFileAndReconcile,
    getRuntimeServiceGetFileQueryKey,
    V1CatalogEntry,
    V1PutFileAndReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
  import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import HorizontalSplitter from "../../../layout/workspace/HorizontalSplitter.svelte";
  import type { LayoutElement } from "../../../layout/workspace/types";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useModelFileIsEmpty } from "../selectors";
  import { sanitizeQuery } from "../utils/sanitize-query";
  import Editor from "./Editor.svelte";

  export let modelName: string;
  export let focusEditorOnMount = false;

  const queryClient = useQueryClient();

  const queryHighlight: Writable<QueryHighlightState> = getContext(
    "rill:app:query-highlight"
  );

  $: runtimeInstanceId = $runtime.instanceId;
  const updateModel = createRuntimeServicePutFileAndReconcile();

  const limit = 150;

  $: tableQuery = createQueryServiceTableRows(runtimeInstanceId, modelName, {
    limit,
  });

  $: runtimeError = ($tableQuery.error as any)?.response.data;

  // track innerHeight to calculate the size of the editor element.
  let innerHeight: number;

  let showPreview = true;
  let modelPath: string;
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;
  $: modelSqlQuery = createRuntimeServiceGetFile(runtimeInstanceId, modelPath);

  $: modelEmpty = useModelFileIsEmpty(runtimeInstanceId, modelName);

  $: modelSql = $modelSqlQuery?.data?.blob;
  $: hasModelSql = typeof modelSql === "string";

  let sanitizedQuery: string;
  $: sanitizedQuery = sanitizeQuery(modelSql ?? "");

  $: sourceCatalogsQuery = useEmbeddedSources($runtime?.instanceId);
  let embeddedSourceCatalogs: Map<string, V1CatalogEntry>;
  $: embeddedSourceCatalogs = getMapFromArray(
    $sourceCatalogsQuery?.data ?? [],
    (entity) => entity.source.properties.path?.toLowerCase()
  ) as Map<string, V1CatalogEntry>;

  let embeddedSourceErrors: Array<string>;

  const outputLayout = getContext(
    "rill:app:output-layout"
  ) as Writable<LayoutElement>;
  const outputPosition = getContext(
    "rill:app:output-height-tween"
  ) as Writable<number>;
  const outputVisibilityTween = getContext(
    "rill:app:output-visibility-tween"
  ) as Writable<number>;

  async function updateModelContent(content: string) {
    const hasChanged = sanitizeQuery(content) !== sanitizedQuery;
    let overlayShown = false;
    let embeddedSources: Array<Reference> = [];

    try {
      if (hasChanged) {
        embeddedSources = getEmbeddedReferences(sanitizeQuery(content));
        const unknownEmbeddedSources = filterKnownEmbeddedSources(
          embeddedSources,
          embeddedSourceCatalogs
        );
        if (unknownEmbeddedSources.length > 0) {
          overlay.set({
            title: `Caching ${unknownEmbeddedSources.join(",")}`,
          });
          overlayShown = true;
        }

        httpRequestQueue.removeByName(modelName);
        // cancel all existing analytical queries currently running.
        await queryClient.cancelQueries({
          fetchStatus: "fetching",
          predicate: (query) => isProfilingQuery(query, modelName),
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

      embeddedSourceErrors = embeddedSourcesError(resp.errors, embeddedSources);
      fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
      if (!resp.errors.length && hasChanged) {
        sanitizedQuery = sanitizeQuery(content);
      }
      await invalidateAfterReconcile(queryClient, $runtime.instanceId, resp);
      if (resp.affectedPaths.length === 0) {
        // when backend detects no change, we need to invalidate the file
        await queryClient.refetchQueries(
          getRuntimeServiceGetFileQueryKey($runtime.instanceId, modelPath)
        );
      }
    } catch (err) {
      console.error(err);
    }

    if (overlayShown) {
      overlay.set(null);
    }
  }
  $: selections = $queryHighlight?.map((selection) => ({
    from: selection?.referenceIndex,
    to: selection?.referenceIndex + selection?.reference?.length,
  })) as SelectionRange[];

  let errors = [];
  $: {
    errors = [];
    if (embeddedSourceErrors?.length) errors.push(...embeddedSourceErrors);
    if (modelError) errors.push(modelError);
    if (runtimeError) errors.push(runtimeError.message);
  }
</script>

<svelte:window bind:innerHeight />

<div class="editor-pane bg-gray-100">
  <div
    style:height="calc({innerHeight}px - {$outputPosition *
      $outputVisibilityTween}px - var(--header-height))"
  >
    {#if hasModelSql}
      <div class="h-full p-5 grid overflow-auto">
        {#key modelName}
          <Editor
            {modelName}
            content={modelSql}
            {selections}
            focusOnMount={focusEditorOnMount}
            on:write={(evt) => updateModelContent(evt.detail.content)}
          />
        {/key}
      </div>
    {/if}
  </div>
  <Portal target=".body">
    {#if $outputLayout.visible}
      <HorizontalSplitter />
    {/if}
  </Portal>

  {#if hasModelSql}
    <div style:height="{$outputPosition}px" class="p-5 flex flex-col gap-6">
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
            <ConnectedPreviewTable objectName={modelName} {limit} />
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
