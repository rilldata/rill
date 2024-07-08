<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
  import { SortType } from "../proto-state/derived-types";
  import {
    LeaderboardItemData,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import { onMount } from "svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import LeaderboardRow from "./LeaderboardRow.svelte";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import DimensionCompareMenu from "./DimensionCompareMenu.svelte";

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
      contextColumn: { isPercentOfTotal },
      dimensions: {
        getDimensionByName,
        getDimensionDisplayName,
        getDimensionDescription,
      },
      activeMeasure: { activeMeasureName, isValidPercentOfTotal },
      dimensionFilters: {
        selectedDimensionValues,
        atLeastOneSelection,
        isFilterExcludeMode,
      },
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
      sorting: { toggleSort, toggleSortByActiveContextColumn },
      dimensions: { setPrimaryDimension },
    },
    metricsViewName,
    runtime,
  } = getStateManagers();

  $: flip = $sortedAscending;

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

  $: isBeingCompared = $isBeingComparedReadable(dimensionName);
  $: filterExcludeMode = $isFilterExcludeMode(dimensionName);
  $: atLeastOneActive = $atLeastOneSelection(dimensionName);

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  // $: excluded = atLeastOneActive
  //   ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
  //   : false;

  let hovered: boolean;
  $: arrowTransform = $sortedAscending ? "scale(1 -1)" : "scale(1 1)";
  $: dimensionDescription = $getDimensionDescription(dimensionName);

  $: tableWidth = 1 * 60 + 190;
</script>

<div
  class="flex-col flex"
  bind:this={container}
  aria-label="{dimensionName} leaderboard"
  role="table"
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
>
  <table>
    <colgroup>
      <col style:width="24px" />
      <col style:width="190px" />
      <col class="col-width" />
      {#if $isTimeComparisonActive}
        <col class="col-width" />
        <col class="col-width" />
      {:else}
        <col class="col-width" />
      {/if}
    </colgroup>
    <thead>
      <tr>
        <th>
          {#if isFetching}
            <Spinner size="16px" status={EntityStatus.Running} />
          {:else if hovered || isBeingCompared}
            <DimensionCompareMenu {dimensionName} />
          {/if}
        </th>
        <th>
          <Tooltip distance={16} location="top">
            <button
              on:click={() => setPrimaryDimension(dimensionName)}
              class="ui-header-primary header-cell"
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
            class="header-cell"
            on:click={() => toggleSort(SortType.VALUE)}
            aria-label="Toggle sort leaderboards by value"
          >
            #{#if $sortType === SortType.VALUE}
              <ArrowDown {flip} />
            {/if}
          </button>
        </th>
        {#if $isTimeComparisonActive}
          <th>
            <button
              class="header-cell"
              on:click={() => toggleSort(SortType.DELTA_ABSOLUTE)}
              aria-label="Toggle sort leaderboards by absolute change"
            >
              <Delta />
              {#if $sortType === SortType.DELTA_ABSOLUTE}
                <ArrowDown {flip} />
              {/if}
            </button>
          </th>

          <th>
            <button
              class="header-cell"
              on:click={() => toggleSort(SortType.DELTA_PERCENT)}
              aria-label="Toggle sort leaderboards by percent change"
            >
              <Delta /> %
              {#if $sortType === SortType.DELTA_PERCENT}
                <ArrowDown {flip} />
              {/if}
            </button>
          </th>
        {:else if $isValidPercentOfTotal}
          <th>
            <button
              on:click={() => toggleSort(SortType.PERCENT)}
              class="header-cell"
              aria-label="Toggle sort leaderboards by percent of total"
            >
              <PieChart /> %
              {#if $sortType === SortType.PERCENT}
                <ArrowDown {flip} />
              {/if}
            </button>
          </th>
        {/if}
      </tr>
    </thead>

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
        {#each { length: 7 } as _, i (i)}
          <tr>
            <td></td>
            <td>
              <div class="loading-bar" />
            </td>
            <td>
              <div class="loading-bar" />
            </td>
          </tr>
        {/each}
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
        class="block flex-row w-full text-left transition-color ui-copy-muted"
        style:padding-left="30px"
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

  td {
    @apply text-right truncate;
    height: 22px;
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
  th:not(:first-of-type) {
    @apply border-b;
  }

  .loading-bar {
    @apply w-11/12 h-2.5 bg-gray-100 animate-pulse rounded-full;
  }

  td:first-of-type {
    @apply text-left;
  }

  .header-cell {
    @apply px-2 flex items-center justify-end;
  }

  th:nth-of-type(2) .header-cell {
    @apply justify-start;
  }

  .col-width {
    width: 64px;
  }
</style>
