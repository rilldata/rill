<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
  import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import { onDestroy } from "svelte";

  let {
    metricsViewName,
    filters,
    dimensionsWithInlistFilter,
    timeRange,
    comparisonTimeRange,
  }: {
    metricsViewName: string;
    filters: V1Expression | undefined;
    dimensionsWithInlistFilter: string[];
    timeRange: V1TimeRange | undefined;
    comparisonTimeRange: V1TimeRange | undefined;
  } = $props();

  const runtimeClient = useRuntimeClient();

  const metricsViewProvider = new MetricsViewsProvider(runtimeClient, []);
  $effect(() => metricsViewProvider.setMetricsViewNames([metricsViewName]));
  const yamlConfigProvider = new YAMLConfigProvider();
  const expressionFilterManager = new ExpressionFilterManager(
    metricsViewProvider,
    yamlConfigProvider,
  );
  $effect(() =>
    expressionFilterManager.setExprForMetricsView(
      metricsViewName,
      filters,
      dimensionsWithInlistFilter,
    ),
  );

  // time range could be an empty object sometimes
  let hasTimeRange = $derived(timeRange && Object.keys(timeRange).length > 0);
  let filtersLength = $derived(
    (filters?.cond?.exprs?.length ?? 0) + (hasTimeRange ? 1 : 0),
  );

  onDestroy(() => {
    metricsViewProvider.cleanup();
    yamlConfigProvider.cleanup?.();
  });
</script>

<div class="flex flex-col gap-y-3" aria-label={m.alert_metadata_filters_aria()}>
  <MetadataLabel>
    {m.alert_filters_label({ count: String(filtersLength) })}
  </MetadataLabel>
  <ReadonlyExpressionFilters
    {expressionFilterManager}
    displayTimeRange={timeRange}
    displayComparisonTimeRange={comparisonTimeRange}
    queryTimeStart={timeRange?.start}
    queryTimeEnd={timeRange?.end}
  />
</div>
