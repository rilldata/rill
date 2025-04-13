<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
  import type { CartesianChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/CartesianChart";
  import type { CircularChartSpec } from "@rilldata/web-common/features/canvas/components/charts/circular-charts/CircularChart";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import type { View } from "vega-typings";
  import { getChartData, validateChartSchema } from "./selector";
  import {
    generateSpec,
    getChartTitle,
    isChartLineLike,
    mergedVlConfig,
    sanitizeFieldName,
  } from "./util";

  export let component: BaseChart<CartesianChartSpec | CircularChartSpec>;

  $: ({
    specStore,
    parent: { name: canvasName },
    timeAndFilterStore,
    chartType: type,
  } = component);

  $: chartType = $type;

  $: store = getCanvasStore(canvasName);
  $: ({
    canvasEntity: {
      spec: { getMeasureForMetricView },
    },
  } = store);

  let viewVL: View;

  $: chartSpec = $specStore;

  $: ({
    title,
    description,
    metrics_view,
    vl_config,
    time_filters,
    dimension_filters,
  } = chartSpec);

  $: schemaStore = validateChartSchema(store, chartSpec);

  $: schema = $schemaStore;

  $: chartQuery = getChartData(store, chartSpec, timeAndFilterStore);

  $: ({ isFetching, data, error } = $chartQuery);
  $: hasNoData = !isFetching && data.length === 0;

  $: spec = generateSpec(chartType, chartSpec, $chartQuery);

  $: filters = {
    time_filters,
    dimension_filters,
  };

  $: measure = getMeasureForMetricView(chartSpec.y?.field, metrics_view);

  $: measureName = sanitizeFieldName($measure?.name || "measure");

  $: measureFormatter = createMeasureValueFormatter<null | undefined>(
    $measure as MetricsViewSpecMeasure,
  );

  $: config = vl_config ? mergedVlConfig(vl_config) : undefined;
</script>

<div class="size-full flex flex-col overflow-hidden">
  {#if schema.isValid}
    {#if isFetching}
      <div class="flex items-center justify-center h-full w-full">
        <Spinner status={EntityStatus.Running} size="20px" />
      </div>
    {:else if error}
      <ComponentError error={error.message} />
    {:else}
      <ComponentHeader
        faint={!title}
        title={title || getChartTitle(chartSpec, $chartQuery)}
        {description}
        {filters}
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
          data={{ "metrics-view": data }}
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
    <ComponentError error={schema.error} />
  {/if}
</div>
