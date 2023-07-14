<script lang="ts">
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import { useIsSourceNotSaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  const sourceStore = useSourceStore();

  // Include `$file.dataUpdatedAt` and `clientYAML` in the reactive statement to recompute
  // the `isSourceNotSaved` value whenever they change
  const file = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );
  $: isSourceNotSaved =
    $file.dataUpdatedAt &&
    $sourceStore.clientYAML &&
    useIsSourceNotSaved($runtime.instanceId, sourceName);
</script>

{#if isSourceNotSaved}
  <div>NOT SAVED</div>
{/if}
