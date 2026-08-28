<script lang="ts">
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateSync";
  import { isUrlTooLong } from "@rilldata/web-common/features/dashboards/url-state/url-length-limits";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import type { TimeRange } from "@rilldata/web-common/lib/time/types";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type DashboardTimeControls,
  } from "@rilldata/web-common/lib/time/types";
  import {
    V1TimeGrain,
    type V1ExploreTimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { invalidationForMetricsViewData } from "@rilldata/web-common/runtime-client/invalidation.ts";
  import { DateTime, Duration, Interval } from "luxon";
  import { getStateManagers } from "../state-managers/state-managers";
  import {
    metricsExplorerStore,
    useExploreState,
  } from "../stores/dashboard-stores";
  import ComparisonPill from "../time-controls/comparison-pill/ComparisonPill.svelte";
  import {
    CUSTOM_TIME_RANGE_ALIAS,
    deriveInterval,
  } from "../time-controls/new-time-controls";
  import { allowedGrainsForInterval } from "@rilldata/web-common/lib/time/new-grains";
  import SuperPill from "../time-controls/super-pill/SuperPill.svelte";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import { featureFlags } from "../../feature-flags";
  import Timestamp from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/Timestamp.svelte";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Metadata from "../time-controls/super-pill/components/Metadata.svelte";
  import { getValidComparisonOption } from "../time-controls/time-range-store";
  import { getPinnedTimeZones } from "../url-state/getDefaultExplorePreset";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import ExpressionFilters from "./ExpressionFilters.svelte";
  import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
  import { untrack } from "svelte";
  import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

  const { rillTime } = featureFlags;

  let {
    timeRanges,
    metricsViewName,
    hasTimeSeries,
  }: {
    timeRanges: V1ExploreTimeRange[];
    metricsViewName: string;
    hasTimeSeries: boolean;
  } = $props();

  const StateManagers = getStateManagers();
  const {
    exploreName,
    validSpecStore,
    selectors: {
      dimensions: { timeDimensions },
      pivot: { showPivot },
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
    dashboardStore,
    expressionFilterManager,
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  const dashboardStateSync = DashboardStateSync.getFromContext();

  let showDefaultItem = false;

  const client = useRuntimeClient();

  let timeRangeQuery = $derived(
    useMetricsViewTimeRange(client, metricsViewName),
  );

  let timeRangeSummary = $derived($timeRangeQuery.data?.timeRangeSummary);

  let watermark = $derived(timeRangeSummary?.watermark);

  let maxQueryTimeRangeMillis = $derived(
    Number($timeRangeQuery.data?.maxQueryTimeRangeMillis ?? 0),
  );
  let maxQueryTimeRange = $derived(
    maxQueryTimeRangeMillis > 0
      ? Duration.fromMillis(maxQueryTimeRangeMillis)
      : undefined,
  );

  let {
    selectedTimeRange,
    allTimeRange,
    showTimeComparison,
    selectedComparisonTimeRange,
    minTimeGrain,
    timeStart,
    timeEnd,
    ready: timeControlsReady,
  } = $derived($timeControlsStore);

  let exploreSpec = $derived($validSpecStore.data?.explore ?? {});
  let metricsViewSpec = $derived($validSpecStore.data?.metricsView ?? {});

  let exploreState = $derived(useExploreState($exploreName));
  let activeTimeZone = $derived($exploreState?.selectedTimezone);

  let selectedRangeAlias = $derived(
    selectedTimeRange?.name === TimeRangePreset.CUSTOM
      ? `${selectedTimeRange.start.toISOString()},${selectedTimeRange.end.toISOString()}`
      : selectedTimeRange?.name,
  );

  let activeTimeGrain = $derived(selectedTimeRange?.interval);
  let defaultTimeRange = $derived(exploreSpec.defaultPreset?.timeRange);

  let { selectedTimeDimension } = $derived($dashboardStore);

  let availableTimeZones = $derived(getPinnedTimeZones(exploreSpec));

  let allTimeRangeInterval = $derived(
    allTimeRange
      ? Interval.fromDateTimes(allTimeRange.start, allTimeRange.end)
      : Interval.invalid("Invalid interval"),
  );

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

  let baseTimeRange = $derived(
    selectedTimeRange?.start &&
      selectedTimeRange?.end && {
        name: selectedTimeRange?.name,
        start: selectedTimeRange.start,
        end: selectedTimeRange.end,
      },
  );

  let primaryTimeDimension = $derived(metricsViewSpec.timeDimension);

  let timeDimensionOptions = $derived(
    $timeDimensions.map((timeDim) => {
      return {
        value: timeDim.name!,
        label: timeDim.displayName || timeDim.name!,
        description: timeDim.description,
      };
    }),
  );

  let maybeMinDate = $derived(
    allTimeRange?.start ? DateTime.fromJSDate(allTimeRange.start) : undefined,
  );
  let maybeMaxDate = $derived(
    allTimeRange?.end ? DateTime.fromJSDate(allTimeRange.end) : undefined,
  );

  let minDate = $derived(maybeMinDate?.isValid ? maybeMinDate : undefined);
  let maxDate = $derived(maybeMaxDate?.isValid ? maybeMaxDate : undefined);

  async function onTimeDimensionSelect(column: string) {
    // Capture time range name before any state changes
    const timeRangeName = selectedTimeRange?.name;

    await queryClient.cancelQueries({
      predicate: (query) =>
        invalidationForMetricsViewData(query, metricsViewName),
    });

    metricsExplorerStore.setTimeDimension($exploreName, column);

    // Re-resolve the time range with the new time dimension
    if (!timeRangeName) return;

    const { interval, grain } = await deriveInterval(
      timeRangeName,
      client,
      metricsViewName,
      activeTimeZone,
      column,
    );

    if (interval.isValid) {
      const validInterval = interval as Interval<true>;
      const baseTimeRange: TimeRange = {
        name: timeRangeName,
        start: validInterval.start.toJSDate(),
        end: validInterval.end.toJSDate(),
      };

      selectRange(baseTimeRange, grain);
    }
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

  async function onSelectRange(alias: string, tz = activeTimeZone) {
    // If we don't have a valid time range, early return
    if (!allTimeRange?.end) return;

    // This should be returned by the API, but it is not yet implemented
    const includesTimeZoneOffset = alias.includes("tz");

    if (includesTimeZoneOffset) {
      const timeZone = alias.match(/tz (.*)/)?.[1];

      if (timeZone) metricsExplorerStore.setTimeZone($exploreName, timeZone);
    }

    await queryClient.cancelQueries({
      predicate: (query) =>
        invalidationForMetricsViewData(query, metricsViewName),
    });

    const { interval, grain } = await deriveInterval(
      alias,
      client,
      metricsViewName,
      tz,
      selectedTimeDimension,
    );

    const allowedGrains = allowedGrainsForInterval(
      interval,
      minTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE,
    );

    const finalGrain =
      activeTimeGrain && allowedGrains.includes(activeTimeGrain)
        ? activeTimeGrain
        : grain && allowedGrains.includes(grain)
          ? grain
          : allowedGrains[0];

    if (interval.isValid) {
      const validInterval = interval as Interval<true>;
      const baseTimeRange: TimeRange = {
        name: alias,
        start: validInterval.start.toJSDate(),
        end: validInterval.end.toJSDate(),
      };

      selectRange(baseTimeRange, finalGrain);
    }
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

  function selectRange(range: TimeRange, grain?: V1TimeGrain) {
    const timeGrain =
      grain ?? getDefaultTimeGrain(range.start, range.end).grain;

    // Get valid option for the new time range
    const validComparison =
      allTimeRange &&
      getValidComparisonOption(
        exploreSpec.timeRanges,
        range,
        $exploreState.selectedComparisonTimeRange?.name as
          | TimeComparisonOption
          | undefined,
        allTimeRange,

        activeTimeZone,
      );

    makeTimeSeriesTimeRangeAndUpdateAppState(range, timeGrain, {
      name: validComparison,
    } as DashboardTimeControls);
  }

  async function onSelectTimeZone(timeZone: string) {
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
    } else if (selectedRangeAlias) {
      // Trigger range selection so that MetricsViewTimeRanges is called with the new time zone
      await onSelectRange(selectedRangeAlias, timeZone);
    }
    metricsExplorerStore.setTimeZone($exploreName, timeZone);
  }

  let usingRillTime = $derived(
    !selectedRangeAlias?.startsWith("P") &&
      !selectedRangeAlias?.startsWith("rill-"),
  );

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (usingRillTime && selectedRangeAlias) {
      metricsExplorerStore.setTimeGrain($exploreName, timeGrain);
    } else if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        $dashboardStore?.selectedComparisonTimeRange,
      );
    }
  }

  function isUrlTooLongAfterInListFilter(
    dimensionName: string,
    values: string[],
  ) {
    if (!dashboardStateSync) return false;

    // The chip calls this from a `$derived`, and the clone below mutates its own state while it
    // applies the filter, so the whole computation has to be untracked.
    return untrack(() => {
      const tempFilterManger = expressionFilterManager.clone();
      tempFilterManger.dimensionFilterAction(dimensionName, (m) =>
        m.setInList(values, m.exclude),
      );

      // Only the filter differs from the current state, and getUrlForExploreState only reads,
      // so a shallow copy is enough.
      const exploreState: ExploreState = {
        ...$dashboardStore,
        whereFilter:
          Object.values(tempFilterManger.topLevelJoiner.expr)[0] ??
          createAndExpression([]),
        dimensionsWithInlistFilter: tempFilterManger.inList,
      };

      const url = dashboardStateSync.getUrlForExploreState(exploreState);
      return isUrlTooLong(url);
    });
  }

  function syncExpressionFilters() {
    metricsExplorerStore.mergePartialExplorerEntity(
      $exploreName,
      {},
      expressionFilterManager,
    );
    return Promise.resolve();
  }
</script>

<div class="flex flex-col gap-y-2 size-full">
  {#if hasTimeSeries}
    <div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center">
      <Tooltip.Root delayDuration={0}>
        <Tooltip.Trigger class="cursor-default text-fg-secondary">
          <Calendar size="16px" />
        </Tooltip.Trigger>
        <Tooltip.Content side="bottom" sideOffset={10}>
          <Metadata
            timeZone={activeTimeZone}
            timeStart={allTimeRange?.start}
            timeEnd={allTimeRange?.end}
          />
        </Tooltip.Content>
      </Tooltip.Root>
      {#if allTimeRange?.start && allTimeRange?.end}
        <SuperPill
          {minDate}
          {maxDate}
          {selectedRangeAlias}
          showPivot={$showPivot}
          {minTimeGrain}
          {defaultTimeRange}
          {availableTimeZones}
          {timeRanges}
          complete={false}
          {interval}
          context={$exploreName}
          {timeStart}
          {timeEnd}
          lockTimeZone={exploreSpec.lockTimeZone}
          allowCustomTimeRange={exploreSpec.allowCustomTimeRange}
          {maxQueryTimeRange}
          {activeTimeGrain}
          {activeTimeZone}
          canPanLeft={$canPanLeft}
          canPanRight={$canPanRight}
          {primaryTimeDimension}
          {selectedTimeDimension}
          {showDefaultItem}
          timeDimensions={timeDimensionOptions}
          watermark={watermark ? DateTime.fromISO(watermark) : undefined}
          applyRange={selectRange}
          {onTimeDimensionSelect}
          {onSelectRange}
          {onTimeGrainSelect}
          {onSelectTimeZone}
          {onPan}
        />
        <ComparisonPill
          {minTimeGrain}
          {allTimeRange}
          {selectedTimeRange}
          showTimeComparison={!!showTimeComparison}
          {selectedComparisonTimeRange}
        />
      {/if}

      {#if !$rillTime && allTimeRangeInterval?.end?.isValid}
        <Tooltip.Root delayDuration={0}>
          <Tooltip.Trigger>
            <span class="text-fg-secondary italic">
              {m.dashboard_as_of()}
              <Timestamp
                id="filter-bar-as-of"
                italic
                suppress
                showDate={false}
                date={allTimeRangeInterval.end}
                zone={activeTimeZone}
              />
            </span>
          </Tooltip.Trigger>
          <Tooltip.Content side="bottom" sideOffset={10}>
            <Metadata
              timeZone={activeTimeZone}
              timeStart={allTimeRange?.start}
              timeEnd={allTimeRange?.end}
            />
          </Tooltip.Content>
        </Tooltip.Root>
      {/if}
    </div>
  {/if}

  <ExpressionFilters
    {expressionFilterManager}
    filteredDimensions={exploreSpec?.dimensions}
    filteredMeasures={exploreSpec?.measures}
    {timeStart}
    {timeEnd}
    timeDimension={$dashboardStore.selectedTimeDimension}
    {timeControlsReady}
    {isUrlTooLongAfterInListFilter}
    {syncExpressionFilters}
  />
</div>
