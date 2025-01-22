<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
  import ComponentTitle from "@rilldata/web-common/features/canvas/ComponentTitle.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type {
    MetricsViewSpecMeasureV2,
    V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
  import type { View } from "svelte-vega";
  import { getChartData } from "./selector";
  import type { ChartType } from "./types";
  import { generateSpec, getChartTitle, mergedVlConfig } from "./util";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let renderer: string;

  const {
    canvasEntity: {
      spec: { getMeasureForMetricView },
    },
  } = getCanvasStateManagers();

  let stateManagers = getCanvasStateManagers();
  let viewVL: View;

  $: chartConfig = rendererProperties as ChartSpec;
  $: chartType = renderer as ChartType;

  $: data = getChartData(stateManagers, chartConfig);
  $: spec = generateSpec(chartType, chartConfig, $data);

  $: measure = getMeasureForMetricView(
    chartConfig.y?.field,
    chartConfig.metrics_view,
  );

  $: measureFormatter = createMeasureValueFormatter<null | undefined>(
    $measure as MetricsViewSpecMeasureV2,
  );

  $: config = chartConfig.vl_config
    ? mergedVlConfig(chartConfig.vl_config)
    : undefined;

  $: title = getChartTitle(chartConfig, $data);
</script>

{#if chartConfig?.x}
  {#if $data.isFetching}
    <div class="flex items-center h-full w-full">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else if $data.error}
    <div class="text-red-500">{$data.error.message}</div>
  {:else}
    {#if !chartConfig.title && !chartConfig.description}
      <ComponentTitle faint {title} />
    {/if}
    <VegaLiteRenderer
      bind:viewVL
      canvasDashboard
      data={{ "metrics-view": $data.data }}
      {spec}
      expressionFunctions={{
        measureFormatter: { fn: (val) => measureFormatter(val) },
      }}
      {config}
    />
  {/if}
{/if}
