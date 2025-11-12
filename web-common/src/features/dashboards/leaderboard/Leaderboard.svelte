<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
    V1Expression,
    V1MetricsViewAggregationMeasure,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import {
    createQueryServiceMetricsViewAggregation,
    V1Operation,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import {
    getComparisonRequestMeasures,
    getURIRequestMeasure,
  } from "../dashboard-utils";
  import { mergeDimensionAndMeasureFilters } from "../filters/measure-filters/measure-filter-utils";
  import { SortType } from "../proto-state/derived-types";
  import { getFiltersForOtherDimensions } from "../selectors";
  import {
    createAndExpression,
    createOrExpression,
    isExpressionUnsupported,
    sanitiseExpression,
  } from "../stores/filter-utils";
  import type { DimensionThresholdFilter } from "web-common/src/features/dashboards/stores/explore-state";
  import DelayedLoadingRows from "./DelayedLoadingRows.svelte";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardRow from "./LeaderboardRow.svelte";
  import {
    cleanUpComparisonValue,
    compareLeaderboardValues,
    getLeaderboardMaxValues,
    getSort,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import { valueColumn, COMPARISON_COLUMN_WIDTH } from "./leaderboard-widths";
  import type { selectedDimensionValues } from "../state-managers/selectors/dimension-filters";
  import { getMeasuresForDimensionOrLeaderboardDisplay } from "../state-managers/selectors/dashboard-queries";

  const gutterWidth = 24;

  export let dimension: MetricsViewSpecDimension;
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let selectedValues: ReturnType<typeof selectedDimensionValues>;
  export let instanceId: string;
  export let whereFilter: V1Expression;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let leaderboardSortByMeasureName: string;
  export let leaderboardMeasures: MetricsViewSpecMeasure[];
  export let leaderboardShowContextForAllMeasures: boolean;
  export let metricsViewName: string;
  export let sortType: SortType;
  export let slice = 7;
  export let tableWidth: number;
  export let sortedAscending: boolean;
  export let timeControlsReady: boolean;
  export let dimensionColumnWidth: number;
  export let filterExcludeMode: boolean;
  export let isBeingCompared: boolean;
  export let parentElement: HTMLElement | undefined = undefined;
  export let allowExpandTable = true;
  export let allowDimensionComparison = true;
  export let visible = false;
  export let formatters: Record<
    string,
    (value: number | string | null | undefined) => string | null | undefined
  >;
  export let isValidPercentOfTotal: (measureName: string) => boolean;
  export let measureLabel: (measureName: string) => string;
  export let toggleDimensionValueSelection: (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean | undefined,
    isExclusiveFilter?: boolean | undefined,
  ) => void;
  export let setPrimaryDimension: (dimensionName: string) => void = () => {};
  export let toggleSort: (sortType: DashboardState_LeaderboardSortType) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void = () => {};

  onMount(() => {
    if (!parentElement) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          visible = true;
          observer.unobserve(container);
        }
      },
      {
        root: parentElement,
        rootMargin: "120px",
        threshold: 0,
      },
    );
    observer.observe(container);
  });

  let container: HTMLElement;

  let hovered: boolean;

  $: queryLimit = slice + 1;
  $: maxValuesToShow = slice * 2;

  $: ({
    name: dimensionName = "",
    description = "",
    displayName = "",
    uri,
  } = dimension);

  $: leaderboardMeasureNames = leaderboardMeasures.map(
    (measure) => measure.name!,
  );

  $: atLeastOneActive = Boolean($selectedValues.data?.length);

  $: isComplexFilter = isExpressionUnsupported(whereFilter);
  $: where = isComplexFilter
    ? whereFilter
    : sanitiseExpression(
        mergeDimensionAndMeasureFilters(
          getFiltersForOtherDimensions(whereFilter, dimensionName),
          dimensionThresholdFilters,
        ),
        undefined,
      );

  $: measures = [
    ...getMeasuresForDimensionOrLeaderboardDisplay(
      leaderboardShowContextForAllMeasures
        ? null
        : leaderboardSortByMeasureName,
      dimensionThresholdFilters,
      leaderboardMeasureNames,
    ).map((name) => ({ name }) as V1MetricsViewAggregationMeasure),

    // Add comparison measures if there's a comparison time range
    ...(comparisonTimeRange
      ? (leaderboardShowContextForAllMeasures
          ? leaderboardMeasureNames
          : [leaderboardSortByMeasureName]
        ).flatMap((name) => getComparisonRequestMeasures(name))
      : []),

    // Add URI measure if URI is present
    ...(uri ? [getURIRequestMeasure(dimensionName)] : []),
  ];

  $: sort = getSort(
    sortedAscending,
    sortType,
    leaderboardSortByMeasureName,
    dimensionName,
    !!comparisonTimeRange,
  );

  $: sortedQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      dimensions: [{ name: dimensionName }],
      measures,
      timeRange,
      comparisonTimeRange,
      sort,
      where,
      limit: queryLimit.toString(),
      offset: "0",
    },
    {
      query: {
        enabled: visible && timeControlsReady,
      },
    },
  );

  $: totalsQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: leaderboardMeasureNames.map((name) => ({ name })),
      where,
      timeStart: timeRange.start,
      timeEnd: timeRange.end,
    },
    {
      query: {
        enabled: timeControlsReady && visible,
      },
    },
  );

  $: ({ data: sortedData, isFetching, isLoading, isPending } = $sortedQuery);
  $: ({ data: totalsData } = $totalsQuery);

  $: leaderboardTotals = totalsData?.data?.[0]
    ? Object.fromEntries(
        leaderboardMeasureNames.map((name) => [
          name,
          (totalsData?.data?.[0]?.[name] as number) ?? null,
        ]),
      )
    : {};

  $: ({ aboveTheFold, belowTheFoldValues, noAvailableValues, showExpandTable } =
    prepareLeaderboardItemData(
      sortedData?.data,
      dimensionName,
      leaderboardMeasureNames,
      slice,
      $selectedValues?.data ?? [],
      leaderboardTotals,
    ));

  $: belowTheFoldDataLimit = maxValuesToShow - aboveTheFold.length;
  $: belowTheFoldDataQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      dimensions: [{ name: dimensionName }],
      where: sanitiseExpression(
        createAndExpression(
          [
            createOrExpression(
              belowTheFoldValues.map((dimensionValue) => ({
                cond: {
                  op: V1Operation.OPERATION_EQ,
                  exprs: [{ ident: dimensionName }, { val: dimensionValue }],
                },
              })),
            ),
          ].concat(where ?? []),
        ),
        undefined,
      ),
      sort,
      timeRange,
      comparisonTimeRange,
      measures,
      limit: belowTheFoldDataLimit.toString(),
    },
    {
      query: {
        enabled:
          !!belowTheFoldValues.length &&
          timeControlsReady &&
          visible &&
          belowTheFoldDataLimit > 0,
      },
    },
  );

  $: ({ data } = $belowTheFoldDataQuery);

  $: belowTheFoldData = data?.data?.length
    ? data?.data
    : belowTheFoldValues
        .map((value) => ({
          [dimensionName]: value,
          [leaderboardSortByMeasureName]: null,
        }))
        .slice(0, belowTheFoldDataLimit);

  $: belowTheFoldRows = belowTheFoldData.map((item) =>
    cleanUpComparisonValue(
      item,
      dimensionName,
      leaderboardMeasureNames,
      leaderboardTotals,
      $selectedValues?.data?.findIndex((value) =>
        compareLeaderboardValues(value, item[dimensionName]),
      ) ?? -1,
    ),
  );

  $: isTimeComparisonActive = !!comparisonTimeRange;

  $: columnCount =
    1 + // Base column (dimension)
    leaderboardMeasureNames.length + // Value column for each measure
    (isTimeComparisonActive
      ? leaderboardMeasureNames.length * // For each measure
        ((isValidPercentOfTotal(leaderboardSortByMeasureName) ? 1 : 0) + // Percent of total column
          (isTimeComparisonActive ? 2 : 0)) // Delta absolute and delta percent columns
      : 0);

  // Calculate maximum values for relative magnitude bar sizing
  // This includes both above-the-fold and below-the-fold data for accurate scaling
  $: maxValues = getLeaderboardMaxValues(
    [...aboveTheFold, ...belowTheFoldRows],
    leaderboardMeasures,
  );

  function shouldShowContextColumns(measureName: string): boolean {
    return (
      leaderboardShowContextForAllMeasures ||
      measureName === leaderboardSortByMeasureName
    );
  }
