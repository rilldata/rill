export const TIME = {
  MILLISECOND: 1,
  get SECOND() {
    return 1000 * this.MILLISECOND;
  },
  get MINUTE() {
    return 60 * this.SECOND;
  },
  get HOUR() {
    return 60 * this.MINUTE;
  },
  get DAY() {
    return 24 * this.HOUR;
  },
  get WEEK() {
    return 7 * this.DAY;
  },
  get MONTH() {
    return 30 * this.DAY;
  },
  get YEAR() {
    return 365 * this.DAY;
  },
};

// Used for luxon's time units
export const TimeUnit = {
  PT1M: "minute",
  PT1H: "hour",
  P1D: "day",
  P1W: "week",
  P1M: "month",
  P3M: "quarter",
  P1Y: "year",
};

/** a Period is a natural duration of time that maps nicely to calendar time.
 * For instance, when we say a day period, we understand this means a 24-hour period
 * that starts at 00:00:00 and ends at 23:59:59.999. These periods are used for
 * time truncation functions.
 */
export enum Period {
  MINUTE = "PT1M",
  HOUR = "PT1H",
  DAY = "P1D",
  WEEK = "P1W",
  MONTH = "P1M",
  QUARTER = "P3M",
  YEAR = "P1Y",
}
