<script lang="ts">
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceGetFile,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import SourceEditor from "../editor/SourceEditor.svelte";
  import SourceWorkspaceErrorStates from "./SourceWorkspaceErrorStates.svelte";

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
</script>

<div
  class="grid pb-3"
  style:grid-template-rows="max-content auto"
  style:height="100vh"
>
  <SourceEditor {yaml} {sourceName} />
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
