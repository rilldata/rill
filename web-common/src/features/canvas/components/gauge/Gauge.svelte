<script lang="ts">
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import {
    type MetricsViewSpecMeasure,
    type V1MetricsViewAggregationResponse,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import { tweened } from "svelte/motion";
  import { cubicOut } from "svelte/easing";

  type AggregationQuery = QueryObserverResult<
    V1MetricsViewAggregationResponse,
    HTTPError
  >;

  export let totalResult: AggregationQuery;
  export let measure: MetricsViewSpecMeasure | undefined;

  $: measureValueFormatter = measure
    ? createMeasureValueFormatter<null>(measure, "big-number")
    : () => "no data";

  $: currentValue = (totalResult?.data?.data?.[0]?.[
    measure?.name ?? ""
  ] ?? null) as number | null;

  // Extract target data from aggregation response
  $: targetsData = totalResult?.data?.targets;
  $: targetForMeasure = targetsData?.find(
    (t) => t.measure === measure?.name,
  );
  $: targetValue = targetForMeasure?.values?.[0]?.["value"] as
    | number
    | null
    | undefined;
  $: targetName =
    targetForMeasure?.target?.targetName ||
    targetForMeasure?.target?.name ||
    undefined;

  // Calculate percentage for gauge (value / target * 100, capped at 200%)
  $: percentage =
    currentValue != null && targetValue != null && targetValue !== 0
      ? Math.min(200, (currentValue / targetValue) * 100)
      : currentValue != null && targetValue === null
        ? null
        : null;

  // Animate the gauge value
  const animatedPercentage = tweened(0, {
    duration: 800,
    easing: cubicOut,
  });

  $: if (percentage !== null) {
    animatedPercentage.set(percentage);
  }

  // Gauge configuration
  const radius = 80;
  const strokeWidth = 16;
  const centerX = 100;
  const centerY = 100;
  const startAngle = -135; // Start angle in degrees
  const endAngle = 135; // End angle in degrees
  const totalAngle = endAngle - startAngle; // 270 degrees

  // Convert percentage to angle (0-200% maps to 0-270 degrees)
  $: currentAngle =
    percentage !== null
      ? ($animatedPercentage / 200) * totalAngle + startAngle
      : startAngle;

  // Calculate arc path
  function getArcPath(startAngleDeg: number, endAngleDeg: number): string {
    const start = polarToCartesian(
      centerX,
      centerY,
      radius,
      startAngleDeg,
    );
    const end = polarToCartesian(centerX, centerY, radius, endAngleDeg);
    const angleDiff = endAngleDeg - startAngleDeg;
    const largeArcFlag = Math.abs(angleDiff) > 180 ? "1" : "0";

    return [
      "M",
      start.x,
      start.y,
      "A",
      radius,
      radius,
      0,
      largeArcFlag,
      1, // clockwise
      end.x,
      end.y,
    ].join(" ");
  }

  function polarToCartesian(
    centerX: number,
    centerY: number,
    radius: number,
    angleInDegrees: number,
  ) {
    const angleInRadians = ((angleInDegrees - 90) * Math.PI) / 180.0;

    return {
      x: centerX + radius * Math.cos(angleInRadians),
      y: centerY + radius * Math.sin(angleInRadians),
    };
  }

  $: backgroundArc = getArcPath(startAngle, endAngle);
  $: valueArc =
    percentage !== null && percentage > 0
      ? getArcPath(startAngle, Math.min(endAngle, currentAngle))
      : "";

  // Color based on percentage
  $: gaugeColor =
    percentage === null
      ? "#9CA3AF"
      : percentage < 50
        ? "#EF4444" // red
        : percentage < 100
          ? "#F59E0B" // amber
          : percentage < 150
            ? "#10B981" // green
            : "#3B82F6"; // blue

  $: isLoading = totalResult.isLoading;
  $: isError = totalResult.isError;
</script>

<div class="gauge-container">
  {#if isLoading}
    <div class="loading">Loading...</div>
  {:else if isError}
    <div class="error">Error loading data</div>
  {:else if currentValue === null}
    <div class="no-data">No data available</div>
  {:else if targetValue === null || targetValue === undefined}
    <div class="no-target">
      <div class="no-target-value">{measureValueFormatter(currentValue)}</div>
      <div class="no-target-message">No target configured</div>
    </div>
  {:else}
    <div class="gauge-wrapper">
      <svg
        viewBox="0 0 200 140"
        class="gauge-svg"
        xmlns="http://www.w3.org/2000/svg"
      >
        <!-- Background arc -->
        <path
          d={backgroundArc}
          fill="none"
          stroke="#E5E7EB"
          stroke-width={strokeWidth}
          stroke-linecap="round"
        />

        <!-- Value arc -->
        {#if percentage !== null && percentage > 0 && valueArc}
          <path
            d={valueArc}
            fill="none"
            stroke={gaugeColor}
            stroke-width={strokeWidth}
            stroke-linecap="round"
            class="gauge-value-arc"
          />
        {/if}

        <!-- Center value display -->
        <text
          x={centerX}
          y={centerY - 10}
          text-anchor="middle"
          class="gauge-value"
        >
          {measureValueFormatter(currentValue)}
        </text>

        <!-- Target display -->
        <text
          x={centerX}
          y={centerY + 15}
          text-anchor="middle"
          class="gauge-target"
        >
          {targetName || "Target"}: {measureValueFormatter(targetValue)}
        </text>
      </svg>

      <!-- Percentage indicator below gauge -->
      {#if percentage !== null}
        <div class="gauge-percentage" style="color: {gaugeColor};">
          {Math.round(percentage)}%
        </div>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .gauge-container {
    @apply flex items-center justify-center w-full h-full p-4;
    min-height: 200px;
  }

  .gauge-wrapper {
    @apply flex flex-col items-center;
  }

  .gauge-svg {
    @apply w-full max-w-xs;
    height: auto;
  }

  .gauge-value {
    @apply text-2xl font-semibold fill-foreground;
    font-family: system-ui, -apple-system, sans-serif;
  }

  .gauge-target {
    @apply text-sm fill-muted-foreground;
    font-family: system-ui, -apple-system, sans-serif;
  }

  .gauge-percentage {
    @apply mt-2 text-lg font-medium;
    color: var(--gauge-color, #9CA3AF);
  }

  .gauge-value-arc {
    transition: stroke-dasharray 0.3s ease;
  }

  .loading,
  .error,
  .no-data {
    @apply text-muted-foreground;
  }

  .no-target {
    @apply flex flex-col items-center justify-center gap-2;
  }

  .no-target-value {
    @apply text-3xl font-semibold;
  }

  .no-target-message {
    @apply text-sm text-muted-foreground;
  }
</style>

