<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
  import {
    LeaderboardItemData,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import { onMount } from "svelte";
  import LeaderboardRow from "./LeaderboardRow.svelte";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LoadingRows from "./LoadingRows.svelte";

  const slice = 7;

  export let parentElement: HTMLElement;
  export let dimensionName: string;

  const observer = new IntersectionObserver(
    ([entry]) => {
      visible = entry.isIntersecting;
    },
    {
      root: parentElement,
      rootMargin: "120px",
      threshold: 0,
    },
  );

  let container: HTMLElement;
  let visible = false;
  let hovered: boolean;

  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */

  onMount(() => {
    observer.observe(container);
  });

  const {
    selectors: {
      dimensions: { getDimensionDisplayName, getDimensionDescription },
      activeMeasure: { activeMeasureName, isValidPercentOfTotal },
      dimensionFilters: { selectedDimensionValues },
      dashboardQueries: {
        leaderboardSortedQueryBody,
        leaderboardSortedQueryOptions,
        leaderboardDimensionTotalQueryBody,
        leaderboardDimensionTotalQueryOptions,
      },
      sorting: { sortedAscending, sortType },
      timeRangeSelectors: { isTimeComparisonActive },
      comparison: { isBeingCompared: isBeingComparedReadable },
    },
    actions: {
      dimensions: { setPrimaryDimension },
      sorting: { toggleSort },
    },
    metricsViewName,
    runtime,
  } = getStateManagers();

  $: sortedQuery = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    $metricsViewName,
    $leaderboardSortedQueryBody(dimensionName),
    $leaderboardSortedQueryOptions(dimensionName, visible),
  );

  $: ({
    data: sortedData,

    isFetching,
  } = $sortedQuery);

  $: totalsQuery = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    $metricsViewName,
    $leaderboardDimensionTotalQueryBody(dimensionName),
    $leaderboardDimensionTotalQueryOptions(dimensionName),
  );

  $: leaderboardTotal = $totalsQuery?.data?.data?.[0]?.[$activeMeasureName];

  let aboveTheFold: LeaderboardItemData[] = [];
  let selectedBelowTheFold: LeaderboardItemData[] = [];
  let showExpandTable = false;

  $: if (sortedData && !isFetching) {
    const leaderboardData = prepareLeaderboardItemData(
      sortedData?.data ?? [],
      dimensionName,
      $activeMeasureName,
      slice,
      $selectedDimensionValues(dimensionName),
      leaderboardTotal,
    );

    aboveTheFold = leaderboardData.aboveTheFold;
    selectedBelowTheFold = leaderboardData.selectedBelowTheFold;

    showExpandTable = leaderboardData.showExpandTable;
  }

  $: isBeingCompared = $isBeingComparedReadable(dimensionName);

  $: dimensionDescription = $getDimensionDescription(dimensionName);

  $: firstColumnWidth =
    !$isTimeComparisonActive && !$isValidPercentOfTotal ? 240 : 190;

  $: columnCount = $isTimeComparisonActive ? 3 : $isValidPercentOfTotal ? 2 : 1;

  $: tableWidth = columnCount * 64 + firstColumnWidth;
</script>

<div
  class="flex-col flex"
  aria-label="{dimensionName} leaderboard"
  role="table"
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  bind:this={container}
>
  <table>
    <colgroup>
      <col style:width="24px" />
      <col style:width="{firstColumnWidth}px" />
      <col class="col-width" />
      {#if $isTimeComparisonActive}
        <col class="col-width" />
        <col class="col-width" />
      {:else}
        <col class="col-width" />
      {/if}
    </colgroup>

    <LeaderboardHeader
      {hovered}
      displayName={$getDimensionDisplayName(dimensionName)}
      {dimensionDescription}
      {dimensionName}
      {isBeingCompared}
      {isFetching}
      sortType={$sortType}
      {toggleSort}
      {setPrimaryDimension}
      isValidPercentOfTotal={$isValidPercentOfTotal}
      sortedAscending={$sortedAscending}
      isTimeComparisonActive={$isTimeComparisonActive}
    />

    <tbody>
      {#each aboveTheFold as itemData (itemData.dimensionValue)}
        <LeaderboardRow
          {tableWidth}
          {dimensionName}
          {itemData}
          isValidPercentOfTotal={$isValidPercentOfTotal}
          isTimeComparisonActive={$isTimeComparisonActive}
        />
      {:else}
        <LoadingRows />
      {/each}

      <!-- place the selected values that are not above the fold here -->
      {#each selectedBelowTheFold as itemData, i (itemData.dimensionValue)}
        <LeaderboardRow
          borderTop={i === 0}
          borderBottom={i === selectedBelowTheFold.length - 1}
          {tableWidth}
          {dimensionName}
          {itemData}
          isValidPercentOfTotal={$isValidPercentOfTotal}
          isTimeComparisonActive={$isTimeComparisonActive}
        />
      {/each}
    </tbody>
  </table>

  {#if showExpandTable}
    <Tooltip location="right">
      <button
        on:click={() => setPrimaryDimension(dimensionName)}
        class="block flex-row w-full text-left transition-color ui-copy-muted pl-7"
      >
        (Expand Table)
      </button>
      <TooltipContent slot="tooltip-content">
        Expand dimension to see more values
      </TooltipContent>
    </Tooltip>
  {/if}
</div>

<style lang="postcss">
  table {
    @apply p-0 m-0 border-spacing-0 border-collapse w-fit;
    @apply font-normal cursor-pointer select-none;
    @apply table-fixed;
  }

  .col-width {
    width: 64px;
  }
</style>
