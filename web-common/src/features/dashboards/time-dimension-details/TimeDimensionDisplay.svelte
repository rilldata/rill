<script lang="ts">
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import TDDHeader from "./TDDHeader.svelte";
  import TDDTable from "./TDDTable.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import {
    chartInteractionColumn,
    tableInteractionStore,
    useTimeDimensionDataStore,
  } from "./time-dimension-data-store";
  import {
    SortDirection,
    SortType,
  } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { createTimeFormat } from "@rilldata/web-common/components/data-graphic/utils";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
  import type { TableData } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let metricViewName;

  const queryClient = useQueryClient();

  const timeDimensionDataStore = useTimeDimensionDataStore(getStateManagers());

  $: metaQuery = useMetaQuery(getStateManagers());
  $: dashboardStore = useDashboardStore(metricViewName);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension ?? "";

  // Get labels for table headers
  $: measureLabel =
    $metaQuery?.data?.measures?.find(
      (m) => m.name === $dashboardStore?.expandedMeasureName
    )?.label ?? "";

  let dimensionLabel = "";
  $: if ($timeDimensionDataStore?.comparing === "dimension") {
    dimensionLabel =
      $metaQuery?.data?.dimensions?.find((d) => d.name === dimensionName)
        ?.label ?? "";
  } else if ($timeDimensionDataStore?.comparing === "time") {
    dimensionLabel = "Time";
  } else {
    dimensionLabel = "No Comparison";
  }

  // Create a copy of the data to avoid flashing of table in transient states
  let timeDimensionDataCopy: TableData;
  $: if (
    $timeDimensionDataStore?.data &&
    $timeDimensionDataStore?.data?.columnHeaderData
  ) {
    timeDimensionDataCopy = $timeDimensionDataStore.data;
  }
  $: formattedData = timeDimensionDataCopy;

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: rowHeaderLabels =
    formattedData?.rowHeaderData?.slice(1)?.map((row) => row[0]?.value) ?? [];

  $: areAllTableRowsSelected = rowHeaderLabels?.every((val) =>
    formattedData?.selectedValues?.includes(val)
  );

  $: columnHeaders = formattedData?.columnHeaderData?.flat();

  // Create a time formatter for the column headers
  $: timeFormatter = (columnHeaders?.length &&
    createTimeFormat(
      [
        new Date(columnHeaders[0]?.value),
        new Date(columnHeaders[columnHeaders?.length - 1]?.value),
      ],
      columnHeaders?.length
    )[0]) as (d: Date) => string;

  function highlightCell(e) {
    const { x, y } = e.detail;

    const dimensionValue = formattedData?.rowHeaderData[y]?.[0]?.value;
    let time: Date | undefined = undefined;
    if (columnHeaders?.[x]?.value) {
      time = new Date(columnHeaders?.[x]?.value);
    }

    tableInteractionStore.set({
      dimensionValue,
      time: time,
    });
  }

  function toggleFilter(e) {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(metricViewName, dimensionName, e.detail);
  }
  function toggleAllSearchItems() {
    cancelDashboardQueries(queryClient, metricViewName);
    if (areAllTableRowsSelected) {
      metricsExplorerStore.deselectItemsInFilter(
        metricViewName,
        dimensionName,
        rowHeaderLabels
      );
      return;
    } else {
      metricsExplorerStore.selectItemsInFilter(
        metricViewName,
        dimensionName,
        rowHeaderLabels
      );
    }
  }
</script>

<TDDHeader
  {dimensionName}
  {metricViewName}
  isFetching={!$timeDimensionDataStore?.data?.columnHeaderData}
  comparing={$timeDimensionDataStore?.comparing}
  {areAllTableRowsSelected}
  isRowsEmpty={!rowHeaderLabels.length}
  on:search={(e) => {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.setSearchText(metricViewName, e.detail);
  }}
  on:toggle-all-search-items={() => toggleAllSearchItems()}
/>

{#if formattedData}
  <TDDTable
    {excludeMode}
    {dimensionLabel}
    {measureLabel}
    scrubPos={{
      start: $chartInteractionColumn?.scrubStart,
      end: $chartInteractionColumn?.scrubEnd,
    }}
    sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
    comparing={$timeDimensionDataStore?.comparing}
    {timeFormatter}
    tableData={formattedData}
    highlightedCol={$chartInteractionColumn?.hover}
    on:toggle-filter={toggleFilter}
    on:toggle-sort={() => {
      cancelDashboardQueries(queryClient, metricViewName);
      metricsExplorerStore.toggleSort(metricViewName, SortType.VALUE);
    }}
    on:highlight={highlightCell}
  />
{/if}

{#if $timeDimensionDataStore?.comparing === "none"}
  <!-- Get height by subtracting table and header heights -->
  <div class="w-full" style:height="calc(100% - 200px)">
    <div class="flex flex-col items-center justify-center h-full text-sm">
      <Compare size="32px" />
      <div class="font-semibold text-gray-600 mt-1">
        No comparison dimension selected
      </div>
      <div class="text-gray-600">
        To see more values, select a comparison dimension above.
      </div>
    </div>
  </div>
{/if}
