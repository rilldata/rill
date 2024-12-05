<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import TabBar from "@rilldata/web-common/features/dashboards/tab-bar/TabBar.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { useExploreStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import Filters from "../filters/Filters.svelte";
  import { selectedMockUserStore } from "../granular-access-policies/stores";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import TimeDimensionDisplay from "../time-dimension-details/TimeDimensionDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";

  export let exploreName: string;
  export let metricsViewName: string;
  export let isEmbedded: boolean = false;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: { visibleMeasures },
      activeMeasure: { activeMeasureName },
      dimensions: { getDimensionByName },
    },
    dashboardStore,
    validSpecStore,
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  const { cloudDataViewer, readOnly } = featureFlags;

  let exploreContainerWidth: number;

  $: ({ whereFilter, dimensionThresholdFilters } = $dashboardStore);

  $: extraLeftPadding = !$navigationOpen;

  $: exploreStore = useExploreStore(exploreName);

  $: selectedDimensionName = $exploreStore?.selectedDimensionName;
  $: selectedDimension =
    selectedDimensionName && $getDimensionByName(selectedDimensionName);
  $: expandedMeasureName = $exploreStore?.tdd?.expandedMeasureName;
  $: showPivot = $exploreStore?.pivot?.active;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricsViewName,
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: isRillDeveloper = $readOnly === false;

  // Check if the mock user (if selected) has access to the explore
  $: explore = useExploreValidSpec($runtime.instanceId, exploreName);

  $: mockUserHasNoAccess =
    $selectedMockUserStore && $explore.error?.response?.status === 404;

  $: hidePivot = isEmbedded && $explore.data?.explore?.embedsHidePivot;

  $: timeControls = $timeControlsStore;

  $: timeRange = {
    start: timeControls.timeStart,
    end: timeControls.timeEnd,
  };

  $: comparisonTimeRange = timeControls.showTimeComparison
    ? {
        start: timeControls.comparisonTimeStart,
        end: timeControls.comparisonTimeEnd,
      }
    : undefined;

  $: metricsView = $validSpecStore.data?.metricsView ?? {};
</script>

<article
  class="flex flex-col size-full overflow-y-hidden dashboard-theme-boundary"
  bind:clientWidth={exploreContainerWidth}
>
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
    class:left-shift={extraLeftPadding}
  >
    {#if mockUserHasNoAccess}
      <div class="mb-3" />
    {:else}
      {#key exploreName}
        <section class="flex relative justify-between gap-x-4 py-4 pb-6 px-4">
          <Filters />
          <div class="absolute bottom-0 flex flex-col right-0">
            <TabBar {hidePivot} />
          </div>
        </section>
      {/key}
    {/if}
  </div>

  {#if mockUserHasNoAccess}
    <!-- Additional safeguard for mock users without dashboard access. -->
    <ErrorPage
      statusCode={$explore.error?.response?.status}
      header="This user can't access this dashboard"
      body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
    />
  {:else if showPivot}
    <PivotDisplay />
  {:else}
    <div
      class="flex gap-x-1 gap-y-2 size-full overflow-hidden pl-4 slide"
      class:flex-col={expandedMeasureName}
      class:flex-row={!expandedMeasureName}
      class:left-shift={extraLeftPadding}
    >
      <div class="pt-2">
        {#key exploreName}
          {#if hasTimeSeries}
            <MetricsTimeSeriesCharts
              {exploreName}
              workspaceWidth={exploreContainerWidth}
              hideStartPivotButton={hidePivot}
            />
          {:else}
            <MeasuresContainer {exploreContainerWidth} {metricsViewName} />
          {/if}
        {/key}
      </div>

      {#if expandedMeasureName}
        <hr class="border-t border-gray-200 -ml-4" />
        <TimeDimensionDisplay
          {exploreName}
          {expandedMeasureName}
          hideStartPivotButton={hidePivot}
        />
      {:else}
        <div class="pt-2 pl-1 border-l overflow-auto w-full">
          {#if selectedDimension}
            <DimensionDisplay
              dimension={selectedDimension}
              {metricsViewName}
              {whereFilter}
              {dimensionThresholdFilters}
              {timeRange}
              {comparisonTimeRange}
              activeMeasureName={$activeMeasureName}
              timeControlsReady={!!timeControls.ready}
              {metricsView}
              visibleMeasureNames={$visibleMeasures.map(
                ({ name }) => name ?? "",
              )}
              hideStartPivotButton={hidePivot}
            />
          {:else}
            <LeaderboardDisplay
              {metricsViewName}
              activeMeasureName={$activeMeasureName}
              {whereFilter}
              {dimensionThresholdFilters}
              {timeRange}
              {comparisonTimeRange}
              {metricsView}
              timeControlsReady={!!timeControls.ready}
            />
          {/if}
        </div>
      {/if}
    </div>
  {/if}
</article>

{#if (isRillDeveloper || $cloudDataViewer) && !expandedMeasureName && !mockUserHasNoAccess}
  <RowsViewerAccordion {metricsViewName} {exploreName} />
{/if}

<style lang="postcss">
  .left-shift {
    @apply pl-8;
  }
</style>
