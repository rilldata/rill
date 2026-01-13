import { getAvailableComparisonsForTimeRange } from "@rilldata/web-common/lib/time/comparisons";
import {
  TimeComparisonOption,
  TimeRangePreset,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";
import { getComparisonInterval } from "@rilldata/web-common/lib/time/comparisons";
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

  if (!interval.isValid) {
    return [];
  }

  const options: {
    name: TimeComparisonOption;
    key: number;
    start: Date;
    end: Date;
  }[] = [];

  timeComparisonOptions.forEach((co, i) => {
    const comparisonTimeRange = getComparisonInterval(
      interval as Interval<true>,
      co,
      timezone,
    );

    if (!comparisonTimeRange) return;
    options.push({
      name: co,
      key: i,
      start: comparisonTimeRange.start.toJSDate(),
      end: comparisonTimeRange.end.toJSDate(),
    });
  });

  return options;
}
