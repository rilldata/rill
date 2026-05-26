<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import type {
    MetricsViewSpecMeasure,
    V1Expression,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { getFontEmbedCSS, toPng } from "html-to-image";
  import { Interval } from "luxon";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import MeasureChart from "./measure-chart/MeasureChart.svelte";
  import MeasureChartXAxis from "./measure-chart/MeasureChartXAxis.svelte";
  import { ScrubController } from "./measure-chart/ScrubController";

  export let open = false;
  export let measure: MetricsViewSpecMeasure;
  export let metricsViewName: string;
  export let where: V1Expression | undefined = undefined;
  export let timeDimension: string | undefined = undefined;
  export let timeStart: string | undefined = undefined;
  export let timeEnd: string | undefined = undefined;
  export let comparisonTimeStart: string | undefined = undefined;
  export let comparisonTimeEnd: string | undefined = undefined;
  export let interval: Interval<true> | undefined = undefined;
  export let comparisonInterval: Interval<true> | undefined = undefined;
  export let timeGranularity: V1TimeGrain | undefined = undefined;
  export let timeZone: string = "UTC";
  export let showComparison = false;
  export let ready = true;

  // Inert scrub controller — interactions are not needed for screenshots.
  const scrubController = new ScrubController();

  let captureNode: HTMLDivElement;
  let downloading = false;
  let url = "";

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
      const fontEmbedCSS = await getFontEmbedCSS(captureNode);
      url = await toPng(captureNode, { fontEmbedCSS, cacheBust: true });
      // const link = document.createElement("a");
      // link.download = `${measure.name ?? "chart"}.png`;
      // link.href = url;
      // link.click();
    } finally {
      downloading = false;
    }
  }
</script>

<Dialog.Root
  bind:open
  onOpenChange={() => {
    url = "";
  }}
>
  <Dialog.Content class="max-w-3xl flex flex-col gap-y-4">
    <Dialog.Header>
      <Dialog.Title>Export chart</Dialog.Title>
    </Dialog.Header>

    <div
      bind:this={captureNode}
      class="flex flex-col gap-y-3 p-4 bg-surface-background border rounded-md"
    >
      <header class="flex flex-col gap-y-0.5">
        <slot name="header">
          <h2 class="text-base font-semibold text-fg-base">
            {measure.displayName || measure.name}
          </h2>
          {#if measure.description}
            <p class="text-xs text-fg-muted">{measure.description}</p>
          {/if}
        </slot>
      </header>

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
        />

        {#if timeGranularity}
          <MeasureChart
            {measure}
            {scrubController}
            tddChartType={TDDChart.DEFAULT}
            {metricsViewName}
            {where}
            {timeDimension}
            {interval}
            {comparisonInterval}
            {timeGranularity}
            {timeZone}
            {ready}
            {showComparison}
            connectNulls={true}
          />
        {/if}
      </div>

      <footer class="flex items-center justify-between text-xs text-fg-muted">
        <slot name="footer">
          <span>Generated {new Date().toLocaleString()}</span>
          <span>Rill</span>
        </slot>
      </footer>
    </div>

    {#if url}
      <div>
        <img src={url} alt="Screenshot" class="w-full h-auto" />
      </div>
    {/if}

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (open = false)}>Cancel</Button>
      <Button
        type="primary"
        disabled={downloading}
        onClick={downloadScreenshot}
      >
        {downloading ? "Generating…" : "Download PNG"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
