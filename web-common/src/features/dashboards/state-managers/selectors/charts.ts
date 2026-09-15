import {
  getDurationFromMS,
  getOffset,
  getTimeWidth,
} from "@rilldata/web-common/lib/time/transforms";
import {
  TimeOffsetType,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";

export const chartSelectors = {};

export function getPanRangeForTimeRange(
  timeRange: DashboardTimeControls | undefined,
  timeZone: string,
) {
  return (direction: "left" | "right") => {
    if (!timeRange) return;
    const { start, end } = timeRange;

    if (!start || !end) return;

    const offsetType =
      direction === "left" ? TimeOffsetType.SUBTRACT : TimeOffsetType.ADD;

    const currentRangeWidth = getTimeWidth(start, end);
    const panAmount = getDurationFromMS(currentRangeWidth);

    const newStart = getOffset(start, panAmount, offsetType, timeZone);
    const newEnd = getOffset(end, panAmount, offsetType, timeZone);

    return { start: newStart, end: newEnd };
  };
}
