import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  getToDateExcludeOptions,
  V1TimeGrainToAlias,
  V1TimeGrainToDateTimeUnit,
} from "@rilldata/web-common/lib/time/new-grains";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";

const defaultLastNValues: Record<V1TimeGrain, number[]> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: [],
  [V1TimeGrain.TIME_GRAIN_SECOND]: [],
  [V1TimeGrain.TIME_GRAIN_MINUTE]: [30, 60, 90],
  [V1TimeGrain.TIME_GRAIN_HOUR]: [3, 6, 12, 24],
  [V1TimeGrain.TIME_GRAIN_DAY]: [3, 7, 14, 30],
  [V1TimeGrain.TIME_GRAIN_WEEK]: [],
  [V1TimeGrain.TIME_GRAIN_MONTH]: [3, 6, 12],
  [V1TimeGrain.TIME_GRAIN_QUARTER]: [],
  [V1TimeGrain.TIME_GRAIN_YEAR]: [2, 5],
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: [],
};

export interface TimeRangeMenuOption {
  string: string;
  label: string;
  // alts: TimeRangeAlt[];
}

export interface TimeRangeAlt {
  string: string;
  label: string;
}

export interface TimeGrainOptions {
  lastN: TimeRangeMenuOption[];
  previous: TimeRangeMenuOption[];
  this: TimeRangeMenuOption[];
}

export function getTimeRangeOptionsByGrain(
  grain: V1TimeGrain,
  smallestTimeGrain: V1TimeGrain,
): TimeGrainOptions {
  const primaryGrainAlias = V1TimeGrainToAlias[grain];
  const primaryGrainUnit = V1TimeGrainToDateTimeUnit[grain];
  // const allowedGrains = getAllowedGrains(smallestTimeGrain);

  const lastN: TimeRangeMenuOption[] = defaultLastNValues[grain].map((v) => {
    const timeRange = `-${v}${primaryGrainAlias}$ to ${primaryGrainAlias}$`;
    const parsed = parseRillTime(timeRange);
    return {
      string: timeRange,
      label: parsed.getLabel(),
    };
  });

  const previous: TimeRangeMenuOption[] = Array.from({ length: 1 }, (_, i) => {
    const timeRange = `-${i + 1}${primaryGrainAlias}^ to ${primaryGrainAlias}^`;
    const parsed = parseRillTime(timeRange);

    return {
      string: timeRange,
      label: parsed.getLabel(),
    };
  });

  const allowedGrains = getToDateExcludeOptions(grain, smallestTimeGrain);

  if (grain === V1TimeGrain.TIME_GRAIN_MINUTE) {
    return {
      lastN,
      this: [],
      previous: [],
    };
  }

  const thisOption = [
    {
      string: `${primaryGrainAlias}!`,
      label: parseRillTime(`${primaryGrainAlias}!`).getLabel(),
      // alts: allowedGrains.map((g) => {
      //   if (g === grain) {
      //     return {
      //       string: `${primaryGrainAlias}^ to ${primaryGrainAlias}$`,
      //       label: `in full`,
      //     };
      //   }

      //   const grainAlias = V1TimeGrainToAlias[g];
      //   const unit = V1TimeGrainToDateTimeUnit[g];
      //   return {
      //     string: `${primaryGrainAlias}^ to ${grainAlias}^`,
      //     label: `excluding this ${unit}`,
      //   };
      // }),
    },
  ];
  return {
    lastN,
    this: thisOption,
    previous,
  };
}
