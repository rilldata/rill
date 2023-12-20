<script lang="ts">
  import CancelCircle from "../../../components/icons/CancelCircle.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import { humanReadableErrorMessage } from "../modal/errors";
  import { useSourceFromYaml } from "../selectors";

  export let sourceName: string;
  export let errorMessage: string | undefined;

  // Parse Source YAML client-side
  $: sourceFromYaml = useSourceFromYaml(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );

  // Try to extract the connector type
  $: connectorType = $sourceFromYaml.data?.type;

  // Try to create an actionable error message
  $: prettyMessage = humanReadableErrorMessage(connectorType, 3, errorMessage);
</script>

<div class="w-full h-full bg-white flex-col justify-center inline-flex p-3">
  <div class="flex-col justify-start items-center gap-1 flex text-red-500">
    <CancelCircle size="24px" />
    <div class="text-center text-sm font-medium">
      {prettyMessage ?? errorMessage}
    </div>
  </div>
</div>
