import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import {
  isStandardGrain,
  V1TimeGrainToAlias,
} from "@rilldata/web-common/lib/time/new-grains";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import type { RangeBuckets } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";

const defaultLastNValues: Record<V1TimeGrain, number[]> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: [],
  [V1TimeGrain.TIME_GRAIN_SECOND]: [],
  [V1TimeGrain.TIME_GRAIN_MINUTE]: [15, 60],
  [V1TimeGrain.TIME_GRAIN_5_MINUTE]: [],
  [V1TimeGrain.TIME_GRAIN_10_MINUTE]: [],
  [V1TimeGrain.TIME_GRAIN_15_MINUTE]: [],
  [V1TimeGrain.TIME_GRAIN_30_MINUTE]: [],
  [V1TimeGrain.TIME_GRAIN_HOUR]: [6, 12, 24],
  [V1TimeGrain.TIME_GRAIN_2_HOUR]: [],
  [V1TimeGrain.TIME_GRAIN_3_HOUR]: [],
  [V1TimeGrain.TIME_GRAIN_4_HOUR]: [],
  [V1TimeGrain.TIME_GRAIN_6_HOUR]: [],
  [V1TimeGrain.TIME_GRAIN_8_HOUR]: [],
  [V1TimeGrain.TIME_GRAIN_12_HOUR]: [],
  [V1TimeGrain.TIME_GRAIN_DAY]: [7, 14, 30],
  [V1TimeGrain.TIME_GRAIN_WEEK]: [4],
  [V1TimeGrain.TIME_GRAIN_MONTH]: [3, 6, 12],
  [V1TimeGrain.TIME_GRAIN_QUARTER]: [],
  [V1TimeGrain.TIME_GRAIN_YEAR]: [],
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: [],
};

export function getDefaultRangeBuckets(
  allowedGrains: V1TimeGrain[],
): RangeBuckets {
  const rangeBuckets: RangeBuckets = {
    latest: [],
    periodToDate: [],
    previous: [],
    custom: [],
    allTime: false,
  };

  allowedGrains.forEach((grain) => {
    const primaryGrainAlias = V1TimeGrainToAlias[grain];

    // Multiples such as 10 minutes share their base unit's ranges.
    if (
      grain === V1TimeGrain.TIME_GRAIN_MILLISECOND ||
      grain === V1TimeGrain.TIME_GRAIN_SECOND ||
      !isStandardGrain(grain)
    ) {
      return;
    }

    defaultLastNValues[grain].forEach((v) => {
      const timeRange = `${v}${primaryGrainAlias}`;

      try {
        const parsed = parseRillTime(timeRange);
        rangeBuckets.latest.push(parsed);
      } catch {
        // no-op
      }
    });

    const timeRange = `-1${primaryGrainAlias}/${primaryGrainAlias} to ref/${primaryGrainAlias}`;

    try {
      const parsed = parseRillTime(timeRange);
      rangeBuckets.previous.push(parsed);
    } catch {
      // no-op
    }

    if (grain === V1TimeGrain.TIME_GRAIN_MINUTE) {
      return;
    }

    const periodToDate = parseRillTime(`${primaryGrainAlias}TD`);

    rangeBuckets.periodToDate.push(periodToDate);
  });

  return rangeBuckets;
}
