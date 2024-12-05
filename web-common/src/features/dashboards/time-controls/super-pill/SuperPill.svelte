<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { getValidComparisonOption } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type DashboardTimeControls,
    type TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import { onMount } from "svelte";
  import {
    metricsExplorerStore,
    useExploreState,
  } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { initLocalUserPreferenceStore } from "../../user-preferences";
  import {
    ALL_TIME_RANGE_ALIAS,
    CUSTOM_TIME_RANGE_ALIAS,
    deriveInterval,
    type ISODurationString,
    type NamedRange,
    type RangeBuckets,
  } from "../new-time-controls";
  import * as Elements from "./components";

  export let allTimeRange: TimeRange;
  export let selectedTimeRange: DashboardTimeControls | undefined;

  const ctx = getStateManagers();
  const {
    exploreName,
    selectors: {
      timeRangeSelectors: { timeRangeSelectorState },
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
    validSpecStore,
  } = ctx;

  $: localUserPreferences = initLocalUserPreferenceStore($exploreName);

  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};
  $: exploreSpec = $validSpecStore.data?.explore ?? {};

  $: exploreState = useExploreState($exploreName);
  $: selectedRange =
    $exploreState?.selectedTimeRange?.name ?? ALL_TIME_RANGE_ALIAS;

  $: defaultTimeRange = exploreSpec?.defaultPreset?.timeRange;

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : Interval.fromDateTimes(allTimeRange.start, allTimeRange.end);

  $: activeTimeZone = $exploreState?.selectedTimezone;

  $: availableTimeZones = exploreSpec.timeZones ?? [];

  $: ({
    latestWindowTimeRanges,
    periodToDateRanges,
    previousCompleteDateRanges,
    showDefaultItem,
  } = $timeRangeSelectorState);

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

    metricsExplorerStore.setTimeZone($exploreName, timeZone);
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

  // This is pulled directly from the old time controls and needs to be refactored
  onMount(() => {
    /**
     * Remove the timezone selector if no timezone key is present
     * or the available timezone list is empty. Set the default
     * timezone to UTC in such cases.
     *
     */
    if (
      !availableTimeZones.length &&
      $exploreState?.selectedTimezone !== "UTC"
    ) {
      metricsExplorerStore.setTimeZone($exploreName, "UTC");
      localUserPreferences.set({ timeZone: "UTC" });
    }
  });

  function selectRange(range: TimeRange) {
    const defaultTimeGrain = getDefaultTimeGrain(range.start, range.end).grain;

    // Get valid option for the new time range
    const validComparison =
      allTimeRange &&
      getValidComparisonOption(
        exploreSpec,
        range,
        $exploreState.selectedComparisonTimeRange?.name as
          | TimeComparisonOption
          | undefined,
        allTimeRange,
      );

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
    metricsExplorerStore.selectTimeRange(
      $exploreName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
      metricsViewSpec,
    );
  }

  function onPan(direction: "left" | "right") {
    const panRange = $getNewPanRange(direction);
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
    metricsExplorerStore.selectTimeRange(
      $exploreName,
      timeRange as TimeRange,
      activeTimeGrain,
      comparisonTimeRange,
      metricsViewSpec,
    );
  }
</script>

<div class="wrapper">
  <Elements.Nudge
    canPanLeft={$canPanLeft}
    canPanRight={$canPanRight}
    {onPan}
    direction="left"
  />
  <Elements.Nudge
    canPanLeft={$canPanLeft}
    canPanRight={$canPanRight}
    {onPan}
    direction="right"
  />
  <!-- TO DO -->
  <!-- <Elements.Zoom /> -->
  {#if interval.isValid && activeTimeGrain}
    <Elements.RangePicker
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
