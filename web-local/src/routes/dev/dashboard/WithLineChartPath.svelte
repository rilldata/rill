<script lang="ts">
  import { contexts } from "@rilldata/web-local/lib/components/data-graphic/constants";
  import { WithTween } from "@rilldata/web-local/lib/components/data-graphic/functional-components";
  import {
    lineFactory,
    pathIsDefined,
  } from "@rilldata/web-local/lib/components/data-graphic/utils";
  import { interpolatePath } from "d3-interpolate-path";
  import { scaleLinear } from "d3-scale";
  import { getContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  export let data;
  export let xAccessor: string;
  export let yAccessor: string;
  // get the scale functions from the data graphic
  const xScale = getContext(contexts.scale("x"));
  const yScale = getContext(contexts.scale("y"));
  let lineFunction;
  $: if ($xScale && $yScale)
    lineFunction = lineFactory({
      xScale: $xScale,
      yScale: $yScale,
      xAccessor,
      pathDefined: pathIsDefined(yAccessor),
    });

  // get segments
  /**
   * Helper function to compute the contiguous segments of the data
   *
   * Derived from https://github.com/pbeshai/d3-line-chunked/blob/master/src/lineChunked.js
   *
   * @param {Array} lineData the line data
   * @param {Function} defined function that takes a data point and returns true if
   *    it is defined, false otherwise
   * @param {Function} isNext function that takes the previous data point and the
   *    current one and returns true if the current point is the expected one to
   *    follow the previous, false otherwise.
   * @return {Array} An array of segments (subarrays) of the line data
   */
  function computeSegments(lineData, defined, isNext = (prev, curr) => true) {
    var startNewSegment = true;

    // split into segments of continuous data
    var segments = lineData.reduce(function (segments, d) {
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

  function stopsFromSegments(segments, xDomain) {
    var gradientScale = scaleLinear()
      .domain(xDomain)
      .range([0, 100])
      .clamp(true);

    var stops = segments.reduce(function (stops, segment) {
      // get first and last points from the segments
      var first = segment[0];
      var last = segment[segment.length - 1];
      // add gap-segment segment-gap stops (4)
      stops.push({
        type: "gap",
        offset: gradientScale(first[xAccessor]) + "%",
      });
      stops.push({
        type: "segment",
        offset: gradientScale(first[xAccessor]) + "%",
      });
      stops.push({
        type: "segment",
        offset: gradientScale(last[xAccessor]) + "%",
      });
      stops.push({
        type: "gap",
        offset: gradientScale(last[xAccessor]) + "%",
      });

      return stops;
    }, []);

    return stops;
  }
  let stops;
  $: segments = computeSegments(data, pathIsDefined(yAccessor));

  $: if ($xScale) stops = stopsFromSegments(segments, $xScale?.domain());
  $: filteredData = data.filter(pathIsDefined(yAccessor));
</script>

<g>
  <WithTween
    value={lineFunction(yAccessor)(filteredData)}
    tweenProps={{
      duration: 1000,
      interpolate: interpolatePath,
      easing: cubicOut,
    }}
    let:output={dt}
  >
    <path stroke="pink" fill="none" stroke-width="1px" d={dt} id="gap-line" />
  </WithTween>
  <WithTween
    value={lineFunction(yAccessor)(filteredData)}
    tweenProps={{
      duration: 1000,
      interpolate: interpolatePath,
      easing: cubicOut,
    }}
    let:output={dt}
  >
    <path
      stroke-width="4px"
      stroke="blue"
      d={dt}
      id="segments-line"
      fill="none"
      style="clip-path: url(#path-segments)"
    />
  </WithTween>
  {#each segments as segment}
    <WithTween
      value={{
        x: $xScale(segment[0][xAccessor]),
        width:
          $xScale(segment.at(-1)[xAccessor]) - $xScale(segment[0][xAccessor]),
      }}
      tweenProps={{
        duration: 1000,
        easing: cubicOut,
      }}
      let:output
    >
      <rect
        fill="hsla(1, 50%, 50%, .1)"
        x={output.x}
        y={0}
        height={$yScale.range()[0]}
        width={output.width}
      />
    </WithTween>
  {/each}
  <defs>
    <clipPath id="path-segments">
      {#each segments as segment}
        <WithTween
          value={{
            x: $xScale(segment[0][xAccessor]),
            width:
              $xScale(segment.at(-1)[xAccessor]) -
              $xScale(segment[0][xAccessor]),
          }}
          tweenProps={{
            duration: 1000,
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
    <!-- <linearGradient id="path-segments">
      {#each stops as stop}
        <stop
          offset={stop.offset}
          stop-color={stop.type === "gap" ? gapColor : segmentColor}
        />
      {/each}
    </linearGradient> -->
  </defs>
</g>

{#if false && lineFunction}
  <slot d={lineFunction(yAccessor)(data)} />
{/if}
