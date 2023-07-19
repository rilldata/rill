<script lang="ts">
  import { cubicOut } from "svelte/easing";
  import { scale } from "svelte/transition";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  const sourceStore = useSourceStore();

  // Include `$file.dataUpdatedAt` and `clientYAML` in the reactive statement to recompute
  // the `isSourceUnsaved` value whenever they change
  const file = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );
  $: isSourceUnsaved =
    $file.dataUpdatedAt &&
    $sourceStore.clientYAML &&
    useIsSourceUnsaved($runtime.instanceId, sourceName);
</script>

{#if isSourceUnsaved}
  <div
    transition:scale={{ duration: 200, easing: cubicOut }}
    class="w-1.5 h-1.5 bg-gray-300 rounded"
  />
{/if}
