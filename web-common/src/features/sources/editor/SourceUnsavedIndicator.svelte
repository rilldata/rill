<script lang="ts">
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { cubicOut } from "svelte/easing";
  import { scale } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;
  // TODO: refactor this once we have moved everything and WorkspaceHeader is
  $: filePath = getFileAPIPathFromNameAndType(sourceName, EntityType.Table);

  $: sourceStore = useSourceStore(filePath);

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    filePath,
    $sourceStore.clientYAML,
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

{#if isSourceUnsaved}
  <div
    transition:scale|global={{ duration: 200, easing: cubicOut }}
    class="w-1.5 h-1.5 bg-gray-300 rounded"
  />
{/if}
