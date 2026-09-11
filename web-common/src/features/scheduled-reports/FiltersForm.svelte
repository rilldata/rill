<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ExpressionFilterManager } from "../dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ExpressionFilters from "../dashboards/filters/ExpressionFilters.svelte";
  import { syncStoreWithSource } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";
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

  // svelte-ignore state_referenced_locally
  syncStoreWithSource(
    expressionFilterManager,
    async (newUrlParams) => expressionFilterManager.setUrlParams(newUrlParams),
    () => expressionFilterManager.metricsViewsProvider.ready,
  );
  // svelte-ignore state_referenced_locally
  syncStoreWithSource(
    timeFilterManager,
    async (newUrlParams) => timeFilterManager.setUrlParams(newUrlParams),
    () => timeFilterManager.ready,
  );

  let { timeStart, timeEnd, timeDimension } = $derived(timeFilterManager);
</script>

<div
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
  aria-label={m.report_form_filters_aria()}
>
  <TimeFilters
    {timeFilterManager}
    {dashboardConfigProvider}
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
