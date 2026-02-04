<!-- @component
Diagnostic plot of count(*) over time for a timestamp column.
Enables zoom (ctrl+drag), pan (drag when zoomed), and shift+click to copy.
Uses index-based scales and TimeSeriesChart for rendering.
-->
<script lang="ts">
  import TimeSeriesChart from "@rilldata/web-common/components/time-series-chart/TimeSeriesChart.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import type {
    ChartScales,
    ChartSeries,
  } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";
  import { snapIndex } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/utils";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import {
    datePortion,
    formatInteger,
    timePortion,
  } from "@rilldata/web-common/lib/formatters";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { timeGrainToDuration } from "@rilldata/web-common/lib/time/grains";
  import { removeLocalTimezoneOffset } from "@rilldata/web-common/lib/time/timezone";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { createLineGenerator } from "@rilldata/web-common/components/data-graphic/utils";
  import { max } from "d3-array";
  import { scaleLinear } from "d3-scale";
  import { onMount } from "svelte";
  import { fade, fly } from "svelte/transition";
  import TimestampBound from "./TimestampBound.svelte";
  import TimestampProfileSummary from "./TimestampProfileSummary.svelte";
  import TimestampTooltipContent from "./TimestampTooltipContent.svelte";
  import type { TimestampDataPoint } from "@rilldata/web-common/features/column-profile/queries";

  const id = guidGenerator();
  const tooltipSparkWidth = 84;
  const tooltipSparkHeight = 12;

  export let data: TimestampDataPoint[];
  export let spark: TimestampDataPoint[];
  export let width = 360;
  export let height = 120;
  export let mouseover = false;
  export let smooth = true;
  export let left = 1;
  export let right = 1;
  export let top = 12;
  export let bottom = 4;
  export let buffer = 0;
  export let fontSize = 12;
  export let textGap = 4;
  export let zoomWindowColor = "var(--surface-active)";
  export let rollupTimeGrain: V1TimeGrain;
  export let estimatedSmallestTimeGrain: V1TimeGrain;

  let devicePixelRatio = 1;
  let zoomStartIdx = 0;
  let zoomEndIdx: number;
  let isZoomed = false;
  let isDraggingZoom = false;
  let isDraggingPan = false;
  let dragStartX = 0;
  let dragCurrentX = 0;
  let hoveredIndex = -1;

  onMount(() => {
    devicePixelRatio = window.devicePixelRatio;
  });

  // Layout
  $: plotLeft = left + buffer;
  $: plotRight = width - right - buffer;
  $: plotTop = top + buffer;
  $: plotBottom = height - bottom - buffer;
  $: plotWidth = plotRight - plotLeft;

  // Zoom resets when data changes
  $: zoomEndIdx = data.length - 1;

  // Scales
  $: xScale = scaleLinear()
    .domain([zoomStartIdx, zoomEndIdx])
    .range([plotLeft, plotRight]);

  $: yMaxVal = Math.max(5, max(data, (d: TimestampDataPoint) => val(d)) ?? 5);
  $: yScale = scaleLinear().domain([0, yMaxVal]).range([plotBottom, plotTop]);

  $: scales = { x: xScale, y: yScale } satisfies ChartScales;

  // Series
  $: primaryValues = data.map((d: TimestampDataPoint): number | null => val(d));

  // Adaptive line density
  $: visibleStart = Math.max(0, Math.floor(zoomStartIdx));
  $: visibleEnd = Math.min(data.length, Math.ceil(zoomEndIdx) + 1);
  $: dataWindow = data.slice(visibleStart, visibleEnd);

  $: totalTravelDistance = dataWindow
    .map((_: TimestampDataPoint, i: number) => {
      if (i === dataWindow.length - 1) return 0;
      return Math.abs(
        yScale(val(dataWindow[i + 1])) - yScale(val(dataWindow[i])),
      );
    })
    .reduce((acc: number, v: number) => acc + v, 0);

  $: lineDensity = Math.min(
    1,
    Math.max(
      2 / (totalTravelDistance / (plotWidth * devicePixelRatio)),
      (plotWidth * devicePixelRatio * 0.7) / dataWindow.length / 1.5,
    ),
  );

  $: opacity = Math.min(
    1,
    1 + (plotWidth * devicePixelRatio) / dataWindow.length / 2,
  );

  $: chartSeries = [
    {
      id: "primary",
      values: primaryValues,
      color: "var(--fg-muted)",
      strokeWidth: lineDensity,
      opacity,
      areaGradient: { dark: "var(--fg-muted)", light: "transparent" },
    },
  ] satisfies ChartSeries[];

  // Smoothed line
  $: windowWithoutZeros = dataWindow.filter(
    (d: TimestampDataPoint) => val(d) !== 0,
  );
  $: windowSize =
    dataWindow.length < 150 ? 30 : Math.trunc(dataWindow.length / 25);

  $: smoothedValues = data.map(
    (_: TimestampDataPoint, i: number, arr: TimestampDataPoint[]) => {
      const w = Math.max(3, Math.min(Math.trunc(windowSize), i));
      const half = Math.trunc(w / 2);
      const prev = arr.slice(i - half, i + half);
      if (prev.length === 0) return val(arr[i]);
      return (
        prev.reduce((a: number, b: TimestampDataPoint) => a + val(b), 0) /
        prev.length
      );
    },
  );

  $: showSmoothed =
    smooth &&
    windowWithoutZeros.length > 0 &&
    windowWithoutZeros.length > width * devicePixelRatio;

  $: smoothedLineGen = createLineGenerator<number>({
    x: (_d, i) => xScale(i),
    y: (d) => yScale(d ?? 0),
  });

  $: smoothedPath = smoothedLineGen(smoothedValues) ?? "";

  $: isDragging = isDraggingZoom || isDraggingPan;

  // Cursor
  $: cursor = isDraggingZoom ? "text" : isDraggingPan ? "grab" : "inherit";

  // Date extents
  $: xExtentStart = dateAt(data[0]);
  $: xExtentEnd = dateAt(data[data.length - 1]);

  // Zoom bounds as dates
  $: zoomMinDate = isDraggingZoom
    ? indexToDate(xScale.invert(Math.min(dragStartX, dragCurrentX)))
    : indexToDate(zoomStartIdx);

  $: zoomMaxDate = isDraggingZoom
    ? indexToDate(xScale.invert(Math.max(dragStartX, dragCurrentX)))
    : indexToDate(zoomEndIdx);

  // Zoomed row count
  $: zoomedRows = computeZoomedRows(
    isDraggingZoom,
    dragStartX,
    dragCurrentX,
    zoomStartIdx,
    zoomEndIdx,
  );

  $: totalRows = Math.trunc(
    data.reduce((a: number, b: TimestampDataPoint) => a + val(b), 0),
  );

  function val(d: TimestampDataPoint): number {
    return d.count;
  }

  function dateAt(d: TimestampDataPoint) {
    return d.ts;
  }

  function indexToDate(idx: number): Date {
    return dateAt(data[snapIndex(idx, data.length)]) ?? new Date();
  }

  function computeZoomedRows(
    dragging: boolean,
    startX: number,
    currentX: number,
    startIdx: number,
    endIdx: number,
  ): number {
    const start = dragging
      ? xScale.invert(Math.min(startX, currentX))
      : startIdx;
    const end = dragging ? xScale.invert(Math.max(startX, currentX)) : endIdx;
    const s = Math.max(0, Math.round(start));
    const e = Math.min(data.length - 1, Math.round(end));
    let sum = 0;
    for (let i = s; i <= e; i++) {
      sum += val(data[i]);
    }
    return Math.trunc(sum);
  }

  function handleMouseDown(event: MouseEvent) {
    if (event.button !== 0) return;
    if (event.shiftKey) return;

    const x = event.offsetX;
    if (x < plotLeft || x > plotRight) return;

    if (event.ctrlKey || event.metaKey) {
      isDraggingZoom = true;
      dragStartX = x;
      dragCurrentX = x;
    } else if (isZoomed) {
      isDraggingPan = true;
      dragStartX = x;
      dragCurrentX = x;
    }
  }

  function handleMouseMove(event: MouseEvent) {
    const x = event.offsetX;

    if (isDraggingZoom) {
      dragCurrentX = x;
    } else if (isDraggingPan) {
      const dx = x - dragCurrentX;
      const idxDelta = xScale.invert(0) - xScale.invert(dx);
      const newStart = zoomStartIdx + idxDelta;
      const newEnd = zoomEndIdx + idxDelta;
      if (newStart >= 0 && newEnd <= data.length - 1) {
        zoomStartIdx = newStart;
        zoomEndIdx = newEnd;
      }
      dragCurrentX = x;
    }

    if (mouseover && x >= plotLeft && x <= plotRight) {
      const fractionalIdx = xScale.invert(x);
      hoveredIndex = Math.max(
        0,
        Math.min(data.length - 1, Math.round(fractionalIdx)),
      );
    }
  }

  function handleMouseUp() {
    if (isDraggingZoom) {
      const startIdx = xScale.invert(Math.min(dragStartX, dragCurrentX));
      const endIdx = xScale.invert(Math.max(dragStartX, dragCurrentX));
      if (Math.abs(dragCurrentX - dragStartX) > 4) {
        zoomStartIdx = Math.max(0, startIdx);
        zoomEndIdx = Math.min(data.length - 1, endIdx);
        isZoomed = true;
      }
      isDraggingZoom = false;
    }
    if (isDraggingPan) {
      isDraggingPan = false;
    }
  }

  function handleMouseLeave() {
    hoveredIndex = -1;
    if (isDraggingZoom) isDraggingZoom = false;
    if (isDraggingPan) isDraggingPan = false;
  }

  function clearZoom() {
    zoomStartIdx = 0;
    zoomEndIdx = data.length - 1;
    isZoomed = false;
  }

  function shiftClick() {
    if (hoveredIndex < 0) return;
    const pt = data[hoveredIndex];
    const exportedValue = `TIMESTAMP '${removeLocalTimezoneOffset(
      dateAt(pt),
      timeGrainToDuration(rollupTimeGrain),
    ).toISOString()}'`;
    copyToClipboard(exportedValue);
  }
