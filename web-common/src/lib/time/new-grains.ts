import {
  RillLegacyDaxInterval,
  RillLegacyIsoInterval,
  type RillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime";
import { reverseMap } from "@rilldata/web-common/lib/map-utils.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { type Interval, type DateTimeUnit } from "luxon";

type Order = 0 | 1 | 2 | 3 | 4 | 5 | 6 | typeof Infinity;

export type TimeGrainAlias =
  | "ms"
  | "MS"
  | "s"
  | "S"
  | "m"
  | "h"
  | "H"
  | "d"
  | "D"
  | "w"
  | "W"
  | "M"
  | "q"
  | "Q"
  | "y"
  | "Y";

export const grainAliasRegex = /(ms|MS|s|S|m|h|H|d|D|w|W|M|q|Q|y|Y)/;

export function getGrainAliasFromString(
  range: string,
): TimeGrainAlias | undefined {
  const match = range.match(grainAliasRegex);
  if (match) {
    return match[0] as TimeGrainAlias;
  }
  return undefined;
}

export function getAllowedEndingGrains(
  syntax: string | undefined,
  smallestTimeGrain?: V1TimeGrain,
) {
  if (!syntax || syntax.startsWith("P") || syntax.startsWith("rill")) {
    return [];
  }
  const alias = getGrainAliasFromString(syntax);

  if (!alias) {
    return [];
  }
  const v1TimeGrain = GrainAliasToV1TimeGrain[alias];
  if (v1TimeGrain === undefined) {
    return [];
  }

  const order = getGrainOrder(v1TimeGrain);

  if (order === undefined) {
    return [];
  }

  return getSmallerGrainsFromOrders(order, getGrainOrder(smallestTimeGrain));
}

export const V1TimeGrainToOrder: Record<V1TimeGrain, Order> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: 0,
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: 0,
  [V1TimeGrain.TIME_GRAIN_SECOND]: 0,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: 0,
  [V1TimeGrain.TIME_GRAIN_HOUR]: 1,
  [V1TimeGrain.TIME_GRAIN_DAY]: 2,
  [V1TimeGrain.TIME_GRAIN_WEEK]: 3,
  [V1TimeGrain.TIME_GRAIN_MONTH]: 4,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: 5,
  [V1TimeGrain.TIME_GRAIN_YEAR]: 6,
};

export const V1TimeGrainToAlias: Record<V1TimeGrain, TimeGrainAlias> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "m",
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "ms",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "s",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "m",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "h",
  [V1TimeGrain.TIME_GRAIN_DAY]: "D",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "W",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "M",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "Q",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "Y",
};

export const V1TimeGrainToDateTimeUnit: Record<V1TimeGrain, DateTimeUnit> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "minute",
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "millisecond",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "second",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "minute",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "hour",
  [V1TimeGrain.TIME_GRAIN_DAY]: "day",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "week",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "month",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "quarter",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "year",
};

export const DateTimeUnitToV1TimeGrain = reverseMap(V1TimeGrainToDateTimeUnit);

export function grainAliasToDateTimeUnit(alias: TimeGrainAlias): DateTimeUnit {
  const v1TimeGrain = GrainAliasToV1TimeGrain[alias];
  if (v1TimeGrain === undefined) {
    throw new Error(`Invalid time grain alias: ${alias}`);
  }
  return V1TimeGrainToDateTimeUnit[v1TimeGrain];
}

// We prevent users from aggregating by second or millisecond
export const allowedAggregationGrains = [
  V1TimeGrain.TIME_GRAIN_MINUTE,
  V1TimeGrain.TIME_GRAIN_HOUR,
  V1TimeGrain.TIME_GRAIN_DAY,
  V1TimeGrain.TIME_GRAIN_WEEK,
  V1TimeGrain.TIME_GRAIN_MONTH,
  V1TimeGrain.TIME_GRAIN_QUARTER,
  V1TimeGrain.TIME_GRAIN_YEAR,
];

