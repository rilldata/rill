<script lang="ts">
  import { goto } from "$app/navigation";
  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { featureFlags } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { createRuntimeServiceGetCatalogEntry } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import { metricsExplorerStore } from "../dashboard-stores";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import DashboardContainer from "./DashboardContainer.svelte";
  import DashboardHeader from "./DashboardHeader.svelte";

  export let metricViewName: string;
  export let hasTitle: boolean;

  export let leftMargin = undefined;

  const switchToMetrics = async (metricViewName: string) => {
    if (!metricViewName) return;

    appStore.setActiveEntity(metricViewName, EntityType.MetricsExplorer);
  };

  $: switchToMetrics(metricViewName);

  $: metricsViewQuery = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    metricViewName,
    {
      query: {
        select: (data) => data?.entry?.metricsView,
      },
    }
  );

  $: if ($metricsViewQuery.data) {
    if (!$featureFlags.readOnly && !$metricsViewQuery.data?.measures?.length) {
      goto(`/dashboard/${metricViewName}/edit`);
    }
    metricsExplorerStore.sync(metricViewName, $metricsViewQuery.data);
  }
  $: if (!$featureFlags.readOnly && $metricsViewQuery.isError) {
    goto(`/dashboard/${metricViewName}/edit`);
  }

  let exploreContainerWidth;

  let width;

  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  $: selectedDimensionName = metricsExplorer?.selectedDimensionName;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;
</script>

<DashboardContainer bind:exploreContainerWidth bind:width {leftMargin}>
  <DashboardHeader {hasTitle} {metricViewName} slot="header" />

  <svelte:fragment let:width slot="metrics">
    {#key metricViewName}
      {#if hasTimeSeries}
        <MetricsTimeSeriesCharts {metricViewName} workspaceWidth={width} />
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
