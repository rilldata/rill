<!-- @component
This component will draw an axis on the specified side.
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { timeFormat } from "d3-time-format";
  import { contexts } from "../constants";
  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";
  import type { AxisSide } from "./types.d";

  export let side: AxisSide = "left";
  export let formatter: (arg0: number | Date) => string = undefined;
  export let tickLength = 4;
  export let tickBuffer = 4;
  export let fontSize: number = undefined;
  export let placement = "middle";

  let container;
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
  $: innerFontSize = $plotConfig.fontSize || fontSize || 12;

  /** make any adjustments to the scale to get what we need */
  $: scale =
    $plotConfig[`${xOrY}Type`] === "date"
      ? $mainScale.copy().nice()
      : $mainScale.copy();

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

  function createTimeFormat(scaleDomain) {
    const diff = Math.abs(scaleDomain[1] - scaleDomain[0]) / 1000;

    const millisecondDiff = diff < 1;
    const secondDiff = diff < 60;
    const dayDiff = diff / (60 * 60) < 24;
    const fourDaysDiff = diff / (60 * 60) < 24 * 4;
    const manyDaysDiff = diff / (60 * 60 * 24) < 60;
    const manyMonthsDiff = diff / (60 * 60 * 24) < 365;

    return millisecondDiff
      ? timeFormat("%M:%S.%L")
      : secondDiff
      ? timeFormat("%M:%S")
      : dayDiff
      ? timeFormat("%H:%M")
      : fourDaysDiff || manyDaysDiff || manyMonthsDiff
      ? timeFormat("%b %d")
      : timeFormat("%Y");
  }

  let formatterFunction;

  $: if ($plotConfig[`${isVertical ? "y" : "x"}Type`] === "date") {
    formatterFunction = createTimeFormat($mainScale.domain());
  } else {
    formatterFunction = formatter || ((v) => v);
  }
  let axisLength;
  let tickCount = 0;
  // FIXME: we should be generalizing anything like this!
  // we also have a similar codeblock in Grid.svelte
  $: if ($plotConfig) {
    if (xOrY === "x") axisLength = $plotConfig.graphicWidth;
    else axisLength = $plotConfig.graphicHeight;
    // use graphicWidth or graphicHeight
    // do we ensure different spacing in one case vs. another?
    tickCount = ~~(axisLength / 20);
    tickCount = Math.max(2, ~~(axisLength / 100));
  }

  $: if (xOrY === "x") console.log(tickCount);
</script>

<g width={$plotConfig.graphicWidth} height={$plotConfig.graphicHeight}>
  {#each scale.ticks(tickCount) as tick}
    {@const tickPlacement = placeTick(side, tick)}
    <text
      x={x(side, tick)}
      y={y(side, tick)}
      dy={dy(side)}
      text-anchor={textAnchor}
      font-size={innerFontSize}
    >
      {formatterFunction(tick)}
    </text>
    <!-- tick mark -->
    <line
      class="stroke-gray-400"
      x1={tickPlacement.x1}
      x2={tickPlacement.x2}
      y1={tickPlacement.y1}
      y2={tickPlacement.y2}
      font-size={innerFontSize}
      stroke="black"
    />
  {/each}
</g>
