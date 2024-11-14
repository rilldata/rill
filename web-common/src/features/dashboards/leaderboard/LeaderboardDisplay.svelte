<script lang="ts">
  import type { LeaderboardItemData } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type {
    V1Expression,
    V1MetricsViewSpec,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { DimensionThresholdFilter } from "../stores/metrics-explorer-entity";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";
  import {
    calculateAndUpdateAllLeaderboardWidths,
    columnWidths,
    resetColumnWidths,
  } from "./leaderboard-widths";

  export let metricsViewName: string;
  export let whereFilter: V1Expression;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let timeControlsReady: boolean;
  export let activeMeasureName: string;
  export let metricsView: V1MetricsViewSpec;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      activeMeasure: { isValidPercentOfTotal, isSummableMeasure },
      numberFormat: { activeMeasureFormatter },
      dimensionFilters: {
        selectedDimensionValues,
        atLeastOneSelection,
        isFilterExcludeMode,
      },
      dimensions: { visibleDimensions },
      comparison: { isBeingCompared: isBeingComparedReadable },
      sorting: { sortedAscending, sortType },
    },
    actions: {
      dimensions: { setPrimaryDimension },
      sorting: { toggleSort },
      dimensionsFilter: { toggleDimensionValueSelection },
    },
    exploreName,
  } = StateManagers;

  let parentElement: HTMLDivElement;

  function estimateAndUpdateLeaderboardWidths(
    dimensionName: string,
    aboveTheFold: LeaderboardItemData[],
    selectedBelowTheFold: LeaderboardItemData[],
  ) {
    const firstColumnWidth =
      !comparisonTimeRange && !$isValidPercentOfTotal ? 240 : 164;
    calculateAndUpdateAllLeaderboardWidths(
      dimensionName,
      firstColumnWidth,
      aboveTheFold,
      selectedBelowTheFold,
      $activeMeasureFormatter,
    );
  }

  // Reset column widths when relevant data changes
  $: {
    activeMeasureName;
    !!comparisonTimeRange;
    !!timeRange;
    $isValidPercentOfTotal;
    resetColumnWidths();
  }
  $: ({ instanceId } = $runtime);
</script>

<div class="flex flex-col overflow-hidden size-full">
  <div class="pl-2.5 pb-3">
    <LeaderboardControls exploreName={$exploreName} />
  </div>
  <div bind:this={parentElement} class="overflow-y-auto leaderboard-display">
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
              {timeRange}
              sortedAscending={$sortedAscending}
              sortType={$sortType}
              filterExcludeMode={$isFilterExcludeMode(dimension.name)}
              atLeastOneActive={$atLeastOneSelection(dimension.name)}
              {comparisonTimeRange}
              {dimension}
              isSummableMeasure={$isSummableMeasure}
              {parentElement}
              {metricsView}
              {timeControlsReady}
              selectedValues={$selectedDimensionValues(dimension.name)}
              isBeingCompared={$isBeingComparedReadable(dimension.name)}
              columnWidths={$columnWidths}
              {estimateAndUpdateLeaderboardWidths}
              formatter={$activeMeasureFormatter}
              {setPrimaryDimension}
              {toggleSort}
              {toggleDimensionValueSelection}
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
