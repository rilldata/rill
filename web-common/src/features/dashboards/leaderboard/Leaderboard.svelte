<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
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
  import LoadingRows from "./LoadingRows.svelte";
  import {
    cleanUpComparisonValue,
    compareLeaderboardValues,
    getSort,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import {
    valueColumn,
    DEFAULT_CONTEXT_COLUMN_WIDTH,
  } from "./leaderboard-widths";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";

  const slice = 7;
  const gutterWidth = 24;
  const queryLimit = 8;
  const maxValuesToShow = 15;

  export let dimension: MetricsViewSpecDimensionV2;
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let selectedValues: CompoundQueryResult<string[]>;
  export let instanceId: string;
  export let whereFilter: V1Expression;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let activeMeasureName: string;
  export let activeMeasureNames: string[];
  export let metricsViewName: string;
  export let sortType: SortType;
  export let sortMeasure: string | null;
  export let tableWidth: number;
  export let sortedAscending: boolean;
  export let isValidPercentOfTotal: boolean;
  export let contextColumns: LeaderboardContextColumn[] = [];
  export let timeControlsReady: boolean;
  export let dimensionColumnWidth: number;
  export let isSummableMeasure: boolean;
  export let filterExcludeMode: boolean;
  export let isBeingCompared: boolean;
  export let parentElement: HTMLElement;
  export let suppressTooltip = false;
  export let measureLabel: (measureName: string) => string;
  export let toggleDimensionValueSelection: (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean | undefined,
    isExclusiveFilter?: boolean | undefined,
  ) => void;
  export let formatter:
    | ((_value: number | undefined) => undefined)
    | ((value: string | number) => string);
  export let setPrimaryDimension: (dimensionName: string) => void;
  export let toggleSort: (sortType: DashboardState_LeaderboardSortType) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void;

  $: console.log("Leaderboard activeMeasureNames: ", activeMeasureNames);

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
    // Get additional measures for each active measure
    ...activeMeasureNames
      .flatMap((name) => additionalMeasures(name, dimensionThresholdFilters))
      .map(
        (name) =>
          ({
            name,
          }) as V1MetricsViewAggregationMeasure,
      ),

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
      measures: activeMeasureNames.map((name) => ({ name })),
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

  $: ({ data: sortedData, isFetching } = $sortedQuery);
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
<<<<<<< HEAD
      $selectedValues.data ?? [],
      leaderboardTotal,
=======
      selectedValues,
      leaderboardTotals,
>>>>>>> 3f73fe2c8 (multiple measure in not expanded leaderboard display, grid tweaks to col)
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
        ...Object.fromEntries(activeMeasureNames.map((name) => [name, null])),
      }));

  $: belowTheFoldRows = belowTheFoldData.map((item) =>
    cleanUpComparisonValue(
      item,
      dimensionName,
      activeMeasureNames,
      leaderboardTotals,
      selectedValues.findIndex((value) =>
        compareLeaderboardValues(value, item[dimensionName]),
      ) ?? -1,
    ),
  );

  $: columnCount =
    1 + // Base column (dimension)
    activeMeasureNames.length *
      (1 + // Value column for each measure
        (showDeltaAbsolute ? 1 : 0) + // Delta absolute column for each measure
        (showDeltaPercent ? 1 : 0) + // Delta percent column for each measure
        (showPercentOfTotal ? 1 : 0)); // Percent of total column for each measure

  $: showDeltaAbsolute =
    !!comparisonTimeRange &&
    contextColumns.includes(LeaderboardContextColumn.DELTA_ABSOLUTE);

  $: showDeltaPercent =
    !!comparisonTimeRange &&
    contextColumns.includes(LeaderboardContextColumn.DELTA_PERCENT);

  $: showPercentOfTotal =
    !!comparisonTimeRange &&
    isValidPercentOfTotal &&
    contextColumns.includes(LeaderboardContextColumn.PERCENT);

  $: if (activeMeasureNames) {
    valueColumn.reset();
  }

  $: dimensionColumnWidth =
    !comparisonTimeRange && !isValidPercentOfTotal ? 240 : 164;

  $: tableWidth =
    dimensionColumnWidth +
    $valueColumn +
    (comparisonTimeRange
      ? DEFAULT_CONTEXT_COLUMN_WIDTH * (showDeltaPercent ? 2 : 1)
      : isValidPercentOfTotal
        ? DEFAULT_CONTEXT_COLUMN_WIDTH
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
      <col style:width="{gutterWidth}px" />
      <col style:width="{dimensionColumnWidth}px" />
      {#each activeMeasureNames as _, index (index)}
        <col style:width="{$valueColumn}px" data-index={index} />
        {#if showDeltaAbsolute}
          <col style:width="{DEFAULT_CONTEXT_COLUMN_WIDTH}px" />
          {#if showDeltaPercent}
            <col style:width="{DEFAULT_CONTEXT_COLUMN_WIDTH}px" />
          {/if}
        {:else if showDeltaPercent}
          <col style:width="{DEFAULT_CONTEXT_COLUMN_WIDTH}px" />
        {/if}
        {#if showPercentOfTotal}
          <col style:width="{DEFAULT_CONTEXT_COLUMN_WIDTH}px" />
        {/if}
      {/each}
    </colgroup>

    <LeaderboardHeader
      {hovered}
      displayName={displayName || dimensionName}
      dimensionDescription={description}
      {dimensionName}
      {isBeingCompared}
      {isFetching}
      {sortType}
      {isValidPercentOfTotal}
      {sortedAscending}
      isTimeComparisonActive={!!comparisonTimeRange}
      {contextColumns}
      {activeMeasureNames}
      {toggleSort}
      {setPrimaryDimension}
      {toggleComparisonDimension}
      {sortMeasure}
      {measureLabel}
    />

    <tbody>
      {#if isFetching}
        <LoadingRows columns={columnCount + 1} />
      {:else}
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
            {contextColumns}
            {activeMeasureNames}
            isTimeComparisonActive={!!comparisonTimeRange}
            {toggleDimensionValueSelection}
            {formatter}
          />
        {/each}
      {/if}

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
          isTimeComparisonActive={!!comparisonTimeRange}
          borderTop={i === 0}
          borderBottom={i === belowTheFoldRows.length - 1}
          {toggleDimensionValueSelection}
          {formatter}
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
  {:else if noAvailableValues && !isFetching}
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
