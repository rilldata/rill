<script lang="ts">
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import {
    SortDirection,
    SortType,
  } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { timeFormat } from "d3-time-format";
  import { onDestroy } from "svelte";
  import TDDHeader from "./TDDHeader.svelte";
  import TDDTable from "./TDDTable.svelte";
  import { useTimeDimensionDataStore } from "./time-dimension-data-store";
  import {
    chartHoverStore,
    hoverIndex,
  } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/hover-index";
  import type { TDDComparison, TableData } from "./types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  interface Props {
    exploreName: string;
    expandedMeasureName: string;
    hideStartPivotButton?: boolean;
  }

  let {
    exploreName,
    expandedMeasureName,
    hideStartPivotButton = false,
  }: Props = $props();

  const isEmbedded = $derived(EmbedStore.isEmbedded());

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      dimensions: { allDimensions },
      measures: { allMeasures },
    },
    actions: {
      sorting: { toggleSort },
    },
    expressionFilterManager,
  } = getStateManagers();

  const timeDimensionDataStore = useTimeDimensionDataStore(stateManagers);
  const timeControlStore = useTimeControlStore(stateManagers);

  const dimensionName = $derived(
    $dashboardStore?.selectedComparisonDimension ?? "",
  );
  const comparing = $derived($timeDimensionDataStore?.comparing);

  const pinIndex = $derived($dashboardStore?.tdd.pinIndex);

  const timeGrain = $derived($timeControlStore.selectedTimeRange?.interval);

  const measure = $derived(
    $allMeasures.find((m) => m.name === expandedMeasureName),
  );

  const measureLabel = $derived(measure?.displayName ?? "");

  const dimensionLabel = $derived.by(() => {
    if (comparing === "dimension") {
      return (
        $allDimensions.find((d) => d.name === dimensionName)?.displayName ?? ""
      );
    } else if (comparing === "time") {
      return m.dashboard_tdd_time();
    } else if (comparing === "none") {
      return m.dashboard_tdd_no_comparison();
    }
    return "";
  });

  // Create a copy of the data to avoid flashing of table in transient states
  let timeDimensionDataCopy = $state<TableData>();
  let comparisonCopy = $state<TDDComparison | undefined>();
  $effect(() => {
    if (
      $timeDimensionDataStore?.data &&
      $timeDimensionDataStore?.data?.columnHeaderData
    ) {
      comparisonCopy = comparing;
      timeDimensionDataCopy = $timeDimensionDataStore.data;
    }
  });
  const formattedData = $derived(timeDimensionDataCopy);
  const excludeMode = $derived(
    expressionFilterManager.filterManagers.dimensions.find(
      (dfm) => dfm.name === dimensionName,
    )?.exclude ?? false,
  );

  const rowHeaderLabels = $derived(
    formattedData?.rowHeaderData?.slice(1)?.map((row) => row[0]?.value) ?? [],
  );

  const areAllTableRowsSelected = $derived(
    rowHeaderLabels?.every(
      (val) =>
        val !== undefined && formattedData?.selectedValues?.includes(val),
    ),
  );

  const columnHeaders = $derived(formattedData?.columnHeaderData?.flat());

  const highlightedColStart = $derived($hoverIndex?.start);
  const highlightedColEnd = $derived($hoverIndex?.end);

  // Create a time formatter for the column headers
  const timeFormatter = $derived(
    timeFormat(timeGrain ? TIME_GRAIN[timeGrain].d3format : "%H:%M") as (
      d: Date,
    ) => string,
  );

  function highlightCell(x: number | undefined, y: number | undefined) {
    if (x === undefined || y === undefined) {
      hoverIndex.clear("table");
      chartHoverStore.set({
        dimensionValue: undefined,
        time: undefined,
      });
      return;
    }
    const dimensionValue = formattedData?.rowHeaderData[y]?.[0]?.value;
    let time: Date | undefined = undefined;
    const colHeader = columnHeaders?.[x]?.value;
    if (colHeader) {
      time = new Date(colHeader);
    }

    hoverIndex.set(x, "table");
    chartHoverStore.set({
      dimensionValue,
      time: time,
    });
  }

  const debounceHighlightCell = debounce(highlightCell, 50);

  function toggleFilter(label: string) {
    expressionFilterManager.dimensionFilterAction(
      dimensionName,
      (dimensionManager) => dimensionManager.toggleValue(label, false),
    );
  }

  function toggleAllSearchItems() {
    const headerHasUndefined = rowHeaderLabels.some(
      (label) => label === undefined,
    );

    if (headerHasUndefined) return;

    if (areAllTableRowsSelected) {
      expressionFilterManager.dimensionFilterAction(
        dimensionName,
        (dimensionManager) =>
          dimensionManager.removeSelectedValues(rowHeaderLabels as string[]),
      );

      eventBus.emit("notification", {
        message: m.dashboard_removed_items_filter({
          count: rowHeaderLabels.length.toString(),
        }),
      });
      return;
    } else {
      const newValuesSelected = expressionFilterManager.dimensionFilterAction(
        dimensionName,
        (dimensionManager) =>
          dimensionManager.appendSelectedValues(rowHeaderLabels as string[]),
      );
      eventBus.emit("notification", {
        message: m.dashboard_added_items_filter({
          count: newValuesSelected.length.toString(),
        }),
      });
    }
  }

  function togglePin() {
    let newPinIndex = -1;
    const selectedCount = formattedData?.selectedValues?.length ?? 0;

    // Pin if some selected items are not pinned yet
    if (pinIndex > -1 && pinIndex < selectedCount - 1) {
      newPinIndex = selectedCount - 1;
    }
    // Pin if no items are pinned yet
    else if (pinIndex === -1) {
      newPinIndex = selectedCount - 1;
    }
    metricsExplorerStore.setPinIndex(exploreName, newPinIndex);
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (comparisonCopy !== "dimension") return;
    // Select all items on Meta+A
    if ((e.ctrlKey || e.metaKey) && e.key === "a") {
      if ((e.target as HTMLElement)?.tagName === "INPUT") return;
      e.preventDefault();
      if (areAllTableRowsSelected) return;
      toggleAllSearchItems();
    }
  }

  onDestroy(() => {
    hoverIndex.clear("table");
    chartHoverStore.set({
      dimensionValue: undefined,
      time: undefined,
    });
  });
