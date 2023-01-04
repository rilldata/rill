<script lang="ts">
  import type { SelectionRange } from "@codemirror/state";
  import Portal from "@rilldata/web-common/components/Portal.svelte";
  import {
    getEmbeddedReferences,
    Reference,
  } from "@rilldata/web-common/features/models/utils/get-table-references";
  import { humanReadableErrorMessage } from "@rilldata/web-common/features/sources/add-source/errors";
  import { useEmbeddedSources } from "@rilldata/web-common/features/sources/selectors";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import {
    useRuntimeServiceGetFile,
    useRuntimeServicePutFileAndReconcile,
    V1CatalogEntry,
    V1PutFileAndReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { SIDE_PAD } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import ConnectedPreviewTable from "@rilldata/web-local/lib/components/preview-table/ConnectedPreviewTable.svelte";
  import { drag } from "@rilldata/web-local/lib/drag";
  import {
    invalidateAfterReconcile,
    invalidationForProfileQueries,
  } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { getMapFromArray } from "@rilldata/web-local/lib/util/arrayUtils";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { slide } from "svelte/transition";
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

  $: modelSql = $modelSqlQuery?.data?.blob;
  $: hasModelSql = typeof modelSql === "string";

  let sanitizedQuery: string;
  $: sanitizedQuery = sanitizeQuery(modelSql ?? "");

  $: sourceCatalogsQuery = useEmbeddedSources($runtimeStore?.instanceId);
  let embeddedSourceCatalogs: Map<string, V1CatalogEntry>;
  $: embeddedSourceCatalogs = getMapFromArray(
    $sourceCatalogsQuery?.data ?? [],
    (entity) => entity.source.properties.path?.toLowerCase()
  ) as Map<string, V1CatalogEntry>;

  let embeddedSourceError: string;

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

  function filterKnownEmbeddedSources(
    embeddedRefs: Array<Reference>
  ): Array<string> {
    const unknownEmbeddedSources = new Array<string>();
    for (const embeddedRef of embeddedRefs) {
      const cleanedRef = embeddedRef.reference.slice(
        1,
        embeddedRef.reference.length - 1
      );
      const ref = cleanedRef.toLowerCase();
      if (embeddedSourceCatalogs.has(ref)) continue;
      unknownEmbeddedSources.push(cleanedRef);
    }
    return unknownEmbeddedSources;
  }

  async function updateModelContent(content: string) {
    const hasChanged = sanitizeQuery(content) !== sanitizedQuery;
    let overlayShown = false;

    try {
      if (hasChanged) {
        const unknownEmbeddedSources = filterKnownEmbeddedSources(
          getEmbeddedReferences(sanitizedQuery)
        );
        if (unknownEmbeddedSources.length > 0) {
          overlay.set({
            title: `Importing embedded sources for the 1st time : ${unknownEmbeddedSources.join(
              ","
            )}`,
          });
          overlayShown = true;
        }

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

      embeddedSourceError = "";
      for (const reconcileError of resp.errors) {
        if (
          reconcileError.filePath.startsWith("/source") &&
          embeddedSourceError === ""
        ) {
          // TODO: add url to this message
          embeddedSourceError = humanReadableErrorMessage(
            reconcileError.filePath.replace(/\/sources\/(.*?)_.*$/, "$1"),
            3,
            reconcileError.message
          );
        }
      }
      fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
      if (!resp.errors.length && hasChanged) {
        sanitizedQuery = sanitizeQuery(content);
      }
      await invalidateAfterReconcile(
        queryClient,
        $runtimeStore.instanceId,
        resp
      );
    } catch (err) {
      console.error(err);
    }

    if (overlayShown) {
      overlay.set(null);
    }
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
      {#if embeddedSourceError || modelError}
        <div
          transition:slide={{ duration: 200 }}
          class="error break-words overflow-auto p-6 border-2 border-gray-300 font-bold text-gray-700 w-full shrink-0 max-h-[60%] z-10 bg-gray-100"
        >
          {embeddedSourceError ?? modelError}
        </div>
      {/if}
    </div>
  {/if}
</div>
