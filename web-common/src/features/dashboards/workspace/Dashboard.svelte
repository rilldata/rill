<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import { useDashboardStore } from "../dashboard-stores";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import DashboardContainer from "./DashboardContainer.svelte";
  import DashboardHeader from "./DashboardHeader.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";

  export let metricViewName: string;
  export let hasTitle: boolean;

  export let leftMargin = undefined;

  $: metricsViewQuery = useMetaQuery($runtime.instanceId, metricViewName);

  $: if ($metricsViewQuery.data) {
    if (!$featureFlags.readOnly && !$metricsViewQuery.data?.measures?.length) {
      goto(`/dashboard/${metricViewName}/edit`);
    }
  }

  $: if (!$featureFlags.readOnly && $metricsViewQuery.isError) {
    goto(`/dashboard/${metricViewName}/edit`);
  }

  let exploreContainerWidth;

  $: metricsExplorer = useDashboardStore(metricViewName);

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

  <svelte:fragment slot="rows">
    {#if !$featureFlags.readOnly}
      <RowsViewerAccordion {metricViewName} />
    {/if}
  </svelte:fragment>
</DashboardContainer>
