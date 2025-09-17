import {
  getAvailableComparisonsForTimeRange,
  getComparisonRange,
} from "@rilldata/web-common/lib/time/comparisons";
import { PREVIOUS_COMPLETE_DATE_RANGES } from "@rilldata/web-common/lib/time/config";
import {
  TimeComparisonOption,
  TimeRangePreset,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";
import type { Duration, Interval } from "luxon";
import { ALL_TIME_RANGE_ALIAS } from "../../dashboards/time-controls/new-time-controls";

const okay = [
  TimeComparisonOption.DAY,
  TimeComparisonOption.WEEK,
  TimeComparisonOption.MONTH,
  TimeComparisonOption.QUARTER,
  TimeComparisonOption.YEAR,
  TimeComparisonOption.CUSTOM,
];

function durationToIndex(duration: Duration) {
  if (duration.as("days") <= 1) return 0;
  if (duration.as("weeks") <= 1) return 1;
  if (duration.as("months") <= 1) return 2;
  if (duration.as("quarters") <= 1) return 3;
  return 4;
}

export function getComparisonOptionsForCanvas(
  interval: Interval<true> | undefined,
  range: string | undefined,
  // allowCustomTimeRange: boolean,
) {
  if (!interval || range == ALL_TIME_RANGE_ALIAS) {
    return [];
  }

  const options: TimeComparisonOption[] = [TimeComparisonOption.CONTIGUOUS];

  const duration = interval.toDuration();

  const index = durationToIndex(duration);

  const slicedOptions = okay.slice(index, -1);

  options.push(...slicedOptions);

  return options;
}
