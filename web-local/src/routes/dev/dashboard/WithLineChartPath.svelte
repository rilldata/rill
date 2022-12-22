<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import type { ScaleStore } from "@rilldata/web-common/components/data-graphic/state/types";
  import {
    areaFactory,
    lineFactory,
    pathIsDefined,
  } from "@rilldata/web-common/components/data-graphic/utils";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { interpolatePath } from "d3-interpolate-path";
  import { getContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { computeSegments, gapsFromSegments } from "./measure-chart/utils";
  import WithDelayedValue from "./WithDelayedValue.svelte";
  export let data;
  export let xAccessor: string;
  export let yAccessor: string;
  export let delay = 0;
  export let duration = 400;

  const id = guidGenerator();

  // get the scale functions from the data graphic
  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;

  let lineFunction;
  let areaFunction;
  $: if ($xScale && $yScale) {
    lineFunction = lineFactory({
      xScale: $xScale,
      yScale: $yScale,
      xAccessor,
      pathDefined: pathIsDefined(yAccessor),
    });
    areaFunction = areaFactory({
      xScale: $xScale,
      yScale: $yScale,
      xAccessor,
      pathDefined: pathIsDefined(yAccessor),
    });
  }

  $: segments = computeSegments(data, pathIsDefined(yAccessor));
  $: gaps = gapsFromSegments(segments);

  $: filteredData = data.filter(pathIsDefined(yAccessor));

  export function zoomOut(
    node,
    { delay = 0, duration = 400, easing = cubicOut, x = 0, y = 0, opacity = 0 }
  ) {
    const style = getComputedStyle(node);
    const target_opacity = +style.opacity;
    const transform = style.transform === "none" ? "" : style.transform;

    const od = target_opacity * (1 - opacity);

    return {
      delay,
      duration,
      easing,
      css: (t, u) => `
			transform: ${transform} translate(${(1 - t) * x}px, ${
        (1 - t) * y
      }px) scale({t});
			opacity: ${target_opacity - od * u}`,
    };
  }
</script>

<WithDelayedValue
  {delay}
  value={[filteredData, segments]}
  let:output={delayedValues}
>
  {@const delayedFilteredData = delayedValues[0]}
  <!-- {@const delayedGaps = delayedValues[1]} -->
  {@const delayedSegments = delayedValues[1]}
  <g>
    <!-- gap line -->
    <!-- {#each delayedGaps as [start, end] (start[xAccessor] + end[xAccessor])}
      <WithTween
        value={{ start, end }}
        tweenProps={{ duration, easing: cubicOut }}
        let:output
      >
        <line
          x1={$xScale(output.start[xAccessor])}
          x2={$xScale(output.end[xAccessor])}
          y1={$yScale(output.start[yAccessor])}
          y2={$yScale(output.end[yAccessor])}
          stroke="red"
          stroke-width="4"
        />
      </WithTween>
    {/each} -->
    {#if false}
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
          stroke="hsl(217,50%,60%)"
          fill="none"
          opacity="1"
          stroke-width=".2px"
          d={dt}
          id="gap-line-{id}"
          stroke-dasharray="1,2"
        />
      </WithTween>
    {/if}
    <!-- segments with actual ata -->
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
        stroke-width="1px"
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
        fill="hsla(217,100%, 50%, 0.1)"
        style="clip-path: url(#path-segments-{id})"
      />
    </WithTween>

    <!-- clip rects for segments -->
    <defs>
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
