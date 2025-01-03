import {
  TimeRangePreset,
  type DashboardTimeControls,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { writable, type Writable } from "svelte/store";

export class CanvasTimeControls {
  selectedTimeRange: Writable<DashboardTimeControls>;
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;

  constructor() {
    this.selectedTimeRange = writable({
      name: TimeRangePreset.ALL_TIME,
      start: new Date(0),
      end: new Date(),
      interval: V1TimeGrain.TIME_GRAIN_DAY,
    });
    this.selectedComparisonTimeRange = writable(undefined);
    this.showTimeComparison = writable(false);
    this.selectedTimezone = writable("UTC");
  }

  setTimeZone(timezone: string) {
    this.selectedTimezone.set(timezone);
  }

  selectTimeRange(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    if (!timeRange.name) return;

    if (timeRange.name === TimeRangePreset.ALL_TIME) {
      this.showTimeComparison.set(false);
    }

    this.selectedTimeRange.set({
      ...timeRange,
      interval: timeGrain,
    });

    this.selectedComparisonTimeRange.set(comparisonTimeRange);
  }

  setSelectedComparisonRange(comparisonTimeRange: DashboardTimeControls) {
    this.selectedComparisonTimeRange.set(comparisonTimeRange);
  }

  displayTimeComparison(showTimeComparison: boolean) {
    this.showTimeComparison.set(showTimeComparison);
  }
}
