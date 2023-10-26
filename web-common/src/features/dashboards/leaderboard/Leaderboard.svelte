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
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { createQueryServiceMetricsViewComparison } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import {
    LeaderboardItemData,
    getLabeledComparisonFromComparisonRow,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import LeaderboardListItem from "./LeaderboardListItem.svelte";
  import { prepareSortedQueryBody } from "../dashboard-utils";

  export let dimensionName: string;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */
  export let referenceValue: number;
  export let unfilteredTotal: number;

  let slice = 7;

  const stateManagers = getStateManagers();

  const {
    selectors: {
      activeMeasure: { activeMeasureName },
      dimensions: { getDimensionByName, getDimensionDisplayName },
      sorting: { sortedAscending, sortType },
      dimensionFilters: {
        getFiltersForOtherDimensions,
        selectedDimensionValues,
        atLeastOneSelection,
      },
    },
    actions,
    metricsViewName,
  } = stateManagers;

  $: dashboardStore = stateManagers.dashboardStore;

  let filterExcludeMode: boolean;
  $: filterExcludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;
  let filterKey: "exclude" | "include";
  $: filterKey = filterExcludeMode ? "exclude" : "include";

  $: dimension = $getDimensionByName(dimensionName);
  $: displayName = $getDimensionDisplayName(dimensionName);
  $: filterForDimension = $getFiltersForOtherDimensions(dimensionName);
  $: activeValues = $selectedDimensionValues(dimensionName);
  $: atLeastOneActive = $atLeastOneSelection(dimensionName);

  const timeControlsStore = useTimeControlStore(stateManagers);

  $: isBeingCompared =
    $dashboardStore?.selectedComparisonDimension === dimensionName;

  $: sortedQueryBody = prepareSortedQueryBody(
    dimensionName,
    [$activeMeasureName],
    $timeControlsStore,
    $activeMeasureName,
    $sortType,
    $sortedAscending,
    filterForDimension
  );

  $: sortedQueryEnabled = $timeControlsStore.ready && !!filterForDimension;

  $: sortedQueryOptions = {
    query: {
      enabled: sortedQueryEnabled,
    },
  };

  $: sortedQuery = sortedQueryBody
    ? createQueryServiceMetricsViewComparison(
        $runtime.instanceId,
        $metricsViewName,
        sortedQueryBody,
        sortedQueryOptions
      )
    : undefined;

  let aboveTheFold: LeaderboardItemData[] = [];
  let selectedBelowTheFold: LeaderboardItemData[] = [];
  let noAvailableValues = true;
  let showExpandTable = false;
  $: if (sortedQuery && !$sortedQuery?.isFetching) {
    const leaderboardData = prepareLeaderboardItemData(
      $sortedQuery?.data?.rows?.map((r) =>
        getLabeledComparisonFromComparisonRow(r, $activeMeasureName)
      ) ?? [],
      slice,
      activeValues,
      unfilteredTotal,
      filterExcludeMode
    );

    aboveTheFold = leaderboardData.aboveTheFold;
    selectedBelowTheFold = leaderboardData.selectedBelowTheFold;
    noAvailableValues = leaderboardData.noAvailableValues;
    showExpandTable = leaderboardData.showExpandTable;
  }

  let hovered: boolean;

  function selectDimension(dimensionName) {
    metricsExplorerStore.setMetricDimensionName(
      $metricsViewName,
      dimensionName
    );
  }

  function toggleComparisonDimension(dimensionName, isBeingCompared) {
    metricsExplorerStore.setComparisonDimension(
      $metricsViewName,
      isBeingCompared ? undefined : dimensionName
    );
  }

  function toggleSort(evt) {
    actions.sorting.toggleSort(evt.detail);
  }
</script>

{#if $sortedQuery !== undefined}
  <div
    style:width="315px"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
  >
    <LeaderboardHeader
      isFetching={$sortedQuery.isFetching}
      {displayName}
      on:toggle-dimension-comparison={() =>
        toggleComparisonDimension(dimensionName, isBeingCompared)}
      {isBeingCompared}
      {hovered}
      dimensionDescription={dimension?.description || ""}
      on:open-dimension-details={() => selectDimension(dimensionName)}
      on:toggle-sort={toggleSort}
    />
    {#if aboveTheFold || selectedBelowTheFold}
      <div class="rounded-b border-gray-200 surface text-gray-800">
        <!-- place the leaderboard entries that are above the fold here -->
        {#each aboveTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardListItem
            {itemData}
            {atLeastOneActive}
            {isBeingCompared}
            {filterExcludeMode}
            {referenceValue}
            on:click
            on:keydown
            on:select-item
          />
        {/each}
        <!-- place the selected values that are not above the fold here -->
        {#if selectedBelowTheFold?.length}
          <hr />
          {#each selectedBelowTheFold as itemData (itemData.dimensionValue)}
            <LeaderboardListItem
              {itemData}
              {atLeastOneActive}
              {isBeingCompared}
              {filterExcludeMode}
              {referenceValue}
              on:click
              on:keydown
              on:select-item
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
              on:click={() => selectDimension(dimensionName)}
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
