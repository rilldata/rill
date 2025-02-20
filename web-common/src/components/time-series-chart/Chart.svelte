<script lang="ts">
  import { min, max, extent } from "d3-array";
  import Line from "./Line.svelte";
  import { scaleLinear, scaleTime } from "d3-scale";
  import { MainLineColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import {
    V1TimeGrain,
    type V1TimeSeriesValue,
  } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import Point from "./Point.svelte";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import Crosshairs from "./Crosshairs.svelte";

  const SNAP_RANGE = 0.05;

  export let primaryData: V1TimeSeriesValue[];
  export let secondaryData: V1TimeSeriesValue[][] = [];
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
  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  let yScale = scaleLinear();

  $: ({ width, height } = contentRect);

  $: data = [primaryData, ...secondaryData];

  $: mappedData = data
    .map((line) => line.map(mapData))
    .filter((line) => line.length > 0);

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

  $: yExtents = [Math.min(0, min(mins) ?? 0), max(maxes) ?? 0];
  $: yScale = yScale.domain(yExtents).range([100, 0]);
  $: ySpan = yExtents[1] - yExtents[0];

  $: hoverIndex =
    offsetPosition === null
      ? null
      : Math.floor((offsetPosition.x / width) * mappedData[0].length);

  $: hoveredPoints = getPoints(hoverIndex);

  $: snappedPoint =
    !!offsetPosition?.y &&
    hoveredPoints[0]?.value !== null &&
    hoveredPoints[0]?.value !== undefined &&
    Math.abs(
      hoveredPoints[0]?.value -
        yScale.invert((offsetPosition?.y / height) * 100),
    ) /
      ySpan <
      SNAP_RANGE &&
    hoveredPoints[0];

  $: svgCoordinateCursor = offsetPosition && {
    x: (offsetPosition.x / width) * 1000,
    y: (offsetPosition.y / height) * 100,
  };

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

  function getPos(pos: number, width: number) {
    const percentage = pos / width;

    if (percentage < 0.1) return "-right-2";
    if (percentage > 0.9) return "-left-2";

    if (percentage <= 0.5) return "-left-2";
    return "-right-2";
  }
</script>

<div role="presentation" class="flex flex-col size-full relative">
  {#if hoveredPoints.length > 0 && offsetPosition}
    <div
      class="{getPos(
        offsetPosition.x,
        width,
      )} w-fit label text-[10px] bg-white border-dashed text-gray-500 border-gray-300 -translate-y-1/2 py-0.5 border rounded-sm px-1 font-medium absolute pointer-events-none"
      style:top="{snappedPoint && snappedPoint.value
        ? (yScale(snappedPoint.value) / 100) * height
        : offsetPosition.y}px"
      class:!text-primary-500={!!snappedPoint}
      class:border-primary-400={!!snappedPoint}
      class:!font-semibold={!!snappedPoint}
    >
      {formatterFunction(
        snappedPoint
          ? snappedPoint.value
          : yScale.invert((offsetPosition?.y / height) * 100),
      )}
    </div>
  {/if}

  <svg
    bind:contentRect
    role="presentation"
    class="cursor-default size-full overflow-visible"
    preserveAspectRatio="none"
    viewBox="0 0 1000 100"
    on:mousemove={(e) => {
      offsetPosition = { x: e.offsetX, y: e.offsetY };
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

    <Crosshairs
      snapped={!!snappedPoint}
      cursor={snappedPoint
        ? {
            x: xScales[0](snappedPoint.interval.start.toJSDate()),
            y: snappedPoint.value ? yScale(snappedPoint.value) : 50,
          }
        : svgCoordinateCursor}
    />

    <g>
      {#each mappedData as mappedDataLine, i (i)}
        {#each mappedDataLine as { interval, value }, j (j)}
          {@const xScale = xScales[i]}
          <Point
            showPoint={hoverIndex === j ||
              (mappedDataLine[j - 1]?.value === null &&
                mappedDataLine[j + 1]?.value === null &&
                mappedDataLine[j]?.value !== null)}
            x={xScale(interval.start.toJSDate())}
            y={value == null ? null : yScale(value)}
            color={getColor(i)}
          />
        {/each}
      {/each}
    </g>
  </svg>

  <div class="w-full h-fit flex justify-between text-gray-500 mt-0.5 relative">
    {#if hoveredPoints.length > 0}
      <span
        class="relative"
        style:transform="translateX(-{xScales[0](
          hoveredPoints[0].interval.start.toJSDate(),
        ) / 10}%)"
        style:left="{xScales[0](hoveredPoints[0].interval.start.toJSDate()) /
          10}%"
      >
        <RangeDisplay interval={hoveredPoints[0].interval} grain={timeGrain} />
      </span>
    {:else}
      <span>
        {mappedData[0][0].interval.start.toLocaleString({
          month: "short",
          day: "numeric",
        })}
      </span>
      <span>
        {mappedData[0][mappedData[0].length - 1].interval.end
          .minus({ millisecond: 1 })
          .toLocaleString({
            month: "short",
            day: "numeric",
          })}
      </span>
    {/if}
  </div>
</div>
