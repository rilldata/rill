<script lang="ts">
  import { metricsSQLUsesTimeRange } from "@rilldata/web-common/features/components/charts/custom/expression-utils";
  import ResolvedComponentChart from "@rilldata/web-common/features/custom-viz/ResolvedComponentChart.svelte";
  import type { CustomChartExportResult } from "@rilldata/web-common/features/exports/xlsx/payload";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ComponentRefComponent } from "./index";

  export let component: ComponentRefComponent;
  export let editable: boolean = false;

  $: ({ specStore, timeAndFilterStore, resource } = component);

  // The component's own `text_size` is a renderer property of its file, not one of the
  // item's bindings in specStore, so the chart resolves it and takes the canvas's as the default.
  $: ({
    parent: { theme, hasTimeSeriesMap, textSize: canvasTextSize },
  } = component);

  // The component's metrics view is the one its item binds (see ComponentRefComponent.syncMetricsViewName).
  $: usesTimeRange = metricsSQLUsesTimeRange(
    [$specStore.metrics_view as string | undefined],
    $hasTimeSeriesMap,
  );

  // Theme mode (light/dark) is separate from which theme is selected.
  let themeMode: "light" | "dark";
  $: themeMode = $themeControl === "dark" ? "dark" : "light";
  $: currentTheme = $theme?.resolvedThemeObject?.[themeMode];

  $: componentName = $specStore.component;
  $: componentSpec =
    $resource?.component?.state?.validSpec ??
    (component.parent.allowUnvalidatedSpec
      ? $resource?.component?.spec
      : undefined);
</script>

<ResolvedComponentChart
  {componentName}
  {componentSpec}
  resolvable={!!componentSpec}
  args={component.args($specStore)}
  stateUpdatedOn={$resource?.meta?.stateUpdatedOn}
  missingError={m.canvas_component_ref_missing({ name: componentName })}
  metricsViewName={$specStore.metrics_view as string | undefined}
  name={component.id}
  timeGrain={$timeAndFilterStore?.timeGrain}
  whereFilter={$timeAndFilterStore?.where}
  timeRange={usesTimeRange ? $timeAndFilterStore?.timeRange : undefined}
  showDataTable={editable}
  {themeMode}
  theme={currentTheme}
  canvasTextSize={$canvasTextSize}
  onResults={(results: CustomChartExportResult[] | undefined) =>
    component.exportPayload.set(
      results ? { kind: "customChart", results } : undefined,
    )}
/>
