<script lang="ts">
  import { metricsSQLUsesTimeRange } from "@rilldata/web-common/features/components/charts/custom/expression-utils";
  import ResolvedComponentChart from "@rilldata/web-common/features/custom-viz/ResolvedComponentChart.svelte";
  import { resolveChartTheme } from "@rilldata/web-common/features/themes/chart-theme";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ComponentRefComponent } from "./index";

  export let component: ComponentRefComponent;
  export let editable: boolean = false;

  $: ({ specStore, timeAndFilterStore, resource } = component);

  $: ({
    parent: {
      theme,
      timeManager: { hasTimeSeriesMap },
    },
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
  chartTheme={resolveChartTheme($theme?.spec, themeMode === "dark")}
/>
