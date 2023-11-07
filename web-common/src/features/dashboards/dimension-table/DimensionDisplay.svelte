<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";

  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceMetricsViewComparison,
    createQueryServiceMetricsViewTotals,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getDimensionFilterWithSearch } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import { metricsExplorerStore } from "../stores/dashboard-stores";

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      dashboardQueries: {
        dimensionTableSortedQueryBody,
        dimensionTableTotalQueryBody,
      },
      comparison: { isBeingCompared },
      dimensions: { dimensionTableDimName },
      dimensionTable: {
        virtualizedTableColumns,
        selectedDimensionValueNames,
        prepareDimTableRows,
      },
      activeMeasure: { activeMeasureName },
    },
    metricsViewName,
    runtime,
  } = stateManagers;

  // cast is safe because dimensionTableDimName must be defined
  // for the dimension table to be open
  $: dimensionName = $dimensionTableDimName as string;

  let searchText = "";

  const queryClient = useQueryClient();

  $: instanceId = $runtime.instanceId;

  const timeControlsStore = useTimeControlStore(stateManagers);

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: filterSet = getDimensionFilterWithSearch(
    $dashboardStore?.filters,
    searchText,
    dimensionName
  );

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    $metricsViewName,
    $dimensionTableTotalQueryBody,
    {
      query: {
        enabled: $timeControlsStore.ready,
      },
    }
  );

  $: unfilteredTotal = $totalsQuery?.data?.data?.[$activeMeasureName] ?? 0;

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

  $: tableRows = $prepareDimTableRows($sortedQuery, unfilteredTotal);

  $: areAllTableRowsSelected = tableRows.every((row) =>
    selectedValues.includes(row[dimensionColumn])
  );

  function onSelectItem(event) {
    const label = tableRows[event.detail][dimensionName] as string;
    cancelDashboardQueries(queryClient, $metricsViewName);
    metricsExplorerStore.toggleFilter($metricsViewName, dimensionName, label);
  }

  function toggleComparisonDimension(dimensionName, isBeingCompared) {
    metricsExplorerStore.setComparisonDimension(
      $metricsViewName,
      isBeingCompared ? undefined : dimensionName
    );
  }

  function toggleAllSearchItems() {
    const labels = tableRows.map((row) => row[dimensionColumn] as string);
    cancelDashboardQueries(queryClient, metricViewName);

    if (areAllTableRowsSelected) {
      metricsExplorerStore.deselectItemsInFilter(
        metricViewName,
        dimensionName,
        labels
      );
      return;
    } else {
      metricsExplorerStore.selectItemsInFilter(
        metricViewName,
        dimensionName,
        labels
      );
    }
  }
</script>

{#if sortedQuery}
  <div class="h-full flex flex-col" style:min-width="365px">
    <div class="flex-none" style:height="50px">
      <DimensionHeader
        {dimensionName}
        {excludeMode}
        {areAllTableRowsSelected}
        isRowsEmpty={!tableRows.length}
        isFetching={$sortedQuery?.isFetching}
        on:search={(event) => {
          searchText = event.detail;
        }}
        on:toggle-all-search-items={() => toggleAllSearchItems()}
      />
    </div>

    {#if tableRows && columns.length && dimensionName}
      <div class="grow" style="overflow-y: hidden;">
        <DimensionTable
          on:select-item={(event) => onSelectItem(event)}
          on:toggle-dimension-comparison={() =>
            toggleComparisonDimension(dimensionName, isBeingCompared)}
          isFetching={$sortedQuery?.isFetching}
          {dimensionName}
          {columns}
          selectedValues={$selectedDimensionValueNames}
          rows={tableRows}
          {excludeMode}
        />
      </div>
    {/if}
  </div>
{/if}
