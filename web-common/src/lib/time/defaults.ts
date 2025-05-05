import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  getAllowedGrains,
  getGrainOrder,
  getLargerGrains,
  V1TimeGrainToAlias,
} from "@rilldata/web-common/lib/time/new-grains";
import type { RillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";

const lastNValues: Record<V1TimeGrain, number[]> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: [100, 200, 500, 1000, 200],
  [V1TimeGrain.TIME_GRAIN_SECOND]: [15, 30, 60, 90, 120],
  [V1TimeGrain.TIME_GRAIN_MINUTE]: [15, 30, 60, 90, 120],
  [V1TimeGrain.TIME_GRAIN_HOUR]: [6, 12, 24, 48, 72],
  [V1TimeGrain.TIME_GRAIN_DAY]: [3, 7, 14, 30, 90],
  [V1TimeGrain.TIME_GRAIN_WEEK]: [2, 4, 6, 12, 52],
  [V1TimeGrain.TIME_GRAIN_MONTH]: [3, 6, 12, 18, 24],
  [V1TimeGrain.TIME_GRAIN_QUARTER]: [2, 4, 6, 8, 12],
  [V1TimeGrain.TIME_GRAIN_YEAR]: [2, 3, 5, 7, 10],
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: [],
};

export interface TimeGrainOptions {
  lastN: TimeRangeMenuOption[];
  previous: TimeRangeMenuOption[];
  this: TimeRangeMenuOption[];
  grainBy: TimeRangeMenuOption[];
}

export function getTimeRangeOptionsByGrain(
  grain: V1TimeGrain,
  smallestTimeGrain: V1TimeGrain = V1TimeGrain.TIME_GRAIN_SECOND,
): TimeGrainOptions {
  const primaryGrainAlias = V1TimeGrainToAlias[grain];
  const grainOrder = getGrainOrder(grain);

  const allowedGrains = getAllowedGrains(smallestTimeGrain);
  const smallerGrains = allowedGrains.filter(
    (g) => getGrainOrder(g) < grainOrder,
  );

  const lastN = lastNValues[grain].map((v) => {
    const timeRange = `${v}${primaryGrainAlias}~`;
    const parsed = parseRillTime(timeRange);
    return {
      string: timeRange,
      parsed,
    };
  });

  const previous = Array.from({ length: 4 }, (_, i) => {
    const timeRange = `-${i + 1}${primaryGrainAlias}~`;
    const parsed = parseRillTime(timeRange);

    return {
      string: timeRange,
      parsed,
    };
  });

  const grainBy = smallerGrains.map((g) => {
    const secondaryGrainAlias = V1TimeGrainToAlias[g];
    const timeRange = `${primaryGrainAlias}T${secondaryGrainAlias}~`;
    const parsed = parseRillTime(timeRange);

    return {
      string: timeRange,
      parsed,
    };
  });

  return {
    lastN,
    this: [
      {
        string: `${primaryGrainAlias}~`,
        parsed: parseRillTime(`${primaryGrainAlias}~`),
      },
    ],
    previous,
    grainBy,
  };
}

export interface TimeRangeMenuOption {
  string: string;
  parsed: RillTime;
}
