<script lang="ts">
  /**
   * TimestampSpark.svelte
   * ---------------------
   * This simple component is a basic sparkline, meant to be used
   * in a table / model profile preview.
   * It optionally enables the user to determine a "window", which
   * is just a box emcompassing the zoomWindowXMin and zoomWindowXMax values.
   */
  import { guidGenerator } from "$lib/util/guid";
  import { fade } from "svelte/transition";
  import { cubicOut as easing } from "svelte/easing";
  import { scaleLinear } from "d3-scale";
  import { extent } from "d3-array";
  import { writable } from "svelte/store";
  import { createExtremumResolutionStore } from "$lib/components/data-graphic/extremum-resolution-store";
  import { lineFactory, areaFactory } from "$lib/components/data-graphic/utils";
  import { tweened } from "svelte/motion";

  const plotID = guidGenerator();

  export let data;

  export let width = 360;
  export let height = 120;
  export let curve = "curveLinear";
  export let area = false;
  export let color = "hsl(217, 10%, 50%)";
  export let tweenIn = false;

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

  export let buffer = 4;
  export let leftBuffer = buffer;
  export let rightBuffer = buffer;
  export let topBuffer = buffer;
  export let bottomBuffer = buffer;

  const X = writable(undefined);
  const Y = writable(undefined);

  $: plotTop = top + topBuffer;
  $: plotBottom = height - bottomBuffer - bottom;
  $: plotLeft = left + leftBuffer;
  $: plotRight = width - rightBuffer - buffer;

  // establish basis values
  let xExtents = extent(data, (d) => d[xAccessor]);
  $: xExtents = extent(data, (d) => d[xAccessor]);

  const xMin = createExtremumResolutionStore(xExtents[0], {
    duration: 300,
    easing,
    direction: "min",
  });
  const xMax = createExtremumResolutionStore(xExtents[1], {
    duration: 300,
    easing,
  });

  $: xMin.setWithKey("x", xExtents[0]);
  $: xMax.setWithKey("x", xExtents[1]);

  // Let's set the X Scale based on the $xMin and $xMax.
  $: $X = scaleLinear()
    .domain([$xMin, $xMax])
    .range([left, width - right]);

  // Generate our Y Scale.
  let yExtents = extent(data, (d) => d[yAccessor]);
  $: yExtents = extent(data, (d) => d[yAccessor]);
  const yMax = createExtremumResolutionStore(Math.max(5, yExtents[1]));

  /** Listen ~ the world needs a little bit of joy. If the user wants to tween in the height
   * of the graph so it looks like it grows, then let them have it.
   * This tweened value is consumed only if the consumer sets tweenIn={true}.
   */
  const tweenInValue = tweened(height, { duration: 600, easing });
  $: tweenInValue.set(plotTop);

  /** we will tween in the upper part of the range if the consumer of the component
   * sets tweenIn={true}. Otherwise This sparkline will just appear.
   */
  $: $Y = scaleLinear()
    .domain([0, $yMax])
    .range([plotBottom, tweenIn ? $tweenInValue : plotTop]);

  $: lineFcn = lineFactory({
    xScale: $X,
    yScale: $Y,
    curve,
    xAccessor,
  });

  $: areaFcn = areaFactory({
    xScale: $X,
    yScale: $Y,
    curve,
    xAccessor,
  });

  // zoom window scrubbing, if used
  $: zoomPreviewXScale = scaleLinear().domain($X.domain()).range([0, width]);
  $: zoomPreviewX = zoomPreviewXScale(zoomWindowXMin);
  $: zoomPreviewWidth = zoomPreviewXScale(zoomWindowXMax) - zoomPreviewX;
</script>

<svg {width} {height}>
  <clipPath id="data-graphic-{plotID}">
    <rect
      x={plotLeft}
      y={plotTop}
      width={plotRight - plotLeft}
      height={plotBottom - plotTop}
    />
  </clipPath>
  <!-- core geoms -->
  <g clip-path="url(#data-graphic-{plotID})">
    {#if area}
      <path d={areaFcn(yAccessor)(data)} fill={color} />
    {/if}

    <path
      d={lineFcn(yAccessor)(data)}
      stroke={color}
      stroke-width={0.2}
      fill="none"
      style:opacity={1}
    />
    <line
      x1={$X.range()[0]}
      x2={$X.range()[1]}
      y1={plotBottom}
      y2={plotBottom}
      stroke={color}
    />
    {#if zoomPreviewWidth}
      <g transition:fade={{ duration: 100 }}>
        <rect
          x={zoomPreviewX}
          y={plotTop}
          width={zoomPreviewWidth}
          {height}
          fill={zoomWindowColor}
          opacity=".9"
          style:mix-blend-mode="lighten"
        />
        <line
          x1={zoomPreviewX}
          x2={zoomPreviewX}
          y1={plotTop}
          y2={plotBottom}
          stroke={zoomWindowBoundaryColor}
        />
        <line
          x1={zoomPreviewX + zoomPreviewWidth}
          x2={zoomPreviewX + zoomPreviewWidth}
          y1={plotTop}
          y2={plotBottom}
          stroke={zoomWindowBoundaryColor}
        />
      </g>
    {/if}
  </g>
</svg>
