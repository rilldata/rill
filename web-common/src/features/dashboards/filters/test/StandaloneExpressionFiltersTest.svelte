<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ExpressionFilters.svelte";
  import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
  import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  /**
   * Test component that renders the filter bar on its own, the way canvas does: the filter manager
   * holds the filter state and the url is written straight from it, with no dashboard state in
   * between. `onManagerCreated` hands the manager to the test, which is how the test asserts on the
   * filter the bar produced.
   */
  let {
    metricsViewNames,
    initUrlSearch,
    onManagerCreated,
  }: {
    metricsViewNames: string[];
    // Url search the bar starts from, for a test that begins with a filter already applied.
    initUrlSearch?: string;
    onManagerCreated?: (
      expressionFilterManager: ExpressionFilterManager,
    ) => void;
  } = $props();

  // The props are fixed for the lifetime of a test, so reading them once at init is enough.
  // svelte-ignore state_referenced_locally
  const metricsViewsProvider = new MetricsViewsProvider(
    useRuntimeClient(),
    metricsViewNames,
  );
  const expressionFilterManager = new ExpressionFilterManager(
    metricsViewsProvider,
    new YAMLConfigProvider(),
  );
  // svelte-ignore state_referenced_locally
  metricsViewsProvider.setMetricsViewNames(metricsViewNames);
  // svelte-ignore state_referenced_locally
  onManagerCreated?.(expressionFilterManager);

  // svelte-ignore state_referenced_locally
  expressionFilterManager.storeSync.setUrlParams(
    new URLSearchParams(initUrlSearch),
  );
</script>

<!-- The app layout supplies the tooltip provider, which the advanced filter pill needs. -->
<Tooltip.Provider>
  <ExpressionFilters
    {expressionFilterManager}
    timeStart={undefined}
    timeEnd={undefined}
    timeControlsReady
  />
  {#if metricsViewsProvider.ready}
    <div>Dashboard loaded!</div>
  {/if}
</Tooltip.Provider>
