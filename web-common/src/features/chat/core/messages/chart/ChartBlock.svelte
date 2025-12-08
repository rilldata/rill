<!--
  Renders a chart block with a collapsible tool call header.
  Shows the chart visualization with expandable request/response details.
-->
<script lang="ts">
  import { page } from "$app/stores";
  import {
    ChartContainer,
    type ChartType,
  } from "@rilldata/web-common/features/components/charts";
  import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard";
  import { readable } from "svelte/store";
  import type { V1Message, V1Tool } from "../../../../../runtime-client";
  import ToolCallHeader from "../shared/ToolCallHeader.svelte";

  export let message: V1Message;
  export let resultMessage: V1Message;
  export let chartType: ChartType;
  export let chartSpec: any;
  export let tools: V1Tool[] | undefined = undefined;

  // Page params for chart
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
        start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
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

<div class="chart-block">
  <ToolCallHeader {message} {resultMessage} {tools} />

  <div class="chart-container">
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
</div>

<style lang="postcss">
  .chart-block {
    @apply w-full max-w-full self-start;
  }

  .chart-container {
    @apply border rounded-md border-gray-200 px-1 py-2;
    @apply w-full h-[400px];
    background: var(--surface);
  }
</style>
