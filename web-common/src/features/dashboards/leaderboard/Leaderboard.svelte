<script lang="ts">
  /**
   * Leaderboard.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    getFilterForDimension,
    useMetaDimension,
    useMetaMeasure,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceMetricsViewComparisonToplist,
    MetricsViewDimension,
    MetricsViewMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { SortDirection } from "../proto-state/derived-types";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import type { FormatPreset } from "../humanize-numbers";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import {
    LeaderboardItemData,
    getLabeledComparisonFromComparisonRow,
    getQuerySortType,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import LeaderboardListItem from "./LeaderboardListItem.svelte";

  export let metricViewName: string;
  export let dimensionName: string;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */
  export let referenceValue: number;
  export let unfilteredTotal: number;

  export let formatPreset: FormatPreset;
  export let isSummableMeasure = false;

  let slice = 7;

  $: dashboardStore = useDashboardStore(metricViewName);

  let filterExcludeMode: boolean;
  $: filterExcludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;
  let filterKey: "exclude" | "include";
  $: filterKey = filterExcludeMode ? "exclude" : "include";

  $: dimensionQuery = useMetaDimension(
    $runtime.instanceId,
    metricViewName,
    dimensionName
  );
  let dimension: MetricsViewDimension;
  $: dimension = $dimensionQuery?.data;
  $: displayName = dimension?.label || dimension?.name;
  $: dimensionColumn = dimension?.column || dimension?.name;

  $: measureQuery = useMetaMeasure(
    $runtime.instanceId,
    metricViewName,
    $dashboardStore?.leaderboardMeasureName
  );
  let measure: MetricsViewMeasure;
  $: measure = $measureQuery?.data;

  $: filterForDimension = getFilterForDimension(
    $dashboardStore?.filters,
    dimensionName
  );

  let activeValues: Array<unknown>;
  $: activeValues =
    $dashboardStore?.filters[filterKey]?.find((d) => d.name === dimension?.name)
      ?.in ?? [];
  $: atLeastOneActive = !!activeValues?.length;

  const timeControlsStore = useTimeControlStore(getStateManagers());

  function selectDimension(dimensionName) {
    metricsExplorerStore.setMetricDimensionName(metricViewName, dimensionName);
  }

  function toggleComparisonDimension(dimensionName, isBeingCompared) {
    metricsExplorerStore.setComparisonDimension(
      metricViewName,
      isBeingCompared ? undefined : dimensionName
    );
  }

  function toggleSort(evt) {
    metricsExplorerStore.toggleSort(metricViewName, evt.detail);
  }

  $: isBeingCompared =
    $dashboardStore?.selectedComparisonDimension === dimensionName;

  $: isBeingCompared =
    $dashboardStore?.selectedComparisonDimension === dimensionName;

  $: sortAscending = $dashboardStore.sortDirection === SortDirection.ASCENDING;
  $: sortType = $dashboardStore.dashboardSortType;

  $: contextColumn = $dashboardStore?.leaderboardContextColumn;

  $: querySortType = getQuerySortType(sortType);

  $: sortedQueryBody = {
    dimensionName: dimensionName,
    measureNames: [measure?.name],
    baseTimeRange: {
      start: $timeControlsStore.timeStart,
      end: $timeControlsStore.timeEnd,
    },
    comparisonTimeRange: {
      start: $timeControlsStore.comparisonTimeStart,
      end: $timeControlsStore.comparisonTimeEnd,
    },
    sort: [
      {
        ascending: sortAscending,
        measureName: measure?.name,
        type: querySortType,
      },
    ],
    filter: filterForDimension,
    limit: "250",
    offset: "0",
  };

  $: sortedQueryEnabled = $timeControlsStore.ready && !!filterForDimension;

  $: sortedQueryOptions = {
    query: {
      enabled: sortedQueryEnabled,
    },
  };

  $: sortedQuery = createQueryServiceMetricsViewComparisonToplist(
    $runtime.instanceId,
    metricViewName,
    sortedQueryBody,
    sortedQueryOptions
  );

  let aboveTheFold: LeaderboardItemData[] = [];
  let selectedBelowTheFold: LeaderboardItemData[] = [];
  let noAvailableValues = true;
  let showExpandTable = false;
  $: if (!$sortedQuery?.isFetching) {
    const leaderboardData = prepareLeaderboardItemData(
      $sortedQuery?.data?.rows?.map((r) =>
        getLabeledComparisonFromComparisonRow(r, measure.name)
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
</script>

{#if sortedQuery}
  <div
    style:width="315px"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
  >
    <LeaderboardHeader
      {contextColumn}
      isFetching={$sortedQuery.isFetching}
      {displayName}
      on:toggle-dimension-comparison={() =>
        toggleComparisonDimension(dimensionName, isBeingCompared)}
      {isBeingCompared}
      {hovered}
      {sortAscending}
      {sortType}
      dimensionDescription={dimension?.description}
      on:open-dimension-details={() => selectDimension(dimensionName)}
      on:toggle-sort={toggleSort}
    />
    {#if aboveTheFold || selectedBelowTheFold}
      <div class="rounded-b border-gray-200 surface text-gray-800">
        <!-- place the leaderboard entries that are above the fold here -->
        {#each aboveTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardListItem
            {itemData}
            {contextColumn}
            {atLeastOneActive}
            {isBeingCompared}
            {filterExcludeMode}
            {isSummableMeasure}
            {referenceValue}
            {formatPreset}
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
              {contextColumn}
              {atLeastOneActive}
              {isBeingCompared}
              {filterExcludeMode}
              {isSummableMeasure}
              {referenceValue}
              {formatPreset}
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
