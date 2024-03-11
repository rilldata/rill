<script lang="ts">
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { runtime } from "../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../entity-management/actions";
  import { EntityType } from "../entity-management/types";
  import { useChartFileNames } from "./selectors";

  export let chartName: string;

  $: chartFileNames = useChartFileNames($runtime.instanceId);

  async function handleDeleteChart() {
    await deleteFileArtifact(
      $runtime.instanceId,
      chartName,
      EntityType.Chart,
      $chartFileNames.data ?? [],
    );
  }
</script>

<MenuItem icon on:select={handleDeleteChart} propogateSelect={false}>
  Delete chart
</MenuItem>
