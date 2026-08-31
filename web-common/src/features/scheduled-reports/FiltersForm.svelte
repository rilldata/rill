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
  import { syncStoreWithSource } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";

  const runtimeClient = useRuntimeClient();

  let {
    filters,
    metricsViewName,
    exploreName,
    timeControls,
    maxWidth = undefined,
    side = "bottom",
  }: {
    filters: ExpressionFilterManager;
    // TODO: make this support canvas. Unifying time controls should allow for this.
    metricsViewName: string;
    exploreName: string;
    timeControls: TimeControls;
    maxWidth?: number | undefined;
    side?: "top" | "right" | "bottom" | "left";
  } = $props();

  // svelte-ignore state_referenced_locally
  filters.createListener();
  // svelte-ignore state_referenced_locally
  syncStoreWithSource(
    filters,
    async (newUrlParams) => filters.setUrlParams(newUrlParams),
    () => filters.metricsViewsProvider.ready,
  );

  let {
    selectedTimezone,
    allTimeRange: allTimeRangeStore,
    timeRangeStateStore,
    comparisonRangeStateStore,
    minTimeGrain: _minTimeGrain,
    setTimeZone,
    selectTimeRange,
    setSelectedComparisonRange,
    displayTimeComparison,
  } = $derived(timeControls);

  let validSpecQuery = $derived(
    useExploreValidSpec(runtimeClient, exploreName),
  );
  let allTimeRange = $derived($allTimeRangeStore);
  let exploreSpec = $derived($validSpecQuery.data?.explore ?? {});
  let metricsViewSpec = $derived($validSpecQuery.data?.metricsView ?? {});
  let timeDimension = $derived(metricsViewSpec.timeDimension);

  let { selectedTimeRange, timeStart, timeEnd } = $derived(
    $timeRangeStateStore || {},
  );
  let selectedComparisonTimeRange = $derived(
    $comparisonRangeStateStore?.selectedComparisonTimeRange,
  );

  let baseTimeRange = $derived(
    selectedTimeRange?.start &&
      selectedTimeRange?.end && {
        name: selectedTimeRange?.name,
        start: selectedTimeRange.start,
        end: selectedTimeRange.end,
      },
  );

  let selectedRangeAlias = $derived(selectedTimeRange?.name);
  let activeTimeGrain = $derived(selectedTimeRange?.interval);
  let defaultTimeRange = $derived(exploreSpec.defaultPreset?.timeRange);
  let availableTimeZones = $derived(exploreSpec.timeZones ?? []);
  let timeRanges = $derived(exploreSpec.timeRanges ?? []);

  let minTimeGrain = $derived($_minTimeGrain);

  let activeTimeZone = $derived($selectedTimezone);

  let maybeInterval = $derived(
    selectedTimeRange
      ? Interval.fromDateTimes(
          DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
          DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
        )
      : allTimeRange
        ? Interval.fromDateTimes(allTimeRange.start, allTimeRange.end)
        : undefined,
  );

  let interval = $derived(maybeInterval?.isValid ? maybeInterval : undefined);

  let maybeMinDate = $derived(
    allTimeRange?.start ? DateTime.fromJSDate(allTimeRange.start) : undefined,
  );
  let maybeMaxDate = $derived(
    allTimeRange?.end ? DateTime.fromJSDate(allTimeRange.end) : undefined,
  );

  let minDate = $derived(maybeMinDate?.isValid ? maybeMinDate : undefined);
  let maxDate = $derived(maybeMaxDate?.isValid ? maybeMaxDate : undefined);

  let maybeComparisonInterval = $derived(
    selectedComparisonTimeRange
      ? Interval.fromDateTimes(
          DateTime.fromJSDate(selectedComparisonTimeRange.start).setZone(
            activeTimeZone,
          ),
          DateTime.fromJSDate(selectedComparisonTimeRange.end).setZone(
            activeTimeZone,
          ),
        )
      : undefined,
  );
  let comparisonInterval = $derived(
    maybeComparisonInterval?.isValid ? maybeComparisonInterval : undefined,
  );

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
