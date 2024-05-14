<script lang="ts" context="module">
  import { writable } from "svelte/store";

  const tracking = (() => {
    const { subscribe, set } = writable<number | null>(null);
    let cursorScaler = scaleLinear().domain([0, 320]).range([0, 320]);
    const writableCursorScaler = writable(cursorScaler);

    return {
      subscribe,
      track: (e: MouseEvent) => set(e.offsetX),
      set: (x: number) => set(x),
      reset: () => set(null),
      updateScaler: (width: number, length: number) => {
        cursorScaler = cursorScaler.domain([0, width]).range([0, length]);
        writableCursorScaler.set(cursorScaler);
      },
      cursorScaler: writableCursorScaler,
    };
  })();

  export type Point = { x: number; y: number; ts: Date };
  import type { Vector } from "@rilldata/web-common/features/custom-dashboards/types";

  export type TimeSeriesLine = {
    data: Point[];
    type: "comparison" | "primary" | "secondary";
    color?: string;
    key: string;
  };

  const scrubStart = writable<number | null>(null);
  const scrubEnd = writable<number | null>(null);

  const buffer = 1.15;

  const timeFormat: Record<V1TimeGrain, Intl.DateTimeFormatOptions> = {
    [V1TimeGrain.TIME_GRAIN_YEAR]: { year: "numeric" },
    [V1TimeGrain.TIME_GRAIN_QUARTER]: { year: "numeric", month: "short" },
    [V1TimeGrain.TIME_GRAIN_MONTH]: { year: "numeric", month: "short" },
    [V1TimeGrain.TIME_GRAIN_WEEK]: DateTime.DATE_FULL,
    [V1TimeGrain.TIME_GRAIN_DAY]: DateTime.DATETIME_FULL,
    [V1TimeGrain.TIME_GRAIN_HOUR]: DateTime.DATETIME_FULL,
    [V1TimeGrain.TIME_GRAIN_MINUTE]: DateTime.DATETIME_FULL,
    [V1TimeGrain.TIME_GRAIN_SECOND]: DateTime.DATETIME_FULL_WITH_SECONDS,
    [V1TimeGrain.TIME_GRAIN_MILLISECOND]: {
      ...DateTime.DATETIME_FULL_WITH_SECONDS,
      fractionalSecondDigits: 2,
    },
    [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: DateTime.DATETIME_FULL,
  };

  const tweenParams = {
    duration: 500,
    easing: cubicOut,
  };
</script>

<script lang="ts">
  import { extent } from "d3-array";

  import { scaleLinear } from "d3-scale";
  import { onDestroy, onMount } from "svelte";

  import { tweened } from "svelte/motion";
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { SingleDigitTimesPowerOfTenFormatter } from "@rilldata/web-common/lib/number-formatting/strategies/SingleDigitTimesPowerOfTen";
  import { formatMsInterval } from "@rilldata/web-common/lib/number-formatting/strategies/intervals";
  import { cubicOut } from "svelte/easing";
  import { DateTime } from "luxon";

  import {
    MetricsViewSpecMeasureV2,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import Chart from "./Chart.svelte";
  import { LINE_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import DataBoundary from "@rilldata/web-common/features/dashboards/time-series/DataBoundary.svelte";
  import { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";

  //   const observer = new ResizeObserver((entries) => {
  //     for (let entry of entries) {
  //       clientWidth = entry.contentRect.width;
  //       clientHeight = entry.contentRect.height;
  //       tracking.updateScaler(clientWidth, referenceLine.length);
  //     }
  //   });
  export let timeSeries: Record<string, () => Promise<TimeSeriesDatum[]>>;
  export let dim: Record<string, () => Promise<DimensionDataItem[]>>;
  //   export let lines: TimeSeriesLine[];
  export let scrubRange: { start: number; end: number } | null;
  //   export let numberKind: NumberKind;
  export let timeZone: string;
  export let timeGrain: V1TimeGrain;
  export let xTicks: number[] = [];
  export let measures: MetricsViewSpecMeasureV2[];
  export let timeStart: number;
  export let timeEnd: number;
  export let allTime: number;
  export let showComparison: boolean;

  //   const initYExtents = extent(lines.map((d) => d.data).flat(), ({ y }) => y);
  //   const initXExtents = extent(lines.map((d) => d.data).flat(), ({ x }) => x);

  const xMin = tweened<number>(timeStart - allTime, tweenParams);
  const xMax = tweened<number>(timeEnd - allTime, tweenParams);

  //   const yMin = tweened<number>(initYExtents[0], tweenParams);
  //   const yMax = tweened<number>(initYExtents[1], tweenParams);

  let element: SVGSVGElement;
  let clientWidth: number;
  //   let clientHeight: number;
  let scrubbing: "start" | "end" | null = null;
  let grabStart: number | null = null;
  let grabbing = false;
  let scrubGrabStart: number | null = null;
  let scrubGrabEnd: number | null = null;
  //   let delayedLines = lines;

  //   onMount(() => {
  //     observer.observe(element);
  //   });

  //   onDestroy(() => {
  //     observer.disconnect();
  //   });

  $: showPoints =
    timeGrain === V1TimeGrain.TIME_GRAIN_HOUR ||
    timeGrain === V1TimeGrain.TIME_GRAIN_MINUTE;

  //   $: referenceLine = lines[0].data;

  //   $: tracking.updateScaler(clientWidth, referenceLine.length);

  $: xExtents = [$xMin, $xMax] as Vector;
  //   $: yExtents = [Math.min(0, $yMin * buffer), $yMax * buffer] as Vector;

  //   $: yScaler = scaleLinear().domain(yExtents).range([0, clientHeight]);
  //   $: xScaler = scaleLinear().domain(xExtents).range([0, clientWidth]);

  //   $: indexScaler = (number: number) =>
  //     Math.floor(
  //       scaleLinear()
  //         .domain([0, clientWidth])
  //         .range([0, referenceLine.length - 1])(number),
  //     );

  $: xMin.set(timeStart - allTime).catch(console.error);

  $: xMax.set(timeEnd - allTime).catch(console.error);

  $: {
    // const [newYMin = 0, newYMax = 0] = extent(
    //   lines.map((d) => d.data).flat(),
    //   (d) => d.y,
    // );
    // const [newXMin = 0, newXMax = 0] = extent(referenceLine, (d) => d.x);
    // if (newYMax >= get(yMax)) {
    //   yMin.set(newYMin, { duration: newYMin === 0 ? 0 : 400 });
    //   yMax.set(newYMax, { duration: newYMin === 0 ? 0 : 400 });
    //   setTimeout(() => {
    //     delayedLines = lines;
    //   }, 500);
    // } else {
    //   setTimeout(() => {
    // yMin.set(newYMin, { duration: !newYMin ? 0 : 400 }).catch(console.error);
    // yMax.set(newYMax, { duration: !newYMin ? 0 : 400 }).catch(console.error);
    //   }, 600);
    // }
    // setTimeout(() => {
    //   delayedLines = lines;
    // }, 500);
  }

  $: if (
    grabbing &&
    grabStart !== null &&
    scrubGrabStart !== null &&
    scrubGrabEnd !== null &&
    $tracking !== null
  ) {
    scrubStart.set(scrubGrabStart + $tracking - grabStart);
    scrubEnd.set(scrubGrabEnd + $tracking - grabStart);
  }

  //   $: if (scrubbing === "start") {
  //     scrubStart.set(referenceLine[$tracking].x);
  //   } else if (scrubbing === "end") {
  //     scrubEnd.set(referenceLine[($tracking ?? 0) + 1].x);
  //   }

  //   $: yTicks = yScaler.ticks(2);

  //   $: formatterFunction = (x: number) =>
  //     numberKind === NumberKind.INTERVAL
  //       ? formatMsInterval(x)
  //       : new SingleDigitTimesPowerOfTenFormatter(yTicks, {
  //           numberKind,
  //           padWithInsignificantZeros: false,
  //         }).stringFormat(x);

  //   $: hoverDate = hoverPoint?.ts
  //     ? DateTime.fromJSDate(hoverPoint.ts)
  //         .setZone(timeZone)
  //         .toLocaleString(timeFormat[timeGrain])
  //     : null;

  //   $: hoverValue = hoverPoint?.y ?? null;

  //   $: hoverPoint = $tracking !== null ? referenceLine[$tracking] : null;

  //   $: yRange = yExtents[1] - yExtents[0];

  //   $: xTicks = getTicks(delayedLines[0].data);

  //   $: xRange = xExtents[1] - xExtents[0];

  function getTicks(array: Point[]) {
    const ticks: number[] = [];

    const tickSpacing = Math.ceil(array.length / 8);

    for (let i = 0; i <= array.length; i += tickSpacing) {
      if (array[i]?.x) ticks.push(array[i]?.x);
    }

    return ticks;
  }

  function handleMouseDown() {
    scrubStart.set(referenceLine[$tracking].x);
    scrubbing = "end";

    window.addEventListener("mouseup", handleMouseUp);
  }

  function handleMouseUp() {
    scrubbing = null;
    if (scrubRange?.end === scrubRange?.start) {
      // currentStore.scrubRange.clear();
      //   dispatch("clear");
    }

    window.removeEventListener("mouseup", handleMouseUp);
  }

  function handleScrubGrab() {
    grabbing = true;
    grabStart = $tracking;
    scrubGrabStart = $scrubStart;
    scrubGrabEnd = $scrubEnd;

    window.addEventListener("mouseup", () => {
      grabbing = false;
      grabStart = null;
    });
  }
</script>

<div>
  {#each measures as { name: measureName } (measureName)}
    {#if measureName}
      <DataBoundary promise={timeSeries[measureName]} defaultData={[]} let:data>
        <DataBoundary
          promise={dim[measureName]}
          defaultData={[]}
          let:data={dimensionData}
        >
          <Chart
            {xExtents}
            {timeZone}
            {showPoints}
            numberKind={NumberKind.INTERVAL}
            {timeGrain}
            scrubRange={{ start: $scrubStart ?? 0, end: $scrubEnd ?? 0 }}
            lines={showComparison && dimensionData && dimensionData.length > 0
              ? dimensionData.map((d, i) => ({
                  data: d.data.slice(1),
                  key: d.value,
                  type: "comparison",
                  color: LINE_COLORS[i],
                }))
              : [
                  {
                    data: data.slice(1),
                    key: measureName,
                    type: "primary",
                  },
                ]}
          />
        </DataBoundary>
      </DataBoundary>
    {/if}
  {/each}
</div>

<style>
  text {
    stroke-linejoin: round;
    fill: black;
    stroke: white;
    stroke-width: 2.5px;
    paint-order: stroke;
  }

  .hover {
    font: 10px;
    font-weight: 500;
    text-anchor: middle;
    transform: scale(1, -1);
  }
</style>
