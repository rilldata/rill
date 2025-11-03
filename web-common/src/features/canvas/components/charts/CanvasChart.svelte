<script lang="ts">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { Chart } from "@rilldata/web-common/features/components/charts";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { resolveThemeObject } from "@rilldata/web-common/features/themes/theme-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { derived } from "svelte/store";
  import type { CanvasChartSpec } from ".";
  import type { BaseChart } from "./BaseChart";
  import { getChartDataForCanvas } from "./selector";
  import { validateChartSchema } from "./validate";
  import { getContext } from "svelte";
  import { Theme } from "@rilldata/web-common/features/themes/theme";

  export let component: BaseChart<CanvasChartSpec>;

  // Theme mode (light/dark) - separate from which theme is selected
  $: isThemeModeDark = derived(
    themeControl,
    ($themeControl) => $themeControl === "dark",
  );

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

  // $: currentTheme = resolveThemeObject($themeSpec, $isThemeModeDark);

  const theme = getContext<Theme | undefined>("themeContext");

  $: chartData = getChartDataForCanvas(
    store,
    component,
    chartSpec,
    timeAndFilterStore,
    isThemeModeDark,
  );

  $: ({ isFetching, error } = $chartData);

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
        themeMode={$isThemeModeDark ? "dark" : "light"}
        {theme}
      />
    {/if}
  {:else}
    <ComponentError error={schema.error} />
  {/if}
</div>
