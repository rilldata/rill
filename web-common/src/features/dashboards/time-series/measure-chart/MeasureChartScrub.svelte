<script lang="ts">
  import type { ChartScales, ChartConfig } from "./types";
  import {
    ScrubArea0Color,
    ScrubArea1Color,
    ScrubArea2Color,
    ScrubBoxColor,
  } from "../chart-colors";

  const chartId = Math.random().toString(36).slice(2, 9);
  const strokeWidth = 1;
  const xLabelBuffer = 8;
  const yLabelBuffer = 10;

  export let scales: ChartScales;
  export let config: ChartConfig;
  /** Scrub start/end as fractional indices */
  export let startIndex: number | null;
  export let endIndex: number | null;
  export let isScrubbing: boolean = false;
  export let showLabels: boolean = false;
  export let onReset: () => void;
  export let formatLabel: (index: number) => string = (i) => String(i);

  $: y1 = config.plotBounds.top + 5;
  $: y2 = config.plotBounds.bottom - 5;

  $: hasSelection = startIndex !== null && endIndex !== null;

  $: orderedStartIdx =
    startIndex !== null && endIndex !== null
      ? Math.min(startIndex, endIndex)
      : null;
  $: orderedEndIdx =
    startIndex !== null && endIndex !== null
      ? Math.max(startIndex, endIndex)
      : null;

  $: xStart = orderedStartIdx !== null ? scales.x(orderedStartIdx) : 0;
  $: xEnd = orderedEndIdx !== null ? scales.x(orderedEndIdx) : 0;
  $: selectionWidth = Math.abs(xEnd - xStart);

  function handleContextMenu(event: MouseEvent) {
    event.preventDefault();
    if (hasSelection) onReset();
  }
</script>

<defs>
  <linearGradient
    gradientUnits="userSpaceOnUse"
    id="scrubbing-gradient-{chartId}"
  >
    <stop stop-color={ScrubArea0Color} />
    <stop offset="0.36" stop-color={ScrubArea1Color} />
    <stop offset="1" stop-color={ScrubArea2Color} />
  </linearGradient>
</defs>

{#if hasSelection && orderedStartIdx !== null && orderedEndIdx !== null}
  <g class="scrub-group" on:contextmenu={handleContextMenu} role="presentation">
    <rect
      class="selection-rect"
      class:scrubbing={isScrubbing}
      x={xStart}
      y={y1}
      width={selectionWidth}
      height={y2 - y1}
      fill="url(#scrubbing-gradient-{chartId})"
      opacity={isScrubbing ? 0.4 : 0.2}
    />

    <line
      x1={xStart}
      x2={xStart}
      {y1}
      {y2}
      stroke={ScrubBoxColor}
      stroke-width={strokeWidth}
    />
    <line
      x1={xEnd}
      x2={xEnd}
      {y1}
      {y2}
      stroke={ScrubBoxColor}
      stroke-width={strokeWidth}
    />

    {#if showLabels}
      <text
        class="scrub-label"
        text-anchor="end"
        x={xStart - xLabelBuffer}
        y={y1 + yLabelBuffer}
      >
        {formatLabel(Math.round(orderedStartIdx))}
      </text>
      <circle
        cx={xStart}
        cy={y1}
        r={3}
        paint-order="stroke"
        class="fill-primary-700 stroke-surface-background"
        stroke-width="3"
      />

      <text
        class="scrub-label"
        text-anchor="start"
        x={xEnd + xLabelBuffer}
        y={y1 + yLabelBuffer}
      >
        {formatLabel(Math.round(orderedEndIdx))}
      </text>
      <circle
        cx={xEnd}
        cy={y1}
        r={3}
        paint-order="stroke"
        class="fill-primary-700 stroke-surface-background"
        stroke-width="3"
      />
    {/if}

    <rect
      class="resize-handle"
      x={xStart - 5}
      y={y1}
      width={10}
      height={y2 - y1}
      fill="transparent"
    />
    <rect
      class="resize-handle"
      x={xEnd - 5}
      y={y1}
      width={10}
      height={y2 - y1}
      fill="transparent"
    />
  </g>
{/if}

<style>
  .scrub-group {
    transition: opacity ease 0.3s;
  }

  .selection-rect.scrubbing {
    filter: drop-shadow(0px 4px 6px rgba(0, 0, 0, 0.1))
      drop-shadow(0px 10px 15px rgba(0, 0, 0, 0.2));
  }

  .scrub-label {
    font-size: 11px;
    fill: var(--color-gray-600);
  }

  .resize-handle {
    cursor: ew-resize;
  }
</style>
