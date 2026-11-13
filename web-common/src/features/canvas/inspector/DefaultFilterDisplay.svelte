<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getCanvasStore } from "../state-managers/state-managers";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ReadonlyFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyFilters.svelte";
  import { V1ExploreComparisonMode } from "@rilldata/web-common/runtime-client";
  import { getDefaultComparisonRange } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
  import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
  import { untrack } from "svelte";

  let { canvasName }: { canvasName: string } = $props();

  const runtimeClient = useRuntimeClient();

  let {
    canvasEntity: { specStore, clearDefaultFilters, dashboardProvider },
  } = $derived(getCanvasStore(canvasName, runtimeClient.instanceId));

  let defaultTimeFilterManager = $derived(
    new TimeFilterManager(
      runtimeClient,
      dashboardProvider.metricsViewsProvider,
      dashboardProvider.yamlConfigProvider,
      // Only a range set in the defaults is shown, not the fallback the canvas would open with.
      { skipDefaultTimeRange: true },
    ),
  );
  $effect(() => {
    const manager = defaultTimeFilterManager;
    return () => manager.cleanup();
  });

  // The default time range is passed as a param since the manager skips its own default.
  let defaultTimeParams = $derived.by(() => {
    const params = new URLSearchParams();
    const defaultTimeRange =
      dashboardProvider.yamlConfigProvider.defaultTimeRange;
    if (defaultTimeRange) {
      params.set(ExploreStateURLParams.TimeRange, defaultTimeRange);
    }

    const defaultPreset = $specStore.data?.canvas?.defaultPreset;
    if (
      defaultPreset?.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
    ) {
      params.set(
        ExploreStateURLParams.ComparisonTimeRange,
        getDefaultComparisonRange(
          defaultPreset.timeRange,
          $specStore.data?.canvas?.timeRanges,
        ),
      );
    }
    return params;
  });
  $effect(() => {
    const manager = defaultTimeFilterManager;
    const params = defaultTimeParams;
    untrack(() => manager.storeSync.setUrlParams(params));
  });

  let defaultFiltersManager = $derived(
    new ExpressionFilterManager(
      dashboardProvider.metricsViewsProvider,
      dashboardProvider.yamlConfigProvider,
    ),
  );
  $effect(() => {
    const filterExpr = $specStore.data?.canvas?.defaultPreset?.filterExpr ?? {};
    dashboardProvider.metricsViewsProvider.metricsViewNames.forEach((key) => {
      defaultFiltersManager.setExprForMetricsView(
        key,
        filterExpr[key]?.expression,
      );
    });
  });
</script>

<div class="flex-col flex h-full">
  <div class="page-param">
    <p class="text-fg-secondary mb-4">
      The filters listed below are saved as your default view and will
      automatically apply each time you open this dashboard in Rill Cloud.
    </p>

    <ReadonlyFilters
      expressionFilterManager={defaultFiltersManager}
      timeFilterManager={defaultTimeFilterManager}
    />
  </div>

  <div class="mt-auto border-t w-full px-5 py-3">
    <Button type="secondary" wide onClick={clearDefaultFilters}>
      <Trash />
      Clear default filters
    </Button>
  </div>
</div>

<style lang="postcss">
  .page-param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
