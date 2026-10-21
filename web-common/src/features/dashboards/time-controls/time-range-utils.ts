/** NOTE:
 *
 * this file should be deprecated in favor of the other time utils.
 *
 * */
import {
  type MetricsViewSpecDimension,
  MetricsViewSpecDimensionType,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { durationToMillis } from "@rilldata/web-common/lib/time/grains";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types.ts";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import {
  RillLegacyDaxInterval,
  RillPeriodToGrainInterval,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";

// Moved
export function getAllowedTimeGrains(timeRangeDurationMs) {
  if (
    timeRangeDurationMs <
    2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return [V1TimeGrain.TIME_GRAIN_MINUTE];
  } else if (
    timeRangeDurationMs <
    6 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return [V1TimeGrain.TIME_GRAIN_MINUTE, V1TimeGrain.TIME_GRAIN_HOUR];
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return [V1TimeGrain.TIME_GRAIN_HOUR];
  } else if (
    timeRangeDurationMs <
    14 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return [V1TimeGrain.TIME_GRAIN_HOUR, V1TimeGrain.TIME_GRAIN_DAY];
  } else if (
    timeRangeDurationMs <
    durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return [
      V1TimeGrain.TIME_GRAIN_HOUR,
      V1TimeGrain.TIME_GRAIN_DAY,
      V1TimeGrain.TIME_GRAIN_WEEK,
    ];
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return [V1TimeGrain.TIME_GRAIN_DAY, V1TimeGrain.TIME_GRAIN_WEEK];
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
  ) {
    return [
      V1TimeGrain.TIME_GRAIN_DAY,
      V1TimeGrain.TIME_GRAIN_WEEK,
      V1TimeGrain.TIME_GRAIN_MONTH,
    ];
  } else {
    return [
      V1TimeGrain.TIME_GRAIN_WEEK,
      V1TimeGrain.TIME_GRAIN_MONTH,
      V1TimeGrain.TIME_GRAIN_YEAR,
    ];
  }
}

// Moved
export function getDefaultTimeGrain(start: Date, end: Date): V1TimeGrain {
  const timeRangeDurationMs = end.getTime() - start.getTime();

  if (
    timeRangeDurationMs <
    2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return V1TimeGrain.TIME_GRAIN_MINUTE;
  } else if (
    timeRangeDurationMs <
    7 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return V1TimeGrain.TIME_GRAIN_HOUR;
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return V1TimeGrain.TIME_GRAIN_DAY;
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
  ) {
    return V1TimeGrain.TIME_GRAIN_WEEK;
  } else {
    return V1TimeGrain.TIME_GRAIN_MONTH;
  }
}

export function getTimeDimensionOptions(
  dimensions: MetricsViewSpecDimension[],
  restrictedDimensions: string[] | undefined,
) {
  const timeDimensions = dimensions.filter(
    (d) =>
      d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME &&
      (!restrictedDimensions || restrictedDimensions.includes(d.name!)),
  );

  if (restrictedDimensions) {
    timeDimensions.sort(
      (a, b) =>
        restrictedDimensions.indexOf(a.name!) -
        restrictedDimensions.indexOf(b.name!),
    );
  }

  return timeDimensions.map((timeDim) => {
    return {
      value: timeDim.name!,
      label: timeDim.displayName || timeDim.name!,
      description: timeDim.description,
    };
  });
}

export function getComparisonTypeFromRangeString(
  range: string | undefined,
): TimeComparisonOption {
  if (!range) {
    return TimeComparisonOption.CONTIGUOUS;
  }
  try {
    const { interval, rangeGrain } = parseRillTime(range);

    if (
      interval instanceof RillLegacyDaxInterval ||
      interval instanceof RillPeriodToGrainInterval
    ) {
      return rangeGrain && rangeGrain in timeGrainToComparisonOptionMap
        ? timeGrainToComparisonOptionMap[rangeGrain]
        : TimeComparisonOption.CONTIGUOUS;
    } else {
      return TimeComparisonOption.CONTIGUOUS;
    }
  } catch {
    return TimeComparisonOption.CONTIGUOUS;
  }
}

const timeGrainToComparisonOptionMap: Record<
  V1TimeGrain,
  TimeComparisonOption
> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: TimeComparisonOption.CONTIGUOUS,
  [V1TimeGrain.TIME_GRAIN_SECOND]: TimeComparisonOption.CONTIGUOUS,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: TimeComparisonOption.CONTIGUOUS,
  [V1TimeGrain.TIME_GRAIN_HOUR]: TimeComparisonOption.CONTIGUOUS,
  [V1TimeGrain.TIME_GRAIN_DAY]: TimeComparisonOption.DAY,
  [V1TimeGrain.TIME_GRAIN_WEEK]: TimeComparisonOption.WEEK,
  [V1TimeGrain.TIME_GRAIN_MONTH]: TimeComparisonOption.MONTH,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: TimeComparisonOption.QUARTER,
  [V1TimeGrain.TIME_GRAIN_YEAR]: TimeComparisonOption.YEAR,
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: TimeComparisonOption.CONTIGUOUS,
};
