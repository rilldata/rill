<script lang="ts">
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getComparisonOptionsForCanvas } from "@rilldata/web-common/features/canvas/filters/util";
  import { Comparison } from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { Interval } from "luxon";
  import type { MinMax } from "../stores/time-control";

  export let interval: Interval<true> | undefined;
  export let range: string | undefined;
  export let comparisonRange: string | undefined;
  export let comparisonInterval: Interval<true> | undefined;
  export let minMaxTimeStamps: MinMax | undefined;
  export let activeTimeGrain: V1TimeGrain | undefined;
  export let showFullRange = true;
  export let showTimeComparison = false;
  export let activeTimeZone: string;
  export let allowCustomTimeRange: boolean = true;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let setComparison: (range: boolean | string) => void;

  $: comparisonOptions = getComparisonOptionsForCanvas(interval, range);

  $: disabled = range === TimeRangePreset.ALL_TIME || undefined;
</script>

<div
  class="wrapper"
  title={disabled && "Comparison not available when viewing all time range"}
>
  <button
    {disabled}
    class="flex gap-x-1.5 cursor-pointer"
    on:click={() => {
      setComparison(!showTimeComparison);
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
  {#if activeTimeGrain && interval?.isValid}
    <Comparison
      minDate={minMaxTimeStamps?.min}
      maxDate={minMaxTimeStamps?.max}
      {comparisonOptions}
      {comparisonInterval}
      {comparisonRange}
      showComparison={showTimeComparison}
      currentInterval={interval}
      grain={activeTimeGrain}
      zone={activeTimeZone}
      {showFullRange}
      disabled={disabled ?? false}
      {allowCustomTimeRange}
      {side}
      onSelectComparisonString={setComparison}
    />
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex w-fit;
    @apply h-7 rounded-full;
    @apply overflow-hidden select-none;
  }

  :global(.wrapper > button) {
    @apply border;
  }

  :global(.wrapper > button:not(:first-child)) {
    @apply -ml-[1px];
  }

  :global(.wrapper > button) {
    @apply border;
    @apply px-2 flex items-center justify-center bg-surface;
  }

  :global(.wrapper > button:first-child) {
    @apply pl-2.5 rounded-l-full;
  }
  :global(.wrapper > button:last-child) {
    @apply pr-2.5 rounded-r-full;
  }

  :global(.wrapper > button:hover:not(:disabled)) {
    @apply bg-gray-50 cursor-pointer;
  }

  /* Doest apply to all instances except alert/report. So this seems unintentional
  :global(.wrapper > [data-state="open"]) {
    @apply bg-gray-50 border-gray-400 z-50;
  }
  */
</style>
