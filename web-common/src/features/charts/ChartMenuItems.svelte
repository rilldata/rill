<script lang="ts">
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../entity-management/actions";
  import { EntityType } from "../entity-management/types";
  import { useChartRoutes } from "./selectors";
  import { getNextRoute } from "../models/utils/navigate-to-next";
  import { goto } from "$app/navigation";

  export let chartName: string;
  export let open: boolean;

  $: chartRoutesQuery = useChartRoutes($runtime.instanceId);
  $: chartRoutes = $chartRoutesQuery.data ?? [];

  async function handleDeleteChart() {
    await deleteFileArtifact(
      $runtime.instanceId,
      getFileAPIPathFromNameAndType(chartName, EntityType.Chart),
      EntityType.Chart,
    );

    if (open) await goto(getNextRoute(chartRoutes));
  }
</script>

<NavigationMenuItem on:click={handleDeleteChart}>
  Delete chart
</NavigationMenuItem>
