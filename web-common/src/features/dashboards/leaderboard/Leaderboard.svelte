<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
  import type {
    MetricsViewSpecDimensionV2,
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
  import {
    additionalMeasures,
    getFiltersForOtherDimensions,
  } from "../selectors";
  import {
    createAndExpression,
    createOrExpression,
    isExpressionUnsupported,
    sanitiseExpression,
  } from "../stores/filter-utils";
  import type { DimensionThresholdFilter } from "../stores/metrics-explorer-entity";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardRow from "./LeaderboardRow.svelte";
  import {
    cleanUpComparisonValue,
    compareLeaderboardValues,
    getSort,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import { valueColumn, COMPARISON_COLUMN_WIDTH } from "./leaderboard-widths";
  import DelayedLoadingRows from "./DelayedLoadingRows.svelte";
  import type { selectedDimensionValuesV2 } from "../state-managers/selectors/dimension-filters";

  const slice = 7;
  const gutterWidth = 24;
  const queryLimit = 8;
  const maxValuesToShow = 15;

  export let dimension: MetricsViewSpecDimensionV2;
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let selectedValues: ReturnType<typeof selectedDimensionValuesV2>;
  export let instanceId: string;
  export let whereFilter: V1Expression;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let activeMeasureName: string;
  export let activeMeasureNames: string[];
  export let metricsViewName: string;
  export let sortType: SortType;
  export let sortBy: string | null;
  export let tableWidth: number;
  export let sortedAscending: boolean;
  export let isValidPercentOfTotal: boolean;
  export let timeControlsReady: boolean;
  export let dimensionColumnWidth: number;
  export let isSummableMeasure: boolean;
  export let filterExcludeMode: boolean;
  export let isBeingCompared: boolean;
  export let parentElement: HTMLElement;
  export let suppressTooltip = false;
  export let leaderboardMeasureCountFeatureFlag: boolean;
  export let measureLabel: (measureName: string) => string;
  export let formatters: Record<
    string,
    (value: number | string | null | undefined) => string | null | undefined
  >;
  export let toggleDimensionValueSelection: (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean | undefined,
    isExclusiveFilter?: boolean | undefined,
  ) => void;
  export let setPrimaryDimension: (dimensionName: string) => void;
  export let toggleSort: (sortType: DashboardState_LeaderboardSortType) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void;

  const observer = new IntersectionObserver(
    ([entry]) => {
      visible = entry.isIntersecting;
    },
    {
      root: parentElement,
      rootMargin: "120px",
      threshold: 0,
    },
  );

  onMount(() => {
    observer.observe(container);
  });

  let container: HTMLElement;
  let visible = false;
  let hovered: boolean;

  $: ({
    name: dimensionName = "",
    description = "",
    displayName = "",
    uri,
  } = dimension);

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
    ...(leaderboardMeasureCountFeatureFlag
      ? activeMeasureNames.map(
          (name) =>
            ({
              name,
            }) as V1MetricsViewAggregationMeasure,
        )
      : additionalMeasures(activeMeasureName, dimensionThresholdFilters).map(
          (n) =>
            ({
              name: n,
            }) as V1MetricsViewAggregationMeasure,
        )),

    // Add comparison measures if there's a comparison time range
    ...(comparisonTimeRange
      ? activeMeasureNames.flatMap((name) => getComparisonRequestMeasures(name))
      : []),

    // Add URI measure if URI is present
    ...(uri ? [getURIRequestMeasure(dimensionName)] : []),
  ];

  $: sort = getSort(
    sortedAscending,
    sortType,
    activeMeasureName,
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
      ...(leaderboardMeasureCountFeatureFlag
        ? {
            measures: activeMeasureNames.map((name) => ({ name })),
          }
        : {
            measures: [{ name: activeMeasureName }],
          }),
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

  $: ({ data: sortedData, isFetching, isLoading } = $sortedQuery);
  $: ({ data: totalsData } = $totalsQuery);

  $: leaderboardTotals = totalsData?.data?.[0]
    ? Object.fromEntries(
        activeMeasureNames.map((name) => [
          name,
          (totalsData?.data?.[0]?.[name] as number) ?? null,
        ]),
      )
    : {};

  $: ({ aboveTheFold, belowTheFoldValues, noAvailableValues, showExpandTable } =
    prepareLeaderboardItemData(
      sortedData?.data,
      dimensionName,
      activeMeasureNames,
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
    : belowTheFoldValues.map((value) => ({
        [dimensionName]: value,
        [activeMeasureName]: null,
      }));

  $: belowTheFoldRows = belowTheFoldData.map((item) =>
    cleanUpComparisonValue(
      item,
      dimensionName,
      activeMeasureNames,
      leaderboardTotals,
      $selectedValues?.data?.findIndex((value) =>
        compareLeaderboardValues(value, item[dimensionName]),
      ) ?? -1,
    ),
  );

  $: isTimeComparisonActive = !!comparisonTimeRange;

  $: columnCount =
    1 + // Base column (dimension)
    activeMeasureNames.length + // Value column for each measure
    (isTimeComparisonActive
      ? activeMeasureNames.length * // For each measure
        ((isValidPercentOfTotal ? 1 : 0) + // Percent of total column
          (isTimeComparisonActive ? 2 : 0)) // Delta absolute and delta percent columns
      : 0);
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
      {#each activeMeasureNames as _, index (index)}
        <col data-measure-column style:width="{$valueColumn}px" />
        {#if isValidPercentOfTotal}
          <col
            data-percent-of-total-column
            style:width="{COMPARISON_COLUMN_WIDTH}px"
          />
        {/if}
        {#if isTimeComparisonActive}
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
      {activeMeasureNames}
      {toggleSort}
      {setPrimaryDimension}
      {toggleComparisonDimension}
      {sortBy}
      {measureLabel}
      {leaderboardMeasureCountFeatureFlag}
    />

    <tbody>
      <DelayedLoadingRows
        {isLoading}
        {isFetching}
        rowCount={aboveTheFold.length}
        columnCount={columnCount + 1}
      >
        {#each aboveTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardRow
            {suppressTooltip}
            {tableWidth}
            {dimensionColumnWidth}
            {isSummableMeasure}
            {isBeingCompared}
            {filterExcludeMode}
            {atLeastOneActive}
            {dimensionName}
            {itemData}
            {isValidPercentOfTotal}
            {isTimeComparisonActive}
            {activeMeasureNames}
            {toggleDimensionValueSelection}
            {formatters}
          />
        {/each}
      </DelayedLoadingRows>

      {#each belowTheFoldRows as itemData, i (itemData.dimensionValue)}
        <LeaderboardRow
          {suppressTooltip}
          {itemData}
          {dimensionColumnWidth}
          {isSummableMeasure}
          {tableWidth}
          {dimensionName}
          {isBeingCompared}
          {filterExcludeMode}
          {atLeastOneActive}
          {isValidPercentOfTotal}
          {isTimeComparisonActive}
          {activeMeasureNames}
          borderTop={i === 0}
          borderBottom={i === belowTheFoldRows.length - 1}
          {toggleDimensionValueSelection}
          {formatters}
        />
      {/each}
    </tbody>
  </table>

  {#if showExpandTable}
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
  {:else if noAvailableValues && !isLoading}
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
