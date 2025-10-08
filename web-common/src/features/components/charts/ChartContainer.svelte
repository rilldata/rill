<script lang="ts">
  import Chart from "@rilldata/web-common/features/components/charts/Chart.svelte";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { Color } from "chroma-js";
  import type { Readable } from "svelte/store";
  import { readable } from "svelte/store";
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
  $: console.log(chartType, $spec);
  $: {
    const chartConfig = CHART_CONFIG[chartType];
    chartProvider = new chartConfig.provider(spec, {});
  }

  $: metricsViewSelectors = new MetricsViewSelectors($runtime.instanceId);

  $: measures = metricsViewSelectors.getMeasuresForMetricView(
    $spec.metrics_view,
  );

  $: chartDataQuery = chartProvider.createChartDataQuery(
    runtime,
    timeAndFilterStore,
  );

  $: chartData = getChartData({
    config: $spec,
    chartDataQuery,
    metricsView: metricsViewSelectors,
    themeStore,
    timeAndFilterStore,
    getDomainValues: () => chartProvider.getChartDomainValues($measures),
    isDarkMode: theme === "dark",
  });
</script>

{#if $spec}
  <div class="size-full">
    <Chart
      {chartType}
      chartSpec={$spec}
      {chartData}
      measures={$measures}
      {theme}
      isCanvas={true}
    />
  </div>
{/if}
