import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import { timeRanges } from "$lib/util/time-ranges";

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

export const defaultTimeRange = timeRanges[0];

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
    })} ${start.getDate()}, ${start.getFullYear()} (${start.toLocaleString(
      undefined,
      { hour12: true, hour: "numeric", minute: "numeric" }
    )} - ${end.toLocaleString(undefined, {
      hour12: true,
      hour: "numeric",
      minute: "numeric",
    })})`;
  }
  // month is the same
  if (
    start.getMonth() === end.getMonth() &&
    start.getFullYear() === end.getFullYear()
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
    })} ${start.getDate()}-${end.getDate()}, ${start.getFullYear()}`;
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

export const getTimeRangeNameForButton = (
  timeRange: TimeSeriesTimeRange
): string => {
  if (timeRange && timeRange.name) return timeRange.name;
  if (timeRange && (timeRange.start || timeRange.end)) return "Custom";
  return "Select a time range";
};
