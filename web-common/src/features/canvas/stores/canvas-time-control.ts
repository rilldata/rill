import type { CanvasValidResponse } from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { getTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  TimeRangePreset,
  type DashboardTimeControls,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewTimeRange,
  V1TimeGrain,
  type RpcStatus,
} from "@rilldata/web-common/runtime-client";
import {
  runtime,
  type Runtime,
} from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";

export class CanvasTimeControls {
  /**
   * Writables
   */
  selectedTimeRange: Writable<DashboardTimeControls>;
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;
  allTimeRange: Readable<TimeRange>;
  isReady: Writable<boolean>;

  constructor(validSpecStore: CanvasSpecResponseStore) {
    // TODO: Refactor this
    this.allTimeRange = writable({
      name: TimeRangePreset.ALL_TIME,
      start: new Date(0),
      end: new Date(),
    });
    this.selectedTimeRange = writable({
      name: TimeRangePreset.ALL_TIME,
      start: new Date(0),
      end: new Date(),
      interval: V1TimeGrain.TIME_GRAIN_DAY,
    });
    this.selectedComparisonTimeRange = writable(undefined);
    this.showTimeComparison = writable(false);
    this.selectedTimezone = writable("UTC");

    this.isReady = writable(true);

    this.setInitialState(validSpecStore);
  }

  setInitialState(validSpecStore: CanvasSpecResponseStore) {
    this.timeRangeSummaryStore(runtime, validSpecStore);
    const store = derived(
      [this.allTimeRange, validSpecStore],
      ([allTimeRange, validSpec]) => {
        if (!validSpec.data) {
          this.isReady.set(false);
        }

        const selectedTimezone = get(this.selectedTimezone);
        const defaultTimeRange = isoDurationToFullTimeRange(
          validSpec.data?.canvas?.defaultPreset?.timeRange,
          allTimeRange.start,
          allTimeRange.end,
          selectedTimezone,
        );

        const newTimeRange: DashboardTimeControls = {
          name: defaultTimeRange.name,
          start: defaultTimeRange.start,
          end: defaultTimeRange.end,
        };

        newTimeRange.interval = getTimeGrain(
          undefined,
          newTimeRange,
          V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        );

        this.selectedTimeRange.set(newTimeRange);
        this.isReady.set(true);
      },
    );

    // Subscribe to ensure the derived code runs
    store.subscribe(() => {});
  }

  timeRangeSummaryStore = (
    runtime: Writable<Runtime>,
    validSpecStore: Readable<
      QueryObserverResult<CanvasValidResponse | undefined, RpcStatus>
    >,
  ) => {
    this.allTimeRange = derived(
      [runtime, validSpecStore],
      ([r, validSpec], set) => {
        const metricsReferred = Object.keys(
          validSpec?.data?.metricsViews || {},
        );
        if (!metricsReferred.length) {
          return set({
            name: TimeRangePreset.ALL_TIME,
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
                queryClient: queryClient,
                staleTime: Infinity,
                cacheTime: Infinity,
              },
            },
          );
        });

        return derived(timeRangeQueries, (timeRanges, querySet) => {
          let start = new Date();
          let end = new Date(0);
          timeRanges.forEach((timeRange) => {
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
          querySet({ name: TimeRangePreset.ALL_TIME, start, end });
        }).subscribe(set);
      },
    );
  };

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
