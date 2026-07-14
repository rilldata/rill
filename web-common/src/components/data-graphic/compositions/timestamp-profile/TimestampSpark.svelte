<script lang="ts">
  import { extent } from "d3-array";
  import { scaleLinear } from "d3-scale";
  import { scaleTime } from "d3-scale";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import {
    createLineGenerator,
    createAreaGenerator,
    pathDoesNotDropToZero,
  } from "../../utils";
  import type { TimestampDataPoint } from "@rilldata/web-common/features/column-profile/queries";

  const gradientId = `spark-gradient-${guidGenerator()}`;

  export let data: TimestampDataPoint[];
  export let color = "var(--color-teal-700)";
  export let zoomWindowXMin: Date | undefined = undefined;
  export let zoomWindowXMax: Date | undefined = undefined;
  export let width = 360;
  export let height = 120;
  export let left = 0;
  export let right = 0;
  export let top = 12;
  export let bottom = 4;

  $: plotLeft = left;
  $: plotRight = width - right;
  $: plotTop = top;
  $: plotBottom = height - bottom;

  $: [xMinVal, xMaxVal] = extent(data, (d) => d.ts);
  $: [, yMaxVal] = extent(data, (d) => d.count);

  $: xScale = scaleTime()
    .domain([xMinVal ?? new Date(), xMaxVal ?? new Date()])
    .range([plotLeft, plotRight]);

  $: yScale = scaleLinear()
    .domain([0, yMaxVal ?? 1])
    .range([plotBottom, plotTop]);

  $: lineGen = createLineGenerator<TimestampDataPoint>({
    x: (d) => xScale(d.ts ?? 0),
    y: (d) => yScale(d.count),
    defined: pathDoesNotDropToZero("count"),
  });

  $: areaGen = createAreaGenerator<TimestampDataPoint>({
    x: (d) => xScale(d.ts ?? 0),
    y0: yScale(0),
    y1: (d) => yScale(d.count),
    defined: pathDoesNotDropToZero("count"),
  });

  $: linePath = lineGen(data);
  $: areaPath = areaGen(data);
</script>

{#if data.length}
  <svg class="overflow-visible" {width} {height}>
    <defs>
      <linearGradient id={gradientId} x1="0" x2="0" y1="0" y2="1">
        <stop offset="5%" stop-color={color} />
        <stop offset="95%" stop-color={color} stop-opacity={0.3} />
      </linearGradient>
    </defs>
    <g>
      {#if linePath}
        <path d={linePath} stroke={color} stroke-width={0.5} fill="none" />
      {/if}
      {#if areaPath}
        <path d={areaPath} fill="url(#{gradientId})" />
      {/if}
    </g>
    {#if zoomWindowXMin && zoomWindowXMax}
      <g>
        <rect
          x={xScale(zoomWindowXMin)}
          y={plotTop}
          width={xScale(zoomWindowXMax) - xScale(zoomWindowXMin)}
          height={plotBottom - plotTop}
          class="fill-gray-200/50"
          opacity=".9"
          style:mix-blend-mode="lighten"
        />
        <line
          x1={xScale(zoomWindowXMin)}
          x2={xScale(zoomWindowXMin)}
          y1={plotTop}
          y2={plotBottom}
          class="stroke-gray-300"
        />
        <line
          x1={xScale(zoomWindowXMax)}
          x2={xScale(zoomWindowXMax)}
          y1={plotTop}
          y2={plotBottom}
          class="stroke-gray-300"
        />
      </g>
    {/if}
  </svg>
{/if}
