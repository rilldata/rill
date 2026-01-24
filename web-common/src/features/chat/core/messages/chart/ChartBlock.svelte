<!--
  Renders a chart block with a collapsible tool call header.
  Shows the chart visualization with expandable request/response details.
-->
<script lang="ts">
  import { page } from "$app/stores";
  import { ChartContainer } from "@rilldata/web-common/features/components/charts";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard";
  import { Theme } from "@rilldata/web-common/features/themes/theme";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { readable } from "svelte/store";
  import type { V1Tool } from "../../../../../runtime-client";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { ChartBlock } from "./chart-block";

  export let block: ChartBlock;
  export let tools: V1Tool[] | undefined = undefined;

  // Page params for chart
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: instanceId = $runtime.instanceId;

  // Cast chartSpec to any for property access (type comes from parsed JSON)
  $: chartSpec = block.chartSpec as any;

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

  $: defaultThemeQuery = createRuntimeServiceGetInstance(
    instanceId,
    undefined,
    {
      query: {
        select: (data) => data?.instance?.theme,
      },
    },
    queryClient,
  );

  $: themeName = $defaultThemeQuery?.data;

  $: themeQuery = useResource(
    instanceId,
    themeName!,
    ResourceKind.Theme,
    {
      enabled: !!themeName,
      select: (data) => {
        if (data.resource?.theme?.spec) {
          return new Theme(data.resource?.theme?.spec);
        } else {
          return undefined;
        }
      },
    },
    queryClient,
  );
</script>

<div class="chart-block">
  <ToolCall
    message={block.message}
    resultMessage={block.resultMessage}
    {tools}
    variant="block"
  />

  <div class="chart-container">
    <ChartContainer
      chartType={block.chartType}
      {spec}
      {timeAndFilterStore}
      {project}
      theme={$themeQuery?.data}
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
