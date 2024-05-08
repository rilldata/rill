import type { Interval } from "@rilldata/web-common/lib/duckdb-data-types";
import type {
  FormatterOptionsCommon,
  NumberParts,
  Formatter,
  FormatterRangeSpecsStrategy,
} from "../humanizer-types";

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

export class IntervalFormatter implements Formatter {
  options: FormatterOptionsCommon & FormatterRangeSpecsStrategy;

  stringFormat(x: number): string {
    return formatMsInterval(x);
  }

  partsFormat(x: number): NumberParts {
    return {
      int: formatMsInterval(x),
      dot: "",
      frac: "",
      suffix: "",
    };
  }
}

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
  if (typeof ms !== "number") {
    // FIXME add these warnings back in when the upstream code is robust enough
    // console.warn(
    //   `input to formatMsInterval must be a number, got: ${ms}. Returning empty string.`
    // );
    return "";
  }

  const format = Intl.NumberFormat("en-US", {
    maximumFractionDigits: 1,
    minimumFractionDigits: 0,
    maximumSignificantDigits: 2,
  }).format;

  let neg: "" | "-" = "";
  if (ms < 0) {
    ms = -ms;
    neg = "-";
  }

  switch (true) {
    case ms < 0:
      // THIS SHOULD NEVER HAPPEN, any negative values should
      // have been made positive above.
      console.warn(
        `formatMsInterval: negative value ${ms} was not converted to positive.`,
      );
      return "0 ms";
    case ms === 0:
      return `0 ${timeUnits.s}`;
    case ms < 1:
      return `~0 ${timeUnits.s}`;
    case ms < 100:
      return `${neg}${format(ms)} ${timeUnits.ms}`;
    case ms < 90 * MS_PER_SEC:
      return `${neg}${format(ms / MS_PER_SEC)} ${timeUnits.s}`;
    case ms < 90 * MS_PER_MIN:
      return `${neg}${format(ms / MS_PER_MIN)} ${timeUnits.m}`;
    case ms < 72 * MS_PER_HOUR:
      return `${neg}${format(ms / MS_PER_HOUR)} ${timeUnits.h}`;
    case ms < 90 * MS_PER_DAY:
      return `${neg}${format(ms / MS_PER_DAY)} ${timeUnits.d}`;
    case ms < 18 * MS_PER_MONTH:
      return `${neg}${format(ms / MS_PER_MONTH)} ${timeUnits.mon}`;
    case ms < 100 * MS_PER_YEAR:
      return `${neg}${format(ms / MS_PER_YEAR)} ${timeUnits.y}`;
    default:
      return neg === "-" ? `< -100 ${timeUnits.y}` : `>100 ${timeUnits.y}`;
  }
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
  style: "short" | "units" | "colon" = "short",
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
    (ms - years * MS_PER_YEAR - months * MS_PER_MONTH) / MS_PER_DAY,
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
    if ((value as number) > 0) {
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
  neg: string,
) {
  return [
    [h, timeUnits.h],
    [m, timeUnits.m],
    [s, timeUnits.s],
    [ms, timeUnits.ms],
  ].reduce((acc, [value, unit]) => {
    if ((value as number) > 0) {
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
  neg: string,
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
