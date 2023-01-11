<script lang="ts">
  import { goto } from "$app/navigation";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { metricsExplorerStore } from "../../../application-state-stores/explorer-stores";
  import { useMetaQuery } from "../../../svelte-query/dashboards";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import ExploreContainer from "./ExploreContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import DimensionDisplay from "./leaderboards/DimensionDisplay.svelte";
  import LeaderboardDisplay from "./leaderboards/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "./time-series-charts/MetricsTimeSeriesCharts.svelte";

  export let metricViewName: string;

  const switchToMetrics = async (metricViewName: string) => {
    if (!metricViewName) return;

    appStore.setActiveEntity(metricViewName, EntityType.MetricsExplorer);
  };

  $: switchToMetrics(metricViewName);

  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  $: selectedDimensionName = metricsExplorer?.selectedDimensionName;

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);
  $: if ($metaQuery.data) {
    metricsExplorerStore.sync(metricViewName, $metaQuery.data);
  }

  $: if ($metaQuery.isError) {
    goto(`/dashboard/${metricViewName}/edit`);
  }
</script>

<WorkspaceContainer
  top="0px"
  assetID={metricViewName}
  bgClass="bg-white"
  inspector={false}
>
  <ExploreContainer slot="body">
    <ExploreHeader {metricViewName} slot="header" />
    <MetricsTimeSeriesCharts {metricViewName} slot="metrics" />
    <svelte:fragment slot="leaderboards">
      {#if selectedDimensionName}
        <DimensionDisplay
          {metricViewName}
          dimensionName={selectedDimensionName}
        />
      {:else}
        <LeaderboardDisplay {metricViewName} />
      {/if}
    </svelte:fragment>
  </ExploreContainer>
</WorkspaceContainer>
