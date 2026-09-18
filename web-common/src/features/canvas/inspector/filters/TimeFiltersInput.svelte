<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ALL_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";
  import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

  let {
    id,
    localTimeFilters,
    metricsViewsProvider,
    yamlConfigProvider,
    showComparison,
    showGrain,
    canvasName,
  }: {
    id: string;
    localTimeFilters: TimeFilterManager;
    metricsViewsProvider: MetricsViewsProvider;
    yamlConfigProvider: YAMLConfigProvider;
    showComparison: boolean;
    showGrain: boolean;
    canvasName: string;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let { instanceId } = $derived(runtimeClient);

  let {
    canvasEntity: { timeFilterManager, dashboardProvider },
  } = $derived(getCanvasStore(canvasName, instanceId));

  let { curParams } = $derived(localTimeFilters);
  $effect(() => console.log(localTimeFilters.timeRange));

  let globalRange = $derived(timeFilterManager.timeRange);

  let localFiltersEnabled = $derived(Boolean(curParams.size));

  let defaultTimeRange = $derived(
    dashboardProvider.yamlConfigProvider.defaultTimeRange,
  );
</script>

<div class="flex flex-col gap-y-1 pt-1">
  <div class="flex justify-between">
    <InputLabel
      capitalize={false}
      small
      label={m.canvas_local_time_range()}
      {id}
      faint={!localFiltersEnabled}
    />
    <Switch
      checked={localFiltersEnabled}
      onclick={() => {
        if (localFiltersEnabled) {
          localTimeFilters.setUrlParams(new URLSearchParams());
        } else {
          void localTimeFilters.onSelectRange(
            globalRange ?? defaultTimeRange ?? ALL_TIME_RANGE_ALIAS,
          );
        }
      }}
      small
    />
  </div>
  <div class="text-fg-secondary">
    {#if localFiltersEnabled}
      {m.canvas_overriding_inherited_time_filters()}
    {:else}
      {m.canvas_override_inherited_time_filters_hint()}
    {/if}
  </div>

  {#if localFiltersEnabled}
    <TimeFilters
      {timeFilterManager}
      {metricsViewsProvider}
      {yamlConfigProvider}
      context="filter-input"
      config={{
        showFullRange: false,
        hidePan: true,
        showGrainSelector: showGrain,
        showComparisonSelector: showComparison,
      }}
    />
  {/if}
</div>
