<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/canvas-components/render/VegaLiteRenderer.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { View } from "svelte-vega";
  import { getChartData } from "./selector";
  import type { ChartConfig, ChartType } from "./types";
  import { generateSpec } from "./util";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let renderer: string;

  $: console.log(renderer, rendererProperties);

  let stateManagers = getCanvasStateManagers();

  const instanceId = $runtime.instanceId;
  $: chartConfig = rendererProperties as ChartConfig;
  $: chartType = renderer as ChartType;

  let viewVL: View;

  $: data = getChartData(stateManagers, instanceId, chartConfig);
  $: spec = generateSpec(chartType, chartConfig);

  $: console.log($data, spec);
</script>

{#if chartConfig?.x}
  <VegaLiteRenderer
    bind:viewVL
    canvasDashboard={true}
    data={{ "metrics-view": $data }}
    {spec}
  />
{/if}
