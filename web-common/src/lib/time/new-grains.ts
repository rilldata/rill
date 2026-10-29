import { reverseMap } from "@rilldata/web-common/lib/map-utils.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import {
  V1TimeGrain,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import type { DateTime, DateTimeUnit, Interval } from "luxon";

const MAX_BUCKETS = 1500;

/** Position of a grain in `aggregationGrainOrder`; larger is coarser. */
type Order = number;

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

/**
 * Every grain a time series can be bucketed by, from fine to coarse.
 * Multiples such as TIME_GRAIN_10_MINUTE sit between their base unit and the next standard grain.
 * A grain's index in this list is its order.
 */
export const aggregationGrainOrder: V1TimeGrain[] = [
  V1TimeGrain.TIME_GRAIN_MINUTE,
  V1TimeGrain.TIME_GRAIN_5_MINUTE,
  V1TimeGrain.TIME_GRAIN_10_MINUTE,
  V1TimeGrain.TIME_GRAIN_15_MINUTE,
  V1TimeGrain.TIME_GRAIN_30_MINUTE,
  V1TimeGrain.TIME_GRAIN_HOUR,
  V1TimeGrain.TIME_GRAIN_2_HOUR,
  V1TimeGrain.TIME_GRAIN_3_HOUR,
  V1TimeGrain.TIME_GRAIN_4_HOUR,
  V1TimeGrain.TIME_GRAIN_6_HOUR,
  V1TimeGrain.TIME_GRAIN_8_HOUR,
  V1TimeGrain.TIME_GRAIN_12_HOUR,
  V1TimeGrain.TIME_GRAIN_DAY,
  V1TimeGrain.TIME_GRAIN_WEEK,
  V1TimeGrain.TIME_GRAIN_MONTH,
  V1TimeGrain.TIME_GRAIN_QUARTER,
  V1TimeGrain.TIME_GRAIN_YEAR,
];

export const V1TimeGrainToOrder: Record<V1TimeGrain, Order> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: 0,
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: 0,
  [V1TimeGrain.TIME_GRAIN_SECOND]: 0,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: 0,
  [V1TimeGrain.TIME_GRAIN_5_MINUTE]: 1,
  [V1TimeGrain.TIME_GRAIN_10_MINUTE]: 2,
  [V1TimeGrain.TIME_GRAIN_15_MINUTE]: 3,
  [V1TimeGrain.TIME_GRAIN_30_MINUTE]: 4,
  [V1TimeGrain.TIME_GRAIN_HOUR]: 5,
  [V1TimeGrain.TIME_GRAIN_2_HOUR]: 6,
  [V1TimeGrain.TIME_GRAIN_3_HOUR]: 7,
  [V1TimeGrain.TIME_GRAIN_4_HOUR]: 8,
  [V1TimeGrain.TIME_GRAIN_6_HOUR]: 9,
  [V1TimeGrain.TIME_GRAIN_8_HOUR]: 10,
  [V1TimeGrain.TIME_GRAIN_12_HOUR]: 11,
  [V1TimeGrain.TIME_GRAIN_DAY]: 12,
  [V1TimeGrain.TIME_GRAIN_WEEK]: 13,
  [V1TimeGrain.TIME_GRAIN_MONTH]: 14,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: 15,
  [V1TimeGrain.TIME_GRAIN_YEAR]: 16,
};

/** Number of base units in one bucket of the grain: 1 for standard grains, 10 for TIME_GRAIN_10_MINUTE. */
export const V1TimeGrainToCount: Record<V1TimeGrain, number> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: 1,
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: 1,
  [V1TimeGrain.TIME_GRAIN_SECOND]: 1,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: 1,
  [V1TimeGrain.TIME_GRAIN_5_MINUTE]: 5,
  [V1TimeGrain.TIME_GRAIN_10_MINUTE]: 10,
  [V1TimeGrain.TIME_GRAIN_15_MINUTE]: 15,
  [V1TimeGrain.TIME_GRAIN_30_MINUTE]: 30,
  [V1TimeGrain.TIME_GRAIN_HOUR]: 1,
  [V1TimeGrain.TIME_GRAIN_2_HOUR]: 2,
  [V1TimeGrain.TIME_GRAIN_3_HOUR]: 3,
  [V1TimeGrain.TIME_GRAIN_4_HOUR]: 4,
  [V1TimeGrain.TIME_GRAIN_6_HOUR]: 6,
  [V1TimeGrain.TIME_GRAIN_8_HOUR]: 8,
  [V1TimeGrain.TIME_GRAIN_12_HOUR]: 12,
  [V1TimeGrain.TIME_GRAIN_DAY]: 1,
  [V1TimeGrain.TIME_GRAIN_WEEK]: 1,
  [V1TimeGrain.TIME_GRAIN_MONTH]: 1,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: 1,
  [V1TimeGrain.TIME_GRAIN_YEAR]: 1,
};

