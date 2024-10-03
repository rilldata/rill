<script lang="ts" context="module">
  import { writable } from "svelte/store";

  const tracking = (() => {
    const { subscribe, set } = writable<number | null>(null);
    let cursorScaler = scaleLinear();

    return {
      subscribe,
      track: (e: MouseEvent) => set(cursorScaler(e.offsetX)),

      reset: () => set(null),
      updateScaler: (width: number) => {
        cursorScaler = cursorScaler.domain([0, width]).range([0, 100]);
      },
    };
  })();
</script>

<script lang="ts">
  import { extent } from "d3-array";
  import Line from "./Line.svelte";
  import Area from "./Area.svelte";
  import type { TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import ScaleBoundary from "./ScaleBoundary.svelte";
  import { scaleLinear, scaleTime } from "d3-scale";
  import { onDestroy, onMount } from "svelte";
  import { roundDownToTimeUnit } from "@rilldata/web-common/features/dashboards/time-series/round-to-nearest-time-unit";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
  import {
    ScrubArea0Color,
    ScrubArea1Color,
    ScrubArea2Color,
    ScrubBoxColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { tweened } from "svelte/motion";
  import Grid from "./Grid.svelte";
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { SingleDigitTimesPowerOfTenFormatter } from "@rilldata/web-common/lib/number-formatting/strategies/SingleDigitTimesPowerOfTen";
  import { formatMsInterval } from "@rilldata/web-common/lib/number-formatting/strategies/intervals";

  const observer = new ResizeObserver((entries) => {
    for (let entry of entries) {
      clientWidth = entry.contentRect.width;
      tracking.updateScaler(clientWidth);
    }
  });

  export let data: TimeSeriesDatum[];
  export let yAccessor: string;
  export let timeGrain: AvailableTimeGrain;
  export let scrubRange: { start: Date; end: Date } | null = null;
  export let numberKind: NumberKind;
  export let xAccessor = "ts_position";
  export let showGrid = true;
  export let showAxis = true;
  export let showBorder = true;

  let element: HTMLDivElement;
  let clientWidth: number;
  let scrubbing: "start" | "end" | null = null;

  let xScale = scaleTime();
  let yScale = scaleLinear();

  onMount(() => {
    observer.observe(element);
  });

  onDestroy(() => {
    observer.disconnect();
  });

  $: xExtents = extent(data, (d) => d[xAccessor] as Date);
  $: yExtents = extent(data, (d) => d[yAccessor] as number);

  $: indexScaler = scaleLinear().domain([0, 100]).range([0, data.length]);

  $: mouseOverIndex = Math.floor(indexScaler($tracking ?? 0));

  $: timeGrainPreset =
    TIME_GRAIN[timeGrain as AvailableTimeGrain] ?? TIME_GRAIN.TIME_GRAIN_HOUR;

  $: mouseOverDate = roundDownToTimeUnit(
    data[mouseOverIndex]?.[xAccessor] as Date,
    timeGrainPreset.label,
  );

  $: mouseOverEntry = data[mouseOverIndex];

  $: mouseOverYValue = mouseOverEntry?.[yAccessor];

  $: mouseOverYScaled = mouseOverYValue ? yScale(Number(mouseOverYValue)) : 0;

  let grabbing = false;
  let previousGrab: number;

  function handleScrubGrab() {
    grabbing = true;
    previousGrab = Number(mouseOverDate);

    window.addEventListener("mouseup", () => {
      grabbing = false;
    });
  }

  $: if (grabbing && scrubRange) {
    const current = Number(mouseOverDate);
    const grabDelta = current - previousGrab;
    previousGrab = current;
    //   currentStore.scrubRange.shift(grabDelta);
  }

  function handleMouseDown() {
    scrubbing = "end";
    scrubRange = { start: mouseOverDate, end: mouseOverDate };

    scrubRange = {
      start: mouseOverDate,
      end: mouseOverDate,
    };

    window.addEventListener("mouseup", handleMouseUp);
  }

  function handleMouseUp() {
    scrubbing = null;
    if (scrubRange?.end === scrubRange?.start) {
      scrubRange = null;
    }

    scrubRange.end = mouseOverDate;
    window.removeEventListener("mouseup", handleMouseUp);
  }

  $: if (scrubRange) {
    if (scrubbing === "start") {
      scrubRange.start = mouseOverDate;
    } else if (scrubbing === "end") {
      scrubRange.end = mouseOverDate;
    }
  }

  $: hasValidRange = scrubRange && scrubRange?.start !== scrubRange?.end;

  let xStart: number;
  let xEnd: number;

  $: if (scrubRange) {
    xStart = xScale(scrubRange.start);
    xEnd = xScale(scrubRange.end);
  }

  const mouseOver = tweened(xScale(mouseOverDate), {
    duration: 50,
  });

  const mouseOverY = tweened(mouseOverYScaled, {
    duration: 50,
  });

  $: mouseOver.set(xScale(mouseOverDate));

  $: mouseOverY.set(mouseOverYScaled);

  $: yTicks = yScale.ticks(2);

  $: formatterFunction = (x: number) =>
    numberKind === NumberKind.INTERVAL
      ? formatMsInterval(x)
      : new SingleDigitTimesPowerOfTenFormatter(yTicks, {
          numberKind,
          padWithInsignificantZeros: false,
        }).stringFormat(x);

  let contentRect = new DOMRect(0, 0, 0, 0);

  $: ({ height } = contentRect);
</script>

<div
  class="wrapper"
  class:cursor-ew-resize={Boolean(scrubbing)}
  bind:contentRect
>
  <div class="flex flex-col w-full gap-y-2" bind:this={element}>
    <svg
      role="presentation"
      class:border-b={showBorder}
      class="border-1 border-gray-300"
      width="100%"
      height="100%"
      viewBox="0 0 100 100"
      preserveAspectRatio="none"
      on:mousemove={tracking.track}
      on:mousedown={handleMouseDown}
      on:mouseleave={() => {
        tracking.reset();
        grabbing = false;
      }}
    >
      {#if xExtents[0] !== undefined && yExtents[0] !== undefined}
        <ScaleBoundary {xExtents} {yExtents} bind:xScale bind:yScale>
          {#if showGrid}
            <Grid {xScale} {yScale} />
          {/if}
          <Line
            {data}
            xKey={xAccessor}
            yKey={yAccessor}
            xScaler={xScale}
            yScaler={yScale}
          />
          <Area
            {data}
            xKey={xAccessor}
            yKey={yAccessor}
            xScaler={xScale}
            yScaler={yScale}
          />

          {#if $tracking !== null}
            <line
              x1={$mouseOver}
              y1={$mouseOverY}
              x2={$mouseOver}
              y2={100}
              class="stroke-primary-300 z-10 pointer-events-none"
              stroke-width="3"
              stroke-linecap="round"
              vector-effect="non-scaling-stroke"
            />

            <path
              d="M {$mouseOver} {$mouseOverY} l 0.0001 0"
              stroke-linecap="round"
              stroke-width="5"
              stroke="black"
              vector-effect="non-scaling-stroke"
            />
          {/if}

          {#if hasValidRange}
            <line
              role="presentation"
              x1={xStart}
              y1={0}
              x2={xStart}
              y2={100}
              stroke={ScrubBoxColor}
              stroke-width="1"
              class="cursor-ew-resize"
              vector-effect="non-scaling-stroke"
              on:mousedown|stopPropagation={() => {
                scrubbing = "start";
                window.addEventListener("mouseup", handleMouseUp);
              }}
            />
            <line
              role="presentation"
              x1={xEnd}
              y1={0}
              x2={xEnd}
              y2={100}
              stroke={ScrubBoxColor}
              class="cursor-ew-resize"
              stroke-width="1"
              vector-effect="non-scaling-stroke"
              on:mousedown|stopPropagation={() => {
                scrubbing = "end";
                window.addEventListener("mouseup", handleMouseUp);
              }}
            />
            <g style:mix-blend-mode="hue">
              <rect width="100%" height="100%" fill="white" />
              <rect
                role="presentation"
                on:mousedown|stopPropagation={handleScrubGrab}
                class:cursor-grab={!scrubbing}
                x={xStart}
                y={0}
                width={xEnd - xStart}
                height={100}
                fill="url(#scrubbing-gradient)"
                opacity="0.2"
              />
            </g>

            <defs>
              <linearGradient
                gradientUnits="userSpaceOnUse"
                id="scrubbing-gradient"
              >
                <stop stop-color={ScrubArea0Color} />
                <stop offset="0.36" stop-color={ScrubArea1Color} />
                <stop offset="1" stop-color={ScrubArea2Color} />
              </linearGradient>
            </defs>

            {#if grabbing}
              <filter id="shadow">
                <feDropShadow
                  dx="0"
                  dy="0"
                  stdDeviation="3"
                  flood-color="rgba(0, 0, 0, 1)"
                />
              </filter>
            {/if}
          {/if}
        </ScaleBoundary>
      {/if}
    </svg>
  </div>

  {#if showAxis}
    <svg class="w-8 h-full overflow-visible">
      {#each yTicks as tick, i (i)}
        <text
          x="0"
          y={(yScale(tick) * height) / 100}
          font-size="0.55rem"
          text-anchor="start"
          dominant-baseline="middle"
          vector-effect="non-scaling-stroke"
        >
          {formatterFunction(tick)}
        </text>
      {/each}
    </svg>
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply size-full relative flex-1 flex gap-2;
  }
</style>
