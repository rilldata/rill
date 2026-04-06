<script lang="ts">
  import Chart from "@rilldata/web-common/features/components/charts/Chart.svelte";
  import { CHART_CONFIG } from "@rilldata/web-common/features/components/charts/config";
  import { getChartData } from "@rilldata/web-common/features/components/charts/data-provider";
  import {
    clearExternalHover,
    setExternalHover,
  } from "@rilldata/web-common/features/components/charts/highlight-controller";
  import type { ChartProvider } from "@rilldata/web-common/features/components/charts/types";
  import { THEME_STORE_CONTEXT_KEY } from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { tableInteractionStore } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
  import type { DimensionSeriesData } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";
  import { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
  import type { Theme } from "@rilldata/web-common/features/themes/theme";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import type {
    MetricsViewSpecMeasure,
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { Interval } from "luxon";
  import { getContext } from "svelte";
  import type { View } from "svelte-vega";
  import type { Readable, Writable } from "svelte/store";
  import { readable } from "svelte/store";
  import type { TDDChart } from "../types";
  import {
    createTDDCartesianSpec,
    TDD_TO_COMPONENT_CHART_TYPE,
  } from "./tdd-chart-config";

  export let metricsViewName: string;
  export let measure: MetricsViewSpecMeasure;
  export let timeDimension: string | undefined = undefined;
  export let interval: Interval<true> | undefined = undefined;
  export let comparisonInterval: Interval<true> | undefined = undefined;
  export let timeGranularity: V1TimeGrain | undefined = undefined;
  export let timeZone: string = "UTC";
  export let where: V1Expression | undefined = undefined;
  export let comparisonDimension: string | undefined = undefined;
  export let dimensionValues: (string | null)[] = [];
  export let dimensionData: DimensionSeriesData[] = [];
  export let showComparison: boolean = false;
  export let chartType: TDDChart;
  export let isScrubbing: boolean;
  export let onChartHover: (
    dimension: undefined | string | null,
    ts: Date | undefined,
  ) => void;
  export let onChartBrush: (interval: { start: Date; end: Date }) => void;
  export let onChartBrushEnd: (interval: { start: Date; end: Date }) => void;
  export let onChartBrushClear: () => void;

  const client = useRuntimeClient();
  const themeStore = getContext<Writable<Theme | undefined>>(
    THEME_STORE_CONTEXT_KEY,
  );

  let provider: ChartProvider;
  let chartView: View;

  $: ({ current } = themeControl);
  $: themeMode = $current;

  $: measureName = measure.name ?? "";

  // Build CartesianChartSpec reactively
  $: cartesianSpec = createTDDCartesianSpec(
    metricsViewName,
    measureName,
    timeDimension ?? "",
    comparisonDimension,
    dimensionValues,
    dimensionData,
  );

  $: componentChartType = TDD_TO_COMPONENT_CHART_TYPE[chartType];

  // Create a TimeAndFilterStore from MeasureChart's props
  $: timeAndFilterStore = createTimeAndFilterStore(
    interval,
    comparisonInterval,
    timeGranularity,
    timeZone,
    where,
    showComparison,
  );

  function createTimeAndFilterStore(
    int: Interval<true> | undefined,
    compInt: Interval<true> | undefined,
    grain: V1TimeGrain | undefined,
    timeZone: string,
    where: V1Expression | undefined,
    showComp: boolean,
  ): Readable<TimeAndFilterStore> {
    const timeRange: V1TimeRange = {
      start: int?.start?.toUTC().toISO() ?? undefined,
      end: int?.end?.toUTC().toISO() ?? undefined,
      timeZone: timeZone,
    };

    const comparisonTimeRange: V1TimeRange | undefined =
      showComp && compInt
        ? {
            start: compInt.start?.toUTC().toISO() ?? undefined,
            end: compInt.end?.toUTC().toISO() ?? undefined,
            timeZone: timeZone,
          }
        : undefined;

    return readable({
      timeRange,
      comparisonTimeRange,
      where,
      timeGrain: grain,
      showTimeComparison: showComp,
      timeRangeState: undefined,
      comparisonTimeRangeState: undefined,
      hasTimeSeries: true,
    });
  }

  $: {
    provider = new CHART_CONFIG[componentChartType].provider(
      readable(cartesianSpec),
      {},
    );
  }

  $: metricsViewSelectors = new MetricsViewSelectors(client);

  $: measures = metricsViewSelectors.getMeasuresForMetricView(metricsViewName);

  $: chartDataQuery = provider.createChartDataQuery(client, timeAndFilterStore);

  $: chartData = getChartData({
    config: cartesianSpec,
    chartDataQuery,
    metricsView: metricsViewSelectors,
    themeStore,
    timeAndFilterStore,
    getDomainValues: () => provider.getChartDomainValues($measures),
    isThemeModeDark: themeMode === "dark",
  });

  // Bidirectional highlighting: table hover → chart highlight
  $: hoveredTime = $tableInteractionStore.time;
  $: hoveredDimensionValue = $tableInteractionStore.dimensionValue;

  $: {
    if (chartView) {
      if (hoveredTime) {
        setExternalHover(chartView, hoveredTime, hoveredDimensionValue);
      } else {
        clearExternalHover(chartView);
      }
    }
  }
</script>

<Chart
  chartType={componentChartType}
  chartSpec={cartesianSpec}
  {chartData}
  measures={$measures}
  {themeMode}
  isCanvas={false}
  temporalField={timeDimension}
  {isScrubbing}
  onBrush={onChartBrush}
  onBrushEnd={onChartBrushEnd}
  onBrushClear={onChartBrushClear}
  onHover={onChartHover}
  bind:view={chartView}
/>
