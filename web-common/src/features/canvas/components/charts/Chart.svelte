<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { View } from "svelte-vega";
  import { getChartData } from "./selector";
  import type { ChartConfig, ChartType } from "./types";
  import { generateSpec } from "./util";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let renderer: string;

  let stateManagers = getCanvasStateManagers();

  const instanceId = $runtime.instanceId;
  $: chartConfig = rendererProperties as ChartConfig;
  $: chartType = renderer as ChartType;

  let viewVL: View;

  $: data = getChartData(stateManagers, instanceId, chartConfig);
  $: spec = generateSpec(chartType, chartConfig, $data);
</script>

{#if chartConfig?.x}
  {#if $data.isFetching}
    <div class="flex items-center h-full w-full">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else if $data.error}
    <div class="text-red-500">{$data.error.message}</div>
  {:else}
    <VegaLiteRenderer
      bind:viewVL
      canvasDashboard
      data={{ "metrics-view": $data.data }}
      {spec}
    />
  {/if}
{/if}
