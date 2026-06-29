<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import type {
    MetricsViewSpecMeasure,
    V1Expression,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { toPng } from "html-to-image";
  import { Interval } from "luxon";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./measure-chart/MeasureChart.svelte";
  import MeasureChartXAxis from "./measure-chart/MeasureChartXAxis.svelte";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
  import ExploreFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/ExploreFilterChipsReadOnly.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { activeDashboardTheme } from "@rilldata/web-common/features/themes/active-dashboard-theme.ts";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let open = false;
  export let measure: MetricsViewSpecMeasure;
  export let metricsViewName: string;
  export let where: V1Expression | undefined = undefined;
  export let tddChartType: TDDChart = TDDChart.DEFAULT;
  export let timeDimension: string | undefined = undefined;
  export let timeStart: string | undefined = undefined;
  export let timeEnd: string | undefined = undefined;
  export let comparisonTimeStart: string | undefined = undefined;
  export let comparisonTimeEnd: string | undefined = undefined;
  export let interval: Interval<true> | undefined = undefined;
  export let comparisonInterval: Interval<true> | undefined = undefined;
  export let timeGranularity: V1TimeGrain | undefined = undefined;
  export let timeZone: string = "UTC";
  export let comparisonDimension: string | undefined = undefined;
  export let dimensionWhere: V1Expression | undefined = undefined;
  export let dimensionValues: (string | null)[] = [];
  export let showComparison = false;
  export let showTimeDimensionDetail: boolean = false;
  export let connectNulls: boolean = true;
  export let dynamicYAxis: boolean = false;
  export let ready = true;

  let captureNode: HTMLDivElement;
  let downloading = false;

  $: formattedTimeRange = interval
    ? prettyFormatTimeRange(interval, timeGranularity)
    : "";
  $: formattedComparisonRange = comparisonInterval
    ? prettyFormatTimeRange(comparisonInterval, timeGranularity)
    : "";
  $: generatedTime = new Date().toISOString();

  const SVG_PROPS = [
    "fill",
    "fill-opacity",
    "stroke",
    "stroke-width",
    "stroke-opacity",
    "stroke-dasharray",
    "stroke-linecap",
    "opacity",
    "font-family",
    "font-size",
    "font-weight",
    "color",
  ];

  function inlineSvgStyles(root: HTMLElement) {
    root.querySelectorAll("svg, svg *").forEach((el) => {
      const cs = getComputedStyle(el);
      const inline = SVG_PROPS.map(
        (p) => `${p}: ${cs.getPropertyValue(p)}`,
      ).join("; ");
      el.setAttribute("style", `${inline}; ${el.getAttribute("style") ?? ""}`);
    });
  }

  async function downloadScreenshot() {
    if (!captureNode) return;
    downloading = true;
    try {
      inlineSvgStyles(captureNode);
      await document.fonts.ready;
      const url = await toPng(captureNode, { cacheBust: true });
      const link = document.createElement("a");
      link.download = `${measure.name ?? "chart"}_${formattedTimeRange || generatedTime}.png`;
      link.href = url;
      link.click();
    } finally {
      downloading = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-3xl flex flex-col gap-y-4">
    <Dialog.Header>
      <Dialog.Title>{m.dashboard_export_chart()}</Dialog.Title>
    </Dialog.Header>

    <ThemeProvider theme={$activeDashboardTheme} applyLayout={false}>
      <div
        bind:this={captureNode}
        class="flex flex-col gap-y-3 p-4 bg-surface-background border rounded-md"
      >
        <header class="flex flex-row gap-y-0.5">
          <div class="flex flex-col">
            <h2 class="text-base font-semibold text-fg-base">
              {measure.displayName || measure.name}
            </h2>
            {#if measure.description}
              <p class="text-xs text-fg-muted">{measure.description}</p>
            {/if}
          </div>
          <div class="grow"></div>
          <div>
            {formattedTimeRange}
            {#if formattedComparisonRange}{m.kpi_vs_comparison({ comparison: formattedComparisonRange })}{/if}
          </div>
        </header>

        <ExploreFilterChipsReadOnly
          metricsViewNames={[metricsViewName]}
          filters={where}
          dimensionsWithInlistFilter={[]}
          dimensionThresholdFilters={[]}
        />

        <div class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2">
          {#if timeGranularity}
            <div class="col-span-2 grid grid-cols-subgrid">
              <div></div>
              <MeasureChartXAxis {interval} {timeGranularity} />
            </div>
          {/if}

          <MeasureBigNumber
            {measure}
            {metricsViewName}
            {where}
            {timeDimension}
            {timeStart}
            {timeEnd}
            {comparisonTimeStart}
            {comparisonTimeEnd}
            {showComparison}
            {ready}
            skipLink
          />

          <MeasureChart
            {measure}
            {connectNulls}
            tddChartType={tddChartType ?? TDDChart.DEFAULT}
            {metricsViewName}
            {where}
            {timeDimension}
            {interval}
            {comparisonInterval}
            {timeGranularity}
            {timeZone}
            {ready}
            {comparisonDimension}
            {dimensionValues}
            {dimensionWhere}
            {showComparison}
            {showTimeDimensionDetail}
            {dynamicYAxis}
          />
        </div>

        <footer class="flex items-center justify-between text-xs text-fg-muted">
          <span>Rill</span>
          <span>{m.dashboard_generated({ time: generatedTime })}</span>
        </footer>
      </div>
    </ThemeProvider>

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (open = false)}
        >{m.dashboard_cancel()}</Button
      >
      <Button
        type="primary"
        disabled={downloading}
        onClick={downloadScreenshot}
      >
        {downloading ? m.dashboard_generating() : m.dashboard_download_png()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
