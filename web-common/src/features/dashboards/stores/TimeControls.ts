import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import { normalizeWeekday } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import {
  calculateComparisonTimeRangePartial,
  calculateTimeRangePartial,
  type ComparisonTimeRangeState,
  type TimeRangeState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { ExploreMetricsViewMetadata } from "@rilldata/web-common/features/dashboards/stores/ExploreMetricsViewMetadata.ts";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges.ts";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { Settings } from "luxon";
import {
  derived,
  get,
  type Readable,
  writable,
  type Writable,
} from "svelte/store";

export type TimeControlState = Pick<
  ExploreState,
  | "selectedTimeRange"
  | "selectedComparisonTimeRange"
  | "showTimeComparison"
  | "selectedTimezone"
>;

export class TimeControls {
  /**
   * Writables which can be updated by the user
   */
  public readonly selectedTimeRange: Writable<
    DashboardTimeControls | undefined
  >;
  public readonly selectedComparisonTimeRange: Writable<
    DashboardTimeControls | undefined
  >;
  public readonly showTimeComparison: Writable<boolean>;
  public readonly selectedTimezone: Writable<string>;

  /**
   * Derived stores based on writables and spec
   */
  public readonly allTimeRange: Readable<TimeRange | undefined>;
  public readonly minTimeGrain: Readable<V1TimeGrain>;
  public readonly hasTimeSeries: Readable<boolean>;
  public readonly timeRangeStateStore: Readable<TimeRangeState | undefined>;
  public readonly comparisonRangeStateStore: Readable<
    ComparisonTimeRangeState | undefined
  >;

  public constructor(
    metricsViewMetadata: ExploreMetricsViewMetadata,
    {
      selectedTimeRange,
      selectedComparisonTimeRange,
      showTimeComparison,
      selectedTimezone,
    }: TimeControlState,
  ) {
    this.selectedTimeRange = writable(selectedTimeRange);
    this.selectedComparisonTimeRange = writable(selectedComparisonTimeRange);
    this.showTimeComparison = writable(showTimeComparison);
    this.selectedTimezone = writable(selectedTimezone);

    this.allTimeRange = derived(
      metricsViewMetadata.timeRangeSummary,
      (timeRangeSummaryResp) => {
        if (!timeRangeSummaryResp.data?.timeRangeSummary) return undefined;
        return {
          name: TimeRangePreset.ALL_TIME,
          start: new Date(timeRangeSummaryResp.data.timeRangeSummary.min!),
          end: new Date(timeRangeSummaryResp.data.timeRangeSummary.max!),
        };
      },
    );

    this.hasTimeSeries = derived(
      metricsViewMetadata.validSpecQuery,
      (validSpecResp) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        return Boolean(metricsViewSpec.timeDimension);
      },
    );

    this.minTimeGrain = derived(
      metricsViewMetadata.validSpecQuery,
      (validSpecResp) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        return (
          metricsViewSpec.smallestTimeGrain ??
          V1TimeGrain.TIME_GRAIN_UNSPECIFIED
        );
      },
    );

    this.timeRangeStateStore = derived(
      [
        metricsViewMetadata.validSpecQuery,
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
        metricsViewMetadata.validSpecQuery,
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

  public setTimeZone = (timezone: string) => {
    this.selectedTimezone.set(timezone);
  };

  public selectTimeRange = (
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

    if (comparisonTimeRange !== undefined)
      this.selectedComparisonTimeRange.set(comparisonTimeRange);
  };

  public setSelectedComparisonRange = (
    comparisonTimeRange: DashboardTimeControls,
  ) => {
    this.selectedComparisonTimeRange.set(comparisonTimeRange);
  };

  public displayTimeComparison = (showTimeComparison: boolean) => {
    this.showTimeComparison.set(showTimeComparison);
    // TODO: find a better fix
    if (!get(this.selectedComparisonTimeRange)?.name) {
      this.selectedComparisonTimeRange.set({
        name: TimeComparisonOption.CONTIGUOUS,
      } as any);
    }
  };

  public toState(): TimeControlState {
    const timeRangeStateStore = get(this.timeRangeStateStore);
    const comparisonRangeStateStore = get(this.comparisonRangeStateStore);
    return {
      selectedTimeRange: timeRangeStateStore?.selectedTimeRange,
      selectedComparisonTimeRange:
        comparisonRangeStateStore?.selectedComparisonTimeRange,
      showTimeComparison:
        comparisonRangeStateStore?.showTimeComparison ?? false,
      selectedTimezone: get(this.selectedTimezone),
    };
  }

  public getStore(): Readable<TimeControlState> {
    return derived(
      [
        this.timeRangeStateStore,
        this.comparisonRangeStateStore,
        this.selectedTimezone,
      ],
      ([timeRangeStateStore, comparisonRangeStateStore, selectedTimezone]) => ({
        selectedTimeRange: timeRangeStateStore?.selectedTimeRange,
        selectedComparisonTimeRange:
          comparisonRangeStateStore?.selectedComparisonTimeRange,
        showTimeComparison:
          comparisonRangeStateStore?.showTimeComparison ?? false,
        selectedTimezone,
      }),
    );
  }
}
