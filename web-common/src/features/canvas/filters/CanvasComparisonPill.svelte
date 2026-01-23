<script lang="ts">
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getComparisonOptionsForCanvas } from "@rilldata/web-common/features/canvas/filters/util";
  import { Comparison } from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";

  export let minDate: DateTime<true> | undefined;
  export let maxDate: DateTime<true> | undefined;
  export let comparisonInterval: Interval<true> | undefined;
  export let comparisonRange: string | undefined;
  export let selectedRange: string | undefined;
  export let interval: Interval<true> | undefined;
  export let activeTimeGrain: V1TimeGrain | undefined;
  export let showFullRange = true;
  export let showTimeComparison = false;
  export let activeTimeZone: string;
  export let allowCustomTimeRange: boolean = true;
  export let minTimeGrain: V1TimeGrain | undefined;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let onDisplayTimeComparison: (show: boolean) => void;
  export let onSetSelectedComparisonRange: (range: TimeRange) => void;

  $: selectedComparisonTimeRange = comparisonInterval
    ? {
        name: comparisonRange,
        start: comparisonInterval.start.toJSDate(),
        end: comparisonInterval.end.toJSDate(),
        interval: activeTimeGrain,
      }
    : undefined;

  $: selectedTimeRange = interval
    ? {
        name: selectedRange,
        start: interval?.start.toJSDate(),
        end: interval?.end.toJSDate(),
        interval: activeTimeGrain,
      }
    : undefined;

  $: comparisonOptions = getComparisonOptionsForCanvas(
    selectedTimeRange,
    allowCustomTimeRange,
    activeTimeZone,
  );

  function onSelectComparisonRange(
    name: TimeComparisonOption,
    start: Date,
    end: Date,
  ) {
    onSetSelectedComparisonRange({
      name,
      start,
      end,
    });

    if (!showTimeComparison) {
      onDisplayTimeComparison(!showTimeComparison);
    }
  }

  $: disabled =
    selectedTimeRange?.name === TimeRangePreset.ALL_TIME || undefined;
</script>

<div
  class="wrapper"
  title={disabled && "Comparison not available when viewing all time range"}
>
  <button
    {disabled}
    class="flex gap-x-1.5 cursor-pointer"
    on:click={() => {
      onDisplayTimeComparison(!showTimeComparison);
    }}
    type="button"
    aria-label="Toggle time comparison"
  >
    <div class="pointer-events-none flex items-center gap-x-1.5">
      <Switch
        checked={showTimeComparison}
        id="comparing"
        small
        disabled={disabled ?? false}
      />

      <Label class="font-normal text-xs cursor-pointer" for="comparing">
        <span class:opacity-50={disabled}>Comparing</span>
      </Label>
    </div>
  </button>
  {#if activeTimeGrain && interval}
    <Comparison
      {minTimeGrain}
      maxDate={minDate}
      minDate={maxDate}
      timeGrain={activeTimeGrain}
      timeComparisonOptionsState={comparisonOptions}
      selectedComparison={selectedComparisonTimeRange}
      showComparison={showTimeComparison}
      currentInterval={interval}
      zone={activeTimeZone}
      {showFullRange}
      disabled={disabled ?? false}
      {allowCustomTimeRange}
      {side}
      {onSelectComparisonRange}
    />
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex w-fit;
    @apply h-[26px] rounded-full;
    @apply overflow-hidden select-none;
  }
  /* 
  :global(.wrapper > button) {
    @apply border;
  }

  :global(.wrapper > button:not(:first-child)) {
    @apply -ml-[1px];
  }

  :global(.wrapper > button) {
    @apply border;
    @apply px-2 flex items-center justify-center bg-surface-container;
  }

  :global(.wrapper > button:first-child) {
    @apply pl-2.5 rounded-l-full;
  }
  :global(.wrapper > button:last-child) {
    @apply pr-2.5 rounded-r-full;
  }

  :global(.wrapper > button:hover:not(:disabled)) {
    @apply bg-surface-background cursor-pointer;
  } */
</style>
