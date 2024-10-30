<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    createQueryServiceMetricsViewAggregation,
    type MetricsViewSpecDimensionV2,
    type V1Expression,
    type V1MetricsViewAggregationMeasure,
    type V1MetricsViewSpec,
    type V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { getDimensionFilterWithSearch } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { sanitiseExpression } from "../stores/filter-utils";
  import { mergeDimensionAndMeasureFilter } from "../filters/measure-filters/measure-filter-utils";
  import { getFiltersForOtherDimensions } from "../selectors";
  import type { DimensionThresholdFilter } from "../stores/metrics-explorer-entity";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getApiSortName } from "../leaderboard/leaderboard-utils";
  import { getMeasuresForDimensionTable } from "../state-managers/selectors/dashboard-queries";
  import { getComparisonRequestMeasures } from "../dashboard-utils";
  import { SortType } from "../proto-state/derived-types";
  import { dimensionSearchText } from "../stores/dashboard-stores";

  const queryLimit = 250;

  export let activeMeasureName: string;
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let whereFilter: V1Expression;
  export let metricsViewName: string;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let metricsView: V1MetricsViewSpec;
  export let visibleMeasureNames: string[];
  export let timeControlsReady: boolean;
  export let dimension: MetricsViewSpecDimensionV2;

  const stateManagers = getStateManagers();
  const {
    selectors: {
      dimensionFilters: { unselectedDimensionValues },
      dimensionTable: {
        virtualizedTableColumns,
        selectedDimensionValueNames,
        prepareDimTableRows,
      },
      sorting: { sortedAscending, sortType },
    },
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        selectItemsInFilter,
        deselectItemsInFilter,
      },
    },
  } = stateManagers;

  $: ({ name: dimensionName = "" } = dimension);

  $: ({ instanceId } = $runtime);

  $: filterSet = getDimensionFilterWithSearch(
    whereFilter,
    $dimensionSearchText,
    dimensionName,
  );

  $: filters = getFiltersForOtherDimensions(whereFilter, dimensionName);

  $: where = sanitiseExpression(
    mergeDimensionAndMeasureFilter(filters, dimensionThresholdFilters),
    undefined,
  );

  $: totalsQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: activeMeasureName }],
      where,
      timeStart: timeRange.start,
      timeEnd: timeRange.end,
    },
    {
      query: {
        enabled: timeControlsReady,
      },
    },
  );

  $: unfilteredTotal = $totalsQuery?.data?.data?.[0]?.[activeMeasureName] ?? 0;

  $: columns = $virtualizedTableColumns($totalsQuery);

  $: measures = getMeasuresForDimensionTable(
    activeMeasureName,
    dimensionThresholdFilters,
    metricsView,
    visibleMeasureNames,
  )
    .map(
      (n) =>
        ({
          name: n,
        }) as V1MetricsViewAggregationMeasure,
    )
    .concat(
      ...(comparisonTimeRange
        ? getComparisonRequestMeasures(activeMeasureName)
        : []),
    );

  $: sort = [
    {
      desc: !$sortedAscending,
      name:
        $sortType === SortType.DIMENSION || !activeMeasureName
          ? dimensionName
          : getApiSortName(activeMeasureName, $sortType),
    },
  ];

  $: sortedQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      dimensions: [{ name: dimensionName }],
      measures,
      timeRange,
      comparisonTimeRange,
      sort,
      where: sanitiseExpression(
        mergeDimensionAndMeasureFilter(filterSet, dimensionThresholdFilters),
        undefined,
      ),
      limit: queryLimit.toString(),
      offset: "0",
    },
    {
      query: {
        enabled: timeControlsReady && !!filterSet,
      },
    },
  );

  $: tableRows = $prepareDimTableRows($sortedQuery, unfilteredTotal);

  $: areAllTableRowsSelected = tableRows.every((row) =>
    $selectedDimensionValueNames.includes(row[dimensionName] as string),
  );

  function onSelectItem(event) {
    const label = tableRows[event.detail.index][dimensionName] as string;
    toggleDimensionValueSelection(
      dimensionName,
      label,
      false,
      event.detail.meta,
    );
  }

  function toggleAllSearchItems() {
    const labels = tableRows.map((row) => row[dimensionName] as string);

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

  // Select all items on Meta+A
  function handleKeyDown(
    e: KeyboardEvent & {
      currentTarget: EventTarget & Window;
    },
  ) {
    if ((e.ctrlKey || e.metaKey) && e.key === "a") {
      if (e.target instanceof HTMLElement && e.target.tagName === "INPUT")
        return;
      e.preventDefault();
      if (areAllTableRowsSelected) return;
      toggleAllSearchItems();
    }
  }
</script>

{#if $sortedQuery}
  <div class="h-full flex flex-col w-full" style:min-width="365px">
    <div class="flex-none" style:height="50px">
      <DimensionHeader
        {dimensionName}
        {areAllTableRowsSelected}
        isRowsEmpty={!tableRows.length}
        isFetching={$sortedQuery?.isFetching}
        bind:searchText={$dimensionSearchText}
        onToggleSearchItems={toggleAllSearchItems}
      />
    </div>

    {#if tableRows && columns.length && dimensionName}
      <div class="grow" style="overflow-y: hidden;">
        <DimensionTable
          on:select-item={(event) => onSelectItem(event)}
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
