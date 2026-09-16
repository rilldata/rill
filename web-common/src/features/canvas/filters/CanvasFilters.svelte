<script lang="ts">
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ExpressionFilters from "../../dashboards/filters/ExpressionFilters.svelte";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";

  let {
    maxWidth,
    canvasName,
  }: {
    readOnly?: boolean;
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

  let {
    hasTimeSeries,
    timeStart,
    timeEnd,
    ready: timeControlsReady,
  } = $derived(timeFilterManager);
</script>

<div
  role="presentation"
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
>
  {#if hasTimeSeries}
    <TimeFilters
      {timeFilterManager}
      dashboardConfigProvider={dashboardProvider}
      config={{
        showTimeDimensionSelector: false,
        showFullRange: false,
        showWatermark: false,
      }}
      context="canvas"
    />
  {/if}

  <div class="pointer-events-auto ml-2">
    <ExpressionFilters
      dashboardConfigProvider={dashboardProvider}
      {expressionFilterManager}
      {timeStart}
      {timeEnd}
      {timeControlsReady}
    />
  </div>
</div>
