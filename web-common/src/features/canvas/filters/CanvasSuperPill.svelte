<script lang="ts">
  import { getTimeRangeForCanvas } from "@rilldata/web-common/features/canvas/filters/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { getPanRangeForTimeRange } from "@rilldata/web-common/features/dashboards/state-managers/selectors/charts";
  import {
    ALL_TIME_RANGE_ALIAS,
    CUSTOM_TIME_RANGE_ALIAS,
    deriveInterval,
    type ISODurationString,
    type NamedRange,
    type RangeBuckets,
  } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import * as Elements from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type DashboardTimeControls,
    type TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";

  export let allTimeRange: TimeRange;
  export let selectedTimeRange: DashboardTimeControls | undefined;
  export let activeTimeZone: string;

  const ctx = getCanvasStateManagers();
  const { canvasName, canvasStore, validSpecStore } = ctx;

  $: localUserPreferences = initLocalUserPreferenceStore($canvasName);

  // $: canvasSpec = $validSpecStore.data;

  $: selectedRange = selectedTimeRange?.name ?? ALL_TIME_RANGE_ALIAS;

  // TODO: Add default timeRange to resource
  // $: defaultTimeRange = $validSpecStore?.data?.defaultPreset?.timeRange;
  let defaultTimeRange = "PT24H";

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : Interval.fromDateTimes(allTimeRange.start, allTimeRange.end);

  // TODO: Add timezone key to resource
  // $: availableTimeZones = canvasSpec?.timeZones ?? [];
  let availableTimeZones = [
    "America/Los_Angeles",
    "America/New_York",
    "Europe/London",
    "Asia/Kolkata",
  ];

  $: ({
    latestWindowTimeRanges,
    periodToDateRanges,
    previousCompleteDateRanges,
    showDefaultItem,
  } = getTimeRangeForCanvas(activeTimeZone));

  $: ranges = <RangeBuckets>{
    latest: latestWindowTimeRanges.map((range) => ({
      range: range.name,
      label: range.label,
    })),
    periodToDate: periodToDateRanges.map((range) => ({
      range: range.name,
      label: range.label,
    })),
    previous: previousCompleteDateRanges.map((range) => ({
      range: range.name,
      label: range.label,
    })),
  };

  $: activeTimeGrain = selectedTimeRange?.interval;

  function onSelectTimeZone(timeZone: string) {
    if (!interval.isValid) return;

    if (selectedRange === "CUSTOM") {
      selectRange({
        name: TimeRangePreset.CUSTOM,
        start: interval.start
          ?.setZone(timeZone, { keepLocalTime: true })
          .toJSDate(),
        end: interval.end
          ?.setZone(timeZone, { keepLocalTime: true })
          .toJSDate(),
      });
    }

    $canvasStore.timeControls.setTimeZone(timeZone);
    localUserPreferences.set({ timeZone });
  }

  function onSelectRange(name: NamedRange | ISODurationString) {
    if (!allTimeRange?.end) {
      return;
    }

    if (name === ALL_TIME_RANGE_ALIAS) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        allTimeRange,
        "TIME_GRAIN_DAY",
        undefined,
      );
      return;
    }

    const interval = deriveInterval(
      name,
      DateTime.fromJSDate(allTimeRange.end),
    );
    if (!interval?.isValid) return;

    const baseTimeRange: TimeRange = {
      name: name as TimeRangePreset,
      start: interval.start.toJSDate(),
      end: interval.end.toJSDate(),
    };

    selectRange(baseTimeRange);
  }

  function selectRange(range: TimeRange) {
    const defaultTimeGrain = getDefaultTimeGrain(range.start, range.end).grain;

    const comparisonOption = DEFAULT_TIME_RANGES[range.name as TimeRangePreset]
      ?.defaultComparison as TimeComparisonOption;

    // Get valid option for the new time range
    const validComparison = allTimeRange && comparisonOption;

    makeTimeSeriesTimeRangeAndUpdateAppState(range, defaultTimeGrain, {
      name: validComparison,
    } as DashboardTimeControls);
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    $canvasStore.timeControls.selectTimeRange(
      timeRange,
      timeGrain,
      comparisonTimeRange,
    );
  }

  function onPan(direction: "left" | "right") {
    const getPanRange = getPanRangeForTimeRange(
      selectedTimeRange,
      activeTimeZone,
    );
    const panRange = getPanRange(direction);
    if (!panRange) return;
    const { start, end } = panRange;

    const timeRange = {
      name: CUSTOM_TIME_RANGE_ALIAS,
      start: start,
      end: end,
    };

    const comparisonTimeRange = {
      name: TimeComparisonOption.CONTIGUOUS,
    } as DashboardTimeControls; // FIXME wrong typecasting across application

    if (!activeTimeGrain) return;
    $canvasStore.timeControls.selectTimeRange(
      timeRange as TimeRange,
      activeTimeGrain,
      comparisonTimeRange,
    );
  }
</script>

<div class="wrapper">
  <Elements.Nudge canPanLeft canPanRight {onPan} direction="left" />
  <Elements.Nudge canPanLeft canPanRight {onPan} direction="right" />
  {#if interval.isValid && activeTimeGrain}
    <Elements.RangePicker
      minDate={DateTime.fromJSDate(allTimeRange.start)}
      maxDate={DateTime.fromJSDate(allTimeRange.end)}
      {ranges}
      {showDefaultItem}
      {defaultTimeRange}
      selected={selectedRange}
      grain={activeTimeGrain}
      {onSelectRange}
      {interval}
      zone={activeTimeZone}
      applyCustomRange={(interval) => {
        selectRange({
          name: TimeRangePreset.CUSTOM,
          start: interval.start.toJSDate(),
          end: interval.end.toJSDate(),
        });
      }}
    />
  {/if}

  {#if availableTimeZones.length}
    <Elements.Zone
      watermark={allTimeRange?.end ?? new Date()}
      {activeTimeZone}
      {availableTimeZones}
      {onSelectTimeZone}
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
    @apply px-2 flex items-center justify-center bg-white;
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

  :global(.wrapper > [data-state="open"]) {
    @apply bg-gray-50 border-gray-400 z-50;
  }
</style>
