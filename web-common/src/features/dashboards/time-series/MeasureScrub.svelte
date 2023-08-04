<script lang="ts">
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import { getContext } from "svelte";
  import { type PlotConfig } from "@rilldata/web-common/components/data-graphic/utils";
  import type { Writable } from "svelte/store";

  export let start;
  export let stop;
  export let isScrubbing = false;
  export let showLabels = false;
  export let mouseoverTimeFormat;

  const plotConfig: Writable<PlotConfig> = getContext(
    "rill:data-graphic:plot-config"
  );

  const strokeWidth = 1;
  const xLabelBuffer = 8;
  const yLabelBuffer = 10;
  const y1 = $plotConfig.plotTop + $plotConfig.top + 5;
  const y2 = $plotConfig.plotBottom - $plotConfig.bottom - 1;
</script>

{#if start && stop}
  <WithGraphicContexts let:xScale let:yScale>
    {@const xStart = xScale(Math.min(start, stop))}
    {@const xEnd = xScale(Math.max(start, stop))}
    <g>
      {#if showLabels}
        <text text-anchor="end" x={xStart - xLabelBuffer} y={y1 + yLabelBuffer}>
          {mouseoverTimeFormat(Math.min(start, stop))}
        </text>
        <circle
          cx={xStart}
          cy={y1}
          r={3}
          paint-order="stroke"
          class="fill-blue-700"
          stroke="white"
          stroke-width="3"
        />
        <text text-anchor="start" x={xEnd + xLabelBuffer} y={y1 + yLabelBuffer}>
          {mouseoverTimeFormat(Math.max(start, stop))}
        </text>
        <circle
          cx={xEnd}
          cy={y1}
          r={3}
          paint-order="stroke"
          class="fill-blue-700"
          stroke="white"
          stroke-width="3"
        />
      {/if}
      <line
        x1={xStart}
        x2={xStart}
        {y1}
        {y2}
        stroke="#60A5FA"
        stroke-width={strokeWidth}
      />
      <line
        x1={xEnd}
        x2={xEnd}
        {y1}
        {y2}
        stroke="#60A5FA"
        stroke-width={strokeWidth}
      />
    </g>
    <g opacity={isScrubbing ? "0.4" : "0.3"}>
      <rect
        class:rect-shadow={isScrubbing}
        x={Math.min(xStart, xEnd)}
        y={y1}
        width={Math.abs(xStart - xEnd)}
        height={y2 - y1}
        fill="url('#scrubbing-gradient')"
      />
    </g>
  </WithGraphicContexts>
{/if}

<defs>
  <linearGradient id="scrubbing-gradient" gradientUnits="userSpaceOnUse">
    <stop stop-color="#558AFF" />
    <stop offset="0.36" stop-color="#4881FF" />
    <stop offset="1" stop-color="#2563EB" />
  </linearGradient>
</defs>

<style>
  .rect-shadow {
    filter: drop-shadow(0px 4px 6px rgba(0, 0, 0, 0.1))
      drop-shadow(0px 10px 15px rgba(0, 0, 0, 0.2));
  }

  g {
    transition: opacity ease 0.4s;
  }
</style>
