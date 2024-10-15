<script lang="ts">
  import type { LeaderboardItemData } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";
  import {
    calculateLeaderboardColumnWidth,
    columnWidths,
    processedDimensions,
    resetColumnWidths,
    updateMaxColumnWidths,
  } from "./leaderboard-widths";

  const {
    selectors: {
      dimensions: { visibleDimensions },
      activeMeasure: { activeMeasureName, isValidPercentOfTotal },
      timeRangeSelectors: { isTimeComparisonActive },
      numberFormat: { activeMeasureFormatter },
    },
    exploreName,
  } = getStateManagers();

  let parentElement: HTMLDivElement;

  function calculateAllLeaderboardWidths(
    dimensionName: string,
    leaderboardData: {
      aboveTheFold: LeaderboardItemData[];
      selectedBelowTheFold: LeaderboardItemData[];
    },
  ) {
    if (
      !processedDimensions.has(dimensionName) &&
      processedDimensions.size < 6
    ) {
      const firstColumnWidth =
        !$isTimeComparisonActive && !$isValidPercentOfTotal ? 240 : 164;
      const widths = calculateLeaderboardColumnWidth(
        firstColumnWidth,
        leaderboardData.aboveTheFold,
        leaderboardData.selectedBelowTheFold,
        $activeMeasureFormatter,
      );
      updateMaxColumnWidths(dimensionName, widths);
    }
  }

  // Reset column widths when relevant data changes
  $: {
    $activeMeasureName;
    $isTimeComparisonActive;
    $isValidPercentOfTotal;
    resetColumnWidths();
  }
</script>

<div class="flex flex-col overflow-hidden size-full">
  <div class="pl-2.5 pb-3">
    <LeaderboardControls exploreName={$exploreName} />
  </div>
  <div bind:this={parentElement} class="overflow-y-auto leaderboard-display">
    {#if parentElement}
      <div class="leaderboard-grid overflow-hidden pb-4">
        {#each $visibleDimensions as item (item.name)}
          {#if item.name}
            <Leaderboard
              dimensionName={item.name}
              {parentElement}
              columnWidths={$columnWidths}
              {calculateAllLeaderboardWidths}
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
