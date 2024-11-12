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
  import { getComparisonRequestMeasures } from "../dashboard-utils";
  import { mergeDimensionAndMeasureFilter } from "../filters/measure-filters/measure-filter-utils";
  import { SortType } from "../proto-state/derived-types";
  import {
    additionalMeasures,
    getFiltersForOtherDimensions,
  } from "../selectors";
  import { getIndependentMeasures } from "../state-managers/selectors/measures";
  import {
    createAndExpression,
    createOrExpression,
    sanitiseExpression,
  } from "../stores/filter-utils";
  import type { DimensionThresholdFilter } from "../stores/metrics-explorer-entity";
  import {
    cleanUpComparisonValue,
    compareLeaderboardValues,
    getSort,
    prepareLeaderboardItemData,
    type LeaderboardItemData,
  } from "./leaderboard-utils";
  import {
    LEADERBOARD_DEFAULT_COLUMN_WIDTHS,
    type ColumnWidths,
  } from "./leaderboard-widths";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardRow from "./LeaderboardRow.svelte";
  import LoadingRows from "./LoadingRows.svelte";

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
  export let sortedAscending: boolean;
  export let isValidPercentOfTotal: boolean;
  export let timeControlsReady: boolean;
  export let isSummableMeasure: boolean;
  export let filterExcludeMode: boolean;
  export let atLeastOneActive: boolean;
  export let isBeingCompared: boolean;
  export let parentElement: HTMLElement;
  export let columnWidths: ColumnWidths = LEADERBOARD_DEFAULT_COLUMN_WIDTHS;
  export let estimateLargestLeaderboardWidth: (
    dimensionName: string,
    aboveTheFold: LeaderboardItemData[],
    belowTheFold: LeaderboardItemData[],
  ) => void;
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

  $: where = sanitiseExpression(
    mergeDimensionAndMeasureFilter(
      getFiltersForOtherDimensions(whereFilter, dimensionName),
      dimensionThresholdFilters,
    ),
    undefined,
  );

  $: measures = getIndependentMeasures(
    metricsView,
    additionalMeasures(activeMeasureName, dimensionThresholdFilters),
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

  $: sort = getSort(
    sortedAscending,
    sortType,
    activeMeasureName,
    dimensionName,
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

  // Estimate the common column widths for all leaderboards
  $: if (aboveTheFold.length || belowTheFoldRows.length) {
    estimateLargestLeaderboardWidth(
      dimensionName,
      aboveTheFold,
      belowTheFoldRows,
    );
  }

  $: tableWidth =
    columnWidths.dimension +
    columnWidths.value +
    (comparisonTimeRange
      ? columnWidths.delta + columnWidths.deltaPercent
      : isValidPercentOfTotal
        ? columnWidths.percentOfTotal
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
      <col style:width="{columnWidths.dimension}px" />
      <col style:width="{columnWidths.value}px" />
      {#if !!comparisonTimeRange}
        <col style:width="{columnWidths.delta}px" />
        <col style:width="{columnWidths.deltaPercent}px" />
      {:else if isValidPercentOfTotal}
        <col style:width="{columnWidths.percentOfTotal}px" />
      {/if}
    </colgroup>

    <LeaderboardHeader
      {hovered}
      displayName={displayName ?? dimensionName}
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
    />

    <tbody>
      {#if isFetching}
        <LoadingRows columns={columnCount + 1} />
      {:else}
        {#each aboveTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardRow
            {tableWidth}
            {isSummableMeasure}
            {isBeingCompared}
            {filterExcludeMode}
            {atLeastOneActive}
            {dimensionName}
            {uri}
            {itemData}
            {isValidPercentOfTotal}
            isTimeComparisonActive={!!comparisonTimeRange}
            {columnWidths}
            {gutterWidth}
            {toggleDimensionValueSelection}
            {formatter}
          />
        {/each}
      {/if}

      {#each belowTheFoldRows as itemData, i (itemData.dimensionValue)}
        <LeaderboardRow
          {itemData}
          {isSummableMeasure}
          {tableWidth}
          {dimensionName}
          {isBeingCompared}
          {uri}
          {filterExcludeMode}
          {atLeastOneActive}
          {isValidPercentOfTotal}
          isTimeComparisonActive={!!comparisonTimeRange}
          {columnWidths}
          {gutterWidth}
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
  {:else if noAvailableValues}
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
