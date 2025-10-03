import { reverseMap } from "@rilldata/web-common/lib/map-utils.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { DateTimeUnit } from "luxon";

type Order = number;

// We prevent users from aggregating by second or millisecond
export const RILL_MIN_GRAIN_VALUE = 2;

export const DateTimeUnitToOrder: Record<DateTimeUnit, Order> = {
  millisecond: 0,
  second: 1,
  minute: 2,
  hour: 3,
  day: 4,
  week: 5,
  month: 6,
  quarter: 7,
  year: 8,
};

export const OrderToDateTimeUnit: DateTimeUnit[] = [
  "millisecond",
  "second",
  "minute",
  "hour",
  "day",
  "week",
  "month",
  "quarter",
  "year",
];

export const OrderToV1TimeGrain: V1TimeGrain[] = [
  V1TimeGrain.TIME_GRAIN_MILLISECOND,
  V1TimeGrain.TIME_GRAIN_SECOND,
  V1TimeGrain.TIME_GRAIN_MINUTE,
  V1TimeGrain.TIME_GRAIN_HOUR,
  V1TimeGrain.TIME_GRAIN_DAY,
  V1TimeGrain.TIME_GRAIN_WEEK,
  V1TimeGrain.TIME_GRAIN_MONTH,
  V1TimeGrain.TIME_GRAIN_QUARTER,
  V1TimeGrain.TIME_GRAIN_YEAR,
];

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

export const V1TimeGrainToDateTimeUnit: Record<V1TimeGrain, DateTimeUnit> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]:
    OrderToDateTimeUnit[RILL_MIN_GRAIN_VALUE],
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

export const V1TimeGrainToOrder: Record<V1TimeGrain, Order> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]:
    DateTimeUnitToOrder[
      V1TimeGrainToDateTimeUnit[V1TimeGrain.TIME_GRAIN_UNSPECIFIED]
    ],
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: DateTimeUnitToOrder["millisecond"],
  [V1TimeGrain.TIME_GRAIN_SECOND]: DateTimeUnitToOrder["second"],
  [V1TimeGrain.TIME_GRAIN_MINUTE]: DateTimeUnitToOrder["minute"],
  [V1TimeGrain.TIME_GRAIN_HOUR]: DateTimeUnitToOrder["hour"],
  [V1TimeGrain.TIME_GRAIN_DAY]: DateTimeUnitToOrder["day"],
  [V1TimeGrain.TIME_GRAIN_WEEK]: DateTimeUnitToOrder["week"],
  [V1TimeGrain.TIME_GRAIN_MONTH]: DateTimeUnitToOrder["month"],
  [V1TimeGrain.TIME_GRAIN_QUARTER]: DateTimeUnitToOrder["quarter"],
  [V1TimeGrain.TIME_GRAIN_YEAR]: DateTimeUnitToOrder["year"],
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

export const DateTimeUnitToV1TimeGrain = reverseMap(V1TimeGrainToDateTimeUnit);

export function grainAliasToDateTimeUnit(alias: TimeGrainAlias): DateTimeUnit {
  const v1TimeGrain = GrainAliasToV1TimeGrain[alias];
  if (v1TimeGrain === undefined) {
    throw new Error(`Invalid time grain alias: ${alias}`);
  }
  return V1TimeGrainToDateTimeUnit[v1TimeGrain];
}

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

export function isGrainAllowed(
  grain: V1TimeGrain | TimeGrainAlias | DateTimeUnit,
  minTimeGrain: V1TimeGrain | TimeGrainAlias | DateTimeUnit | undefined,
) {
  if (!minTimeGrain) return true;
  const grainOrder = getGrainOrder(grain);
  const minimumOrder = getGrainOrder(minTimeGrain);

  if (grainOrder === -1) return false;

  return grainOrder >= minimumOrder;
}

export function getGrainOrder(
  grain: V1TimeGrain | TimeGrainAlias | DateTimeUnit | null | undefined,
): Order {
  if (!grain) return -1;

  if (grain in GrainAliasToOrder) {
    return GrainAliasToOrder[grain as TimeGrainAlias];
  } else if (grain in V1TimeGrainToOrder) {
    return V1TimeGrainToOrder[grain as V1TimeGrain];
  } else if (grain in DateTimeUnitToOrder) {
    return DateTimeUnitToOrder[grain as DateTimeUnit];
  }

  return -1;
}

function getAllowedGrainsFromOrder(order: Order) {
  return OrderToV1TimeGrain.slice(order);
}

function getSmallerGrainsFromOrders(maxOrder: Order, minOrder = 0) {
  return OrderToV1TimeGrain.slice(minOrder, maxOrder + 1);
}

export function getOptionsFromSmallestToLargest(
  largestTimeGrain: V1TimeGrain | undefined,
  smallestTimeGrain?: V1TimeGrain,
  getNextLowest: boolean = false,
) {
  const orderOfReferenceTimeGrain =
    getGrainOrder(largestTimeGrain) - (getNextLowest ? 1 : 0);
  const orderOfSmallestTimeGrain = getGrainOrder(smallestTimeGrain);

  if (orderOfReferenceTimeGrain === -1 || orderOfSmallestTimeGrain === -1) {
    return getAllowedGrainsFromOrder(RILL_MIN_GRAIN_VALUE);
  }

  return getSmallerGrainsFromOrders(
    orderOfReferenceTimeGrain,
    orderOfSmallestTimeGrain,
  );
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
      return V1TimeGrain.TIME_GRAIN_SECOND;
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
      return V1TimeGrain.TIME_GRAIN_QUARTER;
    default:
      return V1TimeGrain.TIME_GRAIN_MINUTE;
  }
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
