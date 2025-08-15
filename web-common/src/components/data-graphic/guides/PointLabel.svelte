<script lang="ts">
  import { outline } from "@rilldata/web-common/components/data-graphic/actions/outline";
  import {
    WithGraphicContexts,
    WithTween,
  } from "@rilldata/web-common/components/data-graphic/functional-components";
  import { justEnoughPrecision } from "@rilldata/web-common/lib/formatters";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { cubicOut } from "svelte/easing";
  import { fade } from "svelte/transition";
  export let point;
  export let xAccessor: string;
  export let yAccessor: string;
  export let location: "left" | "right" = "right";
  export let showText = true;
  export let showComparisonText = false;
  export let showPoint = true;
  export let showReferenceLine = true;
  export let showDistanceLine = true;
  export let yComparisonAccessor: string | undefined = undefined;
  export let format = justEnoughPrecision;

  let lastAvailablePoint;

  const COMPARISON_DIST = 6;
  /**
   * If the point is null, we want to use the last available point to
   * calculate the y position of the label. This is so that the label
   * doesn't jump around when the data is null.
   */
  $: if (point[yAccessor]) {
    lastAvailablePoint = { ...point };
  }

  function scaleFromOrigin(
    node,
    {
      delay = 0,
      duration = 1000,
      easing = cubicOut,
      start = 0,
      opacity = 0,
    } = {},
  ) {
    const style = getComputedStyle(node);
    const target_opacity = +style.opacity;
    const transform = style.transform === "none" ? "" : style.transform;

    const sd = 1 - start;
    const od = target_opacity * (1 - opacity);

    return {
      delay,
      duration,
      easing,
      css: (_t, u) => `
      transform-box: fill-box;
      transform-origin: 100% 100%;
			transform: ${transform} scale(${1 - sd * u});
			opacity: ${target_opacity - od * u}
		`,
    };
  }
</script>

<WithGraphicContexts let:xScale let:yScale let:config>
  {@const isNull = point[yAccessor] == null}
  {@const comparisonIsNull =
    yComparisonAccessor === undefined ||
    point[yComparisonAccessor] === null ||
    point[yComparisonAccessor] === undefined}
  {@const x = xScale(point[xAccessor])}
  {@const y = !isNull
    ? yScale(point[yAccessor])
    : lastAvailablePoint
      ? yScale(lastAvailablePoint[yAccessor])
      : (config.plotBottom - config.plotTop) / 2}
  <!-- these elements aren't used unless we are comparing-->
  {@const comparisonY = yScale(point?.[`comparison.${yAccessor}`] || 0)}
  <WithTween
    value={{
      x,
      y,
      dy: point?.[yAccessor] || 0,
      cdy: comparisonY,
    }}
    tweenProps={{ duration: 50 }}
    let:output
  >
    {@const text = isNull
      ? "no data"
      : format
        ? format(point[yAccessor])
        : point[yAccessor]}
    {@const comparisonText =
      isNull || yComparisonAccessor === undefined
        ? "no data"
        : format
          ? format(point[yAccessor] - point[yComparisonAccessor])
          : point[yAccessor] - point[yComparisonAccessor]}
    {@const percentageDifference =
      (isNull && comparisonIsNull) || yComparisonAccessor === undefined
        ? undefined
        : (point[yAccessor] - point[yComparisonAccessor]) /
          point[yComparisonAccessor]}
    {@const comparisonIsPositive = percentageDifference
      ? percentageDifference >= 0
      : undefined}
    {#if showReferenceLine}
      <line
        transition:fade={{ duration: 100 }}
        x1={output.x}
        x2={output.x}
        y1={config.plotTop}
        y2={config.plotBottom}
        stroke-width="1"
        class="stroke-gray-400"
        stroke-dasharray="2,1"
      />
    {/if}
    {#if showText}
      <text
        class:fill-gray-400={isNull}
        class:italic={isNull}
        use:outline
        x={output.x}
        y={output.y}
        text-anchor={location === "left" ? "end" : "start"}
        dx={8 * (location === "left" ? -1 : 1)}
        dy=".35em"
      >
        {text}
      </text>
    {/if}
    {#if !isNull && showDistanceLine}
      <line
        transition:fade={{ duration: 100 }}
        x1={output.x}
        x2={output.x}
        y1={showComparisonText ? output.cdy : yScale(0)}
        y2={output.y}
        stroke-width="4"
        class={showComparisonText && !comparisonIsPositive
          ? "stroke-red-300"
          : "stroke-primary-300"}
      />
      {#if showComparisonText}
        {@const signedDist = !comparisonIsPositive
          ? -1 * COMPARISON_DIST
          : 1 * COMPARISON_DIST}
        {@const yLoc = output.y + signedDist}
        {@const show = Math.abs(output.y - output.cdy) > 24}
        {#if show}
          <line
            x1={output.x}
            x2={output.x + COMPARISON_DIST}
            y1={yLoc}
            stroke-width="4"
            y2={yLoc + signedDist}
            class={showComparisonText && !comparisonIsPositive
              ? "stroke-red-300"
              : "stroke-primary-300"}
          />
          <line
            x1={output.x}
            x2={output.x - COMPARISON_DIST}
            y1={yLoc}
            stroke-width="4"
            y2={yLoc + signedDist}
            class={showComparisonText && !comparisonIsPositive
              ? "stroke-red-300"
              : "stroke-primary-300"}
          />
        {/if}
      {/if}
    {/if}
    {#if !isNull && showPoint}
      <circle
        transition:scaleFromOrigin
        cx={output.x}
        cy={output.y}
        r="3"
        class={showComparisonText && !comparisonIsPositive
          ? "fill-red-600"
          : "fill-primary-500"}
      />
    {/if}
    {#if !isNull && showPoint && showComparisonText}
      <circle
        transition:scaleFromOrigin
        cx={output.x}
        cy={output.cdy}
        r="3"
        class={showComparisonText && !comparisonIsPositive
          ? "fill-red-600"
          : "fill-primary-500"}
      />
    {/if}
    {#if showComparisonText && percentageDifference}
      {@const diffParts =
        formatMeasurePercentageDifference(percentageDifference)}
      <text
        class:fill-red-500={!comparisonIsPositive}
        class:italic={isNull}
        class="font-normal"
        use:outline
        x={output.x}
        y={output.y + 14}
        text-anchor={location === "left" ? "end" : "start"}
        dx={8 * (location === "left" ? -1 : 1)}
        dy=".35em"
      >
        {comparisonText}
        <tspan
          >{" "}
          ({diffParts?.neg || ""}{diffParts?.int || ""}<tspan class="opacity-50"
            >{diffParts?.percent || ""})</tspan
          >
        </tspan>
      </text>
    {/if}
  </WithTween>
</WithGraphicContexts>
