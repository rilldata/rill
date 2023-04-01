<!-- @component
This is an older component from a different open-source code base.
It is probably not the most up to date code; but it works very well in practice.
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { contexts } from "../constants";

  import { WithTween } from "../functional-components";
  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";
  import { preventVerticalOverlap } from "./prevent-vertical-overlap";

  interface Point {
    x: number;
    y: number;
    label: string;
    key: string;
    valueColorClass?: string;
    valueStyleClass?: string;
    labelColorClass?: string;
    labelStyleClass?: string;
    pointColorClass?: string;
    yOverride?: boolean;
    yOverrideLabel?: string;
    yOverrideStyleClass?: string;
  }

  export let point: Point[] = [];

  export let formatValue = (v) => v;
  export let xOffset = 0;
  export let fontSize = 11;
  export let xBuffer = 8;
  export let elementHeight = 12;
  export let yBuffer = 4;
  export let showLabels = true;

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

  let container;
  let containerWidths = [];
  // let labelWidth = 0;

  // update locations.
  $: nonOverlappingLocations = preventVerticalOverlap(
    point.map((p) => ({
      key: p.key,
      value: $yScale(p.y),
    })),
    plotTop,
    plotBottom,
    elementHeight,
    yBuffer
  );

  $: locations = point.map((p) => {
    return {
      ...p,
      xRange: $xScale(p.x),
      yRange: nonOverlappingLocations.find((l) => l.key === p.key).value,
    };
  });

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
        (q: SVGElement) => q.getBoundingClientRect().width
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
      {#if (location.y || location.yRange) && (location.x || location.xRange)}
        <WithTween
          value={location.xRange}
          let:output={x}
          tweenProps={{ duration: 25 }}
        >
          {@const xText =
            internalDirection === "right"
              ? location.xRange + (xBuffer + xOffset + labelWidth)
              : location.xRange - xBuffer - xOffset}
          <WithTween
            tweenProps={{ duration: 60 }}
            value={{
              label: location.yRange || 0,
              point: $yScale(location?.y) || 0,
            }}
            let:output={y}
          >
            <text
              font-size={fontSize}
              class="text-elements pointer-events-none"
            >
              {#if internalDirection === "right"}
                <tspan
                  dy=".35em"
                  class="widths {location?.valueStyleClass ||
                    'font-bold'} {location?.valueColorClass || ''}"
                  y={y.label}
                  text-anchor="end"
                  x={xText}
                >
                  {#if !location?.yOverride}
                    {formatValue(location.y)}
                  {/if}
                </tspan>
                <tspan
                  dy=".35em"
                  y={y.label}
                  x={xText - (location?.yOverride ? labelWidth : 0)}
                  class="mc-mouseover-label  {location?.yOverride
                    ? location?.yOverrideStyleClass
                    : location?.labelStyleClass || ''} {(!location?.yOverride &&
                    location?.labelColorClass) ||
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
                  y={y.label}
                  x={xText - labelWidth}
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
                  y={y.label}
                  x={xText}
                >
                  {#if !location?.yOverride}
                    {formatValue(location.y)}
                  {/if}
                </tspan>
              {/if}
            </text>
            {#if location.yRange}
              <circle
                cx={x}
                cy={y.point}
                r={3}
                paint-order="stroke"
                class={location.pointColorClass}
                stroke="white"
                stroke-width="3"
                opacity={location?.yOverride ? 0.7 : 1}
              />
            {/if}
          </WithTween>
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
