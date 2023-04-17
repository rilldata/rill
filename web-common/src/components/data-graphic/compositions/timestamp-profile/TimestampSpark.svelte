<script lang="ts">
  import { cubicOut } from "svelte/easing";
  import { fade } from "svelte/transition";
  /**
   * TimestampSpark.svelte
   * ---------------------
   * This simple component is a basic sparkline, meant to be used
   * in a table / model profile preview.
   * It optionally enables the user to determine a "window", which
   * is just a box emcompassing the zoomWindowXMin and zoomWindowXMax values.
   */
  import { SimpleDataGraphic } from "../../elements";
  import { Area, Line } from "../../marks";

  export let data;

  export let width = 360;
  export let height = 120;
  export let color = "hsl(217, 10%, 50%)";
  export let areaColor = color;
  export let stopOpacity: number = undefined;

  // the color of the zoom window
  export let zoomWindowColor = "hsla(217, 90%, 60%, .2)";
  // the color of the zoom window boundaries
  export let zoomWindowBoundaryColor = "rgb(100,100,100)";
  export let zoomWindowXMin: Date = undefined;
  export let zoomWindowXMax: Date = undefined;

  export let xAccessor: string = undefined;
  export let yAccessor: string = undefined;

  // rowsize for table
  export let left = 0;
  export let right = 0;
  export let top = 12;
  export let bottom = 4;

  export function scaleVertical(
    node: Element,
    {
      delay = 0,
      duration = 400,
      easing = cubicOut,
      start = 0,
      opacity = 0,
      scaleDown = false,
    } = {}
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
      css: (_t, u) => {
        const yScale = scaleDown ? ` scaleY(${1 - sd * u})` : "";
        return `
    transform: ${transform} scaleY(${1 - sd * u}) ${yScale};
    transform-origin: 100% calc(100% - ${0}px);
    opacity: ${target_opacity - od * u}
  `;
      },
    };
  }
</script>

{#if data.length}
  <SimpleDataGraphic
    xType="date"
    yType="number"
    {width}
    {height}
    yMin={0}
    bodyBuffer={0}
    marginBuffer={0}
    {left}
    {right}
    {top}
    {bottom}
    shareXScale={false}
    shareYScale={false}
    let:xScale
    let:config
  >
    <g transition:scaleVertical|local={{ duration: 400, start: 0.3 }}>
      <Line {data} {xAccessor} {yAccessor} {color} lineThickness={0.5} />
      <Area {data} {xAccessor} {yAccessor} color={areaColor} {stopOpacity} />
    </g>
    <!-- show zoom boundaries -->
    {#if zoomWindowXMin && zoomWindowXMax}
      <g transition:fade|local={{ duration: 100 }}>
        <rect
          x={xScale(zoomWindowXMin)}
          y={config.plotTop}
          width={xScale(zoomWindowXMax) - xScale(zoomWindowXMin)}
          {height}
          fill={zoomWindowColor}
          opacity=".9"
          style:mix-blend-mode="lighten"
        />
        <line
          x1={xScale(zoomWindowXMin)}
          x2={xScale(zoomWindowXMin)}
          y1={config.plotTop}
          y2={config.plotBottom}
          stroke={zoomWindowBoundaryColor}
        />
        <line
          x1={xScale(zoomWindowXMax)}
          x2={xScale(zoomWindowXMax)}
          y1={config.plotTop}
          y2={config.plotBottom}
          stroke={zoomWindowBoundaryColor}
        />
      </g>
    {/if}
  </SimpleDataGraphic>
{/if}
