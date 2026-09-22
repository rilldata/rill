<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getCanvasStore } from "../state-managers/state-managers";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import { V1ExploreComparisonMode } from "@rilldata/web-common/runtime-client";
  import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

  let { canvasName }: { canvasName: string } = $props();

  const runtimeClient = useRuntimeClient();

  let {
    canvasEntity: { specStore, clearDefaultFilters, dashboardProvider },
  } = $derived(getCanvasStore(canvasName, runtimeClient.instanceId));

  let defaultExpressionFiltersManager = $derived(
    new ExpressionFilterManager(
      dashboardProvider.metricsViewsProvider,
      dashboardProvider.yamlConfigProvider,
    ),
  );
  let defaultTimeFilterManager = $derived(
    new TimeFilterManager(
      runtimeClient,
      dashboardProvider.metricsViewsProvider,
      dashboardProvider.yamlConfigProvider,
      false,
    ),
  );

  $effect(() => {
    const filterExpr = $specStore.data?.canvas?.defaultPreset?.filterExpr ?? {};
    dashboardProvider.metricsViewsProvider.metricsViewNames.forEach((key) => {
      defaultExpressionFiltersManager.setExprForMetricsView(
        key,
        filterExpr[key]?.expression,
      );
    });
  });

  $effect(() => {
    if (!$specStore.data?.canvas?.defaultPreset?.timeRange) {
      defaultTimeFilterManager.setUrlParams(new URLSearchParams());
      return;
    }

    void defaultTimeFilterManager.onSelectRange(
      $specStore.data.canvas.defaultPreset.timeRange,
    );

    if (
      $specStore.data.canvas.defaultPreset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
    ) {
      defaultTimeFilterManager.setShowComparison(true);
    } else {
      defaultTimeFilterManager.setShowComparison(false);
    }
  });
</script>

<div class="flex-col flex h-full">
  <div class="page-param">
    <p class="text-fg-secondary mb-4">
      The filters listed below are saved as your default view and will
      automatically apply each time you open this dashboard in Rill Cloud.
    </p>

    <ReadonlyExpressionFilters
      expressionFilterManager={defaultExpressionFiltersManager}
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
