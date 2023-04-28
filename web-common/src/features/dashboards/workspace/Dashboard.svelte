<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { StateSyncManager } from "@rilldata/web-common/features/dashboards/proto-state/StateSyncManager";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { featureFlags } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
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

  $: metricsViewQuery = useMetaQuery($runtime.instanceId, metricViewName);

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

  $: metricsExplorer = useDashboardStore(metricViewName);

  $: stateSyncManager = new StateSyncManager(metricViewName);
  $: if ($metricsExplorer) {
    stateSyncManager.handleStateChange($metricsExplorer);
  }
  $: if ($page) {
    stateSyncManager.handleUrlChange();
  }

  $: selectedDimensionName = $metricsExplorer?.selectedDimensionName;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;
</script>

<DashboardContainer bind:exploreContainerWidth {leftMargin}>
  <DashboardHeader {hasTitle} {metricViewName} slot="header" />

  <svelte:fragment slot="metrics">
    {#key metricViewName}
      {#if hasTimeSeries}
        <MetricsTimeSeriesCharts
          {metricViewName}
          workspaceWidth={exploreContainerWidth}
        />
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
