<script lang="ts">
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { hasDefinedTimeSeries } from "./utils";

  $: instanceId = $runtimeStore.instanceId;
  $: metaQuery = useMetaQuery(instanceId, metricViewName);
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../application-state-stores/explorer-stores";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import DimensionDisplay from "./leaderboards/DimensionDisplay.svelte";
  import LeaderboardDisplay from "./leaderboards/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "./metrics-container/MetricsTimeSeriesCharts.svelte";
  import DefaultViewContainer from "./view-containers/DefaultViewContainer.svelte";
  import NoTimeSeriesContainer from "./view-containers/NoTimeSeriesContainer.svelte";
  import MeasuresContainer from "./metrics-container/MeasuresContainer.svelte";

  export let metricViewName: string;

  let hasTimeSeries = true;
  $: if (metaQuery && $metaQuery.isSuccess && !$metaQuery.isRefetching) {
    hasTimeSeries = hasDefinedTimeSeries($metaQuery.data);
  }

  const switchToMetrics = async (metricViewName: string) => {
    if (!metricViewName) return;

    appStore.setActiveEntity(metricViewName, EntityType.MetricsExplorer);
  };

  $: switchToMetrics(metricViewName);

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  $: selectedDimensionName = metricsExplorer?.selectedDimensionName;

  $: ExploreContainer = hasTimeSeries
    ? DefaultViewContainer
    : NoTimeSeriesContainer;

  $: MetricsContainer = hasTimeSeries
    ? MetricsTimeSeriesCharts
    : MeasuresContainer;
</script>

<WorkspaceContainer
  top="0px"
  assetID={metricViewName}
  bgClass="bg-white"
  inspector={false}
>
  <svelte:component this={ExploreContainer} slot="body">
    <ExploreHeader {metricViewName} slot="header" />
    <svelte:component this={MetricsContainer} {metricViewName} slot="metrics" />
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
  </svelte:component>
</WorkspaceContainer>
