<script lang="ts">
  import {
    createAreaGenerator,
    createLineGenerator,
  } from "@rilldata/web-common/components/data-graphic/utils";
  import type {
    ChartScales,
    ChartSeries,
  } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";
  import { bridgeSmallGaps } from "./sparse-data-utils";

  const numAccessor = (d: number | null) => d;
  const numClone = (_d: number | null, v: number): number | null => v;

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
    ? bridgeSmallGaps(
        primarySeries.values,
        numAccessor,
        numClone,
        scales.x,
        connectNulls,
      )
    : {
        values: [] as (number | null)[],
        inputSegments: [],
        bridgedSegments: [],
      };
  $: primaryBridged = primaryBridgeResult.values;

  $: primaryLinePath = primarySeries ? (lineGen(primaryBridged) ?? "") : "";
  $: primaryAreaPath = primarySeries ? (areaGen(primaryBridged) ?? "") : "";

  $: primaryRealSegments = primaryBridgeResult.inputSegments;
  $: primarySegments = primaryBridgeResult.bridgedSegments;
  $: primarySingletons = primarySeries
    ? primarySegments
        .filter((s) => s.startIndex === s.endIndex)
        .map((s) => s.startIndex)
    : [];

  $: secondarySeries = series.slice(1).map((s) => {
    const bridged = bridgeSmallGaps(
      s.values,
      numAccessor,
      numClone,
      scales.x,
      connectNulls,
    );
    const singletons = bridged.bridgedSegments
      .filter((seg) => seg.startIndex === seg.endIndex)
      .map((seg) => seg.startIndex);
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
