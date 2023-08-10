import type { Interval } from "@rilldata/web-common/lib/duckdb-data-types";

const MS_PER_MICRO = 0.001;
const MS_PER_SEC = 1000;
const MS_PER_MIN = 60 * MS_PER_SEC;
const MS_PER_HOUR = 60 * MS_PER_MIN;
const MS_PER_DAY = 24 * MS_PER_HOUR;
const MS_PER_MONTH = 30 * MS_PER_DAY;
const MS_PER_YEAR = 365 * MS_PER_DAY;

const timeUnits = {
  ms: "ms",
  s: "s",
  m: "m",
  h: "h",
  d: "d",
  mon: "mon",
  y: "y",
};

const ms_breakpoints = [
  { ms: 0 },
  { ms: 1 },
  { ms: 100, divisor: 1, unit: timeUnits.ms },
  { ms: 90 * MS_PER_SEC, divisor: MS_PER_SEC, unit: timeUnits.s },
  { ms: 90 * MS_PER_MIN, divisor: MS_PER_MIN, unit: timeUnits.m },
  { ms: 72 * MS_PER_HOUR, divisor: MS_PER_HOUR, unit: timeUnits.h },
  { ms: 90 * MS_PER_DAY, divisor: MS_PER_DAY, unit: timeUnits.d },
  { ms: 18 * MS_PER_MONTH, divisor: MS_PER_MONTH, unit: timeUnits.mon },
  { ms: 100 * MS_PER_YEAR, divisor: MS_PER_YEAR, unit: timeUnits.y },
  { ms: Infinity, unit: "TOO_LARGE" },
];

/**
 * Formats a millisecond value into a compact human readable time interval.
 *
 * The strategy is to:
 * - show two digits of precision
 * - prefer to show two integer digits in a smaller unit
 * - if that is not possible, show a floating point number in a larger unit with one digit of precision (e.g. 1.2 days)
 *
 * see https://www.notion.so/rilldata/Support-display-of-intervals-and-formatting-of-intervals-e-g-25-days-in-dashboardsal-data-t-8720522eded648f58f35421ebc28ee2f
 */
export function formatMsInterval(ms: number): string {
  let negative = false;
  if (ms < 0) {
    ms = -ms;
    negative = true;
  }

  if (ms === 0) {
    return `0 ${timeUnits.s}`;
  } else if (ms < 1) {
    return `~0 ${timeUnits.s}`;
  } else if (ms >= 100 * MS_PER_YEAR) {
    return negative ? `< -100 ${timeUnits.y}` : `>100 ${timeUnits.y}`;
  }

  const i = ms_breakpoints.findIndex((b) => ms < b.ms);

  const breakpoint = ms_breakpoints[i];

  if (breakpoint.unit === "TOO_LARGE") {
    return `>100 ${timeUnits.y}`;
  }

  const unit = breakpoint.unit;
  const value = ms / breakpoint.divisor;
  const fmt = Intl.NumberFormat("en-US", {
    maximumFractionDigits: 1,
    minimumFractionDigits: 0,
    maximumSignificantDigits: 2,
  }).format(value);

  return `${negative ? "-" : ""}${fmt} ${unit}`;
}

/**
 * Formats a millisecond value into an expanded interval string
 * that will be parsable by a duckdb INTERVAL constructor.
 * The hour+min+sec portion will use whichever is shorter between the `HH:MM:SS.xxx`
 * format and a sparse format like `2h 4s` for the HMS part of the interval.
 *
 */
