<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import TabBar from "@rilldata/web-common/features/dashboards/tab-bar/TabBar.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { onMount, tick } from "svelte";
  import { useExploreState } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { DashboardState_ActivePage } from "../../../proto/gen/rill/ui/v1/dashboard_pb";
  import { runtime } from "../../../runtime-client/runtime-store";
  import CellInspector from "@rilldata/web-common/components/CellInspector.svelte";
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

  const DEFAULT_TIMESERIES_WIDTH = 580;
  const MIN_TIMESERIES_WIDTH = 440;
  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: { visibleMeasures },
      dimensions: { getDimensionByName },
      pivot: { showPivot },
    },
    dashboardStore,
  } = StateManagers;

  const { cloudDataViewer, readOnly } = featureFlags;

  const timeControlsStore = useTimeControlStore(StateManagers);

  let exploreContainerWidth: number;

  $: ({ instanceId } = $runtime);

  $: ({ whereFilter, dimensionThresholdFilters } = $dashboardStore);

  $: extraLeftPadding = !$navigationOpen;

  $: exploreState = useExploreState(exploreName);

  $: activePage = $exploreState?.activePage;
  $: showTimeDimensionDetail = Boolean(
    activePage === DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
  );
  $: showDimensionTable = Boolean(
    activePage === DashboardState_ActivePage.DIMENSION_TABLE,
  );

  $: selectedDimensionName = $exploreState?.selectedDimensionName;
  $: selectedDimension =
    selectedDimensionName && $getDimensionByName(selectedDimensionName);
  $: expandedMeasureName = $exploreState?.tdd?.expandedMeasureName;

  $: isRillDeveloper = $readOnly === false;

  // Check if the mock user (if selected) has access to the explore
  $: explore = useExploreValidSpec(instanceId, exploreName);

  $: hasTimeSeries = !!$explore.data?.metricsView?.timeDimension;

  $: mockUserHasNoAccess =
    $selectedMockUserStore && $explore.error?.response?.status === 404;

  $: hidePivot = isEmbedded && $explore.data?.explore?.embedsHidePivot;

  $: ({
    timeStart: start,
    timeEnd: end,
    showTimeComparison,
    comparisonTimeStart,
    comparisonTimeEnd,
    ready: timeControlsReady = false,
  } = $timeControlsStore);

  $: timeRange = {
    start,
    end,
  };

  $: comparisonTimeRange = showTimeComparison
    ? {
        start: comparisonTimeStart,
        end: comparisonTimeEnd,
      }
    : undefined;

  $: exploreSpec = $explore.data?.explore;
  $: timeRanges = exploreSpec?.timeRanges ?? [];

  $: visibleMeasureNames = $visibleMeasures.map(({ name }) => name ?? "");

  let metricsWidth = DEFAULT_TIMESERIES_WIDTH;
  let resizing = false;

  let initEmbedPublicAPI;

  // Hacky solution to ensure that the embed public API is initialized after the dashboard is fully loaded
  onMount(async () => {
    if (isEmbedded) {
      initEmbedPublicAPI = (
        await import(
          "@rilldata/web-admin/features/embeds/init-embed-public-api"
        )
      ).default;
    }
    await tick();
  });

  $: if (initEmbedPublicAPI) {
    try {
      initEmbedPublicAPI();
    } catch (error) {
      console.error("Error running initEmbedPublicAPI:", error);
    }
  }
</script>

<article
  class="flex flex-col size-full overflow-y-hidden dashboard-theme-boundary"
  bind:clientWidth={exploreContainerWidth}
>
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-background slide"
    class:left-shift={extraLeftPadding}
  >
    {#if mockUserHasNoAccess}
      <div class="mb-3" />
    {:else}
      {#key exploreName}
        <section class="flex relative justify-between gap-x-4 py-4 pb-6 px-4">
          <Filters {timeRanges} {metricsViewName} {hasTimeSeries} />
          <div class="absolute bottom-0 flex flex-col right-0">
            <TabBar {hidePivot} {exploreName} onPivot={$showPivot} />
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
  {:else if $showPivot}
    <PivotDisplay />
  {:else}
    <div
      class="flex gap-x-1 gap-y-2 size-full overflow-hidden pl-4 slide bg-surface"
      class:flex-col={showTimeDimensionDetail}
      class:flex-row={!showTimeDimensionDetail}
      class:left-shift={extraLeftPadding}
    >
      <div
        class="pt-2 flex-none"
        style:width={showTimeDimensionDetail ? "auto" : `${metricsWidth}px`}
      >
        {#key exploreName}
          {#if hasTimeSeries}
            <MetricsTimeSeriesCharts
              {exploreName}
              timeSeriesWidth={metricsWidth}
              workspaceWidth={exploreContainerWidth}
              hideStartPivotButton={hidePivot}
            />
          {:else}
            <MeasuresContainer {exploreContainerWidth} {metricsViewName} />
          {/if}
        {/key}
      </div>

      {#if showTimeDimensionDetail && expandedMeasureName}
        <hr class="border-t -ml-4" />
        <TimeDimensionDisplay
          {exploreName}
          {expandedMeasureName}
          hideStartPivotButton={hidePivot}
        />
      {:else}
        <div class="relative flex-none bg-gray-200 w-[1px]">
          <Resizer
            dimension={metricsWidth}
            min={MIN_TIMESERIES_WIDTH}
            max={exploreContainerWidth - 500}
            basis={DEFAULT_TIMESERIES_WIDTH}
            bind:resizing
            side="right"
            onUpdate={(width) => {
              metricsWidth = width;
            }}
          />
        </div>
        <div class="pt-2 pl-1 overflow-auto w-full">
          {#if showDimensionTable && selectedDimension}
            <DimensionDisplay
              dimension={selectedDimension}
              {metricsViewName}
              {whereFilter}
              {dimensionThresholdFilters}
              {timeRange}
              {comparisonTimeRange}
              {timeControlsReady}
              {visibleMeasureNames}
              hideStartPivotButton={hidePivot}
            />
          {:else}
            <LeaderboardDisplay
              {metricsViewName}
              {whereFilter}
              {dimensionThresholdFilters}
              {timeRange}
              {comparisonTimeRange}
              {timeControlsReady}
            />
          {/if}
        </div>
      {/if}
    </div>
  {/if}

  <CellInspector />
</article>

{#if (isRillDeveloper || $cloudDataViewer) && !showTimeDimensionDetail && !mockUserHasNoAccess}
  <RowsViewerAccordion {metricsViewName} {exploreName} />
{/if}

<style lang="postcss">
  .left-shift {
    @apply pl-8;
  }
</style>
