<!-- @component
This is an older component from a different open-source code base.
It is probably not the most up to date code; but it works very well in practice.
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { contexts } from "../constants";

  import { WithTween } from "../functional-components";

  interface Point {
    x: number;
    y: number;
    label: string;
    valueColorClass?: string;
    valueStyleClass?: string;
    labelColorClass?: string;
    labelStyleClass?: string;
    pointColorClass?: string;
  }

  export let point = [];

  export let formatValue = (v) => v;
  export let xOffset = 0;
  export let fontSize = 11;
  export let xBuffer = 8;
  export let yBuffer = 3;
  export let alignLabels = true;
  export let showPoints = true;
  export let showLabels = true;

  export let keepPointsTrue = false;

  // plot the middle and push out from there

  const xScale = getContext(contexts.scale("x"));
  const yScale = getContext(contexts.scale("y"));
  const config = getContext(contexts.config);
  $: plotLeft = $config?.plotLeft;
  $: plotRight = $config?.plotRight;
  $: plotTop = $config?.plotTop;
  $: plotBottom = $config?.plotBottom;
  $: width = $config?.plotRight - $config?.plotLeft;

  export let direction = "right";
  export let flipAtEdge: "body" | "graphic" | false = "graphic"; // "body", "graphic", or undefined

  function toLocations(pt, xs, ys, left, right, top, bottom, elementHeight) {
    // this is where the boundary condition lives.
    let locations = [
      ...pt.map((p) => ({
        ...p,
        xRange: Math.max(left, xs(p.x)) || 0,
        yRange: ys(p.y),
      })),
    ];
    // sort order makes all the difference here
    locations.sort((a, b) => {
      if (a.y < b.y) return 1;
      if (a.y > b.y) return -1;
      return 0;
    });
    if (locations.length === 1) {
      locations[0].yRange = Math.min(
        bottom,
        Math.max(top, locations[0].yRange)
      );
      return locations;
    }
    if (!locations.length) return locations;

    const middle = ~~(locations.length / 2); // eslint-disable-line

    // STEP 1: inside up to top label.
    let i = middle;
    while (i >= 0) {
      if (i !== middle) {
        const diff = locations[i + 1].yRange - locations[i].yRange;
        if (diff <= elementHeight + yBuffer) {
          locations[i].yRange -= elementHeight + yBuffer - diff;
        }
      }
      i -= 1;
    }

    // STEP 2: top label shuffle down to reasonable place, shift to middle.
    if (locations[0].yRange < top + yBuffer) {
      locations[0].yRange = top + yBuffer;
      i = 0;
      while (i < middle) {
        const diff = locations[i + 1].yRange - locations[i].yRange;
        if (diff <= elementHeight + yBuffer) {
          locations[i + 1].yRange += elementHeight + yBuffer - diff;
        }
        i += 1;
      }
    }

    // STEP 3: inside down to bottom label;
    i = middle;
    while (i < locations.length) {
      if (i !== middle) {
        const diff = locations[i].yRange - locations[i - 1].yRange;
        if (diff < elementHeight + yBuffer) {
          locations[i].yRange += elementHeight + yBuffer - diff;
        }
      }
      i += 1;
    }
    if (locations[locations.length - 1].yRange > bottom - yBuffer) {
      locations[locations.length - 1].yRange = bottom - yBuffer;
      i = locations.length - 1;
      while (i > 0) {
        const diff = locations[i].yRange - locations[i - 1].yRange;
        if (diff <= fontSize + yBuffer) {
          locations[i - 1].yRange -= elementHeight + yBuffer - diff;
        }
        i -= 1;
      }
    }
    return locations;
  }

  let locations = toLocations(
    point,
    $xScale,
    $yScale,
    plotLeft,
    plotRight,
    plotTop,
    plotBottom,
    fontSize
  );
  let container;
  let containerWidths = [];
  let labelWidth = 0;

  // update locations.
  $: locations = toLocations(
    point,
    $xScale,
    $yScale,
    plotLeft,
    plotRight,
    plotTop,
    plotBottom,
    fontSize
  );
  // update containerWidths. We keep track of the last 6 points.
  $: if (container && locations) {
    containerWidths = [
      ...containerWidths.slice(-6),
      container.getBoundingClientRect().width,
    ];
  }

  // directions: 'left', 'left-plot', 'right-graphic', 'left-graphic'

  // If all the containerWidth histories + the x location are greatre than right plot, then flip.
  // this prevents jitter at the border region of the flip.

  let fcn = () => true;
  let internalDirection = direction;
  $: if (direction === "left") {
    let flip = !!flipAtEdge;
    fcn = (c) =>
      flip && locations[0].xRange - c <= (flipAtEdge === "body" ? plotLeft : 0);
  } else {
    let flip = !!flipAtEdge;
    fcn = (c) =>
      flip &&
      c + locations[0].xRange >= (flipAtEdge === "body" ? plotRight : width);
  }
  $: if (direction === "right" && containerWidths.every(fcn)) {
    internalDirection = "left";
  } else if (direction === "left" && containerWidths.every(fcn)) {
    internalDirection = "right";
  } else {
    internalDirection = direction;
  }

  $: if (container && locations && $xScale && $yScale) {
    labelWidth = Math.max(
      ...Array.from(container.querySelectorAll(".widths")).map(
        (q) => q.getBoundingClientRect().width
      )
    );
    if (!Number.isFinite(labelWidth)) {
      labelWidth = 0;
    }
  }
