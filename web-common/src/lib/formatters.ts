import { format } from "d3-format";
import { timeFormat } from "d3-time-format";
import type { Interval } from "./duckdb-data-types";
import {
  CATEGORICALS,
  FLOATS,
  INTEGERS,
  INTERVALS,
  PreviewRollupInterval,
  TIMESTAMPS,
} from "./duckdb-data-types";

/** This heuristic is courtesy Dominik Moritz.
 * Best used in cases where (1) you have no context for the number, and (2) you
 * want have "enough resolution to distinguish numbers when they should be distinguishable."
 */
export function justEnoughPrecision(n: number) {
  if (typeof n !== "number") throw Error("argument must be a number");
  const str = n.toString();
  // if there are no floating point digits, return the string
  if (n === ~~n) return str;
  const [left, right] = str.split(".");

  // count the integer side
  const leftSideDigits = left
    .split("")
    .filter((l) => l !== "-") // remove the negative sign
    .join("").length;

  // calculate the remaining available precision
  const remainingPrecision = Math.max(0, 5 - leftSideDigits);
  // take the remaining precision from the floating point side.
  const remainingFloatingPoints = right.slice(0, remainingPrecision);
  // format a new string
  return `${left}${remainingFloatingPoints.length ? "." : ""}${
    remainingFloatingPoints || ""
  }`;
}

const zeroPad = format("02d");
const msPad = format("03d");
export const formatInteger = format(",");
const formatRate = format(".1f");

/**  */
export const formatSimplePercentage = format(".0%");
export const formatMetricChangePercentage = format("+.1%");
/**
 * changes precision depending on the
 */
export function formatBigNumberPercentage(v) {
  if (v < 0.0001) {
    const f = format(".4%")(v);
    if (f === "0.0000%") {
      return "~ 0%";
    } else {
      return f;
    }
  } else {
    return format(".2%")(v);
  }
}

export function removeTimezoneOffset(dt: Date) {
  return new Date(dt.getTime() + dt.getTimezoneOffset() * 60000);
}

export const standardTimestampFormat = (v, type = "TIMESTAMP") => {
  let fmt = timeFormat("%Y-%m-%d %I:%M:%S");
  if (type === "DATE") {
    fmt = timeFormat("%Y-%m-%d");
  }
  return fmt(removeTimezoneOffset(new Date(v)));
};

export const fullTimestampFormat = (v) => {
  const fmt = timeFormat("%Y-%m-%d %I:%M:%S.%L");
  return fmt(removeTimezoneOffset(new Date(v)));
};

export const datePortion = timeFormat("%Y-%m-%d");
export const timePortion = timeFormat("%I:%M:%S");

export function microsToTimestring(microseconds: number) {
  // to format micros, we need to translate this to hh:mm:ss.
  // start with hours/
  const sign = Math.sign(microseconds);
  const micros = Math.abs(microseconds);
  const hours = ~~(micros / 1000 / 1000 / 60 / 60);
  let remaining = micros - hours * 1000 * 1000 * 60 * 60;
  const minutes = ~~(remaining / 1000 / 1000 / 60);
  //const seconds = (remaining - (minutes * 1000 * 1000 * 60)) / 1000 / 1000;
  remaining -= minutes * 1000 * 1000 * 60;
  const seconds = ~~(remaining / 1000 / 1000);
  remaining -= seconds * 1000 * 1000;
  const ms = ~~(remaining / 1000);
  if (hours === 0 && minutes === 0 && seconds === 0 && ms > 0) {
    return `${sign == 1 ? "" : "-"}${ms}ms`;
  }
  return `${sign == 1 ? "" : "-"}${zeroPad(hours)}:${zeroPad(
    minutes
  )}:${zeroPad(seconds)}.${msPad(ms)}`;
}

/** convert a start and end date
 * to a human readable time range (e.g. 6 months, 24 years, etc)
 */
export function datesToFormattedTimeRange(start: Date, end: Date) {
  const interval = (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24);
  return intervalToTimestring(interval);
}

export function intervalToTimestring(inputInterval: Interval | number) {
  const interval =
    typeof inputInterval === "number"
      ? { months: 0, days: inputInterval, micros: 0 }
      : inputInterval;
  const months = interval.months
    ? `${formatInteger(interval.months)} month${
        interval.months > 1 ? "s" : ""
      } `
    : "";
  const days = interval.days
    ? `${justEnoughPrecision(interval.days)} day${
        interval.days > 1 ? "s" : ""
      } `
    : "";
  const time =
    interval.months > 0 || interval.days > 1
      ? ""
      : microsToTimestring(interval.micros);
  // if only days && days > 365, convert to years?
  if (interval.months === 0 && interval.days > 0 && interval.days > 365)
    return `${formatRate(interval.days / 365)} years`;
  return `${months}${days}${time}`;
}

export function formatCompactInteger(n: number) {
  let fmt: (number) => string;
  if (n <= 1000) {
    fmt = formatInteger;
    return fmt(~~n);
  } else {
    fmt = format(".3s");
    return fmt(n);
  }
}

export function formatDataType(value: any, type: string) {
  if (INTEGERS.has(type) || type.startsWith("DECIMAL")) {
    return value;
  } else if (FLOATS.has(type)) {
    return value;
  } else if (CATEGORICALS.has(type)) {
    return value;
  } else if (TIMESTAMPS.has(type)) {
    return standardTimestampFormat(value, type);
  } else if (INTERVALS.has(type)) {
    return intervalToTimestring(value);
  }
  // list type
  if (type.includes("[]")) {
    return `[${value
      .map((entry) => (+entry ? +entry : `'${entry}'`))
      .join(", ")}]`;
  }
  if (type === "JSON") {
    return value;
  }
  // use this for structs, maps, etc
  return JSON.stringify(value).replace(/"/g, "'");
}

/** These will be used in the string */
export const PreviewRollupIntervalFormatter = {
  [PreviewRollupInterval.ms]:
    "millisecond-level" /** showing rows binned by ms */,
  [PreviewRollupInterval.second]:
    "second-level" /** showing rows binned by second */,
  [PreviewRollupInterval.minute]:
    "minute-level" /** showing rows binned by minute */,
  [PreviewRollupInterval.hour]: "hourly" /** showing hourly counts */,
  [PreviewRollupInterval.day]: "daily" /** showing daily counts */,
  [PreviewRollupInterval.month]: "monthly" /** showing monthly counts */,
  [PreviewRollupInterval.year]: "yearly" /** showing yearly counts */,
};
