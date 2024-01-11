<!-- @component
This component will draw an axis on the specified side.
-->
<script lang="ts">
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { SingleDigitTimesPowerOfTenFormatter } from "@rilldata/web-common/lib/number-formatting/strategies/SingleDigitTimesPowerOfTen";
  import { formatMsInterval } from "@rilldata/web-common/lib/number-formatting/strategies/intervals";
  import { getContext } from "svelte";
  import { contexts } from "../constants";
  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";
  import { createTimeFormat, getTicks } from "../utils";
  import type { AxisSide } from "./types";

  export let side: AxisSide = "left";

  export let showTicks = false;
  export let tickLength = 4;
  export let tickBuffer = 4;

  export let fontSize: number | undefined = undefined;
  export let placement = "middle";

  export let labelColor = "fill-gray-600 dark:fill-gray-400";
  export let numberKind: NumberKind = NumberKind.ANY;

  // superlabel properties
  export let superlabel = false;
  let superlabelDate = "";
  const superlabelBuffer = side === "top" ? -12 : 12;
  let tickTextPosition;

  let xOrY;
  const isVertical = side === "left" || side === "right";
  if (isVertical) {
    // get Y scale
    xOrY = "y";
  } else {
    // get X Scale
    xOrY = "x";
  }

  const mainScale = getContext(contexts.scale(xOrY)) as ScaleStore;
  const plotConfig = getContext(contexts.config) as SimpleConfigurationStore;

  /** set a font size variable here */
  $: innerFontSize = $plotConfig.fontSize || fontSize || 11;

  /** make any adjustments to the scale to get what we need */
  $: scale = $mainScale;

  // text-anchor
  let textAnchor;
  $: if (side === "left") {
    textAnchor = "end";
  } else if (side === "right") {
    textAnchor = "start";
  } else {
    textAnchor = placement; // middle by default
  }

  function x(side: AxisSide, value) {
    if (side === "left") {
      return $plotConfig.left - tickLength - tickBuffer;
    } else if (side === "right") {
      return $plotConfig.width - $plotConfig.right + tickLength + tickBuffer;
    }
    return scale(value);
  }

  function y(side: AxisSide, value) {
    if (side === "top") {
      return $plotConfig.top - tickLength - tickBuffer;
    } else if (side === "bottom") {
      return (
        $plotConfig.height -
        $plotConfig.bottom +
        (innerFontSize || 0) +
        tickLength
      );
    }
    return scale(value);
  }

  function dy(side: AxisSide) {
    if (side === "top") {
      return 0;
    } else if (side === "bottom") {
      return 0;
    }
    // left and right
    return ".35em";
  }

  function dx(side: AxisSide) {
    if (side === "left" || side === "right") return 0;
    return `${$plotConfig?.bodyBuffer}px`;
  }

  function placeTick(side: AxisSide, value) {
    if (side === "top") {
      return {
        x1: scale(value),
        x2: scale(value),
        y1: $plotConfig.top,
        y2: $plotConfig.top - tickLength,
      };
    } else if (side === "bottom") {
      return {
        x1: scale(value),
        x2: scale(value),
        y1: $plotConfig.height - $plotConfig.bottom,
        y2: $plotConfig.height - $plotConfig.bottom + tickLength,
      };
    } else if (side === "left") {
      return {
        x1: $plotConfig.left,
        x2: $plotConfig.left - tickLength,
        y1: scale(value),
        y2: scale(value),
      };
    }
    // right
    return {
      x1: $plotConfig.width - $plotConfig.right,
      x2: $plotConfig.width - $plotConfig.right + tickLength,
      y1: scale(value),
      y2: scale(value),
    };
  }

  function shouldPlaceSuperLabel(currentDate, i) {
    if ((side === "top" || side === "bottom") && superlabel) {
      if (i === 0 || currentDate !== superlabelDate) {
        superlabelDate = currentDate;
        return true;
      } else return false;
    }
  }

  let axisLength;
  let ticks = [];
  $: if ($plotConfig) {
    if (xOrY === "x") axisLength = $plotConfig.graphicWidth;
    else axisLength = $plotConfig.graphicHeight;

    ticks = getTicks(
      xOrY,
      scale,
      axisLength,
      $plotConfig[`${xOrY}Type`] === "date",
    );
  }

  let formatterFunction;
  let superLabelFormatter;

  $: if ($plotConfig[`${xOrY}Type`] === "date") {
    [formatterFunction, superLabelFormatter] = createTimeFormat(
      $mainScale.domain() as [Date, Date],
      ticks?.length,
    );
  } else {
    superlabel = false;
    // Note that SingleDigitTimesPowerOfTenFormatter expects a single
    // digit time a power of ten. The d3 axis function used upstream
    // in this context _normally_ satisfies this requirement, but
    // there are edge cases where it does not. In these edge cases,
    // this formatter often does the right thing, but may not in some
    // circumstances. See https://github.com/rilldata/rill/issues/3631
    const formatter = new SingleDigitTimesPowerOfTenFormatter(ticks, {
      strategy: "singleDigitTimesPowerOfTen",
      numberKind,
      padWithInsignificantZeros: false,
    });
    formatterFunction = (x) =>
      numberKind === "INTERVAL"
        ? formatMsInterval(x)
        : formatter.stringFormat(x);
  }
</script>

<g>
  {#each ticks as tick, i}
    {@const tickPlacement = placeTick(side, tick)}
    <text
      bind:this={tickTextPosition}
      x={x(side, tick)}
      y={y(side, tick)}
      dx={dx(side)}
      dy={dy(side)}
      text-anchor={textAnchor}
      font-size={innerFontSize}
      class="{labelColor}  pointer-events-none"
    >
      {formatterFunction(tick)}
    </text>
    {#if showTicks}
      <!-- tick mark -->
      <line
        class="stroke-gray-400 dark:stroke-gray-600"
        x1={tickPlacement.x1}
        x2={tickPlacement.x2}
        y1={tickPlacement.y1}
        y2={tickPlacement.y2}
        font-size={innerFontSize}
      />
    {/if}
    {#if superLabelFormatter && shouldPlaceSuperLabel(superLabelFormatter(tick), i)}
      <!-- fix dx placement when tickTextPosition is null  -->
      <text
        font-weight="bold"
        x={x(side, tick)}
        y={y(side, tick) + superlabelBuffer}
        dx={dx(side)}
        text-anchor="start"
        font-size={innerFontSize}
        class="{labelColor} pointer-events-none"
      >
        {superLabelFormatter(tick)}
      </text>
    {/if}
  {/each}
</g>
