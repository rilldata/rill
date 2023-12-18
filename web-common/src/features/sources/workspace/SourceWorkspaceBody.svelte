<script lang="ts">
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { getAllErrorsForFile } from "@rilldata/web-common/features/entity-management/resources-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import HorizontalSplitter from "../../../layout/workspace/HorizontalSplitter.svelte";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import SourceEditor from "../editor/SourceEditor.svelte";
  import ErrorPane from "../errors/ErrorPane.svelte";
  import { useIsSourceUnsaved, useSource } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;
  $: filePath = getFilePathFromNameAndType(sourceName, EntityType.Table);

  const queryClient = useQueryClient();
  const sourceStore = useSourceStore(sourceName);

  $: file = createRuntimeServiceGetFile($runtime.instanceId, filePath, {
    query: {
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  $: yaml = $file.data?.blob || "";

  $: allErrors = getAllErrorsForFile(
    queryClient,
    $runtime.instanceId,
    filePath
  );

  $: sourceQuery = useSource($runtime.instanceId, sourceName);

  // Layout state
  const outputPosition = getContext(
    "rill:app:output-height-tween"
  ) as Writable<number>;
  const outputVisibilityTween = getContext(
    "rill:app:output-visibility-tween"
  ) as Writable<number>;
  // track innerHeight to calculate the size of the editor element.
  let innerHeight: number;

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

<svelte:window bind:innerHeight />

<div class="h-full pb-3">
  <div
    class="p-5"
    style:height="calc({innerHeight}px - {$outputPosition *
      $outputVisibilityTween}px - var(--header-height))"
  >
    <SourceEditor {sourceName} {yaml} />
  </div>
  <HorizontalSplitter className="px-5" />
  <div class="p-5" style:height="{$outputPosition}px">
    <div
      class="h-full border border-gray-300 rounded overflow-auto {isSourceUnsaved &&
        'brightness-90'} transition duration-200"
    >
      {#if !$allErrors?.length}
        {#key sourceName}
          <ConnectedPreviewTable
            objectName={$sourceQuery?.data?.source?.state?.table}
            loading={resourceIsLoading($sourceQuery?.data)}
          />
        {/key}
      {:else}
        <ErrorPane {sourceName} errorMessage={$allErrors[0].message} />
      {/if}
    </div>
  </div>
</div>
