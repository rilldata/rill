import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { DateTime, DateTimeUnit } from "luxon";

type TimeGrainAlias =
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

export const grainAliasRegex = /(ms|s|S|m|h|H|d|D|w|W|M|q|Q|y|Y)/;

const GRAINS = ["Y", "Q", "M", "W", "d", "h", "m", "s"] as const;

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
  console.log("syntax", syntax);
  if (!syntax || syntax.startsWith("P") || syntax.startsWith("rill")) {
    return [];
  }
  const alias = getGrainAliasFromString(syntax);

  console.log("alias", alias);

  if (!alias) {
    return [];
  }
  const v1TimeGrain = GrainAliasToV1TimeGrain[alias];
  if (v1TimeGrain === undefined) {
    return [];
  }

  console.log("v1TimeGrain", v1TimeGrain);

  const order = getGrainOrder(v1TimeGrain);

  if (order === undefined) {
    return [];
  }

  console.log("order", order);
  const okay = getSmallerGrainsFromOrders(
    order,
    getGrainOrder(smallestTimeGrain),
  );

  console.log("okay", okay);
  return okay;
}

type Order = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | typeof Infinity;

export const V1TimeGrainToOrder: Record<V1TimeGrain, Order> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: 0,
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: 0,
  [V1TimeGrain.TIME_GRAIN_SECOND]: 0,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: 1,
  [V1TimeGrain.TIME_GRAIN_HOUR]: 2,
  [V1TimeGrain.TIME_GRAIN_DAY]: 3,
  [V1TimeGrain.TIME_GRAIN_WEEK]: 4,
  [V1TimeGrain.TIME_GRAIN_MONTH]: 5,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: 6,
  [V1TimeGrain.TIME_GRAIN_YEAR]: 7,
};

export const V1TimeGrainToAlias: Record<V1TimeGrain, TimeGrainAlias> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "ms",
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
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "second",
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

export function grainAliasToDateTimeUnit(alias: TimeGrainAlias): DateTimeUnit {
  const v1TimeGrain = GrainAliasToV1TimeGrain[alias];
  if (v1TimeGrain === undefined) {
    throw new Error(`Invalid time grain alias: ${alias}`);
  }
  return V1TimeGrainToDateTimeUnit[v1TimeGrain];
}

const allowedGrains = [
  // V1TimeGrain.TIME_GRAIN_MILLISECOND,
  V1TimeGrain.TIME_GRAIN_SECOND,
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

export function isGrainWithinMinimum(
  grain: V1TimeGrain | TimeGrainAlias | DateTimeUnit,
  minimum: V1TimeGrain | TimeGrainAlias | DateTimeUnit,
) {
  const grainOrder = getGrainOrder(grain);
  const minimumOrder = getGrainOrder(minimum);

  if (grainOrder === undefined || minimumOrder === undefined) {
    return false;
  }
  return grainOrder <= minimumOrder;
}

export function getGrainOrder(
  grain: V1TimeGrain | TimeGrainAlias | DateTimeUnit | undefined,
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
  return allowedGrains.slice(order);
}

export function getLargerGrainsFromOrder(order: Order) {
  return allowedGrains.slice(order + 1);
}

export function getToDateExcludeOptions(
  referenceTimeGrain: V1TimeGrain,
  smallestTimeGrain?: V1TimeGrain,
) {
  const orderOfReferenceTimeGrain = getGrainOrder(referenceTimeGrain);
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

export function getSmallerGrainsFromOrders(maxOrder: Order, minOrder = 0) {
  return allowedGrains.slice(minOrder, maxOrder + 1);
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

export function getNextLowerOrderDuration(syntax: string) {
  const match = syntax.match(/(\d+)([a-zA-Z])/);
  if (!match) {
    return undefined;
  }
  const value = parseInt(match[1], 10);
  const unit = match[2];
  const v1Grain = GrainAliasToV1TimeGrain[unit as TimeGrainAlias];
  const lowerOrderUnit = getLowerOrderGrain(v1Grain);
  if (lowerOrderUnit === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) {
    return undefined;
  }
  const lowerOrderUnitAlias = V1TimeGrainToAlias[lowerOrderUnit];
  const lowerOrderUnitValue = lengths[v1Grain];
  if (!lowerOrderUnitValue) {
    return undefined;
  }
  console.log(
    `lowerOrderUnitValue: ${lowerOrderUnitValue}, value: ${value}, unit: ${unit}`,
  );
  const lowerOrderValue = value * lowerOrderUnitValue;

  return `${lowerOrderValue}${lowerOrderUnitAlias}~`;
}

const lengths = {
  [V1TimeGrain.TIME_GRAIN_SECOND]: 1000,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: 60,
  [V1TimeGrain.TIME_GRAIN_HOUR]: 60,
  [V1TimeGrain.TIME_GRAIN_DAY]: 24,
  [V1TimeGrain.TIME_GRAIN_WEEK]: 7,
  [V1TimeGrain.TIME_GRAIN_MONTH]: 30,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: 3,
  [V1TimeGrain.TIME_GRAIN_YEAR]: 12,
};

export function getLowerOrderGrain(grain: V1TimeGrain) {
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
      return V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  }
}
