<script lang="ts">
  import { tweened } from "svelte/motion";
  import { cubicOut } from "svelte/easing";
  import { fly } from "svelte/transition";
  import type {
    TimeSeriesPoint,
    DimensionSeriesData,
    ChartScales,
    ChartConfig,
  } from "./types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";
  import { formatDateTimeByGrain } from "@rilldata/web-common/lib/time/ranges/formatter";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import MeasureChartPointIndicator from "./MeasureChartPointIndicator.svelte";

  export let scales: ChartScales;
  export let config: ChartConfig;
  export let hoveredIndex: number;
  export let hoveredPoint: TimeSeriesPoint | null = null;
  export let dimensionData: DimensionSeriesData[] = [];
  export let showComparison: boolean = false;
  export let isComparingDimension: boolean = false;
  export let timeGrain: V1TimeGrain | undefined;
  export let formatter: (value: number | null) => string;

  const arrowStrokeWidth = 2;
  const arrowColorClass = "stroke-gray-400";

  $: y = hoveredPoint?.value ?? null;
  $: comparisonY = hoveredPoint?.comparisonValue ?? null;

  $: hasValidComparisonPoint =
    comparisonY !== undefined && comparisonY !== null;
  $: diff =
    y !== null &&
    y !== undefined &&
    comparisonY !== null &&
    comparisonY !== undefined &&
    comparisonY !== 0
      ? (y - comparisonY) / comparisonY
      : NaN;
  $: comparisonIsPositive = diff >= 0;
  $: isDiffValid = !isNaN(diff);
  $: diffLabel =
    isDiffValid && numberPartsToString(formatMeasurePercentageDifference(diff));

  $: currentPointIsNull = y === null;
  $: comparisonPointIsNull = comparisonY === null || comparisonY === undefined;

  // Tweened pixel positions
  const tweenedX = tweened(0, { duration: 25, easing: cubicOut });
  const tweenedY = tweened(0, { duration: 60, easing: cubicOut });
  const tweenedComparisonY = tweened(0, { duration: 60, easing: cubicOut });

  $: tweenedX.set(scales.x(hoveredIndex));
  $: if (y !== null && y !== undefined) tweenedY.set(scales.y(y));
  $: if (comparisonY !== null && comparisonY !== undefined) {
    tweenedComparisonY.set(scales.y(comparisonY));
  }

  $: labelX = config.plotBounds.left + 6;
  $: labelY = config.plotBounds.top + 10;

  // Label spacing for comparison mode — push labels apart when they're too close
  $: minLabelGap = 12;
  $: rawPrimaryLabelY = $tweenedY + 4;
  $: rawCompLabelY = $tweenedComparisonY + 4;
  $: labelGap = Math.abs(rawPrimaryLabelY - rawCompLabelY);
  $: needsSpacing = labelGap < minLabelGap;
  $: midY = (rawPrimaryLabelY + rawCompLabelY) / 2;
  $: primaryLabelY = needsSpacing
    ? rawPrimaryLabelY <= rawCompLabelY
      ? midY - minLabelGap / 2
      : midY + minLabelGap / 2
    : rawPrimaryLabelY;
  $: compLabelY = needsSpacing
    ? rawCompLabelY <= rawPrimaryLabelY
      ? midY - minLabelGap / 2
      : midY + minLabelGap / 2
    : rawCompLabelY;
</script>

