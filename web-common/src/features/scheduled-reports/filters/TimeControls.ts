import { normalizeWeekday } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import {
  calculateComparisonTimeRangePartial,
  calculateTimeRangePartial,
  type ComparisonTimeRangeState,
  type TimeRangeState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { FiltersData } from "@rilldata/web-common/features/scheduled-reports/filters/FiltersData.ts";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges.ts";
import {
  type DashboardTimeControls,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { Settings } from "luxon";
import { derived, type Readable, writable, type Writable } from "svelte/store";

export class TimeControls {
  /**
   * Writables which can be updated by the user
   */
  selectedTimeRange: Writable<DashboardTimeControls | undefined>;
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;

  /**
   * Derived stores based on writables and spec
   */
  allTimeRange: Readable<TimeRange | undefined>;
  minTimeGrain: Readable<V1TimeGrain>;
  hasTimeSeries: Readable<boolean>;
  timeRangeStateStore: Readable<TimeRangeState | undefined>;
  comparisonRangeStateStore: Readable<ComparisonTimeRangeState | undefined>;

  public constructor(data: FiltersData) {
    this.selectedTimeRange = writable(undefined);
    this.selectedComparisonTimeRange = writable(undefined);
    this.showTimeComparison = writable(false);
    this.selectedTimezone = writable("UTC");

    this.allTimeRange = derived(
      data.timeRangeSummary,
      (timeRangeSummaryResp) => {
        if (!timeRangeSummaryResp.data?.timeRangeSummary) return undefined;
        return {
          name: TimeRangePreset.ALL_TIME,
          start: new Date(timeRangeSummaryResp.data.timeRangeSummary.min!),
          end: new Date(timeRangeSummaryResp.data.timeRangeSummary.max!),
        };
      },
    );

    this.hasTimeSeries = derived(data.validSpecQuery, (validSpecResp) => {
      const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
      return Boolean(metricsViewSpec.timeDimension);
    });

    this.timeRangeStateStore = derived(
      [
        data.validSpecQuery,
        this.allTimeRange,
        this.selectedTimeRange,
        this.selectedTimezone,
        this.minTimeGrain,
      ],
      ([
        validSpecResp,
        allTimeRange,
        selectedTimeRange,
        selectedTimezone,
        minTimeGrain,
      ]) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        const exploreSpec = validSpecResp.data?.explore ?? {};

        if (
          !metricsViewSpec ||
          !exploreSpec ||
          !selectedTimeRange ||
          !allTimeRange
        ) {
          return undefined;
        }

        Settings.defaultWeekSettings = {
          firstDay: normalizeWeekday(metricsViewSpec.firstDayOfWeek),
          weekend: [6, 7],
          minimalDays: 4,
        };

        const defaultTimeRange = isoDurationToFullTimeRange(
          exploreSpec.defaultPreset?.timeRange,
          allTimeRange.start,
          allTimeRange.end,
          selectedTimezone,
        );

        const timeRangeState = calculateTimeRangePartial(
          allTimeRange,
          selectedTimeRange,
          undefined, // scrub not present in canvas yet
          selectedTimezone,
          defaultTimeRange,
          minTimeGrain,
        );
        if (!timeRangeState) return undefined;
        return { ...timeRangeState };
      },
    );

    this.comparisonRangeStateStore = derived(
      [
        data.validSpecQuery,
        this.allTimeRange,
        this.selectedComparisonTimeRange,
        this.selectedTimezone,
        this.showTimeComparison,
        this.timeRangeStateStore,
      ],
      ([
        validSpecResp,
        allTimeRange,
        selectedComparisonTimeRange,
        selectedTimezone,
        showTimeComparison,
        timeRangeState,
      ]) => {
        const exploreSpec = validSpecResp.data?.explore ?? {};
        if (!exploreSpec || !timeRangeState || !allTimeRange) return undefined;
        const timeRanges = exploreSpec.timeRanges ?? [];
        return calculateComparisonTimeRangePartial(
          timeRanges,
          allTimeRange,
          selectedComparisonTimeRange,
          selectedTimezone,
          undefined, // scrub not present in canvas yet
          showTimeComparison,
          timeRangeState,
        );
      },
    );
  }

  setTimeZone = (timezone: string) => {
    this.selectedTimezone.set(timezone);
  };

  selectTimeRange = (
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) => {
    if (!timeRange.name) return;

    if (timeRange.name === TimeRangePreset.ALL_TIME) {
      this.showTimeComparison.set(false);
    }

    this.selectedTimeRange.set({
      ...timeRange,
      interval: timeGrain,
    });

    this.selectedComparisonTimeRange.set(comparisonTimeRange);
  };

  setSelectedComparisonRange = (comparisonTimeRange: DashboardTimeControls) => {
    this.selectedComparisonTimeRange.set(comparisonTimeRange);
  };

  displayTimeComparison = (showTimeComparison: boolean) => {
    this.showTimeComparison.set(showTimeComparison);
  };
}
