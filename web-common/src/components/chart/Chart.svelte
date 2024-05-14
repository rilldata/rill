<script lang="ts" context="module">
  import { writable } from "svelte/store";
  import { extent } from "d3-array";
  import Line from "./Line.svelte";
  import Area from "./Area.svelte";
  import { scaleLinear } from "d3-scale";
  import { onDestroy, onMount } from "svelte";
  import { tweened } from "svelte/motion";
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { SingleDigitTimesPowerOfTenFormatter } from "@rilldata/web-common/lib/number-formatting/strategies/SingleDigitTimesPowerOfTen";
  import { formatMsInterval } from "@rilldata/web-common/lib/number-formatting/strategies/intervals";
  import { cubicOut } from "svelte/easing";
  import { DateTime } from "luxon";
  import Grid from "./Grid.svelte";
  import Points from "./Points.svelte";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import ScrubRange from "./ScrubRange.svelte";
  import Grains from "./Grains.svelte";

  export type Point = { x: number; y: number; ts: Date };

  export type TimeSeriesLine = {
    data: Point[];
    type: "comparison" | "primary" | "secondary";
    color?: string;
    key: string;
  };

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

  const scrubStart = writable<number | null>(null);
  const scrubEnd = writable<number | null>(null);
  const scrubbing = writable<"start" | "end" | null>(null);

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
  const observer = new ResizeObserver((entries) => {
    for (let entry of entries) {
      clientWidth = entry.contentRect.width;
      clientHeight = entry.contentRect.height;
      tracking.updateScaler(clientWidth, referenceLine.length);
    }
  });

  export let lines: TimeSeriesLine[];
  export let scrubRange: { start: number; end: number } | null;
  export let numberKind: NumberKind;
  export let timeZone: string;
  export let showPoints: boolean;
  export let timeGrain: V1TimeGrain;
  export let xTicks: number[] = [];

  const [initYMin = 0, initYMax = 0] = extent(
    lines.map((d) => d.data).flat(),
    ({ y }) => y,
  );
  const [initXMin = 0, initXMax = 0] = extent(
    lines.map((d) => d.data).flat(),
    ({ x }) => x,
  );

  const xMinTweened = tweened(initXMin, tweenParams);
  const xMaxTweened = tweened(initXMax, tweenParams);
  const yMinTweened = tweened(Math.min(0, initYMin * buffer), tweenParams);
  const yMaxTweened = tweened(initYMax * buffer, tweenParams);

  let element: SVGSVGElement;
  let clientWidth: number;
  let clientHeight: number;
  let grabStart: number | null = null;
  let grabbing = false;
  let scrubGrabStart: number | null = null;
  let scrubGrabEnd: number | null = null;
  let delayedLines = lines;

  onMount(() => {
    observer.observe(element);
  });

  onDestroy(() => {
    observer.disconnect();
  });

  $: referenceLine = lines[0].data;

  $: tracking.updateScaler(clientWidth, referenceLine.length);

  $: xMin = $xMinTweened;
  $: xMax = $xMaxTweened;
  $: yMin = $yMinTweened;
  $: yMax = $yMaxTweened;

  $: yScaler = scaleLinear().domain([yMin, yMax]).range([0, clientHeight]);
  $: xScaler = scaleLinear().domain([xMin, xMax]).range([0, clientWidth]);

  $: indexScaler = (number: number) =>
    Math.floor(
      scaleLinear()
        .domain([0, clientWidth])
        .range([0, referenceLine.length - 1])(number),
    );

  $: [newYMin = 0, newYMax = 0] = extent(
    lines.map((d) => d.data).flat(),
    (d) => d.y,
  );

  $: [newXMin = 0, newXMax = 0] = extent(referenceLine, (d) => d.x);

  $: xMinTweened
    .set(newXMin, { duration: !newXMin ? 0 : 400 })
    .catch(console.error);

  $: xMaxTweened
    .set(newXMax, { duration: !newXMax ? 0 : 400 })
    .catch(console.error);

  $: {
    // if (newYMax >= get(yMax)) {
    //   yMin.set(newYMin, { duration: newYMin === 0 ? 0 : 400 });
    //   yMax.set(newYMax, { duration: newYMin === 0 ? 0 : 400 });
    //   setTimeout(() => {
    //     delayedLines = lines;
    //   }, 500);
    // } else {

    //   setTimeout(() => {
    yMinTweened
      .set(Math.min(0, newYMin * buffer), { duration: !newYMin ? 0 : 400 })
      .catch(console.error);
    yMaxTweened
      .set(newYMax * buffer, { duration: !newYMin ? 0 : 400 })
      .catch(console.error);
    //   }, 600);
    // }
    setTimeout(() => {
      delayedLines = lines;
    }, 500);
  }

  $: if (
    grabbing &&
    grabStart !== null &&
    scrubGrabStart !== null &&
    scrubGrabEnd !== null &&
    $tracking !== null
  ) {
    const point = referenceLine[$tracking].x;
    scrubStart.set(scrubGrabStart + point - grabStart);
    scrubEnd.set(scrubGrabEnd + point - grabStart);
  }

  $: if ($scrubbing === "start" && $tracking !== null) {
    scrubStart.set(referenceLine[$tracking].x);
  } else if ($scrubbing === "end" && $tracking !== null) {
    scrubEnd.set(referenceLine[($tracking ?? 0) + 1].x);
  }

  $: xTicks = getTicks(delayedLines[0].data);
  $: yTicks = yScaler.ticks(2);

  $: formatterFunction = (x: number) =>
    numberKind === NumberKind.INTERVAL
      ? formatMsInterval(x)
      : new SingleDigitTimesPowerOfTenFormatter(yTicks, {
          numberKind,
          padWithInsignificantZeros: false,
        }).stringFormat(x);

  $: hoverDate = hoverPoint?.ts
    ? DateTime.fromJSDate(hoverPoint.ts)
        .setZone(timeZone)
        .toLocaleString(timeFormat[timeGrain])
    : null;

  $: hoverValue = hoverPoint?.y ?? null;

  $: hoverPoint = $tracking !== null ? referenceLine[$tracking] : null;

  $: yRange = yMax - yMin;
  $: xRange = xMax - xMin;

  function getTicks(array: Point[]) {
    const ticks: number[] = [];

    const tickSpacing = Math.ceil(array.length / 8);

    for (let i = 0; i <= array.length; i += tickSpacing) {
      if (array[i]?.x) ticks.push(array[i]?.x);
    }

    return ticks;
  }

  function handleMouseDown() {
    if ($tracking === null) return;
    scrubStart.set(referenceLine[$tracking].x);
    scrubbing.set("end");

    window.addEventListener("mouseup", handleMouseUp);
  }

  function handleMouseUp() {
    scrubbing.set(null);
    if (scrubRange?.end === scrubRange?.start) {
      // currentStore.scrubRange.clear();
      //   dispatch("clear");
    }

    window.removeEventListener("mouseup", handleMouseUp);
  }

  function handleScrubGrab() {
    if ($tracking === null) return;
    grabbing = true;
    grabStart = referenceLine[$tracking].x;
    scrubGrabStart = $scrubStart;
    scrubGrabEnd = $scrubEnd;

    window.addEventListener("mouseup", () => {
      grabbing = false;
      grabStart = null;
    });
  }
