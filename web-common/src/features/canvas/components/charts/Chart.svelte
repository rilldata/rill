<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { getComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type {
    MetricsViewSpecMeasureV2,
    V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import type { View } from "vega-typings";
  import { getChartData, validateChartSchema } from "./selector";
  import type { ChartType } from "./types";
  import {
    generateSpec,
    getChartTitle,
    isChartLineLike,
    mergedVlConfig,
    sanitizeFieldName,
  } from "./util";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let renderer: string;
  export let canvasName: string;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;

  $: store = getCanvasStore(canvasName);
  $: ({
    canvasEntity: {
      spec: { getMeasureForMetricView },
    },
  } = store);

  let viewVL: View;

  $: chartConfig = rendererProperties as unknown as ChartSpec;
  $: chartType = renderer as ChartType;

  $: schema = validateChartSchema(store, chartConfig);

  $: data = getChartData(store, chartConfig, timeAndFilterStore);
  $: hasNoData = !$data.isFetching && $data.data.length === 0;

  $: spec = generateSpec(chartType, chartConfig, $data);

  $: componentFilters = getComponentFilterProperties(rendererProperties);

  $: measure = getMeasureForMetricView(
    chartConfig.y?.field,
    chartConfig.metrics_view,
  );

  $: measureName = sanitizeFieldName($measure?.name || "measure");

  $: measureFormatter = createMeasureValueFormatter<null | undefined>(
    $measure as MetricsViewSpecMeasureV2,
  );

  $: config = chartConfig.vl_config
    ? mergedVlConfig(chartConfig.vl_config)
    : undefined;

  $: title = chartConfig?.title || getChartTitle(chartConfig, $data);
  $: description = chartConfig?.description;
</script>

<div class="size-full flex flex-col overflow-hidden">
  {#if $schema.isValid}
    {#if $data.isFetching}
      <div class="flex items-center justify-center h-full w-full">
        <Spinner status={EntityStatus.Running} size="20px" />
      </div>
    {:else if $data.error}
      <ComponentError error={$data.error.message} />
    {:else}
      <ComponentHeader
        faint={!chartConfig?.title}
        {title}
        {description}
        filters={componentFilters}
      />
      {#if hasNoData}
        <div
          class="flex w-full h-full p-2 text-xl ui-copy-disabled items-center justify-center"
        >
          No Data to Display
        </div>
      {:else}
        <VegaLiteRenderer
          bind:viewVL
          canvasDashboard
          data={{ "metrics-view": $data.data }}
          {spec}
          renderer={isChartLineLike(chartType) ? "svg" : "canvas"}
          expressionFunctions={{
            [measureName]: { fn: (val) => measureFormatter(val) },
          }}
          {config}
        />
      {/if}
    {/if}
  {:else}
    <ComponentError error={$schema.error} />
  {/if}
</div>