</script>

<filter id="outliner">
  <feMorphology
    operator="dilate"
    radius="2"
    in="SourceGraphic"
    result="THICKNESS"
  />
  <feComposite operator="out" in="THICKNESS" in2="SourceGraphic" />
</filter>

<g bind:this={container}>
  {#if showLabels}
    {#each locations as location, i (location.key || location.label)}
      <WithTween
        value={{
          y: location.yRange || 0,
          x:
            internalDirection === "right"
              ? location.xRange +
                (xBuffer + xOffset + (alignLabels ? labelWidth : 0))
              : location.xRange - xBuffer - xOffset,
        }}
        let:output={v}
        tweenProps={{ duration: 80 }}
      >
        <text
          filter="url(#outliner)"
          fill="white"
          data-location={location.yRange}
          font-size={fontSize}
        >
          {#if internalDirection === "right"}
            <tspan
              dy=".35em"
              style="font-weight: bold;"
              text-anchor="end"
              class="widths"
              y={v.y}
              x={v.x}
            >
              {formatValue(location.y)}
            </tspan>
            <tspan dy=".35em" y={v.y} x={v.x}>{location.label}</tspan>
          {:else}
            <tspan dy=".35em" y={v.y} x={v.x - labelWidth} text-anchor="end">
              {location.label}
            </tspan>
            <tspan
              dy=".35em"
              style="font-weight: bold;"
              class="widths"
              text-anchor="end"
              y={v.y}
              x={v.x}
            >
              {formatValue(location.y)}
            </tspan>
          {/if}
        </text>

        <text font-size={fontSize}>
          {#if internalDirection === "right"}
            <tspan
              dy=".35em"
              style="font-weight: bold;"
              class="widths  {location?.valueStyleClass ||
                ''} {location?.valueColorClass || ''}"
              y={v.y}
              text-anchor="end"
              x={v.x}
            >
              {formatValue(location.y)}
            </tspan>
            <tspan
              dy=".35em"
              y={v.y}
              x={v.x}
              class="mc-mouseover-label  {location?.labelStyleClass ||
                ''} {location?.labelColorClass || ''}"
            >
              {location.label}
            </tspan>
          {:else}
            <tspan
              dy=".35em"
              y={v.y}
              x={v.x - labelWidth}
              class="mc-mouseover-label  {location?.labelStyleClass ||
                ''} {location?.labelColorClass || ''}"
              text-anchor="end"
            >
              {location.label}
            </tspan>
            <tspan
              dy=".35em"
              style="font-weight: bold;"
              class="widths {location?.valueStyleClass ||
                ''} {location?.valueColorClass || ''}"
              text-anchor="end"
              y={v.y}
              x={v.x}
            >
              {formatValue(location.y)}
            </tspan>
          {/if}
        </text>
      </WithTween>
    {/each}
  {/if}
  {#if showPoints}
    {#each locations as { x, y, xRange, yRange, pointColorClass = '', label, key }, i (key || label)}
      <WithTween
        tweenProps={{ duration: 80 }}
        value={[
          keepPointsTrue ? $xScale(x) : xRange,
          keepPointsTrue ? $yScale(y) : yRange,
        ]}
        let:output
      >
        <circle cx={output[0]} cy={output[1]} r={5} fill="white" />
        <circle cx={output[0]} cy={output[1]} r={3} class={pointColorClass} />
      </WithTween>
      <!-- <Point scaling={false} {x} {y} {color} size={3} /> -->
    {/each}
  {/if}
</g>

<style>
  .mc-mouseover-label {
    cursor: pointer;
    transition: fill 200ms;
  }
</style>