</script>

<svelte:window onkeydown={handleKeyDown} />

<div
  class="h-full w-full flex flex-col bg-surface-base"
  aria-label="{expandedMeasureName} Time Dimension Display"
>
  <TDDHeader
    {areAllTableRowsSelected}
    comparing={comparisonCopy}
    {expandedMeasureName}
    {dimensionName}
    isFetching={!$timeDimensionDataStore?.data?.columnHeaderData}
    isRowsEmpty={!rowHeaderLabels.length}
    {exploreName}
    onToggleSearchItems={toggleAllSearchItems}
    {hideStartPivotButton}
  />

  {#if $timeDimensionDataStore?.isError}
    <div
      style:height="calc(100% - 200px)"
      class="w-full flex items-center justify-center text-sm"
    >
      <div class="text-center">
        <div class="text-red-600 mt-1 text-lg">
          {m.dashboard_tdd_error()}
        </div>
        {#if !isEmbedded}
          <div class="text-fg-secondary">
            {m.dashboard_tdd_contact_discord()}
            <a
              target="_blank"
              rel="noopener noreferrer"
              href="https://discord.gg/2ubRfjC7Rh">Discord</a
            >.
          </div>
        {/if}
      </div>
    </div>
  {:else if formattedData && comparisonCopy && measure}
    <TDDTable
      {measure}
      {excludeMode}
      {dimensionLabel}
      {measureLabel}
      sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
      sortType={$dashboardStore.dashboardSortType}
      comparing={comparisonCopy}
      {timeFormatter}
      tableData={formattedData}
      highlightedRow={undefined}
      {highlightedColStart}
      {highlightedColEnd}
      {pinIndex}
      onTogglePin={togglePin}
      onToggleFilter={toggleFilter}
      onToggleSort={(type) => {
        toggleSort(type === "dimension" ? SortType.DIMENSION : SortType.VALUE);
      }}
      onHighlight={debounceHighlightCell}
    />
  {/if}

  {#if comparing === "none"}
    <!-- Get height by subtracting table and header heights -->
    <div class="w-full" style:height="calc(100% - 200px)">
      <div class="flex flex-col items-center justify-center h-full text-sm">
        <Compare size="32px" />
        <div class="font-semibold text-fg-secondary mt-1">
          {m.dashboard_no_comparison_dimension()}
        </div>
        <div class="text-fg-secondary">
          {m.dashboard_select_comparison_hint()}
        </div>
      </div>
    </div>
  {:else if comparing === "dimension" && formattedData?.rowCount === 1}
    <div class="w-full h-full">
      <div class="flex flex-col items-center h-full text-sm">
        <div class="text-fg-secondary">{m.dashboard_no_search_results()}</div>
      </div>
    </div>
  {/if}
</div>