/** Whether the grain is a standard unit rather than a multiple of one. */
export function isStandardGrain(grain: V1TimeGrain): boolean {
  return V1TimeGrainToCount[grain] === 1;
}

/** The rill-time alias of the grain's base unit. */
export const V1TimeGrainToAlias: Record<V1TimeGrain, TimeGrainAlias> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "m",
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "ms",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "s",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "m",
  [V1TimeGrain.TIME_GRAIN_5_MINUTE]: "m",
  [V1TimeGrain.TIME_GRAIN_10_MINUTE]: "m",
  [V1TimeGrain.TIME_GRAIN_15_MINUTE]: "m",
  [V1TimeGrain.TIME_GRAIN_30_MINUTE]: "m",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "h",
  [V1TimeGrain.TIME_GRAIN_2_HOUR]: "h",
  [V1TimeGrain.TIME_GRAIN_3_HOUR]: "h",
  [V1TimeGrain.TIME_GRAIN_4_HOUR]: "h",
  [V1TimeGrain.TIME_GRAIN_6_HOUR]: "h",
  [V1TimeGrain.TIME_GRAIN_8_HOUR]: "h",
  [V1TimeGrain.TIME_GRAIN_12_HOUR]: "h",
  [V1TimeGrain.TIME_GRAIN_DAY]: "D",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "W",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "M",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "Q",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "Y",
};

/** The luxon unit of the grain's base unit; combine with `V1TimeGrainToCount` for multiples. */
export const V1TimeGrainToDateTimeUnit: Record<V1TimeGrain, DateTimeUnit> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "minute",
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "millisecond",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "second",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "minute",
  [V1TimeGrain.TIME_GRAIN_5_MINUTE]: "minute",
  [V1TimeGrain.TIME_GRAIN_10_MINUTE]: "minute",
  [V1TimeGrain.TIME_GRAIN_15_MINUTE]: "minute",
  [V1TimeGrain.TIME_GRAIN_30_MINUTE]: "minute",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "hour",
  [V1TimeGrain.TIME_GRAIN_2_HOUR]: "hour",
  [V1TimeGrain.TIME_GRAIN_3_HOUR]: "hour",
  [V1TimeGrain.TIME_GRAIN_4_HOUR]: "hour",
  [V1TimeGrain.TIME_GRAIN_6_HOUR]: "hour",
  [V1TimeGrain.TIME_GRAIN_8_HOUR]: "hour",
  [V1TimeGrain.TIME_GRAIN_12_HOUR]: "hour",
  [V1TimeGrain.TIME_GRAIN_DAY]: "day",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "week",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "month",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "quarter",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "year",
};

/** Luxon unit to the standard grain of that unit. */
export const DateTimeUnitToV1TimeGrain = reverseMap(
  Object.fromEntries(
    Object.entries(V1TimeGrainToDateTimeUnit).filter(([grain]) =>
      isStandardGrain(grain as V1TimeGrain),
    ),
  ) as Record<V1TimeGrain, DateTimeUnit>,
);

/**
 * The grain's token in URLs (`tg=10m`), explore presets, component `time_filters` and the metrics
 * resolver (`grain: "10m"`): the luxon unit name for standard grains and `<n>m` or `<n>h` for multiples.
 */
export const V1TimeGrainToToken: Record<V1TimeGrain, string> = {
  ...V1TimeGrainToDateTimeUnit,
  [V1TimeGrain.TIME_GRAIN_5_MINUTE]: "5m",
  [V1TimeGrain.TIME_GRAIN_10_MINUTE]: "10m",
  [V1TimeGrain.TIME_GRAIN_15_MINUTE]: "15m",
  [V1TimeGrain.TIME_GRAIN_30_MINUTE]: "30m",
  [V1TimeGrain.TIME_GRAIN_2_HOUR]: "2h",
  [V1TimeGrain.TIME_GRAIN_3_HOUR]: "3h",
  [V1TimeGrain.TIME_GRAIN_4_HOUR]: "4h",
  [V1TimeGrain.TIME_GRAIN_6_HOUR]: "6h",
  [V1TimeGrain.TIME_GRAIN_8_HOUR]: "8h",
  [V1TimeGrain.TIME_GRAIN_12_HOUR]: "12h",
};

