<!-- @component 
A specialized line component that solves a few problems:
1. Tweening between arrays of data that have different lengths
2. Tweening a time series that has gaps in it

I's a re-implementation of Peter Beshai's d3-line-chunked plugin (https://github.com/pbeshai/d3-line-chunked)
which solves a fairly similar set of problems utilizing d3-select.

Use this component when you're rendering a dynamically changing line chart or spark.

Over time, we'll make this the default Line implementation, but it's not quite there yet.
-->
<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import {
    WithDelayedValue,
    WithTween,
  } from "@rilldata/web-common/components/data-graphic/functional-components";
  import { computeSegments } from "@rilldata/web-common/components/data-graphic/marks/segment";
  import type { ScaleStore } from "@rilldata/web-common/components/data-graphic/state/types";
  import {
    areaFactory,
    lineFactory,
  } from "@rilldata/web-common/components/data-graphic/utils";
  import { LineMutedColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { interpolatePath } from "d3-interpolate-path";
  import { getContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  export let data;
  export let xAccessor: string;
  export let yAccessor: string;

  export let isComparingDimension = false;

  /** time in ms to trigger a delay when the underlying data changes */
  export let delay = 0;
  export let duration = 400;

  export let stopOpacity = 0.3;
  // FIXME – this is a different prop than elsewhere
  export let lineColor = LineMutedColor;
  export let areaGradientColors: [string, string] | null = null;

  $: area = areaGradientColors !== null;

  const id = guidGenerator();

  // get the scale functions from the data graphic
  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;

  let lineFunction;
  let areaFunction;

  const curveType = "curveLinear";

  // FIXME:
  $: if ($xScale && $yScale) {
    lineFunction = lineFactory({
      curve: curveType,
      xScale: $xScale,
      // the y should always plot a line segment so that we can
      // smoothly tween
      yScale: (d) => $yScale(d || 0),
      xAccessor,
      // path should always be defined for the chunked line, since
      // we will clip the path to the segments before
      // it reaches zero.
      pathDefined: () => true,
    });
    if (area) {
      areaFunction = areaFactory({
        curve: curveType,
        xScale: $xScale,
        yScale: (d) => $yScale(d || 0),
        xAccessor,
        // path should always be defined for the chunked line, since
        // we will clip the path to the segments before
        // it reaches zero.
        pathDefined: () => true,
      });
    }
  }

  $: segments = computeSegments(data, yAccessor);
  /** plot these as points */
  $: singletons = segments.filter((segment) => segment.length === 1);

  /** use this line thickness heuristic to allow some amount of overplotting
   * FIXME: this needs refinement!
   */
  // let lineThickness = createAdaptiveLineThicknessStore(yAccessor);
  // $: lineThickness.setData(data);
</script>

<WithDelayedValue
  {delay}
  value={[data, segments, singletons]}
  let:output={delayedValues}
>
  {@const delayedFilteredData = delayedValues[0]}
  {@const delayedSegments = delayedValues[1]}
  {@const delayedSingletons = delayedValues[2]}
  {#each delayedSingletons as [singleton]}
    <rect
      x={$xScale(singleton[xAccessor]) - 0.75}
      y={Math.min($yScale(0), $yScale(singleton[yAccessor]))}
      width="1.5"
      height={Math.abs($yScale(0) - $yScale(singleton[yAccessor]))}
      fill={lineColor}
    />
    <circle
      cx={$xScale(singleton[xAccessor])}
      cy={$yScale(singleton[yAccessor])}
      r="1.5"
      fill={lineColor}
    />
  {/each}
  <g>
    <WithTween
      value={lineFunction(yAccessor)(delayedFilteredData)}
      tweenProps={{
        duration,
        interpolate: interpolatePath,
        easing: cubicOut,
      }}
      let:output={dt}
    >
      <!-- line -->
      <path
        stroke-width={isComparingDimension ? 1.5 : 1}
        stroke={lineColor}
        d={dt}
        id="segments-line"
        fill="none"
        style="clip-path: url(#path-segments-{id})"
      />
    </WithTween>
    {#if areaGradientColors !== null}
      <WithTween
        value={areaFunction(yAccessor)(delayedFilteredData)}
        tweenProps={{
          duration,
          interpolate: interpolatePath,
          easing: cubicOut,
        }}
        let:output={at}
      >
        <path
          d={at}
          fill="url(#gradient-{id})"
          style="clip-path: url(#path-segments-{id})"
        />
      </WithTween>
      <defs>
        <linearGradient id="gradient-{id}" x1="0" x2="0" y1="0" y2="1">
          <stop
            offset="5%"
            stop-color={areaGradientColors[0]}
            stop-opacity={stopOpacity}
          />
          <stop
            offset="95%"
            stop-color={areaGradientColors[1]}
            stop-opacity={stopOpacity}
          />
        </linearGradient>
      </defs>
    {/if}
    <!-- clip rects for segments -->
    <defs>
      <clipPath id="path-segments-{id}">
        {#each delayedSegments as segment (segment[0][xAccessor])}
          {@const x = $xScale(segment[0][xAccessor])}
          {@const width =
            $xScale(segment.at(-1)[xAccessor]) - $xScale(segment[0][xAccessor])}
          <rect {x} y={0} height={$yScale.range()[0]} {width} />
        {/each}
      </clipPath>
    </defs>
  </g>
</WithDelayedValue>
