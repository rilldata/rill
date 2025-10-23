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

  export let component: BaseChart<CanvasChartSpec>;

  // Theme mode (light/dark) - separate from which theme is selected
  $: currentThemeMode = $themeControl;
  $: isThemeModeDark = currentThemeMode === "dark";

  // Create a reactive store from theme mode for chart data dependency
  $: themeModeStore = derived([], () => isThemeModeDark);

  $: ({ instanceId } = $runtime);

  $: ({
    specStore,
    parent: { name: canvasName, themeSpec },
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

  // Get the appropriate theme object (light or dark) based on current mode
  // This includes all CSS variables defined in the theme
  $: currentTheme = (() => {
    const spec = $themeSpec;

    // Get the ThemeColors object for current mode (has primary, secondary, and variables)
    const modeTheme = isThemeModeDark
      ? (spec?.dark as
          | {
              primary?: string;
              secondary?: string;
              variables?: Record<string, string>;
            }
          | undefined)
      : (spec?.light as
          | {
              primary?: string;
              secondary?: string;
              variables?: Record<string, string>;
            }
          | undefined);

    if (modeTheme) {
      // Merge primary, secondary, and all variables into a flat object
      const merged: Record<string, string> = { ...modeTheme.variables };
      if (modeTheme.primary) merged.primary = modeTheme.primary;
      if (modeTheme.secondary) merged.secondary = modeTheme.secondary;
      return merged;
    }

    // For legacy themes, construct a theme object with just primary/secondary
    if (spec?.primaryColorRaw || spec?.secondaryColorRaw) {
      const legacyTheme: Record<string, string> = {};
      if (spec.primaryColorRaw) legacyTheme.primary = spec.primaryColorRaw;
      if (spec.secondaryColorRaw)
        legacyTheme.secondary = spec.secondaryColorRaw;
      return legacyTheme;
    }

    return undefined;
  })();

  $: chartData = getChartDataForCanvas(
    store,
    component,
    chartSpec,
    timeAndFilterStore,
    themeModeStore,
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
        themeMode={isThemeModeDark ? "dark" : "light"}
        theme={currentTheme}
      />
    {/if}
  {:else}
    <ComponentError error={schema.error} />
  {/if}
</div>
