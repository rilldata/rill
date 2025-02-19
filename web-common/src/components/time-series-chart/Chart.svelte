<script lang="ts">
  import { min, max, extent } from "d3-array";
  import Line from "./Line.svelte";
  import { scaleLinear, scaleTime } from "d3-scale";
  import {
    LineMutedColor,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";

  import {
    V1TimeGrain,
    type V1TimeSeriesValue,
  } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { writable } from "svelte/store";
  import Point from "./Point.svelte";

  export let data: V1TimeSeriesValue[][];
  export let timeGrain: V1TimeGrain;
  export let selectedTimeZone: string;
  export let yAccessor: string;
  export let formatterFunction: ReturnType<typeof createMeasureValueFormatter>;

  export let showAxis = true;
  export let yMaxPadding = 0.2;

  export const hoverIndexStore = (() => {
    const { subscribe, set } = writable<number | null>(null);
    return {
      subscribe,
      set,
      reset: () => set(null),
    };
  })();

  type MappedPoint = {
    interval: Interval<true>;
    value: number | null | undefined;
  };

  function mapData(point: V1TimeSeriesValue): MappedPoint {
    if (!point.ts)
      return {
        interval: Interval.fromDateTimes(DateTime.now(), DateTime.now()),
        value: null,
      } as MappedPoint;
    return {
      interval: Interval.fromDateTimes(
        DateTime.fromISO(point.ts).setZone(selectedTimeZone),
        DateTime.fromISO(point.ts)
          .setZone(selectedTimeZone)
          .plus({ [TIME_GRAIN[timeGrain].label]: 1 }),
      ),
      value: point.records?.[yAccessor] as number | null | undefined,
    } as MappedPoint;
  }

  $: mappedData = data
    .map((line) => line.map(mapData))
    .filter((line) => line.length > 0);

  $: console.log(mappedData);

  let yScale = scaleLinear();

  $: grainWidth = 1020 / mappedData[0].length;

  $: xExtents = mappedData.map((line) => [
    line?.[0]?.interval.start.toJSDate(),
    line[line.length - 1].interval.start.toJSDate(),
  ]);

  $: xScales = xExtents.map((extents) =>
    scaleTime().domain(extents).range([0, 1000]),
  );

  $: allYExtents = mappedData.map((line) =>
    extent(line, (datum) => datum?.value as number),
  );

  $: mins = allYExtents.map((extents) => extents[0]).filter(isNumber);
  $: maxes = allYExtents.map((extents) => extents[1]).filter(isNumber);

  function isNumber(value: unknown): value is number {
    return value !== undefined && value !== null;
  }

  $: yExtents = [Math.min(0, min(mins) ?? 0), max(maxes) ?? 0];

  $: yScale = yScale
    .domain([yExtents[0], yExtents[1] * (1 + yMaxPadding)])
    .range([1, 0]);

  $: yTicks = yScale.ticks(2);

  $: hoverIndex = $hoverIndexStore;

  $: console.log({ mappedData });
</script>

<div class="flex flex-col size-full overflow-hidden">
  <!-- {#if hoverIndex !== null}
    {@const point = mappedData[hoverIndex]}

    <span class="absolute top-2 pointer-events-none">
      {point.value}
    </span>

    <span class="absolute pointer-events-none">
      {point.interval
        .set({
          end: point.interval.end.minus({ millisecond: 1 }),
        })
        .toLocaleString(DateTime.DATE_MED)}
    </span>
  {/if} -->

  <!-- {#if showAxis}
      <div>
        <svg class="w-full h-8 overflow-visible">
          {#each xScale.ticks(3) as tick, i (i)}
            <text
              x="{((tick.getTime() / DIVISOR - numberXExtents[0]) / dateWidth) *
                100}%"
              y="50%"
              text-anchor="middle"
              dominant-baseline="middle"
              font-size="0.65rem"
              fill="#6B7280"
            >
              {DateTime.fromJSDate(tick).toLocaleString({
                // month: "short",
                // day: "numeric",
                hour: "numeric",
              })}
            </text>
          {/each}
        </svg>
      </div>
    {/if} -->
  <svg
    role="presentation"
    class="cursor-pointer size-full overflow-hidden bg-red-50"
    preserveAspectRatio="none"
    viewBox="{-10} {0} {1020} {1}"
    on:mouseleave={() => {
      hoverIndexStore.reset();
      // grabbing = false;
    }}
  >
    <!-- {#if showGrid}
      <Grid {xScale} {yScale} timeZone={selectedTimeZone} />
    {/if} -->

    {#each mappedData as mappedDataLine, i (i)}
      <Line
        data={mappedDataLine}
        xScale={xScales[i]}
        color={i === 0 ? MainLineColor : LineMutedColor}
        {yScale}
        fill={i === 0}
        strokeWidth={1}
      />
    {/each}

    <g>
      {#each mappedData as mappedDataLine, i (i)}
        {#each mappedDataLine as { interval, value }, j (j)}
          {@const xScale = xScales[i]}

          {#if i === 0}
            <rect
              x={xScale(interval.start.toJSDate()) - grainWidth / 2}
              y={0}
              width={grainWidth}
              height="100%"
              role="presentation"
              class="opacity-0 fill-primary-50"
              on:mouseenter={() => {
                console.log(j);
                hoverIndexStore.set(j);
              }}
            />
          {/if}

          <Point
            showPoint={hoverIndex === j ||
              (mappedDataLine[j - 1]?.value === null &&
                mappedDataLine[j + 1]?.value === null)}
            showLabel={false}
            flipLabel={false}
            x={xScale(interval.start.toJSDate())}
            y={value == null ? null : yScale(value)}
            color="blue"
          />
        {/each}
      {/each}
    </g>
  </svg>
</div>

{#if showAxis}
  <svg class="w-8 h-full overflow-visible">
    {#each yTicks as tick, i (i)}
      <text
        x="0"
        y={yScale(tick)}
        font-size="0.65rem"
        dominant-baseline="middle"
        fill="#6B7280"
      >
        {formatterFunction(tick)}
      </text>
    {/each}
  </svg>
{/if}

<!-- </div> -->

<style lang="postcss">
  .wrapper {
    @apply size-full relative flex-1 flex gap-2;
  }
</style>
