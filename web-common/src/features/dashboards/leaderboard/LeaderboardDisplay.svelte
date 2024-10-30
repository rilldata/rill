<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const StateManagers = getStateManagers();
  const {
    selectors: {
      activeMeasure: {
        activeMeasureName,
        isValidPercentOfTotal,
        isSummableMeasure,
      },
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
    dashboardStore,
    metricsViewName,
    validSpecStore,
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  let parentElement: HTMLDivElement;

  $: timeControls = $timeControlsStore;
  $: ({ instanceId } = $runtime);

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

  $: ({ whereFilter, dimensionThresholdFilters } = $dashboardStore);
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
              metricsViewName={$metricsViewName}
              activeMeasureName={$activeMeasureName}
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
              {toggleDimensionValueSelection}
              isSummableMeasure={$isSummableMeasure}
              {parentElement}
              metricsView={$validSpecStore.data?.metricsView ?? {}}
              timeControlsReady={!!timeControls.ready}
              selectedValues={$selectedDimensionValues(dimension.name)}
              isBeingCompared={$isBeingComparedReadable(dimension.name)}
              {setPrimaryDimension}
              {toggleSort}
              formatter={$activeMeasureFormatter}
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
