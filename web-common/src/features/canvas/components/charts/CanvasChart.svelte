<script lang="ts">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { Chart } from "@rilldata/web-common/features/components/charts";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    requestExactNames,
    withExactVariant,
  } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
  import type { View } from "svelte-vega";
  import { derived } from "svelte/store";
  import type { CanvasChartSpec } from ".";
  import { CANVAS_CHART_CONFIG } from ".";
  import {
    canvasChartPrecisionChip,
    canvasExactMeasureContext,
    chartMeasureNames,
  } from "../exact-measures";
  import type { BaseChart } from "./BaseChart";
  import { getChartDataForCanvas } from "./selector";
  import { validateChartSchema } from "./validate";

  export let component: BaseChart<CanvasChartSpec>;

  const runtimeClient = useRuntimeClient();
  let viewVL: View;

  // Theme mode (light/dark) - separate from which theme is selected
  $: isThemeModeDark = derived(
    themeControl,
    ($themeControl) => $themeControl === "dark",
  );

  $: ({ instanceId } = runtimeClient);

  $: ({
    specStore,
    parent: { name: canvasName, theme },
    timeAndFilterStore,
    chartType: type,
    dataEnabled: visible,
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

  $: ({
    title,
    description,
    show_description_as_tooltip,
    metrics_view,
    time_filters,
    dimension_filters,
  } = chartSpec);

  $: schemaStore = validateChartSchema(metricsView, chartSpec);

  $: schema = $schemaStore;

  $: measures = getMeasuresForMetricView(metrics_view);

  $: currentTheme =
    $theme?.resolvedThemeObject?.[$isThemeModeDark ? "dark" : "light"];

  // Exact measures: the chart's own selection (the chart is its own surface).
  // The query itself reads the set from the component's query spec store;
  // passing the same spec here makes the data re-derive when the selection
  // changes.
  $: measureNames = chartMeasureNames(
    CANVAS_CHART_CONFIG[chartType]?.component.chartInputParams,
    chartSpec,
  );
  $: exactContextStore = canvasExactMeasureContext(
    store.canvasEntity,
    component.id,
    metrics_view,
  );
  $: exactRequestNames = [
    ...requestExactNames(measureNames, $exactContextStore),
  ];
  $: chartSpecForData = exactRequestNames.length
    ? { ...chartSpec, exact_measures: exactRequestNames }
    : chartSpec;

  $: chartData = getChartDataForCanvas(
    store,
    component,
    chartSpecForData,
    timeAndFilterStore,
    isThemeModeDark,
    visible,
  );

  $: ({ isFetching, error } = $chartData);

  $: chartTitle = title || component.chartTitle($chartData.fields);

  // One chip for the whole chart, right of the title: it switches every
  // measure with an exact variant at once.
  $: precisionChip = canvasChartPrecisionChip(
    measureNames,
    chartTitle,
    $exactContextStore,
    isFetching,
    error?.message,
    () =>
      store.canvasEntity.toggleExactSurface(
        component.id,
        withExactVariant(measureNames, $exactContextStore),
      ),
  );

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
    {:else}
      <!-- The header stays when the query fails, so that a failed exact
           query can be switched back from its chip. -->
      <ComponentHeader
        faint={!title}
        title={chartTitle}
        {description}
        showDescriptionAsTooltip={show_description_as_tooltip}
        {filters}
        {component}
        {precisionChip}
      />
      {#if error}
        <ComponentError error={error.message} />
      {:else}
        <Chart
          {chartType}
          {chartSpec}
          {chartData}
          measures={$measures}
          isCanvas
          bind:view={viewVL}
          themeMode={$isThemeModeDark ? "dark" : "light"}
          theme={currentTheme}
        />
      {/if}
    {/if}
  {:else}
    <ComponentError error={schema.error} />
  {/if}
</div>
