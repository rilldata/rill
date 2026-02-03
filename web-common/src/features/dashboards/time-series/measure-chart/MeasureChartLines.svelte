<script lang="ts">
  import { line, area, curveLinear } from "d3-shape";
  import { tweened } from "svelte/motion";
  import { cubicOut } from "svelte/easing";
  import { interpolatePath } from "d3-interpolate-path";
  import type { TimeSeriesPoint, DimensionSeriesData, ChartScales } from "./types";
  import { computeLineSegments, findSingletonPoints } from "./bisect";
  import {
    MainLineColor,
    LineMutedColor,
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    AreaMutedColorGradientDark,
    AreaMutedColorGradientLight,
    TimeComparisonLineColor,
  } from "../chart-colors";

  export let data: TimeSeriesPoint[];
  export let dimensionData: DimensionSeriesData[] = [];
  export let scales: ChartScales;
  export let showComparison: boolean = false;
  export let hasScrubSelection: boolean = false;
  export let scrubStart: Date | null = null;
  export let scrubEnd: Date | null = null;
  export let delay: number = 0;
  export let duration: number = 400;

  // Generate unique ID for this chart instance
  const chartId = Math.random().toString(36).substr(2, 9);

  // Line generator for main data
  $: lineGenerator = line<TimeSeriesPoint>()
    .x((d) => scales.x(d.tsPosition))
    .y((d) => scales.y(d.value ?? 0))
    .curve(curveLinear)
    .defined((d) => d.value !== null);

  // Area generator for main data
  $: areaGenerator = area<TimeSeriesPoint>()
    .x((d) => scales.x(d.tsPosition))
    .y0(scales.y(0))
    .y1((d) => scales.y(d.value ?? 0))
    .curve(curveLinear)
    .defined((d) => d.value !== null);

  // Comparison line generator
  $: comparisonLineGenerator = line<TimeSeriesPoint>()
    .x((d) => scales.x(d.tsPosition))
    .y((d) => scales.y(d.comparisonValue ?? 0))
    .curve(curveLinear)
    .defined((d) => d.comparisonValue !== null && d.comparisonValue !== undefined);

  // Colors based on scrub state
  $: mainLineColor = hasScrubSelection ? LineMutedColor : MainLineColor;
  $: areaGradientStart = hasScrubSelection
    ? AreaMutedColorGradientDark
    : MainAreaColorGradientDark;
  $: areaGradientEnd = hasScrubSelection
    ? AreaMutedColorGradientLight
    : MainAreaColorGradientLight;

  // Compute segments for handling gaps
  $: segments = computeLineSegments(data);
  $: singletons = findSingletonPoints(data);

  // Tweened path for smooth animation
  const tweenedLinePath = tweened("", {
    duration,
    easing: cubicOut,
    interpolate: interpolatePath,
  });

  const tweenedAreaPath = tweened("", {
    duration,
    easing: cubicOut,
    interpolate: interpolatePath,
  });

  // Update tweened paths when data changes
  $: {
    // Apply delay before updating paths
    setTimeout(() => {
      const newLinePath = lineGenerator(data) ?? "";
      const newAreaPath = areaGenerator(data) ?? "";
      tweenedLinePath.set(newLinePath);
      tweenedAreaPath.set(newAreaPath);
    }, delay);
  }

  // Compute scrub clip region
  $: scrubClipX = scrubStart && scrubEnd
    ? Math.min(scales.x(scrubStart), scales.x(scrubEnd))
    : 0;
  $: scrubClipWidth = scrubStart && scrubEnd
    ? Math.abs(scales.x(scrubEnd) - scales.x(scrubStart))
    : 0;

  // Is comparing dimensions
  $: isComparingDimension = dimensionData.length > 0;
</script>

