<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { WorkspaceContainer } from "../../../layout/workspace";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import { metricsExplorerStore } from "../dashboard-stores";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import DashboardContainer from "./DashboardContainer.svelte";
  import DashboardHeader from "./DashboardHeader.svelte";

  export let metricViewName: string;

  const switchToMetrics = async (metricViewName: string) => {
    if (!metricViewName) return;

    appStore.setActiveEntity(metricViewName, EntityType.MetricsExplorer);
  };

  $: switchToMetrics(metricViewName);

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);

  $: if ($metaQuery.data) {
    if (!$metaQuery.data?.measures?.length) {
      goto(`/dashboard/${metricViewName}/edit`);
    }
    metricsExplorerStore.sync(metricViewName, $metaQuery.data);
  }
  $: if ($metaQuery.isError) {
    goto(`/dashboard/${metricViewName}/edit`);
  }

  let exploreContainerWidth;

  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  $: selectedDimensionName = metricsExplorer?.selectedDimensionName;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtimeStore.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: gridConfig = hasTimeSeries
    ? "calc(560px + 1rem) minmax(355px, auto)"
    : "max-content minmax(355px, auto)";
</script>

<WorkspaceContainer
  top="0px"
  assetID={metricViewName}
  bgClass="bg-white"
  inspector={false}
>
  <DashboardContainer bind:exploreContainerWidth {gridConfig} slot="body">
    <DashboardHeader {metricViewName} slot="header" />

    <svelte:fragment slot="metrics">
      {#key metricViewName}
        {#if hasTimeSeries}
          <MetricsTimeSeriesCharts {metricViewName} />
        {:else}
          <MeasuresContainer {exploreContainerWidth} {metricViewName} />
        {/if}
      {/key}
    </svelte:fragment>

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
  </DashboardContainer>
</WorkspaceContainer>
