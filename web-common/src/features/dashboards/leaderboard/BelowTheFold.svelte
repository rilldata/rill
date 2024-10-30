<script lang="ts">
  import {
    createQueryServiceMetricsViewAggregation,
    V1Operation,
    type V1Expression,
    type V1MetricsViewAggregationMeasure,
    type V1MetricsViewAggregationSort,
    type V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import LeaderboardRow from "./LeaderboardRow.svelte";
  import {
    cleanUpComparisonValue,
    compareLeaderboardValues,
  } from "./leaderboard-utils";
  import {
    createAndExpression,
    createOrExpression,
    sanitiseExpression,
  } from "../stores/filter-utils";

  export let instanceId: string;
  export let metricsViewName: string;
  export let dimensionName: string;
  export let dimensionValues: string[];
  export let timeRange: V1TimeRange;
  export let activeMeasureName: string;
  export let enabled: boolean;
  export let where: V1Expression | undefined;
  export let measures: V1MetricsViewAggregationMeasure[];
  export let tableWidth: number;
  export let columnWidth: number;
  export let sort: V1MetricsViewAggregationSort[];
  export let gutterWidth: number;
  export let selectedValues: string[];
  export let isSummableMeasure: boolean;
  export let filterExcludeMode: boolean;
  export let atLeastOneActive: boolean;
  export let isBeingCompared: boolean;
  export let total: number | null;
  export let firstColumnWidth: number;
  export let uri: string | undefined;
  export let comparisonTimeRange: V1TimeRange | undefined = undefined;
  export let toggleDimensionValueSelection: (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean | undefined,
    isExclusiveFilter?: boolean | undefined,
  ) => void;
  export let formatter:
    | ((_value: number | undefined) => undefined)
    | ((value: string | number) => string);

  $: belowTheFoldDataQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      dimensions: [{ name: dimensionName }],
      where: sanitiseExpression(
        createAndExpression(
          [
            createOrExpression(
              dimensionValues.map((dimensionValue) => {
                return {
                  cond: {
                    op: V1Operation.OPERATION_EQ,
                    exprs: [{ ident: dimensionName }, { val: dimensionValue }],
                  },
                };
              }),
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
        enabled,
      },
    },
  );

  $: ({ data } = $belowTheFoldDataQuery);

  $: rows = (data?.data ?? []).map((item) =>
    cleanUpComparisonValue(
      item,
      dimensionName,
      activeMeasureName,
      total,
      selectedValues.findIndex((value) =>
        compareLeaderboardValues(value, item[dimensionName]),
      ),
    ),
  );
</script>

{#each rows as itemData, i (i)}
  <LeaderboardRow
    {itemData}
    {isSummableMeasure}
    {tableWidth}
    {dimensionName}
    {isBeingCompared}
    {uri}
    {filterExcludeMode}
    {atLeastOneActive}
    isValidPercentOfTotal={false}
    isTimeComparisonActive={!!comparisonTimeRange}
    {columnWidth}
    {gutterWidth}
    {firstColumnWidth}
    borderTop={i === 0}
    borderBottom={i === dimensionValues.length - 1}
    {toggleDimensionValueSelection}
    {formatter}
  />
{/each}
