<script lang="ts">
  import { getDimensionType } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/getDimensionType";

  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { STRING_LIKES } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    createQueryServiceMetricsViewComparison,
    createQueryServiceMetricsViewTotals,
  } from "@rilldata/web-common/runtime-client";
  import { getDimensionFilterWithSearch } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
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
      dimensions: { dimensionTableDimName, dimensionTableColumnName },
      dimensionFilters: { unselectedDimensionValues },
      dimensionTable: {
        virtualizedTableColumns,
        selectedDimensionValueNames,
        prepareDimTableRows,
      },
      activeMeasure: { activeMeasureName },
      measureFilters: { getResolvedFilterForMeasureFilters },
    },
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        selectItemsInFilter,
        deselectItemsInFilter,
      },
    },
    metricsViewName,
    runtime,
  } = stateManagers;

  // cast is safe because dimensionTableDimName must be defined
  // for the dimension table to be open
  $: dimensionName = $dimensionTableDimName as string;
  $: dimensionColumnName = $dimensionTableColumnName(dimensionName) as string;

  let searchText = "";

  $: instanceId = $runtime.instanceId;
  $: dimensionType = getDimensionType(
    instanceId,
    $metricsViewName,
    dimensionName,
  );
  $: stringLikeDimension = STRING_LIKES.has($dimensionType.data ?? "");

  const timeControlsStore = useTimeControlStore(stateManagers);

  $: resolvedFilter = $getResolvedFilterForMeasureFilters;

  $: filterSet = getDimensionFilterWithSearch(
    $dashboardStore?.whereFilter,
    searchText,
    dimensionName,
  );

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    $metricsViewName,
    $dimensionTableTotalQueryBody($resolvedFilter),
    {
      query: {
        enabled: $timeControlsStore.ready && $resolvedFilter.ready,
      },
    },
  );

  $: unfilteredTotal = $totalsQuery?.data?.data?.[$activeMeasureName] ?? 0;

  $: columns = $virtualizedTableColumns($totalsQuery);

  $: sortedQuery = createQueryServiceMetricsViewComparison(
    $runtime.instanceId,
    $metricsViewName,
    $dimensionTableSortedQueryBody($resolvedFilter),
    {
      query: {
        enabled:
          $timeControlsStore.ready && !!filterSet && $resolvedFilter.ready,
      },
    },
  );

  $: tableRows = $prepareDimTableRows($sortedQuery, unfilteredTotal);

  $: areAllTableRowsSelected = tableRows.every((row) =>
    $selectedDimensionValueNames.includes(row[dimensionColumnName] as string),
  );

  function onSelectItem(event) {
    const label = tableRows[event.detail.index][dimensionColumnName] as string;
    toggleDimensionValueSelection(
      dimensionName,
      label,
      false,
      event.detail.meta,
    );
  }

  function toggleComparisonDimension(dimensionName, isBeingCompared) {
    metricsExplorerStore.setComparisonDimension(
      $metricsViewName,
      isBeingCompared ? undefined : dimensionName,
    );
  }

  function toggleAllSearchItems() {
    const labels = tableRows.map((row) => row[dimensionColumnName] as string);

    if (areAllTableRowsSelected) {
      deselectItemsInFilter(dimensionName, labels);

      eventBus.emit("notification", {
        message: `Removed ${labels.length} items from filter`,
      });
      return;
    } else {
      const newValuesSelected = $unselectedDimensionValues(
        dimensionName,
        labels,
      );
      selectItemsInFilter(dimensionName, labels);
      eventBus.emit("notification", {
        message: `Added ${newValuesSelected.length} items to filter`,
      });
    }
  }

  function handleKeyDown(e) {
    // Select all items on Meta+A
    if ((e.ctrlKey || e.metaKey) && e.key === "a") {
      if (e.target.tagName === "INPUT") return;
      e.preventDefault();
      if (areAllTableRowsSelected) return;
      toggleAllSearchItems();
    }
  }
</script>

{#if sortedQuery}
  <div class="h-full flex flex-col w-full" style:min-width="365px">
    <div class="flex-none" style:height="50px">
      <DimensionHeader
        {dimensionName}
        {areAllTableRowsSelected}
        isRowsEmpty={!tableRows.length}
        isFetching={$sortedQuery?.isFetching}
        on:search={(event) => {
          if (stringLikeDimension) searchText = event.detail;
        }}
        on:toggle-all-search-items={() => toggleAllSearchItems()}
        enableSearch={stringLikeDimension}
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
        />
      </div>
    {/if}
  </div>
{/if}

<svelte:window on:keydown={handleKeyDown} />
