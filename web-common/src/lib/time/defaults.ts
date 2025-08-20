import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { V1TimeGrainToAlias } from "@rilldata/web-common/lib/time/new-grains";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";

const defaultLastNValues: Record<V1TimeGrain, number[]> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: [],
  [V1TimeGrain.TIME_GRAIN_SECOND]: [],
  [V1TimeGrain.TIME_GRAIN_MINUTE]: [15, 60],
  [V1TimeGrain.TIME_GRAIN_HOUR]: [6, 12, 24],
  [V1TimeGrain.TIME_GRAIN_DAY]: [7, 14, 30],
  [V1TimeGrain.TIME_GRAIN_WEEK]: [4],
  [V1TimeGrain.TIME_GRAIN_MONTH]: [3, 6, 12],
  [V1TimeGrain.TIME_GRAIN_QUARTER]: [],
  [V1TimeGrain.TIME_GRAIN_YEAR]: [],
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: [],
};

export interface TimeRangeMenuOption {
  string: string;
  label: string;
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
): TimeGrainOptions {
  const primaryGrainAlias = V1TimeGrainToAlias[grain];

  const lastN: TimeRangeMenuOption[] = [];
  const previous: TimeRangeMenuOption[] = [];

  if (
    grain === V1TimeGrain.TIME_GRAIN_MILLISECOND ||
    grain === V1TimeGrain.TIME_GRAIN_SECOND
  ) {
    return {
      lastN: [],
      this: [],
      previous: [],
    };
  }

  defaultLastNValues[grain].forEach((v) => {
    const timeRange = `${v}${primaryGrainAlias}`;

    try {
      const parsed = parseRillTime(timeRange);
      lastN.push({
        string: timeRange,
        label: parsed.getLabel(),
      });
    } catch {
      // no-op
    }
  });

  const timeRange = `-1${primaryGrainAlias}/${primaryGrainAlias} to ref/${primaryGrainAlias}`;

  try {
    const parsed = parseRillTime(timeRange);
    previous.push({
      string: timeRange,
      label: parsed.getLabel(),
    });
  } catch {
    // no-op
  }

  if (grain === V1TimeGrain.TIME_GRAIN_MINUTE) {
    return {
      lastN,
      this: [],
      previous: [],
    };
  }

  const thisOption = [
    {
      string: `${primaryGrainAlias}TD`,
      label: parseRillTime(`${primaryGrainAlias}TD`).getLabel(),
    },
  ];

  return {
    lastN,
    this: thisOption,
    previous,
  };
}
