<script lang="ts">
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import HorizontalSplitter from "../../../layout/workspace/HorizontalSplitter.svelte";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import {
    fileArtifactsStore,
    getFileArtifactReconciliationErrors,
  } from "../../entity-management/file-artifacts-store";
  import { EntityType } from "../../entity-management/types";
  import SourceEditor from "../editor/SourceEditor.svelte";
  import ErrorPane from "../errors/ErrorPane.svelte";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  const sourceStore = useSourceStore(sourceName);

  $: file = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table),
    {
      query: {
        // this will ensure that any changes done outside our app is pulled in.
        refetchOnWindowFocus: true,
      },
    }
  );

  $: yaml = $file.data?.blob || "";

  $: reconciliationErrors = getFileArtifactReconciliationErrors(
    $fileArtifactsStore,
    `${sourceName}.yaml`
  );

  const outputPosition = getContext(
    "rill:app:output-height-tween"
  ) as Writable<number>;
  const outputVisibilityTween = getContext(
    "rill:app:output-visibility-tween"
  ) as Writable<number>;

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

<div class="h-full pb-3">
  <div
    class="p-5"
    style:height="calc({innerHeight}px - {$outputPosition *
      $outputVisibilityTween}px - var(--header-height))"
  >
    <SourceEditor {sourceName} {yaml} />
  </div>
  <HorizontalSplitter />
  <div class="p-5" style:height="{$outputPosition}px">
    <div
      class="h-full border border-gray-300 rounded overflow-auto {isSourceUnsaved &&
        'brightness-90'} transition duration-200"
    >
      {#if !reconciliationErrors || reconciliationErrors.length === 0}
        {#key sourceName}
          <ConnectedPreviewTable objectName={sourceName} />
        {/key}
      {:else}
        <ErrorPane error={reconciliationErrors[0]} />
      {/if}
    </div>
  </div>
</div>