</script>

<div
  class="size-full relative cursor-pointer wrapper flex-1 flex gap-2"
  class:cursor-ew-resize={Boolean(scrubRange)}
>
  <div class="flex flex-col size-full relative">
    <svg class="absolute size-full pointer-events-none z-50 overflow-visible">
      <text x="4" baseline-shift="-1.1em" font-size="12px" font-weight="400">
        {hoverDate ?? ""}
      </text>
    </svg>

    <svg
      class="absolute size-full pointer-events-none z-50 overflow-visible"
      transform="scale(1,-1)"
    >
      {#if !scrubbing && hoverValue !== null && !grabbing && $tracking !== null}
        {#each { length: lines.length } as _, i (i)}
          {@const { y: value } = lines[i].data[$tracking]}
          {@const x = xScaler.invert($tracking)}
          {@const y = yScaler(value)}
          <g transform="translate({x}, {y})">
            <text class="hover">{value}</text>
          </g>
        {/each}
      {/if}
    </svg>

    <svg
      bind:this={element}
      role="presentation"
      class="border-t border-gray-300 size-full -scale-y-100"
      viewBox="{xMin} {yMin} {xRange} {yRange}"
      preserveAspectRatio="none"
      on:mousemove={(e) => {
        tracking.set(indexScaler(e.offsetX));
      }}
      on:mousedown={handleMouseDown}
      on:mouseleave={() => {
        tracking.reset();
        grabbing = false;
      }}
    >
      <Grid {xTicks} {yTicks} yExtents={[yMin, yMax]} xExtents={[xMin, xMax]} />

      {#if showPoints}
        <Points lines={delayedLines} />
      {:else}
        {#each delayedLines as { key, type, data, color } (key)}
          <Line {data} {color} {type} />
          {#if type === "primary"}
            <Area {data} />
          {/if}
        {/each}
      {/if}

      <Grains
        {referenceLine}
        {yMin}
        height={yRange}
        hoveredIndex={$scrubbing ? null : $tracking}
      />

      {#if $scrubStart !== null && $scrubEnd !== null}
        <ScrubRange
          {grabbing}
          {yMin}
          {yMax}
          {xMin}
          start={$scrubStart}
          end={$scrubEnd}
          width={$scrubEnd - $scrubStart}
          height={yRange}
          grayscale={lines.length === 1}
          onAdjustScrubEnd={() => {
            scrubbing.set("end");
            window.addEventListener("mouseup", handleMouseUp);
          }}
          onAdjustScrubStart={() => {
            scrubbing.set("start");
            window.addEventListener("mouseup", handleMouseUp);
          }}
          onScrubGrab={handleScrubGrab}
        />
      {/if}
    </svg>
  </div>

  <div class="w-8 h-full relative overflow-visible flex-none">
    {#each yTicks as tick}
      <div class="absolute translate-y-1/2" style:bottom="{yScaler(tick)}px">
        {formatterFunction(tick)}
      </div>
    {/each}
  </div>
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
