<script lang="ts">
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ExpressionFilters.svelte";
  import { syncStoreWithSource } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { DashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";

  /**
   * Test component that renders the filter bar on its own, the way canvas does: the filter manager
   * holds the filter state and the url is written straight from it, with no dashboard state in
   * between. `onManagerCreated` hands the manager to the test, which is how the test asserts on the
   * filter the bar produced.
   */
  let {
    metricsViewNames,
    onManagerCreated,
  }: {
    metricsViewNames: string[];
    onManagerCreated?: (
      expressionFilterManager: ExpressionFilterManager,
    ) => void;
  } = $props();

  const dashboardConfigProvider = new DashboardConfigProvider(
    useRuntimeClient(),
  );
  const { metricsViewsProvider, yamlConfigProvider } = dashboardConfigProvider;
  // The props are fixed for the lifetime of a test, so setting metricsViewNames once is enough.
  // svelte-ignore state_referenced_locally
  metricsViewsProvider.setMetricsViewNames(metricsViewNames);

  const expressionFilterManager = new ExpressionFilterManager(
    metricsViewsProvider,
    yamlConfigProvider,
  );
  // svelte-ignore state_referenced_locally
  onManagerCreated?.(expressionFilterManager);

  syncStoreWithSource(
    expressionFilterManager,
    async (newUrlParams) => expressionFilterManager.setUrlParams(newUrlParams),
    () => metricsViewsProvider.ready,
    undefined,
    true,
  );
</script>

<ExpressionFilters
  {expressionFilterManager}
  {dashboardConfigProvider}
  timeStart={undefined}
  timeEnd={undefined}
  timeControlsReady
/>
{#if metricsViewsProvider.ready}
  <div>Dashboard loaded!</div>
{/if}
