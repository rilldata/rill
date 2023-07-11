import {
  MetricsExplorerEntity,
  metricsExplorerStore,
  useDashboardStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores";
import {
  getComparisonRange,
  getTimeComparisonParametersForComponent,
} from "@rilldata/web-common/lib/time/comparisons";
import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
import {
  checkValidTimeGrain,
  findValidTimeGrain,
  getAllowedTimeGrains,
  getDefaultTimeGrain,
} from "@rilldata/web-common/lib/time/grains";
import {
  convertTimeRangePreset,
  ISODurationToTimePreset,
} from "@rilldata/web-common/lib/time/ranges";
import {
  TimeComparisonOption,
  TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import type { TimeRangeType } from "@rilldata/web-common/lib/time/types";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceColumnTimeRange,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { V1MetricsView } from "@rilldata/web-common/runtime-client";
import { derived, readable } from "svelte/store";
import type { Readable } from "svelte/store";

export type TimeControlState = {
  isFetching: boolean;

  // Computed properties from all time range query
  defaultTimeRange?: TimeRangeType;
  minTimeGrain?: V1TimeGrain;
  allTimeRange?: TimeRange;

  hasTime?: boolean;
  // Selected ranges with start and end filled based on time range type
  selectedTimeRange?: DashboardTimeControls;
  timeStart?: string;
  timeEnd?: string;

  showComparison?: boolean;
  selectedComparisonTimeRange?: DashboardTimeControls;
  comparisonTimeStart?: string;
  comparisonTimeEnd?: string;
};
type TimeControlReducers = {};
export type TimeControlStore = Readable<TimeControlState> & TimeControlReducers;

export function createTimeControlStore(
  instanceId: string,
  metricsViewName: string,
  metricsView: V1MetricsView
) {
  const hasTimeSeries = metricsView ? !!metricsView.timeDimension : false;

  if (!metricsView?.model) {
    return readable({
      hasTime: !hasTimeSeries,
    }) as TimeControlStore;
  }

  return derived(
    [
      createQueryServiceColumnTimeRange(
        instanceId,
        metricsView.model,
        {
          columnName: metricsView.timeDimension,
        },
        {
          query: {
            enabled: !!metricsView.timeDimension,
          },
        }
      ),
      useDashboardStore(metricsViewName),
    ],
    ([timeRangeResponse, metricsExplorer]) => {
      if (!timeRangeResponse || !timeRangeResponse.isSuccess) {
        return {
          isFetching: timeRangeResponse.isRefetching,
          hasTime: !hasTimeSeries,
        } as TimeControlState;
      }
      if (!metricsExplorer.defaultsSelected) {
        metricsExplorerStore.allDefaultsSelected(metricsViewName);
      }
      const allTimeRange = {
        name: TimeRangePreset.ALL_TIME,
        start: new Date(timeRangeResponse.data.timeRangeSummary.min),
        end: new Date(timeRangeResponse.data.timeRangeSummary.max),
      };

      const defaultTimeRange = ISODurationToTimePreset(
        metricsView.defaultTimeRange
      );
      const minTimeGrain =
        (metricsView.defaultTimeRange as V1TimeGrain) ||
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

      const timeRange = getTimeRange(
        metricsExplorer,
        allTimeRange,
        defaultTimeRange
      );
      timeRange.interval = getTimeGrain(
        metricsExplorer,
        timeRange,
        minTimeGrain
      );

      const selectedComparisonTimeRange = getComparisonTimeRange(
        allTimeRange,
        timeRange,
        metricsExplorer.selectedComparisonTimeRange
      );

      return {
        isFetching: false,
        defaultTimeRange,
        minTimeGrain,
        allTimeRange,
        hasTime: true,

        selectedTimeRange: timeRange,
        timeStart: timeRange.start.toISOString(),
        timeEnd: timeRange.end.toISOString(),

        showComparison: Boolean(
          metricsExplorer.showComparison && selectedComparisonTimeRange?.start
        ),
        selectedComparisonTimeRange,
        comparisonTimeStart: selectedComparisonTimeRange?.start.toISOString(),
        comparisonTimeEnd: selectedComparisonTimeRange?.end.toISOString(),
      } as TimeControlState;
    }
  ) as TimeControlStore;
}

function getTimeRange(
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  defaultTimeRange: string
) {
  let timeRange: DashboardTimeControls;
  if (!metricsExplorer?.selectedTimeRange) {
    timeRange = convertTimeRangePreset(
      defaultTimeRange,
      allTimeRange.start,
      allTimeRange.end
    );
  } else {
    if (metricsExplorer.selectedTimeRange.name === TimeRangePreset.CUSTOM) {
      /** set the time range to the fixed custom time range */
      timeRange = {
        name: TimeRangePreset.CUSTOM,
        start: new Date(metricsExplorer.selectedTimeRange.start),
        end: new Date(metricsExplorer.selectedTimeRange.end),
      };
    } else {
      /** rebuild off of relative time range */
      timeRange = convertTimeRangePreset(
        metricsExplorer.selectedTimeRange?.name ?? TimeRangePreset.ALL_TIME,
        allTimeRange.start,
        allTimeRange.end
      );
    }
  }
  return timeRange;
}

function getTimeGrain(
  metricsExplorer: MetricsExplorerEntity,
  timeRange: DashboardTimeControls,
  minTimeGrain: V1TimeGrain
) {
  let timeGrain: V1TimeGrain;

  if (!metricsExplorer?.selectedTimeRange) {
    timeGrain = getDefaultTimeGrain(timeRange.start, timeRange.end).grain;
  } else {
    const timeGrainOptions = getAllowedTimeGrains(
      timeRange.start,
      timeRange.end
    );
    const isValidTimeGrain = checkValidTimeGrain(
      metricsExplorer.selectedTimeRange.interval,
      timeGrainOptions,
      minTimeGrain
    );

    if (isValidTimeGrain) {
      timeGrain = metricsExplorer.selectedTimeRange.interval;
    } else {
      const defaultTimeGrain = getDefaultTimeGrain(
        timeRange.start,
        timeRange.end
      ).grain;
      timeGrain = findValidTimeGrain(
        defaultTimeGrain,
        timeGrainOptions,
        minTimeGrain
      );
    }
  }

  return timeGrain;
}

function getComparisonTimeRange(
  allTimeRange: DashboardTimeControls,
  timeRange: DashboardTimeControls,
  comparisonTimeRange: DashboardTimeControls
) {
  if (!comparisonTimeRange) return undefined;

  let selectedComparisonTimeRange: DashboardTimeControls;
  if (!comparisonTimeRange?.name) {
    const comparisonOption = DEFAULT_TIME_RANGES[timeRange.name]
      ?.defaultComparison as TimeComparisonOption;
    const range = getTimeComparisonParametersForComponent(
      comparisonOption,
      allTimeRange.start,
      allTimeRange.end,
      timeRange.start,
      timeRange.end
    );

    if (range.isComparisonRangeAvailable) {
      selectedComparisonTimeRange = {
        start: range.start,
        end: range.end,
        name: comparisonOption,
      };
    }
  } else if (comparisonTimeRange.name === TimeComparisonOption.CUSTOM) {
    selectedComparisonTimeRange = comparisonTimeRange;
  } else {
    // variable time range of some kind.
    const comparisonOption = comparisonTimeRange.name as TimeComparisonOption;
    const range = getComparisonRange(
      timeRange.start,
      timeRange.end,
      comparisonOption
    );

    selectedComparisonTimeRange = {
      ...range,
      name: comparisonOption,
    };
  }

  return selectedComparisonTimeRange;
}
