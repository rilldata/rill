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
  import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
  import { SortType } from "../proto-state/derived-types";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardListItem from "./LeaderboardListItem.svelte";
  import {
    LeaderboardItemData,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import { onMount } from "svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";
  import LeaderboardValueCell from "./LeaderboardValueCell.svelte";
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import LeaderboardRow from "./LeaderboardRow.svelte";

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
      contextColumn: {
        contextColumn,
        isDeltaAbsolute,
        isDeltaPercent,
        isPercentOfTotal,
        isHidden,
      },
      dimensions: {
        getDimensionByName,
        getDimensionDisplayName,
        getDimensionDescription,
      },
      activeMeasure: { activeMeasureName },
      dimensionFilters: { selectedDimensionValues },
      dashboardQueries: {
        leaderboardSortedQueryBody,
        leaderboardSortedQueryOptions,
        leaderboardDimensionTotalQueryBody,
        leaderboardDimensionTotalQueryOptions,
      },
      sorting: { sortedAscending, sortType },
      timeRangeSelectors: { isTimeComparisonActive },
    },
    actions: {
      sorting: { toggleSort, toggleSortByActiveContextColumn },
      dimensions: { setPrimaryDimension },
    },
    metricsViewName,
    runtime,
  } = getStateManagers();

  $: dimension = $getDimensionByName(dimensionName);

  $: sortedQuery = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    $metricsViewName,
    $leaderboardSortedQueryBody(dimensionName),
    $leaderboardSortedQueryOptions(dimensionName, visible),
  );

  $: ({
    isLoading,
    isError,
    data: sortedData,
    refetch,
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
  let noAvailableValues = true;
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
    noAvailableValues = leaderboardData.noAvailableValues;
    showExpandTable = leaderboardData.showExpandTable;
  }

  let hovered: boolean;
  $: arrowTransform = $sortedAscending ? "scale(1 -1)" : "scale(1 1)";
  $: dimensionDescription = $getDimensionDescription(dimensionName);
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

<table>
  <colgroup>
    <col style:width="20px" />
    <col style:width="184px" />
    <col style:width="50px" />
    <col style:width="50px" />
    <col style:width="50px" />
  </colgroup>
  <thead>
    <tr>
      <th> </th>
      <th>
        <Tooltip distance={16} location="top">
          <button
            on:click={() => setPrimaryDimension(dimensionName)}
            class="ui-header-primary"
            aria-label="Open dimension details"
          >
            {$getDimensionDisplayName(dimensionName)}
          </button>
          <TooltipContent slot="tooltip-content">
            <TooltipTitle>
              <svelte:fragment slot="name">
                {$getDimensionDisplayName(dimensionName)}
              </svelte:fragment>
              <svelte:fragment slot="description" />
            </TooltipTitle>
            <TooltipShortcutContainer>
              <div>
                {#if dimensionDescription}
                  {dimensionDescription}
                {:else}
                  The leaderboard metrics for {$getDimensionDisplayName(
                    dimensionName,
                  )}
                {/if}
              </div>
              <Shortcut />
              <div>Expand leaderboard</div>
              <Shortcut>Click</Shortcut>
            </TooltipShortcutContainer>
          </TooltipContent>
        </Tooltip>
      </th>
      <th>
        <button
          on:click={() => toggleSort(SortType.VALUE)}
          aria-label="Toggle sort leaderboards by value"
        >
          #{#if $sortType === SortType.VALUE}
            <ArrowDown transform={arrowTransform} />
          {/if}
        </button>
      </th>
      {#if $isTimeComparisonActive}
        <th>
          <button
            on:click={toggleSortByActiveContextColumn}
            aria-label="Toggle sort leaderboards by context column"
          >
            <Delta /> %
          </button>
        </th>
        <th>
          <button
            on:click={toggleSortByActiveContextColumn}
            class="size-full"
            aria-label="Toggle sort leaderboards by context column"
          >
            <Delta />
          </button>
        </th>
      {:else if $isPercentOfTotal}
        <th>
          <button
            on:click={toggleSortByActiveContextColumn}
            class="flex flex-row items-center justify-end"
            aria-label="Toggle sort leaderboards by context column"
          >
            <PieChart /> %
          </button>
        </th>
      {/if}
    </tr>
  </thead>
  <tbody>
    {#each aboveTheFold as itemData (itemData.dimensionValue)}
      <LeaderboardRow
        {dimensionName}
        {itemData}
        isPercentOfTotal={$isPercentOfTotal}
        isTimeComparisonActive={$isTimeComparisonActive}
      />
    {/each}
    <!-- place the selected values that are not above the fold here -->
    {#if selectedBelowTheFold?.length}
      <hr />
      {#each selectedBelowTheFold as itemData (itemData.dimensionValue)}
        <LeaderboardListItem {dimensionName} {itemData} on:click on:keydown />
      {/each}
      <hr />
    {/if}
  </tbody>
</table>

<style lang="postcss">
  table {
    @apply p-0 m-0 border-spacing-0 border-collapse w-fit;
    @apply font-normal cursor-pointer select-none;
    /* @apply table-fixed; */
  }

  td {
    @apply text-right truncate;
    height: 22px;
  }

  td:first-of-type {
    @apply text-left;
  }

  tbody tr:hover {
    @apply bg-gray-100;
  }

  td,
  th {
    @apply p-0;
  }

  th {
    @apply text-right;
    height: 32px;
  }

  th:not(:nth-of-type(2)) button {
    @apply justify-end;
  }

  button {
    @apply size-full flex items-center;
  }

  th:first-of-type {
    @apply text-left;
  }
  thead {
    @apply border-b;
  }
</style>
