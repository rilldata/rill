<script lang="ts">
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import { toPng } from "html-to-image";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./measure-chart/MeasureChart.svelte";
  import MeasureChartXAxis from "./measure-chart/MeasureChartXAxis.svelte";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { activeDashboardTheme } from "@rilldata/web-common/features/themes/active-dashboard-theme.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

  let {
    open = $bindable(false),
    measure,
    ephemeralMeasures = undefined,
    metricsViewName,
    expressionFilterManager,
    timeFilterManager,
    tddChartType = TDDChart.DEFAULT,
    timeDimension = undefined,
    comparisonDimension = undefined,
    dimensionValues = [],
    showTimeDimensionDetail = false,
    connectNulls = true,
    dynamicYAxis = false,
  }: {
    open: boolean;
    measure: MetricsViewSpecMeasure;
    ephemeralMeasures: EphemeralMeasureDef[] | undefined;
    metricsViewName: string;
    expressionFilterManager: ExpressionFilterManager;
    timeFilterManager: TimeFilterManager;
    tddChartType?: TDDChart;
    timeDimension?: string | undefined;
    comparisonDimension?: string | undefined;
    dimensionValues?: (string | null)[];
    showTimeDimensionDetail?: boolean;
    connectNulls?: boolean;
    dynamicYAxis?: boolean;
  } = $props();

  let where = $derived(
    expressionFilterManager.exprByMetricsView[metricsViewName],
  );
  let {
    timeStart,
    timeEnd,
    interval,
    timeGrain,

    comparisonTimeStart,
    comparisonTimeEnd,
    comparisonInterval,
    showComparison,

    ready,
  } = $derived(timeFilterManager);

  let captureNode: HTMLDivElement;
  let downloading = $state<boolean>(false);

  let formattedTimeRange = $derived(
    interval ? prettyFormatTimeRange(interval, timeGrain) : "",
  );
  let formattedComparisonRange = $derived(
    comparisonInterval
      ? prettyFormatTimeRange(comparisonInterval, timeGrain)
      : "",
  );
  let generatedTime = $derived(new Date().toISOString());

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
            {#if formattedComparisonRange}{m.kpi_vs_comparison({
                comparison: formattedComparisonRange,
              })}{/if}
          </div>
        </header>

        <ReadonlyExpressionFilters {expressionFilterManager} />

        <div class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2">
          {#if timeGrain}
            <div class="col-span-2 grid grid-cols-subgrid">
              <div></div>
              <MeasureChartXAxis {interval} timeGranularity={timeGrain} />
            </div>
          {/if}

          <MeasureBigNumber
            {measure}
            {ephemeralMeasures}
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

          {#if timeDimension}
            <MeasureChart
              {measure}
              {ephemeralMeasures}
              {expressionFilterManager}
              {timeFilterManager}
              {connectNulls}
              tddChartType={tddChartType ?? TDDChart.DEFAULT}
              {metricsViewName}
              {timeDimension}
              {ready}
              {comparisonDimension}
              {dimensionValues}
              {showTimeDimensionDetail}
              {dynamicYAxis}
            />
          {/if}
        </div>

        <footer class="flex items-center justify-between text-xs text-fg-muted">
          <!-- i18n-ignore: standalone product name -->
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
