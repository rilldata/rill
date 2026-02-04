<script lang="ts">
  import {
    createLineGenerator,
    createAreaGenerator,
  } from "@rilldata/web-common/components/data-graphic/utils";
  import type {
    ChartSeries,
    ChartScales,
  } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";

  export let series: ChartSeries[];
  export let scales: ChartScales;
  export let hasScrubSelection: boolean = false;
  export let scrubStartIndex: number | null = null;
  export let scrubEndIndex: number | null = null;

  const chartId = Math.random().toString(36).slice(2, 11);

  // Index-based line/area generators — x is just the index
  $: lineGen = createLineGenerator<number | null>({
    x: (_d, i) => scales.x(i),
    y: (d) => scales.y(d ?? 0),
    defined: (d) => d !== null,
  });

  $: areaGen = createAreaGenerator<number | null>({
    x: (_d, i) => scales.x(i),
    y0: scales.y(0),
    y1: (d) => scales.y(d ?? 0),
    defined: (d) => d !== null,
  });

  // Contiguous non-null segments as {startIndex, endIndex} pairs
  function computeSegments(
    values: (number | null)[],
  ): { startIndex: number; endIndex: number }[] {
    const segments: { startIndex: number; endIndex: number }[] = [];
    let segStart = -1;
    for (let i = 0; i < values.length; i++) {
      if (values[i] !== null) {
        if (segStart === -1) segStart = i;
      } else if (segStart !== -1) {
        segments.push({ startIndex: segStart, endIndex: i - 1 });
        segStart = -1;
      }
    }
    if (segStart !== -1)
      segments.push({ startIndex: segStart, endIndex: values.length - 1 });
    return segments;
  }

  function findSingletonIndices(values: (number | null)[]): number[] {
    return computeSegments(values)
      .filter((s) => s.startIndex === s.endIndex)
      .map((s) => s.startIndex);
  }

  $: primarySeries = series[0];

  // Direct path computation — no tweening
  $: primaryLinePath = primarySeries
    ? (lineGen(primarySeries.values) ?? "")
    : "";
  $: primaryAreaPath = primarySeries
    ? (areaGen(primarySeries.values) ?? "")
    : "";

  $: primarySegments = primarySeries
    ? computeSegments(primarySeries.values)
    : [];
  $: primarySingletons = primarySeries
    ? findSingletonIndices(primarySeries.values)
    : [];

  // Scrub clip region (index-based)
  $: scrubClipX =
    scrubStartIndex !== null && scrubEndIndex !== null
      ? Math.min(scales.x(scrubStartIndex), scales.x(scrubEndIndex))
      : 0;
  $: scrubClipWidth =
    scrubStartIndex !== null && scrubEndIndex !== null
      ? Math.abs(scales.x(scrubEndIndex) - scales.x(scrubStartIndex))
      : 0;

  // Muted colors for scrub state
  $: primaryLineColor = primarySeries
    ? hasScrubSelection
      ? "var(--color-gray-500)"
      : primarySeries.color
    : "var(--color-gray-500)";

  $: primaryAreaStart = primarySeries?.areaGradient
    ? hasScrubSelection
      ? "var(--color-gray-300)"
      : primarySeries.areaGradient.dark
    : "transparent";
  $: primaryAreaEnd = primarySeries?.areaGradient
    ? hasScrubSelection
      ? "var(--color-gray-50)"
      : primarySeries.areaGradient.light
    : "transparent";
</script>

<defs>
  {#if primarySeries?.areaGradient}
    <linearGradient id="area-grad-{chartId}" x1="0" x2="0" y1="0" y2="1">
      <stop offset="5%" stop-color={primaryAreaStart} stop-opacity="0.3" />
      <stop offset="95%" stop-color={primaryAreaEnd} stop-opacity="0.3" />
    </linearGradient>
    <linearGradient id="scrub-area-grad-{chartId}" x1="0" x2="0" y1="0" y2="1">
      <stop
        offset="5%"
        stop-color={primarySeries.areaGradient.dark}
        stop-opacity="0.3"
      />
      <stop
        offset="95%"
        stop-color={primarySeries.areaGradient.light}
        stop-opacity="0.3"
      />
    </linearGradient>
  {/if}

  {#if primarySeries}
    <clipPath id="seg-clip-{chartId}">
      {#each primarySegments as seg (seg.startIndex)}
        {@const x = scales.x(seg.startIndex)}
        {@const width = scales.x(seg.endIndex) - x}
        <rect {x} y={0} height={scales.y.range()[0]} {width} />
      {/each}
    </clipPath>
  {/if}

  {#if hasScrubSelection && scrubStartIndex !== null && scrubEndIndex !== null}
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

<!-- Area fill for primary series -->
{#if primarySeries?.areaGradient}
  <path
    d={primaryAreaPath}
    fill="url(#area-grad-{chartId})"
    style="clip-path: url(#seg-clip-{chartId})"
  />
{/if}

<!-- Secondary series -->
{#each series.slice(1) as s (s.id)}
  <path
    d={lineGen(s.values) ?? ""}
    stroke={hasScrubSelection ? "var(--color-gray-400)" : s.color}
    stroke-width={s.strokeWidth ?? 1}
    stroke-dasharray={s.strokeDasharray ?? "none"}
    fill="none"
    opacity={s.opacity ?? 1}
  />
  {#if hasScrubSelection && scrubStartIndex !== null && scrubEndIndex !== null}
    <path
      d={lineGen(s.values) ?? ""}
      stroke={s.color}
      stroke-width={s.strokeWidth ?? 1}
      stroke-dasharray={s.strokeDasharray ?? "none"}
      fill="none"
      opacity={s.opacity ?? 1}
      style="clip-path: url(#scrub-clip-{chartId})"
    />
  {/if}
{/each}

<!-- Primary line -->
{#if primarySeries}
  <path
    d={primaryLinePath}
    stroke={primaryLineColor}
    stroke-width={primarySeries.strokeWidth ?? 1}
    fill="none"
    style="clip-path: url(#seg-clip-{chartId})"
  />

  <!-- Singleton points -->
  {#each primarySingletons as idx (idx)}
    {@const v = primarySeries.values[idx] ?? 0}
    <circle
      cx={scales.x(idx)}
      cy={scales.y(v)}
      r={1.5}
      fill={primaryLineColor}
    />
    <rect
      x={scales.x(idx) - 0.75}
      y={Math.min(scales.y(0), scales.y(v))}
      width={1.5}
      height={Math.abs(scales.y(0) - scales.y(v))}
      fill={primaryLineColor}
    />
  {/each}
{/if}

<!-- Highlighted scrub region -->
{#if hasScrubSelection && scrubStartIndex !== null && scrubEndIndex !== null && primarySeries}
  {#if primarySeries.areaGradient}
    <path
      d={areaGen(primarySeries.values) ?? ""}
      fill="url(#scrub-area-grad-{chartId})"
      style="clip-path: url(#scrub-clip-{chartId})"
    />
  {/if}
  <path
    d={lineGen(primarySeries.values) ?? ""}
    stroke={primarySeries.color}
    stroke-width={1}
    fill="none"
    style="clip-path: url(#scrub-clip-{chartId})"
  />
{/if}

<style lang="postcss">
  * {
    @apply pointer-events-none;
  }
</style>
