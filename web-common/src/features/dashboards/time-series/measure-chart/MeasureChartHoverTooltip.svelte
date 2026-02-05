<script lang="ts">
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import { formatGrainBucket } from "@rilldata/web-common/lib/time/ranges/formatter";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { type DateTime, type Interval } from "luxon";

  interface DimTooltipEntry {
    label: string;
    value: number | null;
    color: string;
  }

  export let mouseX: number;
  export let mouseY: number;
  export let currentValue: number | null;
  export let comparisonValue: number | null;
  export let currentTs: DateTime;
  export let comparisonTs: DateTime | undefined;
  export let timeGranularity: V1TimeGrain | undefined;
  export let interval: Interval<true> | undefined = undefined;
  export let comparisonInterval: Interval<true> | undefined = undefined;
  export let isComparingDimension: boolean;
  export let dimTooltipEntries: DimTooltipEntry[] = [];
  export let deltaLabel: string | null;
  export let deltaPositive: boolean;
  export let formatter: (value: number | null) => string;

  const offset = 4;

  $: absoluteDelta =
    currentValue !== null && comparisonValue !== null
      ? currentValue - comparisonValue
      : null;
</script>

<div
  use:portal
  class="tooltip-container"
  style:top="{mouseY + offset}px"
  style:left="{mouseX + offset}px"
>
  {#if isComparingDimension}
    <div class="dimension-tooltip">
      <div class="dimension-date">
        {formatGrainBucket(currentTs, timeGranularity, interval)}
      </div>
      {#each dimTooltipEntries as entry (entry.label)}
        <div class="dimension-entry">
          <span class="dimension-dot" style:background-color={entry.color} />
          <span class="dimension-label">{entry.label}</span>
          <span class="dimension-value">{formatter(entry.value)}</span>
        </div>
      {/each}
    </div>
  {:else}
    <div class="time-comparison">
      <div class="period current">
        <span class="value primary-value">{formatter(currentValue)}</span>
        <span class="date">
          {formatGrainBucket(currentTs, timeGranularity, interval)}</span
        >
      </div>

      <div class="divider">
        <div class="divider-line" />
        <span class="vs-badge">vs</span>
      </div>

      <div class="period comparison">
        <span class="value">{formatter(comparisonValue)}</span>
        <span class="date">
          {#if comparisonTs}
            {formatGrainBucket(
              comparisonTs,
              timeGranularity,
              comparisonInterval,
            )}
          {/if}
        </span>
      </div>
    </div>

    {#if absoluteDelta !== null && deltaLabel}
      <div
        class="delta-footer"
        class:positive={deltaPositive}
        class:negative={!deltaPositive}
      >
        <span class="delta-arrow">{deltaPositive ? "▲" : "▼"}</span>
        <span class="delta-absolute"
          >{deltaPositive ? "+" : ""}{formatter(absoluteDelta)}</span
        >
        <span class="delta-percent">({deltaLabel})</span>
      </div>
    {/if}
  {/if}
</div>

<style lang="postcss">
  .tooltip-container {
    @apply z-50 shadow-md bg-surface-subtle border rounded fixed pointer-events-none overflow-hidden;
  }

  /* Dimension comparison styles */
  .dimension-tooltip {
    @apply px-2 py-1.5 text-[11px];
  }

  .dimension-date {
    @apply text-fg-muted text-[10px] mb-1;
  }

  .dimension-entry {
    @apply flex gap-x-1.5 items-center;
  }

  .dimension-dot {
    @apply size-[7px] rounded-full flex-shrink-0;
  }

  .dimension-label {
    @apply text-fg-muted truncate max-w-[120px];
  }

  .dimension-value {
    @apply font-semibold text-fg-secondary ml-auto;
  }

  /* Time comparison styles */
  .time-comparison {
    @apply flex items-center;
  }

  .period {
    @apply flex flex-col items-center px-2.5 py-1.5;
  }

  .period .value {
    @apply text-[12px];
  }

  .period.current .value {
    @apply font-semibold text-theme-700;
  }

  .period.comparison .value {
    @apply font-medium text-fg-muted;
  }

  .period .date {
    @apply text-[9px] text-fg-muted;
  }

  .divider {
    @apply relative flex items-center justify-center self-stretch;
  }

  .divider-line {
    @apply absolute inset-y-0 w-px bg-gray-200;
  }

  .vs-badge {
    @apply relative size-5 rounded-full border border-gray-300;
    @apply flex items-center justify-center;
    @apply text-[8px] text-fg-muted font-medium bg-surface-background;
  }

  /* Delta footer styles */
  .delta-footer {
    @apply flex items-center justify-center gap-x-1.5 px-2 py-1 border-t text-[10px];
  }

  .delta-footer.positive {
    @apply bg-green-50;
  }

  .delta-footer.negative {
    @apply bg-red-50;
  }

  .delta-footer.positive .delta-arrow,
  .delta-footer.positive .delta-absolute,
  .delta-footer.positive .delta-percent {
    @apply text-green-600;
  }

  .delta-footer.negative .delta-arrow,
  .delta-footer.negative .delta-absolute,
  .delta-footer.negative .delta-percent {
    @apply text-red-600;
  }
</style>
