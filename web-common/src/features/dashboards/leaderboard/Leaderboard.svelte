<script lang="ts">
  /**
   * Leaderboard.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    getFilterForDimension,
    useMetaDimension,
    useMetaMeasure,
    useMetaQuery,
    useModelAllTimeRange,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    createQueryServiceMetricsViewToplist,
    MetricsViewDimension,
    MetricsViewMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { isRangeInsideOther } from "../../../lib/time/ranges";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import { getFilterForComparsion } from "../dimension-table/dimension-table-utils";
  import type { NicelyFormattedTypes } from "../humanize-numbers";
  import DimensionLeaderboardEntrySet from "./DimensionLeaderboardEntrySet.svelte";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardList from "./LeaderboardList.svelte";
  import LeaderboardListItem from "./LeaderboardListItem.svelte";

  export let metricViewName: string;
  export let dimensionName: string;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */
  export let referenceValue: number;

  export let formatPreset: NicelyFormattedTypes;
  export let isSummableMeasure = false;

  let slice = 7;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricViewName);
  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);

  $: allTimeRangeQuery = useModelAllTimeRange(
    $runtime.instanceId,
    $metaQuery.data.model,
    $metaQuery.data.timeDimension,
    {
      query: {
        enabled: !!$metaQuery.data.timeDimension,
      },
    }
  );

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

  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  function toggleFilterMode() {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilterMode(metricViewName, dimensionName);
  }

  function selectDimension(dimensionName) {
    metricsExplorerStore.setMetricDimensionName(metricViewName, dimensionName);
  }

  $: timeStart = $dashboardStore?.selectedTimeRange?.start
    ? $dashboardStore.selectedTimeRange.start.toISOString()
    : undefined;
  $: timeEnd = $dashboardStore?.selectedTimeRange?.end
    ? $dashboardStore.selectedTimeRange.end.toISOString()
    : undefined;
  $: topListQuery = createQueryServiceMetricsViewToplist(
    $runtime.instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: [measure.name],
      timeStart: timeStart,
      timeEnd: timeEnd,
      filter: filterForDimension,
      limit: "250",
      offset: "0",
      sort: [
        {
          name: measure.name,
          ascending: false,
        },
      ],
    },
    {
      query: {
        enabled:
          (hasTimeSeries ? !!timeStart && !!timeEnd : true) &&
          !!filterForDimension,
      },
    }
  );

  let values = [];
  let comparisonValues = [];

  /** replace data after fetched. */
  $: if (!$topListQuery?.isFetching) {
    values =
      $topListQuery?.data?.data.map((val) => ({
        value: val[measure?.name],
        label: val[dimension?.name],
      })) ?? [];
  }

  // get all values that are selected but not visible.
  // we'll put these at the bottom w/ a divider.
  $: selectedValuesThatAreBelowTheFold = activeValues
    ?.filter((label) => {
      return (
        // the value is visible within the fold.
        !values.slice(0, slice).some((value) => {
          return value.label === label;
        })
      );
    })
    .map((label) => {
      const existingValue = values.find((value) => value.label === label);
      // return the existing value, or if it does not exist, just return the label.
      // FIX ME return values for label which are not in the query
      return existingValue ? { ...existingValue } : { label };
    })
    .sort((a, b) => {
      return b.value - a.value;
    });

  // Compose the comparison /toplist query
  $: displayComparison =
    $dashboardStore?.showComparison && isComparisonRangeAvailable;
  $: isComparisonRangeAvailable = isRangeInsideOther(
    $allTimeRangeQuery?.data?.start,
    $allTimeRangeQuery?.data?.end,
    $dashboardStore?.selectedComparisonTimeRange?.start,
    $dashboardStore?.selectedComparisonTimeRange?.end
  );
  // add all sliced and active values to the include filter.
  $: currentVisibleValues =
    $topListQuery?.data?.data
      ?.slice(0, slice)
      ?.concat(selectedValuesThatAreBelowTheFold)
      ?.map((v) => v[dimensionName]) ?? [];
  $: updatedFilters = getFilterForComparsion(
    filterForDimension,
    dimensionName,
    currentVisibleValues
  );
  $: comparisonTimeStart =
    $dashboardStore?.selectedComparisonTimeRange?.start?.toISOString();
  $: comparisonTimeEnd =
    $dashboardStore?.selectedComparisonTimeRange?.end?.toISOString();
  $: comparisonTopListQuery = createQueryServiceMetricsViewToplist(
    $runtime.instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: [measure.name],
      timeStart: comparisonTimeStart,
      timeEnd: comparisonTimeEnd,
      filter: updatedFilters,
      limit: currentVisibleValues.length.toString(),
      offset: "0",
      sort: [
        {
          name: measure.name,
          ascending: false,
        },
      ],
    },
    {
      query: {
        enabled: Boolean(
          displayComparison &&
            !!comparisonTimeStart &&
            !!comparisonTimeEnd &&
            !!updatedFilters
        ),
      },
    }
  );

  $: if (!$comparisonTopListQuery?.isFetching) {
    comparisonValues =
      $comparisonTopListQuery?.data?.data?.map((val) => ({
        value: val[measure?.name],
        label: val[dimension?.name],
      })) ?? [];
  }

  let hovered: boolean;
</script>

{#if topListQuery}
  <div
    style:width="315px"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
  >
    <LeaderboardHeader
      isFetching={$topListQuery.isFetching}
      {displayName}
      on:toggle-filter-mode={toggleFilterMode}
      {filterExcludeMode}
      {hovered}
      dimensionDescription={dimension?.description}
      on:click={() => selectDimension(dimensionName)}
    />
    {#if values}
      <LeaderboardList>
        <!-- place the leaderboard entries that are above the fold here -->
        <DimensionLeaderboardEntrySet
          {formatPreset}
          loading={$topListQuery?.isFetching}
          values={values.slice(0, slice)}
          {comparisonValues}
          showComparison={displayComparison}
          {activeValues}
          {filterExcludeMode}
          {atLeastOneActive}
          {referenceValue}
          {isSummableMeasure}
          on:select-item
        />
        <!-- place the selected values that are not above the fold here -->
        {#if selectedValuesThatAreBelowTheFold?.length}
          <hr />
          <DimensionLeaderboardEntrySet
            {formatPreset}
            loading={$topListQuery?.isFetching}
            values={selectedValuesThatAreBelowTheFold}
            {comparisonValues}
            showComparison={displayComparison}
            {activeValues}
            {filterExcludeMode}
            {atLeastOneActive}
            {referenceValue}
            {isSummableMeasure}
            on:select-item
          />
          <hr />
        {/if}
        {#if $topListQuery?.isError}
          <div class="text-red-500">
            {$topListQuery?.error}
          </div>
        {:else if values.length === 0}
          <div style:padding-left="30px" class="p-1 ui-copy-disabled">
            no available values
          </div>
        {/if}

        {#if values.length > slice}
          <Tooltip location="right">
            <LeaderboardListItem
              value={0}
              color="=ui-label"
              on:click={() => selectDimension(dimensionName)}
            >
              <div class="ui-copy-muted" slot="title">(Expand Table)</div>
            </LeaderboardListItem>
            <TooltipContent slot="tooltip-content"
              >Expand dimension to see more values</TooltipContent
            >
          </Tooltip>
        {/if}
      </LeaderboardList>
    {/if}
  </div>
{/if}
