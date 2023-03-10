<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    protoToBase64,
    toProto,
  } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { featureFlags } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useRuntimeServiceGetCatalogEntry } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import { MEASURE_CONFIG } from "../config";
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

  $: metricsViewQuery = useRuntimeServiceGetCatalogEntry(
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
  $: gridConfig = hasTimeSeries
    ? `${
        width >= MEASURE_CONFIG.breakpoint
          ? MEASURE_CONFIG.container.width.full
          : MEASURE_CONFIG.container.width.breakpoint
      }px minmax(355px, auto)`
    : "max-content minmax(355px, auto)";

  $: if (!$featureFlags.readOnly && metricsExplorer) {
    const binary = toProto(metricsExplorer).toBinary();
    const message = protoToBase64(binary);
    goto(`/dashboard/${metricViewName}?state=${message}`);
  }
</script>

<DashboardContainer bind:exploreContainerWidth {gridConfig} bind:width>
  <DashboardHeader {metricViewName} slot="header" />

  <svelte:fragment slot="metrics" let:width>
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
