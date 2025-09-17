import { timeControlsState } from "@rilldata/web-common/features/dashboards/state-managers/selectors/time-range";
import {
  getDurationFromMS,
  getOffset,
  getTimeWidth,
} from "@rilldata/web-common/lib/time/transforms";
import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
import type { DashboardDataSources } from "./types";
import { Interval } from "luxon";

export const chartSelectors = {
  canPanLeft: (dashData: DashboardDataSources) => {
    const timeControls = timeControlsState(dashData);
    const startRange = timeControls.allTimeRange?.start;
    const selectedStart = timeControls.selectedTimeRange?.start;
    return (
      (selectedStart?.getTime() || Infinity) >=
      (startRange?.getTime() || -Infinity)
    );
  },
  canPanRight: (dashData: DashboardDataSources) => {
    const timeControls = timeControlsState(dashData);
    const endRange = timeControls?.allTimeRange?.end;
    const selectedEnd = timeControls.selectedTimeRange?.end;
    return (
      (selectedEnd?.getTime() || -Infinity) <= (endRange?.getTime() || Infinity)
    );
  },
  getNewPanRange: (dashData: DashboardDataSources) => {
    const timeControls = timeControlsState(dashData);
    const timeZone = dashData.dashboard?.selectedTimezone;
    const interval =
      timeControls.selectedTimeRange &&
      Interval.fromDateTimes(
        timeControls.selectedTimeRange?.start,
        timeControls.selectedTimeRange?.end,
      );
    if (!interval?.isValid || !timeZone) return;

    return getPanRangeForTimeRange(interval, timeZone);
  },
};

export function getPanRangeForTimeRange(
  timeRange: Interval<true> | undefined,
  timeZone: string,
) {
  return (direction: "left" | "right") => {
    if (!timeRange) return;
    const { start, end } = timeRange;

    if (!start || !end) return;

    const offsetType =
      direction === "left" ? TimeOffsetType.SUBTRACT : TimeOffsetType.ADD;

    const currentRangeWidth = getTimeWidth(start.toJSDate(), end.toJSDate());
    const panAmount = getDurationFromMS(currentRangeWidth);

    const newStart = getOffset(
      start.toJSDate(),
      panAmount,
      offsetType,
      timeZone,
    );
    const newEnd = getOffset(end.toJSDate(), panAmount, offsetType, timeZone);

    return { start: newStart, end: newEnd };
  };
}
