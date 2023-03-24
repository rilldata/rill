<script lang="ts">
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import MultiMetricMouseoverLabel from "@rilldata/web-common/components/data-graphic/marks/MultiMetricMouseoverLabel.svelte";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  export let point;
  export let xAccessor;
  export let yAccessor;
  export let showComparison = false;
  export let mouseoverFormat;
  $: comparisonYAccessor = `comparison.${yAccessor}`;

  $: x = point[xAccessor];
  $: y = point[yAccessor];
  $: comparisonY = point?.[comparisonYAccessor];

  $: hasValidComparisonPoint = showComparison && comparisonY !== undefined;

  $: diff = (y - comparisonY) / comparisonY;

  $: comparisonIsPositive = diff >= 0;

  $: diffLabel = formatMeasurePercentageDifference(
    (y - comparisonY) / comparisonY,
    "stringFormat"
  );

  let lastAvailableCurrentY;
  let lastAvailableComparisonY;
  $: if (y !== undefined && y !== null) {
    lastAvailableCurrentY = y;
  }
  $: if (
    point[comparisonYAccessor] !== undefined &&
    point[comparisonYAccessor] !== null
  ) {
    lastAvailableComparisonY = comparisonY;
  }

  $: currentPointIsNull = y === null;
  $: comparisonPointIsNull = comparisonY === null || comparisonY === undefined;

  $: mainPoint = {
    x,
    y: currentPointIsNull ? lastAvailableCurrentY : y,
    yOverride: currentPointIsNull,
    yOverrideLabel: "no data",
    key: "main",
    label:
      hasValidComparisonPoint && !currentPointIsNull && !comparisonPointIsNull
        ? `(${diffLabel})`
        : "",
    pointColorClass: "fill-blue-700",
    valueStyleClass: "font-semibold",
    valueColorClass: "fill-gray-600",
    labelColorClass:
      !comparisonIsPositive && showComparison
        ? "fill-red-500"
        : "fill-gray-600",
  };

  $: comparisonPoint = hasValidComparisonPoint
    ? {
        x,
        y: comparisonPointIsNull ? lastAvailableComparisonY : comparisonY,
        yOverride: comparisonPointIsNull,
        yOverrideLabel: "no comparison data",
        label: "prev.",
        key: "comparison",
        valueStyleClass: "font-normal",
        pointColorClass: "fill-gray-400",
        valueColorClass: "fill-gray-500",
        labelColorClass: "fill-gray-500",
      }
    : undefined;

  /** get the final point set*/
  $: pointSet = hasValidComparisonPoint
    ? [mainPoint, comparisonPoint]
    : [mainPoint];
</script>

<WithGraphicContexts let:xScale let:yScale>
  {@const strokeWidth = showComparison ? 2 : 4}
  {@const colorClass =
    hasValidComparisonPoint && !comparisonIsPositive
      ? "stroke-red-400"
      : "stroke-blue-400"}

  {#if !(currentPointIsNull || comparisonPointIsNull)}
    <WithTween
      tweenProps={{ duration: 80 }}
      value={{
        x: xScale(x),
        y: yScale(y),
        dy: yScale(comparisonY) || yScale(0),
      }}
      let:output
    >
      {#if !showComparison || Math.abs(output.y - output.dy) > 8}
        {@const bufferSize = Math.abs(output.y - output.dy) > 16 ? 8 : 4}
        {@const yBuffer = !hasValidComparisonPoint
          ? 0
          : !comparisonIsPositive
          ? -bufferSize
          : bufferSize}

        <line
          x1={output.x}
          x2={output.x}
          y1={output.y + yBuffer}
          y2={output.dy - yBuffer}
          class={colorClass}
          stroke-width={strokeWidth}
          stroke-linecap="round"
        />
        {@const sign = !comparisonIsPositive ? -1 : 1}
        {@const dist = 3}
        {@const signedDist = sign * dist}
        {@const yLoc = output.y + bufferSize * sign}
        {@const show =
          Math.abs(output.y - output.dy) > 16 && hasValidComparisonPoint}
        <!-- arrows -->
        <g class:opacity-0={!show} class="transition-opacity">
          <!-- {#if show} -->
          <g>
            <line
              x1={output.x}
              x2={output.x + dist}
              y1={yLoc}
              stroke-width={strokeWidth}
              y2={yLoc + signedDist}
              class={colorClass}
              stroke-linecap="round"
            />
            <line
              x1={output.x}
              x2={output.x - dist}
              y1={yLoc}
              stroke-width={strokeWidth}
              y2={yLoc + signedDist}
              class={colorClass}
              stroke-linecap="round"
            />
          </g>
          <!-- {/if} -->
        </g>
      {/if}
    </WithTween>
  {/if}
  <MultiMetricMouseoverLabel
    direction="right"
    flipAtEdge={"graphic"}
    keepPointsTrue
    formatValue={mouseoverFormat}
    point={pointSet || []}
  />

  <!-- {/if} -->
</WithGraphicContexts>
<!-- lines and such -->
