import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import { TimeRangeName } from "$common/database-service/DatabaseTimeSeriesActions";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";

// prepare the activeFilters to be sent to the server
export function prune(
  actives: ActiveValues,
  dimensions: Record<string, DimensionDefinitionEntity>
) {
  const filters: ActiveValues = {};
  for (const activeColumnId in actives) {
    if (!actives[activeColumnId].length) continue;
    filters[dimensions[activeColumnId].dimensionColumn] =
      actives[activeColumnId];
  }
  return filters;
}

const makeSelectableTimeRange = (
  name: TimeRangeName,
  datasetTimeRange: TimeSeriesTimeRange
): TimeSeriesTimeRange => {
  const start = new Date(datasetTimeRange?.start);
  const end = new Date(datasetTimeRange?.end);
  switch (name) {
    case TimeRangeName.LastHour:
      return {
        name,
        start: new Date(end.getTime() - 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last6Hours:
      return {
        name,
        start: new Date(end.getTime() - 6 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.LastDay:
      return {
        name,
        start: new Date(end.getTime() - 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last2Days:
      return {
        name,
        start: new Date(end.getTime() - 2 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last5Days:
      return {
        name,
        start: new Date(end.getTime() - 5 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.LastWeek:
      return {
        name,
        start: new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last2Weeks:
      return {
        name,
        start: new Date(end.getTime() - 14 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last30Days:
      return {
        name,
        start: new Date(end.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last60Days:
      return {
        name,
        start: new Date(end.getTime() - 60 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.AllTime:
      return {
        name,
        start: start.toISOString(),
        end: end.toISOString(),
      };
    default:
      throw new Error(`Unknown time range name: ${name}`);
  }
};

export const makeSelectableTimeRanges = (
  fullTimeRangeInDataset: TimeSeriesTimeRange
): TimeSeriesTimeRange[] => {
  return Object.keys(TimeRangeName).map((name) =>
    makeSelectableTimeRange(TimeRangeName[name], fullTimeRangeInDataset)
  );
};

export const getDefaultSelectedTimeRange = (
  selectableTimeRanges: TimeSeriesTimeRange[]
): TimeSeriesTimeRange => {
  return selectableTimeRanges.find(
    (timeRange) => timeRange.name === TimeRangeName.Last30Days
  );
};

export const getTimeRangeNameForButton = (
  timeRange: TimeSeriesTimeRange
): string => {
  if (timeRange && timeRange.name) return timeRange.name;
  if (timeRange && (timeRange.start || timeRange.end)) return "Custom";
  return "Select a time range";
};

export const prettyFormatTimeRange = (
  timeRange: TimeSeriesTimeRange
): string => {
  if (!timeRange?.start && timeRange?.end) {
    return `- ${timeRange.end}`;
  }

  if (timeRange?.start && !timeRange?.end) {
    return `${timeRange.start} -`;
  }

  if (!timeRange?.start && !timeRange?.end) {
    return "";
  }

  const start = new Date(timeRange.start);
  const end = new Date(timeRange.end);
  // day is the same
  if (
    start.getDate() === end.getDate() &&
    start.getMonth() === end.getMonth() &&
    start.getFullYear() === end.getFullYear()
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
    })} ${start.getDate()}, ${start.getFullYear()} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
      })
      .replace(/\s/g, "")})`;
  }
  // month is the same
  if (
    start.getMonth() === end.getMonth() &&
    start.getFullYear() === end.getFullYear()
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
    })} ${start.getDate()}-${end.getDate()}, ${start.getFullYear()} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
      })
      .replace(/\s/g, "")})`;
  }
  // year is the same
  if (start.getFullYear() === end.getFullYear()) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
    })} - ${end.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
    })}, ${start.getFullYear()}`;
  }
  // year is different
  const dateFormatOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "long",
    day: "numeric",
  };
  return `${start.toLocaleDateString(
    undefined,
    dateFormatOptions
  )} - ${end.toLocaleDateString(undefined, dateFormatOptions)}`;
};
