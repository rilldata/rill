<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    getComparisonRequestMeasures,
    getURIRequestMeasure,
  } from "../dashboard-utils";
  import { getSort } from "../leaderboard/leaderboard-utils";
  import { getFiltersForOtherDimensions } from "../selectors";
  import { getMeasuresForDimensionOrLeaderboardDisplay } from "../state-managers/selectors/dashboard-queries";
  import { dimensionSearchText } from "../stores/dashboard-stores";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import { getDimensionFilterWithSearch } from "./dimension-table-utils";
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";

  const queryLimit = 250;

  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let whereFilter: V1Expression | undefined;
  export let metricsViewName: string;
  export let visibleMeasureNames: string[];
  export let timeControlsReady: boolean;
  export let dimension: MetricsViewSpecDimension;
  export let hideStartPivotButton = false;

  const {
    selectors: {
      dimensionTable: { virtualizedTableColumns, prepareDimTableRows },
      sorting: { sortedAscending, sortType },
      leaderboard: {
        leaderboardShowContextForAllMeasures,
        leaderboardSortByMeasureName,
      },
    },
    dashboardStore,
    validSpecStore,
    expressionFilterManager,
  } = getStateManagers();

  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};

  const client = useRuntimeClient();

  $: ({ name: dimensionName = "" } = dimension);

  $: selectedValues = selectedDimensionValues(
    client,
    [metricsViewName],
    whereFilter,
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
      whereFilter,
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

  // Request the URI measure so the dimension values can be rendered as links.
  // Added after filtering (it is not a real spec measure) and only on the
  // grouped query, since the URI resolves per dimension value.
  $: sortedMeasures = dimension.uri
    ? [...filteredMeasures, getURIRequestMeasure(dimensionName)]
    : filteredMeasures;

  $: totalsQuery = createQueryServiceMetricsViewAggregation(
    client,
    {
      metricsView: metricsViewName,
      measures: filteredMeasures.filter(
        (m) => !m.comparisonValue && !m.comparisonDelta && !m.comparisonRatio,
      ),
      where: sanitiseExpression(
        getFiltersForOtherDimensions(whereFilter, dimensionName),
      ),
      timeRange,
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

  $: sortedQuery = createQueryServiceMetricsViewAggregation(
    client,
    {
      metricsView: metricsViewName,
      dimensions: [{ name: dimensionName }],
      measures: sortedMeasures,
      timeRange,
      comparisonTimeRange,
      sort,
      where: sanitiseExpression(filterSet),
      limit: queryLimit.toString(),
      offset: "0",
    },
    {
      query: {
        enabled: timeControlsReady,
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

  function onSelectItem(data: { index: number; meta: boolean }) {
    const label = tableRows[data.index][dimensionName] as string;
    expressionFilterManager.dimensionFilterAction(
      dimensionName,
      (dimensionManager) => dimensionManager.toggleValue(label, data.meta),
    );
  }

  function toggleAllSearchItems() {
    const labels = tableRows.map((row) => row[dimensionName] as string);

    if (areAllTableRowsSelected) {
      expressionFilterManager.dimensionFilterAction(
        dimensionName,
        (dimensionManager) => dimensionManager.removeSelectedValues(labels),
      );

      eventBus.emit("notification", {
        message: m.dashboard_removed_items_filter({
          count: labels.length.toString(),
        }),
      });
      return;
    } else {
      const newValuesSelected = expressionFilterManager.dimensionFilterAction(
        dimensionName,
        (dimensionManager) => dimensionManager.appendSelectedValues(labels),
      );
      if (newValuesSelected?.length) {
        eventBus.emit("notification", {
          message: m.dashboard_added_items_filter({
            count: newValuesSelected?.length.toString(),
          }),
        });
      }
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
    aria-label={m.dashboard_dimension_display_aria()}
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
          {onSelectItem}
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

<svelte:window onkeydown={handleKeyDown} />
