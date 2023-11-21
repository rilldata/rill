import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { getAvailableComparisonsForTimeRange } from "@rilldata/web-common/lib/time/comparisons";
import type { TimeRangeMetaSet } from "@rilldata/web-common/lib/time/config";
import {
  LATEST_WINDOW_TIME_RANGES,
  PERIOD_TO_DATE_RANGES,
} from "@rilldata/web-common/lib/time/config";
import { getChildTimeRanges } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToTimeRangeMeta } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type { TimeRangeOption } from "@rilldata/web-common/lib/time/types";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import type { Readable } from "svelte/store";
import { derived } from "svelte/store";

export type TimeRangeState = {
  latestWindowTimeRanges: Array<TimeRangeOption>;
  periodToDateRanges: Array<TimeRangeOption>;
  showDefaultItem: boolean;
};

export function createTimeRangeStore(ctx: StateManagers) {
  return derived(
    [useMetaQuery(ctx), useTimeControlStore(ctx), ctx.dashboardStore],
    ([metricsView, timeControlsState, explorer]) => {
      if (
        !metricsView.data ||
        !timeControlsState.ready ||
        !timeControlsState.allTimeRange
      )
        return {
          latestWindowTimeRanges: [],
          periodToDateRanges: [],
          showDefaultItem: false,
        };

      let latestWindowTimeRanges: TimeRangeMetaSet = {};
      let periodToDateRanges: TimeRangeMetaSet = {};
      let hasDefaultInRanges = false;

      if (metricsView.data.availableTimeRanges?.length) {
        for (const availableTimeRange of metricsView.data.availableTimeRanges) {
          if (!availableTimeRange.range) continue;

          // default time range is part of availableTimeRanges.
          // this is used to not show a separate selection for the default
          if (metricsView.data.defaultTimeRange === availableTimeRange.range) {
            hasDefaultInRanges = true;
          }
          if (availableTimeRange.range in LATEST_WINDOW_TIME_RANGES) {
            latestWindowTimeRanges[availableTimeRange.range] =
              LATEST_WINDOW_TIME_RANGES[availableTimeRange.range];
          } else if (availableTimeRange.range in PERIOD_TO_DATE_RANGES) {
            periodToDateRanges[availableTimeRange.range] =
              PERIOD_TO_DATE_RANGES[availableTimeRange.range];
          } else {
            latestWindowTimeRanges[availableTimeRange.range] =
              isoDurationToTimeRangeMeta(availableTimeRange.range);
          }
        }
      } else {
        latestWindowTimeRanges = LATEST_WINDOW_TIME_RANGES;
        periodToDateRanges = PERIOD_TO_DATE_RANGES;
      }

      return {
        latestWindowTimeRanges: getChildTimeRanges(
          timeControlsState.allTimeRange.start,
          timeControlsState.allTimeRange.end,
          latestWindowTimeRanges,
          timeControlsState.minTimeGrain,
          explorer.selectedTimezone
        ),
        periodToDateRanges: getChildTimeRanges(
          timeControlsState.allTimeRange.start,
          timeControlsState.allTimeRange.end,
          periodToDateRanges,
          timeControlsState.minTimeGrain,
          explorer.selectedTimezone
        ),
        showDefaultItem:
          metricsView.data.defaultTimeRange && !hasDefaultInRanges,
      };
    }
  ) as Readable<TimeRangeState>;
}

export type TimeComparisonOptionsState = Array<TimeComparisonOption>;

export function createTimeComparisonOptionsState(ctx: StateManagers) {
  return derived(
    [useMetaQuery(ctx), useTimeControlStore(ctx)],
    ([metricsView, timeControlsState]) => {
      if (
        !metricsView.data ||
        !timeControlsState.ready ||
        !timeControlsState.allTimeRange ||
        !timeControlsState.selectedTimeRange
      )
        return [];

      let allOptions = [...Object.values(TimeComparisonOption)];
      if (metricsView.data.availableTimeRanges?.length) {
        const timeRange = metricsView.data.availableTimeRanges.find(
          (tr) => tr.range === timeControlsState.selectedTimeRange?.name
        );
        if (timeRange?.comparisonOffsets) {
          allOptions =
            timeRange.comparisonOffsets?.map(
              (co) => co.range as TimeComparisonOption
            ) ?? [];
        }
      }

      return getAvailableComparisonsForTimeRange(
        timeControlsState.allTimeRange.start,
        timeControlsState.allTimeRange.end,
        timeControlsState.selectedTimeRange.start,
        timeControlsState.selectedTimeRange.end,
        allOptions,
        [
          timeControlsState.selectedComparisonTimeRange
            ?.name as TimeComparisonOption,
        ]
      );
    }
  ) as Readable<TimeComparisonOptionsState>;
}
