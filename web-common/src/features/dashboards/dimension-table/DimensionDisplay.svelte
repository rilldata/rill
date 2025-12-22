<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { selectedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { filterOutSomeAdvancedAggregationMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    createQueryServiceMetricsViewAggregation,
    type MetricsViewSpecDimension,
    type V1Expression,
    type V1MetricsViewAggregationMeasure,
    type V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getComparisonRequestMeasures } from "../dashboard-utils";
  import { mergeDimensionAndMeasureFilters } from "../filters/measure-filters/measure-filter-utils";
  import { getSort } from "../leaderboard/leaderboard-utils";
  import { getFiltersForOtherDimensions } from "../selectors";
  import { getMeasuresForDimensionOrLeaderboardDisplay } from "../state-managers/selectors/dashboard-queries";
  import { dimensionSearchText } from "../stores/dashboard-stores";
  import { sanitiseExpression } from "../stores/filter-utils";
  import type { DimensionThresholdFilter } from "web-common/src/features/dashboards/stores/explore-state";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import { getDimensionFilterWithSearch } from "./dimension-table-utils";

  const queryLimit = 250;

  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let whereFilter: V1Expression;
  export let metricsViewName: string;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let visibleMeasureNames: string[];
  export let timeControlsReady: boolean;
  export let dimension: MetricsViewSpecDimension;
  export let hideStartPivotButton = false;

  const {
    selectors: {
      dimensionFilters: { unselectedDimensionValues },
      dimensionTable: { virtualizedTableColumns, prepareDimTableRows },
      sorting: { sortedAscending, sortType },
      leaderboard: {
        leaderboardShowContextForAllMeasures,
        leaderboardSortByMeasureName,
      },
    },
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        selectItemsInFilter,
        deselectItemsInFilter,
      },
    },
    dashboardStore,
    validSpecStore,
  } = getStateManagers();

  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};

  $: ({ name: dimensionName = "" } = dimension);

  $: ({ instanceId } = $runtime);

  $: selectedValues = selectedDimensionValues(
    $runtime.instanceId,
    [metricsViewName],
    $dashboardStore.whereFilter,
    dimensionName,
    timeRange.start,
    timeRange.end,
  );

  $: filterSet = getDimensionFilterWithSearch(
    whereFilter,
    $dimensionSearchText,
    dimensionName,
  );

  $: measures = [
    ...getMeasuresForDimensionOrLeaderboardDisplay(
      $leaderboardShowContextForAllMeasures
        ? null
        : $leaderboardSortByMeasureName,
      dimensionThresholdFilters,
      visibleMeasureNames,
    ).map((name) => ({ name }) as V1MetricsViewAggregationMeasure),

    // Add comparison measures if comparison time range exists
    ...(comparisonTimeRange
      ? ($leaderboardShowContextForAllMeasures
          ? visibleMeasureNames
          : [$leaderboardSortByMeasureName]
        ).flatMap((name) => getComparisonRequestMeasures(name))
      : []),
  ];
  $: filteredMeasures = filterOutSomeAdvancedAggregationMeasures(
    $dashboardStore,
    metricsViewSpec,
    measures,
    false,
  );

  $: totalsQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: filteredMeasures,
      where: sanitiseExpression(
        mergeDimensionAndMeasureFilters(
          getFiltersForOtherDimensions(whereFilter, dimensionName),
          dimensionThresholdFilters,
        ),
        undefined,
      ),
      timeRange,
      comparisonTimeRange,
    },
    {
      query: {
        enabled: timeControlsReady,
      },
    },
  );

  $: unfilteredTotal = $leaderboardShowContextForAllMeasures
    ? visibleMeasureNames.reduce(
        (acc, measureName) => {
          acc[measureName] =
            ($totalsQuery?.data?.data?.[0]?.[measureName] as number) ?? 0;
          return acc;
        },
        {} as { [key: string]: number },
      )
    : (($totalsQuery?.data?.data?.[0]?.[
        $leaderboardSortByMeasureName
      ] as number) ?? 0);

  $: sort = getSort(
    $sortedAscending,
    $sortType,
    $leaderboardSortByMeasureName,
    dimensionName,
    !!comparisonTimeRange,
  );

  $: where = sanitiseExpression(
    mergeDimensionAndMeasureFilters(filterSet, dimensionThresholdFilters),
    undefined,
  );

  $: sortedQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      dimensions: [{ name: dimensionName }],
      measures: filteredMeasures,
      timeRange,
      comparisonTimeRange,
      sort,
      where,
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
    $selectedValues.data?.includes(row[dimensionName] as string),
  );

  $: columns = $virtualizedTableColumns(
    tableRows,
    $leaderboardShowContextForAllMeasures ? visibleMeasureNames : undefined,
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
  <div
    class="h-full flex flex-col w-full"
    style:min-width="365px"
    aria-label="Dimension Display"
  >
    <DimensionHeader
      {dimensionName}
      {areAllTableRowsSelected}
      isRowsEmpty={!tableRows.length}
      {hideStartPivotButton}
      bind:searchText={$dimensionSearchText}
      onToggleSearchItems={toggleAllSearchItems}
    />

    {#if tableRows && columns.length && dimensionName}
      <div class="grow" style="overflow-y: hidden;">
        <DimensionTable
          on:select-item={(event) => onSelectItem(event)}
          isFetching={$sortedQuery?.isFetching}
          {dimensionName}
          {columns}
          {selectedValues}
          rows={tableRows}
        />
      </div>
    {/if}
  </div>
{/if}

<svelte:window on:keydown={handleKeyDown} />
