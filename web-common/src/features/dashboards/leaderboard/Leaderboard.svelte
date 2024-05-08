<script lang="ts">
  /**
   * Leaderboard.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    createQueryServiceMetricsViewComparison,
    createQueryServiceMetricsViewTotals,
  } from "@rilldata/web-common/runtime-client";

  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardListItem from "./LeaderboardListItem.svelte";
  import {
    LeaderboardItemData,
    getLabeledComparisonFromComparisonRow,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import { onMount } from "svelte";

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

  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */

  onMount(() => {
    observer.observe(container);
  });

  const {
    selectors: {
      activeMeasure: { activeMeasureName },
      dimensionFilters: { selectedDimensionValues },
      measureFilters: { getResolvedFilterForMeasureFilters },
      dashboardQueries: {
        leaderboardSortedQueryBody,
        leaderboardSortedQueryOptions,
        leaderboardDimensionTotalQueryBody,
        leaderboardDimensionTotalQueryOptions,
      },
    },
    actions: {
      dimensions: { setPrimaryDimension },
    },
    metricsViewName,
    runtime,
  } = getStateManagers();

  $: resolvedFilter = $getResolvedFilterForMeasureFilters;

  $: sortedQuery = createQueryServiceMetricsViewComparison(
    $runtime.instanceId,
    $metricsViewName,
    $leaderboardSortedQueryBody(dimensionName, $resolvedFilter),
    $leaderboardSortedQueryOptions(dimensionName, $resolvedFilter, visible),
  );

  $: ({
    isLoading,
    isError,
    data: sortedData,
    refetch,
    isFetching,
  } = $sortedQuery);

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    $runtime.instanceId,
    $metricsViewName,
    $leaderboardDimensionTotalQueryBody(dimensionName, $resolvedFilter),
    $leaderboardDimensionTotalQueryOptions(dimensionName, $resolvedFilter),
  );

  $: leaderboardTotal = $totalsQuery?.data?.data?.[$activeMeasureName];

  let aboveTheFold: LeaderboardItemData[] = [];
  let selectedBelowTheFold: LeaderboardItemData[] = [];
  let noAvailableValues = true;
  let showExpandTable = false;
  $: if (sortedData && !isFetching) {
    const leaderboardData = prepareLeaderboardItemData(
      sortedData?.rows?.map((r) =>
        getLabeledComparisonFromComparisonRow(r, $activeMeasureName),
      ) ?? [],
      slice,
      $selectedDimensionValues(dimensionName),
      leaderboardTotal,
    );

    aboveTheFold = leaderboardData.aboveTheFold;
    selectedBelowTheFold = leaderboardData.selectedBelowTheFold;
    noAvailableValues = leaderboardData.noAvailableValues;
    showExpandTable = leaderboardData.showExpandTable;
  }

  let hovered: boolean;
</script>

<div
  bind:this={container}
  role="grid"
  aria-label="{dimensionName} leaderboard"
  tabindex="0"
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
>
  <LeaderboardHeader {isFetching} {dimensionName} {hovered} />
  {#if isError}
    <div class="ml-[22px] flex p-2 gap-x-1 items-center">
      <div class="text-gray-500">Unable to load leaderboard.</div>
      <button
        class="text-primary-500 hover:text-primary-600 font-medium"
        disabled={isLoading}
        on:click={() => refetch()}>Try again</button
      >
    </div>
  {:else if isLoading}
    <div class="pl-6 pr-0.5 w-full flex flex-col items-center">
      {#each { length: 7 } as _, i (i)}
        <div class="size-full flex h-[22px] py-1.5 gap-x-1">
          <div
            class="h-full w-10/12 flex-none bg-gray-100 animate-pulse rounded-full"
          />
          <div class="size-full bg-gray-100 animate-pulse rounded-full" />
        </div>
      {/each}
    </div>
  {:else if aboveTheFold || selectedBelowTheFold}
    <div class="rounded-b border-gray-200 surface text-gray-800">
      <!-- place the leaderboard entries that are above the fold here -->
      {#each aboveTheFold as itemData (itemData.dimensionValue)}
        <LeaderboardListItem {dimensionName} {itemData} on:click on:keydown />
      {/each}
      <!-- place the selected values that are not above the fold here -->
      {#if selectedBelowTheFold?.length}
        <hr />
        {#each selectedBelowTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardListItem {dimensionName} {itemData} on:click on:keydown />
        {/each}
        <hr />
      {/if}
      {#if noAvailableValues}
        <div style:padding-left="30px" class="p-1 ui-copy-disabled">
          No available values
        </div>
      {/if}
      {#if showExpandTable}
        <Tooltip location="right">
          <button
            on:click={() => setPrimaryDimension(dimensionName)}
            class="block flex-row w-full text-left transition-color ui-copy-muted"
            style:padding-left="30px"
          >
            (Expand Table)
          </button>
          <TooltipContent slot="tooltip-content"
            >Expand dimension to see more values</TooltipContent
          >
        </Tooltip>
      {/if}
    </div>
  {/if}
</div>
