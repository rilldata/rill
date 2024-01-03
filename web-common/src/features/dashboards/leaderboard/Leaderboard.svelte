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

  export let dimensionName: string;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */

  let slice = 7;

  const {
    selectors: {
      activeMeasure: { activeMeasureName },
      dimensionFilters: { selectedDimensionValues },
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

  $: sortedQuery = createQueryServiceMetricsViewComparison(
    $runtime.instanceId,
    $metricsViewName,
    $leaderboardSortedQueryBody(dimensionName),
    $leaderboardSortedQueryOptions(dimensionName),
  );

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    $runtime.instanceId,
    $metricsViewName,
    $leaderboardDimensionTotalQueryBody(dimensionName),
    $leaderboardDimensionTotalQueryOptions(dimensionName),
  );

  $: leaderboardTotal = $totalsQuery?.data?.data?.[$activeMeasureName];

  let aboveTheFold: LeaderboardItemData[] = [];
  let selectedBelowTheFold: LeaderboardItemData[] = [];
  let noAvailableValues = true;
  let showExpandTable = false;
  $: if (sortedQuery && !$sortedQuery?.isFetching) {
    const leaderboardData = prepareLeaderboardItemData(
      $sortedQuery?.data?.rows?.map((r) =>
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

{#if $sortedQuery !== undefined}
  <div
    role="grid"
    tabindex="0"
    style:width="315px"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
  >
    <LeaderboardHeader
      isFetching={$sortedQuery.isFetching}
      {dimensionName}
      {hovered}
    />
    {#if aboveTheFold || selectedBelowTheFold}
      <div class="rounded-b border-gray-200 surface text-gray-800">
        <!-- place the leaderboard entries that are above the fold here -->
        {#each aboveTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardListItem {dimensionName} {itemData} on:click on:keydown />
        {/each}
        <!-- place the selected values that are not above the fold here -->
        {#if selectedBelowTheFold?.length}
          <hr />
          {#each selectedBelowTheFold as itemData (itemData.dimensionValue)}
            <LeaderboardListItem
              {dimensionName}
              {itemData}
              on:click
              on:keydown
            />
          {/each}

          <hr />
        {/if}
        {#if $sortedQuery?.isError}
          <div class="text-red-500">
            {JSON.stringify($sortedQuery?.error)}
          </div>
        {:else if noAvailableValues}
          <div style:padding-left="30px" class="p-1 ui-copy-disabled">
            no available values
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
{/if}
