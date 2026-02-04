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
  import { cubicOut } from "svelte/easing";
  import { fade } from "svelte/transition";
  import type { TimestampDataPoint } from "@rilldata/web-common/features/column-profile/queries";

  const gradientId = `spark-gradient-${guidGenerator()}`;

  export let data: TimestampDataPoint[];
  export let width = 360;
  export let height = 120;
  export let color = "hsl(217, 10%, 50%)";
  export let zoomWindowColor = "hsla(217, 90%, 60%, .2)";
  export let zoomWindowBoundaryColor = "rgb(100,100,100)";
  export let zoomWindowXMin: Date | undefined = undefined;
  export let zoomWindowXMax: Date | undefined = undefined;
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

  function scaleVertical(
    node: Element,
    {
      delay = 0,
      duration = 400,
      easing = cubicOut,
      start = 0,
      opacity = 0,
    } = {},
  ) {
    const style = getComputedStyle(node);
    const target_opacity = +style.opacity;
    const transform = style.transform === "none" ? "" : style.transform;

    const sd = 1 - start;
    const od = target_opacity * (1 - opacity);

    return {
      delay,
      duration,
      easing,
      css: (_t: number, u: number) => {
        return `
    transform: ${transform} scaleY(${1 - sd * u});
    transform-origin: 100% calc(100% - ${0}px);
    opacity: ${target_opacity - od * u}
  `;
      },
    };
  }
</script>

{#if data.length}
  <svg class="overflow-visible" {width} {height}>
    <defs>
      <linearGradient id={gradientId} x1="0" x2="0" y1="0" y2="1">
        <stop offset="5%" stop-color={color} />
        <stop offset="95%" stop-color={color} stop-opacity={0.3} />
      </linearGradient>
    </defs>
    <g transition:scaleVertical={{ duration: 400, start: 0.3 }}>
      {#if linePath}
        <path d={linePath} stroke={color} stroke-width={0.5} fill="none" />
      {/if}
      {#if areaPath}
        <path d={areaPath} fill="url(#{gradientId})" />
      {/if}
    </g>
    {#if zoomWindowXMin && zoomWindowXMax}
      <g transition:fade={{ duration: 100 }}>
        <rect
          x={xScale(zoomWindowXMin)}
          y={plotTop}
          width={xScale(zoomWindowXMax) - xScale(zoomWindowXMin)}
          height={plotBottom - plotTop}
          fill={zoomWindowColor}
          opacity=".9"
          style:mix-blend-mode="lighten"
        />
        <line
          x1={xScale(zoomWindowXMin)}
          x2={xScale(zoomWindowXMin)}
          y1={plotTop}
          y2={plotBottom}
          stroke={zoomWindowBoundaryColor}
        />
        <line
          x1={xScale(zoomWindowXMax)}
          x2={xScale(zoomWindowXMax)}
          y1={plotTop}
          y2={plotBottom}
          stroke={zoomWindowBoundaryColor}
        />
      </g>
    {/if}
  </svg>
{/if}