</script>

<div
  role="presentation"
  style:max-width="{width}px"
  on:click={modified({ shift: shiftClick })}
>
  <TimestampProfileSummary
    start={xExtentStart}
    end={xExtentEnd}
    {estimatedSmallestTimeGrain}
    {rollupTimeGrain}
  />
  <Tooltip location="right" alignment="middle" distance={32}>
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <svg
      role="img"
      {width}
      {height}
      style:cursor
      on:mousedown={handleMouseDown}
      on:mousemove={handleMouseMove}
      on:mouseup={handleMouseUp}
      on:mouseleave={handleMouseLeave}
      on:contextmenu|preventDefault|stopPropagation
    >
      <defs>
        <clipPath id="clip-{id}">
          <rect
            x={plotLeft}
            y={plotTop}
            width={plotRight - plotLeft}
            height={plotBottom - plotTop}
          />
        </clipPath>
        <linearGradient id="left-side-{id}">
          <stop
            offset="0%"
            stop-color="var(--surface-background)"
            stop-opacity="1"
          />
          <stop
            offset="100%"
            stop-color="var(--surface-background)"
            stop-opacity="0"
          />
        </linearGradient>
        <linearGradient id="right-side-{id}">
          <stop
            offset="0%"
            stop-color="var(--surface-background)"
            stop-opacity="0"
          />
          <stop
            offset="100%"
            stop-color="var(--surface-background)"
            stop-opacity="1"
          />
        </linearGradient>
      </defs>

      <g clip-path="url(#clip-{id})">
        <TimeSeriesChart series={chartSeries} {scales} />

        {#if showSmoothed}
          <path
            d={smoothedPath}
            class="stroke-fg-disabled"
            fill="none"
            stroke-width={3}
            style:opacity={0.5}
          />
          <path
            d={smoothedPath}
            class="stroke-fg-muted"
            fill="none"
            stroke-width={1.5}
            style:opacity={0.85}
          />
        {/if}

        {#if isZoomed}
          <rect
            transition:fade|global
            x={plotLeft}
            y={plotTop}
            width={20}
            height={plotBottom - plotTop}
            fill="url(#left-side-{id})"
          />
          <rect
            transition:fade|global
            x={plotRight - 20}
            y={plotTop}
            width={20}
            height={plotBottom - plotTop}
            fill="url(#right-side-{id})"
          />
        {/if}

        <line
          x1={plotLeft}
          x2={plotRight}
          y1={yScale(0)}
          y2={yScale(0)}
          class="stroke-fg-muted"
        />
      </g>

      {#if isDraggingZoom}
        <rect
          x={Math.min(dragStartX, dragCurrentX)}
          y={plotTop}
          width={Math.abs(dragCurrentX - dragStartX)}
          height={plotBottom - plotTop}
          fill={zoomWindowColor}
          opacity={0.3}
        />
        <line
          x1={dragStartX}
          x2={dragStartX}
          y1={plotTop}
          y2={plotBottom}
          class="stroke-fg-muted"
        />
        <line
          x1={dragCurrentX}
          x2={dragCurrentX}
          y1={plotTop}
          y2={plotBottom}
          class="stroke-fg-muted"
        />
      {/if}

      {#if hoveredIndex >= 0 && !isDragging}
        {@const pt = data[hoveredIndex]}
        {@const cx = xScale(hoveredIndex)}
        {@const cy = yScale(val(pt))}
        {@const label = removeLocalTimezoneOffset(
          dateAt(pt),
          timeGrainToDuration(rollupTimeGrain),
        )}
        <line
          x1={cx}
          x2={cx}
          y1={plotTop}
          y2={plotBottom}
          class="stroke-fg-muted"
        />
        <circle {cx} {cy} r={3} class="fill-fg-muted" />
        <g
          in:fly|global={{ duration: 200, x: -16 }}
          out:fly|global={{ duration: 200, x: -16 }}
          font-size={fontSize}
          style:user-select="none"
        >
          <text
            x={plotLeft}
            y={fontSize}
            class="fill-fg-secondary text-outline"
          >
            {datePortion(label)}
          </text>
          <text
            x={plotLeft}
            y={fontSize * 2 + textGap}
            class="fill-fg-secondary text-outline"
          >
            {timePortion(label)}
          </text>
          <text
            x={plotLeft}
            y={fontSize * 3 + textGap * 2}
            class="fill-fg-secondary text-outline"
          >
            {formatInteger(Math.trunc(val(pt)))} row{val(pt) !== 1 ? "s" : ""}
          </text>
        </g>
      {/if}

      {#if isZoomed}
        <text
          role="button"
          tabindex="0"
          font-size={fontSize}
          x={plotRight}
          y={fontSize}
          text-anchor="end"
          style:user-select="none"
          style:cursor="pointer"
          class="transition-color fill-fg-muted hover:fill-fg-primary text-outline"
          in:fly|global={{ duration: 200, x: 16, delay: 200 }}
          out:fly|global={{ duration: 200, x: 16 }}
          on:keydown={() => {}}
          on:click={clearZoom}
        >
          clear zoom &#x2716;
        </text>
      {/if}
    </svg>

    <div
      slot="tooltip-content"
      in:fly|global={{ duration: 100, y: 4 }}
      out:fly|global={{ duration: 100, y: 4 }}
      style="
        display: grid;
        justify-content: center;
        grid-template-columns: max-content;"
    >
      <TimestampTooltipContent
        data={spark}
        width={tooltipSparkWidth}
        height={tooltipSparkHeight}
        tooltipPanShakeAmount={0}
        {zoomedRows}
        {totalRows}
        zoomed={isDraggingZoom || isZoomed}
        zooming={isZoomed && !isDraggingZoom}
        zoomWindowXMin={isDraggingZoom || isZoomed ? zoomMinDate : undefined}
        zoomWindowXMax={isDraggingZoom || isZoomed ? zoomMaxDate : undefined}
      />
    </div>
  </Tooltip>

  <div class="select-none grid grid-cols-2 space-between">
    <TimestampBound
      grain={rollupTimeGrain}
      align="left"
      value={zoomMinDate}
      label="Min"
    />
    <TimestampBound
      grain={rollupTimeGrain}
      align="right"
      value={zoomMaxDate}
      label="Max"
    />
  </div>
</div>

<style>
  text {
    user-select: none;
  }
</style>