export const TokenToV1TimeGrain: Record<string, V1TimeGrain> = {
  ...DateTimeUnitToV1TimeGrain,
  ...reverseMap(
    Object.fromEntries(
      Object.entries(V1TimeGrainToToken).filter(
        ([grain]) => !isStandardGrain(grain as V1TimeGrain),
      ),
    ) as Record<V1TimeGrain, string>,
  ),
} as Record<string, V1TimeGrain>;

/** Floors a time to the start of its bucket. Multiples floor on the wall clock within the hour or day, like the runtime. */
export function floorToGrain<IsValid extends boolean>(
  dt: DateTime<IsValid>,
  grain: V1TimeGrain,
): DateTime<IsValid> {
  const unit = V1TimeGrainToDateTimeUnit[grain];
  const count = V1TimeGrainToCount[grain];
  const floored = dt.startOf(unit);
  if (count === 1) return floored;
  if (unit === "minute") {
    return floored.set({ minute: floored.minute - (floored.minute % count) });
  }
  return floored.set({ hour: floored.hour - (floored.hour % count) });
}

/** Adds `n` buckets of the grain. */
export function plusGrain<IsValid extends boolean>(
  dt: DateTime<IsValid>,
  grain: V1TimeGrain,
  n = 1,
): DateTime<IsValid> {
  return dt.plus({
    [V1TimeGrainToDateTimeUnit[grain]]: n * V1TimeGrainToCount[grain],
  });
}

/** Number of buckets of the grain in the interval. */
export function grainBucketCount(
  interval: Interval<true>,
  grain: V1TimeGrain,
): number {
  return (
    interval.length(V1TimeGrainToDateTimeUnit[grain]) /
    V1TimeGrainToCount[grain]
  );
}

/**
 * The grains a metrics view offers in grain pickers: its `time_grains` when declared,
 * otherwise the standard grains at or above `smallest_time_grain`.
 */
export function getAvailableTimeGrains(
  metricsViewSpec:
    | Pick<V1MetricsViewSpec, "timeGrains" | "smallestTimeGrain">
    | undefined,
): V1TimeGrain[] {
  if (metricsViewSpec?.timeGrains?.length) return metricsViewSpec.timeGrains;
  return allowedAggregationGrains.filter((grain) =>
    isGrainAllowed(grain, metricsViewSpec?.smallestTimeGrain),
  );
}

export function grainAliasToDateTimeUnit(alias: TimeGrainAlias): DateTimeUnit {
  const v1TimeGrain = GrainAliasToV1TimeGrain[alias];
  if (v1TimeGrain === undefined) {
    throw new Error(`Invalid time grain alias: ${alias}`);
  }
  return V1TimeGrainToDateTimeUnit[v1TimeGrain];
}

/**
 * The standard aggregation grains, which a metrics view offers unless it declares `time_grains`.
 * Seconds and milliseconds are never offered.
 */
export const allowedAggregationGrains =
  aggregationGrainOrder.filter(isStandardGrain);

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
  return aggregationGrainOrder.slice(order).filter(isStandardGrain);
}

export function getLargerGrainsFromOrder(order: Order) {
  return aggregationGrainOrder.slice(order + 1).filter(isStandardGrain);
}

