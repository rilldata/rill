<script lang="ts">
  import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { View } from "vega-typings";
  import type { ChartSpec } from "./";
  import type { BaseChart } from "./BaseChart";
  import { getChartData } from "./selector";
  import { generateSpec } from "./util";
  import { validateChartSchema } from "./validate";

  export let component: BaseChart<ChartSpec>;

  $: themePreference = $themeControl;

  $: ({ instanceId } = $runtime);

  $: ({
    specStore,
    parent: { name: canvasName },
    timeAndFilterStore,
    chartType: type,
  } = component);

  $: chartType = $type;

  $: store = getCanvasStore(canvasName, instanceId);
  $: ({
    canvasEntity: {
      spec: { getMeasuresForMetricView },
    },
  } = store);

  let viewVL: View;

  $: chartSpec = $specStore;

  $: ({ title, description, metrics_view, time_filters, dimension_filters } =
    chartSpec);

  $: schemaStore = validateChartSchema(store, chartSpec);

  $: schema = $schemaStore;

  $: chartQuery = getChartData(store, component, chartSpec, timeAndFilterStore);

  $: ({ isFetching, data, error } = $chartQuery);
  $: hasNoData = !isFetching && data.length === 0;

  $: spec = generateSpec(chartType, chartSpec, $chartQuery);

  $: filters = {
    time_filters,
    dimension_filters,
  };

  $: measures = getMeasuresForMetricView(metrics_view);

  // TODO: Move this to a central cached store
  $: measureFormatters = $measures.reduce(
    (acc, measure) => ({
      ...acc,
      [sanitizeFieldName(measure.name || "measure")]:
        createMeasureValueFormatter<null | undefined>(measure),
    }),
    {},
  );

  $: expressionFunctions = $measures.reduce((acc, measure) => {
    const fieldName = sanitizeFieldName(measure.name || "measure");
    return {
      ...acc,
      [fieldName]: { fn: (val) => measureFormatters[fieldName](val) },
    };
  }, {});
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
        title={title || component.chartTitle($chartQuery?.fields)}
        {description}
        {filters}
        {component}
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
          renderer="canvas"
          {expressionFunctions}
          config={getRillTheme(true, themePreference === "dark")}
        />
      {/if}
    {/if}
  {:else}
    <ComponentError error={schema.error} />
  {/if}
</div>