</script>

<div
  class="flex flex-col"
  aria-label="{dimensionName} leaderboard"
  role="table"
  bind:this={container}
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
>
  <table style:width="{tableWidth + gutterWidth}px">
    <colgroup>
      <col data-gutter-column style:width="{gutterWidth}px" />
      <col data-dimension-column style:width="{dimensionColumnWidth}px" />
      {#each leaderboardMeasureNames as measureName, index (index)}
        <col data-measure-column style:width="{$valueColumn}px" />
        {#if isValidPercentOfTotal(measureName) && shouldShowContextColumns(measureName)}
          <col
            data-percent-of-total-column
            style:width="{COMPARISON_COLUMN_WIDTH}px"
          />
        {/if}
        {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
          <col
            data-absolute-change-column
            style:width="{COMPARISON_COLUMN_WIDTH}px"
          />
          <col
            data-percent-change-column
            style:width="{COMPARISON_COLUMN_WIDTH}px"
          />
        {/if}
      {/each}
    </colgroup>

    <LeaderboardHeader
      {allowDimensionComparison}
      {allowExpandTable}
      {hovered}
      displayName={displayName || dimensionName}
      dimensionDescription={description}
      {dimensionName}
      {isBeingCompared}
      isFetching={isLoading}
      {sortType}
      {isValidPercentOfTotal}
      {isTimeComparisonActive}
      {sortedAscending}
      {leaderboardMeasureNames}
      {leaderboardShowContextForAllMeasures}
      {toggleSort}
      {setPrimaryDimension}
      {toggleComparisonDimension}
      {leaderboardSortByMeasureName}
      {measureLabel}
    />

    <tbody>
      <DelayedLoadingRows
        {isLoading}
        {isPending}
        {isFetching}
        rowCount={aboveTheFold.length}
        columnCount={columnCount + 1}
      >
        {#each aboveTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardRow
            {isBeingCompared}
            {filterExcludeMode}
            {atLeastOneActive}
            {dimensionName}
            {itemData}
            {isValidPercentOfTotal}
            {leaderboardShowContextForAllMeasures}
            {isTimeComparisonActive}
            {leaderboardMeasureNames}
            {toggleDimensionValueSelection}
            {leaderboardSortByMeasureName}
            {formatters}
            {dimensionColumnWidth}
            {maxValues}
          />
        {/each}
      </DelayedLoadingRows>

      {#each belowTheFoldRows as itemData, i (itemData.dimensionValue)}
        <LeaderboardRow
          {itemData}
          {dimensionName}
          {isBeingCompared}
          {filterExcludeMode}
          {atLeastOneActive}
          {isValidPercentOfTotal}
          {leaderboardShowContextForAllMeasures}
          {isTimeComparisonActive}
          {leaderboardMeasureNames}
          borderTop={i === 0}
          borderBottom={i === belowTheFoldRows.length - 1}
          {toggleDimensionValueSelection}
          {leaderboardSortByMeasureName}
          {formatters}
          {dimensionColumnWidth}
          {maxValues}
        />
      {/each}
    </tbody>
  </table>

  {#if allowExpandTable && showExpandTable}
    <Tooltip location="right">
      <button
        class="transition-color ui-copy-muted table-message"
        on:click={() => setPrimaryDimension(dimensionName)}
      >
        <div class="pl-8">(Expand Table)</div>
      </button>
      <TooltipContent slot="tooltip-content">
        Expand dimension to see more values
      </TooltipContent>
    </Tooltip>
  {:else if noAvailableValues}
    <div class="table-message ui-copy-muted">
      <div class="pl-8">(No available values)</div>
    </div>
  {/if}
</div>

<style lang="postcss">
  table {
    @apply p-0 m-0 border-spacing-0 border-collapse w-fit;
    @apply font-normal cursor-pointer select-none;
    @apply table-fixed;
  }

  .table-message {
    @apply h-[22px] flex items-center w-fit;
  }
</style>
