<script lang="ts">
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
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

<NavigationMenuItem on:click={handleDeleteChart}>
  Delete chart
</NavigationMenuItem>
