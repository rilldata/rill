<script lang="ts">
  import { goto } from "$app/navigation";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useModelHasTimeSeries } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { metricsExplorerStore } from "../../../application-state-stores/explorer-stores";
  import { useMetaQuery } from "../../../svelte-query/dashboards";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import ExploreContainer from "./ExploreContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import DimensionDisplay from "./leaderboards/DimensionDisplay.svelte";
  import LeaderboardDisplay from "./leaderboards/LeaderboardDisplay.svelte";
  import MeasuresContainer from "./metrics-container/MeasuresContainer.svelte";
  import MetricsTimeSeriesCharts from "./metrics-container/MetricsTimeSeriesCharts.svelte";

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
    ? "560px minmax(355px, auto)"
    : "max-content minmax(355px, auto)";
</script>

<WorkspaceContainer
  top="0px"
  assetID={metricViewName}
  bgClass="bg-white"
  inspector={false}
>
  <ExploreContainer bind:exploreContainerWidth {gridConfig} slot="body">
    <ExploreHeader {metricViewName} slot="header" />

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
  </ExploreContainer>
</WorkspaceContainer>
