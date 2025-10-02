<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
  import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { Color } from "chroma-js";
  import type { Readable, Writable } from "svelte/store";
  import { derived } from "svelte/store";
  import Chart from "./Chart.svelte";
  import { getChartData } from "./data-provider";
  import type { ChartProvider, ChartType } from "./types";

  // Required props
  export let chartType: ChartType;
  export let chartProvider: ChartProvider;
  export let runtime: Writable<Runtime>;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;

  // Optional props
  export let theme: "light" | "dark" = "light";
  export let themeStore: Readable<{ primary?: Color; secondary?: Color }> =
    derived([], () => ({}));

  // Get the chart spec from the provider
  // All providers have a spec property that's a Readable
  $: spec = (chartProvider as any).spec;

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

  // Create dependencies for getChartData
  $: deps = {
    config: $spec,
    chartDataQuery,
    metricsView: metricsViewSelectors,
    themeStore,
    timeAndFilterStore,
    getDomainValues: () => chartProvider.getChartDomainValues($measures),
  };

  // Create the chart data store
  $: chartData = getChartData(deps);
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
