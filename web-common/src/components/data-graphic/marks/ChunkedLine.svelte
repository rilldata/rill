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
    createAdaptiveLineThicknessStore,
    lineFactory,
    pathIsDefined,
  } from "@rilldata/web-common/components/data-graphic/utils";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { interpolatePath } from "d3-interpolate-path";
  import { getContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  export let data;
  export let xAccessor: string;
  export let yAccessor: string;
  /** time in ms to trigger a delay when the underlying data changes */
  export let delay = 0;
  export let duration = 400;

  export let key;

  export let stopOpacity = 0.3;
  // FIXME – this is a different prop than elsewhere
  export let color = "hsla(217,70%, 80%, .4)";

  const id = guidGenerator();

  // get the scale functions from the data graphic
  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;

  let lineFunction;
  let areaFunction;
  // FIXME:
  $: if ($xScale && $yScale) {
    lineFunction = lineFactory({
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
    areaFunction = areaFactory({
      xScale: $xScale,
      yScale: (d) => $yScale(d || 0),
      xAccessor,
      // path should always be defined for the chunked line, since
      // we will clip the path to the segments before
      // it reaches zero.
      pathDefined: () => true,
    });
  }

  $: segments = computeSegments(data, pathIsDefined(yAccessor));
  /** plot these as points */
  $: singletons = segments.filter((segment) => segment.length === 1);

  /** use this line thickness heuristic to allow some amount of overplotting
   * FIXME: this needs refinement!
   */
  let lineThickness = createAdaptiveLineThicknessStore(yAccessor);
  $: lineThickness.setData(data);
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
      fill="hsla(217,60%, 55%, .5)"
    />
    <circle
      cx={$xScale(singleton[xAccessor])}
      cy={$yScale(singleton[yAccessor])}
      r="1.5"
      fill="hsla(217,60%, 55%, .5)"
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
      <path
        stroke-width={$lineThickness}
        stroke="hsla(217,60%, 55%, 1)"
        d={dt}
        id="segments-line"
        fill="none"
        style="clip-path: url(#path-segments-{id})"
      />
    </WithTween>

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

    <!-- clip rects for segments -->
    <defs>
      <linearGradient id="gradient-{id}" x1="0" x2="0" y1="0" y2="1">
        <stop offset="5%" stop-color={color} />
        <stop offset="95%" stop-color={color} stop-opacity={stopOpacity} />
      </linearGradient>
      <clipPath id="path-segments-{id}">
        {#each delayedSegments as segment (segment[0][xAccessor])}
          {@const x = $xScale(segment[0][xAccessor])}
          {@const width =
            $xScale(segment.at(-1)[xAccessor]) - $xScale(segment[0][xAccessor])}
          <WithTween
            initialValue={{
              x: x - width / 2,
              width: width * 2,
            }}
            value={{
              x,
              width,
            }}
            tweenProps={{
              duration,
              easing: cubicOut,
            }}
            let:output
          >
            <rect
              x={output.x}
              y={0}
              height={$yScale.range()[0]}
              width={output.width}
            />
          </WithTween>
        {/each}
      </clipPath>
    </defs>
  </g>
</WithDelayedValue>
