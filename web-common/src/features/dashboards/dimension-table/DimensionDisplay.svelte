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
    createQueryServiceMetricsViewComparison,
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    getDimensionFilterWithSearch,
    prepareDimensionTableRows,
    prepareVirtualizedDimTableColumns,
  } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import {
    getDimensionColumn,
    isSummableMeasure,
    prepareSortedQueryBody,
  } from "../dashboard-utils";
  import { metricsExplorerStore } from "../stores/dashboard-stores";

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
  const stateManagers = getStateManagers();
  const timeControlsStore = useTimeControlStore(stateManagers);

  const {
    dashboardStore,
    selectors: {
      sorting: { sortedAscending },
    },
  } = stateManagers;

  $: leaderboardMeasureName = $dashboardStore?.leaderboardMeasureName;
  $: isBeingCompared =
    $dashboardStore?.selectedComparisonDimension === dimensionName;

  $: leaderboardMeasureQuery = useMetaMeasure(
    instanceId,
    metricViewName,
    leaderboardMeasureName
  );

  $: validPercentOfTotal = (
    $leaderboardMeasureQuery?.data as MetricsViewSpecMeasureV2
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

  $: visibleMeasures = $metaQuery.data?.measures.filter((m) =>
    $dashboardStore?.visibleMeasureKeys.has(m.name)
  );

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

  let referenceValues: { [key: string]: number } = {};
  $: if ($totalsQuery?.data?.data) {
    visibleMeasures.map((m) => {
      if (isSummableMeasure(m)) {
        referenceValues[m.name] = $totalsQuery.data.data?.[m.name];
      }
    });
  }

  $: columns = prepareVirtualizedDimTableColumns(
    $dashboardStore,
    visibleMeasures,
    referenceValues,
    dimension,
    $timeControlsStore.showComparison,
    validPercentOfTotal
  );

  $: sortedQueryBody = prepareSortedQueryBody(
    dimensionName,
    $dashboardStore?.selectedMeasureNames,
    $timeControlsStore,
    leaderboardMeasureName,
    $dashboardStore.dashboardSortType,
    $sortedAscending,
    filterSet
  );

  $: sortedQuery = createQueryServiceMetricsViewComparison(
    $runtime.instanceId,
    metricViewName,
    sortedQueryBody,
    {
      query: {
        enabled: $timeControlsStore.ready && !!filterSet,
      },
    }
  );

  $: tableRows = prepareDimensionTableRows(
    $sortedQuery?.data?.rows,
    $metaQuery.data?.measures,
    leaderboardMeasureName,
    dimensionColumn,
    $timeControlsStore.showComparison,
    validPercentOfTotal,
    unfilteredTotal
  );

  function onSelectItem(event) {
    const label = tableRows[event.detail][dimensionColumn] as string;
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(metricViewName, dimensionName, label);
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

    {#if tableRows && columns.length}
      <div class="grow" style="overflow-y: hidden;">
        <DimensionTable
          on:select-item={(event) => onSelectItem(event)}
          on:toggle-dimension-comparison={() =>
            toggleComparisonDimension(dimensionName, isBeingCompared)}
          isFetching={$sortedQuery?.isFetching}
          dimensionName={dimensionColumn}
          {isBeingCompared}
          {columns}
          {selectedValues}
          rows={tableRows}
          {excludeMode}
        />
      </div>
    {/if}
  </div>
{/if}
