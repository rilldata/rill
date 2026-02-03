<script lang="ts">
  import { tweened } from "svelte/motion";
  import { cubicOut } from "svelte/easing";
  import type {
    TimeSeriesPoint,
    DimensionSeriesData,
    ChartScales,
    ChartConfig,
  } from "./types";
  import MeasureChartPointIndicator from "./MeasureChartPointIndicator.svelte";

  export let scales: ChartScales;
  export let config: ChartConfig;
  export let hoveredIndex: number;
  export let hoveredPoint: TimeSeriesPoint | null = null;
  export let dimensionData: DimensionSeriesData[] = [];
  export let showComparison: boolean = false;
  export let isComparingDimension: boolean = false;
  export let isBarMode: boolean = false;
  export let visibleStart: number = 0;
  export let visibleEnd: number = 0;
  export let formatter: (value: number | null) => string;

  $: visibleCount = Math.max(1, visibleEnd - visibleStart + 1);
  $: slotWidth = config.plotBounds.width / visibleCount;
  $: gap = slotWidth * 0.2;
  $: bandWidth = slotWidth - gap;

  // Bar count: dimension comparison uses dimensionData.length, time comparison uses 2
  $: barCount = isComparingDimension
    ? dimensionData.length
    : (showComparison ? 2 : 1);
  $: barGap = barCount > 1 ? 2 : 0;
  $: totalGaps = barGap * (barCount - 1);
  $: singleBarWidth = (bandWidth - totalGaps) / barCount;

  // Slot center for bar positioning
  $: slot = hoveredIndex - visibleStart;
  $: slotCenterX = config.plotBounds.left + (slot + 0.5) * slotWidth;

  $: y = hoveredPoint?.value ?? null;
  $: comparisonY = hoveredPoint?.comparisonValue ?? null;
  $: currentPointIsNull = y === null;
  $: hasValidComparisonPoint =
    comparisonY !== undefined && comparisonY !== null;

  // Tweened pixel positions
  const tweenedX = tweened(0, { duration: 25, easing: cubicOut });
  const tweenedY = tweened(0, { duration: 60, easing: cubicOut });
  const tweenedComparisonY = tweened(0, { duration: 60, easing: cubicOut });

  $: tweenedX.set(scales.x(hoveredIndex));
  $: if (y !== null && y !== undefined) tweenedY.set(scales.y(y));
  $: if (comparisonY !== null && comparisonY !== undefined) {
    tweenedComparisonY.set(scales.y(comparisonY));
  }
</script>

{#if hoveredPoint}
  <!-- Primary point indicator (hidden in comparison modes) -->
  {#if !currentPointIsNull && !isComparingDimension && !showComparison}
    <MeasureChartPointIndicator
      x={$tweenedX}
      y={$tweenedY}
      zeroY={scales.y(0)}
      value={formatter(y ?? null)}
    />
  {/if}

  <!-- Time comparison: primary point circle (right bar, index 1) -->
  {#if !isComparingDimension && showComparison && !currentPointIsNull}
    {@const primaryBarX = isBarMode
      ? slotCenterX - bandWidth / 2 + (singleBarWidth + barGap) + singleBarWidth / 2
      : $tweenedX}
    <circle
      cx={primaryBarX}
      cy={$tweenedY}
      r={4}
      class="fill-theme-500 stroke-surface-background stroke-[1.5px]"
    />
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
    {#each dimensionData as dim, i (i)}
      {@const pt = dim.data[hoveredIndex]}
      {@const barX = isBarMode
        ? slotCenterX - bandWidth / 2 + i * (singleBarWidth + barGap) + singleBarWidth / 2
        : $tweenedX}
      {#if pt?.value !== null && pt?.value !== undefined}
        <circle
          cx={barX}
          cy={scales.y(pt.value)}
          r={4}
          fill={dim.color}
          class="stroke-surface-background stroke-[1.5px]"
        />
      {/if}
    {/each}
  {/if}

  <!-- Time comparison: comparison point circle (left bar, index 0) -->
  {#if !isComparingDimension && showComparison && hasValidComparisonPoint}
    {@const compBarX = isBarMode
      ? slotCenterX - bandWidth / 2 + singleBarWidth / 2
      : $tweenedX}
    <circle
      cx={compBarX}
      cy={$tweenedComparisonY}
      r={4}
      class="fill-gray-500 stroke-surface-background stroke-[1.5px]"
    />
  {/if}
{/if}

<style lang="postcss">
  * {
    @apply pointer-events-none;
  }
</style>
