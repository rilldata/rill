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
    useMetaQuery,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceMetricsViewComparison,
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    getDimensionFilterWithSearch,
    prepareDimensionTableRows,
  } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import { getDimensionColumn } from "../dashboard-utils";
  import { metricsExplorerStore } from "../stores/dashboard-stores";

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      dashboardQueries: { dimensionTableSortedQueryBody },
      activeMeasure: { isValidPercentOfTotal },
      comparison: { isBeingCompared },
      dimensions: { dimensionTableDimName },
      dimensionTable: { virtualizedTableColumns },
    },
    metricsViewName,
  } = stateManagers;

  // cast is safe because dimensionTableDimName must be defined
  // for the dimension table to be open
  $: dimensionName = $dimensionTableDimName as string;

  let searchText = "";

  const queryClient = useQueryClient();

  $: instanceId = $runtime.instanceId;

  $: metaQuery = useMetaQuery(instanceId, $metricsViewName);

  $: dimensionQuery = useMetaDimension(
    instanceId,
    $metricsViewName,
    dimensionName
  );

  let dimension: MetricsViewDimension;
  $: dimension = $dimensionQuery?.data as MetricsViewDimension;
  $: dimensionColumn = getDimensionColumn(dimension);
  const timeControlsStore = useTimeControlStore(stateManagers);

  $: leaderboardMeasureName = $dashboardStore?.leaderboardMeasureName;

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: filterSet = getDimensionFilterWithSearch(
    $dashboardStore?.filters,
    searchText,
    dimensionName
  );

  $: selectedValues =
    (excludeMode
      ? $dashboardStore?.filters?.exclude?.find((d) => d.name === dimensionName)
          ?.in
      : $dashboardStore?.filters?.include?.find((d) => d.name === dimensionName)
          ?.in) ?? [];

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    $metricsViewName,
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

  $: columns = $virtualizedTableColumns($totalsQuery);

  $: sortedQuery = createQueryServiceMetricsViewComparison(
    $runtime.instanceId,
    $metricsViewName,
    $dimensionTableSortedQueryBody,
    {
      query: {
        enabled: $timeControlsStore.ready && !!filterSet,
      },
    }
  );

  $: tableRows = prepareDimensionTableRows(
    $sortedQuery?.data?.rows ?? [],
    $metaQuery.data?.measures ?? [],
    leaderboardMeasureName,
    dimensionColumn,
    $timeControlsStore.showComparison ?? false,
    $isValidPercentOfTotal,
    unfilteredTotal
  );

  function onSelectItem(event) {
    const label = tableRows[event.detail][dimensionColumn] as string;
    cancelDashboardQueries(queryClient, $metricsViewName);
    metricsExplorerStore.toggleFilter($metricsViewName, dimensionName, label);
  }

  function toggleComparisonDimension(dimensionName, isBeingCompared) {
    metricsExplorerStore.setComparisonDimension(
      $metricsViewName,
      isBeingCompared ? undefined : dimensionName
    );
  }
</script>

{#if sortedQuery}
  <div class="h-full flex flex-col" style:min-width="365px">
    <div class="flex-none" style:height="50px">
      <DimensionHeader
        {dimensionName}
        {excludeMode}
        isFetching={$sortedQuery?.isFetching}
        on:search={(event) => {
          searchText = event.detail;
        }}
      />
    </div>

    {#if tableRows && columns.length && dimensionColumn}
      <div class="grow" style="overflow-y: hidden;">
        <DimensionTable
          on:select-item={(event) => onSelectItem(event)}
          on:toggle-dimension-comparison={() =>
            toggleComparisonDimension(dimensionName, isBeingCompared)}
          isFetching={$sortedQuery?.isFetching}
          dimensionName={dimensionColumn}
          {columns}
          {selectedValues}
          rows={tableRows}
          {excludeMode}
        />
      </div>
    {/if}
  </div>
{/if}