<!-- Gradient definitions -->
<defs>
  <!-- Main area gradient -->
  <linearGradient id="area-gradient-{chartId}" x1="0" x2="0" y1="0" y2="1">
    <stop offset="5%" stop-color={areaGradientStart} stop-opacity="0.3" />
    <stop offset="95%" stop-color={areaGradientEnd} stop-opacity="0.3" />
  </linearGradient>

  <!-- Highlighted scrub area gradient -->
  <linearGradient id="scrub-area-gradient-{chartId}" x1="0" x2="0" y1="0" y2="1">
    <stop offset="5%" stop-color={MainAreaColorGradientDark} stop-opacity="0.3" />
    <stop offset="95%" stop-color={MainAreaColorGradientLight} stop-opacity="0.3" />
  </linearGradient>

  <!-- Clip path for segments (handles gaps in data) -->
  <clipPath id="segments-clip-{chartId}">
    {#each segments as segment (segment[0].ts.getTime())}
      {@const x = scales.x(segment[0].tsPosition)}
      {@const width = scales.x(segment[segment.length - 1].tsPosition) - x}
      <rect {x} y={0} height={scales.y.range()[0]} {width} />
    {/each}
  </clipPath>

  <!-- Clip path for scrub region -->
  {#if hasScrubSelection && scrubStart && scrubEnd}
    <clipPath id="scrub-clip-{chartId}">
      <rect
        x={scrubClipX}
        y={0}
        width={scrubClipWidth}
        height={scales.y.range()[0]}
      />
    </clipPath>
  {/if}
</defs>

<!-- Dimension comparison lines (multiple colored lines) -->
{#if isComparingDimension}
  {#each dimensionData as dim (dim.dimensionValue)}
    {#if dim.data.length > 0}
      <path
        d={lineGenerator(dim.data) ?? ""}
        stroke={dim.color}
        stroke-width={1.5}
        fill="none"
        class="transition-opacity"
        opacity={dim.isFetching ? 0.5 : 1}
      />
    {/if}
  {/each}
{:else}
  <!-- Main area fill -->
  <path
    d={$tweenedAreaPath}
    fill="url(#area-gradient-{chartId})"
    style="clip-path: url(#segments-clip-{chartId})"
  />

  <!-- Comparison line (if enabled) -->
  {#if showComparison}
    <path
      d={comparisonLineGenerator(data) ?? ""}
      stroke={TimeComparisonLineColor}
      stroke-width={1}
      stroke-dasharray="4,4"
      fill="none"
      class="transition-opacity"
    />
  {/if}

  <!-- Main line -->
  <path
    d={$tweenedLinePath}
    stroke={mainLineColor}
    stroke-width={1}
    fill="none"
    style="clip-path: url(#segments-clip-{chartId})"
  />

  <!-- Singleton points (single data points rendered as circles) -->
  {#each singletons as singleton (singleton.ts.getTime())}
    <circle
      cx={scales.x(singleton.tsPosition)}
      cy={scales.y(singleton.value ?? 0)}
      r={1.5}
      fill={mainLineColor}
    />
    <!-- Small bar below singleton -->
    <rect
      x={scales.x(singleton.tsPosition) - 0.75}
      y={Math.min(scales.y(0), scales.y(singleton.value ?? 0))}
      width={1.5}
      height={Math.abs(scales.y(0) - scales.y(singleton.value ?? 0))}
      fill={mainLineColor}
    />
  {/each}

  <!-- Highlighted scrub region (if scrub is active) -->
  {#if hasScrubSelection && scrubStart && scrubEnd}
    <!-- Highlighted area -->
    <path
      d={areaGenerator(data) ?? ""}
      fill="url(#scrub-area-gradient-{chartId})"
      style="clip-path: url(#scrub-clip-{chartId})"
    />

    <!-- Highlighted line -->
    <path
      d={lineGenerator(data) ?? ""}
      stroke={MainLineColor}
      stroke-width={1}
      fill="none"
      style="clip-path: url(#scrub-clip-{chartId})"
    />
  {/if}
{/if}
