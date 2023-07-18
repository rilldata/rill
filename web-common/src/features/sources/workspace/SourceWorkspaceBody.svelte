<script lang="ts">
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import HorizontalSplitter from "../../../layout/workspace/HorizontalSplitter.svelte";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceGetFile,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import {
    fileArtifactsStore,
    getFileArtifactReconciliationErrors,
  } from "../../entity-management/file-artifacts-store";
  import { EntityType } from "../../entity-management/types";
  import SourceEditor from "../editor/SourceEditor.svelte";
  import ErrorPane from "../errors/ErrorPane.svelte";

  export let sourceName: string;

  $: getSource = createRuntimeServiceGetCatalogEntry(
    $runtime?.instanceId,
    sourceName
  );
  $: isValidSource = $getSource?.data?.entry;

  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table),
    {
      query: {
        // this will ensure that any changes done outside our app is pulled in.
        refetchOnWindowFocus: true,
      },
    }
  );

  $: yaml = $fileQuery.data?.blob || "";

  $: runtimeErrors = getFileArtifactReconciliationErrors(
    $fileArtifactsStore,
    `${sourceName}.yaml`
  );

  const outputPosition = getContext(
    "rill:app:output-height-tween"
  ) as Writable<number>;
  const outputVisibilityTween = getContext(
    "rill:app:output-visibility-tween"
  ) as Writable<number>;
</script>

<div class="h-full pb-3">
  <div
    class="p-5"
    style:height="calc({innerHeight}px - {$outputPosition *
      $outputVisibilityTween}px - var(--header-height))"
  >
    <SourceEditor {yaml} />
  </div>
  <HorizontalSplitter />
  <div class="p-5" style:height="{$outputPosition}px">
    <div class="h-full border border-gray-300 rounded overflow-auto">
      {#if !runtimeErrors || runtimeErrors.length === 0}
        {#key sourceName}
          <ConnectedPreviewTable objectName={sourceName} />
        {/key}
      {:else}
        <ErrorPane error={runtimeErrors[0]} />
      {/if}
    </div>
  </div>
</div>
