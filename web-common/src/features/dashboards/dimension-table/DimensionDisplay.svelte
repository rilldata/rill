<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    getFilterForDimension,
    useMetaDimension,
    useMetaMeasure,
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceMetricsViewComparisonToplist,
    createQueryServiceMetricsViewToplist,
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
    MetricsViewMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { SortDirection } from "../proto-state/derived-types";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import {
    getFilterForComparisonTable,
    prepareDimensionTableRows,
    prepareVirtualizedTableColumns,
    updateFilterOnSearch,
  } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import {
    getDimensionColumn,
    isSummableMeasure,
    prepareSortedQueryBody,
  } from "../dashboard-utils";

  import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
  // import { getQuerySortType } from "../leaderboard/leaderboard-utils";

  export let metricViewName: string;
  export let dimensionName: string;

  let searchText = "";

  const queryClient = useQueryClient();

  $: instanceId = $runtime.instanceId;

  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: dimensionQuery = useMetaDimension(
    instanceId,
    metricViewName,
    dimensionName
  );
  let dimension: MetricsViewDimension;
  $: dimension = $dimensionQuery?.data;
  $: dimensionColumn = getDimensionColumn(dimension);

  $: dashboardStore = useDashboardStore(metricViewName);

  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: leaderboardMeasureName = $dashboardStore?.leaderboardMeasureName;
  $: isBeingCompared =
    $dashboardStore?.selectedComparisonDimension === dimensionName;
  $: leaderboardMeasureQuery = useMetaMeasure(
    instanceId,
    metricViewName,
    leaderboardMeasureName
  );

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: filterForDimension = getFilterForDimension(
    $dashboardStore?.filters,
    dimensionName
  );

  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;

  let selectedValues: Array<unknown>;
  $: selectedValues =
    (excludeMode
      ? $dashboardStore?.filters.exclude.find((d) => d.name === dimension?.name)
          ?.in
      : $dashboardStore?.filters.include.find((d) => d.name === dimension?.name)
          ?.in) ?? [];

  $: console.log("filters", $dashboardStore?.filters);

  $: allMeasures = $metaQuery.data?.measures.filter((m) =>
    $dashboardStore?.visibleMeasureKeys.has(m.name)
  );

  $: sortAscending = $dashboardStore.sortDirection === SortDirection.ASCENDING;

  $: metricTimeSeries = useModelHasTimeSeries(instanceId, metricViewName);
  $: hasTimeSeries = $metricTimeSeries.data;

  $: filterSet = updateFilterOnSearch(
    filterForDimension,
    searchText,
    dimension?.name
  );
  $: topListQuery = createQueryServiceMetricsViewToplist(
    instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: selectedMeasureNames,
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
      filter: filterSet,
      limit: "250",
      offset: "0",
      sort: [
        {
          name: leaderboardMeasureName,
          ascending: sortAscending,
        },
      ],
    },
    {
      query: {
        enabled:
          $timeControlsStore.ready && !!filterSet && !!leaderboardMeasureName,
      },
    }
  );

  // Compose the comparison /toplist query
  $: timeComparison = $timeControlsStore.showComparison;
  $: comparisonFilterSet = getFilterForComparisonTable(
    filterForDimension,
    dimensionName,
    dimensionColumn,
    $topListQuery?.data?.data
  );
  $: comparisonTopListQuery = createQueryServiceMetricsViewToplist(
    $runtime.instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: [leaderboardMeasureName],
      timeStart: $timeControlsStore.comparisonTimeStart,
      timeEnd: $timeControlsStore.comparisonTimeEnd,
      filter: comparisonFilterSet,
      limit: "250",
      offset: "0",
      sort: [
        {
          name: leaderboardMeasureName,
          ascending: sortAscending,
        },
      ],
    },
    {
      query: {
        enabled: Boolean(
          $timeControlsStore.showComparison && !!comparisonFilterSet
        ),
      },
    }
  );

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
    },
    {
      query: {
        enabled: $timeControlsStore.ready,
      },
    }
  );

  let referenceValues: { string: number } = {};
  $: if ($totalsQuery?.data?.data) {
    allMeasures.map((m) => {
      if (isSummableMeasure(m)) {
        referenceValues[m.name] = $totalsQuery.data.data?.[m.name];
      }
    });
  }

  let columns: VirtualizedTableColumns[] = [];

  $: if (!$topListQuery?.isFetching && dimension) {
    columns = prepareVirtualizedTableColumns(
      allMeasures,
      leaderboardMeasureName,
      referenceValues,
      dimension,
      $topListQuery?.data?.meta || [],
      timeComparison,
      validPercentOfTotal,
      $dashboardStore.visibleMeasureKeys
    );
  }

  $: validPercentOfTotal = (
    $leaderboardMeasureQuery?.data as MetricsViewMeasure
  )?.validPercentOfTotal;

  //////////////////////////// SORTED QUERY

  $: sortedQueryBody = prepareSortedQueryBody(
    dimensionName,
    selectedMeasureNames,
    $timeControlsStore,
    leaderboardMeasureName,
    $dashboardStore.dashboardSortType,
    sortAscending,
    filterForDimension
  );

  $: sortedQuery = createQueryServiceMetricsViewComparisonToplist(
    $runtime.instanceId,
    metricViewName,
    sortedQueryBody,
    {
      query: {
        enabled: $timeControlsStore.ready && !!filterForDimension,
      },
    }
  );

  $: unfilteredTotal = $totalsQuery.data?.data?.[leaderboardMeasureName];

  $: newRows = $sortedQuery?.isFetching
    ? []
    : prepareDimensionTableRows(
        $sortedQuery?.data?.rows,
        allMeasures,
        leaderboardMeasureName,
        dimensionColumn,
        timeComparison,
        validPercentOfTotal,
        unfilteredTotal
      );

  ////////////////////////////

  function onSelectItem(event) {
    const label = newRows[event.detail][dimensionColumn];
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(metricViewName, dimension?.name, label);
  }

  function onSortByColumn(event) {
    const columnName = event.detail;
    if (!allMeasures.map((m) => m.name).includes(columnName)) return;

    if (columnName === leaderboardMeasureName) {
      metricsExplorerStore.toggleSort(metricViewName);
    } else {
      metricsExplorerStore.setLeaderboardMeasureName(
        metricViewName,
        columnName
      );
      metricsExplorerStore.setSortDescending(metricViewName);
    }
  }

  function toggleComparisonDimension(dimensionName, isBeingCompared) {
    metricsExplorerStore.setComparisonDimension(
      metricViewName,
      isBeingCompared ? undefined : dimensionName
    );
  }
</script>

{#if topListQuery}
  <div class="h-full flex flex-col" style:min-width="365px">
    <div class="flex-none" style:height="50px">
      <DimensionHeader
        {metricViewName}
        {dimensionName}
        {excludeMode}
        isFetching={$topListQuery?.isFetching}
        on:search={(event) => {
          searchText = event.detail;
        }}
      />
    </div>

    {#if newRows && columns.length}
      <div class="grow" style="overflow-y: hidden;">
        <DimensionTable
          on:select-item={(event) => onSelectItem(event)}
          on:sort={(event) => onSortByColumn(event)}
          on:toggle-dimension-comparison={() =>
            toggleComparisonDimension(dimensionName, isBeingCompared)}
          {sortAscending}
          dimensionName={dimensionColumn}
          {isBeingCompared}
          {columns}
          {selectedValues}
          rows={newRows}
          sortByColumn={leaderboardMeasureName}
          {excludeMode}
        />
      </div>
    {/if}
  </div>
{/if}
