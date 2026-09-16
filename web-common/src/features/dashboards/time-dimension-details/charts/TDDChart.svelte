<script lang="ts">
  import Chart from "@rilldata/web-common/features/components/charts/Chart.svelte";
  import { CHART_CONFIG } from "@rilldata/web-common/features/components/charts/config";
  import { getChartData } from "@rilldata/web-common/features/components/charts/data-provider";
  import {
    clearExternalHover,
    setExternalHover,
  } from "@rilldata/web-common/features/components/charts/highlight-controller";
  import { THEME_STORE_CONTEXT_KEY } from "@rilldata/web-common/features/themes/theme-boundary";
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import {
    chartBrushStore,
    chartHoverStore,
  } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/hover-index";
  import type { DimensionSeriesData } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";
  import { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
  import type { Theme } from "@rilldata/web-common/features/themes/theme";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getContext } from "svelte";
  import type { View } from "svelte-vega";
  import type { Writable } from "svelte/store";
  import { readable } from "svelte/store";
  import type { TDDChart } from "../types";
  import {
    createTDDCartesianSpec,
    TDD_TO_COMPONENT_CHART_TYPE,
  } from "./tdd-chart-config";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

  let {
    metricsViewName,
    measure,
    ephemeralMeasures = undefined,
    expressionFilterManager,
    timeFilterManager,
    timeDimension = undefined,
    comparisonDimension = undefined,
    dimensionValues = [],
    dimensionData = [],
    showTimeDimensionDetail = true,
    chartType,
    dynamicYAxis = false,
    onChartHover,
    onChartBrushEnd,
    onChartBrushClear,
  }: {
    metricsViewName: string;
    measure: MetricsViewSpecMeasure;
    ephemeralMeasures?: EphemeralMeasureDef[] | undefined;
    expressionFilterManager: ExpressionFilterManager;
    timeFilterManager: TimeFilterManager;
    timeDimension?: string | undefined;
    comparisonDimension: string | undefined;
    dimensionValues?: (string | null)[];
    dimensionData?: DimensionSeriesData[];
    showTimeDimensionDetail?: boolean;
    chartType: TDDChart;
    dynamicYAxis?: boolean;
    onChartHover: (
      dimension: undefined | string | null,
      ts: Date | undefined,
    ) => void;
    onChartBrushEnd: (interval: { start: Date; end: Date }) => void;
    onChartBrushClear: () => void;
  } = $props();

  const client = $derived(useRuntimeClient());
  const themeStore = getContext<Writable<Theme | undefined>>(
    THEME_STORE_CONTEXT_KEY,
  );

  let chartView = $state<View>(undefined as unknown as View);

  let { current } = $derived(themeControl);
  let themeMode = $derived($current);

  let measureName = $derived(measure.name ?? "");

  // Build CartesianChartSpec reactively
  let cartesianSpec = $derived(
    createTDDCartesianSpec(
      metricsViewName,
      measureName,
      timeDimension ?? "",
      comparisonDimension,
      dimensionValues,
      dimensionData,
      showTimeDimensionDetail,
      dynamicYAxis,
      ephemeralMeasures,
    ),
  );

  let componentChartType = $derived(TDD_TO_COMPONENT_CHART_TYPE[chartType]);

  let filtersStore = $derived(
    expressionFilterManager.getExprStoreForMetricsView(metricsViewName),
  );
  let timeControlStore = $derived(timeFilterManager.getTimeControlStore());

  let provider = $derived(
    new CHART_CONFIG[componentChartType].provider(readable(cartesianSpec), {}),
  );

  let metricsViewSelectors = $derived(new MetricsViewSelectors(client));

  let measures = $derived(
    metricsViewSelectors.getMeasuresForMetricView(metricsViewName),
  );

  let chartDataQuery = $derived(
    provider.createChartDataQuery(client, filtersStore, timeControlStore),
  );

  let chartData = $derived(
    getChartData({
      config: cartesianSpec,
      chartDataQuery,
      metricsView: metricsViewSelectors,
      themeStore,
      timeControlStore,
      getDomainValues: () => provider.getChartDomainValues($measures),
      isThemeModeDark: themeMode === "dark",
    }),
  );

  // Bidirectional highlighting: table/chart hover → chart highlight
  let hoveredTime = $derived($chartHoverStore.time);
  let hoveredDimensionValue = $derived($chartHoverStore.dimensionValue);

  // Brush sync: apply brush from sibling charts
  let externalBrushStartMs = $derived($chartBrushStore.startMs);
  let externalBrushEndMs = $derived($chartBrushStore.endMs);

  $effect(() => {
    if (chartView) {
      if (hoveredTime) {
        setExternalHover(chartView, hoveredTime, hoveredDimensionValue);
      } else {
        clearExternalHover(chartView);
      }
    }
  });
</script>

<Chart
  chartType={componentChartType}
  chartSpec={cartesianSpec}
  {chartData}
  measures={$measures}
  {themeMode}
  isCanvas={false}
  temporalField={timeDimension}
  {externalBrushStartMs}
  {externalBrushEndMs}
  onBrushEnd={onChartBrushEnd}
  onBrushClear={onChartBrushClear}
  onHover={onChartHover}
  bind:view={chartView}
/>
