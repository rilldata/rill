<script lang="ts">
  import { selectedDimensionValuesV2 } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { DimensionThresholdFilter } from "../stores/metrics-explorer-entity";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";
  import {
    DEFAULT_COL_WIDTH,
    deltaColumn,
    valueColumn,
  } from "./leaderboard-widths";

  export let metricsViewName: string;
  export let whereFilter: V1Expression;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let timeControlsReady: boolean;
  export let activeMeasureName: string;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      activeMeasure: { isValidPercentOfTotal, isSummableMeasure },
      numberFormat: { activeMeasureFormatter },
      dimensionFilters: { atLeastOneSelection, isFilterExcludeMode },
      dimensions: { visibleDimensions },
      comparison: { isBeingCompared: isBeingComparedReadable },
      sorting: { sortedAscending, sortType },
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
  let suppressTooltip = false;

  $: ({ instanceId } = $runtime);

  // Reset column widths when the measure changes
  $: if (activeMeasureName) {
    valueColumn.reset();
    deltaColumn.reset();
  }

  $: firstColumnWidth =
    !comparisonTimeRange && !$isValidPercentOfTotal ? 240 : 164;

  $: tableWidth =
    firstColumnWidth +
    $valueColumn +
    (comparisonTimeRange
      ? $deltaColumn + DEFAULT_COL_WIDTH
      : $isValidPercentOfTotal
        ? DEFAULT_COL_WIDTH
        : 0);
</script>

<div class="flex flex-col overflow-hidden size-full" aria-label="Leaderboards">
  <div class="pl-2.5 pb-3">
    <LeaderboardControls exploreName={$exploreName} />
  </div>
  <div
    bind:this={parentElement}
    class="overflow-y-auto leaderboard-display"
    on:scroll={() => {
      suppressTooltip = true;
    }}
    on:scrollend={() => {
      suppressTooltip = false;
    }}
  >
    {#if parentElement}
      <div class="leaderboard-grid overflow-hidden pb-4">
        {#each $visibleDimensions as dimension (dimension.name)}
          {#if dimension.name}
            <Leaderboard
              isValidPercentOfTotal={$isValidPercentOfTotal}
              {metricsViewName}
              {activeMeasureName}
              {whereFilter}
              {dimensionThresholdFilters}
              {instanceId}
              {tableWidth}
              {timeRange}
              {firstColumnWidth}
              sortedAscending={$sortedAscending}
              sortType={$sortType}
              filterExcludeMode={$isFilterExcludeMode(dimension.name)}
              atLeastOneActive={$atLeastOneSelection(dimension.name)}
              {comparisonTimeRange}
              {dimension}
              isSummableMeasure={$isSummableMeasure}
              {parentElement}
              {suppressTooltip}
              {timeControlsReady}
              selectedValues={selectedDimensionValuesV2(
                $runtime.instanceId,
                [metricsViewName],
                $dashboardStore.whereFilter,
                dimension.name,
                timeRange.start,
                timeRange.end,
              )}
              isBeingCompared={$isBeingComparedReadable(dimension.name)}
              formatter={$activeMeasureFormatter}
              {setPrimaryDimension}
              {toggleSort}
              {toggleDimensionValueSelection}
              {toggleComparisonDimension}
            />
          {/if}
        {/each}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .leaderboard-grid {
    @apply flex flex-row flex-wrap gap-4;
  }
</style>