{#if hoveredPoint}
  <!-- Time label -->
  <g transition:fly={{ duration: 100, x: -4 }}>
    <text
      class="fill-fg-secondary stroke-surface-background text-xs"
      style:paint-order="stroke"
      stroke-width="3px"
      x={labelX}
      y={labelY}
    >
      {formatDateTimeByGrain(hoveredPoint.ts, timeGrain)}
    </text>
    {#if showComparison && hoveredPoint.comparisonTs}
      <text
        style:paint-order="stroke"
        stroke-width="3px"
        class="fill-fg-muted stroke-surface-background text-xs"
        x={labelX}
        y={labelY + 14}
      >
        {formatDateTimeByGrain(hoveredPoint.comparisonTs, timeGrain)} prev.
      </text>
    {/if}
  </g>

  <!-- Vertical line from zero to point (hidden in dimension comparison mode) -->
  {#if !currentPointIsNull && !isComparingDimension}
    <MeasureChartPointIndicator x={$tweenedX} y={$tweenedY} zeroY={scales.y(0)} value={formatter(y ?? null)} />
  {/if}

  <!-- Dimension comparison: guideline + per-series point circles -->
  {#if isComparingDimension}
    <line
      x1={$tweenedX}
      x2={$tweenedX}
      y1={config.plotBounds.top}
      y2={config.plotBounds.top + config.plotBounds.height}
      class="stroke-gray-300"
      stroke-width="1"
      stroke-dasharray="2,2"
    />
    {#each dimensionData as dim}
      {@const pt = dim.data[hoveredIndex]}
      {#if pt?.value !== null && pt?.value !== undefined}
        <circle
          cx={$tweenedX}
          cy={scales.y(pt.value)}
          r={4}
          fill={dim.color}
          class="stroke-surface-background stroke-[1.5px]"
        />
      {/if}
    {/each}
  {/if}

  {#if !isComparingDimension}
    <g>
      <!-- Comparison: arrow between points -->
      {#if showComparison && hasValidComparisonPoint && !comparisonPointIsNull && currentPointIsNull}
        <!-- Null current value with valid comparison -->
        <circle
          cx={$tweenedX}
          cy={$tweenedComparisonY}
          r={3}
          class="fill-theme-300"
        />
        <g class="text-xs">
          <text
            class="fill-fg-muted stroke-surface-background italic"
            style:paint-order="stroke"
            stroke-width="3px"
            x={$tweenedX + 8}
            y={$tweenedComparisonY - 8}
          >
            no current data
          </text>
          <text
            class="fill-fg-muted stroke-surface-background"
            style:paint-order="stroke"
            stroke-width="3px"
            x={$tweenedX + 8}
            y={$tweenedComparisonY + 4}
          >
            {formatter(comparisonY ?? null)} prev.
          </text>
        </g>
      {:else if showComparison && hasValidComparisonPoint && !currentPointIsNull && !comparisonPointIsNull}
        {@const yDiff = Math.abs($tweenedY - $tweenedComparisonY)}
        {#if yDiff > 8}
          {@const bufferSize = yDiff > 16 ? 8 : 4}
          {@const sign = comparisonIsPositive ? 1 : -1}
          {@const yBuffer = sign * bufferSize}

          <line
            x1={$tweenedX}
            x2={$tweenedX}
            y1={$tweenedY + yBuffer}
            y2={$tweenedComparisonY - yBuffer}
            class="stroke-surface-background"
            stroke-width={arrowStrokeWidth + 3}
            stroke-linecap="round"
          />
          <!-- Arrow line -->
          <line
            x1={$tweenedX}
            x2={$tweenedX}
            y1={$tweenedY + yBuffer}
            y2={$tweenedComparisonY - yBuffer}
            class={arrowColorClass}
            stroke-width={arrowStrokeWidth}
            stroke-linecap="round"
          />

          <!-- Arrow head -->
          {#if yDiff > 16}
            {@const yLoc = $tweenedY + bufferSize * sign}
            {@const dist = 3}
            {@const signedDist = sign * dist}
            <line
              x1={$tweenedX}
              x2={$tweenedX + dist}
              y1={yLoc}
              y2={yLoc + signedDist}
              class={arrowColorClass}
              stroke-width={arrowStrokeWidth}
              stroke-linecap="round"
            />
            <line
              x1={$tweenedX}
              x2={$tweenedX - dist}
              y1={yLoc}
              y2={yLoc + signedDist}
              class={arrowColorClass}
              stroke-width={arrowStrokeWidth}
              stroke-linecap="round"
            />
          {/if}
        {/if}

        <!-- Comparison point circle -->
        <circle
          cx={$tweenedX}
          cy={$tweenedComparisonY}
          r={3}
          class="fill-theme-300"
        />

        <!-- Primary value + diff -->
        <g class="text-xs">
          {#if !currentPointIsNull && isDiffValid}
            <text
              class="fill-fg-secondary stroke-surface-background font-semibold"
              style:paint-order="stroke"
              stroke-width="3px"
              x={$tweenedX + 8}
              y={primaryLabelY}
            >
              {formatter(y ?? null)}
              <tspan
                class={comparisonIsPositive ? "fill-gray-600" : "fill-red-500"}
              >
                ({diffLabel})
              </tspan>
            </text>
          {/if}

          <!-- Comparison value -->
          <text
            class="fill-fg-muted stroke-surface-background"
            style:paint-order="stroke"
            stroke-width="3px"
            x={$tweenedX + 8}
            y={compLabelY}
          >
            {#if comparisonPointIsNull}
              <tspan class="italic">no comparison data</tspan>
            {:else}
              {formatter(comparisonY ?? null)} prev.
            {/if}
          </text>
        </g>
      {:else if !currentPointIsNull}
        <!-- No comparison: just show the primary value -->
        <text
          class="fill-fg-secondary stroke-surface-background text-xs font-semibold"
          style:paint-order="stroke"
          stroke-width="3px"
          x={$tweenedX + 8}
          y={$tweenedY + 4}
        >
          {formatter(y ?? null)}
        </text>
      {:else}
        <!-- Null current value -->
        <text
          class="fill-fg-muted stroke-surface-background text-xs italic"
          style:paint-order="stroke"
          stroke-width="3px"
          x={$tweenedX + 8}
          y={scales.y(0) + 4}
        >
          no current data
        </text>
      {/if}
    </g>
  {/if}
{/if}
