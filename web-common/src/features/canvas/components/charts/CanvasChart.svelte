<script lang="ts">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import Chart from "@rilldata/web-common/features/components/charts/Chart.svelte";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { ChartSpec } from ".";
  import type { BaseChart } from "./BaseChart";
  import { getChartDataForCanvas } from "./selector";
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
      metricsView,
      metricsView: { getMeasuresForMetricView },
    },
  } = store);

  $: chartSpec = $specStore;

  $: ({ title, description, metrics_view, time_filters, dimension_filters } =
    chartSpec);

  $: schemaStore = validateChartSchema(metricsView, chartSpec);

  $: schema = $schemaStore;

  $: measures = getMeasuresForMetricView(metrics_view);

  $: chartData = getChartDataForCanvas(
    store,
    component,
    chartSpec,
    timeAndFilterStore,
  );

  $: ({ isFetching, data, error } = $chartData);

  $: filters = {
    time_filters,
    dimension_filters,
  };
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
        title={title || component.chartTitle($chartData?.fields)}
        {description}
        {filters}
        {component}
      />
      <Chart
        {chartType}
        {chartSpec}
        {chartData}
        measures={$measures}
        isCanvas
        theme={themePreference === "dark" ? "dark" : "light"}
      />
    {/if}
  {:else}
    <ComponentError error={schema.error} />
  {/if}
</div>
