<script lang="ts">
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import { MainLineColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import {
    V1TimeGrain,
    type V1TimeSeriesValue,
  } from "@rilldata/web-common/runtime-client";
  import { extent, max, min } from "d3-array";
  import { scaleLinear, scaleTime } from "d3-scale";
  import { DateTime, Interval } from "luxon";
  import Line from "./Line.svelte";
  import Point from "./Point.svelte";
  import { portal } from "@rilldata/web-common/lib/actions/portal";

  const SNAP_RANGE = 0.05;

  export let primaryData: V1TimeSeriesValue[];
  export let secondaryData: V1TimeSeriesValue[] = [];
  export let timeGrain: V1TimeGrain;
  export let selectedTimeZone: string;
  export let yAccessor: string;
  export let formatterFunction: ReturnType<typeof createMeasureValueFormatter>;
  export let hoveredPoints: MappedPoint[] = [];

  type MappedPoint = {
    interval: Interval<true>;
    value: number | null | undefined;
  };

  let offsetPosition: { x: number; y: number } | null = null;
  let clientPosition: { x: number; y: number } = { x: 0, y: 0 };
  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  let yScale = scaleLinear();

  $: ({ width, height } = contentRect);

  $: data = [primaryData, secondaryData];

  $: mappedData = data
    .map((line) => line.map(mapData))
    .filter((line) => line.length > 0);

  $: xExtents = mappedData.map((line) => [
    line?.[0]?.interval.start.toJSDate(),
    line[line.length - 1].interval.start.toJSDate(),
  ]);

  $: xScales = xExtents.map((extents) =>
    scaleTime().domain(extents).range([0, 10000]),
  );

  $: allYExtents = mappedData.map((line) =>
    extent(line, (datum) => datum?.value as number),
  );

  $: mins = allYExtents.map((extents) => extents[0]).filter(isNumber);
  $: maxes = allYExtents.map((extents) => extents[1]).filter(isNumber);

  $: maxDataLength = Math.max(...mappedData.map((line) => line.length));

  $: yExtents = [Math.min(0, min(mins) ?? 0), max(maxes) ?? 0];
  $: yScale = yScale.domain(yExtents).range([100, 0]);
  $: ySpan = yExtents[1] - yExtents[0];

  $: hoverIndex =
    offsetPosition === null
      ? null
      : Math.round((offsetPosition.x / width) * (maxDataLength - 1));

  $: hoveredPoints = getPoints(hoverIndex);

  $: nearPoints = offsetPosition
    ? hoveredPoints
        .map((point, index) => {
          if (
            point === null ||
            point.value === null ||
            point.value === undefined
          )
            return null;

          if (
            Math.abs(
              point?.value -
                yScale.invert(((offsetPosition?.y as number) / height) * 100),
            ) /
              ySpan <
            SNAP_RANGE
          )
            return {
              point,
              index,
            };
          return null;
        })
        .sort((a, b) => {
          if (a === null) return 1;
          if (b === null) return -1;

          return (b.point?.value ?? 0) - (a.point?.value ?? 0);
        })
    : [];

  function getColor(index: number) {
    return index === 0 ? MainLineColor : "rgba(0, 0, 0, 0.22)";
  }

  function isNumber(value: unknown): value is number {
    return value !== undefined && value !== null;
  }

  function getPoints(index: number | null) {
    if (index === null) return [];
    return mappedData.map((line) => line?.[index] || null).filter((x) => x);
  }

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
</script>

{#if mappedData.length}
  <div role="presentation" class="flex flex-col grow h-full relative outline">
    {#if nearPoints.filter(Boolean).length && clientPosition}
      <div
        use:portal
        class=" w-fit label text-[10px] font-semibold flex flex-col z-[1000] shadow-sm bg-white text-gray-500 border-gray-200 -translate-y-1/2 py-0.5 border rounded-sm px-1 absolute pointer-events-none"
        style:top="{clientPosition.y}px"
        style:left="{clientPosition.x + 10}px"
      >
        {#each nearPoints as possiblePoint, i (i)}
          {#if possiblePoint}
            <div class="flex gap-x-1 items-center">
              <span
                class="size-[6.5px] rounded-full"
                style:background-color={getColor(possiblePoint.index)}
              />
              {formatterFunction(possiblePoint?.point.value)}
            </div>
          {/if}
        {/each}
      </div>
    {/if}

    <svg
      bind:contentRect
      role="presentation"
      class="cursor-default size-full overflow-visible"
      preserveAspectRatio="none"
      viewBox="0 0 10000 100"
      on:mousemove={(e) => {
        offsetPosition = { x: e.offsetX, y: e.offsetY };
        clientPosition = { x: e.clientX, y: e.clientY };
      }}
      on:mouseleave={() => {
        offsetPosition = null;
      }}
    >
      {#each mappedData as mappedDataLine, i (i)}
        <Line
          data={mappedDataLine}
          xScale={xScales[i]}
          color={getColor(i)}
          {yScale}
          fill={i === 0}
          strokeWidth={1}
        />
      {/each}

      <g>
        {#each [...mappedData].reverse() as mappedDataLine, reversedIndex (reversedIndex)}
          {@const i = mappedData.length - reversedIndex - 1}
          {#each mappedDataLine as { interval, value }, pointIndex (pointIndex)}
            {@const xScale = xScales[i]}
            {#if value !== null && value !== undefined && (hoverIndex === pointIndex || (mappedDataLine[pointIndex - 1]?.value === null && mappedDataLine[pointIndex + 1]?.value === null))}
              <Point
                x={xScale(interval.start.toJSDate())}
                y={yScale(value)}
                color={getColor(i)}
              />
            {/if}
          {/each}
        {/each}
      </g>
    </svg>

    <div
      class="w-full h-fit flex justify-between text-gray-500 mt-0.5 relative"
    >
      {#if hoveredPoints.length > 0}
        {@const jsDate = hoveredPoints[0].interval.start.toJSDate()}
        {@const percentage = xScales[0](jsDate) / 100}
        <span
          class="relative"
          style:transform="translateX(-{percentage}%)"
          style:left="{percentage}%"
        >
          <RangeDisplay
            interval={hoveredPoints[0].interval}
            grain={timeGrain}
          />
        </span>
      {:else if mappedData.length}
        {@const firstPoint = mappedData?.[0]?.[0]}
        {@const lastPoint = mappedData?.[0]?.[mappedData?.[0]?.length - 1]}
        {#if firstPoint && lastPoint}
          <span>
            {firstPoint.interval.start.toLocaleString({
              month: "short",
              day: "numeric",
            })}
          </span>
          <span>
            {lastPoint.interval.end.minus({ millisecond: 1 }).toLocaleString({
              month: "short",
              day: "numeric",
            })}
          </span>
        {/if}
      {/if}
    </div>
  </div>
{/if}
