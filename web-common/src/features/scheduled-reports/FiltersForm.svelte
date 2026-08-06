<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CanvasComparisonPill from "@rilldata/web-common/features/canvas/filters/CanvasComparisonPill.svelte";
  import { deriveInterval } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config.ts";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import {
    type DashboardTimeControls,
    TimeComparisonOption,
    type TimeRange,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types.ts";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { invalidationForMetricsViewData } from "@rilldata/web-common/runtime-client/invalidation.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { DateTime, Interval } from "luxon";
  import { onMount } from "svelte";
  import type { ExpressionFilterManager } from "../dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import ExpressionFilters from "../dashboards/filters/ExpressionFilters.svelte";

  const runtimeClient = useRuntimeClient();

  export let filters: ExpressionFilterManager;
  // TODO: make this support canvas
  export let metricsViewName: string;
  export let exploreName: string;
  export let timeControls: TimeControls;
  export let maxWidth: number | undefined = undefined;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";

  $: ({
    selectedTimezone,
    allTimeRange: allTimeRangeStore,
    timeRangeStateStore,
    comparisonRangeStateStore,
    minTimeGrain: _minTimeGrain,
    setTimeZone,
    selectTimeRange,
    setSelectedComparisonRange,
    displayTimeComparison,
  } = timeControls);

  const validSpecQuery = useExploreValidSpec(runtimeClient, exploreName);
  $: allTimeRange = $allTimeRangeStore;
  $: exploreSpec = $validSpecQuery.data?.explore ?? {};
  $: metricsViewSpec = $validSpecQuery.data?.metricsView ?? {};
  $: timeDimension = metricsViewSpec.timeDimension;

  $: ({ selectedTimeRange, timeStart, timeEnd } = $timeRangeStateStore || {});
  $: selectedComparisonTimeRange =
    $comparisonRangeStateStore?.selectedComparisonTimeRange;

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = exploreSpec.defaultPreset?.timeRange;
  $: availableTimeZones = exploreSpec.timeZones ?? [];
  $: timeRanges = exploreSpec.timeRanges ?? [];

  $: minTimeGrain = $_minTimeGrain;

  $: activeTimeZone = $selectedTimezone;

  $: maybeInterval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : allTimeRange
      ? Interval.fromDateTimes(allTimeRange.start, allTimeRange.end)
      : undefined;

  $: interval = maybeInterval?.isValid ? maybeInterval : undefined;

  $: maybeMinDate = allTimeRange?.start
    ? DateTime.fromJSDate(allTimeRange.start)
    : undefined;
  $: maybeMaxDate = allTimeRange?.end
    ? DateTime.fromJSDate(allTimeRange.end)
    : undefined;

  $: minDate = maybeMinDate?.isValid ? maybeMinDate : undefined;
  $: maxDate = maybeMaxDate?.isValid ? maybeMaxDate : undefined;

  $: maybeComparisonInterval = selectedComparisonTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedComparisonTimeRange.start).setZone(
          activeTimeZone,
        ),
        DateTime.fromJSDate(selectedComparisonTimeRange.end).setZone(
          activeTimeZone,
        ),
      )
    : undefined;
  $: comparisonInterval = maybeComparisonInterval?.isValid
    ? maybeComparisonInterval
    : undefined;

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    selectTimeRange(timeRange, timeGrain, comparisonTimeRange);
  }

  function selectRange(
    range: TimeRange,
    grain?: V1TimeGrain,
    rangeOnly: boolean = false,
  ) {
    const defaultTimeGrain =
      grain ?? getDefaultTimeGrain(range.start, range.end).grain;

    const comparisonOption = DEFAULT_TIME_RANGES[range.name as TimeRangePreset]
      ?.defaultComparison as TimeComparisonOption;

    // Get valid option for the new time range
    const validComparison = allTimeRange && comparisonOption;

    makeTimeSeriesTimeRangeAndUpdateAppState(
      range,
      defaultTimeGrain,
      rangeOnly
        ? undefined
        : ({
            name: validComparison,
          } as DashboardTimeControls),
    );
  }

  async function onSelectRange(name: string, rangeOnly: boolean = false) {
    if (!allTimeRange?.end) {
      return;
    }

    const includesTimeZoneOffset = name.includes("tz");

    if (includesTimeZoneOffset) {
      const timeZone = name.match(/tz (.*)/)?.[1];

      if (timeZone) setTimeZone(timeZone);
    }

    await queryClient.cancelQueries({
      predicate: (query) =>
        invalidationForMetricsViewData(query, metricsViewName),
    });

    const { interval, grain } = await deriveInterval(
      name,
      runtimeClient,
      metricsViewName,
      $selectedTimezone,
      timeDimension,
    );

    if (interval?.isValid) {
      const validInterval = interval as Interval<true>;
      const baseTimeRange: TimeRange = {
        name: name as TimeRangePreset,
        start: validInterval.start.toJSDate(),
        end: validInterval.end.toJSDate(),
      };

      selectRange(baseTimeRange, grain, rangeOnly);
    }
  }

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        selectedComparisonTimeRange,
      );
    }
  }

  function onSelectTimeZone(timeZone: string) {
    if (!interval?.isValid) return;

    if (selectedRangeAlias === TimeRangePreset.CUSTOM) {
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

    setTimeZone(timeZone);
  }

  onMount(() => {
    if (selectedRangeAlias) onSelectRange(selectedRangeAlias, true);
  });
</script>

<div
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
  aria-label={m.report_form_filters_aria()}
>
  <div
    class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center ml-2 pointer-events-auto w-fit"
  >
    <Calendar size="16px" />
    {#if allTimeRange}
      <SuperPill
        {minDate}
        {maxDate}
        {selectedRangeAlias}
        showPivot={false}
        {defaultTimeRange}
        {availableTimeZones}
        {timeRanges}
        complete={false}
        {interval}
        {timeStart}
        {timeEnd}
        {activeTimeGrain}
        activeTimeZone={$selectedTimezone}
        allowCustomTimeRange={false}
        showDefaultItem
        applyRange={selectRange}
        {onSelectRange}
        {onTimeGrainSelect}
        {onSelectTimeZone}
        hidePan
        onPan={() => {}}
        {minTimeGrain}
        {side}
      />
      <CanvasComparisonPill
        {minTimeGrain}
        {minDate}
        {maxDate}
        {interval}
        selectedRange={selectedRangeAlias}
        comparisonRange={selectedComparisonTimeRange?.name}
        {comparisonInterval}
        {activeTimeGrain}
        showTimeComparison={$comparisonRangeStateStore?.showTimeComparison ??
          false}
        activeTimeZone={$selectedTimezone}
        onDisplayTimeComparison={displayTimeComparison}
        onSetSelectedComparisonRange={setSelectedComparisonRange}
        allowCustomTimeRange={false}
        {side}
      />
    {/if}
  </div>

  <ExpressionFilters
    expressionFilterManager={filters}
    filteredMeasures={exploreSpec?.measures}
    {timeStart}
    {timeEnd}
    {timeDimension}
    timeControlsReady
  />
</div>
