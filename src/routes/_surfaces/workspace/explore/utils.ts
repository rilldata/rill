import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import { TimeRange, timeRanges } from "$lib/util/time-ranges";

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

export const prettyFormatTimeRange = (timeRange: TimeRange): string => {
  // day is the same
  if (
    timeRange.start.getDate() === timeRange.end.getDate() &&
    timeRange.start.getMonth() === timeRange.end.getMonth() &&
    timeRange.start.getFullYear() === timeRange.end.getFullYear()
  ) {
    return `${timeRange.start.toLocaleDateString(undefined, {
      month: "long",
    })} ${timeRange.start.getDate()}, ${timeRange.start.getFullYear()} (${timeRange.start.toLocaleString(
      undefined,
      { hour12: true, hour: "numeric", minute: "numeric" }
    )} - ${timeRange.end.toLocaleString(undefined, {
      hour12: true,
      hour: "numeric",
      minute: "numeric",
    })})`;
  }
  // month is the same
  if (
    timeRange.start.getMonth() === timeRange.end.getMonth() &&
    timeRange.start.getFullYear() === timeRange.end.getFullYear()
  ) {
    return `${timeRange.start.toLocaleDateString(undefined, {
      month: "long",
    })} ${timeRange.start.getDate()}-${timeRange.end.getDate()}, ${timeRange.start.getFullYear()}`;
  }
  // year is the same
  if (timeRange.start.getFullYear() === timeRange.end.getFullYear()) {
    return `${timeRange.start.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
    })} - ${timeRange.end.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
    })}, ${timeRange.start.getFullYear()}`;
  }
  // year is different
  const dateFormatOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "long",
    day: "numeric",
  };
  return `${timeRange.start.toLocaleDateString(
    undefined,
    dateFormatOptions
  )} - ${timeRange.end.toLocaleDateString(undefined, dateFormatOptions)}`;
};
