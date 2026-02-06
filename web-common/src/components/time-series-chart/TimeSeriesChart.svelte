<script lang="ts">
  import {
    createLineGenerator,
    createAreaGenerator,
  } from "@rilldata/web-common/components/data-graphic/utils";
  import type {
    ChartSeries,
    ChartScales,
  } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";

  const MAX_BRIDGE_GAP_PX = 40;

  interface Segment {
    startIndex: number;
    endIndex: number;
  }

  interface BridgeResult {
    values: (number | null)[];
    bridges: Segment[];

    inputSegments: Segment[];
  }

  const chartId = Math.random().toString(36).slice(2, 11);

  export let series: ChartSeries[];
  export let scales: ChartScales;
  export let hasScrubSelection: boolean = false;
  export let scrubStartIndex: number | null = null;
  export let scrubEndIndex: number | null = null;
  export let connectNulls: boolean = true;

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

  $: primarySeries = series[0];

  // Bridge small gaps for smoother rendering
  $: primaryBridgeResult = primarySeries
    ? bridgeSmallGaps(primarySeries.values, scales.x, connectNulls)
    : { values: [] as (number | null)[], bridges: [], inputSegments: [] };
  $: primaryBridged = primaryBridgeResult.values;
  $: primaryBridges = primaryBridgeResult.bridges;

  // Direct path computation — no tweening
  $: primaryLinePath = primarySeries ? (lineGen(primaryBridged) ?? "") : "";
  $: primaryAreaPath = primarySeries ? (areaGen(primaryBridged) ?? "") : "";

  // Original segments (real data only) — reused from bridgeSmallGaps
  $: primaryRealSegments = primaryBridgeResult.inputSegments;
  // Merged segments (real + bridged) — used for area fill clip
  $: primarySegments = primarySeries ? computeSegments(primaryBridged) : [];
  // Singletons from bridged data — only shown when connectNulls is off
  $: primarySingletons =
    primarySeries && !connectNulls
      ? primarySegments
          .filter((s) => s.startIndex === s.endIndex)
          .map((s) => s.startIndex)
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

  // Contiguous non-null segments as {startIndex, endIndex} pairs
  function computeSegments(values: (number | null)[]): Segment[] {
    const segments: Segment[] = [];
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

  /**
   * Bridge small gaps between non-null segments by linearly interpolating.
   * Returns the interpolated values, the bridged gap regions, and the
   * original input segments (so callers don't need to recompute them).
   */
  function bridgeSmallGaps(
    values: (number | null)[],
    xScale: (i: number) => number,
    shouldBridge: boolean,
  ): BridgeResult {
    const inputSegments = computeSegments(values);

    if (!shouldBridge || values.length < 3 || inputSegments.length <= 1) {
      return { values, bridges: [], inputSegments };
    }

    const result = [...values];
    const bridges: Segment[] = [];

    for (let i = 0; i < inputSegments.length - 1; i++) {
      const prev = inputSegments[i];
      const next = inputSegments[i + 1];
      const gapPx = xScale(next.startIndex) - xScale(prev.endIndex);

      if (gapPx <= MAX_BRIDGE_GAP_PX) {
        const v0 = values[prev.endIndex]!;
        const v1 = values[next.startIndex]!;
        const span = next.startIndex - prev.endIndex;
        for (let j = prev.endIndex + 1; j < next.startIndex; j++) {
          const t = (j - prev.endIndex) / span;
          result[j] = v0 + t * (v1 - v0);
        }
        bridges.push({ startIndex: prev.endIndex, endIndex: next.startIndex });
      }
    }

    return { values: result, bridges, inputSegments };
  }
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
    <!-- Clip for real data segments (solid line) -->
    <clipPath id="seg-clip-{chartId}">
      {#each primaryRealSegments as seg (seg.startIndex)}
        {@const x = scales.x(seg.startIndex)}
        {@const width = scales.x(seg.endIndex) - x}
        <rect {x} y={0} height={scales.y.range()[0]} {width} />
      {/each}
    </clipPath>
    <!-- Clip for bridged gap regions (dashed line) -->
    {#if primaryBridges.length > 0}
      <clipPath id="bridge-clip-{chartId}">
        {#each primaryBridges as bridge (bridge.startIndex)}
          {@const x = scales.x(bridge.startIndex)}
          {@const width = scales.x(bridge.endIndex) - x}
          <rect {x} y={0} height={scales.y.range()[0]} {width} />
        {/each}
      </clipPath>
    {/if}
    <!-- Clip for all rendered segments (real + bridged, for area fill) -->
    <clipPath id="full-clip-{chartId}">
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
    style="clip-path: url(#full-clip-{chartId})"
  />
{/if}

<!-- Secondary series -->
{#each series.slice(1) as s (s.id)}
  {@const bridged = bridgeSmallGaps(s.values, scales.x, connectNulls)}
  <path
    d={lineGen(bridged.values) ?? ""}
    stroke={hasScrubSelection ? "var(--color-gray-400)" : s.color}
    stroke-width={s.strokeWidth ?? 1}
    stroke-dasharray={s.strokeDasharray ?? "none"}
    fill="none"
    opacity={s.opacity ?? 1}
  />
  {#if hasScrubSelection && scrubStartIndex !== null && scrubEndIndex !== null}
    <path
      d={lineGen(bridged.values) ?? ""}
      stroke={s.color}
      stroke-width={s.strokeWidth ?? 1}
      stroke-dasharray={s.strokeDasharray ?? "none"}
      fill="none"
      opacity={s.opacity ?? 1}
      style="clip-path: url(#scrub-clip-{chartId})"
    />
  {/if}
{/each}

<!-- Primary line (solid for real data) -->
{#if primarySeries}
  {#if connectNulls}
    <!-- When connecting nulls, draw the full bridged line as solid -->
    <path
      d={primaryLinePath}
      stroke={primaryLineColor}
      stroke-width={primarySeries.strokeWidth ?? 1}
      fill="none"
      style="clip-path: url(#full-clip-{chartId})"
    />
  {:else}
    <path
      d={primaryLinePath}
      stroke={primaryLineColor}
      stroke-width={primarySeries.strokeWidth ?? 1}
      fill="none"
      style="clip-path: url(#seg-clip-{chartId})"
    />
  {/if}

  <!-- Singleton points -->
  {#each primarySingletons as idx (idx)}
    {@const v = primarySeries.values[idx] ?? 0}
    <circle
      cx={scales.x(idx)}
      cy={scales.y(v)}
      r={1.5}
      fill={primaryLineColor}
    />
  {/each}
{/if}

<!-- Highlighted scrub region -->
{#if hasScrubSelection && scrubStartIndex !== null && scrubEndIndex !== null && primarySeries}
  {#if primarySeries.areaGradient}
    <path
      d={areaGen(primaryBridged) ?? ""}
      fill="url(#scrub-area-grad-{chartId})"
      style="clip-path: url(#scrub-clip-{chartId})"
    />
  {/if}
  <path
    d={lineGen(primaryBridged) ?? ""}
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
