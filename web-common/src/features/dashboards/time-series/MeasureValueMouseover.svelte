<script lang="ts">
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import MultiMetricMouseoverLabel from "@rilldata/web-common/components/data-graphic/marks/MultiMetricMouseoverLabel.svelte";
  import type { Point } from "@rilldata/web-common/components/data-graphic/marks/types";

  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";
  import { fade } from "svelte/transition";
  export let point;
  export let xAccessor;
  export let yAccessor;
  export let showComparison = false;
  export let mouseoverFormat;
  export let numberKind: NumberKind;
  $: comparisonYAccessor = `comparison.${yAccessor}`;

  $: x = point?.[xAccessor];
  $: y = point?.[yAccessor];
  $: comparisonY = point?.[comparisonYAccessor];

  $: hasValidComparisonPoint = comparisonY !== undefined;

  $: diff = (y - comparisonY) / comparisonY;

  $: comparisonIsPositive = diff >= 0;

  $: isDiffValid = !isNaN(diff);

  $: diffLabel =
    isDiffValid && numberPartsToString(formatMeasurePercentageDifference(diff));

  let lastAvailableCurrentY = 0;
  let lastAvailableComparisonY;
  $: if (y !== undefined && y !== null) {
    lastAvailableCurrentY = y;
  }
  $: if (
    point?.[comparisonYAccessor] !== undefined &&
    point?.[comparisonYAccessor] !== null
  ) {
    lastAvailableComparisonY = comparisonY;
  }

  $: currentPointIsNull = y === null;
  $: comparisonPointIsNull = comparisonY === null || comparisonY === undefined;

  $: mainPoint = {
    x,
    y: currentPointIsNull ? lastAvailableCurrentY : y,
    yOverride: currentPointIsNull,
    yOverrideLabel: "no current data",
    yOverrideStyleClass: "fill-gray-500 italic",
    key: "main",
    label:
      showComparison &&
      hasValidComparisonPoint &&
      !currentPointIsNull &&
      !comparisonPointIsNull &&
      numberKind !== NumberKind.PERCENT &&
      isDiffValid
        ? `(${diffLabel})`
        : "",
    pointColor: "var(--color-theme-700)",
    valueStyleClass: "font-semibold",
    valueColorClass: "fill-gray-600",
    labelColorClass:
      !comparisonIsPositive && showComparison
        ? "fill-red-500"
        : "fill-gray-600",
  };

  $: comparisonPoint =
    showComparison && hasValidComparisonPoint
      ? {
          x,
          y: comparisonPointIsNull ? lastAvailableComparisonY : comparisonY,
          yOverride: comparisonPointIsNull,
          yOverrideLabel: "no comparison data",
          yOverrideStyleClass: "fill-gray-500 italic",
          label: "prev.",
          key: "comparison",
          valueStyleClass: "font-normal",
          pointColor: "var(--color-theme-300)",
          valueColorClass: "fill-gray-500",
          labelColorClass: "fill-gray-500",
        }
      : undefined;

  /** get the final point set*/
  let pointSet: Point[] = [];
  $: pointSet =
    showComparison && comparisonPoint
      ? [comparisonPoint, mainPoint]
      : [mainPoint];

  /** modes
   * 1. comparison not activated b/c not valid for time range
   * 2. no comparison point available even if comparison is activated
   * 3. no comparison point available, but current is null
   * 4. comparison point available, but current point is null
   * 5. comparison point available, but current point is not null
   * 6. comparisoin point available, neither points are null.
   */
</script>

<WithGraphicContexts let:xScale let:yScale>
  {@const strokeWidth = showComparison ? 2 : 4}
  {@const colorClass = "stroke-gray-400"}
  <WithTween
    tweenProps={{ duration: 25 }}
    value={xScale(x)}
    let:output={xArrow}
  >
    <WithTween
      tweenProps={{ duration: 60 }}
      value={{
        y: yScale(y) || yScale(0),
        dy: yScale(comparisonY) || yScale(0),
      }}
      let:output
    >
      {#if !(currentPointIsNull || comparisonPointIsNull) && x !== undefined && y !== undefined}
        {#if showComparison && Math.abs(output.y - output.dy) > 8}
          {@const bufferSize = Math.abs(output.y - output.dy) > 16 ? 8 : 4}
          {@const yBuffer = !hasValidComparisonPoint
            ? 0
            : !comparisonIsPositive
              ? -bufferSize
              : bufferSize}

          {@const sign = !comparisonIsPositive ? -1 : 1}
          {@const dist = 3}
          {@const signedDist = sign * dist}
          {@const yLoc = output.y + bufferSize * sign}
          {@const show =
            Math.abs(output.y - output.dy) > 16 && hasValidComparisonPoint}
          arrows
          <g>
            {#if show}
              <line
                x1={xArrow}
                x2={xArrow + dist}
                y1={yLoc}
                y2={yLoc + signedDist}
                stroke="var(--surface)"
                stroke-width={strokeWidth + 3}
                stroke-linecap="round"
              />
              <line
                x1={xArrow}
                x2={xArrow - dist}
                y1={yLoc}
                y2={yLoc + signedDist}
                stroke="var(--surface)"
                stroke-width={strokeWidth + 3}
                stroke-linecap="round"
              />
            {/if}

            <line
              x1={xArrow}
              x2={xArrow}
              y1={output.y + yBuffer}
              y2={output.dy - yBuffer}
              stroke="var(--surface)"
              stroke-width={strokeWidth + 3}
              stroke-linecap="round"
            />

            <line
              x1={xArrow}
              x2={xArrow}
              y1={output.y + yBuffer}
              y2={output.dy - yBuffer}
              class={colorClass}
              stroke-width={strokeWidth}
              stroke-linecap="round"
            />

            <g class:opacity-0={!show} class="transition-opacity">
              <g>
                <line
                  x1={xArrow}
                  x2={xArrow + dist}
                  y1={yLoc}
                  stroke-width={strokeWidth}
                  y2={yLoc + signedDist}
                  class={colorClass}
                  stroke-linecap="round"
                />
                <line
                  x1={xArrow}
                  x2={xArrow - dist}
                  y1={yLoc}
                  stroke-width={strokeWidth}
                  y2={yLoc + signedDist}
                  class={colorClass}
                  stroke-linecap="round"
                />
              </g>
            </g>
          </g>
        {/if}
      {/if}
      {#if !showComparison && x !== undefined && y !== null && y !== undefined && !currentPointIsNull}
        <line
          transition:fade={{ duration: 100 }}
          x1={xArrow}
          x2={xArrow}
          y1={yScale(0)}
          y2={output.y}
          stroke-width="4"
          class="stroke-theme-300"
        />
      {/if}
    </WithTween>
  </WithTween>

  <MultiMetricMouseoverLabel
    direction="right"
    flipAtEdge="body"
    formatValue={mouseoverFormat}
    point={pointSet || []}
  />
</WithGraphicContexts>
