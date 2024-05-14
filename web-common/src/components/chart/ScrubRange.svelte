<script lang="ts">
  import {
    ScrubBoxColor,
    ScrubArea0Color,
    ScrubArea1Color,
    ScrubArea2Color,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";

  export let start: number;
  export let end: number;
  export let width: number;
  export let height: number;
  export let yMin: number;
  export let yMax: number;
  export let xMin: number;
  export let grayscale: boolean;
  export let grabbing: boolean;

  export let onAdjustScrubStart: () => void;
  export let onScrubGrab: () => void;
  export let onAdjustScrubEnd: () => void;
</script>

<line
  role="presentation"
  x1={start}
  x2={start}
  y1={yMin}
  y2={yMax}
  stroke={ScrubBoxColor}
  on:mousedown|stopPropagation={onAdjustScrubStart}
/>

<g style:mix-blend-mode="hue">
  {#if grayscale}
    <rect x={xMin} width="100%" height="100%" fill="white" />
  {/if}

  <rect
    role="presentation"
    class="cursor-grab"
    x={width < 0 ? end : start}
    y={yMin}
    width={Math.abs(width)}
    {height}
    fill="url(#scrubbing-gradient)"
    opacity="0.3"
    on:mousedown|stopPropagation={onScrubGrab}
  />
</g>

<line
  role="presentation"
  x1={end}
  x2={end}
  y1={yMin}
  y2={yMax}
  stroke={ScrubBoxColor}
  on:mousedown|stopPropagation={onAdjustScrubEnd}
/>

<defs>
  <linearGradient gradientUnits="userSpaceOnUse" id="scrubbing-gradient">
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

<style>
  line {
    cursor: ew-resize;
    stroke-width: 1px;
    vector-effect: non-scaling-stroke;
  }
</style>
