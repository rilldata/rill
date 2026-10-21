<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ExpressionFilterManager } from "../dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ExpressionFilters from "../dashboards/filters/ExpressionFilters.svelte";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";
  import type { DashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";

  let {
    expressionFilterManager,
    timeFilterManager,
    dashboardConfigProvider,
    maxWidth = undefined,
    // side = "bottom", // TODO
  }: {
    expressionFilterManager: ExpressionFilterManager;
    timeFilterManager: TimeFilterManager;
    dashboardConfigProvider: DashboardConfigProvider;
    maxWidth?: number | undefined;
    side?: "top" | "right" | "bottom" | "left";
  } = $props();

  let { timeStart, timeEnd, timeDimension } = $derived(timeFilterManager);
</script>

<div
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
  aria-label={m.report_form_filters_aria()}
>
  <TimeFilters
    {timeFilterManager}
    metricsViewsProvider={dashboardConfigProvider.metricsViewsProvider}
    yamlConfigProvider={dashboardConfigProvider.yamlConfigProvider}
    config={{
      showTimeDimensionSelector: true,
      showFullRange: true,
      showWatermark: true,
    }}
    context="adhoc"
  />

  <ExpressionFilters
    {expressionFilterManager}
    {dashboardConfigProvider}
    {timeStart}
    {timeEnd}
    {timeDimension}
    timeControlsReady
  />
</div>
