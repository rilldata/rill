<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ExpressionFilterManager } from "../dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ExpressionFilters from "../dashboards/filters/ExpressionFilters.svelte";
  import type { DashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";
  import type { TimeFiltersConfig } from "@rilldata/web-common/features/dashboards/time-controls/time-filters-config.ts";

  let {
    expressionFilterManager,
    timeFilterManager,
    dashboardConfigProvider,
    timeFiltersConfig = {
      showTimeDimensionSelector: true,
      showFullRange: true,
      showWatermark: true,
    },
    maxWidth = undefined,
  }: {
    expressionFilterManager: ExpressionFilterManager;
    timeFilterManager: TimeFilterManager;
    dashboardConfigProvider: DashboardConfigProvider;
    // Defaults to the explore controls; a canvas spans several metrics views, so it has no time dimension selector.
    timeFiltersConfig?: TimeFiltersConfig;
    maxWidth?: number | undefined;
  } = $props();

  let { hasTimeSeries, timeStart, timeEnd, timeDimension } =
    $derived(timeFilterManager);
</script>

<div
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
  aria-label={m.report_form_filters_aria()}
>
  {#if hasTimeSeries}
    <TimeFilters
      {timeFilterManager}
      metricsViewsProvider={dashboardConfigProvider.metricsViewsProvider}
      yamlConfigProvider={dashboardConfigProvider.yamlConfigProvider}
      config={timeFiltersConfig}
      context="adhoc"
    />
  {/if}

  <ExpressionFilters
    {expressionFilterManager}
    {dashboardConfigProvider}
    {timeStart}
    {timeEnd}
    {timeDimension}
    timeControlsReady
  />
</div>
