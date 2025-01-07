import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  TimeRangePreset,
  type DashboardTimeControls,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewTimeRange,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { derived, writable, type Readable, type Writable } from "svelte/store";

export class CanvasTimeControls {
  selectedTimeRange: Writable<DashboardTimeControls>;
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;
  allTimeRange: Readable<TimeRange>;

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

  combineAllTimeRange(ctx: StateManagers) {
    const timeRangeSummaryStore: Readable<TimeRange> = derived(
      [ctx.runtime, ctx.validSpecStore],
      ([r, validSpec], set) => {
        const metricsReferred = new Set<string>();
        if (validSpec?.data?.items?.length) {
          validSpec.data.items.forEach((component) => {
            // TODO: Spec should contain individual component spec
            const metricsView = component["metrics_view"] as string | undefined;
            if (metricsView) {
              metricsReferred.add(metricsView);
            }
          });
        } else {
          return set({
            start: new Date(0),
            end: new Date(),
          });
        }
        console.log(metricsReferred);
        if (metricsReferred.size === 0) {
          return set({
            start: new Date(0),
            end: new Date(),
          });
        }
        const timeRangeQueries = [...metricsReferred].map((metricView) => {
          return createQueryServiceMetricsViewTimeRange(
            r.instanceId,
            metricView,
            {},
            {
              query: {
                queryClient: ctx.queryClient,
                staleTime: Infinity,
                cacheTime: Infinity,
              },
            },
          );
        });

        return derived(timeRangeQueries, (timeRanges, querySet) => {
          let start = new Date(0);
          let end = new Date();
          timeRanges.forEach((timeRange) => {
            console.log(timeRange);
            const metricsStart = timeRange.data?.timeRangeSummary?.min;
            const metricsEnd = timeRange.data?.timeRangeSummary?.max;
            if (metricsStart) {
              const metricsStartDate = new Date(metricsStart);
              start = new Date(
                Math.min(start.getTime(), metricsStartDate.getTime()),
              );
            }
            if (metricsEnd) {
              const metricsEndDate = new Date(metricsEnd);
              end = new Date(Math.max(end.getTime(), metricsEndDate.getTime()));
            }
          });
          querySet({ start, end });
        }).subscribe(set);
      },
    );

    this.allTimeRange = timeRangeSummaryStore;
    return timeRangeSummaryStore;
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
