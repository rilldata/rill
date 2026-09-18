<!--
  Renders a chart block with a collapsible tool call header.
  Shows the chart visualization with expandable request/response details.
-->
<script lang="ts">
  import { page } from "$app/stores";
  import { ChartContainer } from "@rilldata/web-common/features/components/charts";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
  import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard";
  import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import { Theme } from "@rilldata/web-common/features/themes/theme";
  import { DEFAULT_TIMEZONE } from "@rilldata/web-common/lib/time/config";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onDestroy, untrack } from "svelte";
  import { toStore } from "svelte/store";
  import type { V1Tool } from "../../../../../runtime-client";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { ChartBlock } from "./chart-block";

  let {
    block,
    tools = undefined,
  }: {
    block: ChartBlock;
    tools?: V1Tool[] | undefined;
  } = $props();

  const runtimeClient = useRuntimeClient();

  // Page params for chart
  let organization = $derived($page.params.organization);
  let project = $derived($page.params.project);

  // Cast chartSpec to any for property access (type comes from parsed JSON)
  let chartSpec = $derived(block.chartSpec as any);

  const spec = toStore(() => chartSpec);

  // The chart owns its filter state: the spec is a snapshot the agent resolved, not something the
  // user can change, so nothing else shares these managers.
  const metricsViewsProvider = new MetricsViewsProvider(runtimeClient, []);
  const yamlConfigProvider = new YAMLConfigProvider();
  const expressionFilterManager = new ExpressionFilterManager(
    metricsViewsProvider,
    yamlConfigProvider,
  );
  const timeFilterManager = new TimeFilterManager(
    runtimeClient,
    metricsViewsProvider,
    yamlConfigProvider,
    true,
  );
  onDestroy(() => {
    metricsViewsProvider.cleanup();
    yamlConfigProvider.cleanup?.();
  });

  $effect(() => {
    metricsViewsProvider.setMetricsViewNames(
      chartSpec.metrics_view ? [chartSpec.metrics_view] : [],
    );
  });

  // The agent returns an absolute window rather than a URL driven range.
  // `<start>,<end>` is the custom range form of a rill time, so the spec seeds the time manager
  // through the same params a dashboard uses.
  let timeUrlParams = $derived.by(() => {
    const urlParams = new URLSearchParams();

    const start = toRillTimeBound(chartSpec.time_range?.start);
    const end = toRillTimeBound(chartSpec.time_range?.end);
    // A spec without a usable range falls back to the last 7 days, as it did before the managers.
    urlParams.set(
      ExploreStateURLParams.TimeRange,
      start && end ? `${start},${end}` : TimeRangePreset.LAST_7_DAYS,
    );

    urlParams.set(
      ExploreStateURLParams.TimeZone,
      chartSpec.time_range?.time_zone ?? DEFAULT_TIMEZONE,
    );

    // The param is a Luxon unit rather than a V1TimeGrain.
    const grain = V1TimeGrainToDateTimeUnit[chartSpec.time_grain];
    if (grain) urlParams.set(ExploreStateURLParams.TimeGrain, grain);

    const comparisonStart = toRillTimeBound(
      chartSpec.comparison_time_range?.start,
    );
    const comparisonEnd = toRillTimeBound(chartSpec.comparison_time_range?.end);
    if (comparisonStart && comparisonEnd) {
      urlParams.set(
        ExploreStateURLParams.ComparisonTimeRange,
        `${comparisonStart},${comparisonEnd}`,
      );
    }

    return urlParams;
  });

  // The time range is resolved against the metrics views, so seeding no-ops until their specs
  // have loaded. Hence an effect on the provider rather than a one-shot on mount.
  $effect(() => {
    if (!metricsViewsProvider.ready) return;

    const urlParams = timeUrlParams;
    const metricsViewName = chartSpec.metrics_view;
    const where = mapResolverExpressionToV1Expression(chartSpec.where);

    untrack(() => {
      timeFilterManager.setUrlParams(urlParams);
      expressionFilterManager.setExprForMetricsView(metricsViewName, where);
    });
  });

  /** The rill time grammar only accepts UTC timestamps, so normalize whatever the agent emitted. */
  function toRillTimeBound(bound: string | undefined) {
    if (!bound) return undefined;
    const date = new Date(bound);
    if (isNaN(date.getTime())) return undefined;
    return date.toISOString();
  }

  let defaultThemeQuery = $derived(
    createRuntimeServiceGetInstance(
      runtimeClient,
      {},
      {
        query: {
          select: (data) => data?.instance?.theme,
        },
      },
      queryClient,
    ),
  );

  let themeName = $derived($defaultThemeQuery?.data);

  let themeQuery = $derived(
    useResource(
      runtimeClient,
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
    ),
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
      {expressionFilterManager}
      {timeFilterManager}
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
    background: var(--surface-subtle);
  }
</style>
