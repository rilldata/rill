<script lang="ts">
  import { selectedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { DimensionThresholdFilter } from "web-common/src/features/dashboards/stores/explore-state";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";
  import { COMPARISON_COLUMN_WIDTH, valueColumn } from "./leaderboard-widths";

  export let metricsViewName: string;
  export let whereFilter: V1Expression;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let timeControlsReady: boolean;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      numberFormat: { measureFormatters, activeMeasureFormatter },
      dimensionFilters: { isFilterExcludeMode },
      dimensions: { visibleDimensions },
      comparison: { isBeingCompared: isBeingComparedReadable },
      sorting: { sortedAscending, sortType },
      measures: { measureLabel, isMeasureValidPercentOfTotal },
      leaderboard: {
        leaderboardShowContextForAllMeasures,
        leaderboardMeasures,
        leaderboardSortByMeasureName,
      },
    },
    actions: {
      dimensions: { setPrimaryDimension },
      sorting: { toggleSort },
      dimensionsFilter: { toggleDimensionValueSelection },
      comparison: { toggleComparisonDimension },
    },
    exploreName,
    dashboardStore,
  } = StateManagers;

  let parentElement: HTMLDivElement;

  $: ({ instanceId } = $runtime);

  // Reset column widths when the measure changes
  $: if ($leaderboardSortByMeasureName) {
    valueColumn.reset();
  }

  $: dimensionColumnWidth = 164;

  $: showPercentOfTotal = $isMeasureValidPercentOfTotal(
    $leaderboardSortByMeasureName,
  );
  $: showDeltaPercent = !!comparisonTimeRange;

  $: tableWidth =
    dimensionColumnWidth +
    $valueColumn +
    (comparisonTimeRange
      ? COMPARISON_COLUMN_WIDTH * (showDeltaPercent ? 2 : 1)
      : showPercentOfTotal
        ? COMPARISON_COLUMN_WIDTH
        : 0);
</script>

<div class="flex flex-col overflow-hidden size-full" aria-label="Leaderboards">
  <div class="pl-2.5 pb-3">
    <LeaderboardControls exploreName={$exploreName} />
  </div>
  <div bind:this={parentElement} class="overflow-y-auto leaderboard-display">
    {#if parentElement}
      <div class="leaderboard-grid overflow-hidden pb-4">
        {#each $visibleDimensions as dimension (dimension.name)}
          {#if dimension.name}
            <Leaderboard
              isValidPercentOfTotal={$isMeasureValidPercentOfTotal}
              {metricsViewName}
              leaderboardSortByMeasureName={$leaderboardSortByMeasureName}
              leaderboardMeasures={$leaderboardMeasures}
              leaderboardShowContextForAllMeasures={$leaderboardShowContextForAllMeasures}
              {whereFilter}
              {dimensionThresholdFilters}
              {instanceId}
              {tableWidth}
              {timeRange}
              {dimensionColumnWidth}
              sortedAscending={$sortedAscending}
              sortType={$sortType}
              filterExcludeMode={$isFilterExcludeMode(dimension.name)}
              {comparisonTimeRange}
              {dimension}
              {parentElement}
              {timeControlsReady}
              selectedValues={selectedDimensionValues(
                $runtime.instanceId,
                [metricsViewName],
                $dashboardStore.whereFilter,
                dimension.name,
                timeRange.start,
                timeRange.end,
              )}
              isBeingCompared={$isBeingComparedReadable(dimension.name)}
              formatters={$leaderboardMeasures.length > 1
                ? $measureFormatters
                : { [$leaderboardSortByMeasureName]: $activeMeasureFormatter }}
              {setPrimaryDimension}
              {toggleSort}
              {toggleDimensionValueSelection}
              {toggleComparisonDimension}
              measureLabel={$measureLabel}
            />
          {/if}
        {/each}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .leaderboard-grid {
    @apply flex flex-row flex-wrap gap-4 overflow-x-auto;
  }
</style>
