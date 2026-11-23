<script lang="ts">
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ExpressionFilters from "../../dashboards/filters/ExpressionFilters.svelte";

  let {
    maxWidth,
    canvasName,
  }: {
    maxWidth: number;
    canvasName: string;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let {
    canvasEntity: {
      dashboardProvider,
      expressionFilterManager,
      timeFilterManager,
    },
  } = $derived(getCanvasStore(canvasName, runtimeClient.instanceId));
  let { metricsViewsProvider, yamlConfigProvider } =
    $derived(dashboardProvider);

  let { hasTimeSeries, interval, timeStart, timeEnd } =
    $derived(timeFilterManager);
</script>

<div
  role="presentation"
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
>
  {#if hasTimeSeries}
    <div class="px-2 pointer-events-auto">
      <TimeFilters
        {timeFilterManager}
        {metricsViewsProvider}
        {yamlConfigProvider}
        config={{
          allowCustomTimeRange: yamlConfigProvider.allowCustomTimeRange,
          showGrainPicker: true,
        }}
        context={canvasName}
      />
    </div>
  {/if}

  <div class="pointer-events-auto ml-2">
    <ExpressionFilters
      dashboardConfigProvider={dashboardProvider}
      {expressionFilterManager}
      {timeStart}
      {timeEnd}
      timeControlsReady={hasTimeSeries === false || !!interval}
    />
  </div>
</div>
