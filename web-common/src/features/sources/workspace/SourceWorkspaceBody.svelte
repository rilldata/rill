<script lang="ts">
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { getAllErrorsForFile } from "@rilldata/web-common/features/entity-management/resources-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import HorizontalSplitter from "../../../layout/workspace/HorizontalSplitter.svelte";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import SourceEditor from "../editor/SourceEditor.svelte";
  import ErrorPane from "../errors/ErrorPane.svelte";
  import { useIsSourceUnsaved, useSource } from "../selectors";
  import { useSourceStore } from "../sources-store";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { slide } from "svelte/transition";
  import { quintOut } from "svelte/easing";

  export let sourceName: string;

  const queryClient = useQueryClient();
  const sourceStore = useSourceStore(sourceName);

  $: filePath = getFilePathFromNameAndType(sourceName, EntityType.Table);

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
    filePath,
  );

  $: sourceQuery = useSource($runtime.instanceId, sourceName);

  $: workspaceLayout = $workspaces;

  $: tableHeight = workspaceLayout.table.height;

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML,
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

<div class="editor-pane h-full overflow-hidden w-full flex flex-col">
  <div
    class="p-5 size-full flex-shrink-1 overflow-hidden"
    style:min-height="150px"
  >
    <SourceEditor {sourceName} {yaml} />
  </div>

  <div
    class="p-5 w-full relative flex flex-none flex-col gap-2"
    style:height="{$tableHeight}px"
    style:max-height="75%"
    transition:slide={{ duration: 300, easing: quintOut }}
  >
    <Resizer max={600} direction="NS" side="top" bind:dimension={$tableHeight}>
      <HorizontalSplitter />
    </Resizer>
    <div class="table-wrapper" class:brightness-90={isSourceUnsaved}>
      {#if !$allErrors?.length}
        {#key sourceName}
          <ConnectedPreviewTable
            objectName={$sourceQuery?.data?.source?.state?.table}
            loading={resourceIsLoading($sourceQuery?.data)}
          />
        {/key}
      {:else if $allErrors[0].message}
        <ErrorPane {sourceName} errorMessage={$allErrors[0].message} />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .table-wrapper {
    transition: filter 200ms;
    @apply rounded w-full overflow-hidden border-gray-200 border-2 h-full;
  }
</style>
