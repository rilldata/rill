<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
  import type {
    MetricsViewSpecDimensionV2,
    V1Expression,
    V1MetricsViewAggregationMeasure,
    V1MetricsViewSpec,
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
  import { mergeDimensionAndMeasureFilter } from "../filters/measure-filters/measure-filter-utils";
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
    DEFAULT_COL_WIDTH,
    deltaColumn,
    valueColumn,
  } from "./leaderboard-widths";

  const slice = 7;
  const gutterWidth = 24;
  const queryLimit = 8;

  export let dimension: MetricsViewSpecDimensionV2;
  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let selectedValues: string[];
  export let instanceId: string;
  export let whereFilter: V1Expression;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let activeMeasureName: string;
  export let metricsViewName: string;
  export let metricsView: V1MetricsViewSpec;
  export let sortType: SortType;
  export let tableWidth: number;
  export let sortedAscending: boolean;
  export let isValidPercentOfTotal: boolean;
  export let timeControlsReady: boolean;
  export let firstColumnWidth: number;
  export let isSummableMeasure: boolean;
  export let filterExcludeMode: boolean;
  export let atLeastOneActive: boolean;
  export let isBeingCompared: boolean;
  export let parentElement: HTMLElement;
  export let suppressTooltip = false;
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

  $: isComplexFilter = isExpressionUnsupported(whereFilter);
  $: where = isComplexFilter
    ? whereFilter
    : sanitiseExpression(
        mergeDimensionAndMeasureFilter(
          getFiltersForOtherDimensions(whereFilter, dimensionName),
          dimensionThresholdFilters,
        ),
        undefined,
      );

  $: measures = additionalMeasures(activeMeasureName, dimensionThresholdFilters)
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
    )
    .concat(uri ? [getURIRequestMeasure(dimensionName)] : []);

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
      measures: [{ name: activeMeasureName }],
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

  $: leaderboardTotal = totalsData?.data?.[0]?.[activeMeasureName] as
    | number
    | null;

  $: ({ aboveTheFold, belowTheFoldValues, noAvailableValues, showExpandTable } =
    prepareLeaderboardItemData(
      sortedData?.data,
      dimensionName,
      activeMeasureName,
      slice,
      selectedValues,
      leaderboardTotal,
    ));

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
    },
    {
      query: {
        enabled: !!belowTheFoldValues.length && timeControlsReady && visible,
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
      activeMeasureName,
      leaderboardTotal,
      selectedValues.findIndex((value) =>
        compareLeaderboardValues(value, item[dimensionName]),
      ),
    ),
  );

  $: columnCount = comparisonTimeRange ? 3 : isValidPercentOfTotal ? 2 : 1;
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
      <col style:width="{firstColumnWidth}px" />
      <col style:width="{$valueColumn}px" />
      {#if !!comparisonTimeRange}
        <col style:width="{$deltaColumn}px" />
        <col style:width="{DEFAULT_COL_WIDTH}px" />
      {:else if isValidPercentOfTotal}
        <col style:width="{DEFAULT_COL_WIDTH}px" />
      {/if}
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
      {toggleSort}
      {setPrimaryDimension}
      {toggleComparisonDimension}
    />

    <tbody>
      {#if isFetching}
        <LoadingRows columns={columnCount + 1} />
      {:else}
        {#each aboveTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardRow
            {suppressTooltip}
            {tableWidth}
            {firstColumnWidth}
            {isSummableMeasure}
            {isBeingCompared}
            {filterExcludeMode}
            {atLeastOneActive}
            {dimensionName}
            {itemData}
            {isValidPercentOfTotal}
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
          {firstColumnWidth}
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
        (Expand Table)
      </button>
      <TooltipContent slot="tooltip-content">
        Expand dimension to see more values
      </TooltipContent>
    </Tooltip>
  {:else if noAvailableValues && !isFetching}
    <div class="table-message ui-copy-muted">(No available values)</div>
  {/if}
</div>

<style lang="postcss">
  table {
    @apply p-0 m-0 border-spacing-0 border-collapse w-fit;
    @apply font-normal cursor-pointer select-none;
    @apply table-fixed;
  }

  .table-message {
    @apply h-[22px] p-1 flex-row w-full text-left pl-7;
  }
</style>
