<script lang="ts">
  import {
    TimeRangePreset,
    type TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import type {
    V1ExploreTimeRange,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import {
    bucketYamlRanges,
    type ISODurationString,
    type NamedRange,
  } from "../new-time-controls";
  import TimeGrainSelector from "../TimeGrainSelector.svelte";
  import * as Elements from "./components";
  import RangePickerV2 from "./new-time-dropdown/RangePickerV2.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { page } from "$app/stores";
  import { ExploreStateURLParams } from "../../url-state/url-params";

  export let allTimeRange: TimeRange;
  export let selectedRangeAlias: string | undefined;
  export let showPivot: boolean;
  export let minTimeGrain: V1TimeGrain | undefined;
  export let defaultTimeRange: string | undefined;
  export let availableTimeZones: string[];
  export let timeRanges: V1ExploreTimeRange[];
  export let showDefaultItem: boolean;
  export let activeTimeGrain: V1TimeGrain | undefined;
  export let canPanLeft: boolean;
  export let canPanRight: boolean;
  export let interval: Interval;
  export let showPan = false;
  export let lockTimeZone = false;
  export let allowCustomTimeRange = true;
  export let showFullRange = true;
  export let complete: boolean;
  export let context = "dashboard";
  export let activeTimeZone: string;
  export let timeStart: string | undefined;
  export let timeEnd: string | undefined;
  export let watermark: DateTime | undefined = undefined;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;
  export let onPan: (direction: "left" | "right") => void;
  export let onTimeGrainSelect: (timeGrain: V1TimeGrain) => void;
  export let onSelectTimeZone: (timeZone: string) => void;
  export let applyRange: (range: TimeRange) => void;

  const newPicker = featureFlags.rillTime;

  $: rangeBuckets = bucketYamlRanges(timeRanges);

  $: ({
    url: { searchParams },
  } = $page);

  $: rawTimeString = searchParams.get(ExploreStateURLParams.TimeRange);

  $: v2TimeString = normalizeRangeString(rawTimeString);

  function normalizeRangeString(alias: string | null): string | undefined {
    return alias?.replace(",", " to ");
  }
</script>

<div class="wrapper">
  {#if showPan}
    <Elements.Nudge {canPanLeft} {canPanRight} {onPan} direction="left" />
    <Elements.Nudge {canPanLeft} {canPanRight} {onPan} direction="right" />
  {/if}

  {#if $newPicker}
    <RangePickerV2
      {context}
      smallestTimeGrain={minTimeGrain}
      minDate={DateTime.fromJSDate(allTimeRange.start)}
      maxDate={DateTime.fromJSDate(allTimeRange.end)}
      {watermark}
      {showDefaultItem}
      {defaultTimeRange}
      timeString={v2TimeString || selectedRangeAlias}
      {interval}
      zone={activeTimeZone}
      {lockTimeZone}
      {availableTimeZones}
      {onSelectTimeZone}
      {onSelectRange}
      {onTimeGrainSelect}
    />
  {:else if interval.isValid && activeTimeGrain}
    <Elements.RangePicker
      minDate={DateTime.fromJSDate(allTimeRange.start)}
      maxDate={DateTime.fromJSDate(allTimeRange.end)}
      ranges={rangeBuckets}
      {showDefaultItem}
      {showFullRange}
      {defaultTimeRange}
      {allowCustomTimeRange}
      selected={selectedRangeAlias ?? ""}
      {onSelectRange}
      {interval}
      zone={activeTimeZone}
      applyCustomRange={(interval) => {
        applyRange({
          name: TimeRangePreset.CUSTOM,
          start: interval.start.toJSDate(),
          end: interval.end.toJSDate(),
        });
      }}
      {side}
    />
  {/if}

  {#if availableTimeZones.length && !$newPicker}
    <Elements.Zone
      {context}
      watermark={interval.end ?? DateTime.fromJSDate(new Date())}
      {activeTimeZone}
      {lockTimeZone}
      {availableTimeZones}
      {onSelectTimeZone}
      {side}
    />
  {/if}

  {#if !$newPicker && !showPivot && minTimeGrain}
    <TimeGrainSelector
      {activeTimeGrain}
      {minTimeGrain}
      {timeStart}
      {timeEnd}
      {complete}
      {side}
      {onTimeGrainSelect}
    />
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex w-fit;
    @apply h-[26px] rounded-full;
    @apply overflow-hidden;
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

  :global(.wrapper > button:focus) {
    @apply z-50;
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
