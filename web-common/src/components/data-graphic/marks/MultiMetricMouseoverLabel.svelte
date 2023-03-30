<!-- @component
This is an older component from a different open-source code base.
It is probably not the most up to date code; but it works very well in practice.
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { contexts } from "../constants";

  import { WithTween } from "../functional-components";
  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";

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

  export let point: Point[] = [];

  export let formatValue = (v) => v;
  export let xOffset = 0;
  export let fontSize = 11;
  export let xBuffer = 8;
  export let yBuffer = 3;
  export let showPoints = true;
  export let showLabels = true;

  export let keepPointsTrue = false;

  // plot the middle and push out from there

  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;
  const config = getContext(contexts.config) as SimpleConfigurationStore;
  $: plotLeft = $config?.plotLeft;
  $: plotRight = $config?.plotRight;
  $: plotTop = $config?.plotTop;
  $: plotBottom = $config?.plotBottom;
  $: width = $config?.width;

  export let direction = "right";
  export let flipAtEdge: "body" | "graphic" | false = "graphic"; // "body", "graphic", or undefined

  // FIXME – we should replace this with preventVerticalOverlap!
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
  // let labelWidth = 0;

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

  let fcn = (c: any) => true;

  let internalDirection = direction;

  $: if (direction === "left") {
    let flip = !!flipAtEdge;
    fcn = (c) =>
      flip &&
      locations[0].xRange - c <=
        (flipAtEdge === "body"
          ? plotLeft
          : flipAtEdge === "graphic"
          ? 0
          : false);
  } else {
    let flip = !!flipAtEdge;
    fcn = (c) =>
      flip &&
      c + locations[0].xRange >=
        (flipAtEdge === "body"
          ? plotRight
          : flipAtEdge === "graphic"
          ? width
          : false);
  }
  $: if (
    direction === "right" &&
    containerWidths.every(fcn) &&
    flipAtEdge !== false
  ) {
    internalDirection = "left";
  } else if (
    direction === "left" &&
    containerWidths.every(fcn) &&
    flipAtEdge !== false
  ) {
    internalDirection = "right";
  } else {
    internalDirection = direction;
  }

  let labelWidth = 0;
  /** the full text width */
  let textWidths = [];
  let transitionalTimeoutForCalculatingLabelWidth;
  $: if (container && locations && $xScale && $yScale) {
    clearTimeout(transitionalTimeoutForCalculatingLabelWidth);
    transitionalTimeoutForCalculatingLabelWidth = setTimeout(() => {
      labelWidth = Math.max(
        ...Array.from(container.querySelectorAll(".widths")).map(
          (q: SVGElement) => q.getBoundingClientRect().width
        )
      );

      textWidths = Array.from(container.querySelectorAll(".text-elements")).map(
        (q) => q.getBoundingClientRect().width
      );
      if (!Number.isFinite(labelWidth)) {
        labelWidth = 0;
      }
    }, 0);
  }
</script>

<g bind:this={container}>
  {#if showLabels}
    {#each locations as location, i (location.key || location.label)}
      {#if (location.y || location.rangeY) && (location.x || location.rangeX)}
        <WithTween
          value={{
            y: location.yRange || 0,
            x:
              internalDirection === "right"
                ? location.xRange + (xBuffer + xOffset + labelWidth)
                : location.xRange - xBuffer - xOffset,
          }}
          let:output={v}
          tweenProps={{ duration: 50 }}
        >
          <text font-size={fontSize} class="text-elements pointer-events-none">
            {#if internalDirection === "right"}
              <tspan
                dy=".35em"
                class="widths {location?.valueStyleClass ||
                  'font-bold'} {location?.valueColorClass || ''}"
                y={v.y}
                text-anchor="end"
                x={v.x}
              >
                {#if !location?.yOverride}
                  {formatValue(location.y)}
                {/if}
              </tspan>
              <tspan
                dy=".35em"
                y={v.y}
                x={v.x}
                class="mc-mouseover-label  {location?.labelStyleClass ||
                  ''} {(!location?.yOverride && location?.labelColorClass) ||
                  ''}"
              >
                {#if location?.yOverride}
                  {location.yOverrideLabel}
                {:else}
                  {location.label}
                {/if}
              </tspan>
            {:else}
              <tspan
                dy=".35em"
                y={v.y}
                x={v.x - labelWidth}
                class="mc-mouseover-label  {location?.labelStyleClass ||
                  ''} {(!location?.yOverride && location?.labelColorClass) ||
                  ''}"
                text-anchor="end"
              >
                {#if location?.yOverride}
                  {location.yOverrideLabel}
                {:else}
                  {location.label}
                {/if}
              </tspan>
              <tspan
                dy=".35em"
                class="widths {location?.valueStyleClass ||
                  'font-bold'} {location?.valueColorClass || ''}"
                text-anchor="end"
                y={v.y}
                x={v.x}
              >
                {#if !location?.yOverride}
                  {formatValue(location.y)}
                {/if}
              </tspan>
            {/if}
          </text>
        </WithTween>
      {/if}
    {/each}
  {/if}
  {#if showPoints}
    {#each locations as location, i (location.key || location.label)}
      {#if (keepPointsTrue && location.x !== undefined && location.y !== undefined) || (location.xRange !== undefined && location.yRange !== undefined)}
        <WithTween
          tweenProps={{ duration: 50 }}
          value={[
            keepPointsTrue ? $xScale(location.x) : location.xRange,
            keepPointsTrue ? $yScale(location.y) : location.yRange,
          ]}
          let:output
        >
          <circle cx={output[0]} cy={output[1]} r={5} fill="white" />
          <circle
            cx={output[0]}
            cy={output[1]}
            r={3}
            class={location.pointColorClass}
          />
        </WithTween>
      {/if}
    {/each}
  {/if}
</g>

<style>
  .mc-mouseover-label {
    cursor: pointer;
    transition: fill 200ms;
  }

  text {
    paint-order: stroke;
    stroke: white;
    stroke-width: 3px;
  }
</style>