export const GrainAliasToOrder: Record<TimeGrainAlias, Order> = {
  ms: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MILLISECOND],
  MS: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MILLISECOND],
  s: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_SECOND],
  S: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_SECOND],
  m: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MINUTE],
  h: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_HOUR],
  H: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_HOUR],
  d: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_DAY],
  D: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_DAY],
  w: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_WEEK],
  W: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_WEEK],
  M: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MONTH],
  q: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_QUARTER],
  Q: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_QUARTER],
  y: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_YEAR],
  Y: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_YEAR],
};

export const GrainAliasToV1TimeGrain: Record<TimeGrainAlias, V1TimeGrain> = {
  ms: V1TimeGrain.TIME_GRAIN_MILLISECOND,
  MS: V1TimeGrain.TIME_GRAIN_MILLISECOND,
  s: V1TimeGrain.TIME_GRAIN_SECOND,
  S: V1TimeGrain.TIME_GRAIN_SECOND,
  m: V1TimeGrain.TIME_GRAIN_MINUTE,
  h: V1TimeGrain.TIME_GRAIN_HOUR,
  H: V1TimeGrain.TIME_GRAIN_HOUR,
  d: V1TimeGrain.TIME_GRAIN_DAY,
  D: V1TimeGrain.TIME_GRAIN_DAY,
  w: V1TimeGrain.TIME_GRAIN_WEEK,
  W: V1TimeGrain.TIME_GRAIN_WEEK,
  M: V1TimeGrain.TIME_GRAIN_MONTH,
  q: V1TimeGrain.TIME_GRAIN_QUARTER,
  Q: V1TimeGrain.TIME_GRAIN_QUARTER,
  y: V1TimeGrain.TIME_GRAIN_YEAR,
  Y: V1TimeGrain.TIME_GRAIN_YEAR,
};

export const GrainToOrder: Record<DateTimeUnit, Order> = {
  millisecond: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MILLISECOND],
  second: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_SECOND],
  minute: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MINUTE],
  hour: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_HOUR],
  day: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_DAY],
  week: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_WEEK],
  month: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MONTH],
  quarter: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_QUARTER],
  year: V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_YEAR],
};

export function isGrainAllowed(
  grain: V1TimeGrain | TimeGrainAlias | DateTimeUnit | undefined,
  minTimeGrain: V1TimeGrain | TimeGrainAlias | DateTimeUnit | undefined,
) {
  if (!grain) return false;
  if (!minTimeGrain) return true;
  const grainOrder = getGrainOrder(grain);
  const minimumOrder = getGrainOrder(minTimeGrain);

  if (grainOrder === -1) return false;

  return grainOrder >= minimumOrder;
}

export function getGrainOrder(
  grain: V1TimeGrain | TimeGrainAlias | DateTimeUnit | null | undefined,
): Order {
  if (!grain) return Infinity;

  if (grain in GrainAliasToOrder) {
    return GrainAliasToOrder[grain];
  } else if (grain in V1TimeGrainToOrder) {
    return V1TimeGrainToOrder[grain];
  } else if (grain in GrainToOrder) {
    return GrainToOrder[grain];
  }
  return Infinity;
}

export function getAllowedGrainsFromOrder(order: Order) {
  return allowedAggregationGrains.slice(order);
}

export function getLargerGrainsFromOrder(order: Order) {
  return allowedAggregationGrains.slice(order + 1);
}

export function getSmallerGrainsFromOrders(maxOrder: Order, minOrder = 0) {
  return allowedAggregationGrains.slice(minOrder, maxOrder + 1);
}

