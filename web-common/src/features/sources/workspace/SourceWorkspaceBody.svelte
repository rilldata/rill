<script lang="ts">
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import { useRuntimeServiceGetCatalogEntry } from "../../../runtime-client";
  import SourceWorkspaceErrorStates from "./SourceWorkspaceErrorStates.svelte";

  export let sourceName: string;

  $: getSource = useRuntimeServiceGetCatalogEntry(sourceName);
  $: isValidSource = $getSource?.data?.entry;
</script>

<div
  class="grid pb-3"
  style:grid-template-rows="max-content auto"
  style:height="100vh"
>
  {#if isValidSource}
    <div
      style:overflow="auto"
      style:height="calc(100vh - var(--header-height) - 2rem)"
      class="m-4 border border-gray-300 rounded"
    >
      {#key sourceName}
        <ConnectedPreviewTable objectName={sourceName} />
      {/key}
    </div>
  {:else}
    <SourceWorkspaceErrorStates {sourceName} />
  {/if}
</div>
