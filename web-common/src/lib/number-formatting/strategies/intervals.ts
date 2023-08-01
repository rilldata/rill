/**
 * Formats a millisecond value into a human readable interval.
 *
 * The strategy is to:
 * - show two digits of precision
 * - prefer to show two integer digits in a smaller unit
 * - if that is not possible, show a floating point number in a larger unit with one digit of precision (e.g. 1.2 days)
 *
 * see https://www.notion.so/rilldata/Support-display-of-intervals-and-formatting-of-intervals-e-g-25-days-in-dashboardsal-data-t-8720522eded648f58f35421ebc28ee2f
 */

const SEC = 1000;
const MIN = 60 * SEC;
const HOUR = 60 * MIN;
const DAY = 24 * HOUR;
const MONTH = 30 * DAY;
const YEAR = 365 * DAY;

const ms_breakpoints = [
  { ms: 0 },
  { ms: 1 },
  { ms: 100, divisor: 1, unit: "ms" },
  { ms: 90 * SEC, divisor: SEC, unit: "s" },
  { ms: 90 * MIN, divisor: MIN, unit: "m" },
  { ms: 72 * HOUR, divisor: HOUR, unit: "h" },
  { ms: 90 * DAY, divisor: DAY, unit: "d" },
  { ms: 18 * MONTH, divisor: MONTH, unit: "M" },
  { ms: 100 * YEAR, divisor: YEAR, unit: "y" },
  { ms: Infinity, unit: "TOO_LARGE" },
];

export function formatMsInterval(ms: number): string {
  if (ms === 0) {
    return "0s";
  }
  const i = ms_breakpoints.findIndex((b) => ms < b.ms);

  if (i === 0) {
    // this should never happen unless the input is negative,
    // which is not possible for a valid interval.
    return "<0s";
  }
  if (i === 1) {
    return "<1ms";
  }

  const breakpoint = ms_breakpoints[i];

  if (breakpoint.unit === "TOO_LARGE") {
    return ">100y";
  }

  const unit = breakpoint.unit;
  const value = ms / breakpoint.divisor;
  const fmt = Intl.NumberFormat("en-US", {
    maximumFractionDigits: 1,
    minimumFractionDigits: 0,
    maximumSignificantDigits: 2,
  }).format(value);

  return `${fmt}${unit}`;
}
