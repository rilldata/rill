<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { selectedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";
  import {
    COMPARISON_COLUMN_WIDTH,
    dimensionColumn,
    MAX_DIMENSION_COLUMN_WIDTH,
    MIN_DIMENSION_COLUMN_WIDTH,
    valueColumn,
  } from "./leaderboard-widths";

  export let metricsViewName: string;
  export let dimensionOnlyFilter: V1Expression | undefined;
  export let whereFilter: V1Expression | undefined;
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let timeControlsReady: boolean;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      numberFormat: {
        measureFormatters,
        activeMeasureFormatter,
        measureTooltipFormatters,
        activeMeasureTooltipFormatter,
      },
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
      comparison: { toggleComparisonDimension },
    },
    exploreName,
    expressionFilterManager,
  } = StateManagers;

  const client = useRuntimeClient();

  let parentElement: HTMLDivElement;

  // Reset column widths when the measure changes
  $: if ($leaderboardSortByMeasureName) {
    valueColumn.reset();
  }

  $: dimensionColumnWidth = clamp(
    MIN_DIMENSION_COLUMN_WIDTH,
    $dimensionColumn,
    MAX_DIMENSION_COLUMN_WIDTH,
  );

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

<div
  class="flex flex-col overflow-hidden size-full"
  aria-label={m.dashboard_leaderboards_aria()}
>
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
              {dimensionOnlyFilter}
              {whereFilter}
              {tableWidth}
              {timeRange}
              {dimensionColumnWidth}
              sortedAscending={$sortedAscending}
              sortType={$sortType}
              filterExcludeMode={expressionFilterManager.filterManagers.dimensions.find(
                (dfm) => dfm.name === dimension.name,
              )?.exclude ?? false}
              {comparisonTimeRange}
              {dimension}
              {parentElement}
              {timeControlsReady}
              selectedValues={selectedDimensionValues(
                client,
                [metricsViewName],
                whereFilter,
                dimension.name,
                timeRange.start,
                timeRange.end,
              )}
              isBeingCompared={$isBeingComparedReadable(dimension.name)}
              formatters={$leaderboardMeasures.length > 1
                ? $measureFormatters
                : { [$leaderboardSortByMeasureName]: $activeMeasureFormatter }}
              tooltipFormatters={$leaderboardMeasures.length > 1
                ? $measureTooltipFormatters
                : {
                    [$leaderboardSortByMeasureName]:
                      $activeMeasureTooltipFormatter,
                  }}
              {setPrimaryDimension}
              {toggleSort}
              toggleDimensionValueSelection={(_1, value, _2, exclusive) =>
                expressionFilterManager.dimensionFilterAction(
                  dimension.name!,
                  (dimensionManager) =>
                    dimensionManager.toggleValue(value, exclusive ?? false),
                )}
              {toggleComparisonDimension}
              measureLabel={$measureLabel}
              onDimensionColumnResize={dimensionColumn.set}
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