export function getSmallerGrainsFromOrders(maxOrder: Order, minOrder = 0) {
  return aggregationGrainOrder
    .slice(minOrder, maxOrder + 1)
    .filter(isStandardGrain);
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

export function snapToDayOrLargerGrain(
  date: DateTime<true>,
  grain: DateTimeUnit,
  zone: string,
) {
  const snapToGrain = GrainToOrder[grain] > GrainToOrder.day ? grain : "day";
  return date
    .setZone(zone)
    .plus({ [snapToGrain]: 1 })
    .startOf(snapToGrain)
    .toUTC();
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
    case V1TimeGrain.TIME_GRAIN_2_HOUR:
    case V1TimeGrain.TIME_GRAIN_3_HOUR:
    case V1TimeGrain.TIME_GRAIN_4_HOUR:
    case V1TimeGrain.TIME_GRAIN_6_HOUR:
    case V1TimeGrain.TIME_GRAIN_8_HOUR:
    case V1TimeGrain.TIME_GRAIN_12_HOUR:
      return V1TimeGrain.TIME_GRAIN_HOUR;
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
  [V1TimeGrain.TIME_GRAIN_5_MINUTE]: "24h as of latest/h+1h",
  [V1TimeGrain.TIME_GRAIN_10_MINUTE]: "24h as of latest/h+1h",
  [V1TimeGrain.TIME_GRAIN_15_MINUTE]: "24h as of latest/h+1h",
  [V1TimeGrain.TIME_GRAIN_30_MINUTE]: "24h as of latest/h+1h",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "24h as of latest/h",
  [V1TimeGrain.TIME_GRAIN_2_HOUR]: "24h as of latest/h",
  [V1TimeGrain.TIME_GRAIN_3_HOUR]: "24h as of latest/h",
  [V1TimeGrain.TIME_GRAIN_4_HOUR]: "24h as of latest/h",
  [V1TimeGrain.TIME_GRAIN_6_HOUR]: "24h as of latest/h",
  [V1TimeGrain.TIME_GRAIN_8_HOUR]: "24h as of latest/h",
  [V1TimeGrain.TIME_GRAIN_12_HOUR]: "24h as of latest/h",
  [V1TimeGrain.TIME_GRAIN_DAY]: "7d as of latest/d",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "4w as of latest/w",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "3M as of latest/M",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "4Q as of latest/Q",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "5y as of latest/Y",
};

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

/**
 * The grains of `availableGrains` (by default the standard grains) that bucket the interval into a sensible
 * number of points, excluding grains finer than `minTimeGrain`.
 * Falls back to the finest usable grain when none fits.
 */
export function allowedGrainsForInterval(
  interval: Interval<true> | undefined,
  minTimeGrain?: V1TimeGrain,
  availableGrains: V1TimeGrain[] = allowedAggregationGrains,
): V1TimeGrain[] {
  minTimeGrain = minTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE;
  if (!interval) return [];

  const validGrains = availableGrains.filter((grain) =>
    isGrainAllowed(grain, minTimeGrain),
  );

  const allowedGrains = validGrains.filter((grain) => {
    const bucketCount = grainBucketCount(interval, grain);
    return (
      bucketCount >= 1 &&
      (bucketCount <= MAX_BUCKETS || grain === V1TimeGrain.TIME_GRAIN_YEAR)
    );
  });

  if (allowedGrains.length) {
    return allowedGrains;
  } else {
    return [validGrains[0] ?? minTimeGrain];
  }
}

/**
 * Translate a Luxon DateTimeUnit grain name (singular) to a localized display string.
 * Must be called from within a function, reactive block, or template — never at module scope.
 */
export function translateGrainName(grain: string): string {
  switch (grain) {
    case "second":
      return m.time_grain_second();
    case "minute":
      return m.time_grain_minute();
    case "hour":
      return m.time_grain_hour();
    case "day":
      return m.time_grain_day();
    case "week":
      return m.time_grain_week();
    case "month":
      return m.time_grain_month();
    case "quarter":
      return m.time_grain_quarter();
    case "year":
      return m.time_grain_year();
    default:
      return grain;
  }
}

/**
 * Translate a Luxon DateTimeUnit grain name (plural) to a localized display string.
 * Must be called from within a function, reactive block, or template — never at module scope.
 */
export function translateGrainNamePlural(grain: string): string {
  switch (grain) {
    case "second":
    case "seconds":
      return m.time_grain_seconds();
    case "minute":
    case "minutes":
      return m.time_grain_minutes();
    case "hour":
    case "hours":
      return m.time_grain_hours();
    case "day":
    case "days":
      return m.time_grain_days();
    case "week":
    case "weeks":
      return m.time_grain_weeks();
    case "month":
    case "months":
      return m.time_grain_months();
    case "quarter":
    case "quarters":
      return m.time_grain_quarters();
    case "year":
    case "years":
      return m.time_grain_years();
    default:
      return grain;
  }
}

/**
 * Translate a V1TimeGrain enum value to a localized display string.
 * Must be called from within a function, reactive block, or template — never at module scope.
 */
export function translateV1TimeGrain(grain: V1TimeGrain): string {
  const unit = V1TimeGrainToDateTimeUnit[grain];
  const count = V1TimeGrainToCount[grain];
  if (count > 1) {
    return unit === "hour"
      ? m.time_grain_n_hours({ count })
      : m.time_grain_n_minutes({ count });
  }
  return translateGrainName(unit);
}
