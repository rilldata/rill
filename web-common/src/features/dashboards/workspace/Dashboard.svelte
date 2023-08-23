<script lang="ts">
  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import { useDashboardStore } from "../dashboard-stores";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import DashboardContainer from "./DashboardContainer.svelte";
  import DashboardHeader from "./DashboardHeader.svelte";

  export let metricViewName: string;
  export let leftMargin = undefined;

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
  <DashboardHeader {metricViewName} slot="header" />

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
