<script lang="ts">
  import { contexts } from "@rilldata/web-local/lib/components/data-graphic/constants";
  import { WithTween } from "@rilldata/web-local/lib/components/data-graphic/functional-components";
  import type { ScaleStore } from "@rilldata/web-local/lib/components/data-graphic/state/types";
  import {
    areaFactory,
    lineFactory,
    pathIsDefined,
  } from "@rilldata/web-local/lib/components/data-graphic/utils";
  import { interpolatePath } from "d3-interpolate-path";
  import { getContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  export let data;
  export let xAccessor: string;
  export let yAccessor: string;
  export let delay = 0;
  export let duration = 1000;

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

  /**
   * Helper function to compute the contiguous segments of the data
   * based on https://github.com/pbeshai/d3-line-chunked/blob/master/src/lineChunked.js
   */
  function computeSegments(lineData, defined, isNext = (prev, curr) => true) {
    let startNewSegment = true;

    // split into segments of continuous data
    const segments = lineData.reduce(function (segments, d) {
      // skip if this point has no data
      if (!defined(d)) {
        startNewSegment = true;
        return segments;
      }

      // if we are starting a new segment, start it with this point
      if (startNewSegment) {
        segments.push([d]);
        startNewSegment = false;

        // otherwise see if we are adding to the last segment
      } else {
        var lastSegment = segments[segments.length - 1];
        var lastDatum = lastSegment[lastSegment.length - 1];
        // if we expect this point to come next, add it to the segment
        if (isNext(lastDatum, d)) {
          lastSegment.push(d);

          // otherwise create a new segment
        } else {
          segments.push([d]);
        }
      }

      return segments;
    }, []);

    return segments;
  }

  $: segments = computeSegments(data, pathIsDefined(yAccessor));
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

<text x="30" y="60">{$yScale.domain()}</text>
<g>
  <!-- gap line -->
  <WithTween
    value={lineFunction(yAccessor)(filteredData)}
    tweenProps={{
      duration,
      interpolate: interpolatePath,
      easing: cubicOut,
      delay,
    }}
    let:output={dt}
  >
    <path
      stroke="hsl(217,50%,60%)"
      fill="none"
      opacity="1"
      stroke-width="1px"
      d={dt}
      id="gap-line"
      stroke-dasharray="1,2"
    />
  </WithTween>
  <!-- segments with actual ata -->
  <WithTween
    value={lineFunction(yAccessor)(filteredData)}
    tweenProps={{
      duration,
      interpolate: interpolatePath,
      easing: cubicOut,
      delay,
    }}
    let:output={dt}
  >
    <path
      stroke-width="1px"
      stroke="hsla(217,60%, 55%, 1)"
      d={dt}
      id="segments-line"
      fill="none"
      style="clip-path: url(#path-segments)"
    />
  </WithTween>

  <WithTween
    value={areaFunction(yAccessor)(filteredData)}
    tweenProps={{
      duration,
      interpolate: interpolatePath,
      easing: cubicOut,
      delay,
    }}
    let:output={at}
  >
    <path
      d={at}
      fill="hsla(217,100%, 50%, 0.1)"
      style="clip-path: url(#path-segments)"
    />
  </WithTween>

  <!-- 
  {#each segments as segment}
    <WithTween
      value={{
        x: $xScale(segment[0][xAccessor]),
        width:
          $xScale(segment.at(-1)[xAccessor]) - $xScale(segment[0][xAccessor]),
      }}
      tweenProps={{
        duration: 500,
        easing: cubicOut,
      }}
      let:output
    >
      <rect
        x={output.x}
        y={0}
        fill="hsla(1,100%, 50%, 0.1)"
        height={$yScale.range()[0]}
        width={output.width}
      />
    </WithTween>
  {/each} -->

  <!-- clip rects for segments -->
  <defs>
    <clipPath id="path-segments">
      {#each segments as segment (segment[0][xAccessor])}
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
            delay,
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
