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

  $: diff =
    (point[yAccessor] - point[`comparison.${yAccessor}`]) /
    point[`comparison.${yAccessor}`];

  $: comparisonIsPositive = diff >= 0;

  $: diffLabel = formatMeasurePercentageDifference(
    (point[yAccessor] - point[`comparison.${yAccessor}`]) /
      point[`comparison.${yAccessor}`],
    "stringFormat"
  );

  $: mainPoint = {
    x: point[xAccessor],
    y: point[yAccessor],
    label: showComparison ? `(${diffLabel})` : "",
    pointColorClass: "fill-blue-700",
    valueStyleClass: "",
    valueColorClass: "fill-gray-600",
    labelColorClass:
      !comparisonIsPositive && showComparison
        ? "fill-red-500"
        : "fill-gray-600",
  };

  $: comparisonPoint = showComparison
    ? {
        x: point[xAccessor],
        y: point[`comparison.${yAccessor}`],
        label: "prev",
        pointColorClass: "fill-gray-400",
        valueColorClass: "fill-gray-500",
        labelColorClass: "fill-gray-500",
      }
    : undefined;

  /** get the final point set*/
  $: pointSet = showComparison ? [mainPoint, comparisonPoint] : [mainPoint];
</script>

<WithGraphicContexts let:xScale let:yScale let:config>
  {@const strokeWidth = showComparison ? 2 : 4}
  {@const colorClass =
    showComparison && !comparisonIsPositive
      ? "stroke-red-400"
      : "stroke-blue-400"}
  <WithTween
    tweenProps={{ duration: 50 }}
    value={{
      x: xScale(point[xAccessor]),
      y: yScale(point[yAccessor]),
      dy: yScale(point[`comparison.${yAccessor}`]) || yScale(0),
    }}
    let:output
  >
    {#if !showComparison || Math.abs(output.y - output.dy) > 8}
      {@const bufferSize = Math.abs(output.y - output.dy) > 16 ? 6 : 3}
      {@const yBuffer = !showComparison
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
      />
      {@const sign = !comparisonIsPositive ? -1 : 1}
      {@const dist = 4}
      {@const signedDist = sign * 6}
      {@const yLoc = output.y + signedDist}
      {@const show = Math.abs(output.y - output.dy) > 24 && showComparison}
      {#if show}
        <line
          x1={output.x}
          x2={output.x + dist}
          y1={yLoc}
          stroke-width={strokeWidth}
          y2={yLoc + signedDist}
          class={colorClass}
        />
        <line
          x1={output.x}
          x2={output.x - dist}
          y1={yLoc}
          stroke-width={strokeWidth}
          y2={yLoc + signedDist}
          class={colorClass}
        />
      {/if}
    {/if}
  </WithTween>

  <MultiMetricMouseoverLabel
    direction="right"
    flipAtEdge={"graphic"}
    keepPointsTrue
    formatValue={mouseoverFormat}
    point={pointSet}
  />

  <!-- {/if} -->
</WithGraphicContexts>
<!-- lines and such -->
