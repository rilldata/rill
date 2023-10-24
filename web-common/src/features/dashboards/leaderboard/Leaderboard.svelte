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
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { SortDirection } from "../proto-state/derived-types";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import {
    LeaderboardItemData,
    getLabeledComparisonFromComparisonRow,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import LeaderboardListItem from "./LeaderboardListItem.svelte";
  import { prepareSortedQueryBody } from "../dashboard-utils";

  export let metricViewName: string;
  export let dimensionName: string;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */
  export let referenceValue: number;
  export let unfilteredTotal: number;

  let slice = 7;

  const stateManagers = getStateManagers();

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
  $: displayName = dimension?.label || dimension?.name || dimensionName;

  $: measureQuery = useMetaMeasure(
    $runtime.instanceId,
    metricViewName,
    $dashboardStore?.leaderboardMeasureName
  );
  let measure: MetricsViewSpecMeasureV2;
  $: measure = $measureQuery?.data;

  $: filterForDimension = getFilterForDimension(
    $dashboardStore?.filters,
    dimensionName
  );

  // FIXME: it is possible for this way of accessing the filters
  // to return the same value twice, which would seem to indicate
  // a bug in the way we're setting the filters / active values.
  // Need to investigate further to determine whether this is a
  // problem with the runtime or the client, but for now wrapping
  // it in a set dedupes the values.
  $: activeValues = new Set(
    ($dashboardStore?.filters[filterKey]?.find(
      (d) => d.name === dimension?.name
    )?.in as (number | string)[]) ?? []
  );
  $: atLeastOneActive = activeValues?.size > 0;

  const timeControlsStore = useTimeControlStore(stateManagers);

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

  $: sortAscending = $dashboardStore.sortDirection === SortDirection.ASCENDING;
  $: sortType = $dashboardStore.dashboardSortType;

  $: sortedQueryBody = prepareSortedQueryBody(
    dimensionName,
    [measure?.name],
    $timeControlsStore,
    measure?.name,
    sortType,
    sortAscending,
    filterForDimension
  );

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
      [...activeValues],
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
