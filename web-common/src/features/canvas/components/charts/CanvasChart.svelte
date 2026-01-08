<script lang="ts">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { Chart } from "@rilldata/web-common/features/components/charts";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { derived } from "svelte/store";
  import type { CanvasChartSpec } from ".";
  import type { BaseChart } from "./BaseChart";
  import { getChartDataForCanvas } from "./selector";
  import { validateChartSchema } from "./validate";
  import {
    getComponentThemeOverrides,
    mergeComponentTheme,
  } from "../../utils/component-colors";
  import {
    COMPONENT_PATH_ROW_INDEX,
    COMPONENT_PATH_COLUMN_INDEX,
  } from "../../stores/canvas-entity";

  export let component: BaseChart<CanvasChartSpec>;

  // Theme mode (light/dark) - separate from which theme is selected
  $: isThemeModeDark = derived(
    themeControl,
    ($themeControl) => $themeControl === "dark",
  );

  $: ({ instanceId } = $runtime);

  $: ({
    specStore,
    parent: { name: canvasName, theme },
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

  // Get component item to access theme overrides
  $: specStore = component.parent?.specStore;
  $: canvasData = specStore ? $specStore?.data : undefined;
  $: canvasRows = canvasData?.canvas?.rows ?? [];
  $: rowIndex = (() => {
    const idx = component.pathInYAML?.[COMPONENT_PATH_ROW_INDEX];
    return typeof idx === "number" && idx >= 0 ? idx : -1;
  })();
  $: columnIndex = (() => {
    const idx = component.pathInYAML?.[COMPONENT_PATH_COLUMN_INDEX];
    return typeof idx === "number" && idx >= 0 ? idx : -1;
  })();
  $: item = (() => {
    if (
      rowIndex < 0 ||
      columnIndex < 0 ||
      !canvasRows ||
      rowIndex >= canvasRows.length
    ) {
      return undefined;
    }
    const row = canvasRows[rowIndex];
    if (!row?.items || columnIndex >= row.items.length) {
      return undefined;
    }
    return row.items[columnIndex];
  })();

  // Get component theme overrides and merge with global theme
  $: themeOverrides = getComponentThemeOverrides(item);
  $: mergedTheme = mergeComponentTheme(
    $theme?.spec,
    themeOverrides,
    $isThemeModeDark,
  );
</script>

<div
  class="size-full flex flex-col overflow-hidden"
  style="background-color: var(--card);"
>
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
        showDescriptionAsTooltip={show_description_as_tooltip}
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
        theme={mergedTheme || currentTheme}
      />
    {/if}
  {:else}
    <ComponentError error={schema.error} />
  {/if}
</div>
