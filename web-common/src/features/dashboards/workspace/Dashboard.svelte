<script lang="ts">
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import TimeDimensionDisplay from "../time-dimension-details/TimeDimensionDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { page } from "$app/stores";

  export let metricViewName: string;

  const { cloudDataViewer, readOnly } = featureFlags;

  let exploreContainerWidth: number;

  $: extraLeftPadding = !$navigationOpen;

  $: metricsExplorer = useDashboardStore(metricViewName);

  $: selectedDimensionName = $metricsExplorer?.selectedDimensionName;
  $: expandedMeasureName = $metricsExplorer?.expandedMeasureName;
  $: view = $page.params.view;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName,
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: isRillDeveloper = $readOnly === false;
</script>

<div class="size-full overflow-hidden" bind:clientWidth={exploreContainerWidth}>
  {#if view === "pivot"}
    <PivotDisplay />
  {:else}
    <div
      class="flex gap-x-1 gap-y-4 pt-3 size-full overflow-hidden pl-4 slide"
      class:flex-col={expandedMeasureName}
      class:flex-row={!expandedMeasureName}
      class:left-shift={extraLeftPadding}
    >
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

      {#if expandedMeasureName}
        <TimeDimensionDisplay {metricViewName} />
      {:else if selectedDimensionName}
        <DimensionDisplay />
      {:else}
        <LeaderboardDisplay />
      {/if}
    </div>
  {/if}
</div>

{#if (isRillDeveloper || $cloudDataViewer) && !expandedMeasureName && view !== "pivot"}
  <RowsViewerAccordion {metricViewName} />
{/if}

<style lang="postcss">
  .left-shift {
    @apply pl-8;
  }
</style>
