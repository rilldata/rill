<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { Color } from "chroma-js";
  import type { Readable } from "svelte/store";
  import { readable } from "svelte/store";
  import Chart from "./Chart.svelte";
  import { CHART_CONFIG } from "./config";
  import { getChartData } from "./data-provider";
  import type { ChartProvider, ChartSpec, ChartType } from "./types";

  export let chartType: ChartType;
  export let spec: Readable<ChartSpec>;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let theme: "light" | "dark" = "light";
  export let themeStore: Readable<{ primary?: Color; secondary?: Color }> =
    readable({});

  let chartProvider: ChartProvider;
  $: {
    const chartConfig = CHART_CONFIG[chartType];
    chartProvider = new chartConfig.provider(spec, {});
  }

  // Create metrics view selectors from runtime
  $: metricsViewSelectors = new MetricsViewSelectors($runtime.instanceId);

  // Get measures from the metrics view specified in the chart spec
  $: measures = metricsViewSelectors.getMeasuresForMetricView(
    $spec.metrics_view,
  );

  // Create the chart data query using the provider
  $: chartDataQuery = chartProvider.createChartDataQuery(
    runtime,
    timeAndFilterStore,
  );

  // Create the chart data store
  $: chartData = getChartData({
    config: $spec,
    chartDataQuery,
    metricsView: metricsViewSelectors,
    themeStore,
    timeAndFilterStore,
    getDomainValues: () => chartProvider.getChartDomainValues($measures),
  });
</script>

{#if $spec}
  <Chart
    {chartType}
    chartSpec={$spec}
    {chartData}
    measures={$measures}
    {theme}
    isCanvas={false}
  />
{/if}
