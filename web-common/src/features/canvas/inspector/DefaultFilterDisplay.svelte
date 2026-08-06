<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getCanvasStore } from "../state-managers/state-managers";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";

  let { canvasName }: { canvasName: string } = $props();

  const runtimeClient = useRuntimeClient();

  let {
    canvasEntity: {
      specStore,
      clearDefaultFilters,
      timeManager: {
        state: { interval: _interval },
        defaultTimeRangeStore,
        defaultComparisonRangeStore,
      },
      metricsViewsProvider,
      yamlConfigProvider,
    },
  } = $derived(getCanvasStore(canvasName, runtimeClient.instanceId));

  let interval = $derived($_interval);

  let defaultTimeRange: V1TimeRange | undefined = $defaultTimeRangeStore
    ? {
        expression: $defaultTimeRangeStore,
      }
    : undefined;
  let defaultComparisonRange: V1TimeRange | undefined =
    $defaultComparisonRangeStore
      ? {
          expression: $defaultComparisonRangeStore,
        }
      : undefined;

  let defaultFiltersManager = $derived(
    new ExpressionFilterManager(metricsViewsProvider, yamlConfigProvider),
  );
  $effect(() => {
    if (!$specStore.data?.canvas?.defaultPreset?.filterExpr) return;
    Object.entries($specStore.data?.canvas?.defaultPreset?.filterExpr).forEach(
      ([key, value]) => {
        defaultFiltersManager.setExprForMetricsView(key, value?.expression);
      },
    );
  });
</script>

<div class="flex-col flex h-full">
  <div class="page-param">
    <p class="text-fg-secondary mb-4">
      The filters listed below are saved as your default view and will
      automatically apply each time you open this dashboard in Rill Cloud.
    </p>

    <ReadonlyExpressionFilters
      expressionFilterManager={defaultFiltersManager}
      displayTimeRange={defaultTimeRange}
      displayComparisonTimeRange={defaultComparisonRange}
      queryTimeStart={interval?.start?.toISO()}
      queryTimeEnd={interval?.end?.toISO()}
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