export function getOptionsFromSmallestToLargest(
  largestTimeGrain: V1TimeGrain | undefined,
  smallestTimeGrain?: V1TimeGrain,
  getNextLowest: boolean = false,
) {
  const orderOfReferenceTimeGrain =
    getGrainOrder(largestTimeGrain) - (getNextLowest ? 1 : 0);
  const orderOfSmallestTimeGrain = getGrainOrder(smallestTimeGrain);

  if (
    orderOfReferenceTimeGrain === undefined ||
    orderOfSmallestTimeGrain === undefined
  ) {
    return [];
  }

  return getSmallerGrainsFromOrders(
    orderOfReferenceTimeGrain,
    orderOfSmallestTimeGrain,
  );
}

export function getLargerGrains(grain: V1TimeGrain | TimeGrainAlias) {
  const order = getGrainOrder(grain);
  if (order === undefined) {
    return [];
  }
  return getLargerGrainsFromOrder(order);
}

export function getMinGrain(
  grain1: V1TimeGrain | undefined,
  grain2: V1TimeGrain | undefined,
) {
  if (grain1 === undefined) {
    return grain2;
  }
  if (grain2 === undefined) {
    return grain1;
  }
  const order1 = getGrainOrder(grain1);
  const order2 = getGrainOrder(grain2);
  return order1 <= order2 ? grain1 : grain2;
}

export function getMaxGrain(
  grain1: V1TimeGrain | undefined,
  grain2: V1TimeGrain | undefined,
) {
  if (grain1 === undefined) {
    return grain2;
  }
  if (grain2 === undefined) {
    return grain1;
  }
  const order1 = getGrainOrder(grain1);
  const order2 = getGrainOrder(grain2);
  return order1 > order2 ? grain1 : grain2;
}

export function getAllowedGrains(
  grain: V1TimeGrain | TimeGrainAlias | DateTimeUnit | undefined,
) {
  const order = getGrainOrder(grain);
  if (order === undefined) {
    return [];
  }
  return getAllowedGrainsFromOrder(order);
}

export function getLowerOrderGrain(grain: V1TimeGrain): V1TimeGrain {
  switch (grain) {
    case V1TimeGrain.TIME_GRAIN_MILLISECOND:
      return V1TimeGrain.TIME_GRAIN_MILLISECOND;
    case V1TimeGrain.TIME_GRAIN_SECOND:
      return V1TimeGrain.TIME_GRAIN_MILLISECOND;
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return V1TimeGrain.TIME_GRAIN_MINUTE;
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return V1TimeGrain.TIME_GRAIN_MINUTE;
    case V1TimeGrain.TIME_GRAIN_DAY:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return V1TimeGrain.TIME_GRAIN_DAY;
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return V1TimeGrain.TIME_GRAIN_DAY;
    case V1TimeGrain.TIME_GRAIN_QUARTER:
      return V1TimeGrain.TIME_GRAIN_MONTH;
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return V1TimeGrain.TIME_GRAIN_MONTH;
    default:
      return V1TimeGrain.TIME_GRAIN_MINUTE;
  }
}

export function getSmallestGrainFromISODuration(
  duration: string,
): V1TimeGrain | null {
  const grains: V1TimeGrain[] = [];

  const upper = duration.toUpperCase();
  const [datePartRaw, timePartRaw] = upper.split("T");
  const datePart = datePartRaw ?? "";
  const timePart = timePartRaw ?? "";

  if (/\d+Y/.test(datePart)) grains.push(V1TimeGrain.TIME_GRAIN_YEAR);
  if (/\d+M/.test(datePart)) grains.push(V1TimeGrain.TIME_GRAIN_MONTH);
  if (/\d+W/.test(datePart)) grains.push(V1TimeGrain.TIME_GRAIN_WEEK);
  if (/\d+D/.test(datePart)) grains.push(V1TimeGrain.TIME_GRAIN_DAY);

  if (/\d+H/.test(timePart)) grains.push(V1TimeGrain.TIME_GRAIN_HOUR);
  if (/\d+M/.test(timePart)) grains.push(V1TimeGrain.TIME_GRAIN_MINUTE);
  if (/\d+S/.test(timePart)) grains.push(V1TimeGrain.TIME_GRAIN_SECOND);

  if (grains.length === 0) return null;

  return grains.reduce((smallest, current) => {
    return V1TimeGrainToOrder[current] < V1TimeGrainToOrder[smallest]
      ? current
      : smallest;
  });
}

