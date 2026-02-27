<script lang="ts">
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import { MainLineColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import {
    V1TimeGrain,
    type V1TimeSeriesValue,
  } from "@rilldata/web-common/runtime-client";
  import { extent, max, min } from "d3-array";
  import { scaleLinear } from "d3-scale";
  import { DateTime, Interval } from "luxon";
  import { onDestroy } from "svelte";
  import Line from "./Line.svelte";
  import Point from "./Point.svelte";
  import type { ChartDataPoint } from "./types";

  const THROTTLE_MS = 16;
  const MIN_WIDTH_FOR_DYNAMIC_LABEL = 200;

  export let primaryData: V1TimeSeriesValue[];
  export let secondaryData: V1TimeSeriesValue[] = [];
  export let timeGrain: V1TimeGrain;
  export let selectedTimeZone: string;
  export let yAccessor: string;
  export let hideTimeRange: boolean | undefined = false;
  export let formatterFunction: ReturnType<typeof createMeasureValueFormatter>;
  export let hoveredPoints: ChartDataPoint[] = [];

  let offsetPosition: { x: number; y: number } | null = null;
  let clientPosition: { x: number; y: number } = { x: 0, y: 0 };
  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);

  let lastMouseUpdateTime = 0;
  let mouseUpdateScheduled = false;

  let hoveredIntervalCache = new Map<string, Interval>();
  let prevTimeZone = selectedTimeZone;
  let prevTimeGrain = timeGrain;

  $: ({ width } = contentRect);

  // Map data to include index for positioning and originalDate for display
  $: mappedPrimaryData = primaryData.map((point, index) =>
    mapData(point, index),
  );
  $: mappedSecondaryData = secondaryData.map((point, index) =>
    mapData(point, index),
  );

  $: hasComparison = mappedSecondaryData.length > 0;

  // Use the longer of the two datasets for the x-scale
  $: maxDataLength = Math.max(
    mappedPrimaryData.length,
    mappedSecondaryData.length,
  );

  $: mappedData = hasComparison
    ? [mappedPrimaryData, mappedSecondaryData]
    : [mappedPrimaryData];

  $: xScale = scaleLinear()
    .domain([0, Math.max(0, maxDataLength - 1)])
    .range([0, 10000]);

  $: allYExtents = mappedData.map((line) =>
    extent(line, (datum) => datum?.value as number),
  );

  $: mins = allYExtents.map((extents) => extents[0]).filter(isNumber);
  $: maxes = allYExtents.map((extents) => extents[1]).filter(isNumber);

  $: yExtents = [Math.min(0, min(mins) ?? 0), max(maxes) ?? 0];
  $: yScale = scaleLinear().domain(yExtents).range([100, 0]);

  $: hoverIndex = (() => {
    if (offsetPosition === null) return null;
    if (maxDataLength === 0) return null;
    if (maxDataLength === 1) return 0;
    return Math.round((offsetPosition.x / width) * (maxDataLength - 1));
  })();

  $: hoveredPoints = getPoints(hoverIndex);

  // Only show tooltip if there's at least one point with a valid value
  $: hasValidHoveredPoints = hoveredPoints.some(
    (p) => p && p.value !== null && p.value !== undefined,
  );

  function getColor(index: number) {
    return index === 0 ? MainLineColor : "var(--color-gray-400)";
  }

  function isNumber(value: unknown): value is number {
    return value !== undefined && value !== null;
  }

  function getPoints(index: number | null) {
    if (index === null) return [];
    return mappedData.map((line) => line?.[index] || null).filter((x) => x);
  }

  function mapData(point: V1TimeSeriesValue, index: number): ChartDataPoint {
    const originalDate = point.ts ? new Date(point.ts) : new Date();
    return {
      index,
      originalDate,
      value: point.records?.[yAccessor] as number | null | undefined,
    };
  }

  function getHoveredInterval(date: Date): Interval {
    const cacheKey = `${date.getTime()}-${selectedTimeZone}-${timeGrain}`;

    if (hoveredIntervalCache.has(cacheKey)) {
      return hoveredIntervalCache.get(cacheKey)!;
    }

    const interval = Interval.fromDateTimes(
      DateTime.fromJSDate(date).setZone(selectedTimeZone),
      DateTime.fromJSDate(date)
        .setZone(selectedTimeZone)
        .plus({ [TIME_GRAIN[timeGrain]?.label || "minute"]: 1 }),
    );

    hoveredIntervalCache.set(cacheKey, interval);
    return interval;
  }

  function formatDate(date: Date): string {
    return DateTime.fromJSDate(date).setZone(selectedTimeZone).toLocaleString({
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  }

  function handleThrottledMouseMove(e: MouseEvent) {
    const now = performance.now();

    clientPosition = { x: e.clientX, y: e.clientY };

    if (now - lastMouseUpdateTime < THROTTLE_MS) {
      if (!mouseUpdateScheduled) {
        mouseUpdateScheduled = true;
        requestAnimationFrame(() => {
          offsetPosition = { x: e.offsetX, y: e.offsetY };
          mouseUpdateScheduled = false;
          lastMouseUpdateTime = performance.now();
        });
      }
      return;
    }

    offsetPosition = { x: e.offsetX, y: e.offsetY };
    lastMouseUpdateTime = now;
  }

  function handleMouseLeave() {
    offsetPosition = null;
  }

  // Clear cache when timezone or time grain changes
  $: if (selectedTimeZone !== prevTimeZone || timeGrain !== prevTimeGrain) {
    hoveredIntervalCache.clear();
    prevTimeZone = selectedTimeZone;
    prevTimeGrain = timeGrain;
  }

  onDestroy(() => {
    hoveredIntervalCache.clear();
  });
</script>

{#if mappedData.length}
  <div role="presentation" class="flex flex-col grow h-full relative">
    {#if hasValidHoveredPoints && offsetPosition}
      <div
        use:portal
        class=" w-fit label text-[10px] font-semibold flex flex-col z-[1000] shadow-sm bg-surface-subtle text-fg-secondary -translate-y-1/2 py-0.5 border rounded-sm px-1 absolute pointer-events-none"
        style:top="{clientPosition.y}px"
        style:left="{clientPosition.x + 10}px"
      >
        {#each hoveredPoints as point, i (i)}
          {#if point && point.value !== null && point.value !== undefined}
            <div class="flex gap-x-1 items-center">
              <span
                class="size-[6.5px] rounded-full"
                style:background-color={getColor(i)}
              />
              <span>{formatterFunction(point.value)}</span>
              {#if hasComparison}
                <span class="text-fg-muted">
                  {formatDate(point.originalDate)}
                </span>
              {/if}
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
      on:mousemove={handleThrottledMouseMove}
      on:mouseleave={handleMouseLeave}
    >
      {#each mappedData as mappedDataLine, i (i)}
        <Line
          data={mappedDataLine}
          {xScale}
          color={getColor(i)}
          {yScale}
          fill={i === 0}
          strokeWidth={1}
        />
      {/each}

      <g>
        {#each [...mappedData].reverse() as mappedDataLine, reversedIndex (reversedIndex)}
          {@const i = mappedData.length - reversedIndex - 1}
          {#each mappedDataLine as { index, value }, pointIndex (pointIndex)}
            {#if value !== null && value !== undefined && (hoverIndex === pointIndex || (mappedDataLine[pointIndex - 1]?.value === null && mappedDataLine[pointIndex + 1]?.value === null))}
              <Point x={xScale(index)} y={yScale(value)} color={getColor(i)} />
            {/if}
          {/each}
        {/each}
      </g>
    </svg>

    <div
      class="w-full h-fit min-h-[16px] flex justify-between text-fg-secondary mt-0.5 relative"
    >
      {#if hoveredPoints.length > 0}
        {@const percentage = xScale(hoveredPoints[0].index) / 100}
        {@const interval = getHoveredInterval(hoveredPoints[0].originalDate)}
        {@const comparisonPoint =
          hasComparison && hoveredPoints[1] ? hoveredPoints[1] : null}
        {@const useDynamicPosition = width >= MIN_WIDTH_FOR_DYNAMIC_LABEL}
        {#if interval.isValid}
          <span
            class="absolute flex whitespace-nowrap gap-x-1"
            style:transform="translateX(-{useDynamicPosition
              ? percentage
              : 50}%)"
            style:left="{useDynamicPosition ? percentage : 50}%"
          >
            <RangeDisplay {interval} {timeGrain} />
            {#if comparisonPoint}
              <span class="text-fg-muted">
                vs. {formatDate(comparisonPoint.originalDate)}
              </span>
            {/if}
          </span>
        {/if}
      {:else if mappedPrimaryData.length}
        {@const firstPoint = mappedPrimaryData[0]}
        {@const lastPoint = mappedPrimaryData[mappedPrimaryData.length - 1]}
        {#if firstPoint && lastPoint && !hideTimeRange}
          <span>
            {DateTime.fromJSDate(firstPoint.originalDate)
              .setZone(selectedTimeZone)
              .toLocaleString({
                month: "short",
                day: "numeric",
              })}
          </span>
          <span>
            {DateTime.fromJSDate(lastPoint.originalDate)
              .setZone(selectedTimeZone)
              .toLocaleString({
                month: "short",
                day: "numeric",
              })}
          </span>
        {/if}
      {/if}
    </div>
  </div>
{/if}
