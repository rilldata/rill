import { getAvailableComparisonsForTimeRange } from "@rilldata/web-common/lib/time/comparisons";
import {
  TimeComparisonOption,
  TimeRangePreset,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";
import { getComparisonInterval } from "../stores/time-state";
import { DateTime, Interval } from "luxon";

export function getComparisonOptionsForCanvas(
  selectedTimeRange: DashboardTimeControls | undefined,
  allowCustomTimeRange: boolean,
  timezone: string,
) {
  if (!selectedTimeRange) {
    return [];
  }
  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(0),
    end: new Date(),
  };

  let allOptions = [...Object.values(TimeComparisonOption)];
  if (!allowCustomTimeRange) {
    allOptions = allOptions.filter((o) => o !== TimeComparisonOption.CUSTOM);
  }

  const timeComparisonOptions = getAvailableComparisonsForTimeRange(
    allTimeRange.start,
    allTimeRange.end,
    selectedTimeRange.start,
    selectedTimeRange.end,
    allOptions,
    timezone,
  );

  const interval = Interval.fromDateTimes(
    DateTime.fromJSDate(selectedTimeRange.start, { zone: timezone }),
    DateTime.fromJSDate(selectedTimeRange.end, { zone: timezone }),
  );

  return timeComparisonOptions.map((co, i) => {
    const comparisonTimeRange = getComparisonInterval(
      interval as Interval<true>,
      co,
      timezone,
    );
    return {
      name: co,
      key: i,
      start: comparisonTimeRange?.start.toJSDate(),
      end: comparisonTimeRange?.end.toJSDate(),
    };
  });
}
