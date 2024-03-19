<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceMetricsViewComparison,
    createQueryServiceMetricsViewTotals,
  } from "@rilldata/web-common/runtime-client";
  import { getDimensionFilterWithSearch } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";

  import { notifications } from "@rilldata/web-common/components/notifications";
  import { metricsExplorerStore } from "../stores/dashboard-stores";

  import VirtualTable from "@rilldata/web-common/components/table/VirtualTable.svelte";
  import DimensionTableCell from "./DimensionTableCell.svelte";
  import DimensionTableHeaderCell from "./DimensionTableHeaderCell.svelte";
  import DimensionRowHeader from "./DimensionRowHeader.svelte";

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
      dimensionFilters: { unselectedDimensionValues, isFilterExcludeMode },
      dimensionTable: {
        virtualizedTableColumns,
        selectedDimensionValueNames,
        prepareDimTableRows,
        sortedByDimensionValue,
      },
      sorting: { sortedAscending },
      activeMeasure: { activeMeasureName },
      measureFilters: { getResolvedFilterForMeasureFilters },
    },
    actions: {
      sorting: { sortByDimensionValue },
      dimensionTable: { handleMeasureColumnHeaderClick },
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

  $: console.log($totalsQuery);
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

  function onSelectItem(event: CustomEvent<{ index: number; meta?: boolean }>) {
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

      notifications.send({
        message: `Removed ${labels.length} items from filter`,
      });
      return;
    } else {
      const newValuesSelected = $unselectedDimensionValues(
        dimensionName,
        labels,
      );
      selectItemsInFilter(dimensionName, labels);
      notifications.send({
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

  $: console.log(tableRows);
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
          searchText = event.detail;
        }}
        on:toggle-all-search-items={() => toggleAllSearchItems()}
      />
    </div>

    {#if tableRows.length && columns.length && dimensionName}
      <div class="grow flex" style="overflow-y: hidden;">
        <!-- <DimensionTable
          on:select-item={onSelectItem}
          on:toggle-dimension-comparison={() =>
            toggleComparisonDimension(dimensionName, isBeingCompared)}
          on:column-click={(e) => {
            if (e.detail.startsWith("measure")) {
              handleMeasureColumnHeaderClick(e.detail);
            } else {
              sortByDimensionValue();
            }
          }}
          {dimensionName}
          {columns}
          isBeingCompared={$isBeingCompared(dimensionName)}
          excludeMode={$isFilterExcludeMode(dimensionName)}
          selectedValues={$selectedDimensionValueNames}
          rows={tableRows}
          sortedColumn={columns.find((col) => col.sorted)?.name ??
            dimensionName}
          sortedAscending={$sortedAscending}
        /> -->
        <VirtualTable
          stickyBorders
          {columns}
          rows={tableRows}
          headerHeight={44}
          columnAccessor="name"
          PinnedCell={DimensionRowHeader}
          Cell={DimensionTableCell}
          HeaderCell={DimensionTableHeaderCell}
          pinnedColumns={new Map([[0, 0]])}
          valueAccessor={(name) => `__formatted_${name}`}
          sortedColumn={columns.find((col) => col.sorted)?.name ??
            dimensionName}
          sortedAscending={$sortedAscending}
          on:select-item={onSelectItem}
          selectedIndexes={$selectedDimensionValueNames.map((name) => {
            const index = tableRows.findIndex(
              (row) => row[dimensionColumnName] === name,
            );
            return index;
          })}
          on:column-click={(e) => {
            if (e.detail.startsWith("measure")) {
              handleMeasureColumnHeaderClick(e.detail);
            } else {
              sortByDimensionValue();
            }
          }}
        />
      </div>
    {/if}
  </div>
{/if}

<svelte:window on:keydown={handleKeyDown} />
