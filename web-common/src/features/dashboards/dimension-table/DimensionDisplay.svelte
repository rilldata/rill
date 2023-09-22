<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useMetaDimension,
    useMetaMeasure,
    useMetaQuery,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceMetricsViewComparisonToplist,
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
    MetricsViewMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { SortDirection } from "../proto-state/derived-types";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import {
    getDimensionFilterWithSearch,
    prepareDimensionTableRows,
    prepareVirtualizedTableColumns,
  } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import {
    getDimensionColumn,
    isSummableMeasure,
    prepareSortedQueryBody,
  } from "../dashboard-utils";

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

  $: validPercentOfTotal = (
    $leaderboardMeasureQuery?.data as MetricsViewMeasure
  )?.validPercentOfTotal;

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: filterSet = getDimensionFilterWithSearch(
    $dashboardStore?.filters,
    searchText,
    dimensionName
  );

  $: selectedValues =
    (excludeMode
      ? $dashboardStore?.filters.exclude.find((d) => d.name === dimensionName)
          ?.in
      : $dashboardStore?.filters.include.find((d) => d.name === dimensionName)
          ?.in) ?? [];

  $: allMeasures = $metaQuery.data?.measures.filter((m) =>
    $dashboardStore?.visibleMeasureKeys.has(m.name)
  );

  $: sortAscending = $dashboardStore.sortDirection === SortDirection.ASCENDING;

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    metricViewName,
    {
      measureNames: $dashboardStore?.selectedMeasureNames,
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
    },
    {
      query: {
        enabled: $timeControlsStore.ready,
      },
    }
  );

  $: unfilteredTotal = $totalsQuery?.data?.data?.[leaderboardMeasureName] ?? 0;

  let referenceValues: { string: number } = {};
  $: if ($totalsQuery?.data?.data) {
    allMeasures.map((m) => {
      if (isSummableMeasure(m)) {
        referenceValues[m.name] = $totalsQuery.data.data?.[m.name];
      }
    });
  }

  $: columns = prepareVirtualizedTableColumns(
    allMeasures,
    leaderboardMeasureName,
    referenceValues,
    dimension,
    [...$dashboardStore.visibleMeasureKeys],
    $timeControlsStore.showComparison,
    validPercentOfTotal
  );

  $: sortedQueryBody = prepareSortedQueryBody(
    dimensionName,
    $dashboardStore?.selectedMeasureNames,
    $timeControlsStore,
    leaderboardMeasureName,
    $dashboardStore.dashboardSortType,
    sortAscending,
    filterSet
  );

  $: sortedQuery = createQueryServiceMetricsViewComparisonToplist(
    $runtime.instanceId,
    metricViewName,
    sortedQueryBody,
    {
      query: {
        enabled: $timeControlsStore.ready && !!filterSet,
      },
    }
  );

  $: newRows = prepareDimensionTableRows(
    $sortedQuery?.data?.rows,
    allMeasures,
    leaderboardMeasureName,
    dimensionColumn,
    $timeControlsStore.showComparison,
    validPercentOfTotal,
    unfilteredTotal
  );

  ////////////////////////////

  function onSelectItem(event) {
    const label = newRows[event.detail][dimensionColumn];
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(metricViewName, dimensionName, label);
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

{#if sortedQuery}
  <div class="h-full flex flex-col" style:min-width="365px">
    <div class="flex-none" style:height="50px">
      <DimensionHeader
        {metricViewName}
        {dimensionName}
        {excludeMode}
        isFetching={$sortedQuery?.isFetching}
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
