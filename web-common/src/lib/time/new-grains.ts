import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

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

type TimeGrain =
  | "millisecond"
  | "second"
  | "minute"
  | "hour"
  | "day"
  | "week"
  | "month"
  | "quarter"
  | "year";

type Order = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | typeof Infinity;

const V1TimeGrainToOrder: Record<V1TimeGrain, Order> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: 0,
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: 0,
  [V1TimeGrain.TIME_GRAIN_SECOND]: 1,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: 2,
  [V1TimeGrain.TIME_GRAIN_HOUR]: 3,
  [V1TimeGrain.TIME_GRAIN_DAY]: 4,
  [V1TimeGrain.TIME_GRAIN_WEEK]: 5,
  [V1TimeGrain.TIME_GRAIN_MONTH]: 6,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: 7,
  [V1TimeGrain.TIME_GRAIN_YEAR]: 8,
};

const allowedGrains = [
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

export const GrainToOrder: Record<TimeGrain, Order> = {
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
  grain: V1TimeGrain | TimeGrainAlias | TimeGrain,
  minimum: V1TimeGrain | TimeGrainAlias | TimeGrain,
) {
  const grainOrder = getGrainOrder(grain);
  const minimumOrder = getGrainOrder(minimum);

  if (grainOrder === undefined || minimumOrder === undefined) {
    return false;
  }
  return grainOrder <= minimumOrder;
}

export function getGrainOrder(
  grain: V1TimeGrain | TimeGrainAlias | TimeGrain | undefined,
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

export function getAllowedGrains(order: Order) {
  return allowedGrains.slice(order);
}
