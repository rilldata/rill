<script lang="ts">
  import {
    createLineGenerator,
    createAreaGenerator,
  } from "@rilldata/web-common/components/data-graphic/utils";
  import type {
    ChartSeries,
    ChartScales,
  } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";

  /**
   * Rendering sparse data with null gaps
   *
   * 1. Null bridging: `bridgeSmallGaps` linearly interpolates across small
   *    gaps (< MAX_BRIDGE_GAP_PX) when `connectNulls` is on. Large gaps
   *    remain as nulls and produce natural line breaks.
   *
   * 2. Clip paths: The primary series needs clip paths because its area
   *    fill gradient would otherwise render across gaps (`defined` only
   *    affects line generators, not the filled path).
   *      - `seg-clip`:   real data segments only (connectNulls off)
   *      - `full-clip`:  real + bridged segments (connectNulls on, area fill)
   *      - `scrub-clip`: scrub selection rect — chart draws muted, then
   *                      re-draws with original colors inside this clip
   *    Secondary series have no area fill, so they rely on the line
   *    generator's `defined` callback and only use `scrub-clip`.
   *
   * 3. Singletons: When `connectNulls` is off, isolated points (no adjacent
   *    non-null neighbors) are drawn as circles since there's no line
   *    segment to render.
   */

  const MAX_BRIDGE_GAP_PX = 40;

  interface Segment {
    startIndex: number;
    endIndex: number;
  }

  interface BridgeResult {
    values: (number | null)[];
    inputSegments: Segment[];
  }

  const chartId = Math.random().toString(36).slice(2, 11);

  export let series: ChartSeries[];
  export let scales: ChartScales;
  export let hasScrubSelection: boolean = false;
  export let scrubStartIndex: number | null = null;
  export let scrubEndIndex: number | null = null;
  export let connectNulls: boolean = true;

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

  $: primaryBridgeResult = primarySeries
    ? bridgeSmallGaps(primarySeries.values, scales.x, connectNulls)
    : { values: [] as (number | null)[], inputSegments: [] };
  $: primaryBridged = primaryBridgeResult.values;

  $: primaryLinePath = primarySeries ? (lineGen(primaryBridged) ?? "") : "";
  $: primaryAreaPath = primarySeries ? (areaGen(primaryBridged) ?? "") : "";

  $: primaryRealSegments = primaryBridgeResult.inputSegments;
  $: primarySegments = primarySeries ? computeSegments(primaryBridged) : [];
  $: primarySingletons =
    primarySeries && !connectNulls
      ? primarySegments
          .filter((s) => s.startIndex === s.endIndex)
          .map((s) => s.startIndex)
      : [];

  $: secondarySeries = series.slice(1).map((s) => {
    const bridged = bridgeSmallGaps(s.values, scales.x, connectNulls);
    const singletons = !connectNulls
      ? computeSegments(bridged.values)
          .filter((seg) => seg.startIndex === seg.endIndex)
          .map((seg) => seg.startIndex)
      : [];
    return { ...s, bridgedValues: bridged.values, singletons };
  });

  $: scrubClipX =
    scrubStartIndex !== null && scrubEndIndex !== null
      ? Math.min(scales.x(scrubStartIndex), scales.x(scrubEndIndex))
      : 0;
  $: scrubClipWidth =
    scrubStartIndex !== null && scrubEndIndex !== null
      ? Math.abs(scales.x(scrubEndIndex) - scales.x(scrubStartIndex))
      : 0;

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

  function bridgeSmallGaps(
    values: (number | null)[],
    xScale: (i: number) => number,
    shouldBridge: boolean,
  ): BridgeResult {
    const inputSegments = computeSegments(values);

    if (!shouldBridge || values.length < 3 || inputSegments.length <= 1) {
      return { values, inputSegments };
    }

    const result = [...values];

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
      }
    }

    return { values: result, inputSegments };
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
    <!-- seg-clip: real data segments only (connectNulls off) -->
    <clipPath id="seg-clip-{chartId}">
      {#each primaryRealSegments as seg (seg.startIndex)}
        {@const x = scales.x(seg.startIndex)}
        {@const width = scales.x(seg.endIndex) - x}
        <rect {x} y={0} height={scales.y.range()[0]} {width} />
      {/each}
    </clipPath>
    <!-- full-clip: real + bridged segments (connectNulls on, area fill) -->
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

<!-- Primary area fill (clipped to full-clip to avoid filling across large gaps) -->
{#if primarySeries?.areaGradient}
  <path
    d={primaryAreaPath}
    fill="url(#area-grad-{chartId})"
    style="clip-path: url(#full-clip-{chartId})"
  />
{/if}

<!-- Secondary series (no clip paths — line breaks handled by `defined` callback) -->
{#each secondarySeries as s (s.id)}
  <path
    d={lineGen(s.bridgedValues) ?? ""}
    stroke={hasScrubSelection ? "var(--color-gray-400)" : s.color}
    stroke-width={s.strokeWidth ?? 1}
    stroke-dasharray={s.strokeDasharray ?? "none"}
    fill="none"
    opacity={s.opacity ?? 1}
  />
  {#each s.singletons as idx (idx)}
    {@const v = s.values[idx] ?? 0}
    <circle
      cx={scales.x(idx)}
      cy={scales.y(v)}
      r={1.5}
      fill={hasScrubSelection ? "var(--color-gray-400)" : s.color}
      opacity={s.opacity ?? 1}
    />
  {/each}
  {#if hasScrubSelection && scrubStartIndex !== null && scrubEndIndex !== null}
    <g style="clip-path: url(#scrub-clip-{chartId})">
      <path
        d={lineGen(s.bridgedValues) ?? ""}
        stroke={s.color}
        stroke-width={s.strokeWidth ?? 1}
        stroke-dasharray={s.strokeDasharray ?? "none"}
        fill="none"
        opacity={s.opacity ?? 1}
      />
      {#each s.singletons as idx (idx)}
        {@const v = s.values[idx] ?? 0}
        <circle
          cx={scales.x(idx)}
          cy={scales.y(v)}
          r={1.5}
          fill={s.color}
          opacity={s.opacity ?? 1}
        />
      {/each}
    </g>
  {/if}
{/each}

<!-- Primary line (clipped to seg-clip or full-clip depending on connectNulls) -->
{#if primarySeries}
  {#if connectNulls}
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

<!-- Scrub highlight: re-draws primary series with original colors, clipped to scrub-clip -->
{#if hasScrubSelection && scrubStartIndex !== null && scrubEndIndex !== null && primarySeries}
  <g style="clip-path: url(#scrub-clip-{chartId})">
    {#if primarySeries.areaGradient}
      <path
        d={areaGen(primaryBridged) ?? ""}
        fill="url(#scrub-area-grad-{chartId})"
      />
    {/if}
    <path
      d={lineGen(primaryBridged) ?? ""}
      stroke={primarySeries.color}
      stroke-width={1}
      fill="none"
    />
    {#each primarySingletons as idx (idx)}
      {@const v = primarySeries.values[idx] ?? 0}
      <circle
        cx={scales.x(idx)}
        cy={scales.y(v)}
        r={1.5}
        fill={primarySeries.color}
      />
    {/each}
  </g>
{/if}

<style lang="postcss">
  * {
    @apply pointer-events-none;
  }
</style>