export const minTimeGrainToDefaultTimeRange: Record<V1TimeGrain, string> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "24h as of latest/h+1h",
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "24h as of latest/h+1h",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "24h as of latest/h+1h",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "24h as of latest/h+1h",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "24h as of latest/h",
  [V1TimeGrain.TIME_GRAIN_DAY]: "7d as of latest/d",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "4w as of latest/w",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "3M as of latest/M",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "4Q as of latest/Q",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "5y as of latest/Y",
};

export function getRangePrecision(rillTime: RillTime) {
  const asOfSnap = rillTime.asOfLabel?.snap;

  const asOfSnapV1Grain = GrainAliasToV1TimeGrain[asOfSnap as TimeGrainAlias];
  const rangeV1Grain = rillTime.rangeGrain;
  const intervalV1Grain = rillTime.interval.getGrain();

  return getSmallestGrain([asOfSnapV1Grain, rangeV1Grain, intervalV1Grain]);
}

export function getSmallestGrain(grains: (V1TimeGrain | undefined)[]) {
  if (grains.length === 0) {
    return undefined;
  }

  return grains.reduce(
    (smallest, current) => {
      if (!current) return smallest;
      if (!smallest) return current;
      return V1TimeGrainToOrder[current] < V1TimeGrainToOrder[smallest]
        ? current
        : smallest;
    },
    undefined as V1TimeGrain | undefined,
  );
}

export function getAggregationGrain(rillTime: RillTime | undefined) {
  if (!rillTime) return undefined;

  const asOfSnap = rillTime.asOfLabel?.snap;

  const asOfSnapV1Grain = GrainAliasToV1TimeGrain[asOfSnap as TimeGrainAlias];
  const rangeV1Grain = rillTime.rangeGrain;
  const intervalV1Grain = rillTime.interval.getGrain();

  return getSmallestGrain([asOfSnapV1Grain, rangeV1Grain, intervalV1Grain]);
}

export function getTruncationGrain(rillTime: RillTime | undefined) {
  if (!rillTime) return undefined;

  const asOfSnap = rillTime.asOfLabel?.snap;

  if (asOfSnap) return GrainAliasToV1TimeGrain[asOfSnap as TimeGrainAlias];

  if (rillTime.interval instanceof RillLegacyIsoInterval) {
    return rillTime.interval.getGrain();
  }

  if (rillTime.interval instanceof RillLegacyDaxInterval) {
    if (rillTime.interval.name.endsWith("C")) return undefined;
    return V1TimeGrain.TIME_GRAIN_DAY;
  }

  return undefined;
}

const MAX_BUCKETS = 1500;

const ALLOWABLE_AGGREGATION_GRAINS: DateTimeUnit[] =
  allowedAggregationGrains.map((grain) => V1TimeGrainToDateTimeUnit[grain]);

export function allowedGrainsForInterval(
  interval: Interval<true> | undefined,
  minTimeGrain?: V1TimeGrain,
): V1TimeGrain[] {
  minTimeGrain = minTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE;
  if (!interval) return [];

  const validGrains = ALLOWABLE_AGGREGATION_GRAINS.filter((unit) => {
    return isGrainAllowed(unit, minTimeGrain);
  });

  const allowedGrains = validGrains
    .filter((unit) => {
      const grain = DateTimeUnitToV1TimeGrain[unit];
      if (!grain) return false;
      const bucketCount = interval.length(unit);

      return (
        bucketCount >= 1 && (bucketCount <= MAX_BUCKETS || unit === "year")
      );
    })
    .map((unit) => DateTimeUnitToV1TimeGrain[unit]!);

  if (allowedGrains.length) {
    return allowedGrains;
  } else {
    return [minTimeGrain];
  }
}
