<!-- @component
This is an older component from a different open-source code base.
It is probably not the most up to date code; but it works very well in practice.
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { contexts } from "../constants";

  import DelayedLabel from "@rilldata/web-common/components/data-graphic/marks/DelayedLabel.svelte";
  import { WithTween } from "../functional-components";
  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";
  import { preventVerticalOverlap } from "./prevent-vertical-overlap";
  import type { Point } from "./types";

  const DIMENSION_HOVER_DURATION = 350;

  export let point: Point[] = [];

  export let formatValue = (v) => v;
  export let xOffset = 0;
  export let fontSize = 11;
  export let xBuffer = 8;
  export let elementHeight = 12;
  export let yBuffer = 4;
  export let showLabels = true;
  export let isDimension = false;

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
  export let attachPointToLabel = false;

  let container: SVGGElement;
  let containerWidths: number[] = [];
  // let labelWidth = 0;

  let fanOutLabels = true;

  // update locations.
  $: nonOverlappingLocations = preventVerticalOverlap(
    point.map((p) => ({
      key: p.key,
      value: $yScale(p.y),
    })),
    plotTop,
    plotBottom,
    elementHeight,
    yBuffer,
  );

  $: locations = point.map((p) => {
    const locationValue = nonOverlappingLocations.find(
      (l) => l.key === p.key,
    )?.value;

    return {
      ...p,
      xRange: $xScale(p.x),
      yRange:
        fanOutLabels && locationValue !== undefined
          ? locationValue
          : $yScale(p.y),
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

  let fcn: (c: number) => boolean = () => true;

  let internalDirection = direction;

  $: if (direction === "left") {
    let flip = !!flipAtEdge;
    fcn = (c) => {
      const rhs_comparator =
        flipAtEdge === "body" ? plotLeft : flipAtEdge === "graphic" ? 0 : false;
      return (
        flip &&
        typeof rhs_comparator === "number" &&
        locations[0].xRange - c <= rhs_comparator
      );
    };
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
  let transitionalTimeoutForCalculatingLabelWidth;

  $: if (container && locations && $xScale && $yScale) {
    clearTimeout(transitionalTimeoutForCalculatingLabelWidth);
    transitionalTimeoutForCalculatingLabelWidth = setTimeout(() => {
      if (container) {
        labelWidth = Math.max(
          ...Array.from(container.querySelectorAll(".widths")).map(
            (q: SVGElement) => q.getBoundingClientRect().width,
          ),
        );

        if (!Number.isFinite(labelWidth)) {
          labelWidth = 0;
        }
      }
    }, 0);
  }

  let transitionalTimeoutForFanningOutLabels;
  $: if (isDimension && container && point?.[0]?.x) {
    fanOutLabels = false;
    clearTimeout(transitionalTimeoutForFanningOutLabels);
    transitionalTimeoutForFanningOutLabels = setTimeout(() => {
      fanOutLabels = true;
    }, DIMENSION_HOVER_DURATION);
  }
</script>

<g bind:this={container}>
  {#if showLabels}
    {#each locations as location (location.key || location.label)}
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
            <DelayedLabel
              value={location.x}
              {isDimension}
              duration={DIMENSION_HOVER_DURATION}
              let:visibility
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
                    {visibility}
                  >
                    {#if !location?.yOverride}
                      {location.value
                        ? location.value
                        : formatValue(location.y)}
                    {/if}
                  </tspan>

                  <tspan
                    dy=".35em"
                    dx="0.4em"
                    y={y.label}
                    x={xText - (location?.yOverride ? labelWidth : 0)}
                    {visibility}
                    class="mc-mouseover-label {location?.yOverride
                      ? location?.yOverrideStyleClass
                      : location?.labelStyleClass ||
                        ''} {(!location?.yOverride &&
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
                    dx="-0.4em"
                    y={y.label}
                    x={xText - (location?.yOverride ? 0 : labelWidth)}
                    {visibility}
                    class="mc-mouseover-label {location?.yOverride
                      ? location?.yOverrideStyleClass
                      : location?.labelStyleClass ||
                        ''} {(!location?.yOverride &&
                      location?.labelColorClass) ||
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
                    {visibility}
                  >
                    {#if !location?.yOverride}
                      {location.value
                        ? location.value
                        : formatValue(location.y)}
                    {/if}
                  </tspan>
                {/if}
              </text>
            </DelayedLabel>
            {#if location.yRange}
              <circle
                cx={x}
                cy={attachPointToLabel ? y.label : y.point}
                r={3}
                paint-order="stroke"
                fill={location.pointColor}
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

    /* Make all characters and numbers of equal width for easy scanibility */
    font-feature-settings:
      "case" 0,
      "cpsp" 0,
      "dlig" 0,
      "frac" 0,
      "dnom" 0,
      "numr" 0,
      "salt" 0,
      "subs" 0,
      "sups" 0,
      "tnum",
      "zero" 0,
      "ss01",
      "ss02" 0,
      "ss03" 0,
      "ss04" 0,
      "cv01" 0,
      "cv02" 0,
      "cv03" 0,
      "cv04" 0,
      "cv05" 0,
      "cv06" 0,
      "cv07" 0,
      "cv08" 0,
      "cv09" 0,
      "cv10" 0,
      "cv11" 0,
      "calt",
      "ccmp",
      "kern";
  }
</style>