export function formatMsToDuckDbIntervalString(
  ms: number,
  style: "short" | "units" | "colon" = "short"
): string {
  let neg = "";
  if (ms < 0) {
    ms = -ms;
    neg = "-";
  }

  if (ms === 0) {
    return `0${timeUnits.s}`;
  }

  if (ms < 1) {
    return `~0${timeUnits.s}`;
  }

  let string = "";

  const years = Math.floor(ms / MS_PER_YEAR);
  const months = Math.floor((ms - years * MS_PER_YEAR) / MS_PER_MONTH);
  const days = Math.floor(
    (ms - years * MS_PER_YEAR - months * MS_PER_MONTH) / MS_PER_DAY
  );

  const date = new Date(ms);
  const hours = date.getUTCHours();
  const minutes = date.getUTCMinutes();
  const float_seconds = date.getUTCSeconds() + date.getUTCMilliseconds() / 1000;
  const seconds = date.getUTCSeconds();
  const msec = date.getUTCMilliseconds();

  string = [
    [years, timeUnits.y],
    [months, timeUnits.mon],
    [days, timeUnits.d],
  ].reduce((acc, [value, unit]) => {
    if (value > 0) {
      acc += `${neg}${value}${unit} `;
    }
    return acc;
  }, string);

  if (hours === 0 && minutes === 0 && seconds === 0) {
    return string.trim();
  }

  if (style === "units") {
    return string + formatUnitsHMS(hours, minutes, seconds, msec, neg);
  } else if (style === "colon") {
    return string + formatColonHMS(hours, minutes, float_seconds, neg);
  }
  return string + formatShortHMS(hours, minutes, seconds, msec, neg);
}

function formatUnitsHMS(
  h: number,
  m: number,
  s: number,
  ms: number,
  neg: string
) {
  return [
    [h, timeUnits.h],
    [m, timeUnits.m],
    [s, timeUnits.s],
    [ms, timeUnits.ms],
  ].reduce((acc, [value, unit]) => {
    if (value > 0) {
      acc += `${neg}${value}${unit} `;
    }
    return acc;
  }, "");
}

function formatColonHMS(h: number, m: number, s: number, neg: string) {
  const secPad = s < 10 ? "0" : "";
  return `${neg}${h.toString()}:${m
    .toString()
    .padStart(2, "0")}:${secPad}${s.toString()}`;
}

function formatShortHMS(
  h: number,
  m: number,
  s: number,
  msec: number,
  neg: string
) {
  const string1 = formatColonHMS(h, m, s + msec / 1000, neg);
  const string2 = formatUnitsHMS(h, m, s, msec, neg);
  return string1.length < string2.length ? string1 : string2;
}

/**
 * Formats a millisecond value into a human readable time interval in the format HH:MM:SS.xxx, with xxx being milliseconds, and zero padding for minutes and seconds.
 */
export function formatMsIntervalHMS(ms: number) {
  const date = new Date(ms);
  const hours = date.getUTCHours();
  const minutes = date.getUTCMinutes();
  const seconds = date.getUTCSeconds();
  const milliseconds = date.getUTCMilliseconds();

  return `${hours.toString()}:${minutes.toString().padStart(2, "0")}:${seconds
    .toString()
    .padStart(2, "0")}.${milliseconds.toString()}`;
}

function duckdbIntervalToMs(interval: Interval): number {
  return (
    (interval?.months ?? 0) * MS_PER_MONTH +
    (interval?.days ?? 0) * MS_PER_DAY +
    (interval?.micros ?? 0) * MS_PER_MICRO
  );
}

/**
 * Formats a duckdb Interval object of the form
 * `{ months: number, days: number, micros: number }`
 * into a _humanized_ string that can be parsed by a
 * duckdb INTERVAL constructor.
 *
 * NOTE: will be lossy and incorrect in many cases
 * that include a "months" component in the raw interval.
 * It is only intended to be used for approximate dispaly purposes.
 */
export function formatDuckdbIntervalHumane(interval: Interval): string {
  return "~" + formatMsToDuckDbIntervalString(duckdbIntervalToMs(interval));
}

/**
 * Formats a duckdb Interval object of the form
 * `{ months: number, days: number, micros: number }`
 * directly and without humanization into a string that can
 * be parsed by a duckdb INTERVAL constructor.
 *
 * This non-pretty string should theoretically be lossless,
 * as it is a direct representation of duckdb's internal
 * representation of an interval. However, it has not been
 * tested for round-trip correctness.
 */
export function formatDuckdbIntervalLossless(interval: Interval): string {
  return `${interval.months}mon ${interval.days}d ${interval.micros}us`;
}
