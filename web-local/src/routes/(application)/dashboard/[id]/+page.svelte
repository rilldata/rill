<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { MetricsExplorerEntity } from "@rilldata/web-local/lib/application-state-stores/explorer-stores";
  import { metricsExplorerStore } from "@rilldata/web-local/lib/application-state-stores/explorer-stores";
  import ExploreContainer from "@rilldata/web-local/lib/components/workspace/explore/ExploreContainer.svelte";
  import ExploreHeader from "@rilldata/web-local/lib/components/workspace/explore/ExploreHeader.svelte";
  import DimensionDisplay from "@rilldata/web-local/lib/components/workspace/explore/leaderboards/DimensionDisplay.svelte";
  import LeaderboardDisplay from "@rilldata/web-local/lib/components/workspace/explore/leaderboards/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "@rilldata/web-local/lib/components/workspace/explore/time-series-charts/MetricsTimeSeriesCharts.svelte";

  export let data;

  $: metricsDefId = data.metricsDefId;

  $: dataModelerService.dispatch("setActiveAsset", [
    EntityType.MetricsExplorer,
    metricsDefId,
  ]);

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];
  $: selectedDimensionId = metricsExplorer?.selectedDimensionId;
</script>

<svelte:head>
  <!-- TODO: add the dashboard name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<ExploreContainer let:columns>
  <ExploreHeader slot="header" {metricsDefId} />
  <MetricsTimeSeriesCharts slot="metrics" {metricsDefId} />
  <svelte:fragment slot="leaderboards">
    {#if selectedDimensionId}
      <DimensionDisplay {metricsDefId} dimensionId={selectedDimensionId} />
    {:else}
      <LeaderboardDisplay {metricsDefId} />
    {/if}
  </svelte:fragment>
</ExploreContainer>
