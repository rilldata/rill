<script lang="ts">
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import TDDHeader from "./TDDHeader.svelte";
  import TDDTable from "./TDDTable.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import {
    chartInteractionColumn,
    tableInteractionStore,
    useTimeDimensionDataStore,
  } from "./time-dimension-data-store";
  import {
    SortDirection,
    SortType,
  } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
  import type { TDDComparison, TableData } from "./types";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { toFormat } from "@rilldata/web-common/lib/time/timezone";

  export let metricViewName;

  const queryClient = useQueryClient();

  const timeDimensionDataStore = useTimeDimensionDataStore(getStateManagers());
  const timeControlStore = useTimeControlStore(getStateManagers());

  $: metaQuery = useMetaQuery(getStateManagers());
  $: dashboardStore = useDashboardStore(metricViewName);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension ?? "";

  $: timeGrain = $timeControlStore.selectedTimeRange?.interval;

  // Get labels for table headers
  $: measureLabel =
    $metaQuery?.data?.measures?.find(
      (m) => m.name === $dashboardStore?.expandedMeasureName,
    )?.label ?? "";

  let dimensionLabel = "";
  $: if ($timeDimensionDataStore?.comparing === "dimension") {
    dimensionLabel =
      $metaQuery?.data?.dimensions?.find((d) => d.name === dimensionName)
        ?.label ?? "";
  } else if ($timeDimensionDataStore?.comparing === "time") {
    dimensionLabel = "Time";
  } else if ($timeDimensionDataStore?.comparing === "none") {
    dimensionLabel = "No Comparison";
  }

  // Create a copy of the data to avoid flashing of table in transient states
  let timeDimensionDataCopy: TableData;
  let comparisonCopy: TDDComparison;
  $: if (
    $timeDimensionDataStore?.data &&
    $timeDimensionDataStore?.data?.columnHeaderData
  ) {
    comparisonCopy = $timeDimensionDataStore?.comparing;
    timeDimensionDataCopy = $timeDimensionDataStore.data;
  }
  $: formattedData = timeDimensionDataCopy;

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: rowHeaderLabels =
    formattedData?.rowHeaderData?.slice(1)?.map((row) => row[0]?.value) ?? [];

  $: areAllTableRowsSelected = rowHeaderLabels?.every((val) =>
    formattedData?.selectedValues?.includes(val),
  );

  $: columnHeaders = formattedData?.columnHeaderData?.flat();

  $: zone = $dashboardStore?.selectedTimezone ?? "UTC";

  // Create a time formatter for the column headers
  $: timeFormatter = function (d: Date): string {
    let format = timeGrain ? TIME_GRAIN[timeGrain].luxonFormat : "LLL dd";
    return toFormat(d, zone, format);
  };

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
        rowHeaderLabels,
      );
      notifications.send({
        message: `Removed ${rowHeaderLabels.length} items from filter`,
      });
      return;
    } else {
      const newValuesSelected = metricsExplorerStore.selectItemsInFilter(
        metricViewName,
        dimensionName,
        rowHeaderLabels,
      );

      notifications.send({
        message: `Added ${newValuesSelected} items to filter`,
      });
    }
  }

  function togglePin() {
    cancelDashboardQueries(queryClient, metricViewName);

    const pinIndex = $dashboardStore?.pinIndex;
    let newPinIndex = -1;

    // Pin if some selected items are not pinned yet
    if (pinIndex > -1 && pinIndex < formattedData?.selectedValues?.length - 1) {
      newPinIndex = formattedData?.selectedValues?.length - 1;
    }
    // Pin if no items are pinned yet
    else if (pinIndex === -1) {
      newPinIndex = formattedData?.selectedValues?.length - 1;
    }
    metricsExplorerStore.setPinIndex(metricViewName, newPinIndex);
  }

  function handleKeyDown(e) {
    if (comparisonCopy !== "dimension") return;
    // Select all items on Meta+A
    if ((e.ctrlKey || e.metaKey) && e.key === "a") {
      if (e.target.tagName === "INPUT") return;
      e.preventDefault();
      if (areAllTableRowsSelected) return;
      toggleAllSearchItems();
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
    sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
    sortType={$dashboardStore.dashboardSortType}
    comparing={comparisonCopy}
    {timeFormatter}
    tableData={formattedData}
    highlightedCol={$chartInteractionColumn?.hover}
    pinIndex={$dashboardStore?.pinIndex}
    scrubPos={{
      start: $chartInteractionColumn?.scrubStart,
      end: $chartInteractionColumn?.scrubEnd,
    }}
    on:toggle-pin={togglePin}
    on:toggle-filter={toggleFilter}
    on:toggle-sort={(e) => {
      cancelDashboardQueries(queryClient, metricViewName);
      metricsExplorerStore.toggleSort(
        metricViewName,
        e.detail === "dimension" ? SortType.DIMENSION : SortType.VALUE,
      );
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

<svelte:window on:keydown={handleKeyDown} />
