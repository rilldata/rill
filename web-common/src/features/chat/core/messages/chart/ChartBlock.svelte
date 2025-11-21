<script lang="ts">
  import { page } from "$app/stores";
  import {
    ChartContainer,
    type ChartType,
  } from "@rilldata/web-common/features/components/charts";
  import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard";
  import { readable } from "svelte/store";

  export let chartType: ChartType;
  export let chartSpec: any;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: spec = readable(chartSpec);

  // Extract time range from the chart spec or use defaults
  $: timeRange = chartSpec.time_range
    ? {
        start: chartSpec.time_range.start,
        end: chartSpec.time_range.end,
        timeZone: chartSpec.time_range.time_zone || "UTC",
      }
    : {
        start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(), // Default to last 7 days
        end: new Date().toISOString(),
        timeZone: "UTC",
      };

  $: timeAndFilterStore = readable({
    timeRange: timeRange,
    comparisonTimeRange: undefined,
    showTimeComparison: false,
    where: mapResolverExpressionToV1Expression(chartSpec.where) || {
      cond: {
        op: "OPERATION_AND",
        exprs: [],
      },
    },
    timeGrain: chartSpec.time_grain || "TIME_GRAIN_DAY",
    timeRangeState: undefined,
    comparisonTimeRangeState: undefined,
    hasTimeSeries: true,
  });
</script>

<div
  class="border rounded-md border-gray-150 px-1 py-2 bg-surface w-full h-[400px]"
>
  <ChartContainer
    {chartType}
    {spec}
    {timeAndFilterStore}
    {project}
    showExploreLink
    {organization}
    themeMode="light"
  />
</div>
