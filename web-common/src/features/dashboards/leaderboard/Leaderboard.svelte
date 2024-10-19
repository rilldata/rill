<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import {
    type LeaderboardItemData,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardRow from "./LeaderboardRow.svelte";
  import LoadingRows from "./LoadingRows.svelte";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import { getComparisonRequestMeasures } from "../dashboard-utils";

  const slice = 7;
  const columnWidth = 66;
  const gutterWidth = 24;

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

  const StateManagers = getStateManagers();

  const {
    selectors: {
      dimensions: {
        getDimensionDisplayName,
        getDimensionDescription,
        getDimensionByName,
      },
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

  $: dimension = $getDimensionByName(dimensionName);

  const timeControlsStore = useTimeControlStore(StateManagers);

  $: timeControls = $timeControlsStore;

  $: sortedQuery = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    $metricsViewName,
    $leaderboardSortedQueryBody(dimensionName),
    $leaderboardSortedQueryOptions(dimensionName, visible),
  );

  $: belowTheFoldDataQuery = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    $metricsViewName,
    {
      dimensions: [{ name: dimensionName }],
      whereSql: selectedBelowTheFold
        .map((item) => {
          return `${dimensionName} = '${item.dimensionValue}'`;
        })
        .join(" OR "),
      timeRange: {
        start: timeControls.timeStart,
        end: timeControls.timeEnd,
      },
      comparisonTimeRange: {
        start: timeControls.comparisonTimeStart,
        end: timeControls.comparisonTimeEnd,
      },

      measures: [
        { name: $activeMeasureName },
        ...(timeControls.showTimeComparison
          ? getComparisonRequestMeasures($activeMeasureName)
          : []),
      ],
    },
    $leaderboardSortedQueryOptions(
      dimensionName,
      visible && selectedBelowTheFold.length > 0,
    ),
  );

  $: ({ data: sortedData, isFetching } = $sortedQuery);
  $: belowTheFoldData = $belowTheFoldDataQuery?.data?.data ?? [];

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
  let noAvailableValues = true;

  $: if (sortedData && !isFetching) {
    const leaderboardData = prepareLeaderboardItemData(
      (sortedData?.data ?? []).concat(belowTheFoldData),
      dimensionName,
      $activeMeasureName,
      slice,
      $selectedDimensionValues(dimensionName),
      leaderboardTotal,
    );

    aboveTheFold = leaderboardData.aboveTheFold;
    selectedBelowTheFold = leaderboardData.selectedBelowTheFold;
    noAvailableValues = leaderboardData.noAvailableValues;
    showExpandTable = leaderboardData.showExpandTable;
  }

  $: isBeingCompared = $isBeingComparedReadable(dimensionName);

  $: dimensionDescription = $getDimensionDescription(dimensionName);

  $: firstColumnWidth =
    !$isTimeComparisonActive && !$isValidPercentOfTotal ? 240 : 164;

  $: columnCount = $isTimeComparisonActive ? 3 : $isValidPercentOfTotal ? 2 : 1;

  $: tableWidth = columnCount * columnWidth + firstColumnWidth;
</script>

<div
  class="flex flex-col"
  aria-label="{dimensionName} leaderboard"
  role="table"
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  bind:this={container}
>
  <table style:width="{tableWidth + gutterWidth}px">
    <colgroup>
      <col style:width="{gutterWidth}px" />
      <col style:width="{firstColumnWidth}px" />
      <col style:width="{columnWidth}px" />
      {#if $isTimeComparisonActive}
        <col style:width="{columnWidth}px" />
        <col style:width="{columnWidth}px" />
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
      {#if isFetching}
        <LoadingRows columns={columnCount + 1} />
      {:else}
        {#each aboveTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardRow
            {tableWidth}
            {dimensionName}
            uri={dimension?.uri}
            {itemData}
            isValidPercentOfTotal={$isValidPercentOfTotal}
            isTimeComparisonActive={$isTimeComparisonActive}
            {columnWidth}
            {gutterWidth}
            {firstColumnWidth}
          />
        {/each}
      {/if}

      {#each selectedBelowTheFold as itemData, i (itemData.dimensionValue)}
        <LeaderboardRow
          {itemData}
          {tableWidth}
          {dimensionName}
          uri={dimension?.uri}
          isValidPercentOfTotal={$isValidPercentOfTotal}
          isTimeComparisonActive={$isTimeComparisonActive}
          borderTop={i === 0}
          borderBottom={i === selectedBelowTheFold.length - 1}
          {columnWidth}
          {gutterWidth}
          {firstColumnWidth}
        />
      {/each}
    </tbody>
  </table>

  {#if showExpandTable}
    <Tooltip location="right">
      <button
        class="transition-color ui-copy-muted table-message"
        on:click={() => setPrimaryDimension(dimensionName)}
      >
        (Expand Table)
      </button>
      <TooltipContent slot="tooltip-content">
        Expand dimension to see more values
      </TooltipContent>
    </Tooltip>
  {:else if noAvailableValues}
    <div class="table-message ui-copy-muted">(No available values)</div>
  {/if}
</div>

<style lang="postcss">
  table {
    @apply p-0 m-0 border-spacing-0 border-collapse w-fit;
    @apply font-normal cursor-pointer select-none;
    @apply table-fixed;
  }

  tbody {
    /* @apply bg-gray-50; */
  }

  .table-message {
    @apply h-[22px] p-1 flex-row w-full text-left pl-7;
  }
</style>
